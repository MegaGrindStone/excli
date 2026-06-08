package main

import (
	"fmt"
	"io"
)

// sheetSummary describes one workbook sheet in JSON output.
type sheetSummary struct {
	Index   int    `json:"index"`
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Visible bool   `json:"visible"`
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

// writeRuntimeError writes a runtime error payload and returns its exit code.
func writeRuntimeError(stderr io.Writer, pretty bool, err error) int {
	if writeErr := writeErrorJSON(stderr, errorCodeRuntime, err.Error(), pretty); writeErr != nil {
		return exitRuntime
	}

	return exitRuntime
}
