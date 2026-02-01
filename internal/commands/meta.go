package commands

import (
	"fmt"
	"path/filepath"

	"github.com/lgbarn/pdf-cli/internal/cli"
	"github.com/lgbarn/pdf-cli/internal/fileio"
	"github.com/lgbarn/pdf-cli/internal/output"
	"github.com/lgbarn/pdf-cli/internal/pdf"
	"github.com/lgbarn/pdf-cli/internal/pdferrors"
	"github.com/spf13/cobra"
)

func init() {
	cli.AddCommand(metaCmd)
	cli.AddOutputFlag(metaCmd, "Output file path (for setting metadata)")
	cli.AddPasswordFlag(metaCmd, "Password for encrypted PDFs")
	cli.AddPasswordFileFlag(metaCmd, "")
	cli.AddFormatFlag(metaCmd)
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
	// Sanitize input paths
	sanitizedArgs, err := fileio.SanitizePaths(args)
	if err != nil {
		return fmt.Errorf("invalid file path: %w", err)
	}
	args = sanitizedArgs

	outputFile := cli.GetOutput(cmd)
	// Sanitize output path if provided
	if outputFile != "" {
		outputFile, err = fileio.SanitizePath(outputFile)
		if err != nil {
			return fmt.Errorf("invalid output path: %w", err)
		}
	}

	password, err := cli.GetPasswordSecure(cmd, "Enter PDF password: ")
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}
	format := cli.GetFormat(cmd)
	formatter := output.NewOutputFormatter(format)

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
		return setMetadata(args[0], outputFile, password, title, author, subject, keywords, creator)
	}

	if len(args) == 1 {
		return viewMetadata(args[0], password, formatter)
	}

	return viewBatchMetadata(args, password, formatter)
}

// MetadataOutput represents PDF metadata for structured output.
type MetadataOutput struct {
	File     string `json:"file"`
	Title    string `json:"title,omitempty"`
	Author   string `json:"author,omitempty"`
	Subject  string `json:"subject,omitempty"`
	Keywords string `json:"keywords,omitempty"`
	Creator  string `json:"creator,omitempty"`
	Producer string `json:"producer,omitempty"`
}

func viewMetadata(inputFile, password string, formatter *output.OutputFormatter) error {
	if err := fileio.ValidatePDFFile(inputFile); err != nil {
		return err
	}

	cli.PrintVerbose("Reading metadata from %s", inputFile)

	meta, err := pdf.GetMetadata(inputFile, password)
	if err != nil {
		return pdferrors.WrapError("reading metadata", inputFile, err)
	}

	// Structured output (JSON/CSV/TSV)
	if formatter.IsStructured() {
		output := MetadataOutput{
			File:     inputFile,
			Title:    meta.Title,
			Author:   meta.Author,
			Subject:  meta.Subject,
			Keywords: meta.Keywords,
			Creator:  meta.Creator,
			Producer: meta.Producer,
		}
		return formatter.Print(output)
	}

	// Human-readable output
	fmt.Printf("File: %s\n\n", inputFile)

	if !hasMetadata(meta) {
		fmt.Println("No metadata found")
		return nil
	}

	printMetadataFields(meta)
	return nil
}

func viewBatchMetadata(files []string, password string, formatter *output.OutputFormatter) error {
	// Structured output (JSON/CSV/TSV)
	if formatter.IsStructured() {
		var outputs []MetadataOutput
		for _, file := range files {
			if err := fileio.ValidatePDFFile(file); err != nil {
				continue
			}
			meta, err := pdf.GetMetadata(file, password)
			if err != nil {
				continue
			}
			outputs = append(outputs, MetadataOutput{
				File:     file,
				Title:    meta.Title,
				Author:   meta.Author,
				Subject:  meta.Subject,
				Keywords: meta.Keywords,
				Creator:  meta.Creator,
				Producer: meta.Producer,
			})
		}

		if formatter.Format == output.FormatJSON {
			return formatter.Print(outputs)
		}

		// CSV/TSV: use table format
		headers := []string{"file", "title", "author", "subject", "keywords", "creator", "producer"}
		var rows [][]string
		for _, o := range outputs {
			rows = append(rows, []string{
				o.File, o.Title, o.Author, o.Subject, o.Keywords, o.Creator, o.Producer,
			})
		}
		return formatter.PrintTable(headers, rows)
	}

	// Human-readable output
	for i, file := range files {
		if i > 0 {
			fmt.Println()
		}

		fmt.Printf("=== %s ===\n", filepath.Base(file))

		if err := fileio.ValidatePDFFile(file); err != nil {
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

func setMetadata(inputFile, outputFile, password, title, author, subject, keywords, creator string) error {
	if err := fileio.ValidatePDFFile(inputFile); err != nil {
		return err
	}

	outputFile = outputOrDefault(outputFile, inputFile, "_updated")

	// Handle dry-run mode
	if cli.IsDryRun() {
		cli.DryRunPrint("Would set metadata on: %s", inputFile)
		if title != "" {
			cli.DryRunPrint("  Title: \"%s\"", title)
		}
		if author != "" {
			cli.DryRunPrint("  Author: \"%s\"", author)
		}
		if subject != "" {
			cli.DryRunPrint("  Subject: \"%s\"", subject)
		}
		if keywords != "" {
			cli.DryRunPrint("  Keywords: \"%s\"", keywords)
		}
		if creator != "" {
			cli.DryRunPrint("  Creator: \"%s\"", creator)
		}
		cli.DryRunPrint("  Output: %s", outputFile)
		return nil
	}

	if err := checkOutputFile(outputFile); err != nil {
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

	if err := pdf.SetMetadata(inputFile, outputFile, meta, password); err != nil {
		return pdferrors.WrapError("setting metadata", inputFile, err)
	}

	fmt.Printf("Metadata updated in %s\n", outputFile)
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
