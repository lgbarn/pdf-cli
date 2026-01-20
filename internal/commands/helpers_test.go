package commands

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lgbarn/pdf-cli/internal/pdf"
)

// testdataDir returns the path to the testdata directory
func testdataDir() string {
	return filepath.Join("..", "..", "testdata")
}

// samplePDF returns the path to the sample PDF file
func samplePDF() string {
	return filepath.Join(testdataDir(), "sample.pdf")
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
