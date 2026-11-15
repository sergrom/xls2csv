package app

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"unicode/utf8"

	"github.com/sergrom/xls-reader"
)

// Xls2CsvConverter converts selected worksheets into a UTF-8 CSV file.
type Xls2CsvConverter struct {
	xlsFileName        string
	csvFileName        string
	csvDelimiter       rune
	sheetIndex         int
	sheetName          string
	allSheets          bool
	skipRepeatedHeader bool
}

// NewXls2CsvConverter configures input, output, and a single-character delimiter.
func NewXls2CsvConverter(xlsFileName, csvFileName, csvDelimiter string) (*Xls2CsvConverter, error) {
	if xlsFileName == "" || csvFileName == "" {
		return nil, fmt.Errorf("xls-file-name and csv-file-name must not be empty")
	}
	delimiter, size := utf8.DecodeRuneInString(csvDelimiter)
	if size != len(csvDelimiter) || size == 0 || delimiter == utf8.RuneError || delimiter == 0 || delimiter == '"' || delimiter == '\r' || delimiter == '\n' {
		return nil, fmt.Errorf("csv delimiter must be one valid character other than a quote, newline, or NUL")
	}
	return &Xls2CsvConverter{xlsFileName: xlsFileName, csvFileName: csvFileName, csvDelimiter: delimiter}, nil
}

// WithSheetIndex selects a zero-based tab index, replacing a previous name selection.
func (c *Xls2CsvConverter) WithSheetIndex(index int) *Xls2CsvConverter {
	c.sheetIndex, c.sheetName, c.allSheets = index, "", false
	return c
}

// WithSheetName selects a case-insensitive worksheet name. Empty selects by index.
func (c *Xls2CsvConverter) WithSheetName(name string) *Xls2CsvConverter {
	c.sheetName, c.allSheets = name, false
	return c
}

// WithAllSheets selects every worksheet in tab order, including hidden sheets.
func (c *Xls2CsvConverter) WithAllSheets() *Xls2CsvConverter {
	c.allSheets = true
	return c
}

// WithSkipRepeatedHeader skips matching first rows of subsequent nonempty sheets.
// It requires all-sheets mode. Matching is exact, including the column count.
func (c *Xls2CsvConverter) WithSkipRepeatedHeader(skip bool) *Xls2CsvConverter {
	c.skipRepeatedHeader = skip
	return c
}

// Convert reads the workbook and replaces the output only after CSV writing succeeds.
func (c *Xls2CsvConverter) Convert() error {
	if c.skipRepeatedHeader && !c.allSheets {
		return fmt.Errorf("skip-repeated-header requires all-sheets")
	}
	if !c.allSheets && c.sheetName == "" && c.sheetIndex < 0 {
		return fmt.Errorf("sheet-index must not be negative")
	}
	inputInfo, err := os.Stat(c.xlsFileName)
	if err != nil {
		return fmt.Errorf("stat XLS: %w", err)
	}
	outputInfo, err := os.Stat(c.csvFileName)
	if err == nil && os.SameFile(inputInfo, outputInfo) {
		return fmt.Errorf("input XLS and output CSV must be different files")
	}
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("stat CSV: %w", err)
	}
	book, err := xlsreader.Open(c.xlsFileName)
	if err != nil {
		return fmt.Errorf("read XLS: %w", err)
	}
	sheets := book.Sheets
	if !c.allSheets {
		var sheet *xlsreader.Sheet
		if c.sheetName != "" {
			var ok bool
			sheet, ok = book.SheetByName(c.sheetName)
			if !ok {
				return fmt.Errorf("worksheet %q not found", c.sheetName)
			}
		} else {
			if c.sheetIndex >= len(book.Sheets) {
				return fmt.Errorf("sheet index %d out of range (workbook has %d sheets)", c.sheetIndex, len(book.Sheets))
			}
			sheet = book.Sheets[c.sheetIndex]
		}
		sheets = []*xlsreader.Sheet{sheet}
	}
	f, err := os.CreateTemp(filepath.Dir(c.csvFileName), ".xls2csv-*.tmp")
	if err != nil {
		return fmt.Errorf("create CSV: %w", err)
	}
	defer os.Remove(f.Name())
	defer f.Close()
	if outputInfo != nil && outputInfo.Mode().IsRegular() {
		if err := f.Chmod(outputInfo.Mode().Perm()); err != nil {
			return fmt.Errorf("set CSV permissions: %w", err)
		}
	}
	w := csv.NewWriter(f)
	w.Comma = c.csvDelimiter
	var header []string
	haveHeader := false
	for _, sheet := range sheets {
		rows := sheet.Rows()
		if len(rows) == 0 {
			continue
		}
		start := 0
		if c.skipRepeatedHeader {
			if !haveHeader {
				header, haveHeader = rows[0], true
			} else if equalRow(header, rows[0]) {
				start = 1
			}
		}
		for i := start; i < len(rows); i++ {
			if err := w.Write(rows[i]); err != nil {
				return fmt.Errorf("write CSV sheet %q row %d: %w", sheet.Name, i+1, err)
			}
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return fmt.Errorf("flush CSV: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("close CSV: %w", err)
	}
	if err := os.Rename(f.Name(), c.csvFileName); err != nil {
		return fmt.Errorf("save CSV: %w", err)
	}
	return nil
}

func equalRow(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
