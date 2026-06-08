package main

import (
	"fmt"
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
