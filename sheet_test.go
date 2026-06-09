package main

import (
	"testing"

	"github.com/xuri/excelize/v2"
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

func TestResolveActiveSheet(t *testing.T) {
	t.Parallel()

	file := excelize.NewFile()
	defer closeTestWorkbook(t, file)

	got, err := resolveActiveSheet(file)
	if err != nil {
		t.Fatalf("resolveActiveSheet returned error: %v", err)
	}

	if got != "Sheet1" {
		t.Fatalf("active sheet = %q, want %q", got, "Sheet1")
	}
}

func TestEnsureMutationSheetOmittedUsesActiveSheet(t *testing.T) {
	t.Parallel()

	file := excelize.NewFile()
	defer closeTestWorkbook(t, file)

	got, err := ensureMutationSheet(file, "", true)
	if err != nil {
		t.Fatalf("ensureMutationSheet returned error: %v", err)
	}

	if got != "Sheet1" {
		t.Fatalf("sheet = %q, want %q", got, "Sheet1")
	}

	assertSheetList(t, file, []string{"Sheet1"})
}

func TestEnsureMutationSheetCreatesMissingSheetForExistingWorkbook(t *testing.T) {
	t.Parallel()

	file := excelize.NewFile()
	defer closeTestWorkbook(t, file)

	got, err := ensureMutationSheet(file, "Budget", false)
	if err != nil {
		t.Fatalf("ensureMutationSheet returned error: %v", err)
	}

	if got != "Budget" {
		t.Fatalf("sheet = %q, want %q", got, "Budget")
	}

	assertSheetList(t, file, []string{"Sheet1", "Budget"})
	assertActiveSheet(t, file, "Sheet1")
}

func TestEnsureMutationSheetRenamesDefaultForCreatedWorkbook(t *testing.T) {
	t.Parallel()

	file := excelize.NewFile()
	defer closeTestWorkbook(t, file)

	got, err := ensureMutationSheet(file, "Budget", true)
	if err != nil {
		t.Fatalf("ensureMutationSheet returned error: %v", err)
	}

	if got != "Budget" {
		t.Fatalf("sheet = %q, want %q", got, "Budget")
	}

	assertSheetList(t, file, []string{"Budget"})
	assertActiveSheet(t, file, "Budget")
}

func TestResolveOptionalSheetRequiresExistingSheet(t *testing.T) {
	t.Parallel()

	file := excelize.NewFile()
	defer closeTestWorkbook(t, file)

	got, err := resolveOptionalSheet(file, "")
	if err != nil {
		t.Fatalf("resolveOptionalSheet returned error: %v", err)
	}

	if got != "Sheet1" {
		t.Fatalf("optional sheet = %q, want %q", got, "Sheet1")
	}

	_, err = resolveOptionalSheet(file, "Missing")
	if err == nil {
		t.Fatal("resolveOptionalSheet missing sheet error = nil, want non-nil")
	}

	want := "sheet not found: \"Missing\""
	if err.Error() != want {
		t.Fatalf("error message = %q, want %q", err.Error(), want)
	}
}

func assertActiveSheet(t *testing.T, file *excelize.File, want string) {
	t.Helper()

	got, err := resolveActiveSheet(file)
	if err != nil {
		t.Fatalf("resolveActiveSheet returned error: %v", err)
	}

	if got != want {
		t.Fatalf("active sheet = %q, want %q", got, want)
	}
}
