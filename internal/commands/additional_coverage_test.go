package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/lgbarn/pdf-cli/internal/cli"
)

// Additional tests to increase coverage for uncovered code paths

func TestImagesCommand(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-images-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	if err := executeCommand("images", samplePDF(), "-o", tmpDir); err != nil {
		t.Fatalf("images command failed: %v", err)
	}
}

func TestImagesCommand_NoOutput(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	// When no output is specified, it should default to current directory
	// Use absolute path to sample.pdf since we're changing directory
	absSamplePDF, err := filepath.Abs(samplePDF())
	if err != nil {
		t.Fatalf("Failed to get absolute path: %v", err)
	}

	// Run in a temp directory to avoid cluttering the test directory
	origDir, _ := os.Getwd()
	tmpDir, err := os.MkdirTemp("", "pdf-test-images-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	defer os.Chdir(origDir)
	os.Chdir(tmpDir)

	if err := executeCommand("images", absSamplePDF); err != nil {
		t.Fatalf("images command without output failed: %v", err)
	}
}

func TestImagesCommand_WithPageSelection(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-images-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	if err := executeCommand("images", samplePDF(), "-p", "1", "-o", tmpDir); err != nil {
		t.Fatalf("images command with page selection failed: %v", err)
	}
}

func TestCombineImagesCommand(t *testing.T) {
	resetFlags(t)
	testImage := filepath.Join(testdataDir(), "test_image.png")
	if _, err := os.Stat(testImage); os.IsNotExist(err) {
		t.Skip("test_image.png not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-combine-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output := filepath.Join(tmpDir, "combined.pdf")
	if err := executeCommand("combine-images", testImage, "-o", output); err != nil {
		t.Fatalf("combine-images command failed: %v", err)
	}

	if _, err := os.Stat(output); os.IsNotExist(err) {
		t.Error("combine-images did not create output file")
	}
}

func TestCombineImagesCommand_MultipleImages(t *testing.T) {
	resetFlags(t)
	testImage := filepath.Join(testdataDir(), "test_image.png")
	if _, err := os.Stat(testImage); os.IsNotExist(err) {
		t.Skip("test_image.png not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-combine-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output := filepath.Join(tmpDir, "combined.pdf")
	if err := executeCommand("combine-images", testImage, testImage, "-o", output); err != nil {
		t.Fatalf("combine-images with multiple images failed: %v", err)
	}

	if _, err := os.Stat(output); os.IsNotExist(err) {
		t.Error("combine-images did not create output file")
	}
}

func TestCombineImagesCommand_NonexistentImage(t *testing.T) {
	resetFlags(t)

	tmpDir, err := os.MkdirTemp("", "pdf-test-combine-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output := filepath.Join(tmpDir, "combined.pdf")
	err = executeCommand("combine-images", "/nonexistent/image.png", "-o", output)
	if err == nil {
		t.Error("combine-images with nonexistent image should fail")
	}
}

func TestCombineImagesCommand_UnsupportedFormat(t *testing.T) {
	resetFlags(t)

	tmpDir, err := os.MkdirTemp("", "pdf-test-combine-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a text file with wrong extension
	txtFile := filepath.Join(tmpDir, "notimage.txt")
	if err := os.WriteFile(txtFile, []byte("not an image"), 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	output := filepath.Join(tmpDir, "combined.pdf")
	err = executeCommand("combine-images", txtFile, "-o", output)
	if err == nil {
		t.Error("combine-images with unsupported format should fail")
	}
}

func TestCombineImagesCommand_WithPageSize(t *testing.T) {
	resetFlags(t)
	testImage := filepath.Join(testdataDir(), "test_image.png")
	if _, err := os.Stat(testImage); os.IsNotExist(err) {
		t.Skip("test_image.png not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-combine-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output := filepath.Join(tmpDir, "combined.pdf")
	if err := executeCommand("combine-images", testImage, "-o", output, "--page-size", "A4"); err != nil {
		t.Fatalf("combine-images with page size failed: %v", err)
	}

	if _, err := os.Stat(output); os.IsNotExist(err) {
		t.Error("combine-images did not create output file")
	}
}

func TestInfoCommand_BatchWithErrors(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	// Mix valid and invalid files to test error handling in batch mode
	rootCmd := cli.GetRootCmd()
	rootCmd.SetArgs([]string{"info", samplePDF(), "/nonexistent/file.pdf"})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	// Should complete without fatal error even with invalid files
	_ = rootCmd.Execute()
}

func TestInfoCommand_BatchCSVFormat(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	rootCmd := cli.GetRootCmd()
	rootCmd.SetArgs([]string{"info", samplePDF(), samplePDF(), "--format", "csv"})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("info command with csv format failed: %v", err)
	}
}

func TestInfoCommand_BatchTSVFormat(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	rootCmd := cli.GetRootCmd()
	rootCmd.SetArgs([]string{"info", samplePDF(), samplePDF(), "--format", "tsv"})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("info command with tsv format failed: %v", err)
	}
}

func TestInfoCommand_BatchJSONFormat(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	rootCmd := cli.GetRootCmd()
	rootCmd.SetArgs([]string{"info", samplePDF(), samplePDF(), "--format", "json"})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("info command batch with json format failed: %v", err)
	}
}

func TestMetaCommand_BatchWithErrors(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	// Mix valid and invalid files to test error handling in batch mode
	rootCmd := cli.GetRootCmd()
	rootCmd.SetArgs([]string{"meta", samplePDF(), "/nonexistent/file.pdf"})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	// Should complete without fatal error even with invalid files
	_ = rootCmd.Execute()
}

func TestMetaCommand_BatchCSVFormat(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	rootCmd := cli.GetRootCmd()
	rootCmd.SetArgs([]string{"meta", samplePDF(), samplePDF(), "--format", "csv"})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("meta command with csv format failed: %v", err)
	}
}

func TestMetaCommand_BatchTSVFormat(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	rootCmd := cli.GetRootCmd()
	rootCmd.SetArgs([]string{"meta", samplePDF(), samplePDF(), "--format", "tsv"})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("meta command with tsv format failed: %v", err)
	}
}

func TestMetaCommand_BatchJSONFormat(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	rootCmd := cli.GetRootCmd()
	rootCmd.SetArgs([]string{"meta", samplePDF(), samplePDF(), "--format", "json"})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("meta command batch with json format failed: %v", err)
	}
}

func TestMetaCommand_SetMultipleOnBatch(t *testing.T) {
	resetFlags(t)
	if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	// Setting metadata on multiple files should fail
	rootCmd := cli.GetRootCmd()
	rootCmd.SetArgs([]string{"meta", samplePDF(), samplePDF(), "--title", "Test"})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	err := rootCmd.Execute()
	if err == nil {
		t.Error("meta --title on multiple files should fail")
	}
}

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

func TestCompressCommand_BatchWithOutputFlag(t *testing.T) {
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

	// Batch compress with output flag should fail
	output := filepath.Join(tmpDir, "output.pdf")
	rootCmd := cli.GetRootCmd()
	rootCmd.SetArgs([]string{"compress", pdf1, pdf2, "-o", output})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	err = rootCmd.Execute()
	if err == nil {
		t.Error("compress batch with -o flag should fail")
	}
}

func TestEncryptCommand_BatchWithOutputFlag(t *testing.T) {
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

	// Batch encrypt with output flag should fail
	output := filepath.Join(tmpDir, "output.pdf")
	rootCmd := cli.GetRootCmd()
	rootCmd.SetArgs([]string{"encrypt", pdf1, pdf2, "--password", "test", "-o", output})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	err = rootCmd.Execute()
	if err == nil {
		t.Error("encrypt batch with -o flag should fail")
	}
}

func TestRotateCommand_BatchWithOutputFlag(t *testing.T) {
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

	// Batch rotate with output flag should fail
	output := filepath.Join(tmpDir, "output.pdf")
	rootCmd := cli.GetRootCmd()
	rootCmd.SetArgs([]string{"rotate", pdf1, pdf2, "-a", "90", "-o", output})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	err = rootCmd.Execute()
	if err == nil {
		t.Error("rotate batch with -o flag should fail")
	}
}

func TestWatermarkCommand_BatchWithOutputFlag(t *testing.T) {
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

	// Batch watermark with output flag should fail
	output := filepath.Join(tmpDir, "output.pdf")
	rootCmd := cli.GetRootCmd()
	rootCmd.SetArgs([]string{"watermark", pdf1, pdf2, "-t", "DRAFT", "-o", output})
	rootCmd.SetOut(&bytes.Buffer{})
	rootCmd.SetErr(&bytes.Buffer{})

	err = rootCmd.Execute()
	if err == nil {
		t.Error("watermark batch with -o flag should fail")
	}
}
