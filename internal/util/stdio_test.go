package util

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsStdinInput(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"-", true},
		{"", false},
		{"/path/to/file.pdf", false},
		{"file.pdf", false},
		{"--", false},
		{"/path/to/my-file.pdf", false},
		{" - ", false},
		{"C:\\path\\to\\my-file.pdf", false},
		{".hidden", false},
		{"file-", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := IsStdinInput(tt.input); got != tt.want {
				t.Errorf("IsStdinInput(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestStdinIndicator(t *testing.T) {
	if StdinIndicator != "-" {
		t.Errorf("StdinIndicator = %q, want %q", StdinIndicator, "-")
	}
}

func TestIsStdinPiped(t *testing.T) {
	_ = IsStdinPiped()
}

func TestReadFromStdinNotPiped(t *testing.T) {
	if !IsStdinPiped() {
		t.Skip("stdin is not piped in test environment")
	}
	path, cleanup, err := ReadFromStdin()
	if err != nil {
		return
	}
	defer cleanup()
	if path == "" {
		t.Error("ReadFromStdin() returned empty path")
	}
}

func TestWriteToStdout(t *testing.T) {
	if err := WriteToStdout("/nonexistent/file.txt"); err == nil {
		t.Error("WriteToStdout() should return error for non-existent file")
	}

	tmpDir, err := os.MkdirTemp("", "test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := []byte("Hello, World!")
	if err := os.WriteFile(testFile, testContent, 0644); err != nil {
		t.Fatal(err)
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err = WriteToStdout(testFile)

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Errorf("WriteToStdout() error = %v", err)
	}

	buf := make([]byte, 100)
	n, _ := r.Read(buf)
	if string(buf[:n]) != string(testContent) {
		t.Errorf("output = %q, want %q", string(buf[:n]), string(testContent))
	}
}

func TestResolveInputPath(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.pdf")
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	t.Run("regular file", func(t *testing.T) {
		path, cleanup, err := ResolveInputPath(tmpFile.Name())
		if err != nil {
			t.Fatalf("error = %v", err)
		}
		if path != tmpFile.Name() {
			t.Errorf("path = %q, want %q", path, tmpFile.Name())
		}
		cleanup()
		cleanup()
		cleanup()
		if !FileExists(tmpFile.Name()) {
			t.Error("cleanup should not remove regular files")
		}
	})

	t.Run("stdin not piped", func(t *testing.T) {
		if IsStdinPiped() {
			t.Skip("stdin is piped")
		}
		if _, _, err := ResolveInputPath("-"); err == nil {
			t.Error("should return error when stdin is not piped")
		}
	})

	t.Run("non-stdin dash patterns", func(t *testing.T) {
		for _, tc := range []string{"--", "-file.pdf", "file-.pdf"} {
			path, cleanup, err := ResolveInputPath(tc)
			if err != nil {
				continue
			}
			defer cleanup()
			if path != tc {
				t.Errorf("ResolveInputPath(%q) = %q", tc, path)
			}
		}
	})
}
