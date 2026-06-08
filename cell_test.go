package main

import "testing"

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
