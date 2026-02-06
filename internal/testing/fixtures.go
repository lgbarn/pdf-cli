package testing

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// TestdataDir returns the path to the testdata directory.
// It handles being called from any package by finding the project root.
func TestdataDir() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("failed to get caller information")
	}
	// Go up from internal/testing to project root
	projectRoot := filepath.Dir(filepath.Dir(filepath.Dir(filename)))
	return filepath.Join(projectRoot, "testdata")
}

// SamplePDF returns the path to sample.pdf in testdata.
func SamplePDF() string {
	return filepath.Join(TestdataDir(), "sample.pdf")
}

// TestImage returns the path to test_image.png in testdata.
func TestImage() string {
	return filepath.Join(TestdataDir(), "test_image.png")
}

// TempDir creates a temporary directory for test artifacts.
// Returns the path and a cleanup function.
func TempDir(t testing.TB, prefix string) (string, func()) {
	dir, err := os.MkdirTemp("", "pdf-cli-test-"+prefix+"-")
	if err != nil {
		t.Fatal("failed to create temp dir: " + err.Error())
	}
	return dir, func() { _ = os.RemoveAll(dir) }
}

// TempFile creates a temporary file with the given content.
// Returns the path and a cleanup function.
func TempFile(t testing.TB, prefix, content string) (string, func()) {
	f, err := os.CreateTemp("", "pdf-cli-test-"+prefix+"-*.pdf")
	if err != nil {
		t.Fatal("failed to create temp file: " + err.Error())
	}
	if content != "" {
		if _, err := f.WriteString(content); err != nil {
			_ = f.Close()
			_ = os.Remove(f.Name())
			t.Fatal("failed to write temp file: " + err.Error())
		}
	}
	_ = f.Close()
	return f.Name(), func() { _ = os.Remove(f.Name()) }
}

// CopyFile copies a file from src to dst for test isolation.
func CopyFile(src, dst string) error {
	data, err := os.ReadFile(src) // #nosec G304 - test fixture, paths are controlled
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644) // #nosec G306 - test fixture, permissive permissions OK
}
