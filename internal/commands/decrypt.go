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
	cli.AddCommand(decryptCmd)
	cli.AddOutputFlag(decryptCmd, "Output file path (only with single file)")
	cli.AddPasswordFlag(decryptCmd, "Password for the encrypted PDF (required)")
	cli.AddPasswordFileFlag(decryptCmd, "")
	cli.AddStdoutFlag(decryptCmd)
}

var decryptCmd = &cobra.Command{
	Use:   "decrypt <file.pdf> [file2.pdf...]",
	Short: "Remove password protection from PDF(s)",
	Long: `Remove password protection from encrypted PDF file(s).

Requires the correct password to decrypt the files.
The output files will be unprotected PDFs.

Supports batch processing of multiple files. When processing
multiple files, output files are named with '_decrypted' suffix.
Use "-" to read from stdin. Use --stdout for binary output.

Examples:
  pdf decrypt secure.pdf --password secret -o unlocked.pdf
  pdf decrypt protected.pdf --password mypassword
  cat secure.pdf | pdf decrypt - --password secret --stdout > unlocked.pdf`,
	Args: cobra.MinimumNArgs(1),
	RunE: runDecrypt,
}

func runDecrypt(cmd *cobra.Command, args []string) error {
	// Sanitize input paths
	sanitizedArgs, err := fileio.SanitizePaths(args)
	if err != nil {
		return fmt.Errorf("invalid file path: %w", err)
	}
	args = sanitizedArgs

	password, err := cli.GetPasswordSecure(cmd, "Enter PDF password: ")
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}
	if password == "" {
		return fmt.Errorf("password is required (use --password-file, PDF_CLI_PASSWORD env var, or interactive prompt)")
	}

	output := cli.GetOutput(cmd)
	toStdout := cli.GetStdout(cmd)

	// Sanitize output path
	if output != "" && output != "-" {
		output, err = fileio.SanitizePath(output)
		if err != nil {
			return fmt.Errorf("invalid output path: %w", err)
		}
	}

	// Handle dry-run mode
	if cli.IsDryRun() {
		return decryptDryRun(args, output, password)
	}

	// Handle stdin/stdout for single file
	if len(args) == 1 && (fileio.IsStdinInput(args[0]) || toStdout) {
		return decryptWithStdio(args[0], output, password, toStdout)
	}

	if err := validateBatchOutput(args, output, "_decrypted"); err != nil {
		return err
	}

	return processBatch(args, func(inputFile string) error {
		return decryptFile(inputFile, output, password)
	})
}

func decryptDryRun(args []string, explicitOutput, password string) error {
	for _, inputFile := range args {
		if fileio.IsStdinInput(inputFile) {
			cli.DryRunPrint("Would decrypt: stdin")
			continue
		}

		info, err := pdf.GetInfo(inputFile, password)
		if err != nil {
			cli.DryRunPrint("Would decrypt: %s (unable to read info - may need password)", inputFile)
			continue
		}

		output := outputOrDefault(explicitOutput, inputFile, "_decrypted")
		cli.DryRunPrint("Would decrypt: %s (%d pages)", inputFile, info.Pages)
		cli.DryRunPrint("  Encrypted: %t", info.Encrypted)
		cli.DryRunPrint("  Output: %s", output)
	}
	return nil
}

func decryptWithStdio(inputArg, explicitOutput, password string, toStdout bool) error {
	handler := &patterns.StdioHandler{
		InputArg:       inputArg,
		ExplicitOutput: explicitOutput,
		ToStdout:       toStdout,
		DefaultSuffix:  "_decrypted",
		Operation:      "decrypt",
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

	if err := pdf.Decrypt(input, output, password); err != nil {
		return pdferrors.WrapError("decrypting file", inputArg, err)
	}

	if err := handler.Finalize(); err != nil {
		return err
	}

	if !toStdout {
		fmt.Fprintf(os.Stderr, "Decrypted to %s\n", output)
	}
	return nil
}

func decryptFile(inputFile, explicitOutput, password string) error {
	if err := fileio.ValidatePDFFile(inputFile); err != nil {
		return err
	}

	output := outputOrDefault(explicitOutput, inputFile, "_decrypted")

	if err := checkOutputFile(output); err != nil {
		return err
	}

	cli.PrintVerbose("Decrypting %s to %s", inputFile, output)

	if err := pdf.Decrypt(inputFile, output, password); err != nil {
		return pdferrors.WrapError("decrypting file", inputFile, err)
	}

	fmt.Printf("Decrypted %s to %s\n", inputFile, output)
	return nil
}
