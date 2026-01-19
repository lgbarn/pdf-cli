package util

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileExists(t *testing.T) {
	// Create a temp file
	tmpFile, err := os.CreateTemp("", "test-*")
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "existing file",
			path: tmpFile.Name(),
			want: true,
		},
		{
			name: "non-existing file",
			path: "/nonexistent/path/to/file",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FileExists(tt.path); got != tt.want {
				t.Errorf("FileExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsDir(t *testing.T) {
	// Create a temp directory
	tmpDir, err := os.MkdirTemp("", "test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a temp file
	tmpFile, err := os.CreateTemp("", "test-*")
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	tests := []struct {
		name string
		path string
		want bool
	}{
		{
			name: "directory",
			path: tmpDir,
			want: true,
		},
		{
			name: "file",
			path: tmpFile.Name(),
			want: false,
		},
		{
			name: "non-existing",
			path: "/nonexistent/path",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsDir(tt.path); got != tt.want {
				t.Errorf("IsDir() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEnsureDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	newDir := filepath.Join(tmpDir, "new", "nested", "dir")

	if err := EnsureDir(newDir); err != nil {
		t.Errorf("EnsureDir() error = %v", err)
	}

	if !IsDir(newDir) {
		t.Error("EnsureDir() did not create directory")
	}
}

func TestAtomicWrite(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	testPath := filepath.Join(tmpDir, "test.txt")
	testData := []byte("Hello, World!")

	if err := AtomicWrite(testPath, testData); err != nil {
		t.Errorf("AtomicWrite() error = %v", err)
	}

	// Verify the file exists and has correct content
	data, err := os.ReadFile(testPath)
	if err != nil {
		t.Errorf("Failed to read written file: %v", err)
	}

	if string(data) != string(testData) {
		t.Errorf("AtomicWrite() wrote %q, want %q", string(data), string(testData))
	}
}

func TestGenerateOutputFilename(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		suffix string
		want   string
	}{
		{
			name:   "simple pdf",
			input:  "document.pdf",
			suffix: "_output",
			want:   "document_output.pdf",
		},
		{
			name:   "with path",
			input:  "/path/to/document.pdf",
			suffix: "_compressed",
			want:   "/path/to/document_compressed.pdf",
		},
		{
			name:   "uppercase extension",
			input:  "document.PDF",
			suffix: "_rotated",
			want:   "document_rotated.PDF",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GenerateOutputFilename(tt.input, tt.suffix); got != tt.want {
				t.Errorf("GenerateOutputFilename() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatFileSize(t *testing.T) {
	tests := []struct {
		name  string
		bytes int64
		want  string
	}{
		{
			name:  "bytes",
			bytes: 500,
			want:  "500 B",
		},
		{
			name:  "kilobytes",
			bytes: 1024,
			want:  "1.00 KB",
		},
		{
			name:  "megabytes",
			bytes: 1024 * 1024,
			want:  "1.00 MB",
		},
		{
			name:  "gigabytes",
			bytes: 1024 * 1024 * 1024,
			want:  "1.00 GB",
		},
		{
			name:  "mixed",
			bytes: 1536 * 1024,
			want:  "1.50 MB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatFileSize(tt.bytes); got != tt.want {
				t.Errorf("FormatFileSize() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidatePDFFile(t *testing.T) {
	// Create a temp PDF file
	tmpDir, err := os.MkdirTemp("", "test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	pdfFile := filepath.Join(tmpDir, "test.pdf")
	if err := os.WriteFile(pdfFile, []byte("dummy"), 0644); err != nil {
		t.Fatal(err)
	}

	txtFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(txtFile, []byte("dummy"), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid pdf file",
			path:    pdfFile,
			wantErr: false,
		},
		{
			name:    "non-pdf file",
			path:    txtFile,
			wantErr: true,
		},
		{
			name:    "non-existing file",
			path:    "/nonexistent/file.pdf",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePDFFile(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePDFFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
