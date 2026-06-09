package main

import (
	"fmt"
	"io"

	"github.com/xuri/excelize/v2"
)

// cellValue is the JSON shape for one addressed cell.
type cellValue struct {
	Cell    string `json:"cell"`
	Value   string `json:"value"`
	Formula string `json:"formula,omitempty"`
}

// cellReadResult is the success payload for cell read.
type cellReadResult struct {
	File  string `json:"file"`
	Sheet string `json:"sheet"`

	cellValue
}

// normalizeCellRef canonicalizes an A1-style cell reference.
func normalizeCellRef(cell string) (string, error) {
	col, row, err := excelize.CellNameToCoordinates(cell)
	if err != nil {
		return "", fmt.Errorf("invalid cell reference: %s", cell)
	}

	normalized, err := excelize.CoordinatesToCellName(col, row)
	if err != nil {
		return "", fmt.Errorf("normalize cell reference %q: %w", cell, err)
	}

	return normalized, nil
}

// runCellRead executes the cell read command.
func runCellRead(cmd parsedArgs, stdout io.Writer, stderr io.Writer) int {
	result, err := readCellResult(cmd.file, cmd.sheet, cmd.cell)
	if err != nil {
		return writeRuntimeError(stderr, cmd.pretty, err)
	}

	if err := writeJSON(stdout, result, cmd.pretty); err != nil {
		return writeRuntimeError(stderr, cmd.pretty, err)
	}

	return exitSuccess
}

// readCellResult builds the cell read result for a file, sheet, and cell.
func readCellResult(path, sheetName, cell string) (cellReadResult, error) {
	file, err := openWorkbook(path)
	if err != nil {
		return cellReadResult{}, err
	}

	result, readErr := readCellResultFromWorkbook(file, path, sheetName, cell)
	closeErr := closeWorkbook(file)
	if err := workbookReadError(readErr, closeErr); err != nil {
		return cellReadResult{}, err
	}

	return result, nil
}

// readCellResultFromWorkbook reads one cell from an opened workbook.
func readCellResultFromWorkbook(
	file *excelize.File,
	path string,
	sheetName string,
	cell string,
) (cellReadResult, error) {
	_, sheet, err := resolveSheet(file, sheetName)
	if err != nil {
		return cellReadResult{}, err
	}

	value, err := readCell(file, sheet, cell)
	if err != nil {
		return cellReadResult{}, err
	}

	return cellReadResult{
		File:      path,
		Sheet:     sheet,
		cellValue: value,
	}, nil
}

// readCell reads a normalized cell value and formula.
func readCell(file *excelize.File, sheet, cell string) (cellValue, error) {
	normalized, err := normalizeCellRef(cell)
	if err != nil {
		return cellValue{}, err
	}

	value, err := file.GetCellValue(sheet, normalized)
	if err != nil {
		return cellValue{}, fmt.Errorf("read cell value for %q in %q: %w", normalized, sheet, err)
	}

	formula, err := file.GetCellFormula(sheet, normalized)
	if err != nil {
		return cellValue{}, fmt.Errorf("read cell formula for %q in %q: %w", normalized, sheet, err)
	}

	return cellValue{
		Cell:    normalized,
		Value:   value,
		Formula: formula,
	}, nil
}
