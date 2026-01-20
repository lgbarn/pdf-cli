package pdf

import (
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
