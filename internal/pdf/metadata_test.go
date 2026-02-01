package pdf

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetMetadata(t *testing.T) {
	pdf := samplePDF()
	if _, err := os.Stat(pdf); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	meta, err := GetMetadata(pdf, "")
	if err != nil {
		t.Fatalf("GetMetadata() error = %v", err)
	}

	if meta == nil {
		t.Fatal("GetMetadata() returned nil")
	}
}

func TestGetMetadataNonExistent(t *testing.T) {
	_, err := GetMetadata("/nonexistent/file.pdf", "")
	if err == nil {
		t.Error("GetMetadata() expected error for non-existent file")
	}
}

func TestSetMetadata(t *testing.T) {
	pdfFile := samplePDF()
	if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output := filepath.Join(tmpDir, "metadata.pdf")
	meta := &Metadata{
		Title:    "Test Title",
		Author:   "Test Author",
		Subject:  "Test Subject",
		Keywords: "test, keywords",
	}

	err = SetMetadata(pdfFile, output, meta, "")
	if err != nil {
		t.Fatalf("SetMetadata() error = %v", err)
	}

	if _, err := os.Stat(output); os.IsNotExist(err) {
		t.Error("SetMetadata() did not create output file")
	}
}

func TestSetMetadataPartial(t *testing.T) {
	pdfFile := samplePDF()
	if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output := filepath.Join(tmpDir, "metadata.pdf")
	meta := &Metadata{
		Title: "Only Title",
	}

	err = SetMetadata(pdfFile, output, meta, "")
	if err != nil {
		t.Fatalf("SetMetadata() with partial metadata error = %v", err)
	}

	if _, err := os.Stat(output); os.IsNotExist(err) {
		t.Error("SetMetadata() did not create output file")
	}
}

func TestSetMetadataEmpty(t *testing.T) {
	pdfFile := samplePDF()
	if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output := filepath.Join(tmpDir, "metadata.pdf")
	meta := &Metadata{}

	// Empty metadata should still work
	err = SetMetadata(pdfFile, output, meta, "")
	if err != nil {
		t.Fatalf("SetMetadata() with empty metadata error = %v", err)
	}
}

func TestSetMetadataNonExistent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output := filepath.Join(tmpDir, "metadata.pdf")
	meta := &Metadata{Title: "Test"}
	err = SetMetadata("/nonexistent/file.pdf", output, meta, "")
	if err == nil {
		t.Error("SetMetadata() expected error for non-existent file")
	}
}

func TestSetMetadataAllFields(t *testing.T) {
	pdfFile := samplePDF()
	if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output := filepath.Join(tmpDir, "metadata_all.pdf")
	meta := &Metadata{
		Title:    "Full Test Title",
		Author:   "Full Test Author",
		Subject:  "Full Test Subject",
		Keywords: "full, test, keywords",
		Creator:  "Full Test Creator",
		Producer: "Full Test Producer",
	}

	err = SetMetadata(pdfFile, output, meta, "")
	if err != nil {
		t.Fatalf("SetMetadata() with all fields error = %v", err)
	}

	if _, err := os.Stat(output); os.IsNotExist(err) {
		t.Error("SetMetadata() did not create output file")
	}
}

func TestInfoStruct(t *testing.T) {
	info := Info{
		FilePath:    "/path/to/file.pdf",
		FileSize:    1024,
		Pages:       10,
		Version:     "1.4",
		Title:       "Test PDF",
		Author:      "Test Author",
		Subject:     "Test Subject",
		Keywords:    "test, pdf",
		Creator:     "Test Creator",
		Producer:    "Test Producer",
		CreatedDate: "2024-01-01",
		ModDate:     "2024-01-02",
		Encrypted:   false,
	}

	if info.FilePath != "/path/to/file.pdf" {
		t.Errorf("Info.FilePath = %q, want %q", info.FilePath, "/path/to/file.pdf")
	}
	if info.Pages != 10 {
		t.Errorf("Info.Pages = %d, want %d", info.Pages, 10)
	}
}

func TestMetadataStruct(t *testing.T) {
	meta := Metadata{
		Title:       "Test Title",
		Author:      "Test Author",
		Subject:     "Test Subject",
		Keywords:    "test, keywords",
		Creator:     "Test Creator",
		Producer:    "Test Producer",
		CreatedDate: "2024-01-01",
		ModDate:     "2024-01-02",
	}

	if meta.Title != "Test Title" {
		t.Errorf("Metadata.Title = %q, want %q", meta.Title, "Test Title")
	}
}

func TestPDFAValidationResultStruct(t *testing.T) {
	result := PDFAValidationResult{
		IsValid:  false,
		Level:    "1b",
		Errors:   []string{"error1", "error2"},
		Warnings: []string{"warning1"},
	}

	if result.IsValid {
		t.Error("PDFAValidationResult.IsValid should be false")
	}
	if result.Level != "1b" {
		t.Errorf("PDFAValidationResult.Level = %q, want %q", result.Level, "1b")
	}
	if len(result.Errors) != 2 {
		t.Errorf("PDFAValidationResult.Errors length = %d, want 2", len(result.Errors))
	}
}

func TestValidatePDFA(t *testing.T) {
	pdfFile := samplePDF()
	if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	result, err := ValidatePDFA(pdfFile, "1b", "")
	if err != nil {
		t.Fatalf("ValidatePDFA() error = %v", err)
	}

	if result == nil {
		t.Fatal("ValidatePDFA() returned nil result")
		return
	}

	if result.Level != "1b" {
		t.Errorf("ValidatePDFA() Level = %q, want %q", result.Level, "1b")
	}

	// Note: regular PDFs are unlikely to pass PDF/A validation
	// We just verify the function works without crashing
}

func TestValidatePDFALevels(t *testing.T) {
	pdfFile := samplePDF()
	if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	levels := []string{"1a", "1b", "2a", "2b", "3a", "3b"}

	for _, level := range levels {
		t.Run(level, func(t *testing.T) {
			result, err := ValidatePDFA(pdfFile, level, "")
			if err != nil {
				t.Fatalf("ValidatePDFA(%q) error = %v", level, err)
			}
			if result == nil {
				t.Fatalf("ValidatePDFA(%q) returned nil", level)
				return
			}
			if result.Level != level {
				t.Errorf("ValidatePDFA() Level = %q, want %q", result.Level, level)
			}
		})
	}
}

func TestValidatePDFANonExistent(t *testing.T) {
	result, err := ValidatePDFA("/nonexistent/file.pdf", "1b", "")
	// ValidatePDFA may return a result with IsValid=false instead of an error
	if err == nil && result != nil && result.IsValid {
		t.Error("ValidatePDFA() should report invalid for non-existent file")
	}
}

func TestValidatePDFAInvalidPDF(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create an invalid PDF file
	invalidPDF := filepath.Join(tmpDir, "invalid.pdf")
	err = os.WriteFile(invalidPDF, []byte("not a valid pdf"), 0644)
	if err != nil {
		t.Fatalf("Failed to write invalid PDF: %v", err)
	}

	result, err := ValidatePDFA(invalidPDF, "1b", "")
	// Either returns error or result with IsValid=false
	if err == nil && result != nil && result.IsValid {
		t.Error("ValidatePDFA() should report invalid for invalid PDF")
	}
}

func TestConvertToPDFA(t *testing.T) {
	pdfFile := samplePDF()
	if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output := filepath.Join(tmpDir, "pdfa.pdf")
	err = ConvertToPDFA(pdfFile, output, "1b", "")
	if err != nil {
		t.Fatalf("ConvertToPDFA() error = %v", err)
	}

	// Verify output exists
	if _, err := os.Stat(output); os.IsNotExist(err) {
		t.Error("ConvertToPDFA() did not create output file")
	}
}

func TestConvertToPDFANonExistent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output := filepath.Join(tmpDir, "pdfa.pdf")
	err = ConvertToPDFA("/nonexistent/file.pdf", output, "1b", "")
	if err == nil {
		t.Error("ConvertToPDFA() expected error for non-existent file")
	}
}
