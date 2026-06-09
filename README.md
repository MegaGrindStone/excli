# excli

`excli` is a small command-line tool for reading and editing local Excel `.xlsx` workbooks while printing predictable JSON for workbook commands.

Use it when you want to inspect or make focused in-place updates to spreadsheets from shell scripts, CI jobs, data-processing pipelines, or AI agents without automating Excel itself. `excli` is non-interactive and intentionally focused on a compact set of workbook, sheet, cell, and range commands.

## Features

- Read workbook and worksheet metadata from local `.xlsx` files
- List sheets, including hidden/visible status
- Read individual cells, including formula text when present
- Read rectangular ranges in row-major order
- Set a single cell to a literal string value or formula
- Clear cell or range value/formula content
- Emit compact JSON by default, with `--pretty` for formatted output
- Return structured JSON errors on `stderr` with stable exit codes
- Apply mutation commands directly to the positional workbook path

## Installation

### Download a GitHub Release archive

Prebuilt archives are published on the [GitHub Releases](https://github.com/MegaGrindStone/excli/releases) page. Download the archive for your operating system and CPU architecture, extract it, and place `excli` (`excli.exe` on Windows) on your `PATH`.

Binaries from GitHub Release archives report their release tag, such as `v0.1.0`, from `excli version`.

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

Binaries built from source with `go build` or `go install` print `dev` from `excli version` unless a version is supplied manually with linker flags, for example:

```bash
go build -ldflags "-X main.version=v0.1.0" -o excli .
```

## Commands

| Command | Description |
| --- | --- |
| `excli version` | Print the plain-text version string. |
| `excli workbook info <file>` | Show workbook-level info and a summary of all sheets. |
| `excli sheet list <file>` | List all sheets in workbook order. |
| `excli sheet info <file> --sheet <name>` | Show one sheet's metadata and used dimension. |
| `excli cell read <file> --sheet <name> --cell <cell>` | Read a single A1-style cell reference. |
| `excli cell set <file> --cell <cell> [--sheet <name>] --value <text>` | Write a literal string to one cell. |
| `excli cell set <file> --cell <cell> [--sheet <name>] --formula <formula>` | Write formula text to one cell. A leading `=` is optional. |
| `excli cell clear <file> --cell <cell> [--sheet <name>]` | Clear one cell's value/formula content. |
| `excli range read <file> --sheet <name> --range <range>` | Read a rectangular A1-style range. |
| `excli range clear <file> --range <range> [--sheet <name>]` | Clear value/formula content for every cell in a range. |

Workbook, sheet, cell, and range commands accept `--pretty` to format JSON output with two-space indentation.

`excli version` prints the version string followed by a newline.

## Write behavior

Mutation commands edit the positional `<file>` directly. Make your own copy first if you need a backup or rollback path.

`cell set` can create a missing workbook. If `--sheet` is omitted, it writes to the workbook's active sheet. If `--sheet` names a missing sheet, `cell set` creates that sheet; for a brand-new workbook, the default sheet is renamed to the requested sheet.

`--value` writes the supplied text as a literal string only. Values such as `001`, `true`, `-1`, or `=SUM(A1:A2)` are not inferred as numbers, booleans, or formulas.

`--formula` writes formula text. You may include or omit one leading `=`; `=SUM(A1:A2)` and `SUM(A1:A2)` write the same formula. Empty formulas are usage errors, and formulas are not evaluated by `excli`.

`cell clear` and `range clear` require an existing workbook and existing named sheet when `--sheet` is provided. If `--sheet` is omitted, they use the active sheet. Clear operations remove only cell value/formula content; they do not intentionally delete sheets, rows, columns, styles, comments, or other workbook structure.

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
- Mutation success payloads contain stable `file`, `operation`, and `success` fields.

Mutation commands return minimal success JSON:

```json
{
  "file": "book.xlsx",
  "operation": "cell_set",
  "success": true
}
```

`operation` is one of `cell_set`, `cell_clear`, or `range_clear`.

## Exit codes and errors

| Exit code | Meaning |
| --- | --- |
| `0` | Success. Result JSON is written to `stdout`. |
| `1` | Runtime error, such as an unreadable file, missing workbook for clear commands, missing sheet for clear commands, or save failure. Error JSON is written to `stderr`. |
| `2` | Usage or argument validation error, such as a missing flag, invalid cell/range reference, invalid `--value`/`--formula` combination, empty formula, unknown flag, or range over `10,000` cells. Error JSON is written to `stderr`. |

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

`excli` currently focuses on reliable local workbook inspection and simple direct edits.

Supported:

- Local `.xlsx` files
- Workbook, sheet, cell, and range reads
- Sheet visibility and used sheet dimensions
- Formula text for read cells
- Single-cell literal string writes
- Single-cell formula writes
- Cell and range value/formula clearing

Not supported:

- Legacy `.xls` files
- Google Sheets, Office 365, or remote URLs
- Writing edited output to a separate destination file
- Automatic backups or rollback workflow
- Typed value inference for numbers, booleans, dates, or nulls
- Formula evaluation or calculated-value caching
- Styles, comments, rich text, charts, images, validations, pivot tables, or other advanced workbook artifact editing
- Reading or clearing ranges larger than `10,000` cells

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
