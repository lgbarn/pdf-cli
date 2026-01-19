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
	cli.AddOutputFlag(encryptCmd, "Output file path")
	cli.AddPasswordFlag(encryptCmd, "User password (required)")
	encryptCmd.Flags().String("owner-password", "", "Owner password (defaults to user password)")
	_ = encryptCmd.MarkFlagRequired("password")
}

var encryptCmd = &cobra.Command{
	Use:   "encrypt <file.pdf>",
	Short: "Add password protection to a PDF",
	Long: `Add password protection to a PDF file.

The user password is required to open the document.
The owner password (optional) controls editing permissions.

Examples:
  pdf encrypt document.pdf --password secret -o secure.pdf
  pdf encrypt document.pdf --password user123 --owner-password admin456 -o protected.pdf`,
	Args: cobra.ExactArgs(1),
	RunE: runEncrypt,
}

func runEncrypt(cmd *cobra.Command, args []string) error {
	inputFile := args[0]
	userPassword := cli.GetPassword(cmd)
	ownerPassword, _ := cmd.Flags().GetString("owner-password")

	if err := util.ValidatePDFFile(inputFile); err != nil {
		return err
	}

	if userPassword == "" {
		return fmt.Errorf("password is required for encryption")
	}

	output := outputOrDefault(cli.GetOutput(cmd), inputFile, "_encrypted")

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
