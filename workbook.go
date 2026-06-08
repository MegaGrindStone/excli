package main

import (
	"fmt"
	"io"

	"github.com/xuri/excelize/v2"
)

// workbookInfoResult is the success payload for workbook info.
type workbookInfoResult struct {
	File       string         `json:"file"`
	SheetCount int            `json:"sheet_count"`
	Sheets     []sheetSummary `json:"sheets"`
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

// workbookReadError combines read and close errors from workbook operations.
func workbookReadError(readErr, closeErr error) error {
	if readErr != nil {
		if closeErr != nil {
			return fmt.Errorf("%w; %s", readErr, closeErr.Error())
		}

		return readErr
	}

	if closeErr != nil {
		return closeErr
	}

	return nil
}
