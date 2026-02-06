package commands

import (
	"fmt"
	"os"
	"strings"

	"github.com/lgbarn/pdf-cli/internal/cli"
	"github.com/lgbarn/pdf-cli/internal/config"
	"github.com/lgbarn/pdf-cli/internal/fileio"
	"github.com/lgbarn/pdf-cli/internal/ocr"
	"github.com/lgbarn/pdf-cli/internal/pdf"
	"github.com/lgbarn/pdf-cli/internal/pdferrors"
	"github.com/spf13/cobra"
)

func init() {
	cli.AddCommand(textCmd)
	cli.AddOutputFlag(textCmd, "Output file path (default: stdout)")
	cli.AddPagesFlag(textCmd, "Pages to extract text from (default: all)")
	cli.AddPasswordFlag(textCmd, "Password for encrypted PDFs")
	cli.AddPasswordFileFlag(textCmd, "")
	cli.AddAllowInsecurePasswordFlag(textCmd)
	textCmd.Flags().Bool("ocr", false, "Use OCR for image-based PDFs")
	textCmd.Flags().String("ocr-lang", "eng", "OCR language(s), e.g., 'eng' or 'eng+fra'")
	textCmd.Flags().String("ocr-backend", "auto", "OCR backend: auto (native if available, else wasm), native, or wasm")
}

var textCmd = &cobra.Command{
	Use:   "text <file.pdf>",
	Short: "Extract text content from a PDF",
	Long: `Extract text content from a PDF file.

By default, extracts text from all pages and prints to stdout.
Use -o to save to a file, or -p to extract from specific pages.
Use "-" to read from stdin.

For scanned or image-based PDFs, use --ocr to enable OCR text extraction.
OCR requires downloading tessdata on first use (~15MB per language).

Examples:
  pdf text document.pdf
  pdf text document.pdf -o content.txt
  pdf text document.pdf -p 1-5 -o chapter1.txt
  pdf text scanned.pdf --ocr                    # OCR for scanned PDF
  pdf text scanned.pdf --ocr --ocr-lang eng+fra # Multi-language OCR
  cat document.pdf | pdf text -                 # Read from stdin`,
	Args: cobra.ExactArgs(1),
	RunE: runText,
}

func runText(cmd *cobra.Command, args []string) error {
	// Sanitize input path
	sanitizedPath, err := fileio.SanitizePath(args[0])
	if err != nil {
		return fmt.Errorf("invalid file path: %w", err)
	}
	inputArg := sanitizedPath

	output := cli.GetOutput(cmd)
	output, err = sanitizeOutputPath(output)
	if err != nil {
		return err
	}

	pagesStr := cli.GetPages(cmd)
	password, err := cli.GetPasswordSecure(cmd, "Enter PDF password: ")
	if err != nil {
		return fmt.Errorf("failed to read password: %w", err)
	}
	useOCR, _ := cmd.Flags().GetBool("ocr")
	ocrLang, _ := cmd.Flags().GetString("ocr-lang")
	ocrBackend, _ := cmd.Flags().GetString("ocr-backend")

	// Handle stdin input
	inputFile, cleanup, err := fileio.ResolveInputPath(inputArg)
	if err != nil {
		return err
	}
	defer cleanup()

	if !fileio.IsStdinInput(inputArg) {
		if err := fileio.ValidatePDFFile(inputFile); err != nil {
			return err
		}
	}

	pages, err := parseAndValidatePages(pagesStr, inputFile, password)
	if err != nil {
		return err
	}

	var text string

	if useOCR {
		backendType := ocr.ParseBackendType(ocrBackend)
		cli.PrintVerbose("Extracting text from %s using OCR (language: %s, backend: %s)", inputFile, ocrLang, ocrBackend)

		cfg := config.Get()
		engine, err := ocr.NewEngineWithOptions(ocr.EngineOptions{
			Lang:              ocrLang,
			BackendType:       backendType,
			ParallelThreshold: cfg.Performance.OCRParallelThreshold,
			MaxWorkers:        cfg.Performance.MaxWorkers,
		})
		if err != nil {
			return pdferrors.WrapError("initializing OCR", inputFile, err)
		}
		defer engine.Close()

		cli.PrintVerbose("Using OCR backend: %s", engine.BackendName())

		text, err = engine.ExtractTextFromPDF(cmd.Context(), inputFile, pages, password, cli.Progress())
		if err != nil {
			return pdferrors.WrapError("extracting text with OCR", inputFile, err)
		}
	} else {
		cli.PrintVerbose("Extracting text from %s", inputFile)

		text, err = pdf.ExtractTextWithProgress(cmd.Context(), inputFile, pages, password, cli.Progress())
		if err != nil {
			return pdferrors.WrapError("extracting text", inputFile, err)
		}

		if strings.TrimSpace(text) == "" {
			cli.PrintStatus("No text found. Try using --ocr for scanned/image-based PDFs.")
		}
	}

	if output == "" {
		fmt.Print(text)
		return nil
	}

	if err := checkOutputFile(output); err != nil {
		return err
	}

	if err := os.WriteFile(output, []byte(text), fileio.DefaultFilePerm); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}
	fmt.Printf("Extracted text saved to %s\n", output)
	return nil
}
