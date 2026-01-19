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
	cli.AddOutputFlag(compressCmd, "Output file path")
	cli.AddPasswordFlag(compressCmd, "Password for encrypted PDFs")
}

var compressCmd = &cobra.Command{
	Use:   "compress <file.pdf>",
	Short: "Compress and optimize a PDF",
	Long: `Compress and optimize a PDF file to reduce its size.

This removes redundant data, optimizes internal structures,
and can significantly reduce file size without losing quality.

Examples:
  pdf compress large.pdf -o smaller.pdf
  pdf compress document.pdf
  pdf compress scan.pdf -o optimized.pdf`,
	Args: cobra.ExactArgs(1),
	RunE: runCompress,
}

func runCompress(cmd *cobra.Command, args []string) error {
	inputFile := args[0]
	password := cli.GetPassword(cmd)

	if err := util.ValidatePDFFile(inputFile); err != nil {
		return err
	}

	originalSize, _ := util.GetFileSize(inputFile)
	output := outputOrDefault(cli.GetOutput(cmd), inputFile, "_compressed")

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
