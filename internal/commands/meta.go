package commands

import (
	"fmt"

	"github.com/lgbarn/pdf-cli/internal/cli"
	"github.com/lgbarn/pdf-cli/internal/pdf"
	"github.com/lgbarn/pdf-cli/internal/util"
	"github.com/spf13/cobra"
)

func init() {
	cli.AddCommand(metaCmd)
	cli.AddOutputFlag(metaCmd, "Output file path (for setting metadata)")
	cli.AddPasswordFlag(metaCmd, "Password for encrypted PDFs")
	metaCmd.Flags().String("title", "", "Set document title")
	metaCmd.Flags().String("author", "", "Set document author")
	metaCmd.Flags().String("subject", "", "Set document subject")
	metaCmd.Flags().String("keywords", "", "Set document keywords")
	metaCmd.Flags().String("creator", "", "Set document creator")
}

var metaCmd = &cobra.Command{
	Use:   "meta <file.pdf>",
	Short: "View or modify PDF metadata",
	Long: `View or modify PDF document metadata.

Without options, displays current metadata.
With options, sets the specified metadata fields.

Examples:
  pdf meta document.pdf
  pdf meta document.pdf --title "My Document" -o updated.pdf
  pdf meta document.pdf --author "John Doe" --subject "Report" -o updated.pdf`,
	Args: cobra.ExactArgs(1),
	RunE: runMeta,
}

func runMeta(cmd *cobra.Command, args []string) error {
	inputFile := args[0]
	output := cli.GetOutput(cmd)
	password := cli.GetPassword(cmd)

	title, _ := cmd.Flags().GetString("title")
	author, _ := cmd.Flags().GetString("author")
	subject, _ := cmd.Flags().GetString("subject")
	keywords, _ := cmd.Flags().GetString("keywords")
	creator, _ := cmd.Flags().GetString("creator")

	// Validate input file
	if err := util.ValidatePDFFile(inputFile); err != nil {
		return err
	}

	// Check if we're setting or viewing metadata
	isSettingMeta := title != "" || author != "" || subject != "" || keywords != "" || creator != ""

	if isSettingMeta {
		return setMetadata(inputFile, output, password, title, author, subject, keywords, creator)
	}

	return viewMetadata(inputFile, password)
}

func viewMetadata(inputFile, password string) error {
	cli.PrintVerbose("Reading metadata from %s", inputFile)

	meta, err := pdf.GetMetadata(inputFile, password)
	if err != nil {
		return util.WrapError("reading metadata", inputFile, err)
	}

	fmt.Printf("File: %s\n\n", inputFile)

	hasMetadata := meta.Title != "" || meta.Author != "" || meta.Subject != "" ||
		meta.Keywords != "" || meta.Creator != "" || meta.Producer != ""

	if !hasMetadata {
		fmt.Println("No metadata found")
		return nil
	}

	printIfSet("Title", meta.Title)
	printIfSet("Author", meta.Author)
	printIfSet("Subject", meta.Subject)
	printIfSet("Keywords", meta.Keywords)
	printIfSet("Creator", meta.Creator)
	printIfSet("Producer", meta.Producer)

	return nil
}

func setMetadata(inputFile, output, password, title, author, subject, keywords, creator string) error {
	output = outputOrDefault(output, inputFile, "_updated")

	if err := checkOutputFile(output); err != nil {
		return err
	}

	meta := &pdf.Metadata{
		Title:    title,
		Author:   author,
		Subject:  subject,
		Keywords: keywords,
		Creator:  creator,
	}

	cli.PrintVerbose("Setting metadata on %s", inputFile)

	if err := pdf.SetMetadata(inputFile, output, meta, password); err != nil {
		return util.WrapError("setting metadata", inputFile, err)
	}

	fmt.Printf("Metadata updated in %s\n", output)
	return nil
}
