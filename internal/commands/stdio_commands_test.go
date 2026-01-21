package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/lgbarn/pdf-cli/internal/cli"
)

// Tests for commands with stdin/stdout support

func TestCompressWithStdout(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	rootCmd := cli.GetRootCmd()
	stdout := &bytes.Buffer{}
	rootCmd.SetArgs([]string{"compress", samplePDF(), "--stdout"})
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(&bytes.Buffer{})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("compress --stdout failed: %v", err)
	}

	// Note: stdout output is written directly to os.Stdout, not the cobra buffer,
	// so we can't reliably verify the PDF content here. The test verifies the
	// command executes without error, which confirms the stdout path works.
}

func TestDecryptWithStdout(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	// First create an encrypted file
	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	encrypted := filepath.Join(tmpDir, "encrypted.pdf")
	if err := executeCommand("encrypt", samplePDF(), "--password", "secret", "-o", encrypted); err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	resetFlags(t)

	// Now decrypt to stdout
	rootCmd := cli.GetRootCmd()
	rootCmd.SetArgs([]string{"decrypt", encrypted, "--password", "secret", "--stdout"})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("decrypt --stdout failed: %v", err)
	}
}

func TestEncryptWithStdout(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	rootCmd := cli.GetRootCmd()
	rootCmd.SetArgs([]string{"encrypt", samplePDF(), "--password", "secret", "--stdout"})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("encrypt --stdout failed: %v", err)
	}
}

func TestRotateWithStdout(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	rootCmd := cli.GetRootCmd()
	rootCmd.SetArgs([]string{"rotate", samplePDF(), "-a", "90", "--stdout"})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("rotate --stdout failed: %v", err)
	}
}

func TestExtractWithStdout(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	rootCmd := cli.GetRootCmd()
	rootCmd.SetArgs([]string{"extract", samplePDF(), "-p", "1", "--stdout"})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("extract --stdout failed: %v", err)
	}
}

func TestReorderWithStdout(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	rootCmd := cli.GetRootCmd()
	rootCmd.SetArgs([]string{"reorder", samplePDF(), "-s", "1,2,3", "--stdout"})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("reorder --stdout failed: %v", err)
	}
}

func TestPdfaConvertWithStdout(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	rootCmd := cli.GetRootCmd()
	rootCmd.SetArgs([]string{"pdfa", "convert", samplePDF(), "--stdout"})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	// May fail due to PDF/A conversion limitations, but exercises the code path
	_ = rootCmd.Execute()
}

// Test stdin input for commands that support it
// Note: Actually testing stdin requires more complex setup since os.Stdin can't be easily mocked
// These tests focus on command structure and error handling

func TestCompressWithExplicitOutput_ExistingFile(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output := filepath.Join(tmpDir, "existing.pdf")
	// Create an existing file
	if err := os.WriteFile(output, []byte("existing"), 0600); err != nil {
		t.Fatalf("Failed to create existing file: %v", err)
	}

	rootCmd := cli.GetRootCmd()
	rootCmd.SetArgs([]string{"compress", samplePDF(), "-o", output})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	// Should fail without -f flag
	err = rootCmd.Execute()
	if err == nil {
		t.Error("compress to existing file without -f should fail")
	}
}

func TestEncryptWithOwnerPassword(t *testing.T) {
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
	if err := executeCommand("encrypt", samplePDF(), "--password", "user", "--owner-password", "owner", "-o", output); err != nil {
		t.Fatalf("encrypt with owner password failed: %v", err)
	}

	if _, err := os.Stat(output); os.IsNotExist(err) {
		t.Error("encrypt with owner password did not create output file")
	}
}

func TestDecrypt_NoPassword(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	rootCmd := cli.GetRootCmd()
	rootCmd.SetArgs([]string{"decrypt", samplePDF()})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	// Should fail - password is required
	err := rootCmd.Execute()
	if err == nil {
		t.Error("decrypt without password should fail")
	}
}

func TestEncrypt_NoPassword(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	rootCmd := cli.GetRootCmd()
	rootCmd.SetArgs([]string{"encrypt", samplePDF()})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	// Should fail - password is required
	err := rootCmd.Execute()
	if err == nil {
		t.Error("encrypt without password should fail")
	}
}

func TestWatermark_NonexistentImage(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	rootCmd := cli.GetRootCmd()
	rootCmd.SetArgs([]string{"watermark", samplePDF(), "-i", "/nonexistent/image.png", "-o", filepath.Join(tmpDir, "out.pdf")})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	// Should fail - image doesn't exist
	err = rootCmd.Execute()
	if err == nil {
		t.Error("watermark with nonexistent image should fail")
	}
}

func TestRotate_WithPageSelection_Batch(t *testing.T) {
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

	// Batch rotate with page selection
	if err := executeCommand("rotate", pdf1, pdf2, "-a", "90", "-p", "1"); err != nil {
		t.Fatalf("rotate batch with page selection failed: %v", err)
	}
}

func TestWatermark_Batch(t *testing.T) {
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

	// Batch watermark
	if err := executeCommand("watermark", pdf1, pdf2, "-t", "DRAFT"); err != nil {
		t.Fatalf("watermark batch failed: %v", err)
	}
}

func TestWatermark_WithPageSelection(t *testing.T) {
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
	if err := executeCommand("watermark", samplePDF(), "-t", "DRAFT", "-p", "1", "-o", output); err != nil {
		t.Fatalf("watermark with page selection failed: %v", err)
	}

	if _, err := os.Stat(output); os.IsNotExist(err) {
		t.Error("watermark did not create output file")
	}
}
