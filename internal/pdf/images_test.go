package pdf

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreatePDFFromImages(t *testing.T) {
	testImage := filepath.Join(testdataDir(), "test_image.png")
	if _, err := os.Stat(testImage); os.IsNotExist(err) {
		t.Skip("test_image.png not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output := filepath.Join(tmpDir, "from_images.pdf")
	err = CreatePDFFromImages([]string{testImage}, output, "")
	if err != nil {
		t.Fatalf("CreatePDFFromImages() error = %v", err)
	}

	if _, err := os.Stat(output); os.IsNotExist(err) {
		t.Error("CreatePDFFromImages() did not create output file")
	}
}

func TestCreatePDFFromImagesMultiple(t *testing.T) {
	testImage := filepath.Join(testdataDir(), "test_image.png")
	if _, err := os.Stat(testImage); os.IsNotExist(err) {
		t.Skip("test_image.png not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Use the same image multiple times
	output := filepath.Join(tmpDir, "multi_images.pdf")
	err = CreatePDFFromImages([]string{testImage, testImage}, output, "")
	if err != nil {
		t.Fatalf("CreatePDFFromImages() error = %v", err)
	}

	// Verify output has multiple pages
	count, err := PageCount(output, "")
	if err != nil {
		t.Fatalf("PageCount() error = %v", err)
	}
	if count < 2 {
		t.Errorf("CreatePDFFromImages() pages = %d, want >= 2", count)
	}
}

func TestCreatePDFFromImagesPageSize(t *testing.T) {
	testImage := filepath.Join(testdataDir(), "test_image.png")
	if _, err := os.Stat(testImage); os.IsNotExist(err) {
		t.Skip("test_image.png not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	pageSizes := []string{"A4", "Letter", "LETTER", "a4"}

	for _, size := range pageSizes {
		t.Run(size, func(t *testing.T) {
			output := filepath.Join(tmpDir, "pagesize_"+size+".pdf")
			err := CreatePDFFromImages([]string{testImage}, output, size)
			if err != nil {
				t.Fatalf("CreatePDFFromImages() with pageSize %q error = %v", size, err)
			}
			if _, err := os.Stat(output); os.IsNotExist(err) {
				t.Errorf("CreatePDFFromImages() did not create output file")
			}
		})
	}
}

func TestCreatePDFFromImagesMissing(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output := filepath.Join(tmpDir, "from_images.pdf")
	err = CreatePDFFromImages([]string{"/nonexistent/image.png"}, output, "")
	if err == nil {
		t.Error("CreatePDFFromImages() expected error for missing image")
	}
}

func TestCreatePDFFromImagesDefaultPageSize(t *testing.T) {
	testImage := filepath.Join(testdataDir(), "test_image.png")
	if _, err := os.Stat(testImage); os.IsNotExist(err) {
		t.Skip("test_image.png not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test with custom page size (not A4 or Letter)
	output := filepath.Join(tmpDir, "custom_size.pdf")
	err = CreatePDFFromImages([]string{testImage}, output, "Legal")
	if err != nil {
		t.Fatalf("CreatePDFFromImages() with custom size error = %v", err)
	}

	if _, err := os.Stat(output); os.IsNotExist(err) {
		t.Error("CreatePDFFromImages() did not create output file")
	}
}

func TestExtractImagesNonExistent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	err = ExtractImages("/nonexistent/file.pdf", tmpDir, nil, "")
	if err == nil {
		t.Error("ExtractImages() expected error for non-existent file")
	}
}

func TestExtractImagesValid(t *testing.T) {
	pdfFile := samplePDF()
	if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// This may or may not find images depending on the PDF
	err = ExtractImages(pdfFile, tmpDir, nil, "")
	// Just verify it doesn't panic
	_ = err
}

func TestExtractImagesWithPages(t *testing.T) {
	pdfFile := samplePDF()
	if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Extract images from specific pages
	err = ExtractImages(pdfFile, tmpDir, []int{1}, "")
	// This may succeed or fail depending on the PDF content
	_ = err
}
