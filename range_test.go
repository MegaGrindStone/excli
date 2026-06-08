package main

import "testing"

func TestRangeReadResultJSON(t *testing.T) {
	t.Parallel()

	result := rangeReadResult{
		File:  "book.xlsx",
		Sheet: "Sheet1",
		Range: "A1:B2",
		Cells: []cellValue{
			{Cell: "A1", Value: "Name"},
			{Cell: "B1", Value: "Total"},
			{Cell: "A2", Value: "Alpha"},
			{Cell: "B2", Value: "100", Formula: "SUM(A2:A3)"},
		},
	}

	jsonBytes, err := marshalJSON(result, false)
	if err != nil {
		t.Fatalf("marshalJSON returned error: %v", err)
	}

	want := "{\"file\":\"book.xlsx\",\"sheet\":\"Sheet1\",\"range\":\"A1:B2\",\"cells\":[" +
		"{\"cell\":\"A1\",\"value\":\"Name\"}," +
		"{\"cell\":\"B1\",\"value\":\"Total\"}," +
		"{\"cell\":\"A2\",\"value\":\"Alpha\"}," +
		"{\"cell\":\"B2\",\"value\":\"100\",\"formula\":\"SUM(A2:A3)\"}]}\n"
	if string(jsonBytes) != want {
		t.Fatalf("range read JSON = %q, want %q", string(jsonBytes), want)
	}
}

func TestParseCellRange(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		ref         string
		want        cellRange
		wantCells   []string
		wantMessage string
	}{
		{
			name: "normalizes reversed lowercase range",
			ref:  "b2:a1",
			want: cellRange{
				startCol: 1,
				startRow: 1,
				endCol:   2,
				endRow:   2,
				ref:      "A1:B2",
				count:    4,
			},
			wantCells: []string{"A1", "B1", "A2", "B2"},
		},
		{
			name:        "rejects invalid syntax",
			ref:         "A1",
			wantMessage: "invalid range reference: A1",
		},
		{
			name:        "rejects oversize range",
			ref:         "A1:J1001",
			wantMessage: "range exceeds 10000 cells: A1:J1001",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertParsedRange(t, tt)
		})
	}
}

func assertParsedRange(t *testing.T, tt struct {
	name        string
	ref         string
	want        cellRange
	wantCells   []string
	wantMessage string
}) {
	t.Helper()

	got, err := parseCellRange(tt.ref)
	if tt.wantMessage != "" {
		assertRangeError(t, err, tt.wantMessage)
		return
	}

	if err != nil {
		t.Fatalf("parseCellRange returned error: %v", err)
	}

	if got != tt.want {
		t.Fatalf("parsed range = %#v, want %#v", got, tt.want)
	}

	assertRangeCells(t, got, tt.wantCells)
}

func assertRangeError(t *testing.T, err error, want string) {
	t.Helper()

	if err == nil {
		t.Fatal("parseCellRange error = nil, want non-nil")
	}

	if err.Error() != want {
		t.Fatalf("error message = %q, want %q", err.Error(), want)
	}
}

func assertRangeCells(t *testing.T, rng cellRange, want []string) {
	t.Helper()

	cells, err := rng.cellNames()
	if err != nil {
		t.Fatalf("cellNames returned error: %v", err)
	}

	if len(cells) != len(want) {
		t.Fatalf("len(cellNames) = %d, want %d", len(cells), len(want))
	}

	for i := range cells {
		if cells[i] != want[i] {
			t.Fatalf("cellNames[%d] = %q, want %q", i, cells[i], want[i])
		}
	}
}
