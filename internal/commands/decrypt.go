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
	cli.AddOutputFlag(decryptCmd, "Output file path")
	cli.AddPasswordFlag(decryptCmd, "Password for the encrypted PDF (required)")
	decryptCmd.MarkFlagRequired("password")
}

var decryptCmd = &cobra.Command{
	Use:   "decrypt <file.pdf>",
	Short: "Remove password protection from a PDF",
	Long: `Remove password protection from an encrypted PDF.

Requires the correct password to decrypt the file.
The output file will be an unprotected PDF.

Examples:
  pdf decrypt secure.pdf --password secret -o unlocked.pdf
  pdf decrypt protected.pdf --password mypassword`,
	Args: cobra.ExactArgs(1),
	RunE: runDecrypt,
}

func runDecrypt(cmd *cobra.Command, args []string) error {
	inputFile := args[0]
	password := cli.GetPassword(cmd)

	if err := util.ValidatePDFFile(inputFile); err != nil {
		return err
	}

	if password == "" {
		return fmt.Errorf("password is required for decryption")
	}

	output := outputOrDefault(cli.GetOutput(cmd), inputFile, "_decrypted")

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
