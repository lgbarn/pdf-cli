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
	cli.AddCommand(watermarkCmd)
	cli.AddOutputFlag(watermarkCmd, "Output file path (only with single file)")
	cli.AddPagesFlag(watermarkCmd, "Pages to watermark (default: all)")
	cli.AddPasswordFlag(watermarkCmd, "Password for encrypted PDFs")
	cli.AddPasswordFileFlag(watermarkCmd, "")
	watermarkCmd.Flags().StringP("text", "t", "", "Text watermark content")
	watermarkCmd.Flags().StringP("image", "i", "", "Image file for image watermark")
}

var watermarkCmd = &cobra.Command{
	Use:   "watermark <file.pdf> [file2.pdf...]",
	Short: "Add a watermark to PDF(s)",
	Long: `Add a text or image watermark to PDF file(s).

Text watermarks are rendered diagonally across each page.
Image watermarks are centered on each page.

Either --text or --image must be specified.

Supports batch processing of multiple files. When processing
multiple files, output files are named with '_watermarked' suffix.

Examples:
  pdf watermark document.pdf -t "CONFIDENTIAL" -o marked.pdf
  pdf watermark document.pdf -t "DRAFT" -p 1-5 -o draft.pdf
  pdf watermark document.pdf -i logo.png -o branded.pdf
  pdf watermark *.pdf -t "CONFIDENTIAL"       # Batch watermark
  pdf watermark doc1.pdf doc2.pdf -t "DRAFT"  # Multiple files`,
	Args: cobra.MinimumNArgs(1),
	RunE: runWatermark,
}

func runWatermark(cmd *cobra.Command, args []string) error {
	pagesStr := cli.GetPages(cmd)
	password, err := cli.GetPasswordSecure(cmd, "Enter PDF password: ")
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}
	output := cli.GetOutput(cmd)
	text, _ := cmd.Flags().GetString("text")
	image, _ := cmd.Flags().GetString("image")

	if text == "" && image == "" {
		return fmt.Errorf("must specify either --text or --image for watermark")
	}
	if text != "" && image != "" {
		return fmt.Errorf("cannot specify both --text and --image")
	}

	if image != "" && !fileio.FileExists(image) {
		return fmt.Errorf("image file not found: %s", image)
	}

	// Handle dry-run mode
	if cli.IsDryRun() {
		return watermarkDryRun(args, output, pagesStr, password, text, image)
	}

	if err := validateBatchOutput(args, output, "_watermarked"); err != nil {
		return err
	}

	return processBatch(args, func(inputFile string) error {
		return watermarkFile(inputFile, output, pagesStr, password, text, image)
	})
}

func watermarkDryRun(args []string, explicitOutput, pagesStr, password, text, image string) error {
	for _, inputFile := range args {
		info, err := pdf.GetInfo(inputFile, password)
		if err != nil {
			cli.DryRunPrint("Would watermark: %s (unable to read info)", inputFile)
			continue
		}

		output := outputOrDefault(explicitOutput, inputFile, "_watermarked")
		pageDesc := "all pages"
		if pagesStr != "" {
			pageDesc = "pages " + pagesStr
		}

		cli.DryRunPrint("Would watermark: %s (%d pages)", inputFile, info.Pages)
		if text != "" {
			cli.DryRunPrint("  Text: \"%s\"", text)
		} else {
			cli.DryRunPrint("  Image: %s", image)
		}
		cli.DryRunPrint("  Pages: %s", pageDesc)
		cli.DryRunPrint("  Output: %s", output)
	}
	return nil
}

func watermarkFile(inputFile, explicitOutput, pagesStr, password, text, image string) error {
	if err := fileio.ValidatePDFFile(inputFile); err != nil {
		return err
	}

	pages, err := parseAndValidatePages(pagesStr, inputFile, password)
	if err != nil {
		return err
	}

	output := outputOrDefault(explicitOutput, inputFile, "_watermarked")

	if err := checkOutputFile(output); err != nil {
		return err
	}

	if text != "" {
		cli.PrintVerbose("Adding text watermark '%s' to %s", text, inputFile)
		if err := pdf.AddWatermark(inputFile, output, text, pages, password); err != nil {
			return pdferrors.WrapError("adding watermark", inputFile, err)
		}
	} else {
		cli.PrintVerbose("Adding image watermark '%s' to %s", image, inputFile)
		if err := pdf.AddImageWatermark(inputFile, output, image, pages, password); err != nil {
			return pdferrors.WrapError("adding watermark", inputFile, err)
		}
	}

	fmt.Printf("Watermark added to %s\n", output)
	return nil
}
