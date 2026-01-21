package commands

import (
	"fmt"
	"os"

	"github.com/lgbarn/pdf-cli/internal/cli"
	"github.com/lgbarn/pdf-cli/internal/fileio"
	"github.com/lgbarn/pdf-cli/internal/pdf"
	"github.com/lgbarn/pdf-cli/internal/pdferrors"
	"github.com/spf13/cobra"
)

func init() {
	cli.AddCommand(extractCmd)
	cli.AddOutputFlag(extractCmd, "Output file path")
	cli.AddPagesFlag(extractCmd, "Pages to extract (e.g., 1-5,7,10-12)")
	cli.AddPasswordFlag(extractCmd, "Password for encrypted PDFs")
	cli.AddStdoutFlag(extractCmd)
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

Use "-" to read from stdin. Use --stdout for binary output.

Examples:
  pdf extract document.pdf -p 1-5 -o first5.pdf
  pdf extract document.pdf -p 1,3,5,7 -o odds.pdf
  cat input.pdf | pdf extract - -p 1-5 --stdout > pages.pdf`,
	Args: cobra.ExactArgs(1),
	RunE: runExtract,
}

func runExtract(cmd *cobra.Command, args []string) error {
	inputArg := args[0]
	pagesStr := cli.GetPages(cmd)
	password := cli.GetPassword(cmd)
	toStdout := cli.GetStdout(cmd)

	// Handle stdin input
	inputFile, cleanup, err := fileio.ResolveInputPath(inputArg)
	if err != nil {
		return err
	}
	defer cleanup()

	if !fileio.IsStdinInput(inputArg) {
		if err := fileio.ValidatePDFFile(inputFile); err != nil {
			return err
		}
	}

	pages, err := parseAndValidatePages(pagesStr, inputFile, password)
	if err != nil {
		return err
	}

	if len(pages) == 0 {
		return fmt.Errorf("no pages specified")
	}

	// Handle stdout output
	var output string
	var outputCleanup func()
	if toStdout {
		tmpFile, err := os.CreateTemp("", "pdf-cli-extract-*.pdf")
		if err != nil {
			return fmt.Errorf("failed to create temp file: %w", err)
		}
		output = tmpFile.Name()
		_ = tmpFile.Close()
		outputCleanup = func() { _ = os.Remove(output) }
		defer outputCleanup()
	} else {
		output = outputOrDefault(cli.GetOutput(cmd), inputArg, "_extracted")
		if err := checkOutputFile(output); err != nil {
			return err
		}
	}

	cli.PrintVerbose("Extracting pages %s from %s to %s", pagesStr, inputArg, output)

	if err := pdf.ExtractPages(inputFile, output, pages, password); err != nil {
		return pdferrors.WrapError("extracting pages", inputArg, err)
	}

	if toStdout {
		return fileio.WriteToStdout(output)
	}

	fmt.Printf("Extracted %d pages to %s\n", len(pages), output)
	return nil
}
