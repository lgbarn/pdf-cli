package commands

import (
	"fmt"
	"path/filepath"

	"github.com/lgbarn/pdf-cli/internal/cli"
	"github.com/lgbarn/pdf-cli/internal/pdf"
	"github.com/lgbarn/pdf-cli/internal/util"
	"github.com/spf13/cobra"
)

func init() {
	cli.AddCommand(splitCmd)
	cli.AddOutputFlag(splitCmd, "Output directory for split files")
	cli.AddPasswordFlag(splitCmd, "Password for encrypted PDFs")
	splitCmd.Flags().IntP("pages", "n", 1, "Number of pages per output file")
}

var splitCmd = &cobra.Command{
	Use:   "split <file.pdf>",
	Short: "Split PDF into multiple files",
	Long: `Split a PDF file into multiple smaller PDF files.

By default, splits into individual pages. Use -n to specify
how many pages per output file.

Output files are named based on the input file with page numbers appended.

Examples:
  pdf split document.pdf -o output/
  pdf split document.pdf -n 5 -o chunks/
  pdf split large.pdf`,
	Args: cobra.ExactArgs(1),
	RunE: runSplit,
}

func runSplit(cmd *cobra.Command, args []string) error {
	inputFile := args[0]
	outputDir := cli.GetOutput(cmd)
	password := cli.GetPassword(cmd)
	pagesPerFile, _ := cmd.Flags().GetInt("pages")

	// Validate input file
	if err := util.ValidatePDFFile(inputFile); err != nil {
		return err
	}

	// Default output directory
	if outputDir == "" {
		outputDir = filepath.Dir(inputFile)
	}

	// Ensure output directory exists
	if err := util.EnsureDir(outputDir); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	cli.PrintVerbose("Splitting %s into %s (%d pages per file)", inputFile, outputDir, pagesPerFile)

	var err error
	if pagesPerFile == 1 {
		err = pdf.Split(inputFile, outputDir, password)
	} else {
		err = pdf.SplitByPageCount(inputFile, outputDir, pagesPerFile, password)
	}

	if err != nil {
		return util.WrapError("splitting file", inputFile, err)
	}

	fmt.Printf("Split %s into %s\n", inputFile, outputDir)
	return nil
}
