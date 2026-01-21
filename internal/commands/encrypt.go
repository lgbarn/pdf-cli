package commands

import (
	"fmt"
	"os"

	"github.com/lgbarn/pdf-cli/internal/cli"
	"github.com/lgbarn/pdf-cli/internal/fileio"
	"github.com/lgbarn/pdf-cli/internal/pdf"
	"github.com/lgbarn/pdf-cli/internal/pdferrors"
	"github.com/spf13/cobra"
)

func init() {
	cli.AddCommand(encryptCmd)
	cli.AddOutputFlag(encryptCmd, "Output file path (only with single file)")
	cli.AddPasswordFlag(encryptCmd, "User password (required)")
	cli.AddStdoutFlag(encryptCmd)
	encryptCmd.Flags().String("owner-password", "", "Owner password (defaults to user password)")
	_ = encryptCmd.MarkFlagRequired("password")
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
	userPassword := cli.GetPassword(cmd)
	ownerPassword, _ := cmd.Flags().GetString("owner-password")
	output := cli.GetOutput(cmd)
	toStdout := cli.GetStdout(cmd)

	if userPassword == "" {
		return fmt.Errorf("password is required for encryption")
	}

	// Handle stdin/stdout for single file
	if len(args) == 1 && (fileio.IsStdinInput(args[0]) || toStdout) {
		return encryptWithStdio(args[0], output, userPassword, ownerPassword, toStdout)
	}

	if err := validateBatchOutput(args, output, "_encrypted"); err != nil {
		return err
	}

	return processBatch(args, func(inputFile string) error {
		return encryptFile(inputFile, output, userPassword, ownerPassword)
	})
}

func encryptWithStdio(inputArg, explicitOutput, userPassword, ownerPassword string, toStdout bool) error {
	// Handle stdin input
	inputFile, cleanup, err := fileio.ResolveInputPath(inputArg)
	if err != nil {
		return err
	}
	defer cleanup()

	// Handle stdout output
	var output string
	var outputCleanup func()
	if toStdout {
		tmpFile, err := os.CreateTemp("", "pdf-cli-encrypt-*.pdf")
		if err != nil {
			return fmt.Errorf("failed to create temp file: %w", err)
		}
		output = tmpFile.Name()
		_ = tmpFile.Close()
		outputCleanup = func() { _ = os.Remove(output) }
		defer outputCleanup()
	} else {
		output = outputOrDefault(explicitOutput, inputArg, "_encrypted")
		if err := checkOutputFile(output); err != nil {
			return err
		}
	}

	if err := pdf.Encrypt(inputFile, output, userPassword, ownerPassword); err != nil {
		return pdferrors.WrapError("encrypting file", inputArg, err)
	}

	if toStdout {
		return fileio.WriteToStdout(output)
	}

	fmt.Fprintf(os.Stderr, "Encrypted to %s\n", output)
	return nil
}

func encryptFile(inputFile, explicitOutput, userPassword, ownerPassword string) error {
	if err := fileio.ValidatePDFFile(inputFile); err != nil {
		return err
	}

	output := outputOrDefault(explicitOutput, inputFile, "_encrypted")

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
