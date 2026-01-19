package commands

import (
	"os"
	"path/filepath"
	"testing"
)

// testdataDir returns the path to the testdata directory
func testdataDir() string {
	return filepath.Join("..", "..", "testdata")
}

// samplePDF returns the path to the sample PDF file
func samplePDF() string {
	return filepath.Join(testdataDir(), "sample.pdf")
}

func TestParseAndValidatePages_EmptyString(t *testing.T) {
	// Empty string should return nil (meaning all pages)
	pages, err := parseAndValidatePages("", samplePDF(), "")
	if err != nil {
		t.Fatalf("parseAndValidatePages('') error = %v, want nil", err)
	}
	if pages != nil {
		t.Errorf("parseAndValidatePages('') = %v, want nil", pages)
	}
}

func TestParseAndValidatePages_SinglePage(t *testing.T) {
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	pages, err := parseAndValidatePages("1", samplePDF(), "")
	if err != nil {
		t.Fatalf("parseAndValidatePages('1') error = %v", err)
	}
	if len(pages) != 1 || pages[0] != 1 {
		t.Errorf("parseAndValidatePages('1') = %v, want [1]", pages)
	}
}

func TestParseAndValidatePages_PageRange(t *testing.T) {
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	pages, err := parseAndValidatePages("1-3", samplePDF(), "")
	if err != nil {
		t.Fatalf("parseAndValidatePages('1-3') error = %v", err)
	}
	expected := []int{1, 2, 3}
	if len(pages) != len(expected) {
		t.Fatalf("parseAndValidatePages('1-3') length = %d, want %d", len(pages), len(expected))
	}
	for i, p := range pages {
		if p != expected[i] {
			t.Errorf("parseAndValidatePages('1-3')[%d] = %d, want %d", i, p, expected[i])
		}
	}
}

func TestParseAndValidatePages_MixedSelection(t *testing.T) {
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	pages, err := parseAndValidatePages("1,3", samplePDF(), "")
	if err != nil {
		t.Fatalf("parseAndValidatePages('1,3') error = %v", err)
	}
	expected := []int{1, 3}
	if len(pages) != len(expected) {
		t.Fatalf("parseAndValidatePages('1,3') length = %d, want %d", len(pages), len(expected))
	}
	for i, p := range pages {
		if p != expected[i] {
			t.Errorf("parseAndValidatePages('1,3')[%d] = %d, want %d", i, p, expected[i])
		}
	}
}

func TestParseAndValidatePages_InvalidReversedRange(t *testing.T) {
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	_, err := parseAndValidatePages("5-1", samplePDF(), "")
	if err == nil {
		t.Error("parseAndValidatePages('5-1') expected error for reversed range")
	}
}

func TestParseAndValidatePages_InvalidZeroPage(t *testing.T) {
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	_, err := parseAndValidatePages("0", samplePDF(), "")
	if err == nil {
		t.Error("parseAndValidatePages('0') expected error for zero page")
	}
}

func TestParseAndValidatePages_InvalidPageExceedsCount(t *testing.T) {
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	// Use a very large page number that won't exist
	_, err := parseAndValidatePages("9999", samplePDF(), "")
	if err == nil {
		t.Error("parseAndValidatePages('9999') expected error for page exceeding count")
	}
}

func TestParseAndValidatePages_InvalidFormat(t *testing.T) {
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	_, err := parseAndValidatePages("abc", samplePDF(), "")
	if err == nil {
		t.Error("parseAndValidatePages('abc') expected error for invalid format")
	}
}

func TestParseAndValidatePages_NonExistentFile(t *testing.T) {
	_, err := parseAndValidatePages("1", "/nonexistent/file.pdf", "")
	if err == nil {
		t.Error("parseAndValidatePages() expected error for non-existent file")
	}
}

func TestParseAndValidatePages_ComplexRange(t *testing.T) {
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	// Test a complex but valid range for page 1
	pages, err := parseAndValidatePages("1-1", samplePDF(), "")
	if err != nil {
		t.Fatalf("parseAndValidatePages('1-1') error = %v", err)
	}
	if len(pages) != 1 || pages[0] != 1 {
		t.Errorf("parseAndValidatePages('1-1') = %v, want [1]", pages)
	}
}
