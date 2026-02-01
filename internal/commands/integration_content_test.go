package commands

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTextCommand_AllPages(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	// Should not error even if PDF has no text
	if err := executeCommand("text", samplePDF()); err != nil {
		t.Fatalf("text command failed: %v", err)
	}
}

func TestTextCommand_SpecificPages(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	if err := executeCommand("text", samplePDF(), "-p", "1"); err != nil {
		t.Fatalf("text command with page selection failed: %v", err)
	}
}

func TestTextCommand_OutputToFile(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output := filepath.Join(tmpDir, "text.txt")
	if err := executeCommand("text", samplePDF(), "-o", output); err != nil {
		t.Fatalf("text command failed: %v", err)
	}

	if _, err := os.Stat(output); os.IsNotExist(err) {
		t.Error("text did not create output file")
	}
}
func TestInfoCommand(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	if err := executeCommand("info", samplePDF()); err != nil {
		t.Fatalf("info command failed: %v", err)
	}
}

func TestExtractCommand(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output := filepath.Join(tmpDir, "extracted.pdf")
	if err := executeCommand("extract", samplePDF(), "-p", "1", "-o", output); err != nil {
		t.Fatalf("extract command failed: %v", err)
	}

	if _, err := os.Stat(output); os.IsNotExist(err) {
		t.Error("extract did not create output file")
	}
}

func TestMetaCommand(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	if err := executeCommand("meta", samplePDF()); err != nil {
		t.Fatalf("meta command failed: %v", err)
	}
}
func TestPdfaValidateCommand(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	// Just verify it runs without panicking - validation may return errors
	// since sample.pdf may not be PDF/A compliant
	_ = executeCommand("pdfa", "validate", samplePDF())
}

func TestPdfaValidateCommand_WithLevel(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	// Just verify it runs - may report non-compliance
	_ = executeCommand("pdfa", "validate", samplePDF(), "--level", "1b")
}

func TestPdfaConvertCommand(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output := filepath.Join(tmpDir, "pdfa.pdf")
	// This may fail or succeed depending on PDF/A support
	_ = executeCommand("pdfa", "convert", samplePDF(), "-o", output)
}
func TestInfoCommand_WithFormat(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	if err := executeCommand("info", samplePDF(), "--format", "json"); err != nil {
		t.Fatalf("info command with json format failed: %v", err)
	}
}

func TestInfoCommand_MultipleFiles(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	// Test batch processing by passing multiple files
	if err := executeCommand("info", samplePDF(), samplePDF()); err != nil {
		t.Fatalf("info command with multiple files failed: %v", err)
	}
}

func TestMetaCommand_WithFormat(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	if err := executeCommand("meta", samplePDF(), "--format", "json"); err != nil {
		t.Fatalf("meta command with json format failed: %v", err)
	}
}

func TestMetaCommand_MultipleFiles(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	// Test batch processing by passing multiple files
	if err := executeCommand("meta", samplePDF(), samplePDF()); err != nil {
		t.Fatalf("meta command with multiple files failed: %v", err)
	}
}

func TestMetaCommand_SetTitle(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Copy sample to temp so we can modify it
	tmpPDF := filepath.Join(tmpDir, "test.pdf")
	input, _ := os.ReadFile(samplePDF())
	os.WriteFile(tmpPDF, input, 0644)

	// Set metadata
	if err := executeCommand("meta", tmpPDF, "--title", "Test Title"); err != nil {
		t.Fatalf("meta set command failed: %v", err)
	}
}
