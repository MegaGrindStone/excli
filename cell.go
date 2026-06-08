package main

import (
	"fmt"

	"github.com/xuri/excelize/v2"
)

// cellData holds a cell read result.
//
//nolint:unused // Reserved for later.
type cellData struct {
	cell    string
	value   string
	formula string
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

// readCell reads a normalized cell value and formula.
//
//nolint:unused // Reserved for later.
func readCell(file *excelize.File, sheet, cell string) (cellData, error) {
	normalized, err := normalizeCellRef(cell)
	if err != nil {
		return cellData{}, err
	}

	value, err := file.GetCellValue(sheet, normalized)
	if err != nil {
		return cellData{}, fmt.Errorf("read cell value for %q in %q: %w", normalized, sheet, err)
	}

	formula, err := file.GetCellFormula(sheet, normalized)
	if err != nil {
		return cellData{}, fmt.Errorf("read cell formula for %q in %q: %w", normalized, sheet, err)
	}

	return cellData{
		cell:    normalized,
		value:   value,
		formula: formula,
	}, nil
}
