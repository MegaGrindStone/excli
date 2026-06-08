package main

import (
	"testing"
)

func TestRunSheetListMissingFilePretty(t *testing.T) {
	t.Parallel()

	assertRuntimeJSONErrorForMissingWorkbook(t, []string{"sheet", "list", "--pretty"})
}

func TestRunSheetInfoMissingFile(t *testing.T) {
	t.Parallel()

	assertRuntimeJSONErrorForMissingWorkbook(t, []string{"sheet", "info", "--sheet", "Budget"})
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
