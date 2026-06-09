# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Introduced `excli`, a Go CLI for read-only inspection of local `.xlsx` workbooks.
- Added `workbook info`, `sheet list`, `sheet info`, `cell read`, and `range read` commands with deterministic JSON output.
- Added top-level help, compact/pretty JSON formatting, structured JSON error payloads, and documented exit codes.
- Added cell and range reference validation/normalization, including a 10,000-cell safety limit for range reads.
- Added `cell clear` command to remove a cell's value and formula from a workbook.
- Added `range clear` command to clear all cells within a rectangular range.

