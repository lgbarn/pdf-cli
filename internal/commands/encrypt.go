package commands

import (
	"fmt"

	"github.com/lgbarn/pdf-cli/internal/cli"
	"github.com/lgbarn/pdf-cli/internal/pdf"
	"github.com/lgbarn/pdf-cli/internal/util"
	"github.com/spf13/cobra"
)

func init() {
	cli.AddCommand(encryptCmd)
	cli.AddOutputFlag(encryptCmd, "Output file path (only with single file)")
	cli.AddPasswordFlag(encryptCmd, "User password (required)")
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

Examples:
  pdf encrypt document.pdf --password secret -o secure.pdf
  pdf encrypt document.pdf --password user123 --owner-password admin456 -o protected.pdf
  pdf encrypt *.pdf --password secret        # Batch encrypt
  pdf encrypt doc1.pdf doc2.pdf --password s # Multiple files`,
	Args: cobra.MinimumNArgs(1),
	RunE: runEncrypt,
}

func runEncrypt(cmd *cobra.Command, args []string) error {
	userPassword := cli.GetPassword(cmd)
	ownerPassword, _ := cmd.Flags().GetString("owner-password")
	output := cli.GetOutput(cmd)

	if userPassword == "" {
		return fmt.Errorf("password is required for encryption")
	}

	if err := validateBatchOutput(args, output, "_encrypted"); err != nil {
		return err
	}

	return processBatch(args, func(inputFile string) error {
		return encryptFile(inputFile, output, userPassword, ownerPassword)
	})
}

func encryptFile(inputFile, explicitOutput, userPassword, ownerPassword string) error {
	if err := util.ValidatePDFFile(inputFile); err != nil {
		return err
	}

	output := outputOrDefault(explicitOutput, inputFile, "_encrypted")

	if err := checkOutputFile(output); err != nil {
		return err
	}

	cli.PrintVerbose("Encrypting %s to %s", inputFile, output)

	if err := pdf.Encrypt(inputFile, output, userPassword, ownerPassword); err != nil {
		return util.WrapError("encrypting file", inputFile, err)
	}

	fmt.Printf("Encrypted %s to %s\n", inputFile, output)
	return nil
}
