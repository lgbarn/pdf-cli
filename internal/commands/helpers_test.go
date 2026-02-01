package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/lgbarn/pdf-cli/internal/cli"
	"github.com/lgbarn/pdf-cli/internal/pdf"
)

// testdataDir returns the absolute path to the testdata directory
func testdataDir() string {
	// Use absolute path to avoid triggering path sanitization on ".." components
	abs, err := filepath.Abs(filepath.Join("..", "..", "testdata"))
	if err != nil {
		return filepath.Join("..", "..", "testdata")
	}
	return abs
}

// samplePDF returns the path to the sample PDF file
func samplePDF() string {
	return filepath.Join(testdataDir(), "sample.pdf")
}

// resetFlags resets flag values to their defaults
// This is needed because cobra persists flag values between tests
func resetFlags(t *testing.T) {
	t.Helper()
	rootCmd := cli.GetRootCmd()
	// Reset global flags
	_ = rootCmd.PersistentFlags().Set("verbose", "false")
	_ = rootCmd.PersistentFlags().Set("force", "false")
	_ = rootCmd.PersistentFlags().Set("dry-run", "false")

	// Reset subcommand flags by finding and resetting each one
	for _, cmd := range rootCmd.Commands() {
		// Reset common flags if they exist
		if f := cmd.Flags().Lookup("output"); f != nil {
			_ = cmd.Flags().Set("output", "")
		}
		// Note: split uses "pages" differently (-n not -p)
		if f := cmd.Flags().Lookup("pages"); f != nil {
			if cmd.Name() == "split" {
				_ = cmd.Flags().Set("pages", "1")
			} else {
				_ = cmd.Flags().Set("pages", "")
			}
		}
		if f := cmd.Flags().Lookup("password"); f != nil {
			_ = cmd.Flags().Set("password", "")
		}
		if f := cmd.Flags().Lookup("text"); f != nil {
			_ = cmd.Flags().Set("text", "")
		}
		if f := cmd.Flags().Lookup("image"); f != nil {
			_ = cmd.Flags().Set("image", "")
		}
		if f := cmd.Flags().Lookup("angle"); f != nil {
			_ = cmd.Flags().Set("angle", "90")
		}
		if f := cmd.Flags().Lookup("owner-password"); f != nil {
			_ = cmd.Flags().Set("owner-password", "")
		}
		if f := cmd.Flags().Lookup("stdout"); f != nil {
			_ = cmd.Flags().Set("stdout", "false")
		}
		if f := cmd.Flags().Lookup("ocr"); f != nil {
			_ = cmd.Flags().Set("ocr", "false")
		}
		if f := cmd.Flags().Lookup("format"); f != nil {
			_ = cmd.Flags().Set("format", "")
		}
		// Reset meta flags
		if f := cmd.Flags().Lookup("title"); f != nil {
			_ = cmd.Flags().Set("title", "")
		}
		if f := cmd.Flags().Lookup("author"); f != nil {
			_ = cmd.Flags().Set("author", "")
		}
		if f := cmd.Flags().Lookup("subject"); f != nil {
			_ = cmd.Flags().Set("subject", "")
		}
		if f := cmd.Flags().Lookup("keywords"); f != nil {
			_ = cmd.Flags().Set("keywords", "")
		}
		if f := cmd.Flags().Lookup("creator"); f != nil {
			_ = cmd.Flags().Set("creator", "")
		}
	}
}

// executeCommand runs a command and captures output
func executeCommand(args ...string) error {
	rootCmd := cli.GetRootCmd()
	rootCmd.SetArgs(args)
	// Capture output to avoid polluting test output
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})
	return rootCmd.Execute()
}

func TestParseAndValidatePages(t *testing.T) {
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tests := []struct {
		name    string
		spec    string
		file    string
		want    []int
		wantErr bool
	}{
		{"empty string", "", samplePDF(), nil, false},
		{"single page", "1", samplePDF(), []int{1}, false},
		{"page range", "1-3", samplePDF(), []int{1, 2, 3}, false},
		{"mixed selection", "1,3", samplePDF(), []int{1, 3}, false},
		{"same page range", "1-1", samplePDF(), []int{1}, false},
		{"reversed range", "5-1", samplePDF(), nil, true},
		{"zero page", "0", samplePDF(), nil, true},
		{"page exceeds count", "9999", samplePDF(), nil, true},
		{"invalid format", "abc", samplePDF(), nil, true},
		{"non-existent file", "1", "/nonexistent/file.pdf", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pages, err := parseAndValidatePages(tt.spec, tt.file, "")
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(pages) != len(tt.want) {
				t.Errorf("got %v, want %v", pages, tt.want)
			}
		})
	}
}

func TestTruncateString(t *testing.T) {
	tests := []struct {
		s      string
		maxLen int
		want   string
	}{
		{"hello", 10, "hello"},
		{"hello", 5, "hello"},
		{"hello world", 8, "hello..."},
		{"hello", 3, "hel"},
		{"hello", 2, "he"},
		{"", 10, ""},
		{"test", 4, "test"},
	}

	for _, tt := range tests {
		if got := truncateString(tt.s, tt.maxLen); got != tt.want {
			t.Errorf("truncateString(%q, %d) = %q, want %q", tt.s, tt.maxLen, got, tt.want)
		}
	}
}

func TestPrintIfSet(t *testing.T) {
	for _, tc := range []struct{ label, value string }{
		{"Label", ""},
		{"Label", "Value"},
		{"", ""},
	} {
		printIfSet(tc.label, tc.value)
	}
}

func TestHasMetadata(t *testing.T) {
	tests := []struct {
		name string
		meta *pdf.Metadata
		want bool
	}{
		{"empty", &pdf.Metadata{}, false},
		{"title", &pdf.Metadata{Title: "Test"}, true},
		{"author", &pdf.Metadata{Author: "John"}, true},
		{"subject", &pdf.Metadata{Subject: "Testing"}, true},
		{"keywords", &pdf.Metadata{Keywords: "test"}, true},
		{"creator", &pdf.Metadata{Creator: "App"}, true},
		{"producer", &pdf.Metadata{Producer: "Prod"}, true},
		{"all fields", &pdf.Metadata{Title: "T", Author: "A", Subject: "S", Keywords: "K", Creator: "C", Producer: "P"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasMetadata(tt.meta); got != tt.want {
				t.Errorf("hasMetadata() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPrintMetadataFields(t *testing.T) {
	for _, meta := range []*pdf.Metadata{
		{},
		{Title: "T", Author: "A", Subject: "S", Keywords: "K", Creator: "C", Producer: "P"},
		{Title: "Only Title", Author: "Only Author"},
	} {
		printMetadataFields(meta)
	}
}
