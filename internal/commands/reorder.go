package commands

import (
	"fmt"

	"github.com/lgbarn/pdf-cli/internal/cli"
	"github.com/lgbarn/pdf-cli/internal/commands/patterns"
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
	cli.AddPasswordFileFlag(reorderCmd, "")
	cli.AddAllowInsecurePasswordFlag(reorderCmd)
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
	// Sanitize input path
	sanitizedPath, err := fileio.SanitizePath(args[0])
	if err != nil {
		return fmt.Errorf("invalid file path: %w", err)
	}
	inputArg := sanitizedPath

	explicitOutput := cli.GetOutput(cmd)
	// Sanitize output path
	if explicitOutput != "" && explicitOutput != "-" {
		explicitOutput, err = fileio.SanitizePath(explicitOutput)
		if err != nil {
			return fmt.Errorf("invalid output path: %w", err)
		}
	}

	password, err := cli.GetPasswordSecure(cmd, "Enter PDF password: ")
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}
	toStdout := cli.GetStdout(cmd)
	sequence, _ := cmd.Flags().GetString("sequence")

	// Handle dry-run mode early
	if cli.IsDryRun() {
		return reorderDryRun(inputArg, explicitOutput, sequence, password)
	}

	handler := &patterns.StdioHandler{
		InputArg:       inputArg,
		ExplicitOutput: explicitOutput,
		ToStdout:       toStdout,
		DefaultSuffix:  SuffixReordered,
		Operation:      "reorder",
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

	pageCount, err := pdf.PageCount(input, password)
	if err != nil {
		return pdferrors.WrapError("reading file", inputArg, err)
	}

	pageList, err := pages.ParseReorderSequence(sequence, pageCount)
	if err != nil {
		return fmt.Errorf("invalid sequence: %w", err)
	}

	if !toStdout {
		if err := checkOutputFile(output); err != nil {
			return err
		}
	}

	cli.PrintVerbose("Reordering %d pages from %s -> %s", len(pageList), inputArg, output)
	cli.PrintVerbose("Page order: %v", pageList)

	if err := pdf.ExtractPages(input, output, pageList, password); err != nil {
		return pdferrors.WrapError("reordering pages", inputArg, err)
	}

	if err := handler.Finalize(); err != nil {
		return err
	}

	if !toStdout {
		fmt.Printf("Reordered PDF saved to %s (%d pages)\n", output, len(pageList))
	}
	return nil
}

func reorderDryRun(inputArg, explicitOutput, sequence, password string) error {
	if fileio.IsStdinInput(inputArg) {
		cli.DryRunPrint("Would reorder: stdin")
		cli.DryRunPrint("  Sequence: %s", sequence)
		return nil
	}

	info, err := pdf.GetInfo(inputArg, password)
	if err != nil {
		cli.DryRunPrint("Would reorder: %s (unable to read info)", inputArg)
		cli.DryRunPrint("  Sequence: %s", sequence)
		return nil
	}

	pageCount := info.Pages
	pageList, err := pages.ParseReorderSequence(sequence, pageCount)
	if err != nil {
		cli.DryRunPrint("Would reorder: %s (%d pages)", inputArg, pageCount)
		cli.DryRunPrint("  Sequence: %s (invalid: %v)", sequence, err)
		return nil
	}

	output := outputOrDefault(explicitOutput, inputArg, SuffixReordered)
	cli.DryRunPrint("Would reorder: %s (%d pages)", inputArg, pageCount)
	cli.DryRunPrint("  Sequence: %s", sequence)
	cli.DryRunPrint("  Result: %d pages in order: %v", len(pageList), pageList)
	cli.DryRunPrint("  Output: %s", output)
	return nil
}
