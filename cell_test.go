package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCellReadResultJSON(t *testing.T) {
	t.Parallel()

	result := cellReadResult{
		File:  "book.xlsx",
		Sheet: "Sheet1",
		cellValue: cellValue{
			Cell:  "A1",
			Value: "",
		},
	}

	jsonBytes, err := marshalJSON(result, false)
	if err != nil {
		t.Fatalf("marshalJSON returned error: %v", err)
	}

	want := "{\"file\":\"book.xlsx\",\"sheet\":\"Sheet1\",\"cell\":\"A1\",\"value\":\"\"}\n"
	if string(jsonBytes) != want {
		t.Fatalf("cell read JSON = %q, want %q", string(jsonBytes), want)
	}
}

func TestCellReadResultJSONIncludesFormula(t *testing.T) {
	t.Parallel()

	result := cellReadResult{
		File:  "book.xlsx",
		Sheet: "Sheet1",
		cellValue: cellValue{
			Cell:    "C2",
			Value:   "150",
			Formula: "SUM(A2:B2)",
		},
	}

	jsonBytes, err := marshalJSON(result, false)
	if err != nil {
		t.Fatalf("marshalJSON returned error: %v", err)
	}

	want := "{\"file\":\"book.xlsx\",\"sheet\":\"Sheet1\",\"cell\":\"C2\",\"value\":\"150\",\"formula\":\"SUM(A2:B2)\"}\n"
	if string(jsonBytes) != want {
		t.Fatalf("cell read JSON = %q, want %q", string(jsonBytes), want)
	}
}

func TestRunCellSetCreatesWorkbookWithDefaultSheet(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "book.xlsx")
	args := []string{"cell", "set", path, "--cell", "A1", "--value", "created"}

	assertRunStdout(t, args, cellSetSuccessJSON(t, path))
	assertWorkbookSheets(t, path, []string{"Sheet1"})
	assertWorkbookCellValue(t, path, "Sheet1", "A1", "created")
}

func TestRunCellSetCreatesWorkbookWithNamedSheet(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "book.xlsx")
	args := []string{"cell", "set", path, "--sheet", "Budget", "--cell", "B2", "--value", "001"}

	assertRunStdout(t, args, cellSetSuccessJSON(t, path))
	assertWorkbookSheets(t, path, []string{"Budget"})
	assertWorkbookCellValue(t, path, "Budget", "B2", "001")
}

func TestRunCellSetCreatesSheetInExistingWorkbook(t *testing.T) {
	t.Parallel()

	path := createTempWorkbook(t)
	args := []string{"cell", "set", path, "--sheet", "Budget", "--cell", "A1", "--value", "budget"}

	assertRunStdout(t, args, cellSetSuccessJSON(t, path))
	assertWorkbookSheets(t, path, []string{"Sheet1", "Budget"})
	assertWorkbookCellValue(t, path, "Budget", "A1", "budget")
}

func TestRunCellSetWritesLiteralStringValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value string
	}{
		{
			name:  "leading zero",
			value: "001",
		},
		{
			name:  "boolean word",
			value: "true",
		},
		{
			name:  "dash-prefixed number",
			value: "-1",
		},
		{
			name:  "formula-looking string",
			value: "=SUM(A1:A2)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			path := createTempWorkbook(t)
			args := []string{"cell", "set", path, "--cell", "C3", "--value", tt.value}

			assertRunStdout(t, args, cellSetSuccessJSON(t, path))
			assertWorkbookCellValue(t, path, "Sheet1", "C3", tt.value)
			assertWorkbookCellFormula(t, path, "Sheet1", "C3", "")
		})
	}
}

func TestRunCellSetWritesFormulas(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		formula string
	}{
		{
			name:    "without leading equals",
			formula: "SUM(A1:A2)",
		},
		{
			name:    "with leading equals",
			formula: "=SUM(A1:A2)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			path := createTempWorkbook(t)
			args := []string{"cell", "set", path, "--cell", "D4", "--formula", tt.formula}

			assertRunStdout(t, args, cellSetSuccessJSON(t, path))
			assertWorkbookCellFormula(t, path, "Sheet1", "D4", "SUM(A1:A2)")
		})
	}
}

func TestRunCellSetReportsSaveErrors(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "missing-dir", "book.xlsx")
	args := []string{"cell", "set", path, "--cell", "A1", "--value", "created"}

	assertRunRuntimeErrorContains(t, args, "save workbook", path)
}

func TestRunCellClearClearsPlainValue(t *testing.T) {
	t.Parallel()

	path := copyFixtureWorkbook(t, basicFixture)
	args := []string{"cell", "clear", path, "--sheet", "Data", "--cell", "B2"}

	assertRunStdout(t, args, cellClearSuccessJSON(t, path))
	assertWorkbookCellValue(t, path, "Data", "B2", "")
	assertWorkbookCellFormula(t, path, "Data", "B2", "")
	assertWorkbookCellValue(t, path, "Data", "A2", "Alpha")
}

func TestRunCellClearClearsFormula(t *testing.T) {
	t.Parallel()

	path := copyFixtureWorkbook(t, basicFixture)
	args := []string{"cell", "clear", path, "--sheet", "Data", "--cell", "D2"}

	assertRunStdout(t, args, cellClearSuccessJSON(t, path))
	assertWorkbookCellValue(t, path, "Data", "D2", "")
	assertWorkbookCellFormula(t, path, "Data", "D2", "")
	assertWorkbookCellValue(t, path, "Data", "B2", "100")
}

func TestRunCellClearSucceedsForAlreadyEmptyCell(t *testing.T) {
	t.Parallel()

	path := copyFixtureWorkbook(t, basicFixture)
	args := []string{"cell", "clear", path, "--sheet", "Data", "--cell", "C3"}

	assertRunStdout(t, args, cellClearSuccessJSON(t, path))
	assertWorkbookCellValue(t, path, "Data", "C3", "")
	assertWorkbookCellFormula(t, path, "Data", "C3", "")
}

func TestRunCellClearUsesActiveSheetWhenSheetOmitted(t *testing.T) {
	t.Parallel()

	path := createTempWorkbook(t)
	setArgs := []string{"cell", "set", path, "--cell", "A1", "--value", "active"}
	clearArgs := []string{"cell", "clear", path, "--cell", "A1"}

	assertRunStdout(t, setArgs, cellSetSuccessJSON(t, path))
	assertRunStdout(t, clearArgs, cellClearSuccessJSON(t, path))
	assertWorkbookCellValue(t, path, "Sheet1", "A1", "")
	assertWorkbookSheets(t, path, []string{"Sheet1"})
}

func TestRunCellClearMissingWorkbookDoesNotCreateFile(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "missing.xlsx")
	args := []string{"cell", "clear", path, "--cell", "A1"}

	assertRunRuntimeErrorContains(t, args, "open workbook", path)
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("os.Stat error = %v, want missing file", err)
	}
}

func TestRunCellClearMissingSheetDoesNotCreateSheet(t *testing.T) {
	t.Parallel()

	path := createTempWorkbook(t)
	args := []string{"cell", "clear", path, "--sheet", "Missing", "--cell", "A1"}

	assertRunError(t, args, exitRuntime, errorCodeRuntime, "sheet not found: \"Missing\"")
	assertWorkbookSheets(t, path, []string{"Sheet1"})
}

func TestRunCellClearInvalidCellDoesNotMutateWorkbook(t *testing.T) {
	t.Parallel()

	path := createTempWorkbook(t)
	setArgs := []string{"cell", "set", path, "--cell", "A1", "--value", "keep"}
	clearArgs := []string{"cell", "clear", path, "--cell", "A0"}

	assertRunStdout(t, setArgs, cellSetSuccessJSON(t, path))
	assertRunError(t, clearArgs, exitUsage, errorCodeUsage, "invalid cell reference: A0")
	assertWorkbookCellValue(t, path, "Sheet1", "A1", "keep")
}

func TestNormalizeCellRef(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		cell        string
		want        string
		wantMessage string
	}{
		{
			name: "normalizes lowercase",
			cell: "c3",
			want: "C3",
		},
		{
			name:        "rejects invalid row",
			cell:        "A0",
			wantMessage: "invalid cell reference: A0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := normalizeCellRef(tt.cell)
			if tt.wantMessage != "" {
				if err == nil {
					t.Fatal("normalizeCellRef error = nil, want non-nil")
				}

				if err.Error() != tt.wantMessage {
					t.Fatalf("error message = %q, want %q", err.Error(), tt.wantMessage)
				}

				return
			}

			if err != nil {
				t.Fatalf("normalizeCellRef returned error: %v", err)
			}

			if got != tt.want {
				t.Fatalf("normalized cell = %q, want %q", got, tt.want)
			}
		})
	}
}

func cellSetSuccessJSON(t *testing.T, path string) string {
	t.Helper()

	jsonBytes, err := marshalJSON(mutationResult{
		File:      path,
		Operation: operationCellSet,
		Success:   true,
	}, false)
	if err != nil {
		t.Fatalf("marshalJSON returned error: %v", err)
	}

	return string(jsonBytes)
}

func cellClearSuccessJSON(t *testing.T, path string) string {
	t.Helper()

	jsonBytes, err := marshalJSON(mutationResult{
		File:      path,
		Operation: operationCellClear,
		Success:   true,
	}, false)
	if err != nil {
		t.Fatalf("marshalJSON returned error: %v", err)
	}

	return string(jsonBytes)
}

func assertWorkbookSheets(t *testing.T, path string, want []string) {
	t.Helper()

	file, err := openWorkbook(path)
	if err != nil {
		t.Fatalf("openWorkbook returned error: %v", err)
	}
	defer closeTestWorkbook(t, file)

	assertSheetList(t, file, want)
}

func assertWorkbookCellFormula(t *testing.T, path, sheet, cell, want string) {
	t.Helper()

	file, err := openWorkbook(path)
	if err != nil {
		t.Fatalf("openWorkbook returned error: %v", err)
	}
	defer closeTestWorkbook(t, file)

	got, err := file.GetCellFormula(sheet, cell)
	if err != nil {
		t.Fatalf("GetCellFormula returned error: %v", err)
	}

	if got != want {
		t.Fatalf("cell formula = %q, want %q", got, want)
	}
}

func assertRunRuntimeErrorContains(t *testing.T, args []string, substrings ...string) {
	t.Helper()

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

	for _, substring := range substrings {
		if !strings.Contains(payload.Error.Message, substring) {
			t.Fatalf("error message = %q, want to contain %q", payload.Error.Message, substring)
		}
	}
}
