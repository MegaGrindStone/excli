package main

import (
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

func TestLookupSheetID(t *testing.T) {
	t.Parallel()

	ids := map[int]string{
		1: "Sheet1",
		2: "Budget",
	}

	id, ok := lookupSheetID(ids, "Budget")
	if !ok {
		t.Fatal("lookupSheetID ok = false, want true")
	}

	if id != 2 {
		t.Fatalf("lookupSheetID id = %d, want %d", id, 2)
	}

	missingID, missingOK := lookupSheetID(ids, "Missing")
	if missingOK {
		t.Fatal("lookupSheetID missing ok = true, want false")
	}

	if missingID != 0 {
		t.Fatalf("lookupSheetID missing id = %d, want %d", missingID, 0)
	}
}
