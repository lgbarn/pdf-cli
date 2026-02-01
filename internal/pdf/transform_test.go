package pdf

import (
	"fmt"
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

func TestExtractPages(t *testing.T) {
	pdf := samplePDF()
	if _, err := os.Stat(pdf); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output := filepath.Join(tmpDir, "extracted.pdf")
	err = ExtractPages(pdf, output, []int{1}, "")
	if err != nil {
		t.Fatalf("ExtractPages() error = %v", err)
	}

	// Verify output exists
	if _, err := os.Stat(output); os.IsNotExist(err) {
		t.Error("ExtractPages() did not create output file")
	}

	// Verify extracted file has 1 page
	count, err := PageCount(output, "")
	if err != nil {
		t.Fatalf("PageCount() error = %v", err)
	}
	if count != 1 {
		t.Errorf("Extracted PDF has %d pages, want 1", count)
	}
}

func TestExtractPagesEmptyList(t *testing.T) {
	pdfFile := samplePDF()
	if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output := filepath.Join(tmpDir, "extracted.pdf")
	// Empty pages list - behavior depends on implementation
	err = ExtractPages(pdfFile, output, []int{}, "")
	// Just verify it doesn't panic
	_ = err
}

func TestExtractPagesNonExistent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output := filepath.Join(tmpDir, "extracted.pdf")
	err = ExtractPages("/nonexistent/file.pdf", output, []int{1}, "")
	if err == nil {
		t.Error("ExtractPages() expected error for non-existent file")
	}
}

func TestRotate(t *testing.T) {
	pdf := samplePDF()
	if _, err := os.Stat(pdf); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output := filepath.Join(tmpDir, "rotated.pdf")
	err = Rotate(pdf, output, 90, nil, "")
	if err != nil {
		t.Fatalf("Rotate() error = %v", err)
	}

	// Verify output exists
	if _, err := os.Stat(output); os.IsNotExist(err) {
		t.Error("Rotate() did not create output file")
	}
}

func TestRotateWithPages(t *testing.T) {
	pdfFile := samplePDF()
	if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output := filepath.Join(tmpDir, "rotated.pdf")
	err = Rotate(pdfFile, output, 90, []int{1}, "")
	if err != nil {
		t.Fatalf("Rotate() with pages error = %v", err)
	}

	if _, err := os.Stat(output); os.IsNotExist(err) {
		t.Error("Rotate() with pages did not create output file")
	}
}

func TestRotateAllAngles(t *testing.T) {
	pdfFile := samplePDF()
	if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	angles := []int{0, 90, 180, 270}

	for _, angle := range angles {
		t.Run(fmt.Sprintf("%d_degrees", angle), func(t *testing.T) {
			output := filepath.Join(tmpDir, fmt.Sprintf("rotated_%d.pdf", angle))
			err := Rotate(pdfFile, output, angle, nil, "")
			if err != nil {
				t.Fatalf("Rotate() angle=%d error = %v", angle, err)
			}
			if _, err := os.Stat(output); os.IsNotExist(err) {
				t.Errorf("Rotate() did not create output file")
			}
		})
	}
}

func TestRotateNonExistent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output := filepath.Join(tmpDir, "rotated.pdf")
	err = Rotate("/nonexistent/file.pdf", output, 90, nil, "")
	if err == nil {
		t.Error("Rotate() expected error for non-existent file")
	}
}

func TestCompress(t *testing.T) {
	pdf := samplePDF()
	if _, err := os.Stat(pdf); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output := filepath.Join(tmpDir, "compressed.pdf")
	err = Compress(pdf, output, "")
	if err != nil {
		t.Fatalf("Compress() error = %v", err)
	}

	// Verify output exists
	if _, err := os.Stat(output); os.IsNotExist(err) {
		t.Error("Compress() did not create output file")
	}
}

func TestCompressNonExistent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output := filepath.Join(tmpDir, "compressed.pdf")
	err = Compress("/nonexistent/file.pdf", output, "")
	if err == nil {
		t.Error("Compress() expected error for non-existent file")
	}
}

func TestEncryptDecrypt(t *testing.T) {
	pdf := samplePDF()
	if _, err := os.Stat(pdf); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Encrypt
	encrypted := filepath.Join(tmpDir, "encrypted.pdf")
	password := "testpassword123"
	err = Encrypt(pdf, encrypted, password, "")
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	// Verify encrypted file exists
	if _, err := os.Stat(encrypted); os.IsNotExist(err) {
		t.Fatal("Encrypt() did not create output file")
	}

	// Decrypt
	decrypted := filepath.Join(tmpDir, "decrypted.pdf")
	err = Decrypt(encrypted, decrypted, password)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}

	// Verify decrypted file exists
	if _, err := os.Stat(decrypted); os.IsNotExist(err) {
		t.Error("Decrypt() did not create output file")
	}
}

func TestEncryptNonExistent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output := filepath.Join(tmpDir, "encrypted.pdf")
	err = Encrypt("/nonexistent/file.pdf", output, "password", "")
	if err == nil {
		t.Error("Encrypt() expected error for non-existent file")
	}
}

func TestDecryptNonExistent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output := filepath.Join(tmpDir, "decrypted.pdf")
	err = Decrypt("/nonexistent/file.pdf", output, "password")
	if err == nil {
		t.Error("Decrypt() expected error for non-existent file")
	}
}

func TestDecryptWrongPassword(t *testing.T) {
	pdfFile := samplePDF()
	if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// First encrypt the file
	encrypted := filepath.Join(tmpDir, "encrypted.pdf")
	err = Encrypt(pdfFile, encrypted, "correctpassword", "")
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	// Try to decrypt with wrong password
	decrypted := filepath.Join(tmpDir, "decrypted.pdf")
	err = Decrypt(encrypted, decrypted, "wrongpassword")
	if err == nil {
		t.Error("Decrypt() should return error for wrong password")
	}
}

func TestEncryptSpecialPassword(t *testing.T) {
	pdfFile := samplePDF()
	if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test with special characters in password
	passwords := []string{
		"p@ss!w0rd#$%",
		"pass with spaces",
		"unicode:日本語",
	}

	for _, pw := range passwords {
		t.Run(pw, func(t *testing.T) {
			encrypted := filepath.Join(tmpDir, "encrypted_special.pdf")
			err := Encrypt(pdfFile, encrypted, pw, "")
			if err != nil {
				t.Fatalf("Encrypt() with special password error = %v", err)
			}
		})
	}
}

func TestEncryptWithOwnerPassword(t *testing.T) {
	pdfFile := samplePDF()
	if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	encrypted := filepath.Join(tmpDir, "encrypted_owner.pdf")
	userPW := "user123"
	ownerPW := "owner456"

	// Encrypt with both user and owner passwords
	err = Encrypt(pdfFile, encrypted, userPW, ownerPW)
	if err != nil {
		t.Fatalf("Encrypt() with owner password error = %v", err)
	}

	// Verify encrypted file exists
	if _, err := os.Stat(encrypted); os.IsNotExist(err) {
		t.Fatal("Encrypt() did not create output file")
	}

	// Verify we can decrypt with owner password
	decrypted := filepath.Join(tmpDir, "decrypted.pdf")
	err = Decrypt(encrypted, decrypted, ownerPW)
	if err != nil {
		t.Fatalf("Decrypt() with owner password error = %v", err)
	}
}
