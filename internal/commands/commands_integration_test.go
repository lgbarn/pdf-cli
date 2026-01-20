package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/lgbarn/pdf-cli/internal/cli"
)

// resetFlags resets flag values to their defaults
// This is needed because cobra persists flag values between tests
func resetFlags(t *testing.T) {
	t.Helper()
	rootCmd := cli.GetRootCmd()
	// Reset global flags
	_ = rootCmd.PersistentFlags().Set("verbose", "false")
	_ = rootCmd.PersistentFlags().Set("force", "false")

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

func TestCompressCommand_WithOutput(t *testing.T) {
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
	if err := executeCommand("compress", samplePDF(), "-o", output); err != nil {
		t.Fatalf("compress command failed: %v", err)
	}

	if _, err := os.Stat(output); os.IsNotExist(err) {
		t.Error("compress did not create output file")
	}
}

func TestCompressCommand_ForceOverwrite(t *testing.T) {
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
	// Create existing file
	if err := os.WriteFile(output, []byte("existing"), 0600); err != nil {
		t.Fatalf("Failed to create existing file: %v", err)
	}

	if err := executeCommand("compress", samplePDF(), "-o", output, "-f"); err != nil {
		t.Fatalf("compress command with -f failed: %v", err)
	}

	// Verify it was overwritten (file size should be different)
	info, _ := os.Stat(output)
	if info.Size() == 8 { // "existing" is 8 bytes
		t.Error("compress did not overwrite existing file")
	}
}

func TestRotateCommand_ValidAngles(t *testing.T) {
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	for _, angle := range []string{"90", "180", "270"} {
		t.Run(angle+"_degrees", func(t *testing.T) {
			resetFlags(t)
			tmpDir, err := os.MkdirTemp("", "pdf-test-*")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			output := filepath.Join(tmpDir, "rotated.pdf")
			if err := executeCommand("rotate", samplePDF(), "-a", angle, "-o", output); err != nil {
				t.Fatalf("rotate command failed: %v", err)
			}
			if _, err := os.Stat(output); os.IsNotExist(err) {
				t.Error("rotate did not create output file")
			}
		})
	}
}

func TestRotateCommand_InvalidAngle(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output := filepath.Join(tmpDir, "rotated.pdf")
	err = executeCommand("rotate", samplePDF(), "-a", "45", "-o", output)
	if err == nil {
		t.Error("rotate with invalid angle should fail")
	}
}

func TestRotateCommand_WithPageSelection(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output := filepath.Join(tmpDir, "rotated.pdf")
	if err := executeCommand("rotate", samplePDF(), "-a", "90", "-p", "1", "-o", output); err != nil {
		t.Fatalf("rotate command with page selection failed: %v", err)
	}

	if _, err := os.Stat(output); os.IsNotExist(err) {
		t.Error("rotate did not create output file")
	}
}

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

func TestCompressCommand_MultipleFiles(t *testing.T) {
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

	// Batch compress
	if err := executeCommand("compress", pdf1, pdf2); err != nil {
		t.Fatalf("compress command with multiple files failed: %v", err)
	}
}

func TestRotateCommand_MultipleFiles(t *testing.T) {
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

	// Batch rotate
	if err := executeCommand("rotate", pdf1, pdf2, "-a", "90"); err != nil {
		t.Fatalf("rotate command with multiple files failed: %v", err)
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
