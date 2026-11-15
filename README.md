[![build](https://github.com/sergrom/xls2csv/actions/workflows/build.yml/badge.svg)](https://github.com/sergrom/xls2csv/actions/workflows/build.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/sergrom/xls2csv.svg)](https://pkg.go.dev/github.com/sergrom/xls2csv)
[![Latest release](https://img.shields.io/github/v/release/sergrom/xls2csv)](https://github.com/sergrom/xls2csv/releases)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

# xls2csv

A command-line tool for converting binary Excel XLS worksheets to UTF-8 CSV.

Workbooks are read by [xls-reader](https://github.com/sergrom/xls-reader).
This project handles worksheet selection, CSV output, and command-line options.

## Build

Go 1.20 or later is required. From the project directory:

```sh
go build -o xls2csv .
./xls2csv --help
```

## Usage

Convert the first worksheet using the default semicolon delimiter:

```sh
xls2csv --xls-file-name="cities.xls" --csv-file-name="cities.csv"
```

Select a worksheet by name and use a comma delimiter:

```sh
xls2csv --xls-file-name="cities.xls" --csv-file-name="cities.csv" --sheet-name="Cities" --csv-delimiter=","
```

Select the second worksheet by its zero-based index:

```sh
xls2csv --xls-file-name="cities.xls" --csv-file-name="notes.csv" --sheet-index=1
```

Combine all worksheets into one CSV, skipping repeated headers that exactly match
the first nonempty worksheet's header:

```sh
xls2csv --xls-file-name="report.xls" --csv-file-name="report.csv" --all-sheets --skip-repeated-header
```

Omit `--skip-repeated-header` to retain every row, including repeated headers.

## Options

| Flag | Description | Default |
| --- | --- | --- |
| `--xls-file-name` | Input XLS file (required). | — |
| `--csv-file-name` | Output CSV file (required). An existing file is replaced after successful writing. | — |
| `--csv-delimiter` | One valid Unicode character other than a quote, newline, NUL, or replacement character. | `;` |
| `--sheet-index` | Zero-based worksheet index, including hidden sheets. | `0` |
| `--sheet-name` | Case-insensitive worksheet name; the first match is selected. | Empty |
| `--all-sheets` | Combine all worksheets, including hidden sheets, in tab order. | `false` |
| `--skip-repeated-header` | Skip matching first rows on subsequent sheets; requires `--all-sheets`. | `false` |

`--sheet-index` and `--sheet-name` cannot be specified together. Explicitly empty
file names, delimiters, or sheet names are rejected. Missing sheets are errors.
`--all-sheets` cannot be combined with an explicit `--sheet-index` or `--sheet-name`.
Without `--all-sheets`, only the selected worksheet is exported.

In all-sheets mode, worksheets are appended in tab order without sheet names or
extra separator rows. Sheets with no rows are skipped. With
`--skip-repeated-header`, the first row of the first sheet containing rows is the
reference header. Only the first row of each subsequent sheet is compared against
it. A row is skipped only when every value and the number of columns match exactly;
case and whitespace are significant. Different headers and matching rows within
the data are preserved. Leading blank rows are not searched past to find a header.

## Example

A worksheet containing European cities:

| ID | City | Country |
| --- | --- | --- |
| 1 | Paris | France |
| 2 | Berlin | Germany |
| 3 | Madrid | Spain |

Produces:

```csv
ID;City;Country
1;Paris;France
2;Berlin;Germany
3;Madrid;Spain
```

## Conversion behavior

- Output is UTF-8 without a BOM, with LF line endings. Fields containing quotes,
  delimiters, or line breaks are escaped using Go's standard CSV writer.
- Text and textual leading zeros are preserved. Numeric formatting is not applied.
  Dates are exported as Excel serial numbers; currency and percentage display
  formats are not retained.
- Booleans become `TRUE` or `FALSE`; Excel errors retain their names, such as
  `#DIV/0!`. Formulas use their saved results and are not recalculated.
  Unsupported values remain empty strings.
- Rows may have different numbers of columns. Empty rows produce blank lines;
  many CSV readers skip these lines. An empty worksheet produces an empty file.
- Workbook properties, cell formatting, and sheet visibility are not stored in CSV.
- xls-reader loads the entire workbook, including other worksheets and metadata,
  before conversion. Errors in other sheets can prevent export. Input buffering
  is limited to 128 MiB; decoded data requires additional memory.
- XLSX, encrypted files, and other formats unsupported by xls-reader are rejected.
- Input and output must be different files, including symbolic and hard-link aliases.
- Output is written to a temporary file in the destination directory and renamed
  after writing succeeds. The destination directory must already exist. New files
  have private permissions (`0600` on Unix); existing regular-file permissions are
  preserved. An output symlink is replaced rather than followed when saving.
- Failures are reported to standard error with a nonzero exit status.

## Development

```sh
go build ./...
go test -race ./...
go vet ./...
```

Tests cover worksheet selection, combining sheets, repeated headers, CLI validation,
delimiters, CSV escaping, Unicode,
blank rows, numeric and formula results, output replacement, and input protection.
Checked-in fixtures include files generated by xls-writer and the independent
Python xlwt writer. Neither writer is required to build or run the tests.

## License

[MIT](LICENSE).
