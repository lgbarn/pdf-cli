package commands

import (
	"fmt"

	"github.com/lgbarn/pdf-cli/internal/cli"
	"github.com/lgbarn/pdf-cli/internal/pdf"
	"github.com/lgbarn/pdf-cli/internal/util"
	"github.com/spf13/cobra"
)

func init() {
	cli.AddCommand(mergeCmd)
	cli.AddOutputFlag(mergeCmd, "Output file path (required)")
	cli.AddPasswordFlag(mergeCmd, "Password for encrypted input PDFs")
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
	password := cli.GetPassword(cmd)

	if err := util.ValidatePDFFiles(args); err != nil {
		return err
	}

	if err := checkOutputFile(output); err != nil {
		return err
	}

	cli.PrintVerbose("Merging %d files into %s", len(args), output)

	if err := pdf.MergeWithProgress(args, output, password, cli.Progress()); err != nil {
		return util.WrapError("merging files", output, err)
	}

	fmt.Printf("Merged %d files into %s\n", len(args), output)
	return nil
}
