package ocr

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// parallelThreshold is an alias for test compatibility
const parallelThreshold = DefaultParallelThreshold

func TestEnsureTessdata(t *testing.T) {
	// Create a temp directory for tessdata
	tmpDir, err := os.MkdirTemp("", "tessdata-ensure-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a fake traineddata file so we don't try to download
	if err := os.WriteFile(filepath.Join(tmpDir, "eng.traineddata"), []byte("fake"), 0600); err != nil {
		t.Fatalf("Failed to create fake traineddata: %v", err)
	}

	engine := &Engine{
		dataDir: tmpDir,
		lang:    "eng",
	}

	// Should succeed since file exists
	if err := engine.EnsureTessdata(context.Background()); err != nil {
		t.Errorf("EnsureTessdata() error = %v", err)
	}
}

func TestEnsureTessdataMultipleLanguages(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tessdata-multi-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create fake traineddata files for multiple languages
	for _, lang := range []string{"eng", "fra", "deu"} {
		if err := os.WriteFile(filepath.Join(tmpDir, lang+".traineddata"), []byte("fake"), 0600); err != nil {
			t.Fatalf("Failed to create fake traineddata: %v", err)
		}
	}

	engine := &Engine{
		dataDir: tmpDir,
		lang:    "eng+fra+deu",
	}

	if err := engine.EnsureTessdata(context.Background()); err != nil {
		t.Errorf("EnsureTessdata() error = %v", err)
	}
}

func TestSelectBackendNative(t *testing.T) {
	engine := &Engine{
		dataDir:     "",
		lang:        "eng",
		backendType: BackendNative,
	}

	backend, err := engine.selectBackend()
	if err != nil {
		// Native backend might not be available
		t.Logf("Native backend not available: %v", err)
		return
	}
	defer backend.Close()

	if backend.Name() != "native" {
		t.Errorf("selectBackend() returned %q, want %q", backend.Name(), "native")
	}
}

func TestSelectBackendWASM(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "wasm-backend-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	engine := &Engine{
		dataDir:     tmpDir,
		lang:        "eng",
		backendType: BackendWASM,
	}

	backend, err := engine.selectBackend()
	if err != nil {
		t.Fatalf("selectBackend() error = %v", err)
	}
	defer backend.Close()

	if backend.Name() != "wasm" {
		t.Errorf("selectBackend() returned %q, want %q", backend.Name(), "wasm")
	}
}

func TestSelectBackendAuto(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "auto-backend-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	engine := &Engine{
		dataDir:     tmpDir,
		lang:        "eng",
		backendType: BackendAuto,
	}

	backend, err := engine.selectBackend()
	if err != nil {
		t.Fatalf("selectBackend() error = %v", err)
	}
	defer backend.Close()

	// Auto will pick native if available, otherwise wasm
	name := backend.Name()
	if name != "native" && name != "wasm" {
		t.Errorf("selectBackend() returned %q, want 'native' or 'wasm'", name)
	}
}

func TestExtractImagesToDir(t *testing.T) {
	// Create temp directory for output
	tmpDir, err := os.MkdirTemp("", "extract-images-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	engine := &Engine{
		lang: "eng",
	}

	// Test with a non-existent PDF - should return error
	err = engine.extractImagesToDir("/nonexistent/file.pdf", tmpDir, []int{1}, "")
	if err == nil {
		t.Error("extractImagesToDir() expected error for non-existent file")
	}
}

func TestResolvePagesEmpty(t *testing.T) {
	engine := &Engine{
		lang: "eng",
	}

	// With empty pages and non-existent PDF, should return error
	_, err := engine.resolvePages("/nonexistent/file.pdf", nil, "")
	if err == nil {
		t.Error("resolvePages() expected error for non-existent file with empty pages")
	}
}

func TestImageResultStruct(t *testing.T) {
	// Just verify the struct works as expected
	result := imageResult{
		index: 5,
		text:  "sample text",
	}
	if result.index != 5 {
		t.Errorf("imageResult.index = %d, want 5", result.index)
	}
	if result.text != "sample text" {
		t.Errorf("imageResult.text = %q, want %q", result.text, "sample text")
	}
}

func TestParallelThreshold(t *testing.T) {
	// Verify the constant is set appropriately
	if parallelThreshold <= 0 {
		t.Errorf("parallelThreshold = %d, want > 0", parallelThreshold)
	}
	if parallelThreshold > 100 {
		t.Errorf("parallelThreshold = %d, seems too high", parallelThreshold)
	}
}

func TestEngineFields(t *testing.T) {
	engine := &Engine{
		dataDir:     "/test/data",
		lang:        "eng+fra",
		backendType: BackendNative,
		backend:     newMockBackend("test", true),
	}

	if engine.dataDir != "/test/data" {
		t.Errorf("engine.dataDir = %q, want %q", engine.dataDir, "/test/data")
	}
	if engine.lang != "eng+fra" {
		t.Errorf("engine.lang = %q, want %q", engine.lang, "eng+fra")
	}
	if engine.backendType != BackendNative {
		t.Errorf("engine.backendType = %v, want %v", engine.backendType, BackendNative)
	}
	if engine.BackendName() != "test" {
		t.Errorf("engine.BackendName() = %q, want %q", engine.BackendName(), "test")
	}
}
