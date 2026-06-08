package main

import (
	"errors"
	"fmt"
	"io"
)

// Exit codes map process results to the Phase 1 CLI contract.
const (
	exitSuccess = 0
	exitRuntime = 1
	exitUsage   = 2
)

// run executes the CLI with injected streams for testing.
func run(args []string, stdout, stderr io.Writer) int {
	if isHelpCommand(args) {
		if err := writeHelp(stdout); err != nil {
			return exitRuntime
		}

		return exitSuccess
	}

	cmd, err := parseArgs(args)
	if err != nil {
		pretty := parseErrorPretty(err)
		if writeErr := writeErrorJSON(stderr, errorCodeUsage, err.Error(), pretty); writeErr != nil {
			return exitRuntime
		}

		return exitUsage
	}

	return dispatch(cmd, stdout, stderr)
}

// isHelpCommand reports whether args request top-level help output.
func isHelpCommand(args []string) bool {
	if len(args) != 1 {
		return false
	}

	switch args[0] {
	case "help", "--help", "-h":
		return true
	default:
		return false
	}
}

// writeHelp writes the top-level help text.
func writeHelp(w io.Writer) error {
	// helpText is the top-level CLI help output.
	const helpText = `excli

Commands:
  excli workbook info <file>
  excli sheet list <file>
  excli sheet info <file> --sheet <name>
  excli cell read <file> --sheet <name> --cell <cell>
  excli range read <file> --sheet <name> --range <range>
`

	if _, err := io.WriteString(w, helpText); err != nil {
		return fmt.Errorf("write help: %w", err)
	}

	return nil
}

// parseErrorPretty reports whether a parse error requested pretty JSON.
func parseErrorPretty(err error) bool {
	var parseErr parseError
	if !errors.As(err, &parseErr) {
		return false
	}

	return parseErr.pretty
}

// dispatch routes a parsed command to its handler.
func dispatch(cmd parsedArgs, stdout, stderr io.Writer) int {
	switch {
	case cmd.resource == resourceWorkbook && cmd.action == actionInfo:
		return runWorkbookInfo(cmd, stdout, stderr)
	case cmd.resource == resourceSheet && cmd.action == actionList:
		return runSheetList(cmd, stdout, stderr)
	case cmd.resource == resourceSheet && cmd.action == actionInfo:
		return runSheetInfo(cmd, stdout, stderr)
	case cmd.resource == resourceCell && cmd.action == actionRead:
		return runCellRead(cmd, stdout, stderr)
	case cmd.resource == resourceRange && cmd.action == actionRead:
		return runRangeRead(cmd, stdout, stderr)
	default:
		if err := writeErrorJSON(stderr, errorCodeRuntime, "internal dispatch error", false); err != nil {
			return exitRuntime
		}

		return exitRuntime
	}
}

// runCellRead returns the current placeholder result for cell reads.
func runCellRead(cmd parsedArgs, _ io.Writer, stderr io.Writer) int {
	return writePlaceholderError(stderr, cmd, "cell read")
}

// runRangeRead returns the current placeholder result for range reads.
func runRangeRead(cmd parsedArgs, _ io.Writer, stderr io.Writer) int {
	return writePlaceholderError(stderr, cmd, "range read")
}

// writePlaceholderError writes a not-yet-implemented runtime error.
func writePlaceholderError(stderr io.Writer, cmd parsedArgs, name string) int {
	message := "command not implemented: " + name
	if err := writeErrorJSON(stderr, errorCodeRuntime, message, cmd.pretty); err != nil {
		return exitRuntime
	}

	return exitRuntime
}
