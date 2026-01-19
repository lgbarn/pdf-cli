package commands

import (
	"fmt"

	"github.com/lgbarn/pdf-cli/internal/cli"
	"github.com/lgbarn/pdf-cli/internal/pdf"
	"github.com/lgbarn/pdf-cli/internal/util"
	"github.com/spf13/cobra"
)

func init() {
	cli.AddCommand(rotateCmd)
	cli.AddOutputFlag(rotateCmd, "Output file path (only with single file)")
	cli.AddPagesFlag(rotateCmd, "Pages to rotate (default: all pages)")
	cli.AddPasswordFlag(rotateCmd, "Password for encrypted PDFs")
	rotateCmd.Flags().IntP("angle", "a", 90, "Rotation angle (90, 180, or 270)")
}

var rotateCmd = &cobra.Command{
	Use:   "rotate <file.pdf> [file2.pdf...]",
	Short: "Rotate pages in PDF(s)",
	Long: `Rotate pages in PDF file(s) by a specified angle.

Valid rotation angles are 90, 180, and 270 degrees (clockwise).
By default, all pages are rotated. Use -p to specify specific pages.

Supports batch processing of multiple files. When processing
multiple files, output files are named with '_rotated' suffix.

Examples:
  pdf rotate document.pdf -a 90 -o rotated.pdf
  pdf rotate document.pdf -a 180 -p 1-5 -o rotated.pdf
  pdf rotate scanned.pdf -a 270
  pdf rotate *.pdf -a 90                    # Batch rotate
  pdf rotate doc1.pdf doc2.pdf -a 180       # Multiple files`,
	Args: cobra.MinimumNArgs(1),
	RunE: runRotate,
}

func runRotate(cmd *cobra.Command, args []string) error {
	pagesStr := cli.GetPages(cmd)
	password := cli.GetPassword(cmd)
	output := cli.GetOutput(cmd)
	angle, _ := cmd.Flags().GetInt("angle")

	if angle != 90 && angle != 180 && angle != 270 {
		return fmt.Errorf("invalid rotation angle: %d (must be 90, 180, or 270)", angle)
	}

	if err := validateBatchOutput(args, output, "_rotated"); err != nil {
		return err
	}

	return processBatch(args, func(inputFile string) error {
		return rotateFile(inputFile, output, pagesStr, password, angle)
	})
}

func rotateFile(inputFile, explicitOutput, pagesStr, password string, angle int) error {
	if err := util.ValidatePDFFile(inputFile); err != nil {
		return err
	}

	pages, err := parseAndValidatePages(pagesStr, inputFile, password)
	if err != nil {
		return err
	}

	output := outputOrDefault(explicitOutput, inputFile, "_rotated")

	if err := checkOutputFile(output); err != nil {
		return err
	}

	pageDesc := "all pages"
	if len(pages) > 0 {
		pageDesc = fmt.Sprintf("%d pages", len(pages))
	}
	cli.PrintVerbose("Rotating %s by %d degrees in %s", pageDesc, angle, inputFile)

	if err := pdf.Rotate(inputFile, output, angle, pages, password); err != nil {
		return util.WrapError("rotating pages", inputFile, err)
	}

	fmt.Printf("Rotated %s by %d degrees to %s\n", pageDesc, angle, output)
	return nil
}
