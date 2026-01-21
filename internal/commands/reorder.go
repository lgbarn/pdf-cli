package commands

import (
	"fmt"
	"os"

	"github.com/lgbarn/pdf-cli/internal/cli"
	"github.com/lgbarn/pdf-cli/internal/fileio"
	"github.com/lgbarn/pdf-cli/internal/pages"
	"github.com/lgbarn/pdf-cli/internal/pdf"
	"github.com/lgbarn/pdf-cli/internal/pdferrors"
	"github.com/spf13/cobra"
)

func init() {
	cli.AddCommand(reorderCmd)
	cli.AddOutputFlag(reorderCmd, "Output file path")
	cli.AddPasswordFlag(reorderCmd, "Password for encrypted PDFs")
	cli.AddStdoutFlag(reorderCmd)
	reorderCmd.Flags().StringP("sequence", "s", "", "Page sequence (required)")
	_ = reorderCmd.MarkFlagRequired("sequence")
}

var reorderCmd = &cobra.Command{
	Use:   "reorder <file.pdf>",
	Short: "Reorder pages in a PDF",
	Long: `Reorder pages in a PDF file.

Use -s to specify the new page order. Supports:
  - Individual pages: 1,3,5
  - Ranges: 1-10
  - Special values: end (last page)
  - Reverse ranges: 10-1, end-1
  - Page duplication: repeat a page number to include it multiple times

Use "-" to read from stdin. Use --stdout for binary output.

Examples:
  pdf reorder doc.pdf -s "1,5,2,3,4" -o out.pdf   # Move page 5 to position 2
  pdf reorder doc.pdf -s "end-1" -o reversed.pdf  # Reverse all pages
  cat in.pdf | pdf reorder - -s "end-1" --stdout > reversed.pdf`,
	Args: cobra.ExactArgs(1),
	RunE: runReorder,
}

func runReorder(cmd *cobra.Command, args []string) error {
	inputArg := args[0]
	output := cli.GetOutput(cmd)
	password := cli.GetPassword(cmd)
	toStdout := cli.GetStdout(cmd)
	sequence, _ := cmd.Flags().GetString("sequence")

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

	pageCount, err := pdf.PageCount(inputFile, password)
	if err != nil {
		return pdferrors.WrapError("reading file", inputArg, err)
	}

	pages, err := pages.ParseReorderSequence(sequence, pageCount)
	if err != nil {
		return fmt.Errorf("invalid sequence: %w", err)
	}

	// Handle stdout output
	var actualOutput string
	var outputCleanup func()
	if toStdout {
		tmpFile, err := os.CreateTemp("", "pdf-cli-reorder-*.pdf")
		if err != nil {
			return fmt.Errorf("failed to create temp file: %w", err)
		}
		actualOutput = tmpFile.Name()
		_ = tmpFile.Close()
		outputCleanup = func() { _ = os.Remove(actualOutput) }
		defer outputCleanup()
	} else {
		actualOutput = outputOrDefault(output, inputArg, "_reordered")
		if err := checkOutputFile(actualOutput); err != nil {
			return err
		}
	}

	cli.PrintVerbose("Reordering %d pages from %s -> %s", len(pages), inputArg, actualOutput)
	cli.PrintVerbose("Page order: %v", pages)

	if err := pdf.ExtractPages(inputFile, actualOutput, pages, password); err != nil {
		return pdferrors.WrapError("reordering pages", inputArg, err)
	}

	if toStdout {
		return fileio.WriteToStdout(actualOutput)
	}

	fmt.Printf("Reordered PDF saved to %s (%d pages)\n", actualOutput, len(pages))
	return nil
}
