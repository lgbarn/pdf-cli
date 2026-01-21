package pdferrors

import (
	"errors"
	"fmt"
	"strings"
)

// PDFError represents a user-friendly error for PDF operations
type PDFError struct {
	Operation string
	File      string
	Cause     error
	Hint      string
}

func (e *PDFError) Error() string {
	var sb strings.Builder
	sb.WriteString(e.Operation)
	if e.File != "" {
		sb.WriteString(fmt.Sprintf(" '%s'", e.File))
	}
	sb.WriteString(": ")
	sb.WriteString(e.Cause.Error())
	if e.Hint != "" {
		sb.WriteString(fmt.Sprintf("\nHint: %s", e.Hint))
	}
	return sb.String()
}

func (e *PDFError) Unwrap() error {
	return e.Cause
}

// NewPDFError creates a new PDFError
func NewPDFError(operation, file string, cause error) *PDFError {
	return &PDFError{
		Operation: operation,
		File:      file,
		Cause:     cause,
	}
}

// WithHint adds a hint to a PDFError
func (e *PDFError) WithHint(hint string) *PDFError {
	e.Hint = hint
	return e
}

// Common error types
var (
	ErrFileNotFound     = errors.New("file not found")
	ErrNotPDF           = errors.New("not a valid PDF file")
	ErrInvalidPages     = errors.New("invalid page specification")
	ErrPasswordRequired = errors.New("password required")
	ErrWrongPassword    = errors.New("incorrect password")
	ErrCorruptPDF       = errors.New("PDF file is corrupted")
	ErrOutputExists     = errors.New("output file already exists")
)

// WrapError wraps an error with additional context
func WrapError(operation string, file string, err error) error {
	if err == nil {
		return nil
	}

	// Check for known error patterns and provide user-friendly messages
	errStr := err.Error()

	switch {
	case strings.Contains(errStr, "no such file"):
		return &PDFError{
			Operation: operation,
			File:      file,
			Cause:     ErrFileNotFound,
		}
	case strings.Contains(errStr, "encrypted") || strings.Contains(errStr, "password"):
		return &PDFError{
			Operation: operation,
			File:      file,
			Cause:     ErrPasswordRequired,
			Hint:      "Use --password to provide the document password",
		}
	case strings.Contains(errStr, "invalid PDF") || strings.Contains(errStr, "malformed"):
		return &PDFError{
			Operation: operation,
			File:      file,
			Cause:     ErrCorruptPDF,
		}
	default:
		return &PDFError{
			Operation: operation,
			File:      file,
			Cause:     err,
		}
	}
}

// FormatError formats an error for display to the user
func FormatError(err error) string {
	if err == nil {
		return ""
	}

	var pdfErr *PDFError
	if errors.As(err, &pdfErr) {
		return pdfErr.Error()
	}

	return fmt.Sprintf("Error: %v", err)
}

// IsFileNotFound checks if an error is a file not found error
func IsFileNotFound(err error) bool {
	var pdfErr *PDFError
	if errors.As(err, &pdfErr) {
		return errors.Is(pdfErr.Cause, ErrFileNotFound)
	}
	return false
}

// IsPasswordRequired checks if an error indicates a password is required
func IsPasswordRequired(err error) bool {
	var pdfErr *PDFError
	if errors.As(err, &pdfErr) {
		return errors.Is(pdfErr.Cause, ErrPasswordRequired)
	}
	return false
}
