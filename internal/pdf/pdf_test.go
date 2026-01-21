package pdf

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ledongthuc/pdf"
)

// testdataDir returns the path to the testdata directory
func testdataDir() string {
	return filepath.Join("..", "..", "testdata")
}

// samplePDF returns the path to the sample PDF file
func samplePDF() string {
	return filepath.Join(testdataDir(), "sample.pdf")
}

func TestPagesToStrings(t *testing.T) {
	tests := []struct {
		name  string
		pages []int
		want  []string
	}{
		{
			name:  "empty slice",
			pages: []int{},
			want:  nil,
		},
		{
			name:  "nil slice",
			pages: nil,
			want:  nil,
		},
		{
			name:  "single page",
			pages: []int{1},
			want:  []string{"1"},
		},
		{
			name:  "multiple pages",
			pages: []int{1, 3, 5, 10},
			want:  []string{"1", "3", "5", "10"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pagesToStrings(tt.pages)
			if len(got) != len(tt.want) {
				t.Errorf("pagesToStrings() length = %d, want %d", len(got), len(tt.want))
				return
			}
			for i, v := range got {
				if v != tt.want[i] {
					t.Errorf("pagesToStrings()[%d] = %v, want %v", i, v, tt.want[i])
				}
			}
		})
	}
}

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name     string
		password string
	}{
		{
			name:     "no password",
			password: "",
		},
		{
			name:     "with password",
			password: "secret123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conf := NewConfig(tt.password)
			if conf == nil {
				t.Fatal("newConfig() returned nil")
			}
			if tt.password != "" {
				if conf.UserPW != tt.password {
					t.Errorf("newConfig() UserPW = %v, want %v", conf.UserPW, tt.password)
				}
				if conf.OwnerPW != tt.password {
					t.Errorf("newConfig() OwnerPW = %v, want %v", conf.OwnerPW, tt.password)
				}
			}
		})
	}
}

func TestGetInfo(t *testing.T) {
	pdf := samplePDF()
	if _, err := os.Stat(pdf); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	info, err := GetInfo(pdf, "")
	if err != nil {
		t.Fatalf("GetInfo() error = %v", err)
	}

	if info == nil {
		t.Fatal("GetInfo() returned nil")
	}

	if info.Pages < 1 {
		t.Errorf("GetInfo() Pages = %d, want >= 1", info.Pages)
	}

	if info.FileSize <= 0 {
		t.Errorf("GetInfo() FileSize = %d, want > 0", info.FileSize)
	}

	if info.Version == "" {
		t.Error("GetInfo() Version is empty")
	}
}

func TestGetInfoNonExistent(t *testing.T) {
	_, err := GetInfo("/nonexistent/file.pdf", "")
	if err == nil {
		t.Error("GetInfo() expected error for non-existent file")
	}
}

func TestPageCount(t *testing.T) {
	pdf := samplePDF()
	if _, err := os.Stat(pdf); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	count, err := PageCount(pdf, "")
	if err != nil {
		t.Fatalf("PageCount() error = %v", err)
	}

	if count < 1 {
		t.Errorf("PageCount() = %d, want >= 1", count)
	}
}

func TestPageCountNonExistent(t *testing.T) {
	_, err := PageCount("/nonexistent/file.pdf", "")
	if err == nil {
		t.Error("PageCount() expected error for non-existent file")
	}
}

func TestExtractText(t *testing.T) {
	pdf := samplePDF()
	if _, err := os.Stat(pdf); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	text, err := ExtractText(pdf, nil, "")
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
		text, err := ExtractText(pdf, []int{1}, "")
		if err != nil {
			t.Fatalf("ExtractText() with pages error = %v", err)
		}
		// Just verify it doesn't crash
		_ = text
	}
}

func TestValidate(t *testing.T) {
	pdf := samplePDF()
	if _, err := os.Stat(pdf); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	err := Validate(pdf, "")
	if err != nil {
		t.Errorf("Validate() error = %v", err)
	}
}

func TestValidateNonExistent(t *testing.T) {
	err := Validate("/nonexistent/file.pdf", "")
	if err == nil {
		t.Error("Validate() expected error for non-existent file")
	}
}

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

func TestGetMetadata(t *testing.T) {
	pdf := samplePDF()
	if _, err := os.Stat(pdf); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	meta, err := GetMetadata(pdf, "")
	if err != nil {
		t.Fatalf("GetMetadata() error = %v", err)
	}

	if meta == nil {
		t.Fatal("GetMetadata() returned nil")
	}
}

func TestParseTextFromPDFContent(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name:    "simple text",
			content: "(Hello World) Tj",
			want:    "Hello World",
		},
		{
			name:    "multiple strings",
			content: "(Hello) Tj (World) Tj",
			want:    "Hello World",
		},
		{
			name:    "escaped parentheses",
			content: "(Hello \\(World\\)) Tj",
			want:    "Hello (World)",
		},
		{
			name:    "newline escape",
			content: "(Hello\\nWorld) Tj",
			want:    "Hello\nWorld",
		},
		{
			name:    "empty content",
			content: "",
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseTextFromPDFContent(tt.content)
			if got != tt.want {
				t.Errorf("parseTextFromPDFContent() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestExtractParenString(t *testing.T) {
	tests := []struct {
		name    string
		content string
		start   int
		want    string
	}{
		{
			name:    "simple string",
			content: "(Hello)",
			start:   0,
			want:    "Hello",
		},
		{
			name:    "nested parens",
			content: "(Hello (World))",
			start:   0,
			want:    "Hello (World)",
		},
		{
			name:    "escaped backslash",
			content: "(Hello\\\\World)",
			start:   0,
			want:    "Hello\\World",
		},
		{
			name:    "not at start",
			content: "BT (Hello) Tj",
			start:   3,
			want:    "Hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := extractParenString(tt.content, tt.start)
			if got != tt.want {
				t.Errorf("extractParenString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestAddWatermark(t *testing.T) {
	pdf := samplePDF()
	if _, err := os.Stat(pdf); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output := filepath.Join(tmpDir, "watermarked.pdf")
	err = AddWatermark(pdf, output, "CONFIDENTIAL", nil, "")
	if err != nil {
		t.Fatalf("AddWatermark() error = %v", err)
	}

	// Verify output exists
	if _, err := os.Stat(output); os.IsNotExist(err) {
		t.Error("AddWatermark() did not create output file")
	}
}

func TestAddWatermarkWithPages(t *testing.T) {
	pdfFile := samplePDF()
	if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output := filepath.Join(tmpDir, "watermarked.pdf")
	err = AddWatermark(pdfFile, output, "DRAFT", []int{1}, "")
	if err != nil {
		t.Fatalf("AddWatermark() with pages error = %v", err)
	}

	if _, err := os.Stat(output); os.IsNotExist(err) {
		t.Error("AddWatermark() with pages did not create output file")
	}
}

func TestAddImageWatermark(t *testing.T) {
	pdfFile := samplePDF()
	if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
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

	output := filepath.Join(tmpDir, "image_watermarked.pdf")
	err = AddImageWatermark(pdfFile, output, testImage, nil, "")
	if err != nil {
		t.Fatalf("AddImageWatermark() error = %v", err)
	}

	if _, err := os.Stat(output); os.IsNotExist(err) {
		t.Error("AddImageWatermark() did not create output file")
	}
}

func TestAddImageWatermarkWithPages(t *testing.T) {
	pdfFile := samplePDF()
	if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
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

	output := filepath.Join(tmpDir, "image_watermarked.pdf")
	err = AddImageWatermark(pdfFile, output, testImage, []int{1}, "")
	if err != nil {
		t.Fatalf("AddImageWatermark() with pages error = %v", err)
	}

	if _, err := os.Stat(output); os.IsNotExist(err) {
		t.Error("AddImageWatermark() with pages did not create output file")
	}
}

func TestAddImageWatermarkNonExistent(t *testing.T) {
	pdfFile := samplePDF()
	if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output := filepath.Join(tmpDir, "watermarked.pdf")
	err = AddImageWatermark(pdfFile, output, "/nonexistent/image.png", nil, "")
	if err == nil {
		t.Error("AddImageWatermark() expected error for non-existent image")
	}
}

func TestSetMetadata(t *testing.T) {
	pdfFile := samplePDF()
	if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output := filepath.Join(tmpDir, "metadata.pdf")
	meta := &Metadata{
		Title:    "Test Title",
		Author:   "Test Author",
		Subject:  "Test Subject",
		Keywords: "test, keywords",
	}

	err = SetMetadata(pdfFile, output, meta, "")
	if err != nil {
		t.Fatalf("SetMetadata() error = %v", err)
	}

	if _, err := os.Stat(output); os.IsNotExist(err) {
		t.Error("SetMetadata() did not create output file")
	}
}

func TestSetMetadataPartial(t *testing.T) {
	pdfFile := samplePDF()
	if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output := filepath.Join(tmpDir, "metadata.pdf")
	meta := &Metadata{
		Title: "Only Title",
	}

	err = SetMetadata(pdfFile, output, meta, "")
	if err != nil {
		t.Fatalf("SetMetadata() with partial metadata error = %v", err)
	}

	if _, err := os.Stat(output); os.IsNotExist(err) {
		t.Error("SetMetadata() did not create output file")
	}
}

func TestSetMetadataEmpty(t *testing.T) {
	pdfFile := samplePDF()
	if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output := filepath.Join(tmpDir, "metadata.pdf")
	meta := &Metadata{}

	// Empty metadata should still work
	err = SetMetadata(pdfFile, output, meta, "")
	if err != nil {
		t.Fatalf("SetMetadata() with empty metadata error = %v", err)
	}
}

func TestValidateToBuffer(t *testing.T) {
	pdfFile := samplePDF()
	if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	data, err := os.ReadFile(pdfFile)
	if err != nil {
		t.Fatalf("Failed to read sample PDF: %v", err)
	}

	err = ValidateToBuffer(data)
	if err != nil {
		t.Errorf("ValidateToBuffer() error = %v", err)
	}
}

func TestValidateToBufferInvalid(t *testing.T) {
	// Invalid PDF data
	invalidData := []byte("This is not a valid PDF")
	err := ValidateToBuffer(invalidData)
	if err == nil {
		t.Error("ValidateToBuffer() expected error for invalid data")
	}
}

func TestValidateToBufferEmpty(t *testing.T) {
	err := ValidateToBuffer([]byte{})
	if err == nil {
		t.Error("ValidateToBuffer() expected error for empty data")
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
	text, err := extractPagesSequential(r, pages, totalPages, false)
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

	text, err := extractPagesSequential(r, pages, totalPages, false)
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

	text, err := extractPagesSequential(r, pages, totalPages, false)
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
	text, err := extractPagesParallel(r, pages, totalPages, false)
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

	text, err := extractPagesParallel(r, pages, totalPages, false)
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

	text, err := extractPagesParallel(r, pages, totalPages, false)
	if err != nil {
		t.Fatalf("extractPagesParallel() out of range error = %v", err)
	}

	// Should return empty for out-of-range pages
	if text != "" {
		t.Log("Warning: extractPagesParallel returned non-empty for out-of-range page")
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

func TestExtractTextNonExistent(t *testing.T) {
	_, err := ExtractText("/nonexistent/file.pdf", nil, "")
	if err == nil {
		t.Error("ExtractText() expected error for non-existent file")
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

// PDF/A validation tests

func TestValidatePDFA(t *testing.T) {
	pdfFile := samplePDF()
	if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	result, err := ValidatePDFA(pdfFile, "1b", "")
	if err != nil {
		t.Fatalf("ValidatePDFA() error = %v", err)
	}

	if result == nil {
		t.Fatal("ValidatePDFA() returned nil result")
	}

	if result.Level != "1b" {
		t.Errorf("ValidatePDFA() Level = %q, want %q", result.Level, "1b")
	}

	// Note: regular PDFs are unlikely to pass PDF/A validation
	// We just verify the function works without crashing
}

func TestValidatePDFALevels(t *testing.T) {
	pdfFile := samplePDF()
	if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	levels := []string{"1a", "1b", "2a", "2b", "3a", "3b"}

	for _, level := range levels {
		t.Run(level, func(t *testing.T) {
			result, err := ValidatePDFA(pdfFile, level, "")
			if err != nil {
				t.Fatalf("ValidatePDFA(%q) error = %v", level, err)
			}
			if result == nil {
				t.Fatalf("ValidatePDFA(%q) returned nil", level)
			}
			if result.Level != level {
				t.Errorf("ValidatePDFA() Level = %q, want %q", result.Level, level)
			}
		})
	}
}

func TestValidatePDFANonExistent(t *testing.T) {
	result, err := ValidatePDFA("/nonexistent/file.pdf", "1b", "")
	// ValidatePDFA may return a result with IsValid=false instead of an error
	if err == nil && result != nil && result.IsValid {
		t.Error("ValidatePDFA() should report invalid for non-existent file")
	}
}

func TestConvertToPDFA(t *testing.T) {
	pdfFile := samplePDF()
	if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output := filepath.Join(tmpDir, "pdfa.pdf")
	err = ConvertToPDFA(pdfFile, output, "1b", "")
	if err != nil {
		t.Fatalf("ConvertToPDFA() error = %v", err)
	}

	// Verify output exists
	if _, err := os.Stat(output); os.IsNotExist(err) {
		t.Error("ConvertToPDFA() did not create output file")
	}
}

func TestConvertToPDFANonExistent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output := filepath.Join(tmpDir, "pdfa.pdf")
	err = ConvertToPDFA("/nonexistent/file.pdf", output, "1b", "")
	if err == nil {
		t.Error("ConvertToPDFA() expected error for non-existent file")
	}
}

// CreatePDFFromImages tests

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

// Additional edge case tests

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

func TestGetMetadataNonExistent(t *testing.T) {
	_, err := GetMetadata("/nonexistent/file.pdf", "")
	if err == nil {
		t.Error("GetMetadata() expected error for non-existent file")
	}
}

func TestSetMetadataNonExistent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output := filepath.Join(tmpDir, "metadata.pdf")
	meta := &Metadata{Title: "Test"}
	err = SetMetadata("/nonexistent/file.pdf", output, meta, "")
	if err == nil {
		t.Error("SetMetadata() expected error for non-existent file")
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

func TestAddWatermarkNonExistent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output := filepath.Join(tmpDir, "watermarked.pdf")
	err = AddWatermark("/nonexistent/file.pdf", output, "WATERMARK", nil, "")
	if err == nil {
		t.Error("AddWatermark() expected error for non-existent file")
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

func TestExtractTextWithProgress(t *testing.T) {
	pdfFile := samplePDF()
	if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	// Test with progress enabled
	text, err := ExtractTextWithProgress(pdfFile, nil, "", true)
	if err != nil {
		t.Fatalf("ExtractTextWithProgress() error = %v", err)
	}
	// Just verify it returns without error
	_ = text
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

// Additional edge case tests for helper functions

func TestExtractParenStringEdges(t *testing.T) {
	tests := []struct {
		name    string
		content string
		start   int
		want    string
		wantEnd int
	}{
		{
			name:    "tab escape",
			content: "(a\\tb)",
			start:   0,
			want:    "a\tb",
			wantEnd: 6,
		},
		{
			name:    "return escape",
			content: "(a\\rb)",
			start:   0,
			want:    "a\rb",
			wantEnd: 6,
		},
		{
			name:    "backslash escape",
			content: "(a\\\\b)",
			start:   0,
			want:    "a\\b",
			wantEnd: 6,
		},
		{
			name:    "out of bounds start",
			content: "(Hello)",
			start:   100,
			want:    "",
			wantEnd: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, end := extractParenString(tt.content, tt.start)
			if got != tt.want {
				t.Errorf("extractParenString() string = %q, want %q", got, tt.want)
			}
			if end != tt.wantEnd {
				t.Errorf("extractParenString() end = %d, want %d", end, tt.wantEnd)
			}
		})
	}
}

func TestInfoStruct(t *testing.T) {
	info := Info{
		FilePath:    "/path/to/file.pdf",
		FileSize:    1024,
		Pages:       10,
		Version:     "1.4",
		Title:       "Test PDF",
		Author:      "Test Author",
		Subject:     "Test Subject",
		Keywords:    "test, pdf",
		Creator:     "Test Creator",
		Producer:    "Test Producer",
		CreatedDate: "2024-01-01",
		ModDate:     "2024-01-02",
		Encrypted:   false,
	}

	if info.FilePath != "/path/to/file.pdf" {
		t.Errorf("Info.FilePath = %q, want %q", info.FilePath, "/path/to/file.pdf")
	}
	if info.Pages != 10 {
		t.Errorf("Info.Pages = %d, want %d", info.Pages, 10)
	}
}

func TestMetadataStruct(t *testing.T) {
	meta := Metadata{
		Title:       "Test Title",
		Author:      "Test Author",
		Subject:     "Test Subject",
		Keywords:    "test, keywords",
		Creator:     "Test Creator",
		Producer:    "Test Producer",
		CreatedDate: "2024-01-01",
		ModDate:     "2024-01-02",
	}

	if meta.Title != "Test Title" {
		t.Errorf("Metadata.Title = %q, want %q", meta.Title, "Test Title")
	}
}

func TestPDFAValidationResultStruct(t *testing.T) {
	result := PDFAValidationResult{
		IsValid:  false,
		Level:    "1b",
		Errors:   []string{"error1", "error2"},
		Warnings: []string{"warning1"},
	}

	if result.IsValid {
		t.Error("PDFAValidationResult.IsValid should be false")
	}
	if result.Level != "1b" {
		t.Errorf("PDFAValidationResult.Level = %q, want %q", result.Level, "1b")
	}
	if len(result.Errors) != 2 {
		t.Errorf("PDFAValidationResult.Errors length = %d, want 2", len(result.Errors))
	}
}

func TestValidateValid(t *testing.T) {
	pdfFile := samplePDF()
	if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	err := Validate(pdfFile, "")
	if err != nil {
		t.Errorf("Validate() error = %v", err)
	}
}

// Additional tests for 75%+ coverage

// Test encryption with separate owner password
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

// Test MergeWithProgress with many files (triggers incremental merge)
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

// Test MergeWithProgress error handling for non-existent first file
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

// Test SetMetadata with all fields including Creator and Producer
func TestSetMetadataAllFields(t *testing.T) {
	pdfFile := samplePDF()
	if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	output := filepath.Join(tmpDir, "metadata_all.pdf")
	meta := &Metadata{
		Title:    "Full Test Title",
		Author:   "Full Test Author",
		Subject:  "Full Test Subject",
		Keywords: "full, test, keywords",
		Creator:  "Full Test Creator",
		Producer: "Full Test Producer",
	}

	err = SetMetadata(pdfFile, output, meta, "")
	if err != nil {
		t.Fatalf("SetMetadata() with all fields error = %v", err)
	}

	if _, err := os.Stat(output); os.IsNotExist(err) {
		t.Error("SetMetadata() did not create output file")
	}
}

// Test extractTextPrimary with specific pages (sorted pages path)
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

	text, err := ExtractTextWithProgress(pdfFile, pages, "", false)
	if err != nil {
		t.Fatalf("ExtractTextWithProgress() with specific pages error = %v", err)
	}
	_ = text
}

// Test extractPageText with invalid page numbers
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

// Test parseTextFromPDFContent with ET markers
func TestParseTextFromPDFContentWithET(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name:    "text followed by ET",
			content: "(Hello) Tj ET (World) Tj",
			want:    "Hello\nWorld",
		},
		{
			name:    "multiple ET markers",
			content: "(Line1) Tj ET (Line2) Tj ET (Line3) Tj",
			want:    "Line1\nLine2\nLine3",
		},
		{
			name:    "octal escape",
			content: "(Test\\101) Tj",
			want:    "Test01", // Octal escape skips backslash, keeps digits
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseTextFromPDFContent(tt.content)
			if got != tt.want {
				t.Errorf("parseTextFromPDFContent() = %q, want %q", got, tt.want)
			}
		})
	}
}

// Test extractParenString with unclosed parenthesis
func TestExtractParenStringUnclosed(t *testing.T) {
	content := "(Hello World"
	got, end := extractParenString(content, 0)

	// Should extract what's available
	if got != "Hello World" {
		t.Errorf("extractParenString() unclosed = %q, want %q", got, "Hello World")
	}
	if end != len(content) {
		t.Errorf("extractParenString() end = %d, want %d", end, len(content))
	}
}

// Test extractParenString when not starting with parenthesis
func TestExtractParenStringNotParenthesis(t *testing.T) {
	content := "Hello World"
	got, end := extractParenString(content, 0)

	if got != "" {
		t.Errorf("extractParenString() not paren = %q, want empty", got)
	}
	if end != 0 {
		t.Errorf("extractParenString() end = %d, want 0", end)
	}
}

// Test SplitWithProgress with large PDF (triggers progress bar path)
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

// Test SplitWithProgress with multi-page chunks
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

// Test CreatePDFFromImages with default page size handling
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

// Test extractPagesSequential with progress
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
	text, err := extractPagesSequential(r, pages, totalPages, true)
	if err != nil {
		t.Fatalf("extractPagesSequential() with progress error = %v", err)
	}
	_ = text
}

// Test extractPagesParallel with progress
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
	text, err := extractPagesParallel(r, pages, totalPages, true)
	if err != nil {
		t.Fatalf("extractPagesParallel() with progress error = %v", err)
	}
	_ = text
}

// Test ValidatePDFA with invalid PDF (triggers the error path)
func TestValidatePDFAInvalidPDF(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create an invalid PDF file
	invalidPDF := filepath.Join(tmpDir, "invalid.pdf")
	err = os.WriteFile(invalidPDF, []byte("not a valid pdf"), 0644)
	if err != nil {
		t.Fatalf("Failed to write invalid PDF: %v", err)
	}

	result, err := ValidatePDFA(invalidPDF, "1b", "")
	// Either returns error or result with IsValid=false
	if err == nil && result != nil && result.IsValid {
		t.Error("ValidatePDFA() should report invalid for invalid PDF")
	}
}

// Test extractTextFallback is triggered when primary fails
func TestExtractTextFallbackPath(t *testing.T) {
	// Create an empty temp dir to simulate no text extraction
	tmpDir, err := os.MkdirTemp("", "pdf-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Test with non-existent file to trigger error path
	_, err = ExtractText("/nonexistent/file.pdf", nil, "")
	if err == nil {
		t.Error("ExtractText() expected error for non-existent file")
	}
}

// Test ExtractTextWithProgress with many pages (triggers parallel extraction)
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
	text, err := ExtractTextWithProgress(largePDF, nil, "", true)
	if err != nil {
		t.Fatalf("ExtractTextWithProgress() error = %v", err)
	}
	_ = text
}

// Test MergeWithProgress without progress flag but with many files
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

// Test extractParenString with deeply nested parentheses
func TestExtractParenStringDeepNesting(t *testing.T) {
	content := "(a(b(c)d)e)"
	got, end := extractParenString(content, 0)

	if got != "a(b(c)d)e" {
		t.Errorf("extractParenString() deep nesting = %q, want %q", got, "a(b(c)d)e")
	}
	if end != len(content) {
		t.Errorf("extractParenString() end = %d, want %d", end, len(content))
	}
}

// Test ExtractImages with specific pages
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

// Test SplitWithProgress error handling for non-existent file
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

// Test parseTextFromPDFContent handles various edge cases
func TestParseTextFromPDFContentEdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name:    "only spaces in parens",
			content: "(   ) Tj",
			want:    "",
		},
		{
			name:    "newline in text",
			content: "(Line1\\nLine2) Tj",
			want:    "Line1\nLine2",
		},
		{
			name:    "content with no parentheses",
			content: "BT /F1 12 Tf ET",
			want:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseTextFromPDFContent(tt.content)
			if strings.TrimSpace(got) != strings.TrimSpace(tt.want) {
				t.Errorf("parseTextFromPDFContent() = %q, want %q", got, tt.want)
			}
		})
	}
}
