package ocr

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultLang(t *testing.T) {
	tests := []struct {
		name     string
		lang     string
		fallback string
		want     string
	}{
		{"uses lang when provided", "fra", "eng", "fra"},
		{"uses fallback when lang empty", "", "deu", "deu"},
		{"uses eng when both empty", "", "", "eng"},
		{"lang takes precedence", "spa", "", "spa"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := defaultLang(tt.lang, tt.fallback); got != tt.want {
				t.Errorf("defaultLang(%q, %q) = %q, want %q", tt.lang, tt.fallback, got, tt.want)
			}
		})
	}
}

func TestNativeBackendName(t *testing.T) {
	n := &NativeBackend{}
	if got := n.Name(); got != "native" {
		t.Errorf("NativeBackend.Name() = %q, want %q", got, "native")
	}
}

func TestNativeBackendAvailable(t *testing.T) {
	tests := []struct {
		name          string
		tesseractPath string
		want          bool
	}{
		{"available when path set", "/usr/bin/tesseract", true},
		{"unavailable when path empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &NativeBackend{tesseractPath: tt.tesseractPath}
			if got := n.Available(); got != tt.want {
				t.Errorf("NativeBackend.Available() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNativeBackendClose(t *testing.T) {
	n := &NativeBackend{}
	if err := n.Close(); err != nil {
		t.Errorf("NativeBackend.Close() error = %v", err)
	}
}

func TestNativeBackendBuildArgs(t *testing.T) {
	tests := []struct {
		name        string
		tessdataDir string
		imagePath   string
		outputBase  string
		lang        string
		wantLen     int
	}{
		{"without tessdata dir", "", "image.png", "output", "eng", 4},
		{"with tessdata dir", "/path/to/tessdata", "image.png", "output", "eng", 6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &NativeBackend{tessdataDir: tt.tessdataDir}
			args := n.buildArgs(tt.imagePath, tt.outputBase, tt.lang)
			if len(args) != tt.wantLen {
				t.Errorf("buildArgs() returned %d args, want %d", len(args), tt.wantLen)
				t.Logf("args: %v", args)
			}
		})
	}
}

func TestNativeBackendHasLanguage(t *testing.T) {
	// Create temp tessdata directory
	tmpDir, err := os.MkdirTemp("", "tessdata-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a fake eng.traineddata file
	if err := os.WriteFile(filepath.Join(tmpDir, "eng.traineddata"), []byte("fake"), 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name        string
		tessdataDir string
		lang        string
		want        bool
	}{
		{"has language", tmpDir, "eng", true},
		{"missing language", tmpDir, "fra", false},
		{"empty tessdata dir", "", "eng", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := &NativeBackend{tessdataDir: tt.tessdataDir}
			if got := n.HasLanguage(tt.lang); got != tt.want {
				t.Errorf("HasLanguage(%q) = %v, want %v", tt.lang, got, tt.want)
			}
		})
	}
}

func TestNativeBackendVersion(t *testing.T) {
	// Create a backend without a valid tesseract path
	n := &NativeBackend{tesseractPath: ""}
	version := n.Version()
	// Should return either a version or "unknown"
	if version == "" {
		t.Error("Version() returned empty string")
	}
}

func TestNewNativeBackend(t *testing.T) {
	backend, err := NewNativeBackend("eng", "")
	if err != nil {
		// Native tesseract may not be installed
		if err == ErrNativeNotFound {
			t.Log("Native Tesseract not found, skipping test")
			return
		}
		t.Logf("NewNativeBackend() error = %v", err)
		return
	}
	defer backend.Close()

	if backend.Name() != "native" {
		t.Errorf("backend.Name() = %q, want %q", backend.Name(), "native")
	}
	if !backend.Available() {
		t.Error("backend.Available() = false, want true")
	}
}

func TestNewNativeBackendWithDataDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "native-datadir-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	backend, err := NewNativeBackend("eng", tmpDir)
	if err != nil {
		if err == ErrNativeNotFound {
			t.Log("Native Tesseract not found, skipping test")
			return
		}
		t.Logf("NewNativeBackend() error = %v", err)
		return
	}
	defer backend.Close()

	if backend.tessdataDir != tmpDir {
		t.Errorf("backend.tessdataDir = %q, want %q", backend.tessdataDir, tmpDir)
	}
}

func TestNativeBackendProcessImageWithCancelledContext(t *testing.T) {
	backend, err := NewNativeBackend("eng", "")
	if err != nil {
		if err == ErrNativeNotFound {
			t.Log("Native Tesseract not found, skipping test")
			return
		}
		t.Logf("NewNativeBackend() error = %v", err)
		return
	}
	defer backend.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err = backend.ProcessImage(ctx, "/nonexistent/image.png", "eng")
	if err == nil {
		t.Error("ProcessImage() expected error with canceled context")
	}
}
