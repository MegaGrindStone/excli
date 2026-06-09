# excli Command Reference

Use this reference after loading the main `excli` skill instructions. Examples use `book.xlsx` and a sheet named `Data`; adjust paths, sheet names, cells, and ranges for the user's workbook.

## Global conventions

- `excli` works with local `.xlsx` files.
- Quote workbook paths, sheet names, values, and formulas when the shell requires it.
- Workbook, sheet, cell, and range commands accept `--pretty` for two-space-indented JSON.
- Successful workbook commands write JSON to `stdout` followed by a newline.
- Errors write JSON to `stderr` followed by a newline.
- `excli version` and help output are plain text, not JSON.
- Cell references are normalized to uppercase A1 notation, such as `a1` -> `A1`.
- Range references are normalized to uppercase top-left-to-bottom-right A1 notation, such as `b2:a1` -> `A1:B2`.
- `range read` and `range clear` reject ranges larger than `10,000` cells.

## Exit codes and errors

| Exit code | Meaning |
| --- | --- |
| `0` | Success. Workbook commands write result JSON to `stdout`. |
| `1` | Runtime error, such as an unreadable workbook, missing workbook for clear commands, missing sheet, or save failure. |
| `2` | Usage error, such as missing flags, unknown flags, invalid cell/range references, empty formulas, or invalid `--value`/`--formula` combinations. |

Example usage error:

```bash
excli cell read book.xlsx --sheet Data --pretty
```

```json
{
  "error": {
    "code": "usage_error",
    "message": "missing --cell"
  }
}
```

`error.code` is stable and is either `usage_error` or `runtime_error`. Treat `error.message` as human-readable detail; avoid depending on exact wording when possible.

## JSON key reference

| Key | Meaning |
| --- | --- |
| `file` | Workbook path exactly as supplied to the command. |
| `sheet_count` | Number of sheets in the workbook. |
| `sheets` | Array of sheet summary objects in workbook order. |
| `sheet` | For `sheet info`, a sheet detail object. For cell/range reads, the resolved sheet name string. |
| `index` | Zero-based sheet position in workbook order. |
| `id` | Workbook sheet ID. Use as metadata; do not confuse with zero-based `index`. |
| `name` | Sheet name. |
| `visible` | Boolean sheet visibility. |
| `dimension` | Used sheet dimension reported by the workbook, such as `A1:D4`. |
| `cell` | Normalized A1-style cell reference. |
| `value` | Cell display value as a string. Values are not cast to numbers, booleans, dates, or null. Formula cells may have an empty or cached display value because `excli` does not evaluate formulas. |
| `formula` | Formula text when a cell has a formula. Omitted when the cell has no formula. |
| `range` | Normalized A1-style range reference. |
| `cells` | Row-major array of addressed cell objects, including empty cells. |
| `operation` | Mutation operation name: `cell_set`, `cell_clear`, or `range_clear`. |
| `success` | `true` for successful mutation payloads. |
| `error.code` | Machine-readable error category: `usage_error` or `runtime_error`. |
| `error.message` | Human-readable error detail. |

## Commands

### `excli version`

Use this first to verify that `excli` is installed and on `PATH`.

```bash
excli version
```

Example output:

```text
v0.1.0
```

Source builds may print `dev` instead of a release tag.

### `excli workbook info <file>`

Shows workbook-level info and sheet summaries. Use this early to discover the workbook shape.

```bash
excli workbook info book.xlsx --pretty
```

```json
{
  "file": "book.xlsx",
  "sheet_count": 1,
  "sheets": [
    {
      "index": 0,
      "id": 1,
      "name": "Data",
      "visible": true
    }
  ]
}
```

Key notes:

- `sheet_count` is the number of sheets.
- `sheets` preserves workbook order.
- Use `name` values exactly when passing `--sheet`.

### `excli sheet list <file>`

Lists sheets in workbook order.

```bash
excli sheet list book.xlsx --pretty
```

```json
{
  "file": "book.xlsx",
  "sheets": [
    {
      "index": 0,
      "id": 1,
      "name": "Data",
      "visible": true
    }
  ]
}
```

Use this when sheet names or visibility are the main need and workbook-level `sheet_count` is not necessary.

### `excli sheet info <file> --sheet <name>`

Shows one sheet's metadata and used dimension.

```bash
excli sheet info book.xlsx --sheet Data --pretty
```

```json
{
  "file": "book.xlsx",
  "sheet": {
    "index": 0,
    "id": 1,
    "name": "Data",
    "visible": true,
    "dimension": "A1:D4"
  }
}
```

Key notes:

- `dimension` is useful for choosing safe ranges to inspect.
- `--sheet` is required for `sheet info`.

### `excli cell read <file> --sheet <name> --cell <cell>`

Reads one cell's display value and formula text when present.

```bash
excli cell read book.xlsx --sheet Data --cell A1 --pretty
```

```json
{
  "file": "book.xlsx",
  "sheet": "Data",
  "cell": "A1",
  "value": "Name"
}
```

Formula cell example:

```bash
excli cell read book.xlsx --sheet Data --cell D2 --pretty
```

```json
{
  "file": "book.xlsx",
  "sheet": "Data",
  "cell": "D2",
  "value": "150",
  "formula": "SUM(B2:C2)"
}
```

Key notes:

- `formula` appears only when the cell has formula text.
- `value` is always a string.
- `excli` does not evaluate formulas.
- `--sheet` and `--cell` are required for `cell read`.

### `excli range read <file> --sheet <name> --range <range>`

Reads every cell in a rectangular range in row-major order.

```bash
excli range read book.xlsx --sheet Data --range A1:B2 --pretty
```

```json
{
  "file": "book.xlsx",
  "sheet": "Data",
  "range": "A1:B2",
  "cells": [
    {
      "cell": "A1",
      "value": "Name"
    },
    {
      "cell": "B1",
      "value": "Amount"
    },
    {
      "cell": "A2",
      "value": "Alpha"
    },
    {
      "cell": "B2",
      "value": "100"
    }
  ]
}
```

Key notes:

- `cells` is row-major: A1, B1, A2, B2 for `A1:B2`.
- Empty cells are included with `value: ""`.
- `formula` appears per cell only when present.
- `--sheet` and `--range` are required for `range read`.

### `excli cell set <file> --cell <cell> [--sheet <name>] --value <text>`

Writes a literal string to one cell. This mutates the workbook in place.

```bash
excli cell set book.xlsx --sheet Data --cell E1 --value Status --pretty
```

```json
{
  "file": "book.xlsx",
  "operation": "cell_set",
  "success": true
}
```

Verify after writing:

```bash
excli cell read book.xlsx --sheet Data --cell E1 --pretty
```

```json
{
  "file": "book.xlsx",
  "sheet": "Data",
  "cell": "E1",
  "value": "Status"
}
```

Key notes:

- Recommend a backup or temp copy before mutation, but do not require one if the user wants direct edits.
- `--value` writes the exact text as a string. Values such as `001`, `true`, `-1`, or `=SUM(A1:A2)` are not inferred as numbers, booleans, or formulas.
- `--sheet` is optional for `cell set`; when omitted, `excli` writes to the active sheet.
- `cell set` can create a missing workbook. If `--sheet` names a missing sheet, it can create that sheet.
- Use either `--value` or `--formula`, never both.

### `excli cell set <file> --cell <cell> [--sheet <name>] --formula <formula>`

Writes formula text to one cell. This mutates the workbook in place.

```bash
excli cell set book.xlsx --sheet Data --cell E2 --formula '=SUM(B2:C2)' --pretty
```

```json
{
  "file": "book.xlsx",
  "operation": "cell_set",
  "success": true
}
```

Verify after writing:

```bash
excli cell read book.xlsx --sheet Data --cell E2 --pretty
```

```json
{
  "file": "book.xlsx",
  "sheet": "Data",
  "cell": "E2",
  "value": "",
  "formula": "SUM(B2:C2)"
}
```

Key notes:

- A leading `=` is optional for `--formula`.
- Empty formulas are usage errors.
- `excli` writes formula text but does not evaluate formulas or update calculated-value caches.
- `--sheet` is optional for `cell set`; when omitted, `excli` writes to the active sheet.

### `excli cell clear <file> --cell <cell> [--sheet <name>]`

Clears one cell's value/formula content. This mutates the workbook in place.

```bash
excli cell clear book.xlsx --sheet Data --cell E1 --pretty
```

```json
{
  "file": "book.xlsx",
  "operation": "cell_clear",
  "success": true
}
```

Verify after clearing:

```bash
excli cell read book.xlsx --sheet Data --cell E1 --pretty
```

```json
{
  "file": "book.xlsx",
  "sheet": "Data",
  "cell": "E1",
  "value": ""
}
```

Key notes:

- `cell clear` requires an existing workbook.
- If `--sheet` is omitted, it clears from the active sheet.
- If `--sheet` is provided, the sheet must exist.
- Clear operations remove value/formula content only; they do not intentionally remove styles, comments, rows, columns, or other workbook structure.

### `excli range clear <file> --range <range> [--sheet <name>]`

Clears value/formula content for every cell in a rectangular range. This mutates the workbook in place.

```bash
excli range clear book.xlsx --sheet Data --range E1:E2 --pretty
```

```json
{
  "file": "book.xlsx",
  "operation": "range_clear",
  "success": true
}
```

Verify after clearing:

```bash
excli range read book.xlsx --sheet Data --range E1:E2 --pretty
```

```json
{
  "file": "book.xlsx",
  "sheet": "Data",
  "range": "E1:E2",
  "cells": [
    {
      "cell": "E1",
      "value": ""
    },
    {
      "cell": "E2",
      "value": ""
    }
  ]
}
```

Key notes:

- `range clear` requires an existing workbook.
- If `--sheet` is omitted, it clears from the active sheet.
- If `--sheet` is provided, the sheet must exist.
- The range must be rectangular A1 notation and no larger than `10,000` cells.
- Prefer the smallest range that satisfies the user request.
