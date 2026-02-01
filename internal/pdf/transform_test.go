package pdf

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMerge(t *testing.T) {
	pdf := samplePDF()
	if _, err := os.Stat(pdf); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output := filepath.Join(tmpDir, "merged.pdf")
	err = Merge([]string{pdf, pdf}, output, "")
	if err != nil {
		t.Fatalf("Merge() error = %v", err)
	}

	// Verify output exists
	if _, err := os.Stat(output); os.IsNotExist(err) {
		t.Error("Merge() did not create output file")
	}

	// Verify merged file has double the pages
	origCount, _ := PageCount(pdf, "")
	mergedCount, err := PageCount(output, "")
	if err != nil {
		t.Fatalf("PageCount() on merged file error = %v", err)
	}
	if mergedCount != origCount*2 {
		t.Errorf("Merged PDF has %d pages, want %d", mergedCount, origCount*2)
	}
}

func TestMergeEmpty(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output := filepath.Join(tmpDir, "merged.pdf")

	// MergeWithProgress now returns an error for empty input
	err = Merge([]string{}, output, "")
	if err == nil {
		t.Error("Merge() expected error for empty input list")
	}
}

func TestMergeSingleFile(t *testing.T) {
	pdfFile := samplePDF()
	if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output := filepath.Join(tmpDir, "merged.pdf")
	err = Merge([]string{pdfFile}, output, "")
	if err != nil {
		t.Fatalf("Merge() single file error = %v", err)
	}

	// Verify output exists
	if _, err := os.Stat(output); os.IsNotExist(err) {
		t.Error("Merge() did not create output file")
	}
}

func TestMergeMultipleFiles(t *testing.T) {
	pdfFile := samplePDF()
	if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output := filepath.Join(tmpDir, "merged.pdf")
	// Merge 3 copies of the same file
	err = Merge([]string{pdfFile, pdfFile, pdfFile}, output, "")
	if err != nil {
		t.Fatalf("Merge() multiple files error = %v", err)
	}

	// Verify merged file has triple the pages
	origCount, _ := PageCount(pdfFile, "")
	mergedCount, err := PageCount(output, "")
	if err != nil {
		t.Fatalf("PageCount() error = %v", err)
	}
	if mergedCount != origCount*3 {
		t.Errorf("Merged PDF has %d pages, want %d", mergedCount, origCount*3)
	}
}

func TestMergeWithProgress(t *testing.T) {
	pdfFile := samplePDF()
	if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test with progress enabled but few files (should use standard merge)
	output := filepath.Join(tmpDir, "merged.pdf")
	err = MergeWithProgress([]string{pdfFile, pdfFile}, output, "", true)
	if err != nil {
		t.Fatalf("MergeWithProgress() error = %v", err)
	}

	if _, err := os.Stat(output); os.IsNotExist(err) {
		t.Error("MergeWithProgress() did not create output file")
	}
}

func TestMergeWithProgressManyFiles(t *testing.T) {
	pdfFile := samplePDF()
	if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a list of 5 files to trigger incremental merge path
	inputs := make([]string, 5)
	for i := 0; i < 5; i++ {
		inputs[i] = pdfFile
	}

	output := filepath.Join(tmpDir, "merged_many.pdf")
	err = MergeWithProgress(inputs, output, "", true)
	if err != nil {
		t.Fatalf("MergeWithProgress() with many files error = %v", err)
	}

	if _, err := os.Stat(output); os.IsNotExist(err) {
		t.Error("MergeWithProgress() did not create output file")
	}

	// Verify page count
	origCount, _ := PageCount(pdfFile, "")
	mergedCount, _ := PageCount(output, "")
	if mergedCount != origCount*5 {
		t.Errorf("Merged PDF has %d pages, want %d", mergedCount, origCount*5)
	}
}

func TestMergeWithProgressNonExistentFirst(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output := filepath.Join(tmpDir, "merged.pdf")
	inputs := []string{"/nonexistent/first.pdf", samplePDF()}

	err = MergeWithProgress(inputs, output, "", true)
	if err == nil {
		t.Error("MergeWithProgress() expected error for non-existent first file")
	}
}

func TestMergeWithProgressNoFlag(t *testing.T) {
	pdfFile := samplePDF()
	if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	inputs := make([]string, 5)
	for i := 0; i < 5; i++ {
		inputs[i] = pdfFile
	}

	output := filepath.Join(tmpDir, "merged.pdf")
	// showProgress=false should use standard merge even with many files
	err = MergeWithProgress(inputs, output, "", false)
	if err != nil {
		t.Fatalf("MergeWithProgress() without progress error = %v", err)
	}

	if _, err := os.Stat(output); os.IsNotExist(err) {
		t.Error("MergeWithProgress() did not create output file")
	}
}

func TestSplit(t *testing.T) {
	pdf := samplePDF()
	if _, err := os.Stat(pdf); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	err = Split(pdf, tmpDir, "")
	if err != nil {
		t.Fatalf("Split() error = %v", err)
	}

	// Verify at least one file was created
	files, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("ReadDir() error = %v", err)
	}

	pdfCount := 0
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".pdf") {
			pdfCount++
		}
	}

	if pdfCount == 0 {
		t.Error("Split() did not create any PDF files")
	}
}

func TestSplitNonExistent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	err = Split("/nonexistent/file.pdf", tmpDir, "")
	if err == nil {
		t.Error("Split() expected error for non-existent file")
	}
}

func TestSplitByPageCount(t *testing.T) {
	pdfFile := samplePDF()
	if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	err = SplitByPageCount(pdfFile, tmpDir, 2, "")
	if err != nil {
		t.Fatalf("SplitByPageCount() error = %v", err)
	}

	// Verify at least one file was created
	files, err := os.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("ReadDir() error = %v", err)
	}

	pdfCount := 0
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".pdf") {
			pdfCount++
		}
	}

	if pdfCount == 0 {
		t.Error("SplitByPageCount() did not create any PDF files")
	}
}

func TestSplitByPageCountEdges(t *testing.T) {
	pdfFile := samplePDF()
	if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	totalPages, err := PageCount(pdfFile, "")
	if err != nil {
		t.Fatalf("PageCount() error = %v", err)
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test with pageCount = 1 (split into individual pages)
	t.Run("pageCount=1", func(t *testing.T) {
		outDir := filepath.Join(tmpDir, "split1")
		os.MkdirAll(outDir, 0755)
		err := SplitByPageCount(pdfFile, outDir, 1, "")
		if err != nil {
			t.Fatalf("SplitByPageCount(1) error = %v", err)
		}

		files, _ := os.ReadDir(outDir)
		pdfCount := 0
		for _, f := range files {
			if strings.HasSuffix(f.Name(), ".pdf") {
				pdfCount++
			}
		}
		if pdfCount != totalPages {
			t.Errorf("SplitByPageCount(1) created %d files, want %d", pdfCount, totalPages)
		}
	})

	// Test with pageCount > total pages
	t.Run("pageCount>total", func(t *testing.T) {
		outDir := filepath.Join(tmpDir, "splithigh")
		os.MkdirAll(outDir, 0755)
		err := SplitByPageCount(pdfFile, outDir, totalPages+10, "")
		if err != nil {
			t.Fatalf("SplitByPageCount(high) error = %v", err)
		}

		// Should create just 1 file
		files, _ := os.ReadDir(outDir)
		pdfCount := 0
		for _, f := range files {
			if strings.HasSuffix(f.Name(), ".pdf") {
				pdfCount++
			}
		}
		if pdfCount != 1 {
			t.Errorf("SplitByPageCount(high) created %d files, want 1", pdfCount)
		}
	})
}

func TestSplitWithProgress(t *testing.T) {
	pdfFile := samplePDF()
	if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	err = SplitWithProgress(pdfFile, tmpDir, 1, "", true)
	if err != nil {
		t.Fatalf("SplitWithProgress() error = %v", err)
	}

	// Verify files were created
	files, _ := os.ReadDir(tmpDir)
	pdfCount := 0
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".pdf") {
			pdfCount++
		}
	}
	if pdfCount == 0 {
		t.Error("SplitWithProgress() did not create any files")
	}
}

func TestSplitWithProgressLargePDF(t *testing.T) {
	pdfFile := samplePDF()
	if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	// First, create a larger PDF by merging multiple copies
	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Merge 6 copies to create a PDF with more pages
	largePDF := filepath.Join(tmpDir, "large.pdf")
	inputs := make([]string, 6)
	for i := 0; i < 6; i++ {
		inputs[i] = pdfFile
	}
	err = Merge(inputs, largePDF, "")
	if err != nil {
		t.Fatalf("Merge() error = %v", err)
	}

	// Now test SplitWithProgress with showProgress=true
	splitDir := filepath.Join(tmpDir, "split")
	os.MkdirAll(splitDir, 0755)

	err = SplitWithProgress(largePDF, splitDir, 1, "", true)
	if err != nil {
		t.Fatalf("SplitWithProgress() error = %v", err)
	}

	// Verify files were created
	files, _ := os.ReadDir(splitDir)
	pdfCount := 0
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".pdf") {
			pdfCount++
		}
	}
	if pdfCount == 0 {
		t.Error("SplitWithProgress() did not create any files")
	}
}

func TestSplitWithProgressMultiPageChunks(t *testing.T) {
	pdfFile := samplePDF()
	if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	// Create a larger PDF
	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	largePDF := filepath.Join(tmpDir, "large.pdf")
	inputs := make([]string, 8)
	for i := 0; i < 8; i++ {
		inputs[i] = pdfFile
	}
	err = Merge(inputs, largePDF, "")
	if err != nil {
		t.Fatalf("Merge() error = %v", err)
	}

	// Split with 3-page chunks and progress
	splitDir := filepath.Join(tmpDir, "split")
	os.MkdirAll(splitDir, 0755)

	err = SplitWithProgress(largePDF, splitDir, 3, "", true)
	if err != nil {
		t.Fatalf("SplitWithProgress() with chunks error = %v", err)
	}

	// Verify files were created
	files, _ := os.ReadDir(splitDir)
	pdfCount := 0
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".pdf") {
			pdfCount++
		}
	}
	if pdfCount == 0 {
		t.Error("SplitWithProgress() did not create any files")
	}
}

func TestSplitWithProgressNonExistent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	err = SplitWithProgress("/nonexistent/file.pdf", tmpDir, 1, "", true)
	if err == nil {
		t.Error("SplitWithProgress() expected error for non-existent file")
	}
}
