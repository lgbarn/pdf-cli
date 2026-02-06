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
	cli.AddCommand(compressCmd)
	cli.AddOutputFlag(compressCmd, "Output file path (only with single file)")
	cli.AddPasswordFlag(compressCmd, "Password for encrypted PDFs")
	cli.AddPasswordFileFlag(compressCmd, "")
	cli.AddAllowInsecurePasswordFlag(compressCmd)
	cli.AddStdoutFlag(compressCmd)
}

var compressCmd = &cobra.Command{
	Use:   "compress <file.pdf> [file2.pdf...]",
	Short: "Compress and optimize PDF(s)",
	Long: `Compress and optimize PDF file(s) to reduce their size.

This removes redundant data, optimizes internal structures,
and can significantly reduce file size without losing quality.

Supports batch processing of multiple files. When processing
multiple files, output files are named with '_compressed' suffix.
Use "-" to read from stdin (single file only).
Use --stdout to write binary output to stdout.

Examples:
  pdf compress large.pdf -o smaller.pdf
  pdf compress document.pdf
  pdf compress *.pdf                      # Batch compress
  cat input.pdf | pdf compress - --stdout > out.pdf  # stdin/stdout`,
	Args: cobra.MinimumNArgs(1),
	RunE: runCompress,
}

func runCompress(cmd *cobra.Command, args []string) error {
	args, err := sanitizeInputArgs(args)
	if err != nil {
		return err
	}

	password, err := cli.GetPasswordSecure(cmd, "Enter PDF password: ")
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}
	output := cli.GetOutput(cmd)
	toStdout := cli.GetStdout(cmd)

	output, err = sanitizeOutputPath(output)
	if err != nil {
		return err
	}

	// Handle dry-run mode
	if cli.IsDryRun() {
		return compressDryRun(args, output, password)
	}

	// Handle stdin/stdout for single file
	if len(args) == 1 && (fileio.IsStdinInput(args[0]) || toStdout) {
		return compressWithStdio(args[0], output, password, toStdout)
	}

	if err := validateBatchOutput(args, output, SuffixCompressed); err != nil {
		return err
	}

	return processBatch(args, func(inputFile string) error {
		return compressFile(inputFile, output, password)
	})
}

func compressDryRun(args []string, explicitOutput, password string) error {
	for _, inputFile := range args {
		if fileio.IsStdinInput(inputFile) {
			cli.DryRunPrint("Would compress: stdin")
			continue
		}

		info, err := pdf.GetInfo(inputFile, password)
		if err != nil {
			cli.DryRunPrint("Would compress: %s (unable to read info)", inputFile)
			continue
		}

		output := outputOrDefault(explicitOutput, inputFile, SuffixCompressed)
		cli.DryRunPrint("Would compress: %s", inputFile)
		cli.DryRunPrint("  Size: %s (%d pages)", fileio.FormatFileSize(info.FileSize), info.Pages)
		cli.DryRunPrint("  Output: %s", output)
	}
	return nil
}

func compressWithStdio(inputArg, explicitOutput, password string, toStdout bool) error {
	handler := &patterns.StdioHandler{
		InputArg:       inputArg,
		ExplicitOutput: explicitOutput,
		ToStdout:       toStdout,
		DefaultSuffix:  SuffixCompressed,
		Operation:      "compress",
	}
	defer handler.Cleanup()

	input, output, err := handler.Setup()
	if err != nil {
		return err
	}

	if !toStdout {
		if err := checkOutputFile(output); err != nil {
			return err
		}
	}

	if err := pdf.Compress(input, output, password); err != nil {
		return pdferrors.WrapError("compressing file", inputArg, err)
	}

	if err := handler.Finalize(); err != nil {
		return err
	}

	if !toStdout {
		newSize, _ := fileio.GetFileSize(output)
		fmt.Fprintf(os.Stderr, "Compressed to %s (%s)\n", output, fileio.FormatFileSize(newSize))
	}
	return nil
}

func compressFile(inputFile, explicitOutput, password string) error {
	if err := fileio.ValidatePDFFile(inputFile); err != nil {
		return err
	}

	originalSize, _ := fileio.GetFileSize(inputFile)
	output := outputOrDefault(explicitOutput, inputFile, SuffixCompressed)

	if err := checkOutputFile(output); err != nil {
		return err
	}

	cli.PrintVerbose("Compressing %s to %s", inputFile, output)

	if err := pdf.Compress(inputFile, output, password); err != nil {
		return pdferrors.WrapError("compressing file", inputFile, err)
	}

	newSize, _ := fileio.GetFileSize(output)
	savings := originalSize - newSize
	savingsPercent := float64(savings) / float64(originalSize) * 100

	fmt.Printf("Compressed %s to %s\n", inputFile, output)
	fmt.Printf("Original:   %s\n", fileio.FormatFileSize(originalSize))
	fmt.Printf("Compressed: %s\n", fileio.FormatFileSize(newSize))
	if savings > 0 {
		fmt.Printf("Saved:      %s (%.1f%%)\n", fileio.FormatFileSize(savings), savingsPercent)
	} else {
		fmt.Println("Note: File size increased (already optimized)")
	}

	return nil
}
