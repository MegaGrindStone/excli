package main

import "testing"

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
