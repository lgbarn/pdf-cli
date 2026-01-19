package commands

import (
	"fmt"

	"github.com/lgbarn/pdf-cli/internal/cli"
	"github.com/lgbarn/pdf-cli/internal/pdf"
	"github.com/lgbarn/pdf-cli/internal/util"
	"github.com/spf13/cobra"
)

func init() {
	cli.AddCommand(infoCmd)
	cli.AddPasswordFlag(infoCmd, "Password for encrypted PDFs")
}

var infoCmd = &cobra.Command{
	Use:   "info <file.pdf>",
	Short: "Display PDF information",
	Long: `Display detailed information about a PDF file.

Shows file size, page count, PDF version, encryption status,
and metadata like title, author, subject, and keywords.

Examples:
  pdf info document.pdf
  pdf info encrypted.pdf --password secret`,
	Args: cobra.ExactArgs(1),
	RunE: runInfo,
}

func runInfo(cmd *cobra.Command, args []string) error {
	inputFile := args[0]
	password := cli.GetPassword(cmd)

	if err := util.ValidatePDFFile(inputFile); err != nil {
		return err
	}

	cli.PrintVerbose("Reading PDF info from %s", inputFile)

	info, err := pdf.GetInfo(inputFile, password)
	if err != nil {
		return util.WrapError("reading info", inputFile, err)
	}

	fmt.Printf("File:       %s\n", info.FilePath)
	fmt.Printf("Size:       %s\n", util.FormatFileSize(info.FileSize))
	fmt.Printf("Pages:      %d\n", info.Pages)
	fmt.Printf("Version:    PDF %s\n", info.Version)
	fmt.Printf("Encrypted:  %t\n", info.Encrypted)

	printIfSet("Title", info.Title)
	printIfSet("Author", info.Author)
	printIfSet("Subject", info.Subject)
	printIfSet("Keywords", info.Keywords)
	printIfSet("Creator", info.Creator)
	printIfSet("Producer", info.Producer)

	return nil
}

// printIfSet prints a labeled value only if the value is non-empty
func printIfSet(label, value string) {
	if value != "" {
		fmt.Printf("%-11s %s\n", label+":", value)
	}
}
