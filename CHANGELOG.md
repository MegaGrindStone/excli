# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Added a testable Excel CLI runner with help output, structured JSON success/error responses, and compact or pretty formatting.
- Added workbook inspection commands for workbook metadata and sheet listings backed by `excelize`.
- Added argument validation and normalization for workbook, sheet, cell, and range commands, including cell/range reference handling and range-size limits.
- Added unit test coverage for argument parsing, JSON output, CLI execution, workbook helpers, cell references, and range parsing.

### Changed
- Changed the executable entry point to run the CLI and return process exit codes.
