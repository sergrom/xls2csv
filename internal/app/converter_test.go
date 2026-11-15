package app

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestConversion(t *testing.T) {
	cases := []struct {
		name, input, delimiter, sheetName, want string
		index                                   int
	}{
		{"default", "sheets.xls", ";", "", "City;Code;Note\nParis;001;\"a;b\"\n\nBerlin;002;\"say \"\"hello\"\"\"\nMadrid;003;\"line1\nline2\"\nZürich;004;🐧\n", 0},
		{"by index", "sheets.xls", ",", "", "Other sheet\n\"a,b\",=1+1\n", 1},
		{"by name", "sheets.xls", ";", "nOtEs", "Other sheet\na,b;=1+1\n", 0},
		{"unicode delimiter", "sheets.xls", "§", "Notes", "Other sheet\na,b§=1+1\n", 0},
		{"empty sheet", "sheets.xls", ";", "Empty", "", 0},
		{"independent writer", "mixed-xlwt.xls", ";", "", "City;42;-12.5;TRUE;FALSE;001;🐧\n123.45;1.23456789012345;45292;;rich text;after formula\n", 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			output := filepath.Join(t.TempDir(), "out.csv")
			if err := os.WriteFile(output, []byte("old data"), 0600); err != nil {
				t.Fatal(err)
			}
			c, err := NewXls2CsvConverter(filepath.Join("../../testdata", tc.input), output, tc.delimiter)
			if err != nil {
				t.Fatal(err)
			}
			c.WithSheetIndex(tc.index)
			if tc.sheetName != "" {
				c.WithSheetName(tc.sheetName)
			}
			if err := c.Convert(); err != nil {
				t.Fatal(err)
			}
			got, err := os.ReadFile(output)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
		})
	}
}

func TestInvalidOptions(t *testing.T) {
	for _, delimiter := range []string{"", "ab", "\x00", "\n", "\r", "\"", "\xff", "�"} {
		if _, err := NewXls2CsvConverter("in", "out", delimiter); err == nil {
			t.Fatalf("accepted delimiter %q", delimiter)
		}
	}
	for _, paths := range [][2]string{{"", "out"}, {"in", ""}} {
		if _, err := NewXls2CsvConverter(paths[0], paths[1], ";"); err == nil {
			t.Fatal("accepted empty path")
		}
	}
}

func TestFailuresPreserveOutput(t *testing.T) {
	for _, name := range []string{"negative index", "missing index", "missing name", "corrupt input", "missing input"} {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			output := filepath.Join(dir, "out.csv")
			original := []byte("keep me\n")
			if err := os.WriteFile(output, original, 0600); err != nil {
				t.Fatal(err)
			}
			input := "../../testdata/sheets.xls"
			if name == "missing input" {
				input = filepath.Join(dir, "missing.xls")
			}
			if name == "corrupt input" {
				input = filepath.Join(dir, "bad.xls")
				if err := os.WriteFile(input, []byte("bad"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			c, err := NewXls2CsvConverter(input, output, ";")
			if err != nil {
				t.Fatal(err)
			}
			switch name {
			case "negative index":
				c.WithSheetIndex(-1)
			case "missing index":
				c.WithSheetIndex(3)
			case "missing name":
				c.WithSheetName("absent")
			}
			if err := c.Convert(); err == nil {
				t.Fatal("expected error")
			}
			got, err := os.ReadFile(output)
			if err != nil || !bytes.Equal(got, original) {
				t.Fatal("output modified", err)
			}
		})
	}
}

func TestInputCannotBeOverwritten(t *testing.T) {
	for _, alias := range []string{"same", "hard link", "symlink"} {
		t.Run(alias, func(t *testing.T) {
			dir := t.TempDir()
			input := filepath.Join(dir, "in.xls")
			original, err := os.ReadFile("../../testdata/sheets.xls")
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(input, original, 0600); err != nil {
				t.Fatal(err)
			}
			output := input
			if alias != "same" {
				output = filepath.Join(dir, "alias.csv")
				if alias == "hard link" {
					err = os.Link(input, output)
				} else {
					err = os.Symlink(input, output)
				}
				if err != nil {
					t.Skip(err)
				}
			}
			c, _ := NewXls2CsvConverter(input, output, ";")
			if err := c.Convert(); err == nil {
				t.Fatal("input overwrite accepted")
			}
			got, err := os.ReadFile(input)
			if err != nil || !bytes.Equal(got, original) {
				t.Fatal("input modified", err)
			}
		})
	}
}

func TestOutputFailureCleanup(t *testing.T) {
	dir := t.TempDir()
	for _, output := range []string{filepath.Join(dir, "missing", "out.csv"), dir} {
		c, _ := NewXls2CsvConverter("../../testdata/sheets.xls", output, ";")
		if err := c.Convert(); err == nil {
			t.Fatal("expected output error")
		}
	}
	matches, err := filepath.Glob(filepath.Join(filepath.Dir(dir), ".xls2csv-*.tmp"))
	if err != nil || len(matches) != 0 {
		t.Fatal("temporary files left behind", matches, err)
	}
}

func TestAllSheets(t *testing.T) {
	for _, skip := range []bool{false, true} {
		t.Run(fmt.Sprintf("skip=%v", skip), func(t *testing.T) {
			out := filepath.Join(t.TempDir(), "all.csv")
			c, err := NewXls2CsvConverter("../../testdata/headers.xls", out, ";")
			if err != nil {
				t.Fatal(err)
			}
			if err := c.WithAllSheets().WithSkipRepeatedHeader(skip).Convert(); err != nil {
				t.Fatal(err)
			}
			want := "City;Code\nParis;001\nCity;Code\nBerlin;002\nCity;Code\ncity;Code\nMadrid;003\nCity;Code;\nCity;Code\n\nRome;004\n"
			if skip {
				want = "City;Code\nParis;001\nBerlin;002\nCity;Code\ncity;Code\nMadrid;003\nCity;Code;\n\nRome;004\n"
			}
			got, err := os.ReadFile(out)
			if err != nil || string(got) != want {
				t.Fatalf("got %q want %q: %v", got, want, err)
			}
		})
	}
}

func TestSkipHeaderRequiresAllSheets(t *testing.T) {
	out := filepath.Join(t.TempDir(), "out.csv")
	c, err := NewXls2CsvConverter("../../testdata/headers.xls", out, ";")
	if err != nil {
		t.Fatal(err)
	}
	if err := c.WithSkipRepeatedHeader(true).Convert(); err == nil {
		t.Fatal("expected error")
	}
	if _, err := os.Stat(out); !os.IsNotExist(err) {
		t.Fatal("output created", err)
	}
}
