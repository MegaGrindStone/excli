package main

import (
	"errors"
	"testing"
)

func TestParseArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		args        []string
		want        parsedArgs
		wantMessage string
		wantPretty  bool
	}{
		{
			name: "workbook info valid",
			args: []string{"workbook", "info", "book.xlsx"},
			want: parsedArgs{
				resource: "workbook",
				action:   "info",
				file:     "book.xlsx",
			},
		},
		{
			name: "sheet info valid",
			args: []string{"sheet", "info", "book.xlsx", "--sheet", "Budget"},
			want: parsedArgs{
				resource: "sheet",
				action:   "info",
				file:     "book.xlsx",
				sheet:    "Budget",
			},
		},
		{
			name: "range read valid with pretty",
			args: []string{"range", "read", "book.xlsx", "--sheet", "Budget", "--range", "b2:a1", "--pretty"},
			want: parsedArgs{
				resource:  "range",
				action:    "read",
				file:      "book.xlsx",
				sheet:     "Budget",
				cellRange: "A1:B2",
				pretty:    true,
			},
		},
		{
			name: "cell read normalizes cell reference",
			args: []string{"cell", "read", "book.xlsx", "--sheet", "Budget", "--cell", "b2"},
			want: parsedArgs{
				resource: "cell",
				action:   "read",
				file:     "book.xlsx",
				sheet:    "Budget",
				cell:     "B2",
			},
		},
		{
			name: "cell set valid with empty value and sheet",
			args: []string{"cell", "set", "book.xlsx", "--sheet", "Budget", "--cell", "b2", "--value", ""},
			want: parsedArgs{
				resource: "cell",
				action:   "set",
				file:     "book.xlsx",
				sheet:    "Budget",
				cell:     "B2",
				valueSet: true,
			},
		},
		{
			name: "cell set valid with dash-prefixed value",
			args: []string{"cell", "set", "book.xlsx", "--cell", "b2", "--value", "-1"},
			want: parsedArgs{
				resource: "cell",
				action:   "set",
				file:     "book.xlsx",
				cell:     "B2",
				value:    "-1",
				valueSet: true,
			},
		},
		{
			name: "cell set valid with formula",
			args: []string{"cell", "set", "book.xlsx", "--cell", "b2", "--formula", "SUM(A1:A2)"},
			want: parsedArgs{
				resource:   "cell",
				action:     "set",
				file:       "book.xlsx",
				cell:       "B2",
				formula:    "SUM(A1:A2)",
				formulaSet: true,
			},
		},
		{
			name: "cell set strips one leading formula equals",
			args: []string{"cell", "set", "book.xlsx", "--cell", "b2", "--formula", "=SUM(A1:A2)"},
			want: parsedArgs{
				resource:   "cell",
				action:     "set",
				file:       "book.xlsx",
				cell:       "B2",
				formula:    "SUM(A1:A2)",
				formulaSet: true,
			},
		},
		{
			name: "cell clear valid without sheet",
			args: []string{"cell", "clear", "book.xlsx", "--cell", "b2"},
			want: parsedArgs{
				resource: "cell",
				action:   "clear",
				file:     "book.xlsx",
				cell:     "B2",
			},
		},
		{
			name: "range clear valid with sheet",
			args: []string{"range", "clear", "book.xlsx", "--sheet", "Budget", "--range", "b2:a1"},
			want: parsedArgs{
				resource:  "range",
				action:    "clear",
				file:      "book.xlsx",
				sheet:     "Budget",
				cellRange: "A1:B2",
			},
		},
		{
			name:        "unknown command",
			args:        []string{"workbook", "read", "book.xlsx"},
			wantMessage: "unknown command: workbook read",
		},
		{
			name:        "missing file",
			args:        []string{"sheet", "list"},
			wantMessage: "missing file",
		},
		{
			name:        "unknown flag",
			args:        []string{"sheet", "list", "book.xlsx", "--bogus"},
			wantMessage: "unknown flag: --bogus",
		},
		{
			name:        "missing flag value",
			args:        []string{"sheet", "info", "book.xlsx", "--sheet"},
			wantMessage: "missing value for --sheet",
		},
		{
			name:        "duplicate singleton flag",
			args:        []string{"cell", "read", "book.xlsx", "--sheet", "Budget", "--sheet", "Other", "--cell", "A1"},
			wantMessage: "duplicate flag: --sheet",
		},
		{
			name:        "extra positional argument",
			args:        []string{"workbook", "info", "book.xlsx", "other.xlsx"},
			wantMessage: "unexpected argument: other.xlsx",
		},
		{
			name:        "missing required flag",
			args:        []string{"cell", "read", "book.xlsx", "--sheet", "Budget"},
			wantMessage: "missing --cell",
		},
		{
			name:        "invalid cell reference",
			args:        []string{"cell", "read", "book.xlsx", "--sheet", "Budget", "--cell", "A0"},
			wantMessage: "invalid cell reference: A0",
		},
		{
			name:        "invalid range reference",
			args:        []string{"range", "read", "book.xlsx", "--sheet", "Budget", "--range", "A0:B2"},
			wantMessage: "invalid range reference: A0:B2",
		},
		{
			name:        "cell set rejects both value and formula",
			args:        []string{"cell", "set", "book.xlsx", "--cell", "A1", "--value", "text", "--formula", "SUM(A1:A2)"},
			wantMessage: "cannot combine --value and --formula",
		},
		{
			name:        "cell set rejects missing value and formula",
			args:        []string{"cell", "set", "book.xlsx", "--cell", "A1"},
			wantMessage: "missing --value or --formula",
		},
		{
			name:        "cell set rejects empty formula",
			args:        []string{"cell", "set", "book.xlsx", "--cell", "A1", "--formula", ""},
			wantMessage: "empty --formula",
		},
		{
			name:        "cell set rejects formula containing only equals",
			args:        []string{"cell", "set", "book.xlsx", "--cell", "A1", "--formula", "="},
			wantMessage: "empty --formula",
		},
		{
			name:        "cell set rejects duplicate value flag",
			args:        []string{"cell", "set", "book.xlsx", "--cell", "A1", "--value", "one", "--value", "two"},
			wantMessage: "duplicate flag: --value",
		},
		{
			name: "cell set rejects duplicate formula flag",
			args: []string{
				"cell", "set", "book.xlsx", "--cell", "A1", "--formula", "SUM(A1:A2)", "--formula", "SUM(B1:B2)",
			},
			wantMessage: "duplicate flag: --formula",
		},
		{
			name:        "cell clear rejects value flag",
			args:        []string{"cell", "clear", "book.xlsx", "--cell", "A1", "--value", "text"},
			wantMessage: "flag not allowed: --value",
		},
		{
			name:        "cell clear rejects invalid cell reference",
			args:        []string{"cell", "clear", "book.xlsx", "--cell", "A0"},
			wantMessage: "invalid cell reference: A0",
		},
		{
			name:        "range clear rejects invalid range reference",
			args:        []string{"range", "clear", "book.xlsx", "--range", "A0:B2"},
			wantMessage: "invalid range reference: A0:B2",
		},
		{
			name:        "range clear rejects oversized range",
			args:        []string{"range", "clear", "book.xlsx", "--range", "A1:J1001"},
			wantMessage: "range exceeds 10000 cells: A1:J1001",
		},
		{
			name:        "read command still rejects dash-prefixed sheet value",
			args:        []string{"cell", "read", "book.xlsx", "--sheet", "-Budget", "--cell", "A1"},
			wantMessage: "missing value for --sheet",
		},
		{
			name:        "cell read still requires sheet",
			args:        []string{"cell", "read", "book.xlsx", "--cell", "A1"},
			wantMessage: "missing --sheet",
		},
		{
			name:        "irrelevant flag",
			args:        []string{"workbook", "info", "book.xlsx", "--sheet", "Budget"},
			wantMessage: "flag not allowed: --sheet",
		},
		{
			name:        "pretty preserved on usage error",
			args:        []string{"sheet", "info", "book.xlsx", "--pretty"},
			wantMessage: "missing --sheet",
			wantPretty:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseArgs(tt.args)
			if tt.wantMessage != "" {
				assertParseError(t, err, tt.wantMessage, tt.wantPretty)
				return
			}

			if err != nil {
				t.Fatalf("parseArgs returned error: %v", err)
			}

			if got != tt.want {
				t.Fatalf("parsed args = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func assertParseError(t *testing.T, err error, wantMessage string, wantPretty bool) {
	t.Helper()

	if err == nil {
		t.Fatal("parseArgs error = nil, want non-nil")
	}

	var got parseError
	if !errors.As(err, &got) {
		t.Fatalf("error type = %T, want parseError", err)
	}

	if got.message != wantMessage {
		t.Fatalf("error message = %q, want %q", got.message, wantMessage)
	}

	if got.pretty != wantPretty {
		t.Fatalf("error pretty = %t, want %t", got.pretty, wantPretty)
	}
}
