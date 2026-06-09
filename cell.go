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

// runCellSet executes the cell set command.
func runCellSet(cmd parsedArgs, stdout io.Writer, stderr io.Writer) int {
	result, err := setCellResult(
		cmd.file,
		cmd.sheet,
		cmd.cell,
		cmd.value,
		cmd.formula,
		cmd.valueSet,
		cmd.formulaSet,
	)
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

// setCellResult writes one cell and returns the mutation success payload.
func setCellResult(
	path string,
	sheetName string,
	cell string,
	value string,
	formula string,
	valueSet bool,
	formulaSet bool,
) (mutationResult, error) {
	file, created, err := openOrCreateWorkbook(path)
	if err != nil {
		return mutationResult{}, err
	}

	mutationErr := setCellInWorkbook(file, sheetName, cell, value, formula, valueSet, formulaSet, created)
	var saveErr error
	if mutationErr == nil {
		saveErr = saveWorkbook(file, path, created)
	}

	closeErr := closeWorkbook(file)
	if err := workbookWriteError(mutationErr, saveErr, closeErr); err != nil {
		return mutationResult{}, err
	}

	return mutationResult{
		File:      path,
		Operation: operationCellSet,
		Success:   true,
	}, nil
}

// setCellInWorkbook applies a cell set mutation to an opened workbook.
func setCellInWorkbook(
	file *excelize.File,
	sheetName string,
	cell string,
	value string,
	formula string,
	valueSet bool,
	formulaSet bool,
	createdWorkbook bool,
) error {
	sheet, err := ensureMutationSheet(file, sheetName, createdWorkbook)
	if err != nil {
		return err
	}

	return setCellValue(file, sheet, cell, value, formula, valueSet, formulaSet)
}

// setCellValue writes a literal string value or formula into one normalized cell.
func setCellValue(
	file *excelize.File,
	sheet string,
	cell string,
	value string,
	formula string,
	valueSet bool,
	formulaSet bool,
) error {
	normalized, err := normalizeCellRef(cell)
	if err != nil {
		return err
	}

	if valueSet {
		if err := file.SetCellStr(sheet, normalized, value); err != nil {
			return fmt.Errorf("set cell value for %q in %q: %w", normalized, sheet, err)
		}

		return nil
	}

	if formulaSet {
		if err := file.SetCellFormula(sheet, normalized, formula); err != nil {
			return fmt.Errorf("set cell formula for %q in %q: %w", normalized, sheet, err)
		}

		return nil
	}

	return fmt.Errorf("missing cell set value or formula")
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
