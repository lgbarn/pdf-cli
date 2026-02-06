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
	cli.AddCommand(encryptCmd)
	cli.AddOutputFlag(encryptCmd, "Output file path (only with single file)")
	cli.AddPasswordFlag(encryptCmd, "User password (required)")
	cli.AddPasswordFileFlag(encryptCmd, "")
	cli.AddAllowInsecurePasswordFlag(encryptCmd)
	cli.AddStdoutFlag(encryptCmd)
	encryptCmd.Flags().String("owner-password", "", "Owner password (defaults to user password)")
}

var encryptCmd = &cobra.Command{
	Use:   "encrypt <file.pdf> [file2.pdf...]",
	Short: "Add password protection to PDF(s)",
	Long: `Add password protection to PDF file(s).

The user password is required to open the document.
The owner password (optional) controls editing permissions.

Supports batch processing of multiple files. When processing
multiple files, output files are named with '_encrypted' suffix.
Use "-" to read from stdin. Use --stdout for binary output.

Examples:
  pdf encrypt document.pdf --password secret -o secure.pdf
  pdf encrypt document.pdf --password user123 --owner-password admin456
  cat in.pdf | pdf encrypt - --password secret --stdout > secure.pdf`,
	Args: cobra.MinimumNArgs(1),
	RunE: runEncrypt,
}

func runEncrypt(cmd *cobra.Command, args []string) error {
	args, err := sanitizeInputArgs(args)
	if err != nil {
		return err
	}

	userPassword, err := cli.GetPasswordSecure(cmd, "Enter PDF password: ")
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}
	if userPassword == "" {
		return fmt.Errorf("password is required (use --password-file, PDF_CLI_PASSWORD env var, or interactive prompt)")
	}

	ownerPassword, _ := cmd.Flags().GetString("owner-password")
	output := cli.GetOutput(cmd)
	toStdout := cli.GetStdout(cmd)

	output, err = sanitizeOutputPath(output)
	if err != nil {
		return err
	}

	// Handle dry-run mode
	if cli.IsDryRun() {
		return encryptDryRun(args, output, ownerPassword != "")
	}

	// Handle stdin/stdout for single file
	if len(args) == 1 && (fileio.IsStdinInput(args[0]) || toStdout) {
		return encryptWithStdio(args[0], output, userPassword, ownerPassword, toStdout)
	}

	if err := validateBatchOutput(args, output, SuffixEncrypted); err != nil {
		return err
	}

	return processBatch(args, func(inputFile string) error {
		return encryptFile(inputFile, output, userPassword, ownerPassword)
	})
}

func encryptDryRun(args []string, explicitOutput string, hasOwnerPassword bool) error {
	for _, inputFile := range args {
		if fileio.IsStdinInput(inputFile) {
			cli.DryRunPrint("Would encrypt: stdin")
			continue
		}

		info, err := pdf.GetInfo(inputFile, "")
		if err != nil {
			cli.DryRunPrint("Would encrypt: %s (unable to read info)", inputFile)
			continue
		}

		output := outputOrDefault(explicitOutput, inputFile, SuffixEncrypted)
		cli.DryRunPrint("Would encrypt: %s (%d pages)", inputFile, info.Pages)
		cli.DryRunPrint("  Output: %s", output)
		if hasOwnerPassword {
			cli.DryRunPrint("  Owner password: set")
		}
	}
	return nil
}

func encryptWithStdio(inputArg, explicitOutput, userPassword, ownerPassword string, toStdout bool) error {
	handler := &patterns.StdioHandler{
		InputArg:       inputArg,
		ExplicitOutput: explicitOutput,
		ToStdout:       toStdout,
		DefaultSuffix:  SuffixEncrypted,
		Operation:      "encrypt",
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

	if err := pdf.Encrypt(input, output, userPassword, ownerPassword); err != nil {
		return pdferrors.WrapError("encrypting file", inputArg, err)
	}

	if err := handler.Finalize(); err != nil {
		return err
	}

	if !toStdout {
		fmt.Fprintf(os.Stderr, "Encrypted to %s\n", output)
	}
	return nil
}

func encryptFile(inputFile, explicitOutput, userPassword, ownerPassword string) error {
	if err := fileio.ValidatePDFFile(inputFile); err != nil {
		return err
	}

	output := outputOrDefault(explicitOutput, inputFile, SuffixEncrypted)

	if err := checkOutputFile(output); err != nil {
		return err
	}

	cli.PrintVerbose("Encrypting %s to %s", inputFile, output)

	if err := pdf.Encrypt(inputFile, output, userPassword, ownerPassword); err != nil {
		return pdferrors.WrapError("encrypting file", inputFile, err)
	}

	fmt.Printf("Encrypted %s to %s\n", inputFile, output)
	return nil
}
