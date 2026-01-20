package util

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileExists(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*")
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	if !FileExists(tmpFile.Name()) {
		t.Error("FileExists() = false for existing file")
	}
	if FileExists("/nonexistent/path/to/file") {
		t.Error("FileExists() = true for non-existing file")
	}
}

func TestIsDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	tmpFile, err := os.CreateTemp("", "test-*")
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	if !IsDir(tmpDir) {
		t.Error("IsDir() = false for directory")
	}
	if IsDir(tmpFile.Name()) {
		t.Error("IsDir() = true for file")
	}
	if IsDir("/nonexistent/path") {
		t.Error("IsDir() = true for non-existing path")
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

	if err := EnsureDir(tmpDir); err != nil {
		t.Errorf("EnsureDir() on existing directory error = %v", err)
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
		t.Errorf("error = %v", err)
	}
	data, err := os.ReadFile(testPath)
	if err != nil {
		t.Fatalf("Failed to read: %v", err)
	}
	if string(data) != string(testData) {
		t.Errorf("wrote %q, want %q", data, testData)
	}

	nestedPath := filepath.Join(tmpDir, "nested", "dirs", "file.txt")
	if err := AtomicWrite(nestedPath, []byte("content")); err != nil {
		t.Errorf("AtomicWrite nested error = %v", err)
	}
	if !FileExists(nestedPath) {
		t.Error("AtomicWrite() did not create file in nested directory")
	}
}

func TestGenerateOutputFilename(t *testing.T) {
	tests := []struct {
		input, suffix, want string
	}{
		{"document.pdf", "_output", "document_output.pdf"},
		{"/path/to/document.pdf", "_compressed", "/path/to/document_compressed.pdf"},
		{"document.PDF", "_rotated", "document_rotated.PDF"},
	}
	for _, tt := range tests {
		if got := GenerateOutputFilename(tt.input, tt.suffix); got != tt.want {
			t.Errorf("GenerateOutputFilename(%q, %q) = %q, want %q", tt.input, tt.suffix, got, tt.want)
		}
	}
}

func TestFormatFileSize(t *testing.T) {
	tests := []struct {
		bytes int64
		want  string
	}{
		{0, "0 B"},
		{1, "1 B"},
		{500, "500 B"},
		{1023, "1023 B"},
		{1024, "1.00 KB"},
		{1024 * 1024, "1.00 MB"},
		{1536 * 1024, "1.50 MB"},
		{1024 * 1024 * 1024, "1.00 GB"},
		{1024*1024 - 1, "1024.00 KB"},
		{1024*1024*1024 - 1, "1024.00 MB"},
	}
	for _, tt := range tests {
		if got := FormatFileSize(tt.bytes); got != tt.want {
			t.Errorf("FormatFileSize(%d) = %q, want %q", tt.bytes, got, tt.want)
		}
	}
}

func TestValidatePDFFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	pdfFile := filepath.Join(tmpDir, "test.pdf")
	upperPDF := filepath.Join(tmpDir, "test.PDF")
	txtFile := filepath.Join(tmpDir, "test.txt")
	for _, f := range []string{pdfFile, upperPDF, txtFile} {
		if err := os.WriteFile(f, []byte("dummy"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		path    string
		wantErr bool
	}{
		{pdfFile, false},
		{upperPDF, false},
		{txtFile, true},
		{"/nonexistent/file.pdf", true},
	}
	for _, tt := range tests {
		t.Run(filepath.Base(tt.path), func(t *testing.T) {
			if err := ValidatePDFFile(tt.path); (err != nil) != tt.wantErr {
				t.Errorf("ValidatePDFFile(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}
		})
	}
}

func TestCopyFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	srcFile := filepath.Join(tmpDir, "source.txt")
	srcContent := []byte("Hello, World!")
	if err := os.WriteFile(srcFile, srcContent, 0644); err != nil {
		t.Fatal(err)
	}

	t.Run("success", func(t *testing.T) {
		dstFile := filepath.Join(tmpDir, "dest.txt")
		if err := CopyFile(srcFile, dstFile); err != nil {
			t.Fatalf("error = %v", err)
		}
		if content, _ := os.ReadFile(dstFile); string(content) != string(srcContent) {
			t.Errorf("content = %q, want %q", content, srcContent)
		}
	})

	t.Run("nested destination", func(t *testing.T) {
		dstFile := filepath.Join(tmpDir, "nested", "dir", "dest.txt")
		if err := CopyFile(srcFile, dstFile); err != nil {
			t.Fatalf("error = %v", err)
		}
		if !FileExists(dstFile) {
			t.Error("did not create destination file")
		}
	})

	t.Run("non-existent source", func(t *testing.T) {
		if err := CopyFile("/nonexistent/file.txt", filepath.Join(tmpDir, "out.txt")); err == nil {
			t.Error("should return error for non-existent source")
		}
	})

	t.Run("overwrite", func(t *testing.T) {
		dstFile := filepath.Join(tmpDir, "overwrite.txt")
		os.WriteFile(dstFile, []byte("old"), 0644)
		if err := CopyFile(srcFile, dstFile); err != nil {
			t.Fatalf("error = %v", err)
		}
		if content, _ := os.ReadFile(dstFile); string(content) != string(srcContent) {
			t.Error("did not overwrite content")
		}
	})

	t.Run("large file", func(t *testing.T) {
		largeFile := filepath.Join(tmpDir, "large.bin")
		largeContent := make([]byte, 1024*1024)
		os.WriteFile(largeFile, largeContent, 0644)

		dstFile := filepath.Join(tmpDir, "large_copy.bin")
		if err := CopyFile(largeFile, dstFile); err != nil {
			t.Fatalf("error = %v", err)
		}
		srcInfo, _ := os.Stat(largeFile)
		dstInfo, _ := os.Stat(dstFile)
		if srcInfo.Size() != dstInfo.Size() {
			t.Errorf("size mismatch: src=%d, dst=%d", srcInfo.Size(), dstInfo.Size())
		}
	})
}

func TestValidatePDFFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	pdfs := make([]string, 10)
	for i := 0; i < 10; i++ {
		pdfs[i] = filepath.Join(tmpDir, "file"+string(rune('a'+i))+".pdf")
		os.WriteFile(pdfs[i], []byte("dummy"), 0644)
	}
	txtFile := filepath.Join(tmpDir, "file.txt")
	os.WriteFile(txtFile, []byte("dummy"), 0644)

	tests := []struct {
		name    string
		paths   []string
		wantErr bool
	}{
		{"empty", []string{}, false},
		{"single valid", pdfs[:1], false},
		{"two valid", pdfs[:2], false},
		{"five valid", pdfs[:5], false},
		{"ten valid", pdfs, false},
		{"invalid", []string{txtFile}, true},
		{"mixed", []string{pdfs[0], txtFile}, true},
		{"non-existent", []string{"/nonexistent/file.pdf"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidatePDFFiles(tt.paths); (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetFileSize(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("1234567890"), 0644)

	if size, err := GetFileSize(testFile); err != nil || size != 10 {
		t.Errorf("GetFileSize() = %d, %v; want 10, nil", size, err)
	}
	if _, err := GetFileSize("/nonexistent/file.txt"); err == nil {
		t.Error("should error for non-existent file")
	}
	if _, err := GetFileSize(tmpDir); err != nil {
		t.Errorf("GetFileSize() on directory error = %v", err)
	}
}

func TestEnsureParentDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	paths := []string{
		filepath.Join(tmpDir, "a", "b", "c", "file.txt"),
		filepath.Join(tmpDir, "single", "file.txt"),
		"file.txt",
		"/file.txt",
		filepath.Join(tmpDir, "file.txt"),
	}
	for _, path := range paths {
		if err := EnsureParentDir(path); err != nil {
			t.Errorf("EnsureParentDir(%q) error = %v", path, err)
		}
	}
}
