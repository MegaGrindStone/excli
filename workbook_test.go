package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"
)

func TestOpenWorkbookMissingFile(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "missing.xlsx")

	_, err := openWorkbook(path)
	if err == nil {
		t.Fatal("openWorkbook error = nil, want non-nil")
	}

	if !strings.Contains(err.Error(), "open workbook") {
		t.Fatalf("error message = %q, want to contain %q", err.Error(), "open workbook")
	}
}

func TestOpenOrCreateWorkbookMissingFile(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "created.xlsx")

	file, created, err := openOrCreateWorkbook(path)
	if err != nil {
		t.Fatalf("openOrCreateWorkbook returned error: %v", err)
	}
	defer closeTestWorkbook(t, file)

	if !created {
		t.Fatal("created = false, want true")
	}

	assertSheetList(t, file, []string{"Sheet1"})
}

func TestOpenOrCreateWorkbookExistingFile(t *testing.T) {
	t.Parallel()

	path := createTempWorkbook(t)

	file, created, err := openOrCreateWorkbook(path)
	if err != nil {
		t.Fatalf("openOrCreateWorkbook returned error: %v", err)
	}
	defer closeTestWorkbook(t, file)

	if created {
		t.Fatal("created = true, want false")
	}
}

func TestSaveWorkbookCreatedSavesAs(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "created.xlsx")
	file, created, err := openOrCreateWorkbook(path)
	if err != nil {
		t.Fatalf("openOrCreateWorkbook returned error: %v", err)
	}

	if !created {
		t.Fatal("created = false, want true")
	}

	if err := file.SetCellStr("Sheet1", "A1", "created"); err != nil {
		t.Fatalf("SetCellStr returned error: %v", err)
	}

	if err := saveWorkbook(file, path, created); err != nil {
		t.Fatalf("saveWorkbook returned error: %v", err)
	}
	closeTestWorkbook(t, file)

	assertWorkbookCellValue(t, path, "Sheet1", "A1", "created")
}

func TestSaveWorkbookExistingSaves(t *testing.T) {
	t.Parallel()

	path := createTempWorkbook(t)
	file, created, err := openOrCreateWorkbook(path)
	if err != nil {
		t.Fatalf("openOrCreateWorkbook returned error: %v", err)
	}

	if created {
		t.Fatal("created = true, want false")
	}

	if err := file.SetCellStr("Sheet1", "A1", "saved"); err != nil {
		t.Fatalf("SetCellStr returned error: %v", err)
	}

	if err := saveWorkbook(file, path, created); err != nil {
		t.Fatalf("saveWorkbook returned error: %v", err)
	}
	closeTestWorkbook(t, file)

	assertWorkbookCellValue(t, path, "Sheet1", "A1", "saved")
}

func TestWorkbookWriteErrorPriority(t *testing.T) {
	t.Parallel()

	mutationErr := errors.New("mutation failed")
	saveErr := errors.New("save failed")
	closeErr := errors.New("close failed")
	tests := []struct {
		name        string
		mutationErr error
		saveErr     error
		closeErr    error
		wantErr     error
		wantMessage string
	}{
		{
			name:        "mutation wins and appends close",
			mutationErr: mutationErr,
			saveErr:     saveErr,
			closeErr:    closeErr,
			wantErr:     mutationErr,
			wantMessage: "mutation failed; close failed",
		},
		{
			name:        "save wins and appends close",
			saveErr:     saveErr,
			closeErr:    closeErr,
			wantErr:     saveErr,
			wantMessage: "save failed; close failed",
		},
		{
			name:        "close returned alone",
			closeErr:    closeErr,
			wantErr:     closeErr,
			wantMessage: "close failed",
		},
		{
			name: "nil when no errors",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := workbookWriteError(tt.mutationErr, tt.saveErr, tt.closeErr)
			if tt.wantErr == nil {
				if got != nil {
					t.Fatalf("workbookWriteError = %v, want nil", got)
				}

				return
			}

			if !errors.Is(got, tt.wantErr) {
				t.Fatalf("workbookWriteError = %v, want to wrap %v", got, tt.wantErr)
			}

			if got.Error() != tt.wantMessage {
				t.Fatalf("error message = %q, want %q", got.Error(), tt.wantMessage)
			}
		})
	}
}

func TestCloseWorkbookNil(t *testing.T) {
	t.Parallel()

	if err := closeWorkbook(nil); err != nil {
		t.Fatalf("closeWorkbook(nil) returned error: %v", err)
	}
}

func TestRunWorkbookInfoMissingFile(t *testing.T) {
	t.Parallel()

	assertRuntimeJSONErrorForMissingWorkbook(t, []string{"workbook", "info"})
}

func TestReadWorkbookInfoUsesUserPath(t *testing.T) {
	t.Parallel()

	result := workbookInfoResult{
		File:       "./book.xlsx",
		SheetCount: 2,
		Sheets: []sheetSummary{
			{Index: 0, ID: 1, Name: "Sheet1", Visible: true},
			{Index: 1, ID: 2, Name: "Sheet2", Visible: false},
		},
	}

	jsonBytes, err := marshalJSON(result, false)
	if err != nil {
		t.Fatalf("marshalJSON returned error: %v", err)
	}

	var decoded struct {
		File       string         `json:"file"`
		SheetCount int            `json:"sheet_count"`
		Sheets     []sheetSummary `json:"sheets"`
	}
	if err := json.Unmarshal(jsonBytes, &decoded); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}

	if decoded.File != result.File {
		t.Fatalf("file = %q, want %q", decoded.File, result.File)
	}

	if decoded.SheetCount != 2 {
		t.Fatalf("sheet_count = %d, want %d", decoded.SheetCount, 2)
	}

	if len(decoded.Sheets) != 2 {
		t.Fatalf("len(sheets) = %d, want %d", len(decoded.Sheets), 2)
	}
}

func createTempWorkbook(t *testing.T) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "book.xlsx")
	file := excelize.NewFile()
	if err := file.SaveAs(path); err != nil {
		t.Fatalf("SaveAs returned error: %v", err)
	}
	closeTestWorkbook(t, file)

	return path
}

func closeTestWorkbook(t *testing.T, file *excelize.File) {
	t.Helper()

	if err := closeWorkbook(file); err != nil {
		t.Fatalf("closeWorkbook returned error: %v", err)
	}
}

func assertSheetList(t *testing.T, file *excelize.File, want []string) {
	t.Helper()

	got := file.GetSheetList()
	if len(got) != len(want) {
		t.Fatalf("sheet list = %v, want %v", got, want)
	}

	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("sheet list = %v, want %v", got, want)
		}
	}
}

func assertWorkbookCellValue(t *testing.T, path, sheet, cell, want string) {
	t.Helper()

	file, err := openWorkbook(path)
	if err != nil {
		t.Fatalf("openWorkbook returned error: %v", err)
	}
	defer closeTestWorkbook(t, file)

	got, err := file.GetCellValue(sheet, cell)
	if err != nil {
		t.Fatalf("GetCellValue returned error: %v", err)
	}

	if got != want {
		t.Fatalf("cell value = %q, want %q", got, want)
	}
}

func assertRuntimeJSONErrorForMissingWorkbook(t *testing.T, prefixArgs []string) {
	t.Helper()

	path := filepath.Join(t.TempDir(), "missing.xlsx")
	args := append(append([]string{}, prefixArgs...), path)

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run(args, &stdout, &stderr)
	if exitCode != exitRuntime {
		t.Fatalf("exit code = %d, want %d", exitCode, exitRuntime)
	}

	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}

	var payload errorPayload
	if err := json.Unmarshal(stderr.Bytes(), &payload); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}

	if payload.Error.Code != errorCodeRuntime {
		t.Fatalf("error code = %q, want %q", payload.Error.Code, errorCodeRuntime)
	}

	if !strings.Contains(payload.Error.Message, "open workbook") {
		t.Fatalf("error message = %q, want to contain %q", payload.Error.Message, "open workbook")
	}

	if !strings.Contains(payload.Error.Message, path) {
		t.Fatalf("error message = %q, want to contain %q", payload.Error.Message, path)
	}
}
