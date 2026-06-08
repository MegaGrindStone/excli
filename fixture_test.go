package main

import (
	"bytes"
	"encoding/json"
	"testing"
)

const (
	basicFixture      = "testdata/basic.xlsx"
	multisheetFixture = "testdata/multisheet.xlsx"

	expectedBasicRangeReadJSON = "{\"file\":\"testdata/basic.xlsx\",\"sheet\":\"Data\"," +
		"\"range\":\"A1:D4\",\"cells\":[" +
		"{\"cell\":\"A1\",\"value\":\"Name\"}," +
		"{\"cell\":\"B1\",\"value\":\"Amount\"}," +
		"{\"cell\":\"C1\",\"value\":\"Bonus\"}," +
		"{\"cell\":\"D1\",\"value\":\"Total\"}," +
		"{\"cell\":\"A2\",\"value\":\"Alpha\"}," +
		"{\"cell\":\"B2\",\"value\":\"100\"}," +
		"{\"cell\":\"C2\",\"value\":\"50\"}," +
		"{\"cell\":\"D2\",\"value\":\"150\",\"formula\":\"SUM(B2:C2)\"}," +
		"{\"cell\":\"A3\",\"value\":\"Beta\"}," +
		"{\"cell\":\"B3\",\"value\":\"200\"}," +
		"{\"cell\":\"C3\",\"value\":\"\"}," +
		"{\"cell\":\"D3\",\"value\":\"200\",\"formula\":\"SUM(B3:C3)\"}," +
		"{\"cell\":\"A4\",\"value\":\"Gamma\"}," +
		"{\"cell\":\"B4\",\"value\":\"0\"}," +
		"{\"cell\":\"C4\",\"value\":\"25\"}," +
		"{\"cell\":\"D4\",\"value\":\"25\",\"formula\":\"SUM(B4:C4)\"}]}\n"
)

func TestRunFixtureBackedHappyPaths(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		args []string
		want string
	}{
		{
			name: "workbook info",
			args: []string{"workbook", "info", basicFixture},
			want: "{\"file\":\"testdata/basic.xlsx\",\"sheet_count\":1," +
				"\"sheets\":[{\"index\":0,\"id\":1,\"name\":\"Data\",\"visible\":true}]}\n",
		},
		{
			name: "sheet list",
			args: []string{"sheet", "list", multisheetFixture},
			want: "{\"file\":\"testdata/multisheet.xlsx\",\"sheets\":[" +
				"{\"index\":0,\"id\":1,\"name\":\"Visible\",\"visible\":true}," +
				"{\"index\":1,\"id\":2,\"name\":\"Hidden\",\"visible\":false}]}\n",
		},
		{
			name: "sheet info",
			args: []string{"sheet", "info", basicFixture, "--sheet", "Data"},
			want: "{\"file\":\"testdata/basic.xlsx\",\"sheet\":{" +
				"\"index\":0,\"id\":1,\"name\":\"Data\",\"visible\":true," +
				"\"dimension\":\"A1:D4\"}}\n",
		},
		{
			name: "cell read formula",
			args: []string{"cell", "read", basicFixture, "--sheet", "Data", "--cell", "d2"},
			want: "{\"file\":\"testdata/basic.xlsx\",\"sheet\":\"Data\"," +
				"\"cell\":\"D2\",\"value\":\"150\"," +
				"\"formula\":\"SUM(B2:C2)\"}\n",
		},
		{
			name: "range read",
			args: []string{"range", "read", basicFixture, "--sheet", "Data", "--range", "a1:d4"},
			want: expectedBasicRangeReadJSON,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assertRunStdout(t, tt.args, tt.want)
		})
	}
}

func TestRunPrettyFixtureOutput(t *testing.T) {
	t.Parallel()

	args := []string{"cell", "read", basicFixture, "--sheet", "Data", "--cell", "d2", "--pretty"}
	want := "{\n" +
		"  \"file\": \"testdata/basic.xlsx\",\n" +
		"  \"sheet\": \"Data\",\n" +
		"  \"cell\": \"D2\",\n" +
		"  \"value\": \"150\",\n" +
		"  \"formula\": \"SUM(B2:C2)\"\n" +
		"}\n"

	assertRunStdout(t, args, want)
}

func TestRunFixtureBackedEmptyCell(t *testing.T) {
	t.Parallel()

	args := []string{"cell", "read", basicFixture, "--sheet", "Data", "--cell", "c3"}
	want := "{\"file\":\"testdata/basic.xlsx\",\"sheet\":\"Data\"," +
		"\"cell\":\"C3\",\"value\":\"\"}\n"

	assertRunStdout(t, args, want)
}

func TestRunUsageErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		args    []string
		message string
	}{
		{
			name:    "unknown command",
			args:    []string{"workbook", "read", basicFixture},
			message: "unknown command: workbook read",
		},
		{
			name:    "missing file",
			args:    []string{"sheet", "list"},
			message: "missing file",
		},
		{
			name:    "missing sheet",
			args:    []string{"cell", "read", basicFixture, "--cell", "A1"},
			message: "missing --sheet",
		},
		{
			name:    "invalid cell",
			args:    []string{"cell", "read", basicFixture, "--sheet", "Data", "--cell", "A0"},
			message: "invalid cell reference: A0",
		},
		{
			name:    "invalid range",
			args:    []string{"range", "read", basicFixture, "--sheet", "Data", "--range", "A0:B2"},
			message: "invalid range reference: A0:B2",
		},
		{
			name:    "unknown flag",
			args:    []string{"sheet", "list", basicFixture, "--bogus"},
			message: "unknown flag: --bogus",
		},
		{
			name: "duplicate singleton flag",
			args: []string{
				"cell", "read", basicFixture, "--sheet", "Data",
				"--sheet", "Other", "--cell", "A1",
			},
			message: "duplicate flag: --sheet",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assertRunError(t, tt.args, exitUsage, errorCodeUsage, tt.message)
		})
	}
}

func TestRunRuntimeMissingSheetWithFixture(t *testing.T) {
	t.Parallel()

	args := []string{"sheet", "info", basicFixture, "--sheet", "Missing"}

	assertRunError(t, args, exitRuntime, errorCodeRuntime, "sheet not found: \"Missing\"")
}

func assertRunStdout(t *testing.T, args []string, want string) {
	t.Helper()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run(args, &stdout, &stderr)
	if exitCode != exitSuccess {
		t.Fatalf("exit code = %d, want %d; stderr = %q", exitCode, exitSuccess, stderr.String())
	}

	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}

	if stdout.String() != want {
		t.Fatalf("stdout = %q, want %q", stdout.String(), want)
	}
}

func assertRunError(t *testing.T, args []string, wantExit int, wantCode, wantMessage string) {
	t.Helper()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run(args, &stdout, &stderr)
	if exitCode != wantExit {
		t.Fatalf("exit code = %d, want %d", exitCode, wantExit)
	}

	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}

	var payload errorPayload
	if err := json.Unmarshal(stderr.Bytes(), &payload); err != nil {
		t.Fatalf("json.Unmarshal returned error: %v", err)
	}

	if payload.Error.Code != wantCode {
		t.Fatalf("error code = %q, want %q", payload.Error.Code, wantCode)
	}

	if payload.Error.Message != wantMessage {
		t.Fatalf("error message = %q, want %q", payload.Error.Message, wantMessage)
	}
}
