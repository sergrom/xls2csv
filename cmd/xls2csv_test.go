package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLI(t *testing.T) {
	out := filepath.Join(t.TempDir(), "out.csv")
	command := NewRootCommand()
	command.SetArgs([]string{"--xls-file-name=../testdata/sheets.xls", "--csv-file-name=" + out, "--sheet-name=Notes", "--csv-delimiter=,"})
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(out)
	if err != nil || string(got) != "Other sheet\n\"a,b\",=1+1\n" {
		t.Fatal(string(got), err)
	}
	// A fresh command must not inherit the previous sheet or delimiter.
	command = NewRootCommand()
	command.SetArgs([]string{"--xls-file-name=../testdata/sheets.xls", "--csv-file-name=" + out})
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	got, err = os.ReadFile(out)
	if err != nil || !strings.HasPrefix(string(got), "City;Code;Note\n") {
		t.Fatal(string(got), err)
	}
}

func TestInvalidArguments(t *testing.T) {
	base := []string{"--xls-file-name=../testdata/sheets.xls", "--csv-file-name=" + filepath.Join(t.TempDir(), "out.csv")}
	for _, args := range [][]string{
		{}, {"--xls-file-name=x"}, {"--unknown"},
		append(append([]string{}, base...), "extra"),
		append(append([]string{}, base...), "--sheet-index=0", "--sheet-name=Cities"),
		append(append([]string{}, base...), "--sheet-name="),
		append(append([]string{}, base...), "--csv-delimiter="),
		append(append([]string{}, base...), "--sheet-index=-1"),
		append(append([]string{}, base...), "--all-sheets", "--sheet-index=0"),
		append(append([]string{}, base...), "--all-sheets", "--sheet-name=Cities"),
		append(append([]string{}, base...), "--skip-repeated-header"),
		append(append([]string{}, base...), "--all-sheets=false", "--skip-repeated-header"),
	} {
		command := NewRootCommand()
		command.SetArgs(args)
		if err := command.Execute(); err == nil {
			t.Fatalf("accepted %v", args)
		}
	}
}

func TestHelp(t *testing.T) {
	command := NewRootCommand()
	command.SetArgs([]string{"--help"})
	var output bytes.Buffer
	command.SetOut(&output)
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "--sheet-name") {
		t.Fatal(output.String())
	}
}

func TestAllSheetsCLI(t *testing.T) {
	out := filepath.Join(t.TempDir(), "all.csv")
	for _, skip := range []bool{false, true} {
		command := NewRootCommand()
		args := []string{"--xls-file-name=../testdata/headers.xls", "--csv-file-name=" + out, "--all-sheets"}
		if skip {
			args = append(args, "--skip-repeated-header")
		}
		command.SetArgs(args)
		if err := command.Execute(); err != nil {
			t.Fatal(err)
		}
		got, err := os.ReadFile(out)
		if err != nil {
			t.Fatal(err)
		}
		count := 4
		if skip {
			count = 2
		}
		if strings.Count(string(got), "City;Code\n") != count || !strings.HasSuffix(string(got), "Rome;004\n") {
			t.Fatal(string(got))
		}
	}
	command := NewRootCommand()
	command.SetArgs([]string{"--xls-file-name=../testdata/headers.xls", "--csv-file-name=" + out})
	if err := command.Execute(); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(out)
	if err != nil || len(got) != 0 {
		t.Fatal("flags leaked into new command", string(got), err)
	}
}
