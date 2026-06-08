package main

import "testing"

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
