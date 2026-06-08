package main

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
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
