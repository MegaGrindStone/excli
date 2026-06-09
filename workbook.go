package main

import (
	"errors"
	"fmt"
	"io"
	"os"

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

// openOrCreateWorkbook opens an existing workbook or creates a new one when missing.
func openOrCreateWorkbook(path string) (*excelize.File, bool, error) {
	file, err := excelize.OpenFile(path)
	if err == nil {
		return file, false, nil
	}

	if !errors.Is(err, os.ErrNotExist) {
		return nil, false, fmt.Errorf("open workbook %q: %w", path, err)
	}

	return excelize.NewFile(), true, nil
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

// saveWorkbook persists a workbook through SaveAs for new files or Save for opened files.
func saveWorkbook(file *excelize.File, path string, created bool) error {
	if file == nil {
		return fmt.Errorf("save workbook %q: no workbook open", path)
	}

	if created {
		if err := file.SaveAs(path); err != nil {
			return fmt.Errorf("save workbook %q: %w", path, err)
		}

		return nil
	}

	if err := file.Save(); err != nil {
		return fmt.Errorf("save workbook %q: %w", path, err)
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

// workbookWriteError combines mutation, save, and close errors from write operations.
func workbookWriteError(mutationErr, saveErr, closeErr error) error {
	if mutationErr != nil {
		if closeErr != nil {
			return fmt.Errorf("%w; %s", mutationErr, closeErr.Error())
		}

		return mutationErr
	}

	if saveErr != nil {
		if closeErr != nil {
			return fmt.Errorf("%w; %s", saveErr, closeErr.Error())
		}

		return saveErr
	}

	return closeErr
}
