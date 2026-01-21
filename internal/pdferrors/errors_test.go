package pdferrors

import (
	"errors"
	"strings"
	"testing"
)

func TestPDFError_Error(t *testing.T) {
	tests := []struct {
		name string
		err  *PDFError
		want string
	}{
		{
			name: "basic error",
			err: &PDFError{
				Operation: "reading",
				File:      "test.pdf",
				Cause:     errors.New("file not found"),
			},
			want: "reading 'test.pdf': file not found",
		},
		{
			name: "error with hint",
			err: &PDFError{
				Operation: "decrypting",
				File:      "secure.pdf",
				Cause:     ErrPasswordRequired,
				Hint:      "Use --password to provide the document password",
			},
			want: "decrypting 'secure.pdf': password required\nHint: Use --password to provide the document password",
		},
		{
			name: "error without file",
			err: &PDFError{
				Operation: "processing",
				Cause:     errors.New("invalid input"),
			},
			want: "processing: invalid input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("PDFError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewPDFError(t *testing.T) {
	err := NewPDFError("reading", "test.pdf", ErrFileNotFound)

	if err.Operation != "reading" {
		t.Errorf("NewPDFError() Operation = %v, want %v", err.Operation, "reading")
	}
	if err.File != "test.pdf" {
		t.Errorf("NewPDFError() File = %v, want %v", err.File, "test.pdf")
	}
	if !errors.Is(err.Cause, ErrFileNotFound) {
		t.Errorf("NewPDFError() Cause = %v, want %v", err.Cause, ErrFileNotFound)
	}
}

func TestPDFError_WithHint(t *testing.T) {
	err := NewPDFError("reading", "test.pdf", ErrFileNotFound).WithHint("Check the file path")

	if err.Hint != "Check the file path" {
		t.Errorf("WithHint() Hint = %v, want %v", err.Hint, "Check the file path")
	}
}

func TestWrapError(t *testing.T) {
	tests := []struct {
		name      string
		operation string
		file      string
		err       error
		wantNil   bool
		wantCause error
	}{
		{
			name:    "nil error",
			err:     nil,
			wantNil: true,
		},
		{
			name:      "no such file error",
			operation: "reading",
			file:      "test.pdf",
			err:       errors.New("no such file"),
			wantCause: ErrFileNotFound,
		},
		{
			name:      "encrypted error",
			operation: "reading",
			file:      "secure.pdf",
			err:       errors.New("document is encrypted"),
			wantCause: ErrPasswordRequired,
		},
		{
			name:      "invalid PDF error",
			operation: "reading",
			file:      "corrupt.pdf",
			err:       errors.New("invalid PDF"),
			wantCause: ErrCorruptPDF,
		},
		{
			name:      "generic error",
			operation: "processing",
			file:      "test.pdf",
			err:       errors.New("some error"),
			wantCause: nil, // Will be the original error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WrapError(tt.operation, tt.file, tt.err)

			if tt.wantNil {
				if result != nil {
					t.Errorf("WrapError() = %v, want nil", result)
				}
				return
			}

			var pdfErr *PDFError
			if !errors.As(result, &pdfErr) {
				t.Errorf("WrapError() did not return PDFError")
				return
			}

			if tt.wantCause != nil && !errors.Is(pdfErr.Cause, tt.wantCause) {
				t.Errorf("WrapError() cause = %v, want %v", pdfErr.Cause, tt.wantCause)
			}
		})
	}
}

func TestFormatError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "nil error",
			err:  nil,
			want: "",
		},
		{
			name: "PDFError",
			err: &PDFError{
				Operation: "reading",
				File:      "test.pdf",
				Cause:     ErrFileNotFound,
			},
			want: "reading 'test.pdf': file not found",
		},
		{
			name: "regular error",
			err:  errors.New("something went wrong"),
			want: "Error: something went wrong",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatError(tt.err); got != tt.want {
				t.Errorf("FormatError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsFileNotFound(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "file not found",
			err: &PDFError{
				Cause: ErrFileNotFound,
			},
			want: true,
		},
		{
			name: "other error",
			err: &PDFError{
				Cause: ErrCorruptPDF,
			},
			want: false,
		},
		{
			name: "regular error",
			err:  errors.New("not found"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsFileNotFound(tt.err); got != tt.want {
				t.Errorf("IsFileNotFound() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsPasswordRequired(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "password required",
			err: &PDFError{
				Cause: ErrPasswordRequired,
			},
			want: true,
		},
		{
			name: "other error",
			err: &PDFError{
				Cause: ErrFileNotFound,
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsPasswordRequired(tt.err); got != tt.want {
				t.Errorf("IsPasswordRequired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPDFError_Unwrap(t *testing.T) {
	cause := errors.New("original error")
	err := &PDFError{
		Operation: "test",
		Cause:     cause,
	}

	if unwrapped := err.Unwrap(); unwrapped != cause {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, cause)
	}

	// Test with errors.Unwrap
	if unwrapped := errors.Unwrap(err); unwrapped != cause {
		t.Errorf("errors.Unwrap() = %v, want %v", unwrapped, cause)
	}
}

func TestErrorMessages(t *testing.T) {
	// Verify error messages are user-friendly
	errorTests := []struct {
		err  error
		want string
	}{
		{ErrFileNotFound, "file not found"},
		{ErrNotPDF, "not a valid PDF file"},
		{ErrInvalidPages, "invalid page specification"},
		{ErrPasswordRequired, "password required"},
		{ErrWrongPassword, "incorrect password"},
		{ErrCorruptPDF, "PDF file is corrupted"},
		{ErrOutputExists, "output file already exists"},
	}

	for _, tt := range errorTests {
		if !strings.Contains(tt.err.Error(), tt.want) {
			t.Errorf("Error %v should contain %q", tt.err, tt.want)
		}
	}
}
