package commands

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/lgbarn/pdf-cli/internal/cli"
	"github.com/lgbarn/pdf-cli/internal/fileio"
	"github.com/lgbarn/pdf-cli/internal/output"
	"github.com/lgbarn/pdf-cli/internal/pdf"
	"github.com/lgbarn/pdf-cli/internal/pdferrors"
	"github.com/spf13/cobra"
)

func init() {
	cli.AddCommand(infoCmd)
	cli.AddPasswordFlag(infoCmd, "Password for encrypted PDFs")
	cli.AddPasswordFileFlag(infoCmd, "")
	cli.AddAllowInsecurePasswordFlag(infoCmd)
	cli.AddFormatFlag(infoCmd)
}

var infoCmd = &cobra.Command{
	Use:   "info <file.pdf> [file2.pdf...]",
	Short: "Display PDF information",
	Long: `Display detailed information about a PDF file.

Shows file size, page count, PDF version, encryption status,
and metadata like title, author, subject, and keywords.

When multiple files are provided, displays a summary table.
Use "-" to read a single file from stdin.

Examples:
  pdf info document.pdf
  pdf info encrypted.pdf --password secret
  pdf info *.pdf                           # Batch mode: show summary table
  cat document.pdf | pdf info -            # Read from stdin
  pdf info document.pdf --format json      # JSON output`,
	Args: cobra.MinimumNArgs(1),
	RunE: runInfo,
}

func runInfo(cmd *cobra.Command, args []string) error {
	args, err := sanitizeInputArgs(args)
	if err != nil {
		return err
	}

	password, err := cli.GetPasswordSecure(cmd, "Enter PDF password: ")
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}

	format := cli.GetFormat(cmd)
	formatter := output.NewOutputFormatter(format)

	// Single file: detailed output (supports stdin)
	if len(args) == 1 {
		inputArg := args[0]

		// Handle stdin input
		inputFile, cleanup, err := fileio.ResolveInputPath(inputArg)
		if err != nil {
			return err
		}
		defer cleanup()

		return displaySingleInfo(inputFile, password, formatter, fileio.IsStdinInput(inputArg))
	}

	// Multiple files: table output
	return displayBatchInfo(args, password, formatter)
}

// InfoOutput represents PDF info for structured output.
type InfoOutput struct {
	File      string            `json:"file"`
	Size      int64             `json:"size"`
	SizeHuman string            `json:"size_human"`
	Pages     int               `json:"pages"`
	Version   string            `json:"version"`
	Encrypted bool              `json:"encrypted"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

func displaySingleInfo(inputFile, password string, formatter *output.OutputFormatter, isStdin bool) error {
	if !isStdin {
		if err := fileio.ValidatePDFFile(inputFile); err != nil {
			return err
		}
	}

	cli.PrintVerbose("Reading PDF info from %s", inputFile)

	info, err := pdf.GetInfo(inputFile, password)
	if err != nil {
		return pdferrors.WrapError("reading info", inputFile, err)
	}

	// Structured output (JSON/CSV/TSV)
	if formatter.IsStructured() {
		output := InfoOutput{
			File:      info.FilePath,
			Size:      info.FileSize,
			SizeHuman: fileio.FormatFileSize(info.FileSize),
			Pages:     info.Pages,
			Version:   info.Version,
			Encrypted: info.Encrypted,
			Metadata:  make(map[string]string),
		}
		if info.Title != "" {
			output.Metadata["title"] = info.Title
		}
		if info.Author != "" {
			output.Metadata["author"] = info.Author
		}
		if info.Subject != "" {
			output.Metadata["subject"] = info.Subject
		}
		if info.Keywords != "" {
			output.Metadata["keywords"] = info.Keywords
		}
		if info.Creator != "" {
			output.Metadata["creator"] = info.Creator
		}
		if info.Producer != "" {
			output.Metadata["producer"] = info.Producer
		}
		return formatter.Print(output)
	}

	// Human-readable output
	fmt.Printf("File:       %s\n", info.FilePath)
	fmt.Printf("Size:       %s\n", fileio.FormatFileSize(info.FileSize))
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

func displayBatchInfo(files []string, password string, formatter *output.OutputFormatter) error {
	// Structured output (JSON/CSV/TSV)
	if formatter.IsStructured() {
		var outputs []InfoOutput
		for _, file := range files {
			if err := fileio.ValidatePDFFile(file); err != nil {
				continue
			}
			info, err := pdf.GetInfo(file, password)
			if err != nil {
				continue
			}
			output := InfoOutput{
				File:      info.FilePath,
				Size:      info.FileSize,
				SizeHuman: fileio.FormatFileSize(info.FileSize),
				Pages:     info.Pages,
				Version:   info.Version,
				Encrypted: info.Encrypted,
				Metadata:  make(map[string]string),
			}
			if info.Title != "" {
				output.Metadata["title"] = info.Title
			}
			if info.Author != "" {
				output.Metadata["author"] = info.Author
			}
			outputs = append(outputs, output)
		}

		if formatter.Format == output.FormatJSON {
			return formatter.Print(outputs)
		}

		// CSV/TSV: use table format
		headers := []string{"file", "pages", "version", "size", "encrypted"}
		var rows [][]string
		for _, o := range outputs {
			rows = append(rows, []string{
				o.File,
				strconv.Itoa(o.Pages),
				o.Version,
				o.SizeHuman,
				strconv.FormatBool(o.Encrypted),
			})
		}
		return formatter.PrintTable(headers, rows)
	}

	// Human-readable output
	fmt.Printf("%-40s %8s %6s %10s\n", "FILE", "PAGES", "VER", "SIZE")
	fmt.Println(strings.Repeat("-", 70))

	var hasErrors bool
	for _, file := range files {
		if err := fileio.ValidatePDFFile(file); err != nil {
			fmt.Printf("%-40s ERROR: %v\n", truncateString(filepath.Base(file), 40), err)
			hasErrors = true
			continue
		}

		info, err := pdf.GetInfo(file, password)
		if err != nil {
			fmt.Printf("%-40s ERROR: %v\n", truncateString(filepath.Base(file), 40), err)
			hasErrors = true
			continue
		}

		fmt.Printf("%-40s %8d %6s %10s\n",
			truncateString(filepath.Base(file), 40),
			info.Pages,
			info.Version,
			fileio.FormatFileSize(info.FileSize))
	}

	if hasErrors {
		fmt.Println()
		fmt.Println("Some files had errors. Use -v for details.")
	}

	return nil
}

func printIfSet(label, value string) {
	if value != "" {
		fmt.Printf("%-11s %s\n", label+":", value)
	}
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
