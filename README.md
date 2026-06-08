# excli

`excli` is an agent- and tool-friendly Command Line Interface (CLI) for inspecting local `.xlsx` files. Designed to be easily invoked by both automated AI agents/tools and humans in shell scripts, it exposes workbook/sheet structure, sheet visibility, dimensions, and cell/range values as structured, deterministic JSON.

---

## Installation & Build

Build and run `excli` locally with standard Go development tools.

### Prerequisites

- [Go](https://go.dev/doc/install) (version declared in `go.mod`, currently `1.26.4` or later is recommended)

### Build the Binary

To compile the `excli` binary in the current directory:

```bash
go build -o excli .
```

### Install Globally

To install the binary into your `$GOPATH/bin`:

```bash
go install .
```

---

## Supported Commands

`excli` provides five high-level read-only commands. By default, outputs are emitted as single-line JSON. You can pass the `--pretty` flag to format the JSON output with two-space indentation.

### 1. `excli workbook info <file>`

Inspects the workbook and returns high-level information, including a summary of all worksheets.

* **Example Invocation:**
  ```bash
  excli workbook info testdata/basic.xlsx
  ```
* **Success Output:**
  ```json
  {"file":"testdata/basic.xlsx","sheet_count":1,"sheets":[{"index":0,"id":1,"name":"Data","visible":true}]}
  ```

---

### 2. `excli sheet list <file>`

Lists all sheets in the workbook in sheet order.

* **Example Invocation:**
  ```bash
  excli sheet list testdata/multisheet.xlsx
  ```
* **Success Output:**
  ```json
  {"file":"testdata/multisheet.xlsx","sheets":[{"index":0,"id":1,"name":"Visible","visible":true},{"index":1,"id":2,"name":"Hidden","visible":false}]}
  ```

---

### 3. `excli sheet info <file> --sheet <name>`

Inspects summary details and the used dimension for a specific worksheet.

* **Example Invocation:**
  ```bash
  excli sheet info testdata/basic.xlsx --sheet Data
  ```
* **Success Output:**
  ```json
  {"file":"testdata/basic.xlsx","sheet":{"index":0,"id":1,"name":"Data","visible":true,"dimension":"A1:D4"}}
  ```

---

### 4. `excli cell read <file> --sheet <name> --cell <cell>`

Reads the displayed value and optional formula for a specific cell.

* **Example Invocation:**
  ```bash
  excli cell read testdata/basic.xlsx --sheet Data --cell D2
  ```
* **Success Output:**
  ```json
  {"file":"testdata/basic.xlsx","sheet":"Data","cell":"D2","value":"150","formula":"SUM(B2:C2)"}
  ```
* **Empty Cell Output:**
  If a cell has no value, it returns an empty string. The `formula` key is omitted when no formula is present.
  ```bash
  excli cell read testdata/basic.xlsx --sheet Data --cell C3
  ```
  ```json
  {"file":"testdata/basic.xlsx","sheet":"Data","cell":"C3","value":""}
  ```

---

### 5. `excli range read <file> --sheet <name> --range <range>`

Reads a contiguous rectangular range of cells in row-major order.

* **Example Invocation:**
  ```bash
  excli range read testdata/basic.xlsx --sheet Data --range A1:D4
  ```
* **Success Output:**
  ```json
  {"file":"testdata/basic.xlsx","sheet":"Data","range":"A1:D4","cells":[{"cell":"A1","value":"Name"},{"cell":"B1","value":"Amount"},{"cell":"C1","value":"Bonus"},{"cell":"D1","value":"Total"},{"cell":"A2","value":"Alpha"},{"cell":"B2","value":"100"},{"cell":"C2","value":"50"},{"cell":"D2","value":"150","formula":"SUM(B2:C2)"},{"cell":"A3","value":"Beta"},{"cell":"B3","value":"200"},{"cell":"C3","value":""},{"cell":"D3","value":"200","formula":"SUM(B3:C3)"},{"cell":"A4","value":"Gamma"},{"cell":"B4","value":"0"},{"cell":"C4","value":"25"},{"cell":"D4","value":"25","formula":"SUM(B4:C4)"}]}
  ```

---

## JSON Contract Notes

* **Output Stream & Formatting:** Successful JSON output is written to `stdout` ending with a single trailing newline. It is printed in compact single-line JSON by default, or with a two-space indentation when using `--pretty`.
* **File Path:** The `file` property echoes the user-provided workbook path string exactly.
* **Sheet Indexing & ID:** Worksheet `index` is zero-based (reflecting sheet list order), and `id` corresponds to the workbook sheet ID.
* **Reference Normalization:** Cell references and range syntax are validated and normalized to uppercase, top-left A1-style notation (e.g., `a1` becomes `A1` and `a1:b2` becomes `A1:B2`).
* **Value Extraction:** The `value` fields are strings returned by Excelize based on default formatted cell values. There is no JSON type-inference or casting (e.g., to JSON numbers, booleans, or dates).
* **Formulas:** The `formula` property is included only when a formula is present in the cell; empty formula strings are omitted entirely.
* **Range Reading:** For `range read`, cell objects are emitted in row-major order. Every cell within the requested bounding box is included, with empty cells represented as `"value": ""`.

---

## Global Options

### Pretty Printing (`--pretty`)

Any JSON-producing command supports the `--pretty` flag. This formats success payloads and errors with clear indentation and line breaks.

* **Example Invocation:**
  ```bash
  excli cell read testdata/basic.xlsx --sheet Data --cell D2 --pretty
  ```
* **Output:**
  ```json
  {
    "file": "testdata/basic.xlsx",
    "sheet": "Data",
    "cell": "D2",
    "value": "150",
    "formula": "SUM(B2:C2)"
  }
  ```

---

## Exit Codes and Errors

`excli` uses standardized exit codes and a structured, predictable error contract written to `stderr` in JSON.

### Exit Code Contract

* `0`: Success (Result JSON is written to `stdout`).
* `1`: Runtime / general error (Error JSON is written to `stderr`), e.g., workbook file not found, unreadable sheet, or Excelize runtime errors.
* `2`: Usage / argument validation error (Error JSON is written to `stderr`), e.g., missing required flags, unknown flags, duplicate singleton flags, or invalid cell/range syntax.

### Error Schema

All CLI errors use this exact top-level JSON structure:

```json
{
  "error": {
    "code": "usage_error",
    "message": "missing --sheet"
  }
}
```

* `error.code` is either `"usage_error"` (for exit code `2`) or `"runtime_error"` (for exit code `1`).
* `error.message` is a clear, single-line human-readable error description.

---

## Scope & Boundaries

To ensure stability and reliability when being parsed by agents or shell scripts, `excli` operates under explicit boundaries:

* **Local `.xlsx` files only:** There is no support for legacy `.xls` formats, cloud spreadsheet APIs (e.g., Google Sheets, Office 365), or remote URLs.
* **Read-only inspection:** This is a read-first tool. No mutation or writing capabilities are supported.
* **Value Formatting:** Values are exposed as displayed strings returned by Excelize. No type-inference or casting (e.g., to JSON numbers, booleans, or ISO dates) is performed.
* **Excluded Metadata:** Style details, custom formatting, rich-text annotations, comment threads, embedded charts, and images are intentionally excluded.
* **Maximum cell safety limit:** To avoid resource exhaustion when reading massive ranges, `excli range read` will reject ranges exceeding `10,000` cells with a `usage_error` (exit code `2`).

---

## Shell Scripting Example with `jq`

Since `excli` communicates in structured JSON, it integrates seamlessly into shell pipelines. You can use tools like [`jq`](https://jqlang.github.io/jq/) to parse its outputs.

### Extract all sheet names from a workbook:
```bash
excli sheet list testdata/multisheet.xlsx | jq -r '.sheets[].name'
```

### Retrieve only the cell formula of a target cell:
```bash
excli cell read testdata/basic.xlsx --sheet Data --cell D2 | jq -r '.formula'
```

---

## Development & Verification

### Running Tests

Execute the unit and integration test suite to verify CLI behavior:

```bash
go test ./...
```

### Local Build

Verify that the binary compiles successfully:

```bash
go build
```

### Code Linting

Run static code analysis checks locally:

```bash
golangci-lint run
```

---

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE).
