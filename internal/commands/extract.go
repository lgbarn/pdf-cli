package commands

import (
	"fmt"

	"github.com/lgbarn/pdf-cli/internal/cli"
	"github.com/lgbarn/pdf-cli/internal/commands/patterns"
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
	explicitOutput := cli.GetOutput(cmd)

	// Handle dry-run mode early
	if cli.IsDryRun() {
		return extractDryRun(inputArg, explicitOutput, pagesStr, password)
	}

	handler := &patterns.StdioHandler{
		InputArg:       inputArg,
		ExplicitOutput: explicitOutput,
		ToStdout:       toStdout,
		DefaultSuffix:  "_extracted",
		Operation:      "extract",
	}
	defer handler.Cleanup()

	input, output, err := handler.Setup()
	if err != nil {
		return err
	}

	if !fileio.IsStdinInput(inputArg) {
		if err := fileio.ValidatePDFFile(input); err != nil {
			return err
		}
	}

	pages, err := parseAndValidatePages(pagesStr, input, password)
	if err != nil {
		return err
	}

	if len(pages) == 0 {
		return fmt.Errorf("no pages specified")
	}

	if !toStdout {
		if err := checkOutputFile(output); err != nil {
			return err
		}
	}

	cli.PrintVerbose("Extracting pages %s from %s to %s", pagesStr, inputArg, output)

	if err := pdf.ExtractPages(input, output, pages, password); err != nil {
		return pdferrors.WrapError("extracting pages", inputArg, err)
	}

	if err := handler.Finalize(); err != nil {
		return err
	}

	if !toStdout {
		fmt.Printf("Extracted %d pages to %s\n", len(pages), output)
	}
	return nil
}

func extractDryRun(inputArg, explicitOutput, pagesStr, password string) error {
	if fileio.IsStdinInput(inputArg) {
		cli.DryRunPrint("Would extract pages %s from: stdin", pagesStr)
		return nil
	}

	info, err := pdf.GetInfo(inputArg, password)
	if err != nil {
		cli.DryRunPrint("Would extract pages %s from: %s (unable to read info)", pagesStr, inputArg)
		return nil
	}

	output := outputOrDefault(explicitOutput, inputArg, "_extracted")
	cli.DryRunPrint("Would extract from: %s (%d pages total)", inputArg, info.Pages)
	cli.DryRunPrint("  Pages: %s", pagesStr)
	cli.DryRunPrint("  Output: %s", output)
	return nil
}
