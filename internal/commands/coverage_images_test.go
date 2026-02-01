package commands

import (
	"os"
	"path/filepath"
	"testing"
)

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
