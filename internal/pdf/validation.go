package pdf

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pdfcpu/pdfcpu/pkg/api"
)

// Validate validates a PDF file
func Validate(path, password string) error {
	return api.ValidateFile(path, NewConfig(password))
}

// ValidateToBuffer validates a PDF from bytes
func ValidateToBuffer(data []byte) error {
	return api.Validate(bytes.NewReader(data), NewConfig(""))
}

// PDFAValidationResult contains the result of PDF/A validation.
type PDFAValidationResult struct {
	IsValid  bool
	Level    string
	Errors   []string
	Warnings []string
}

// ValidatePDFA validates a PDF for PDF/A compliance.
// Note: pdfcpu provides basic validation; full PDF/A validation requires specialized tools.
func ValidatePDFA(path, level, password string) (*PDFAValidationResult, error) {
	conf := NewConfig(password)

	if err := api.ValidateFile(path, conf); err != nil {
		return &PDFAValidationResult{
			IsValid: false,
			Level:   level,
			Errors:  []string{fmt.Sprintf("PDF validation failed: %v", err)},
		}, nil
	}

	cleanPath := filepath.Clean(path)
	f, err := os.Open(cleanPath) // #nosec G304 -- path is cleaned
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() { _ = f.Close() }()

	info, err := api.PDFInfo(f, path, nil, false, conf)
	if err != nil {
		return nil, fmt.Errorf("failed to read PDF info: %w", err)
	}

	result := &PDFAValidationResult{
		Level:    level,
		Warnings: []string{"Note: This is basic PDF/A validation. Full compliance testing requires specialized tools like veraPDF."},
	}

	if info.Encrypted {
		result.Errors = append(result.Errors, "PDF/A documents should not use standard encryption")
	}

	if (level == "1b" || level == "1a") && info.Version != "1.4" {
		result.Warnings = append(result.Warnings, fmt.Sprintf("PDF/A-1 recommends PDF version 1.4, found %s", info.Version))
	}

	result.IsValid = len(result.Errors) == 0
	return result, nil
}

// ConvertToPDFA attempts to convert a PDF to PDF/A format.
// Note: pdfcpu has limited PDF/A conversion capabilities. This performs optimization
// which may help with some PDF/A requirements, but full conversion requires
// specialized tools like Ghostscript or Adobe Acrobat.
func ConvertToPDFA(input, output, _ /* level */, password string) error {
	return api.OptimizeFile(input, output, NewConfig(password))
}
