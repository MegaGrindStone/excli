# excli

`excli` is a small command-line tool for reading local Excel `.xlsx` workbooks and printing predictable JSON.

Use it when you want to inspect spreadsheets from shell scripts, CI jobs, data-processing pipelines, or AI agents without automating Excel itself. `excli` is read-only, non-interactive, and intentionally focused on a compact set of workbook, sheet, cell, and range inspection commands.

## Features

- Read workbook and worksheet metadata from local `.xlsx` files
- List sheets, including hidden/visible status
- Read individual cells, including formula text when present
- Read rectangular ranges in row-major order
- Emit compact JSON by default, with `--pretty` for formatted output
- Return structured JSON errors on `stderr` with stable exit codes
- Avoid workbook mutation entirely

## Installation

### Install with Go

```bash
go install github.com/MegaGrindStone/excli@latest
```

Make sure your Go bin directory is on your `PATH`:

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
```

### Build from source

```bash
git clone https://github.com/MegaGrindStone/excli.git
cd excli
go build -o excli .
```

`excli` requires the Go version declared in [`go.mod`](go.mod), currently Go `1.26.4`.

## Commands

| Command | Description |
| --- | --- |
| `excli workbook info <file>` | Show workbook-level info and a summary of all sheets. |
| `excli sheet list <file>` | List all sheets in workbook order. |
| `excli sheet info <file> --sheet <name>` | Show one sheet's metadata and used dimension. |
| `excli cell read <file> --sheet <name> --cell <cell>` | Read a single A1-style cell reference. |
| `excli range read <file> --sheet <name> --range <range>` | Read a rectangular A1-style range. |

All commands accept `--pretty` to format JSON output with two-space indentation.

## JSON behavior

`excli` is designed to be easy to parse:

- Successful command output is written to `stdout` as JSON followed by a newline.
- Errors are written to `stderr` as JSON followed by a newline.
- The `file` field echoes the workbook path exactly as provided.
- Sheet `index` values are zero-based and follow workbook sheet order.
- Cell and range references are normalized to uppercase A1 notation. For example, `a1:b2` becomes `A1:B2`.
- Cell `value` fields are always strings and reflect the workbook's formatted display values.
- Values are not cast to JSON numbers, booleans, dates, or nulls.
- `formula` appears only when a cell has formula text.
- `range read` returns every cell in the requested rectangle in row-major order, including empty cells.

## Exit codes and errors

| Exit code | Meaning |
| --- | --- |
| `0` | Success. Result JSON is written to `stdout`. |
| `1` | Runtime error, such as an unreadable file or missing sheet. Error JSON is written to `stderr`. |
| `2` | Usage or argument validation error, such as a missing flag or invalid cell reference. Error JSON is written to `stderr`. |

Error payloads use this shape:

```json
{
  "error": {
    "code": "usage_error",
    "message": "missing --sheet"
  }
}
```

`error.code` is either `usage_error` or `runtime_error`.

## Scope

`excli` currently focuses on reliable read-only inspection.

Supported:

- Local `.xlsx` files
- Workbook, sheet, cell, and range reads
- Sheet visibility and used sheet dimensions
- Formula text for read cells

Not supported:

- Editing or writing workbooks
- Legacy `.xls` files
- Google Sheets, Office 365, or remote URLs
- Styles, comments, rich text, charts, images, validations, pivot tables, or other advanced workbook artifacts
- Reading ranges larger than `10,000` cells

## Development

```bash
go test ./...
go build
```

If you use `golangci-lint` locally:

```bash
golangci-lint run
```

Contributions and issue reports are welcome.

## License

`excli` is released under the MIT License. See [LICENSE](LICENSE).
