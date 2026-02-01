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
	cli.AddCommand(combineImagesCmd)
	cli.AddOutputFlag(combineImagesCmd, "Output PDF file (required)")
	_ = combineImagesCmd.MarkFlagRequired("output")
	combineImagesCmd.Flags().String("page-size", "", "Page size (A4, Letter, or leave empty to use image dimensions)")
}

var combineImagesCmd = &cobra.Command{
	Use:   "combine-images <image1> <image2> [image3...]",
	Short: "Create a PDF from multiple images",
	Long: `Create a PDF file from multiple images.

Each image becomes one page in the output PDF.
Supported formats: PNG, JPEG, TIFF

Examples:
  pdf combine-images photo1.jpg photo2.jpg -o album.pdf
  pdf combine-images *.png -o scans.pdf
  pdf combine-images scan1.png scan2.png -o doc.pdf --page-size A4`,
	Args: cobra.MinimumNArgs(1),
	RunE: runCombineImages,
}

func runCombineImages(cmd *cobra.Command, args []string) error {
	// Sanitize input paths
	sanitizedArgs, err := fileio.SanitizePaths(args)
	if err != nil {
		return fmt.Errorf("invalid file path: %w", err)
	}
	args = sanitizedArgs

	output := cli.GetOutput(cmd)
	// Sanitize output path
	output, err = fileio.SanitizePath(output)
	if err != nil {
		return fmt.Errorf("invalid output path: %w", err)
	}

	pageSize, _ := cmd.Flags().GetString("page-size")

	// Validate all input files exist and are images
	for _, img := range args {
		if !fileio.FileExists(img) {
			return fmt.Errorf("image file not found: %s", img)
		}
		if !fileio.IsImageFile(img) {
			return fmt.Errorf("not a supported image format: %s (supported: png, jpg, jpeg, tif, tiff)", img)
		}
	}

	if err := checkOutputFile(output); err != nil {
		return err
	}

	cli.PrintVerbose("Creating PDF from %d images...", len(args))

	if err := pdf.CreatePDFFromImages(args, output, pageSize); err != nil {
		return pdferrors.WrapError("creating PDF from images", output, err)
	}

	fmt.Printf("Created %s from %d image(s)\n", output, len(args))
	return nil
}
