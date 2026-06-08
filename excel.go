package main

import (
	"fmt"

	"github.com/xuri/excelize/v2"
)

// sheetInfo holds workbook sheet metadata.
type sheetInfo struct {
	index     int
	id        int
	name      string
	visible   bool
	dimension string
}

// openWorkbook opens an .xlsx workbook from disk.
func openWorkbook(path string) (*excelize.File, error) {
	file, err := excelize.OpenFile(path)
	if err != nil {
		return nil, fmt.Errorf("open workbook %q: %w", path, err)
	}

	return file, nil
}

// closeWorkbook closes an opened workbook when present.
func closeWorkbook(file *excelize.File) error {
	if file == nil {
		return nil
	}

	if err := file.Close(); err != nil {
		return fmt.Errorf("close workbook: %w", err)
	}

	return nil
}

// listWorkbookSheets returns sheet metadata in workbook order.
func listWorkbookSheets(file *excelize.File) ([]sheetInfo, error) {
	names := file.GetSheetList()
	ids := file.GetSheetMap()
	sheets := make([]sheetInfo, 0, len(names))

	for index, name := range names {
		sheet, err := buildSheetInfo(file, ids, index, name, false)
		if err != nil {
			return nil, err
		}

		sheets = append(sheets, sheet)
	}

	return sheets, nil
}

// lookupSheetInfo resolves sheet metadata for one sheet name.
//
//nolint:unused // Reserved for later.
func lookupSheetInfo(file *excelize.File, sheet string) (sheetInfo, error) {
	index, err := file.GetSheetIndex(sheet)
	if err != nil {
		return sheetInfo{}, fmt.Errorf("get sheet index for %q: %w", sheet, err)
	}

	if index < 0 {
		return sheetInfo{}, fmt.Errorf("sheet not found: %q", sheet)
	}

	name := file.GetSheetName(index)
	if name == "" {
		return sheetInfo{}, fmt.Errorf("sheet index %d has no name", index)
	}

	return buildSheetInfo(file, file.GetSheetMap(), index, name, true)
}

// buildSheetInfo constructs metadata for one workbook sheet.
func buildSheetInfo(
	file *excelize.File,
	ids map[int]string,
	index int,
	name string,
	includeDimension bool,
) (sheetInfo, error) {
	id, ok := lookupSheetID(ids, name)
	if !ok {
		return sheetInfo{}, fmt.Errorf("sheet ID not found: %q", name)
	}

	visible, err := file.GetSheetVisible(name)
	if err != nil {
		return sheetInfo{}, fmt.Errorf("get sheet visibility for %q: %w", name, err)
	}

	sheet := sheetInfo{
		index:   index,
		id:      id,
		name:    name,
		visible: visible,
	}

	if !includeDimension {
		return sheet, nil
	}

	dimension, err := file.GetSheetDimension(name)
	if err != nil {
		return sheetInfo{}, fmt.Errorf("get sheet dimension for %q: %w", name, err)
	}

	sheet.dimension = dimension
	return sheet, nil
}

// lookupSheetID finds the workbook sheet ID for a name.
func lookupSheetID(ids map[int]string, name string) (int, bool) {
	for id, candidate := range ids {
		if candidate == name {
			return id, true
		}
	}

	return 0, false
}
