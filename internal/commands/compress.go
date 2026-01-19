package commands

import (
	"fmt"

	"github.com/lgbarn/pdf-cli/internal/cli"
	"github.com/lgbarn/pdf-cli/internal/pdf"
	"github.com/lgbarn/pdf-cli/internal/util"
	"github.com/spf13/cobra"
)

func init() {
	cli.AddCommand(compressCmd)
	cli.AddOutputFlag(compressCmd, "Output file path (only with single file)")
	cli.AddPasswordFlag(compressCmd, "Password for encrypted PDFs")
}

var compressCmd = &cobra.Command{
	Use:   "compress <file.pdf> [file2.pdf...]",
	Short: "Compress and optimize PDF(s)",
	Long: `Compress and optimize PDF file(s) to reduce their size.

This removes redundant data, optimizes internal structures,
and can significantly reduce file size without losing quality.

Supports batch processing of multiple files. When processing
multiple files, output files are named with '_compressed' suffix.

Examples:
  pdf compress large.pdf -o smaller.pdf
  pdf compress document.pdf
  pdf compress *.pdf                      # Batch compress
  pdf compress doc1.pdf doc2.pdf doc3.pdf # Multiple files`,
	Args: cobra.MinimumNArgs(1),
	RunE: runCompress,
}

func runCompress(cmd *cobra.Command, args []string) error {
	password := cli.GetPassword(cmd)
	output := cli.GetOutput(cmd)

	if err := validateBatchOutput(args, output, "_compressed"); err != nil {
		return err
	}

	return processBatch(args, func(inputFile string) error {
		return compressFile(inputFile, output, password)
	})
}

func compressFile(inputFile, explicitOutput, password string) error {
	if err := util.ValidatePDFFile(inputFile); err != nil {
		return err
	}

	originalSize, _ := util.GetFileSize(inputFile)
	output := outputOrDefault(explicitOutput, inputFile, "_compressed")

	if err := checkOutputFile(output); err != nil {
		return err
	}

	cli.PrintVerbose("Compressing %s to %s", inputFile, output)

	if err := pdf.Compress(inputFile, output, password); err != nil {
		return util.WrapError("compressing file", inputFile, err)
	}

	newSize, _ := util.GetFileSize(output)
	savings := originalSize - newSize
	savingsPercent := float64(savings) / float64(originalSize) * 100

	fmt.Printf("Compressed %s to %s\n", inputFile, output)
	fmt.Printf("Original:   %s\n", util.FormatFileSize(originalSize))
	fmt.Printf("Compressed: %s\n", util.FormatFileSize(newSize))
	if savings > 0 {
		fmt.Printf("Saved:      %s (%.1f%%)\n", util.FormatFileSize(savings), savingsPercent)
	} else {
		fmt.Println("Note: File size increased (already optimized)")
	}

	return nil
}
