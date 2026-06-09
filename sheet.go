package main

import (
	"fmt"
	"io"

	"github.com/xuri/excelize/v2"
)

// sheetData holds workbook sheet metadata.
type sheetData struct {
	index     int
	id        int
	name      string
	visible   bool
	dimension string
}

// sheetSummary describes one workbook sheet summary in JSON output.
type sheetSummary struct {
	Index   int    `json:"index"`
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Visible bool   `json:"visible"`
}

// sheetDetail describes one workbook sheet with sheet-specific details.
type sheetDetail struct {
	sheetSummary

	Dimension string `json:"dimension"`
}

// sheetListResult is the success payload for sheet list.
type sheetListResult struct {
	File   string         `json:"file"`
	Sheets []sheetSummary `json:"sheets"`
}

// sheetInfoResult is the success payload for sheet info.
type sheetInfoResult struct {
	File  string      `json:"file"`
	Sheet sheetDetail `json:"sheet"`
}

// runSheetList executes the sheet list command.
func runSheetList(cmd parsedArgs, stdout io.Writer, stderr io.Writer) int {
	result, err := readSheetList(cmd.file)
	if err != nil {
		return writeRuntimeError(stderr, cmd.pretty, err)
	}

	if err := writeJSON(stdout, result, cmd.pretty); err != nil {
		return writeRuntimeError(stderr, cmd.pretty, err)
	}

	return exitSuccess
}

// readSheetList builds the sheet list result for a file.
func readSheetList(path string) (sheetListResult, error) {
	sheets, err := readWorkbookSheets(path)
	if err != nil {
		return sheetListResult{}, err
	}

	return sheetListResult{
		File:   path,
		Sheets: sheetSummaries(sheets),
	}, nil
}

// readWorkbookSheets loads sheet metadata from a workbook file.
func readWorkbookSheets(path string) ([]sheetData, error) {
	file, err := openWorkbook(path)
	if err != nil {
		return nil, err
	}

	sheets, readErr := listWorkbookSheets(file)
	closeErr := closeWorkbook(file)
	if err := workbookReadError(readErr, closeErr); err != nil {
		return nil, err
	}

	return sheets, nil
}

// listWorkbookSheets returns sheet metadata in workbook order.
func listWorkbookSheets(file *excelize.File) ([]sheetData, error) {
	names := file.GetSheetList()
	ids := file.GetSheetMap()
	sheets := make([]sheetData, 0, len(names))

	for index, name := range names {
		sheet, err := buildSheetData(file, ids, index, name, false)
		if err != nil {
			return nil, err
		}

		sheets = append(sheets, sheet)
	}

	return sheets, nil
}

// runSheetInfo executes the sheet info command.
func runSheetInfo(cmd parsedArgs, stdout io.Writer, stderr io.Writer) int {
	result, err := readSheetInfo(cmd.file, cmd.sheet)
	if err != nil {
		return writeRuntimeError(stderr, cmd.pretty, err)
	}

	if err := writeJSON(stdout, result, cmd.pretty); err != nil {
		return writeRuntimeError(stderr, cmd.pretty, err)
	}

	return exitSuccess
}

// readSheetInfo builds the sheet info result for a file and sheet name.
func readSheetInfo(path, sheetName string) (sheetInfoResult, error) {
	file, err := openWorkbook(path)
	if err != nil {
		return sheetInfoResult{}, err
	}

	result, readErr := readSheetInfoFromWorkbook(file, path, sheetName)
	closeErr := closeWorkbook(file)
	if err := workbookReadError(readErr, closeErr); err != nil {
		return sheetInfoResult{}, err
	}

	return result, nil
}

// readSheetInfoFromWorkbook builds sheet info from an opened workbook.
func readSheetInfoFromWorkbook(file *excelize.File, path, sheetName string) (sheetInfoResult, error) {
	index, name, err := resolveSheet(file, sheetName)
	if err != nil {
		return sheetInfoResult{}, err
	}

	sheet, err := buildSheetData(file, file.GetSheetMap(), index, name, true)
	if err != nil {
		return sheetInfoResult{}, err
	}

	return sheetInfoResult{
		File: path,
		Sheet: sheetDetail{
			sheetSummary: sheetSummary{
				Index:   sheet.index,
				ID:      sheet.id,
				Name:    sheet.name,
				Visible: sheet.visible,
			},
			Dimension: sheet.dimension,
		},
	}, nil
}

// resolveSheet resolves a sheet reference to its canonical workbook sheet.
func resolveSheet(file *excelize.File, sheet string) (int, string, error) {
	index, err := file.GetSheetIndex(sheet)
	if err != nil {
		return 0, "", fmt.Errorf("get sheet index for %q: %w", sheet, err)
	}

	if index < 0 {
		return 0, "", fmt.Errorf("sheet not found: %q", sheet)
	}

	name := file.GetSheetName(index)
	if name == "" {
		return 0, "", fmt.Errorf("sheet index %d has no name", index)
	}

	return index, name, nil
}

// resolveActiveSheet resolves the workbook's active sheet name.
func resolveActiveSheet(file *excelize.File) (string, error) {
	index := file.GetActiveSheetIndex()
	if index < 0 {
		return "", fmt.Errorf("active sheet not found")
	}

	name := file.GetSheetName(index)
	if name == "" {
		return "", fmt.Errorf("sheet index %d has no name", index)
	}

	return name, nil
}

// ensureMutationSheet resolves or creates a sheet for mutation commands.
func ensureMutationSheet(file *excelize.File, sheetName string, createdWorkbook bool) (string, error) {
	if sheetName == "" {
		return resolveActiveSheet(file)
	}

	index, err := file.GetSheetIndex(sheetName)
	if err != nil {
		return "", fmt.Errorf("get sheet index for %q: %w", sheetName, err)
	}

	if index >= 0 {
		name := file.GetSheetName(index)
		if name == "" {
			return "", fmt.Errorf("sheet index %d has no name", index)
		}

		return name, nil
	}

	if createdWorkbook {
		activeSheet, err := resolveActiveSheet(file)
		if err != nil {
			return "", err
		}

		if err := file.SetSheetName(activeSheet, sheetName); err != nil {
			return "", fmt.Errorf("rename sheet %q to %q: %w", activeSheet, sheetName, err)
		}

		return resolveOptionalSheet(file, sheetName)
	}

	index, err = file.NewSheet(sheetName)
	if err != nil {
		return "", fmt.Errorf("create sheet %q: %w", sheetName, err)
	}

	name := file.GetSheetName(index)
	if name == "" {
		return "", fmt.Errorf("sheet index %d has no name", index)
	}

	return name, nil
}

// resolveOptionalSheet resolves an optional sheet name to an existing sheet.
func resolveOptionalSheet(file *excelize.File, sheetName string) (string, error) {
	if sheetName == "" {
		return resolveActiveSheet(file)
	}

	_, name, err := resolveSheet(file, sheetName)
	if err != nil {
		return "", err
	}

	return name, nil
}

// buildSheetData constructs metadata for one workbook sheet.
func buildSheetData(
	file *excelize.File,
	ids map[int]string,
	index int,
	name string,
	includeDimension bool,
) (sheetData, error) {
	id, ok := lookupSheetID(ids, name)
	if !ok {
		return sheetData{}, fmt.Errorf("sheet ID not found: %q", name)
	}

	visible, err := file.GetSheetVisible(name)
	if err != nil {
		return sheetData{}, fmt.Errorf("get sheet visibility for %q: %w", name, err)
	}

	sheet := sheetData{
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
		return sheetData{}, fmt.Errorf("get sheet dimension for %q: %w", name, err)
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

// sheetSummaries converts workbook sheet metadata to JSON summaries.
func sheetSummaries(sheets []sheetData) []sheetSummary {
	summaries := make([]sheetSummary, 0, len(sheets))

	for _, sheet := range sheets {
		summaries = append(summaries, sheetSummary{
			Index:   sheet.index,
			ID:      sheet.id,
			Name:    sheet.name,
			Visible: sheet.visible,
		})
	}

	return summaries
}
