---
name: excli
description: Use the excli command-line tool to inspect and make focused edits to local Excel .xlsx workbooks with deterministic JSON output. Use when a user asks an AI agent to read workbook, sheet, cell, or range data; update a literal value or formula; or clear cells/ranges in local spreadsheets.
license: MIT
compatibility: Requires the excli binary on PATH for command execution; supports local .xlsx workbooks.
---

# excli

Use `excli` for deterministic, non-interactive inspection and focused edits of local Excel `.xlsx` workbooks.

## Before using `excli`

1. Verify the binary is installed and on `PATH`:

   ```bash
   excli version
   ```

2. If `excli` is unavailable, visit `https://github.com/MegaGrindStone/excli` and read its README for the current installation instructions. Offer to help the user install `excli` using those README instructions, then retry after installation.

3. For exact command syntax, output shapes, and examples, read [the command reference](references/commands.md).

## When to use this skill

Use this skill when the user wants to:

- Inspect a local `.xlsx` workbook, sheets, cells, formulas, or rectangular ranges.
- Make a focused in-place cell update with a literal string or formula.
- Clear one cell or a rectangular range of cell value/formula content.
- Automate spreadsheet inspection/editing from an AI agent, script, CI job, or shell workflow.

Do not use `excli` for unsupported spreadsheet work such as `.xls` files, remote spreadsheets, Google Sheets, formula evaluation, styles, comments, charts, images, validations, pivot tables, or rich workbook artifact editing.

## Recommended workflow

1. **Clarify the file and intent.** Confirm the user is working with a local `.xlsx` path and whether they only want inspection or also want edits.
2. **Discover before reading deeply.** Start with `workbook info`, `sheet list`, and/or `sheet info` to identify sheet names and used dimensions.
3. **Read before mutating.** Use `cell read` or `range read` to inspect the current target values before any edit.
4. **Mutate only when needed.** Use `cell set`, `cell clear`, or `range clear` for focused in-place edits.
5. **Verify after mutation.** Read the affected cell or range after writing/clearing and summarize what changed.

## Mutation safety

`excli` mutation commands edit the positional workbook path directly. Recommend that the user make a backup or temp copy before mutations, especially for important files, but do not require one if the user clearly wants direct in-place edits.

When mutating:

- Keep edits narrow and explicit.
- Prefer targeted cells/ranges over broad ranges.
- Remind the user that edits are in-place when the risk matters.
- Use `--value` for literal strings only; it does not infer numbers, booleans, dates, or formulas.
- Use `--formula` for formulas. A leading `=` is accepted but not required.
- Do not claim formulas were recalculated by `excli`; it does not evaluate formulas.

## Parsing results

Workbook, sheet, cell, and range commands return JSON on success. Errors return JSON on `stderr` with a stable `error.code`. `excli version` and help output are plain text exceptions.

Use exit codes as the primary status signal:

- `0`: success
- `1`: runtime error, such as unreadable file or missing sheet
- `2`: usage error, such as missing flags or invalid cell/range references

When JSON output includes a `formula` key, the cell has formula text. The key is omitted for cells without formulas.
