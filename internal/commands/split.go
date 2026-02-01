package commands

import (
	"fmt"
	"path/filepath"

	"github.com/lgbarn/pdf-cli/internal/cli"
	"github.com/lgbarn/pdf-cli/internal/fileio"
	"github.com/lgbarn/pdf-cli/internal/pdf"
	"github.com/lgbarn/pdf-cli/internal/pdferrors"
	"github.com/spf13/cobra"
)

func init() {
	cli.AddCommand(splitCmd)
	cli.AddOutputFlag(splitCmd, "Output directory for split files")
	cli.AddPasswordFlag(splitCmd, "Password for encrypted PDFs")
	cli.AddPasswordFileFlag(splitCmd, "")
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
	// Sanitize input path
	sanitizedPath, err := fileio.SanitizePath(args[0])
	if err != nil {
		return fmt.Errorf("invalid file path: %w", err)
	}
	inputFile := sanitizedPath

	outputDir := cli.GetOutput(cmd)
	password, err := cli.GetPasswordSecure(cmd, "Enter PDF password: ")
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}
	pagesPerFile, _ := cmd.Flags().GetInt("pages")

	// Validate input file
	if err := fileio.ValidatePDFFile(inputFile); err != nil {
		return err
	}

	// Default output directory
	if outputDir == "" {
		outputDir = filepath.Dir(inputFile)
	}

	// Sanitize output directory path
	outputDir, err = fileio.SanitizePath(outputDir)
	if err != nil {
		return fmt.Errorf("invalid output path: %w", err)
	}

	// Handle dry-run mode
	if cli.IsDryRun() {
		info, err := pdf.GetInfo(inputFile, password)
		if err != nil {
			cli.DryRunPrint("Would split: %s (unable to read info)", inputFile)
		} else {
			outputFiles := (info.Pages + pagesPerFile - 1) / pagesPerFile
			cli.DryRunPrint("Would split: %s (%d pages)", inputFile, info.Pages)
			cli.DryRunPrint("Pages per file: %d", pagesPerFile)
			cli.DryRunPrint("Output directory: %s", outputDir)
			cli.DryRunPrint("Result: ~%d output files", outputFiles)
		}
		return nil
	}

	// Ensure output directory exists
	if err := fileio.EnsureDir(outputDir); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	cli.PrintVerbose("Splitting %s into %s (%d pages per file)", inputFile, outputDir, pagesPerFile)

	if err := pdf.SplitWithProgress(inputFile, outputDir, pagesPerFile, password, cli.Progress()); err != nil {
		return pdferrors.WrapError("splitting file", inputFile, err)
	}

	fmt.Printf("Split %s into %s\n", inputFile, outputDir)
	return nil
}
