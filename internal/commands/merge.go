package commands

import (
	"fmt"

	"github.com/lgbarn/pdf-cli/internal/cli"
	"github.com/lgbarn/pdf-cli/internal/fileio"
	"github.com/lgbarn/pdf-cli/internal/pdf"
	"github.com/lgbarn/pdf-cli/internal/pdferrors"
	"github.com/spf13/cobra"
)

func init() {
	cli.AddCommand(mergeCmd)
	cli.AddOutputFlag(mergeCmd, "Output file path (required)")
	cli.AddPasswordFlag(mergeCmd, "Password for encrypted input PDFs")
	cli.AddPasswordFileFlag(mergeCmd, "")
	_ = mergeCmd.MarkFlagRequired("output")
}

var mergeCmd = &cobra.Command{
	Use:   "merge <file1.pdf> <file2.pdf> [file3.pdf...]",
	Short: "Merge multiple PDFs into one",
	Long: `Merge multiple PDF files into a single PDF.

Files are merged in the order they are specified.
The output file must be specified with the -o flag.

Examples:
  pdf merge -o combined.pdf file1.pdf file2.pdf
  pdf merge -o output.pdf *.pdf
  pdf merge -o combined.pdf doc1.pdf doc2.pdf doc3.pdf`,
	Args: cobra.MinimumNArgs(2),
	RunE: runMerge,
}

func runMerge(cmd *cobra.Command, args []string) error {
	output := cli.GetOutput(cmd)
	password, err := cli.GetPasswordSecure(cmd, "Enter PDF password: ")
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}

	if err := fileio.ValidatePDFFiles(args); err != nil {
		return err
	}

	// Handle dry-run mode
	if cli.IsDryRun() {
		totalPages := 0
		cli.DryRunPrint("Would merge %d files:", len(args))
		for _, f := range args {
			info, err := pdf.GetInfo(f, password)
			if err == nil {
				cli.DryRunPrint("  - %s (%d pages)", f, info.Pages)
				totalPages += info.Pages
			} else {
				cli.DryRunPrint("  - %s (unable to read info)", f)
			}
		}
		cli.DryRunPrint("Output: %s (%d pages total)", output, totalPages)
		return nil
	}

	if err := checkOutputFile(output); err != nil {
		return err
	}

	cli.PrintVerbose("Merging %d files into %s", len(args), output)

	if err := pdf.MergeWithProgress(args, output, password, cli.Progress()); err != nil {
		return pdferrors.WrapError("merging files", output, err)
	}

	fmt.Printf("Merged %d files into %s\n", len(args), output)
	return nil
}
