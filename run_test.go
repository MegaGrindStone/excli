package main

import (
	"bytes"
	"testing"
)

func TestRunWritesUsageErrorForMissingCommand(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run(nil, &stdout, &stderr)

	if exitCode != exitUsage {
		t.Fatalf("exit code = %d, want %d", exitCode, exitUsage)
	}

	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}

	want := "{\"error\":{\"code\":\"usage_error\",\"message\":\"missing command\"}}\n"
	if stderr.String() != want {
		t.Fatalf("stderr = %q, want %q", stderr.String(), want)
	}
}

func TestRunWritesPrettyUsageError(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run([]string{"sheet", "info", "book.xlsx", "--pretty"}, &stdout, &stderr)

	if exitCode != exitUsage {
		t.Fatalf("exit code = %d, want %d", exitCode, exitUsage)
	}

	want := "{\n  \"error\": {\n    \"code\": \"usage_error\",\n    \"message\": \"missing --sheet\"\n  }\n}\n"
	if stderr.String() != want {
		t.Fatalf("stderr = %q, want %q", stderr.String(), want)
	}
}

func TestRunDispatchesValidCommandToPlaceholder(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run([]string{"cell", "read", "book.xlsx", "--sheet", "Budget", "--cell", "A1"}, &stdout, &stderr)

	if exitCode != exitRuntime {
		t.Fatalf("exit code = %d, want %d", exitCode, exitRuntime)
	}

	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}

	want := "{\"error\":{\"code\":\"runtime_error\",\"message\":\"command not implemented: cell read\"}}\n"
	if stderr.String() != want {
		t.Fatalf("stderr = %q, want %q", stderr.String(), want)
	}
}

func TestRunWritesHelp(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := run([]string{"--help"}, &stdout, &stderr)

	if exitCode != exitSuccess {
		t.Fatalf("exit code = %d, want %d", exitCode, exitSuccess)
	}

	want := "excli\n\n" +
		"Commands:\n" +
		"  excli workbook info <file>\n" +
		"  excli sheet list <file>\n" +
		"  excli sheet info <file> --sheet <name>\n" +
		"  excli cell read <file> --sheet <name> --cell <cell>\n" +
		"  excli range read <file> --sheet <name> --range <range>\n"
	if stdout.String() != want {
		t.Fatalf("stdout = %q, want %q", stdout.String(), want)
	}

	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunReturnsRuntimeOnHelpWriteError(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer

	exitCode := run([]string{"--help"}, errWriter{}, &stderr)

	if exitCode != exitRuntime {
		t.Fatalf("exit code = %d, want %d", exitCode, exitRuntime)
	}

	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}
