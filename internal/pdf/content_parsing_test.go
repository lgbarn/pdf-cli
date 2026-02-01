package pdf

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
