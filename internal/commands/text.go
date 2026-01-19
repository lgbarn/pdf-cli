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
	cli.AddCommand(textCmd)
	cli.AddOutputFlag(textCmd, "Output file path (default: stdout)")
	cli.AddPagesFlag(textCmd, "Pages to extract text from (default: all)")
	cli.AddPasswordFlag(textCmd, "Password for encrypted PDFs")
}

var textCmd = &cobra.Command{
	Use:   "text <file.pdf>",
	Short: "Extract text content from a PDF",
	Long: `Extract text content from a PDF file.

By default, extracts text from all pages and prints to stdout.
Use -o to save to a file, or -p to extract from specific pages.

Examples:
  pdf text document.pdf
  pdf text document.pdf -o content.txt
  pdf text document.pdf -p 1-5 -o chapter1.txt`,
	Args: cobra.ExactArgs(1),
	RunE: runText,
}

func runText(cmd *cobra.Command, args []string) error {
	inputFile := args[0]
	output := cli.GetOutput(cmd)
	pagesStr := cli.GetPages(cmd)
	password := cli.GetPassword(cmd)

	if err := util.ValidatePDFFile(inputFile); err != nil {
		return err
	}

	pages, err := parseAndValidatePages(pagesStr, inputFile, password)
	if err != nil {
		return err
	}

	cli.PrintVerbose("Extracting text from %s", inputFile)

	text, err := pdf.ExtractText(inputFile, pages, password)
	if err != nil {
		return util.WrapError("extracting text", inputFile, err)
	}

	if output == "" {
		fmt.Print(text)
		return nil
	}

	if err := checkOutputFile(output); err != nil {
		return err
	}

	if err := os.WriteFile(output, []byte(text), 0600); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}
	fmt.Printf("Extracted text saved to %s\n", output)
	return nil
}
