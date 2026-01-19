package commands

import (
	"fmt"

	"github.com/lgbarn/pdf-cli/internal/cli"
	"github.com/lgbarn/pdf-cli/internal/pdf"
	"github.com/lgbarn/pdf-cli/internal/util"
	"github.com/spf13/cobra"
)

func init() {
	cli.AddCommand(imagesCmd)
	cli.AddOutputFlag(imagesCmd, "Output directory for extracted images")
	cli.AddPagesFlag(imagesCmd, "Pages to extract images from (default: all)")
	cli.AddPasswordFlag(imagesCmd, "Password for encrypted PDFs")
}

var imagesCmd = &cobra.Command{
	Use:   "images <file.pdf>",
	Short: "Extract images from a PDF",
	Long: `Extract all images from a PDF file.

Images are saved to the specified output directory.
Original image format and quality are preserved where possible.

Examples:
  pdf images document.pdf -o images/
  pdf images document.pdf -p 1-5 -o chapter1_images/
  pdf images presentation.pdf -o slides/`,
	Args: cobra.ExactArgs(1),
	RunE: runImages,
}

func runImages(cmd *cobra.Command, args []string) error {
	inputFile := args[0]
	pagesStr := cli.GetPages(cmd)
	password := cli.GetPassword(cmd)

	if err := util.ValidatePDFFile(inputFile); err != nil {
		return err
	}

	outputDir := cli.GetOutput(cmd)
	if outputDir == "" {
		outputDir = "."
	}

	if err := util.EnsureDir(outputDir); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	pages, err := parseAndValidatePages(pagesStr, inputFile, password)
	if err != nil {
		return err
	}

	cli.PrintVerbose("Extracting images from %s to %s", inputFile, outputDir)

	if err := pdf.ExtractImages(inputFile, outputDir, pages, password); err != nil {
		return util.WrapError("extracting images", inputFile, err)
	}

	fmt.Printf("Images extracted to %s\n", outputDir)
	return nil
}
