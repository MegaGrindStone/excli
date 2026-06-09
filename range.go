package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/xuri/excelize/v2"
)

// Range limits cap the number of cells a command may read.
const maxRangeCells = 10000

// cellRange stores a normalized rectangular cell range.
type cellRange struct {
	startCol int
	startRow int
	endCol   int
	endRow   int
	ref      string
	count    int
}

// rangeReadResult is the success payload for range read.
type rangeReadResult struct {
	File  string      `json:"file"`
	Sheet string      `json:"sheet"`
	Range string      `json:"range"`
	Cells []cellValue `json:"cells"`
}

// runRangeRead executes the range read command.
func runRangeRead(cmd parsedArgs, stdout io.Writer, stderr io.Writer) int {
	rng, err := parseCellRange(cmd.cellRange)
	if err != nil {
		return writeUsageError(stderr, cmd.pretty, err)
	}

	result, err := readRangeResult(cmd.file, cmd.sheet, rng)
	if err != nil {
		return writeRuntimeError(stderr, cmd.pretty, err)
	}

	if err := writeJSON(stdout, result, cmd.pretty); err != nil {
		return writeRuntimeError(stderr, cmd.pretty, err)
	}

	return exitSuccess
}

// runRangeClear executes the range clear command.
func runRangeClear(cmd parsedArgs, stdout io.Writer, stderr io.Writer) int {
	rng, err := parseCellRange(cmd.cellRange)
	if err != nil {
		return writeUsageError(stderr, cmd.pretty, err)
	}

	result, err := clearRangeResult(cmd.file, cmd.sheet, rng)
	if err != nil {
		return writeRuntimeError(stderr, cmd.pretty, err)
	}

	if err := writeJSON(stdout, result, cmd.pretty); err != nil {
		return writeRuntimeError(stderr, cmd.pretty, err)
	}

	return exitSuccess
}

// readRangeResult builds the range read result for a file, sheet, and range.
func readRangeResult(path, sheetName string, rng cellRange) (rangeReadResult, error) {
	file, err := openWorkbook(path)
	if err != nil {
		return rangeReadResult{}, err
	}

	result, readErr := readRangeResultFromWorkbook(file, path, sheetName, rng)
	closeErr := closeWorkbook(file)
	if err := workbookReadError(readErr, closeErr); err != nil {
		return rangeReadResult{}, err
	}

	return result, nil
}

// readRangeResultFromWorkbook reads one cell range from an opened workbook.
func readRangeResultFromWorkbook(
	file *excelize.File,
	path string,
	sheetName string,
	rng cellRange,
) (rangeReadResult, error) {
	_, sheet, err := resolveSheet(file, sheetName)
	if err != nil {
		return rangeReadResult{}, err
	}

	cells, err := readRangeCells(file, sheet, rng)
	if err != nil {
		return rangeReadResult{}, err
	}

	return rangeReadResult{
		File:  path,
		Sheet: sheet,
		Range: rng.ref,
		Cells: cells,
	}, nil
}

// readRangeCells reads all cells addressed by a range in row-major order.
func readRangeCells(file *excelize.File, sheet string, rng cellRange) ([]cellValue, error) {
	names, err := rng.cellNames()
	if err != nil {
		return nil, err
	}

	cells := make([]cellValue, 0, len(names))
	for _, name := range names {
		value, err := readCell(file, sheet, name)
		if err != nil {
			return nil, err
		}

		cells = append(cells, value)
	}

	return cells, nil
}

// clearRangeResult clears a cell range and returns the mutation success payload.
func clearRangeResult(path, sheetName string, rng cellRange) (mutationResult, error) {
	file, err := openWorkbook(path)
	if err != nil {
		return mutationResult{}, err
	}

	mutationErr := clearRangeInWorkbook(file, sheetName, rng)
	var saveErr error
	if mutationErr == nil {
		saveErr = saveWorkbook(file, path, false)
	}

	closeErr := closeWorkbook(file)
	if err := workbookWriteError(mutationErr, saveErr, closeErr); err != nil {
		return mutationResult{}, err
	}

	return mutationResult{
		File:      path,
		Operation: operationRangeClear,
		Success:   true,
	}, nil
}

// clearRangeInWorkbook applies a range clear mutation to an opened workbook.
func clearRangeInWorkbook(file *excelize.File, sheetName string, rng cellRange) error {
	sheet, err := resolveOptionalSheet(file, sheetName)
	if err != nil {
		return err
	}

	return clearRangeCells(file, sheet, rng)
}

// clearRangeCells clears each cell addressed by a range in row-major order.
func clearRangeCells(file *excelize.File, sheet string, rng cellRange) error {
	names, err := rng.cellNames()
	if err != nil {
		return err
	}

	for _, name := range names {
		if err := clearCellValue(file, sheet, name); err != nil {
			return err
		}
	}

	return nil
}

// parseCellRange validates and normalizes a cell range reference.
func parseCellRange(ref string) (cellRange, error) {
	startRef, endRef, err := splitRangeRef(ref)
	if err != nil {
		return cellRange{}, err
	}

	startCol, startRow, err := excelize.CellNameToCoordinates(startRef)
	if err != nil {
		return cellRange{}, fmt.Errorf("invalid range reference: %s", ref)
	}

	endCol, endRow, err := excelize.CellNameToCoordinates(endRef)
	if err != nil {
		return cellRange{}, fmt.Errorf("invalid range reference: %s", ref)
	}

	normalized := cellRange{
		startCol: min(startCol, endCol),
		startRow: min(startRow, endRow),
		endCol:   max(startCol, endCol),
		endRow:   max(startRow, endRow),
	}

	normalized.count = normalized.cellCount()
	if normalized.count > maxRangeCells {
		return cellRange{}, fmt.Errorf("range exceeds %d cells: %s", maxRangeCells, ref)
	}

	normalizedRef, err := normalized.normalizedRef()
	if err != nil {
		return cellRange{}, err
	}

	normalized.ref = normalizedRef
	return normalized, nil
}

// splitRangeRef splits a range into start and end references.
func splitRangeRef(ref string) (string, string, error) {
	start, end, found := strings.Cut(ref, ":")
	if !found || start == "" || end == "" || strings.Contains(end, ":") {
		return "", "", fmt.Errorf("invalid range reference: %s", ref)
	}

	return start, end, nil
}

// normalizedRef returns the canonical A1-style range string.
func (r cellRange) normalizedRef() (string, error) {
	start, err := excelize.CoordinatesToCellName(r.startCol, r.startRow)
	if err != nil {
		return "", fmt.Errorf("normalize range start: %w", err)
	}

	end, err := excelize.CoordinatesToCellName(r.endCol, r.endRow)
	if err != nil {
		return "", fmt.Errorf("normalize range end: %w", err)
	}

	return start + ":" + end, nil
}

// cellCount returns the number of cells in the range.
func (r cellRange) cellCount() int {
	cols := int64(r.endCol-r.startCol) + 1
	rows := int64(r.endRow-r.startRow) + 1

	return int(cols * rows)
}

// cellNames expands the range into row-major cell names.
func (r cellRange) cellNames() ([]string, error) {
	cells := make([]string, 0, r.count)

	for row := r.startRow; row <= r.endRow; row++ {
		for col := r.startCol; col <= r.endCol; col++ {
			cell, err := excelize.CoordinatesToCellName(col, row)
			if err != nil {
				return nil, fmt.Errorf("build cell name for (%d,%d): %w", col, row, err)
			}

			cells = append(cells, cell)
		}
	}

	return cells, nil
}
