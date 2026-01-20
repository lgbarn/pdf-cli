package commands

import (
	"fmt"
	"path/filepath"

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
	Use:   "meta <file.pdf> [file2.pdf...]",
	Short: "View or modify PDF metadata",
	Long: `View or modify PDF document metadata.

Without options, displays current metadata.
With options, sets the specified metadata fields.

When viewing multiple files, displays metadata for each file.
Setting metadata only works with a single file.

Examples:
  pdf meta document.pdf
  pdf meta document.pdf --title "My Document" -o updated.pdf
  pdf meta document.pdf --author "John Doe" --subject "Report" -o updated.pdf
  pdf meta *.pdf                                                # View all`,
	Args: cobra.MinimumNArgs(1),
	RunE: runMeta,
}

func runMeta(cmd *cobra.Command, args []string) error {
	output := cli.GetOutput(cmd)
	password := cli.GetPassword(cmd)

	title, _ := cmd.Flags().GetString("title")
	author, _ := cmd.Flags().GetString("author")
	subject, _ := cmd.Flags().GetString("subject")
	keywords, _ := cmd.Flags().GetString("keywords")
	creator, _ := cmd.Flags().GetString("creator")

	isSettingMeta := title != "" || author != "" || subject != "" || keywords != "" || creator != ""

	if isSettingMeta {
		if len(args) > 1 {
			return fmt.Errorf("cannot set metadata on multiple files; use a single file")
		}
		return setMetadata(args[0], output, password, title, author, subject, keywords, creator)
	}

	if len(args) == 1 {
		return viewMetadata(args[0], password)
	}

	return viewBatchMetadata(args, password)
}

func viewMetadata(inputFile, password string) error {
	if err := util.ValidatePDFFile(inputFile); err != nil {
		return err
	}

	cli.PrintVerbose("Reading metadata from %s", inputFile)

	meta, err := pdf.GetMetadata(inputFile, password)
	if err != nil {
		return util.WrapError("reading metadata", inputFile, err)
	}

	fmt.Printf("File: %s\n\n", inputFile)

	if !hasMetadata(meta) {
		fmt.Println("No metadata found")
		return nil
	}

	printMetadataFields(meta)
	return nil
}

func viewBatchMetadata(files []string, password string) error {
	for i, file := range files {
		if i > 0 {
			fmt.Println()
		}

		fmt.Printf("=== %s ===\n", filepath.Base(file))

		if err := util.ValidatePDFFile(file); err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		meta, err := pdf.GetMetadata(file, password)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		if !hasMetadata(meta) {
			fmt.Println("No metadata found")
			continue
		}

		printMetadataFields(meta)
	}

	return nil
}

func setMetadata(inputFile, output, password, title, author, subject, keywords, creator string) error {
	if err := util.ValidatePDFFile(inputFile); err != nil {
		return err
	}

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

func hasMetadata(meta *pdf.Metadata) bool {
	return meta.Title != "" || meta.Author != "" || meta.Subject != "" ||
		meta.Keywords != "" || meta.Creator != "" || meta.Producer != ""
}

func printMetadataFields(meta *pdf.Metadata) {
	printIfSet("Title", meta.Title)
	printIfSet("Author", meta.Author)
	printIfSet("Subject", meta.Subject)
	printIfSet("Keywords", meta.Keywords)
	printIfSet("Creator", meta.Creator)
	printIfSet("Producer", meta.Producer)
}
