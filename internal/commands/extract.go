package commands

import (
	"fmt"

	"github.com/lgbarn/pdf-cli/internal/cli"
	"github.com/lgbarn/pdf-cli/internal/pdf"
	"github.com/lgbarn/pdf-cli/internal/util"
	"github.com/spf13/cobra"
)

func init() {
	cli.AddCommand(extractCmd)
	cli.AddOutputFlag(extractCmd, "Output file path")
	cli.AddPagesFlag(extractCmd, "Pages to extract (e.g., 1-5,7,10-12)")
	cli.AddPasswordFlag(extractCmd, "Password for encrypted PDFs")
	_ = extractCmd.MarkFlagRequired("pages")
}

var extractCmd = &cobra.Command{
	Use:   "extract <file.pdf>",
	Short: "Extract specific pages from a PDF",
	Long: `Extract specific pages from a PDF into a new file.

Specify pages using ranges and individual numbers:
  - Single pages: 1,3,5
  - Ranges: 1-5,10-15
  - Combined: 1-3,7,10-12

Examples:
  pdf extract document.pdf -p 1-5 -o first5.pdf
  pdf extract document.pdf -p 1,3,5,7 -o odds.pdf
  pdf extract document.pdf -p 10-20 -o chapter2.pdf`,
	Args: cobra.ExactArgs(1),
	RunE: runExtract,
}

func runExtract(cmd *cobra.Command, args []string) error {
	inputFile := args[0]
	pagesStr := cli.GetPages(cmd)
	password := cli.GetPassword(cmd)

	if err := util.ValidatePDFFile(inputFile); err != nil {
		return err
	}

	pages, err := parseAndValidatePages(pagesStr, inputFile, password)
	if err != nil {
		return err
	}

	if len(pages) == 0 {
		return fmt.Errorf("no pages specified")
	}

	output := outputOrDefault(cli.GetOutput(cmd), inputFile, "_extracted")

	if err := checkOutputFile(output); err != nil {
		return err
	}

	cli.PrintVerbose("Extracting pages %s from %s to %s", pagesStr, inputFile, output)

	if err := pdf.ExtractPages(inputFile, output, pages, password); err != nil {
		return util.WrapError("extracting pages", inputFile, err)
	}

	fmt.Printf("Extracted %d pages to %s\n", len(pages), output)
	return nil
}
