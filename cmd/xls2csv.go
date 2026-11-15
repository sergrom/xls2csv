package cmd

import (
	"fmt"

	"github.com/sergrom/xls2csv/internal/app"
	"github.com/spf13/cobra"
)

// NewRootCommand creates an independent command with its own flag values.
func NewRootCommand() *cobra.Command {
	var input, output, delimiter, sheetName string
	var sheetIndex int
	var allSheets, skipRepeatedHeader bool
	command := &cobra.Command{
		Use:           "xls2csv",
		Short:         "Convert XLS worksheets to CSV",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Flags().Changed("sheet-index") && cmd.Flags().Changed("sheet-name") {
				return fmt.Errorf("sheet-index and sheet-name are mutually exclusive")
			}
			if cmd.Flags().Changed("sheet-name") && sheetName == "" {
				return fmt.Errorf("sheet-name must not be empty")
			}
			if allSheets && (cmd.Flags().Changed("sheet-index") || cmd.Flags().Changed("sheet-name")) {
				return fmt.Errorf("all-sheets cannot be combined with sheet-index or sheet-name")
			}
			if skipRepeatedHeader && !allSheets {
				return fmt.Errorf("skip-repeated-header requires all-sheets")
			}
			converter, err := app.NewXls2CsvConverter(input, output, delimiter)
			if err != nil {
				return err
			}
			if allSheets {
				converter.WithAllSheets().WithSkipRepeatedHeader(skipRepeatedHeader)
			} else if sheetName != "" {
				converter.WithSheetName(sheetName)
			} else {
				converter.WithSheetIndex(sheetIndex)
			}
			return converter.Convert()
		},
	}
	flags := command.Flags()
	flags.StringVar(&input, "xls-file-name", "", "Input XLS file (required)")
	flags.StringVar(&output, "csv-file-name", "", "Output CSV file (required)")
	_ = command.MarkFlagRequired("xls-file-name")
	_ = command.MarkFlagRequired("csv-file-name")
	flags.StringVar(&delimiter, "csv-delimiter", ";", "CSV field delimiter (one character)")
	flags.IntVar(&sheetIndex, "sheet-index", 0, "Zero-based worksheet index")
	flags.StringVar(&sheetName, "sheet-name", "", "Worksheet name (case-insensitive; replaces sheet-index)")
	flags.BoolVar(&allSheets, "all-sheets", false, "Export all worksheets in tab order to one CSV")
	flags.BoolVar(&skipRepeatedHeader, "skip-repeated-header", false, "Skip matching first rows on subsequent sheets (requires all-sheets)")
	return command
}

// Execute runs the command and returns errors to the caller.
func Execute() error { return NewRootCommand().Execute() }
