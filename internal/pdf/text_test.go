package pdf

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/ledongthuc/pdf"
)

func TestExtractText(t *testing.T) {
	pdf := samplePDF()
	if _, err := os.Stat(pdf); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	text, err := ExtractText(context.Background(), pdf, nil, "")
	if err != nil {
		t.Fatalf("ExtractText() error = %v", err)
	}

	// The sample.pdf should have some text content
	if text == "" {
		t.Log("Warning: ExtractText() returned empty string (PDF may be image-based)")
	}
}

func TestExtractTextWithPages(t *testing.T) {
	pdf := samplePDF()
	if _, err := os.Stat(pdf); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	// Get page count first
	count, err := PageCount(pdf, "")
	if err != nil {
		t.Fatalf("PageCount() error = %v", err)
	}

	if count > 0 {
		text, err := ExtractText(context.Background(), pdf, []int{1}, "")
		if err != nil {
			t.Fatalf("ExtractText() with pages error = %v", err)
		}
		// Just verify it doesn't crash
		_ = text
	}
}

func TestExtractTextNonExistent(t *testing.T) {
	_, err := ExtractText(context.Background(), "/nonexistent/file.pdf", nil, "")
	if err == nil {
		t.Error("ExtractText() expected error for non-existent file")
	}
}

func TestExtractPagesSequentialDirect(t *testing.T) {
	pdfFile := samplePDF()
	if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	f, r, err := pdf.Open(pdfFile)
	if err != nil {
		// Some simple PDFs may not be parseable by this library
		t.Skipf("PDF not parseable by ledongthuc/pdf: %v", err)
	}
	defer f.Close()

	totalPages := r.NumPage()
	if totalPages < 1 {
		t.Skip("PDF has no pages")
	}

	pages := []int{1}
	text, err := extractPagesSequential(context.Background(), r, pages, totalPages, false)
	if err != nil {
		t.Fatalf("extractPagesSequential() error = %v", err)
	}

	// Just verify it returns without error, text might be empty
	_ = text
}

func TestExtractPagesSequentialMultiple(t *testing.T) {
	pdfFile := samplePDF()
	if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	f, r, err := pdf.Open(pdfFile)
	if err != nil {
		t.Skipf("PDF not parseable by ledongthuc/pdf: %v", err)
	}
	defer f.Close()

	totalPages := r.NumPage()
	if totalPages < 1 {
		t.Skip("PDF has no pages")
	}

	// Request only valid pages
	pages := []int{1}
	if totalPages >= 2 {
		pages = []int{1, 2}
	}

	text, err := extractPagesSequential(context.Background(), r, pages, totalPages, false)
	if err != nil {
		t.Fatalf("extractPagesSequential() multiple pages error = %v", err)
	}
	_ = text
}

func TestExtractPagesSequentialOutOfRange(t *testing.T) {
	pdfFile := samplePDF()
	if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	f, r, err := pdf.Open(pdfFile)
	if err != nil {
		t.Skipf("PDF not parseable by ledongthuc/pdf: %v", err)
	}
	defer f.Close()

	totalPages := r.NumPage()
	// Request out-of-range pages - should be skipped, not error
	pages := []int{9999}

	text, err := extractPagesSequential(context.Background(), r, pages, totalPages, false)
	if err != nil {
		t.Fatalf("extractPagesSequential() out of range error = %v", err)
	}

	// Should return empty for out-of-range pages
	if text != "" {
		t.Log("Warning: extractPagesSequential returned non-empty for out-of-range page")
	}
}

func TestExtractPagesParallelDirect(t *testing.T) {
	pdfFile := samplePDF()
	if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	f, r, err := pdf.Open(pdfFile)
	if err != nil {
		t.Skipf("PDF not parseable by ledongthuc/pdf: %v", err)
	}
	defer f.Close()

	totalPages := r.NumPage()
	if totalPages < 1 {
		t.Skip("PDF has no pages")
	}

	pages := []int{1}
	text, err := extractPagesParallel(context.Background(), r, pages, totalPages, false)
	if err != nil {
		t.Fatalf("extractPagesParallel() error = %v", err)
	}
	_ = text
}

func TestExtractPagesParallelMultiple(t *testing.T) {
	pdfFile := samplePDF()
	if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	f, r, err := pdf.Open(pdfFile)
	if err != nil {
		t.Skipf("PDF not parseable by ledongthuc/pdf: %v", err)
	}
	defer f.Close()

	totalPages := r.NumPage()
	if totalPages < 1 {
		t.Skip("PDF has no pages")
	}

	// Request only valid pages
	pages := []int{1}
	if totalPages >= 2 {
		pages = []int{1, 2}
	}

	text, err := extractPagesParallel(context.Background(), r, pages, totalPages, false)
	if err != nil {
		t.Fatalf("extractPagesParallel() multiple pages error = %v", err)
	}
	_ = text
}

func TestExtractPagesParallelOutOfRange(t *testing.T) {
	pdfFile := samplePDF()
	if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	f, r, err := pdf.Open(pdfFile)
	if err != nil {
		t.Skipf("PDF not parseable by ledongthuc/pdf: %v", err)
	}
	defer f.Close()

	totalPages := r.NumPage()
	// Request out-of-range pages - should return empty, not error
	pages := []int{9999}

	text, err := extractPagesParallel(context.Background(), r, pages, totalPages, false)
	if err != nil {
		t.Fatalf("extractPagesParallel() out of range error = %v", err)
	}

	// Should return empty for out-of-range pages
	if text != "" {
		t.Log("Warning: extractPagesParallel returned non-empty for out-of-range page")
	}
}

func TestExtractTextWithProgress(t *testing.T) {
	pdfFile := samplePDF()
	if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	// Test with progress enabled
	text, err := ExtractTextWithProgress(context.Background(), pdfFile, nil, "", true)
	if err != nil {
		t.Fatalf("ExtractTextWithProgress() error = %v", err)
	}
	// Just verify it returns without error
	_ = text
}

func TestExtractTextPrimaryWithSpecificPages(t *testing.T) {
	pdfFile := samplePDF()
	if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	// Get page count first
	count, err := PageCount(pdfFile, "")
	if err != nil {
		t.Fatalf("PageCount() error = %v", err)
	}

	if count < 1 {
		t.Skip("PDF has no pages")
	}

	// Test with unsorted pages to trigger sort path
	pages := []int{1}
	if count >= 2 {
		pages = []int{2, 1} // Unsorted to trigger sort
	}

	text, err := ExtractTextWithProgress(context.Background(), pdfFile, pages, "", false)
	if err != nil {
		t.Fatalf("ExtractTextWithProgress() with specific pages error = %v", err)
	}
	_ = text
}

func TestExtractPageTextInvalidPages(t *testing.T) {
	pdfFile := samplePDF()
	if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	f, r, err := pdf.Open(pdfFile)
	if err != nil {
		t.Skipf("PDF not parseable by ledongthuc/pdf: %v", err)
	}
	defer f.Close()

	totalPages := r.NumPage()

	// Test with page number < 1
	text := extractPageText(r, 0, totalPages)
	if text != "" {
		t.Error("extractPageText() should return empty for page 0")
	}

	// Test with page number < 1
	text = extractPageText(r, -1, totalPages)
	if text != "" {
		t.Error("extractPageText() should return empty for negative page")
	}
}

func TestExtractTextFallbackPath(t *testing.T) {
	// Create an empty temp dir to simulate no text extraction
	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test with non-existent file to trigger error path
	_, err = ExtractText(context.Background(), "/nonexistent/file.pdf", nil, "")
	if err == nil {
		t.Error("ExtractText() expected error for non-existent file")
	}
}

func TestExtractTextWithProgressManyPages(t *testing.T) {
	pdfFile := samplePDF()
	if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	// Create a large PDF for testing parallel extraction
	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Merge multiple copies to get more than 5 pages
	largePDF := filepath.Join(tmpDir, "large.pdf")
	inputs := make([]string, 6)
	for i := 0; i < 6; i++ {
		inputs[i] = pdfFile
	}
	err = Merge(inputs, largePDF, "")
	if err != nil {
		t.Fatalf("Merge() error = %v", err)
	}

	// Test extraction on larger PDF (should trigger parallel extraction)
	text, err := ExtractTextWithProgress(context.Background(), largePDF, nil, "", true)
	if err != nil {
		t.Fatalf("ExtractTextWithProgress() error = %v", err)
	}
	_ = text
}

func TestExtractPagesSequentialWithProgress(t *testing.T) {
	pdfFile := samplePDF()
	if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	f, r, err := pdf.Open(pdfFile)
	if err != nil {
		t.Skipf("PDF not parseable: %v", err)
	}
	defer f.Close()

	totalPages := r.NumPage()
	if totalPages < 1 {
		t.Skip("PDF has no pages")
	}

	pages := []int{1}
	// Test with showProgress=true
	text, err := extractPagesSequential(context.Background(), r, pages, totalPages, true)
	if err != nil {
		t.Fatalf("extractPagesSequential() with progress error = %v", err)
	}
	_ = text
}

func TestExtractPagesParallelWithProgress(t *testing.T) {
	pdfFile := samplePDF()
	if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	f, r, err := pdf.Open(pdfFile)
	if err != nil {
		t.Skipf("PDF not parseable: %v", err)
	}
	defer f.Close()

	totalPages := r.NumPage()
	if totalPages < 1 {
		t.Skip("PDF has no pages")
	}

	pages := []int{1}
	// Test with showProgress=true
	text, err := extractPagesParallel(context.Background(), r, pages, totalPages, true)
	if err != nil {
		t.Fatalf("extractPagesParallel() with progress error = %v", err)
	}
	_ = text
}
