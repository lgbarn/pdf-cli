package commands

import (
	"fmt"
	"os"

	"github.com/lgbarn/pdf-cli/internal/cli"
	"github.com/lgbarn/pdf-cli/internal/pdf"
	"github.com/lgbarn/pdf-cli/internal/util"
	"github.com/spf13/cobra"
)

func init() {
	cli.AddCommand(decryptCmd)
	cli.AddOutputFlag(decryptCmd, "Output file path (only with single file)")
	cli.AddPasswordFlag(decryptCmd, "Password for the encrypted PDF (required)")
	cli.AddStdoutFlag(decryptCmd)
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
Use "-" to read from stdin. Use --stdout for binary output.

Examples:
  pdf decrypt secure.pdf --password secret -o unlocked.pdf
  pdf decrypt protected.pdf --password mypassword
  cat secure.pdf | pdf decrypt - --password secret --stdout > unlocked.pdf`,
	Args: cobra.MinimumNArgs(1),
	RunE: runDecrypt,
}

func runDecrypt(cmd *cobra.Command, args []string) error {
	password := cli.GetPassword(cmd)
	output := cli.GetOutput(cmd)
	toStdout := cli.GetStdout(cmd)

	if password == "" {
		return fmt.Errorf("password is required for decryption")
	}

	// Handle stdin/stdout for single file
	if len(args) == 1 && (util.IsStdinInput(args[0]) || toStdout) {
		return decryptWithStdio(args[0], output, password, toStdout)
	}

	if err := validateBatchOutput(args, output, "_decrypted"); err != nil {
		return err
	}

	return processBatch(args, func(inputFile string) error {
		return decryptFile(inputFile, output, password)
	})
}

func decryptWithStdio(inputArg, explicitOutput, password string, toStdout bool) error {
	// Handle stdin input
	inputFile, cleanup, err := util.ResolveInputPath(inputArg)
	if err != nil {
		return err
	}
	defer cleanup()

	// Handle stdout output
	var output string
	var outputCleanup func()
	if toStdout {
		tmpFile, err := os.CreateTemp("", "pdf-cli-decrypt-*.pdf")
		if err != nil {
			return fmt.Errorf("failed to create temp file: %w", err)
		}
		output = tmpFile.Name()
		_ = tmpFile.Close()
		outputCleanup = func() { _ = os.Remove(output) }
		defer outputCleanup()
	} else {
		output = outputOrDefault(explicitOutput, inputArg, "_decrypted")
		if err := checkOutputFile(output); err != nil {
			return err
		}
	}

	if err := pdf.Decrypt(inputFile, output, password); err != nil {
		return util.WrapError("decrypting file", inputArg, err)
	}

	if toStdout {
		return util.WriteToStdout(output)
	}

	fmt.Fprintf(os.Stderr, "Decrypted to %s\n", output)
	return nil
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
