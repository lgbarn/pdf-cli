package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/lgbarn/pdf-cli/internal/cli"
)

func TestMerge_NonexistentFiles(t *testing.T) {
	resetFlags(t)

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output := filepath.Join(tmpDir, "merged.pdf")
	err = executeCommand("merge", "/nonexistent/a.pdf", "/nonexistent/b.pdf", "-o", output)
	if err == nil {
		t.Error("merge with nonexistent files should fail")
	}
}

func TestSplit_InvalidPagesPerFile(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test with zero pages per file (should probably fail or default to 1)
	rootCmd := cli.GetRootCmd()
	rootCmd.SetArgs([]string{"split", samplePDF(), "-o", tmpDir, "-n", "0"})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	_ = rootCmd.Execute()
}
func TestTextCommand_WithOCR(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	rootCmd := cli.GetRootCmd()
	rootCmd.SetArgs([]string{"text", samplePDF(), "--ocr"})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	// May fail if tesseract is not installed, but exercises the code path
	_ = rootCmd.Execute()
}

func TestTextCommand_WithOCRLang(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	rootCmd := cli.GetRootCmd()
	rootCmd.SetArgs([]string{"text", samplePDF(), "--ocr", "--ocr-lang", "eng"})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	// May fail if tesseract is not installed, but exercises the code path
	_ = rootCmd.Execute()
}

func TestTextCommand_WithOCRBackend(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	rootCmd := cli.GetRootCmd()
	rootCmd.SetArgs([]string{"text", samplePDF(), "--ocr", "--ocr-backend", "wasm"})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	// May fail if tesseract is not available, but exercises the code path
	_ = rootCmd.Execute()
}
func TestPdfaValidate_JSONFormat(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	rootCmd := cli.GetRootCmd()
	rootCmd.SetArgs([]string{"pdfa", "validate", samplePDF(), "--format", "json"})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	// Just verify it runs
	_ = rootCmd.Execute()
}
func TestCompressCommand_FileSizeIncrease(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output := filepath.Join(tmpDir, "compressed.pdf")
	// Compress an already small/optimized file - may result in size increase
	// This tests the "file size increased" output path
	if err := executeCommand("compress", samplePDF(), "-o", output); err != nil {
		t.Fatalf("compress command failed: %v", err)
	}
}
func TestReorderCommand_EndKeyword(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output := filepath.Join(tmpDir, "reordered.pdf")
	// Test with 'end' keyword
	if err := executeCommand("reorder", samplePDF(), "-s", "end-1", "-o", output); err != nil {
		t.Fatalf("reorder with end keyword failed: %v", err)
	}

	if _, err := os.Stat(output); os.IsNotExist(err) {
		t.Error("reorder did not create output file")
	}
}
func TestExtract_NoPages(t *testing.T) {
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
	// Extract without pages flag should fail (pages is required)
	rootCmd := cli.GetRootCmd()
	rootCmd.SetArgs([]string{"extract", samplePDF(), "-o", output})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	err = rootCmd.Execute()
	if err == nil {
		t.Error("extract without pages should fail")
	}
}
