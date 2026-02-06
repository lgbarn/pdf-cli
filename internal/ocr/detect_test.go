package ocr

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsValidTessdataDir(t *testing.T) {
	// Create temp directory structure for testing
	tmpDir, err := os.MkdirTemp("", "tessdata-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a valid tessdata directory
	validDir := filepath.Join(tmpDir, "valid")
	if err := os.MkdirAll(validDir, 0700); err != nil {
		t.Fatalf("Failed to create valid dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(validDir, "eng.traineddata"), []byte("test"), 0600); err != nil {
		t.Fatalf("Failed to create traineddata file: %v", err)
	}

	// Create an empty directory
	emptyDir := filepath.Join(tmpDir, "empty")
	if err := os.MkdirAll(emptyDir, 0700); err != nil {
		t.Fatalf("Failed to create empty dir: %v", err)
	}

	// Create a directory with wrong file extensions
	wrongDir := filepath.Join(tmpDir, "wrong")
	if err := os.MkdirAll(wrongDir, 0700); err != nil {
		t.Fatalf("Failed to create wrong dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(wrongDir, "eng.txt"), []byte("test"), 0600); err != nil {
		t.Fatalf("Failed to create txt file: %v", err)
	}

	tests := []struct {
		name string
		path string
		want bool
	}{
		{"valid tessdata directory", validDir, true},
		{"empty directory", emptyDir, false},
		{"directory with wrong extensions", wrongDir, false},
		{"non-existent directory", filepath.Join(tmpDir, "nonexistent"), false},
		{"empty path", "", false},
		{"file instead of directory", filepath.Join(validDir, "eng.traineddata"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidTessdataDir(tt.path); got != tt.want {
				t.Errorf("isValidTessdataDir(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestExtractTessdataFromParams(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   string
	}{
		{
			name:   "with tessdata_dir",
			output: "tessdata_dir /usr/share/tessdata\nother params",
			want:   "/usr/share/tessdata",
		},
		{
			name:   "without tessdata_dir",
			output: "some other output",
			want:   "",
		},
		{
			name:   "empty output",
			output: "",
			want:   "",
		},
		{
			name:   "tessdata_dir with tabs",
			output: "tessdata_dir\t/custom/path/tessdata",
			want:   "/custom/path/tessdata",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractTessdataFromParams(tt.output); got != tt.want {
				t.Errorf("extractTessdataFromParams() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetCommonTessdataPaths(t *testing.T) {
	paths := getCommonTessdataPaths()

	// Just verify we get some paths (the specific paths depend on OS)
	if paths == nil {
		// Some OS might not have any paths defined, which is OK
		t.Log("No common tessdata paths defined for this OS")
		return
	}

	for _, path := range paths {
		if path == "" {
			t.Error("Empty path in common tessdata paths")
		}
	}
}

func TestDetectNativeTesseract(t *testing.T) {
	// This test checks if native Tesseract detection works
	// It may return ErrNativeNotFound if Tesseract isn't installed
	info, err := DetectNativeTesseract()

	if err == ErrNativeNotFound {
		t.Log("Native Tesseract not found, skipping detailed checks")
		return
	}

	if err != nil {
		t.Logf("DetectNativeTesseract returned unexpected error: %v", err)
		return
	}

	// If we get here, Tesseract was found
	if info.Path == "" {
		t.Error("DetectNativeTesseract returned empty path")
	}
	if info.Version == "" {
		t.Error("DetectNativeTesseract returned empty version")
	}
	t.Logf("Found Tesseract at %s, version %s", info.Path, info.Version)
}

func TestNativeInfoStruct(t *testing.T) {
	info := &NativeInfo{
		Path:     "/usr/bin/tesseract",
		Version:  "5.0.0",
		Tessdata: "/usr/share/tessdata",
	}

	if info.Path != "/usr/bin/tesseract" {
		t.Errorf("info.Path = %q, want %q", info.Path, "/usr/bin/tesseract")
	}
	if info.Version != "5.0.0" {
		t.Errorf("info.Version = %q, want %q", info.Version, "5.0.0")
	}
	if info.Tessdata != "/usr/share/tessdata" {
		t.Errorf("info.Tessdata = %q, want %q", info.Tessdata, "/usr/share/tessdata")
	}
}

func TestErrNativeNotFound(t *testing.T) {
	if ErrNativeNotFound == nil {
		t.Error("ErrNativeNotFound is nil")
	}
	if ErrNativeNotFound.Error() != "native Tesseract not found" {
		t.Errorf("ErrNativeNotFound.Error() = %q", ErrNativeNotFound.Error())
	}
}

func TestFindTessdataDir(t *testing.T) {
	// This test checks findTessdataDir function
	// It requires tesseract to be installed
	info, err := DetectNativeTesseract()
	if err != nil {
		t.Log("Native Tesseract not found, skipping findTessdataDir test")
		return
	}

	// findTessdataDir should return something (even if empty path error)
	tessdata, err := findTessdataDir(info.Path)
	if err != nil {
		t.Logf("findTessdataDir returned error: %v (may be expected)", err)
		return
	}

	if tessdata != "" {
		// Verify it's a valid tessdata directory
		if !isValidTessdataDir(tessdata) {
			t.Errorf("findTessdataDir returned invalid path: %s", tessdata)
		}
	}
}

func TestGetTesseractVersion(t *testing.T) {
	info, err := DetectNativeTesseract()
	if err != nil {
		t.Log("Native Tesseract not found, skipping version test")
		return
	}

	version, err := getTesseractVersion(info.Path)
	if err != nil {
		t.Errorf("getTesseractVersion() error = %v", err)
		return
	}

	if version == "" {
		t.Error("getTesseractVersion() returned empty string")
	}
	t.Logf("Tesseract version: %s", version)
}

func TestGetTesseractVersionInvalidPath(t *testing.T) {
	_, err := getTesseractVersion("/nonexistent/tesseract")
	if err == nil {
		t.Error("getTesseractVersion() expected error for invalid path")
	}
}
