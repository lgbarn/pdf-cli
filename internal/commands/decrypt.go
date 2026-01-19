package commands

import (
	"fmt"

	"github.com/lgbarn/pdf-cli/internal/cli"
	"github.com/lgbarn/pdf-cli/internal/pdf"
	"github.com/lgbarn/pdf-cli/internal/util"
	"github.com/spf13/cobra"
)

func init() {
	cli.AddCommand(decryptCmd)
	cli.AddOutputFlag(decryptCmd, "Output file path (only with single file)")
	cli.AddPasswordFlag(decryptCmd, "Password for the encrypted PDF (required)")
	_ = decryptCmd.MarkFlagRequired("password")
}

var decryptCmd = &cobra.Command{
	Use:   "decrypt <file.pdf> [file2.pdf...]",
	Short: "Remove password protection from PDF(s)",
	Long: `Remove password protection from encrypted PDF file(s).

Requires the correct password to decrypt the files.
The output files will be unprotected PDFs.

Supports batch processing of multiple files. When processing
multiple files, output files are named with '_decrypted' suffix.

Examples:
  pdf decrypt secure.pdf --password secret -o unlocked.pdf
  pdf decrypt protected.pdf --password mypassword
  pdf decrypt *.pdf --password secret        # Batch decrypt
  pdf decrypt doc1.pdf doc2.pdf --password s # Multiple files`,
	Args: cobra.MinimumNArgs(1),
	RunE: runDecrypt,
}

func runDecrypt(cmd *cobra.Command, args []string) error {
	password := cli.GetPassword(cmd)
	output := cli.GetOutput(cmd)

	if password == "" {
		return fmt.Errorf("password is required for decryption")
	}

	if err := validateBatchOutput(args, output, "_decrypted"); err != nil {
		return err
	}

	return processBatch(args, func(inputFile string) error {
		return decryptFile(inputFile, output, password)
	})
}

func decryptFile(inputFile, explicitOutput, password string) error {
	if err := util.ValidatePDFFile(inputFile); err != nil {
		return err
	}

	output := outputOrDefault(explicitOutput, inputFile, "_decrypted")

	if err := checkOutputFile(output); err != nil {
		return err
	}

	cli.PrintVerbose("Decrypting %s to %s", inputFile, output)

	if err := pdf.Decrypt(inputFile, output, password); err != nil {
		return util.WrapError("decrypting file", inputFile, err)
	}

	fmt.Printf("Decrypted %s to %s\n", inputFile, output)
	return nil
}
