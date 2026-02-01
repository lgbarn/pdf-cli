package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/lgbarn/pdf-cli/internal/cli"
)

func TestEncryptCommand_WithPassword(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output := filepath.Join(tmpDir, "encrypted.pdf")
	if err := executeCommand("encrypt", samplePDF(), "--password", "secret123", "-o", output); err != nil {
		t.Fatalf("encrypt command failed: %v", err)
	}

	if _, err := os.Stat(output); os.IsNotExist(err) {
		t.Error("encrypt did not create output file")
	}
}

func TestEncryptDecryptCommand_RoundTrip(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	encrypted := filepath.Join(tmpDir, "encrypted.pdf")
	decrypted := filepath.Join(tmpDir, "decrypted.pdf")
	password := "testpassword"

	// Encrypt
	if err := executeCommand("encrypt", samplePDF(), "--password", password, "-o", encrypted); err != nil {
		t.Fatalf("encrypt command failed: %v", err)
	}

	// Reset flags before decrypt
	resetFlags(t)

	// Decrypt
	if err := executeCommand("decrypt", encrypted, "--password", password, "-o", decrypted); err != nil {
		t.Fatalf("decrypt command failed: %v", err)
	}

	if _, err := os.Stat(decrypted); os.IsNotExist(err) {
		t.Error("decrypt did not create output file")
	}
}
func TestWatermarkCommand_TextWatermark(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output := filepath.Join(tmpDir, "watermarked.pdf")
	if err := executeCommand("watermark", samplePDF(), "-t", "CONFIDENTIAL", "-o", output); err != nil {
		t.Fatalf("watermark command failed: %v", err)
	}

	if _, err := os.Stat(output); os.IsNotExist(err) {
		t.Error("watermark did not create output file")
	}
}

func TestWatermarkCommand_ImageWatermark(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	testImage := filepath.Join(testdataDir(), "test_image.png")
	if _, err := os.Stat(testImage); os.IsNotExist(err) {
		t.Skip("test_image.png not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output := filepath.Join(tmpDir, "watermarked.pdf")
	if err := executeCommand("watermark", samplePDF(), "-i", testImage, "-o", output); err != nil {
		t.Fatalf("watermark command failed: %v", err)
	}

	if _, err := os.Stat(output); os.IsNotExist(err) {
		t.Error("watermark did not create output file")
	}
}

func TestWatermarkCommand_NoWatermarkSpecified(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output := filepath.Join(tmpDir, "watermarked.pdf")
	err = executeCommand("watermark", samplePDF(), "-o", output)
	if err == nil {
		t.Error("watermark without text or image should fail")
	}
}

func TestWatermarkCommand_BothTextAndImage(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	testImage := filepath.Join(testdataDir(), "test_image.png")
	if _, err := os.Stat(testImage); os.IsNotExist(err) {
		t.Skip("test_image.png not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output := filepath.Join(tmpDir, "watermarked.pdf")
	err = executeCommand("watermark", samplePDF(), "-t", "TEST", "-i", testImage, "-o", output)
	if err == nil {
		t.Error("watermark with both text and image should fail")
	}
}
func TestMergeCommand_TwoFiles(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output := filepath.Join(tmpDir, "merged.pdf")
	if err := executeCommand("merge", "-o", output, samplePDF(), samplePDF()); err != nil {
		t.Fatalf("merge command failed: %v", err)
	}

	if _, err := os.Stat(output); os.IsNotExist(err) {
		t.Error("merge did not create output file")
	}
}
func TestSplitCommand_SinglePageChunks(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-split-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Run command directly, resetting args explicitly
	rootCmd := cli.GetRootCmd()
	rootCmd.SetArgs([]string{"split", samplePDF(), "-o", tmpDir})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("split command failed: %v", err)
	}

	// Verify at least one PDF was created
	files, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("ReadDir error: %v", err)
	}

	pdfCount := 0
	for _, f := range files {
		if filepath.Ext(f.Name()) == ".pdf" {
			pdfCount++
		}
	}

	if pdfCount == 0 {
		// Debug: list what's in the directory
		t.Logf("Directory %s contents:", tmpDir)
		for _, f := range files {
			t.Logf("  - %s (isDir: %v)", f.Name(), f.IsDir())
		}
		t.Logf("Expected split files in: %s", tmpDir)
		t.Error("split did not create any PDF files")
	}
}

func TestSplitCommand_MultiPageChunks(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	if err := executeCommand("split", samplePDF(), "-o", tmpDir, "-n", "2"); err != nil {
		t.Fatalf("split command failed: %v", err)
	}

	// Verify at least one PDF was created (may be in subdirectories)
	pdfCount := 0
	err = filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".pdf" {
			pdfCount++
		}
		return nil
	})
	if err != nil {
		t.Fatalf("filepath.Walk error: %v", err)
	}

	if pdfCount == 0 {
		t.Error("split did not create any PDF files")
	}
}
func TestReorderCommand(t *testing.T) {
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
	if err := executeCommand("reorder", samplePDF(), "-s", "3,2,1", "-o", output); err != nil {
		t.Fatalf("reorder command failed: %v", err)
	}

	if _, err := os.Stat(output); os.IsNotExist(err) {
		t.Error("reorder did not create output file")
	}
}

func TestReorderCommand_InvalidSequence(t *testing.T) {
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
	err = executeCommand("reorder", samplePDF(), "-s", "abc", "-o", output)
	if err == nil {
		t.Error("reorder with invalid sequence should fail")
	}
}

func TestReorderCommand_OutOfRangeSequence(t *testing.T) {
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
	err = executeCommand("reorder", samplePDF(), "-s", "1,2,100", "-o", output)
	if err == nil {
		t.Error("reorder with out-of-range sequence should fail")
	}
}
func TestEncryptCommand_MultipleFiles(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Copy sample.pdf twice
	pdf1 := filepath.Join(tmpDir, "test1.pdf")
	pdf2 := filepath.Join(tmpDir, "test2.pdf")
	input, _ := os.ReadFile(samplePDF())
	os.WriteFile(pdf1, input, 0644)
	os.WriteFile(pdf2, input, 0644)

	// Batch encrypt
	if err := executeCommand("encrypt", pdf1, pdf2, "--password", "secret"); err != nil {
		t.Fatalf("encrypt command with multiple files failed: %v", err)
	}
}

func TestDecryptCommand_MultipleFiles(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Copy sample.pdf twice and encrypt them
	pdf1 := filepath.Join(tmpDir, "test1.pdf")
	pdf2 := filepath.Join(tmpDir, "test2.pdf")
	input, _ := os.ReadFile(samplePDF())
	os.WriteFile(pdf1, input, 0644)
	os.WriteFile(pdf2, input, 0644)

	// First encrypt both
	executeCommand("encrypt", pdf1, "--password", "secret")
	resetFlags(t)
	executeCommand("encrypt", pdf2, "--password", "secret")
	resetFlags(t)

	// Batch decrypt
	encrypted1 := filepath.Join(tmpDir, "test1_encrypted.pdf")
	encrypted2 := filepath.Join(tmpDir, "test2_encrypted.pdf")
	if _, err := os.Stat(encrypted1); err == nil {
		_ = executeCommand("decrypt", encrypted1, encrypted2, "--password", "secret")
	}
}

func TestExtractCommand_MultiplePages(t *testing.T) {
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
	if err := executeCommand("extract", samplePDF(), "-p", "1,2", "-o", output); err != nil {
		t.Fatalf("extract command failed: %v", err)
	}

	if _, err := os.Stat(output); os.IsNotExist(err) {
		t.Error("extract did not create output file")
	}
}
func TestCommands_NonExistentFile(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"info", []string{"info", "/nonexistent/file.pdf"}},
		{"text", []string{"text", "/nonexistent/file.pdf"}},
		{"compress", []string{"compress", "/nonexistent/file.pdf"}},
		{"rotate", []string{"rotate", "/nonexistent/file.pdf", "-a", "90"}},
		{"extract", []string{"extract", "/nonexistent/file.pdf", "-p", "1"}},
		{"decrypt", []string{"decrypt", "/nonexistent/file.pdf", "--password", "test"}},
		{"meta", []string{"meta", "/nonexistent/file.pdf"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetFlags(t)
			if err := executeCommand(tt.args...); err == nil {
				t.Errorf("%s with non-existent file should fail", tt.name)
			}
		})
	}
}
func TestMergeCommand_SingleFile(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output := filepath.Join(tmpDir, "merged.pdf")
	_ = executeCommand("merge", "-o", output, samplePDF())
}
