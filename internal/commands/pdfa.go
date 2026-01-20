package commands

import (
	"fmt"

	"github.com/lgbarn/pdf-cli/internal/cli"
	"github.com/lgbarn/pdf-cli/internal/pdf"
	"github.com/lgbarn/pdf-cli/internal/util"
	"github.com/spf13/cobra"
)

func init() {
	cli.AddCommand(pdfaCmd)
	pdfaCmd.AddCommand(pdfaValidateCmd)
	pdfaCmd.AddCommand(pdfaConvertCmd)

	cli.AddPasswordFlag(pdfaValidateCmd, "Password for encrypted PDFs")
	pdfaValidateCmd.Flags().String("level", "", "PDF/A level to validate: 1b, 2b, 3b (default: any)")

	cli.AddOutputFlag(pdfaConvertCmd, "Output file path")
	cli.AddPasswordFlag(pdfaConvertCmd, "Password for encrypted PDFs")
	pdfaConvertCmd.Flags().String("level", "2b", "Target PDF/A level: 1b, 2b, 3b")
}

var pdfaCmd = &cobra.Command{
	Use:   "pdfa",
	Short: "PDF/A validation and conversion",
	Long: `PDF/A validation and conversion commands.

PDF/A is an ISO-standardized version of PDF specialized for digital preservation.

Note: Full PDF/A validation and conversion may require specialized tools like
veraPDF (validation) or Ghostscript/Adobe Acrobat (conversion). This tool provides
basic validation and optimization that can help with PDF/A compliance.

Available subcommands:
  validate - Check PDF/A compliance
  convert  - Convert/optimize a PDF toward PDF/A format`,
}

var pdfaValidateCmd = &cobra.Command{
	Use:   "validate <file.pdf>",
	Short: "Validate PDF/A compliance",
	Long: `Validate a PDF file for PDF/A compliance.

Performs basic PDF/A compliance checks including:
- PDF structure validation
- Encryption detection (PDF/A prohibits standard encryption)
- PDF version compatibility checks

Note: This is basic validation. For full PDF/A compliance testing,
use specialized tools like veraPDF.

Examples:
  pdf pdfa validate document.pdf
  pdf pdfa validate document.pdf --level 1b
  pdf pdfa validate document.pdf --level 2b --password secret`,
	Args: cobra.ExactArgs(1),
	RunE: runPdfaValidate,
}

var pdfaConvertCmd = &cobra.Command{
	Use:   "convert <file.pdf>",
	Short: "Convert to PDF/A format",
	Long: `Convert a PDF to PDF/A format.

Optimizes the PDF to improve PDF/A compliance. This includes:
- Removing unused objects
- Optimizing internal structure

Note: Full PDF/A conversion may require specialized tools like Ghostscript
or Adobe Acrobat. This tool performs optimization which can help with some
PDF/A requirements but may not achieve full compliance for complex documents.

Examples:
  pdf pdfa convert document.pdf -o archive.pdf
  pdf pdfa convert document.pdf --level 2b -o archive.pdf`,
	Args: cobra.ExactArgs(1),
	RunE: runPdfaConvert,
}

func runPdfaValidate(cmd *cobra.Command, args []string) error {
	inputFile := args[0]
	password := cli.GetPassword(cmd)
	level, _ := cmd.Flags().GetString("level")

	if err := util.ValidatePDFFile(inputFile); err != nil {
		return err
	}

	cli.PrintVerbose("Validating PDF/A compliance for %s", inputFile)
	if level != "" {
		cli.PrintVerbose("Target level: PDF/A-%s", level)
	}

	result, err := pdf.ValidatePDFA(inputFile, level, password)
	if err != nil {
		return util.WrapError("validating PDF/A", inputFile, err)
	}

	if result.IsValid {
		fmt.Printf("✓ %s passes basic PDF/A validation\n", inputFile)
	} else {
		fmt.Printf("✗ %s has PDF/A compliance issues\n", inputFile)
	}

	if len(result.Errors) > 0 {
		fmt.Println("\nErrors:")
		for _, e := range result.Errors {
			fmt.Printf("  - %s\n", e)
		}
	}

	if len(result.Warnings) > 0 {
		fmt.Println("\nWarnings:")
		for _, w := range result.Warnings {
			fmt.Printf("  - %s\n", w)
		}
	}

	return nil
}

func runPdfaConvert(cmd *cobra.Command, args []string) error {
	inputFile := args[0]
	output := cli.GetOutput(cmd)
	password := cli.GetPassword(cmd)
	level, _ := cmd.Flags().GetString("level")

	if err := util.ValidatePDFFile(inputFile); err != nil {
		return err
	}

	output = outputOrDefault(output, inputFile, "_pdfa")

	if err := checkOutputFile(output); err != nil {
		return err
	}

	cli.PrintVerbose("Converting %s to PDF/A-%s format", inputFile, level)

	if err := pdf.ConvertToPDFA(inputFile, output, level, password); err != nil {
		return util.WrapError("converting to PDF/A", inputFile, err)
	}

	fmt.Printf("PDF optimized and saved to %s\n", output)
	fmt.Println("\nNote: Full PDF/A conversion may require specialized tools.")
	fmt.Println("Consider using veraPDF to validate the output.")

	return nil
}
