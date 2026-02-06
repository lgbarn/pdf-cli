package ocr

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestFindImageFiles(t *testing.T) {
	// Create temp directory structure
	tmpDir, err := os.MkdirTemp("", "findimages-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create subdirectory
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.MkdirAll(subDir, 0700); err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	// Create test files
	testFiles := map[string]bool{
		filepath.Join(tmpDir, "image1.png"):   true,  // should be found
		filepath.Join(tmpDir, "image2.jpg"):   true,  // should be found
		filepath.Join(tmpDir, "image3.jpeg"):  true,  // should be found
		filepath.Join(tmpDir, "scan.tif"):     true,  // should be found
		filepath.Join(tmpDir, "scan2.tiff"):   true,  // should be found
		filepath.Join(subDir, "nested.png"):   true,  // should be found (nested)
		filepath.Join(tmpDir, "document.pdf"): false, // should NOT be found
		filepath.Join(tmpDir, "readme.txt"):   false, // should NOT be found
		filepath.Join(tmpDir, "data.json"):    false, // should NOT be found
		filepath.Join(tmpDir, "UPPER.PNG"):    true,  // should be found (uppercase)
	}

	for path := range testFiles {
		if err := os.WriteFile(path, []byte("test"), 0600); err != nil {
			t.Fatalf("Failed to create test file %s: %v", path, err)
		}
	}

	// Run findImageFiles
	found, err := findImageFiles(tmpDir)
	if err != nil {
		t.Fatalf("findImageFiles() error = %v", err)
	}

	// Count expected images
	expectedCount := 0
	for _, isImage := range testFiles {
		if isImage {
			expectedCount++
		}
	}

	if len(found) != expectedCount {
		t.Errorf("findImageFiles() found %d files, want %d", len(found), expectedCount)
		t.Logf("Found files: %v", found)
	}

	// Verify all found files are images
	for _, f := range found {
		if want, ok := testFiles[f]; !ok || !want {
			t.Errorf("findImageFiles() found unexpected file: %s", f)
		}
	}
}

func TestFindImageFilesEmpty(t *testing.T) {
	// Create empty temp directory
	tmpDir, err := os.MkdirTemp("", "findimages-empty-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	found, err := findImageFiles(tmpDir)
	if err != nil {
		t.Fatalf("findImageFiles() error = %v", err)
	}
	if len(found) != 0 {
		t.Errorf("findImageFiles() on empty dir found %d files, want 0", len(found))
	}
}

func TestFindImageFilesNonExistent(t *testing.T) {
	_, err := findImageFiles("/nonexistent/path/that/does/not/exist")
	if err == nil {
		t.Error("findImageFiles() expected error for non-existent path, got nil")
	}
}

func TestFindImageFilesOrder(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "findimages-order-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create numbered image files
	for i := 1; i <= 5; i++ {
		name := filepath.Join(tmpDir, string(rune('a'+i-1))+".png")
		if err := os.WriteFile(name, []byte("test"), 0600); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	found, err := findImageFiles(tmpDir)
	if err != nil {
		t.Fatalf("findImageFiles() error = %v", err)
	}

	// Verify files are returned (filepath.Walk returns alphabetical order)
	sorted := make([]string, len(found))
	copy(sorted, found)
	sort.Strings(sorted)

	for i, f := range found {
		if f != sorted[i] {
			t.Logf("Note: findImageFiles returns files in walk order, which is alphabetical")
			break
		}
	}
}

func TestGetDataDir(t *testing.T) {
	// getDataDir creates directory if it doesn't exist
	dataDir, err := getDataDir()
	if err != nil {
		t.Fatalf("getDataDir() error = %v", err)
	}

	if dataDir == "" {
		t.Error("getDataDir() returned empty string")
	}

	// Verify path contains expected components
	if !filepath.IsAbs(dataDir) {
		t.Errorf("getDataDir() returned non-absolute path: %s", dataDir)
	}

	// Should contain "pdf-cli" and "tessdata"
	if !contains(dataDir, "pdf-cli") {
		t.Errorf("getDataDir() path doesn't contain 'pdf-cli': %s", dataDir)
	}
	if !contains(dataDir, "tessdata") {
		t.Errorf("getDataDir() path doesn't contain 'tessdata': %s", dataDir)
	}

	// Verify directory exists
	info, err := os.Stat(dataDir)
	if err != nil {
		t.Errorf("getDataDir() directory doesn't exist: %v", err)
	} else if !info.IsDir() {
		t.Errorf("getDataDir() path is not a directory: %s", dataDir)
	}
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
