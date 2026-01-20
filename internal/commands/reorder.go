package commands

import (
	"fmt"

	"github.com/lgbarn/pdf-cli/internal/cli"
	"github.com/lgbarn/pdf-cli/internal/pdf"
	"github.com/lgbarn/pdf-cli/internal/util"
	"github.com/spf13/cobra"
)

func init() {
	cli.AddCommand(reorderCmd)
	cli.AddOutputFlag(reorderCmd, "Output file path")
	cli.AddPasswordFlag(reorderCmd, "Password for encrypted PDFs")
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

Examples:
  pdf reorder doc.pdf -s "1,5,2,3,4" -o out.pdf   # Move page 5 to position 2
  pdf reorder doc.pdf -s "end-1" -o reversed.pdf  # Reverse all pages
  pdf reorder doc.pdf -s "1-end,1" -o dup.pdf     # Duplicate page 1 at the end
  pdf reorder doc.pdf -s "2-end" -o skip.pdf      # Remove first page`,
	Args: cobra.ExactArgs(1),
	RunE: runReorder,
}

func runReorder(cmd *cobra.Command, args []string) error {
	inputFile := args[0]
	output := cli.GetOutput(cmd)
	password := cli.GetPassword(cmd)
	sequence, _ := cmd.Flags().GetString("sequence")

	if err := util.ValidatePDFFile(inputFile); err != nil {
		return err
	}

	pageCount, err := pdf.PageCount(inputFile, password)
	if err != nil {
		return util.WrapError("reading file", inputFile, err)
	}

	pages, err := util.ParseReorderSequence(sequence, pageCount)
	if err != nil {
		return fmt.Errorf("invalid sequence: %w", err)
	}

	output = outputOrDefault(output, inputFile, "_reordered")

	if err := checkOutputFile(output); err != nil {
		return err
	}

	cli.PrintVerbose("Reordering %d pages from %s -> %s", len(pages), inputFile, output)
	cli.PrintVerbose("Page order: %v", pages)

	if err := pdf.ExtractPages(inputFile, output, pages, password); err != nil {
		return util.WrapError("reordering pages", inputFile, err)
	}

	fmt.Printf("Reordered PDF saved to %s (%d pages)\n", output, len(pages))
	return nil
}
