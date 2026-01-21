package commands

import (
	"fmt"
	"os"

	"github.com/lgbarn/pdf-cli/internal/cli"
	"github.com/lgbarn/pdf-cli/internal/commands/patterns"
	"github.com/lgbarn/pdf-cli/internal/fileio"
	"github.com/lgbarn/pdf-cli/internal/pdf"
	"github.com/lgbarn/pdf-cli/internal/pdferrors"
	"github.com/spf13/cobra"
)

func init() {
	cli.AddCommand(rotateCmd)
	cli.AddOutputFlag(rotateCmd, "Output file path (only with single file)")
	cli.AddPagesFlag(rotateCmd, "Pages to rotate (default: all pages)")
	cli.AddPasswordFlag(rotateCmd, "Password for encrypted PDFs")
	cli.AddStdoutFlag(rotateCmd)
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
Use "-" to read from stdin. Use --stdout for binary output.

Examples:
  pdf rotate document.pdf -a 90 -o rotated.pdf
  pdf rotate document.pdf -a 180 -p 1-5 -o rotated.pdf
  cat input.pdf | pdf rotate - -a 90 --stdout > rotated.pdf`,
	Args: cobra.MinimumNArgs(1),
	RunE: runRotate,
}

func runRotate(cmd *cobra.Command, args []string) error {
	pagesStr := cli.GetPages(cmd)
	password := cli.GetPassword(cmd)
	output := cli.GetOutput(cmd)
	toStdout := cli.GetStdout(cmd)
	angle, _ := cmd.Flags().GetInt("angle")

	if angle != 90 && angle != 180 && angle != 270 {
		return fmt.Errorf("invalid rotation angle: %d (must be 90, 180, or 270)", angle)
	}

	// Handle dry-run mode
	if cli.IsDryRun() {
		return rotateDryRun(args, output, pagesStr, password, angle)
	}

	// Handle stdin/stdout for single file
	if len(args) == 1 && (fileio.IsStdinInput(args[0]) || toStdout) {
		return rotateWithStdio(args[0], output, pagesStr, password, angle, toStdout)
	}

	if err := validateBatchOutput(args, output, "_rotated"); err != nil {
		return err
	}

	return processBatch(args, func(inputFile string) error {
		return rotateFile(inputFile, output, pagesStr, password, angle)
	})
}

func rotateDryRun(args []string, explicitOutput, pagesStr, password string, angle int) error {
	for _, inputFile := range args {
		if fileio.IsStdinInput(inputFile) {
			cli.DryRunPrint("Would rotate: stdin by %d degrees", angle)
			continue
		}

		info, err := pdf.GetInfo(inputFile, password)
		if err != nil {
			cli.DryRunPrint("Would rotate: %s (unable to read info)", inputFile)
			continue
		}

		output := outputOrDefault(explicitOutput, inputFile, "_rotated")
		pageDesc := "all pages"
		if pagesStr != "" {
			pageDesc = "pages " + pagesStr
		}

		cli.DryRunPrint("Would rotate: %s (%d pages)", inputFile, info.Pages)
		cli.DryRunPrint("  Angle: %d degrees", angle)
		cli.DryRunPrint("  Pages: %s", pageDesc)
		cli.DryRunPrint("  Output: %s", output)
	}
	return nil
}

func rotateWithStdio(inputArg, explicitOutput, pagesStr, password string, angle int, toStdout bool) error {
	handler := &patterns.StdioHandler{
		InputArg:       inputArg,
		ExplicitOutput: explicitOutput,
		ToStdout:       toStdout,
		DefaultSuffix:  "_rotated",
		Operation:      "rotate",
	}
	defer handler.Cleanup()

	input, output, err := handler.Setup()
	if err != nil {
		return err
	}

	pages, err := parseAndValidatePages(pagesStr, input, password)
	if err != nil {
		return err
	}

	if !toStdout {
		if err := checkOutputFile(output); err != nil {
			return err
		}
	}

	if err := pdf.Rotate(input, output, angle, pages, password); err != nil {
		return pdferrors.WrapError("rotating pages", inputArg, err)
	}

	if err := handler.Finalize(); err != nil {
		return err
	}

	if !toStdout {
		fmt.Fprintf(os.Stderr, "Rotated by %d degrees to %s\n", angle, output)
	}
	return nil
}

func rotateFile(inputFile, explicitOutput, pagesStr, password string, angle int) error {
	if err := fileio.ValidatePDFFile(inputFile); err != nil {
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
		return pdferrors.WrapError("rotating pages", inputFile, err)
	}

	fmt.Printf("Rotated %s by %d degrees to %s\n", pageDesc, angle, output)
	return nil
}
