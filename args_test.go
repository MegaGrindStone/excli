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
