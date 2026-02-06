package ocr

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestWASMBackendName(t *testing.T) {
	w := &WASMBackend{}
	if got := w.Name(); got != "wasm" {
		t.Errorf("WASMBackend.Name() = %q, want %q", got, "wasm")
	}
}

func TestWASMBackendAvailable(t *testing.T) {
	w := &WASMBackend{}
	// WASM backend is always available
	if got := w.Available(); !got {
		t.Errorf("WASMBackend.Available() = %v, want %v", got, true)
	}
}

func TestWASMBackendClose(t *testing.T) {
	w := &WASMBackend{}
	// Close on uninitialized backend should succeed
	if err := w.Close(); err != nil {
		t.Errorf("WASMBackend.Close() error = %v", err)
	}
}

func TestNewWASMBackend(t *testing.T) {
	backend, err := NewWASMBackend("eng", "")
	if err != nil {
		t.Fatalf("NewWASMBackend() error = %v", err)
	}
	defer backend.Close()

	if backend.Name() != "wasm" {
		t.Errorf("backend.Name() = %q, want %q", backend.Name(), "wasm")
	}
	if !backend.Available() {
		t.Error("backend.Available() = false, want true")
	}
}

func TestNewWASMBackendWithDataDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "wasm-datadir-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	backend, err := NewWASMBackend("eng", tmpDir)
	if err != nil {
		t.Fatalf("NewWASMBackend() error = %v", err)
	}
	defer backend.Close()

	if backend.dataDir != tmpDir {
		t.Errorf("backend.dataDir = %q, want %q", backend.dataDir, tmpDir)
	}
}

func TestWASMBackendEnsureTessdata(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "wasm-tessdata-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create fake traineddata file
	if err := os.WriteFile(filepath.Join(tmpDir, "eng.traineddata"), []byte("fake"), 0600); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	backend := &WASMBackend{
		dataDir: tmpDir,
		lang:    "eng",
	}

	// Should succeed because file exists
	if err := backend.EnsureTessdata(context.Background(), "eng"); err != nil {
		t.Errorf("EnsureTessdata() error = %v", err)
	}

	// Test with empty lang (should use backend lang)
	if err := backend.EnsureTessdata(context.Background(), ""); err != nil {
		t.Errorf("EnsureTessdata('') error = %v", err)
	}
}

func TestWASMBackendFields(t *testing.T) {
	backend := &WASMBackend{
		dataDir: "/test/data",
		lang:    "fra+deu",
	}

	if backend.dataDir != "/test/data" {
		t.Errorf("backend.dataDir = %q, want %q", backend.dataDir, "/test/data")
	}
	if backend.lang != "fra+deu" {
		t.Errorf("backend.lang = %q, want %q", backend.lang, "fra+deu")
	}
}
