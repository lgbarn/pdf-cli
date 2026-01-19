package pdf

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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
			conf := newConfig(tt.password)
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
