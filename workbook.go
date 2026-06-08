package main

import (
	"fmt"
	"io"
)

// sheetSummary describes one workbook sheet summary in JSON output.
type sheetSummary struct {
	Index   int    `json:"index"`
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Visible bool   `json:"visible"`
}

// sheetDetail describes one workbook sheet with sheet-specific details.
type sheetDetail struct {
	Index     int    `json:"index"`
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Visible   bool   `json:"visible"`
	Dimension string `json:"dimension"`
}

// workbookInfoResult is the success payload for workbook info.
type workbookInfoResult struct {
	File       string         `json:"file"`
	SheetCount int            `json:"sheet_count"`
	Sheets     []sheetSummary `json:"sheets"`
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

// runWorkbookInfo executes the workbook info command.
func runWorkbookInfo(cmd parsedArgs, stdout io.Writer, stderr io.Writer) int {
	result, err := readWorkbookInfo(cmd.file)
	if err != nil {
		return writeRuntimeError(stderr, cmd.pretty, err)
	}

	if err := writeJSON(stdout, result, cmd.pretty); err != nil {
		return writeRuntimeError(stderr, cmd.pretty, err)
	}

	return exitSuccess
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

// readWorkbookInfo builds the workbook info result for a file.
func readWorkbookInfo(path string) (workbookInfoResult, error) {
	sheets, err := readWorkbookSheets(path)
	if err != nil {
		return workbookInfoResult{}, err
	}

	return workbookInfoResult{
		File:       path,
		SheetCount: len(sheets),
		Sheets:     sheetSummaries(sheets),
	}, nil
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

// readSheetInfo builds the sheet info result for a file and sheet name.
func readSheetInfo(path, sheetName string) (sheetInfoResult, error) {
	file, err := openWorkbook(path)
	if err != nil {
		return sheetInfoResult{}, err
	}

	sheet, readErr := lookupSheetInfo(file, sheetName)
	closeErr := closeWorkbook(file)
	if readErr != nil {
		if closeErr != nil {
			return sheetInfoResult{}, fmt.Errorf("%w; %s", readErr, closeErr.Error())
		}

		return sheetInfoResult{}, readErr
	}

	if closeErr != nil {
		return sheetInfoResult{}, closeErr
	}

	return sheetInfoResult{
		File:  path,
		Sheet: sheetDetailFromInfo(sheet),
	}, nil
}

// readWorkbookSheets loads sheet metadata from a workbook file.
func readWorkbookSheets(path string) ([]sheetInfo, error) {
	file, err := openWorkbook(path)
	if err != nil {
		return nil, err
	}

	sheets, readErr := listWorkbookSheets(file)
	closeErr := closeWorkbook(file)
	if readErr != nil {
		if closeErr != nil {
			return nil, fmt.Errorf("%w; %s", readErr, closeErr.Error())
		}

		return nil, readErr
	}

	if closeErr != nil {
		return nil, closeErr
	}

	return sheets, nil
}

// sheetSummaries converts workbook sheet metadata to JSON summaries.
func sheetSummaries(sheets []sheetInfo) []sheetSummary {
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

// sheetDetailFromInfo converts workbook sheet metadata to a JSON detail.
func sheetDetailFromInfo(sheet sheetInfo) sheetDetail {
	return sheetDetail{
		Index:     sheet.index,
		ID:        sheet.id,
		Name:      sheet.name,
		Visible:   sheet.visible,
		Dimension: sheet.dimension,
	}
}

// writeRuntimeError writes a runtime error payload and returns its exit code.
func writeRuntimeError(stderr io.Writer, pretty bool, err error) int {
	if writeErr := writeErrorJSON(stderr, errorCodeRuntime, err.Error(), pretty); writeErr != nil {
		return exitRuntime
	}

	return exitRuntime
}
