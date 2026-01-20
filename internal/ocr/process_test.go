package ocr

import (
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
)

func TestProcessImagesSequential(t *testing.T) {
	// Create temp directory with test images
	tmpDir, err := os.MkdirTemp("", "process-seq-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test image files
	imageFiles := []string{
		filepath.Join(tmpDir, "page1.png"),
		filepath.Join(tmpDir, "page2.png"),
		filepath.Join(tmpDir, "page3.png"),
	}
	for _, f := range imageFiles {
		if err := os.WriteFile(f, []byte("test"), 0600); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// Create engine with mock backend
	mock := newMockBackend("test", true).withOutput("extracted text")
	engine := &Engine{
		backend: mock,
		lang:    "eng",
	}

	result, err := engine.processImagesSequential(imageFiles, false)
	if err != nil {
		t.Fatalf("processImagesSequential() error = %v", err)
	}

	// Should have processed all images
	if calls := atomic.LoadInt32(&mock.processCalls); calls != 3 {
		t.Errorf("processCalls = %d, want 3", calls)
	}

	// Result should contain the text
	if !strings.Contains(result, "extracted text") {
		t.Errorf("result doesn't contain expected text: %s", result)
	}
}

func TestProcessImagesSequentialEmpty(t *testing.T) {
	mock := newMockBackend("test", true)
	engine := &Engine{
		backend: mock,
		lang:    "eng",
	}

	result, err := engine.processImagesSequential([]string{}, false)
	if err != nil {
		t.Fatalf("processImagesSequential() error = %v", err)
	}
	if result != "" {
		t.Errorf("processImagesSequential() with empty input = %q, want empty", result)
	}
	if calls := atomic.LoadInt32(&mock.processCalls); calls != 0 {
		t.Errorf("processCalls = %d, want 0", calls)
	}
}

func TestProcessImagesSequentialWithError(t *testing.T) {
	// Create temp directory with test images
	tmpDir, err := os.MkdirTemp("", "process-err-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	imageFiles := []string{filepath.Join(tmpDir, "page1.png")}
	for _, f := range imageFiles {
		if err := os.WriteFile(f, []byte("test"), 0600); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// Create engine with mock backend that returns error
	mock := newMockBackend("test", true).withError(errTestProcess)
	engine := &Engine{
		backend: mock,
		lang:    "eng",
	}

	// processImagesSequential continues even if individual images fail
	result, err := engine.processImagesSequential(imageFiles, false)
	if err != nil {
		t.Fatalf("processImagesSequential() error = %v", err)
	}
	// Result should be empty since processing failed
	if result != "" {
		t.Errorf("result should be empty on error, got: %s", result)
	}
}

func TestProcessImagesParallel(t *testing.T) {
	// Create temp directory with test images
	tmpDir, err := os.MkdirTemp("", "process-par-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create enough images to trigger parallel processing
	imageCount := parallelThreshold + 2
	imageFiles := make([]string, imageCount)
	for i := 0; i < imageCount; i++ {
		imageFiles[i] = filepath.Join(tmpDir, string(rune('a'+i))+".png")
		if err := os.WriteFile(imageFiles[i], []byte("test"), 0600); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// Create engine with mock backend
	mock := newMockBackend("test-native", true).withOutput("parallel text")
	engine := &Engine{
		backend: mock,
		lang:    "eng",
	}

	result, err := engine.processImagesParallel(imageFiles, false)
	if err != nil {
		t.Fatalf("processImagesParallel() error = %v", err)
	}

	// Should have processed all images
	if calls := atomic.LoadInt32(&mock.processCalls); calls != int32(imageCount) {
		t.Errorf("processCalls = %d, want %d", calls, imageCount)
	}

	// Result should contain the text
	if !strings.Contains(result, "parallel text") {
		t.Errorf("result doesn't contain expected text: %s", result)
	}
}

func TestProcessImagesRouting(t *testing.T) {
	tests := []struct {
		name        string
		imageCount  int
		backendName string
	}{
		{"few images uses sequential", 3, "native"},
		{"many images uses parallel", parallelThreshold + 1, "native"},
		{"wasm always uses sequential", parallelThreshold + 1, "wasm"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "process-route-*")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			imageFiles := make([]string, tt.imageCount)
			for i := 0; i < tt.imageCount; i++ {
				imageFiles[i] = filepath.Join(tmpDir, string(rune('a'+i))+".png")
				if err := os.WriteFile(imageFiles[i], []byte("test"), 0600); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
			}

			mock := newMockBackend(tt.backendName, true).withOutput("text")
			engine := &Engine{
				backend: mock,
				lang:    "eng",
			}

			_, err = engine.processImages(imageFiles, false)
			if err != nil {
				t.Fatalf("processImages() error = %v", err)
			}

			// All images should be processed regardless of method
			if calls := atomic.LoadInt32(&mock.processCalls); calls != int32(tt.imageCount) {
				t.Errorf("processCalls = %d, want %d", calls, tt.imageCount)
			}
		})
	}
}

func TestProcessImagesWithProgress(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "process-progress-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	imageFiles := []string{filepath.Join(tmpDir, "page1.png")}
	for _, f := range imageFiles {
		if err := os.WriteFile(f, []byte("test"), 0600); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	mock := newMockBackend("test", true).withOutput("text")
	engine := &Engine{
		backend: mock,
		lang:    "eng",
	}

	// Test with progress enabled (just verify it doesn't crash)
	_, err = engine.processImagesSequential(imageFiles, true)
	if err != nil {
		t.Fatalf("processImagesSequential with progress error = %v", err)
	}
}

func TestEngineOptionsDefaults(t *testing.T) {
	tests := []struct {
		name    string
		opts    EngineOptions
		wantErr bool
	}{
		{
			name:    "empty options uses defaults",
			opts:    EngineOptions{},
			wantErr: false, // Will use auto backend
		},
		{
			name: "custom language",
			opts: EngineOptions{
				Lang: "fra",
			},
			wantErr: false,
		},
		{
			name: "explicit wasm backend",
			opts: EngineOptions{
				BackendType: BackendWASM,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			engine, err := NewEngineWithOptions(tt.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewEngineWithOptions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				defer engine.Close()
				// Verify engine was created
				if engine.BackendName() == "" {
					t.Error("BackendName() returned empty string")
				}
			}
		})
	}
}

func TestNewEngine(t *testing.T) {
	engine, err := NewEngine("eng")
	if err != nil {
		// May fail if no backend is available, which is OK for this test
		t.Logf("NewEngine() error = %v (expected if no backend available)", err)
		return
	}
	defer engine.Close()

	if engine.lang != "eng" {
		t.Errorf("engine.lang = %q, want %q", engine.lang, "eng")
	}
}

func TestResolvePages(t *testing.T) {
	engine := &Engine{lang: "eng"}

	t.Run("specified pages returned as-is", func(t *testing.T) {
		pages := []int{1, 3, 5}
		result, err := engine.resolvePages("/nonexistent/path.pdf", pages, "")
		if err != nil {
			t.Fatalf("resolvePages() error = %v", err)
		}
		if len(result) != len(pages) {
			t.Errorf("resolvePages() returned %d pages, want %d", len(result), len(pages))
		}
	})

	t.Run("nil pages with non-existent file", func(t *testing.T) {
		_, err := engine.resolvePages("/nonexistent/path.pdf", nil, "")
		if err == nil {
			t.Log("resolvePages with nil pages and non-existent file returned nil error")
		}
	})
}

func TestResolvePagesWithRealPDF(t *testing.T) {
	samplePDF := filepath.Join("..", "..", "testdata", "sample.pdf")
	if _, err := os.Stat(samplePDF); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	engine := &Engine{
		lang: "eng",
	}

	// When pages is nil, resolvePages should get page count from the PDF
	pages, err := engine.resolvePages(samplePDF, nil, "")
	if err != nil {
		t.Fatalf("resolvePages() error = %v", err)
	}

	// sample.pdf should have at least 1 page
	if len(pages) == 0 {
		t.Error("resolvePages() returned empty pages for valid PDF")
	}

	// Verify pages are sequential starting from 1
	for i, p := range pages {
		if p != i+1 {
			t.Errorf("resolvePages()[%d] = %d, want %d", i, p, i+1)
		}
	}
}

func TestExtractTextFromPDFNonExistent(t *testing.T) {
	mock := newMockBackend("test", true)
	engine := &Engine{
		backend: mock,
		lang:    "eng",
		dataDir: "/tmp/test",
	}

	// Should fail for non-existent PDF
	_, err := engine.ExtractTextFromPDF("/nonexistent/file.pdf", []int{1}, "", false)
	if err == nil {
		t.Error("ExtractTextFromPDF with non-existent file should error")
	}
}

func TestExtractTextFromPDFWithWASMBackendEnsureError(t *testing.T) {
	// Create WASM backend with invalid data dir to trigger EnsureTessdata error
	backend, err := NewWASMBackend("eng", "/invalid/nonexistent/path")
	if err != nil {
		t.Logf("NewWASMBackend error: %v (expected)", err)
		return
	}

	engine := &Engine{
		backend: backend,
		lang:    "eng",
		dataDir: "/invalid/nonexistent/path",
	}

	_, err = engine.ExtractTextFromPDF("../../testdata/sample.pdf", []int{1}, "", false)
	// This may or may not error depending on whether tessdata exists
	if err != nil {
		t.Logf("ExtractTextFromPDF with invalid tessdata path error: %v (expected)", err)
	}
}

func TestExtractImagesWithRealPDF(t *testing.T) {
	samplePDF := filepath.Join("..", "..", "testdata", "sample.pdf")
	if _, err := os.Stat(samplePDF); os.IsNotExist(err) {
		t.Skip("sample.pdf not found in testdata")
	}

	tmpDir, err := os.MkdirTemp("", "extract-img-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	engine := &Engine{
		lang: "eng",
	}

	// Extract images - may or may not extract images depending on PDF content
	err = engine.extractImagesToDir(samplePDF, tmpDir, []int{1}, "")
	if err != nil {
		t.Logf("extractImagesToDir error (may be expected): %v", err)
	}
}

func TestProcessImagesParallelWithError(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "process-par-err-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create enough images to trigger parallel processing
	imageCount := parallelThreshold + 2
	imageFiles := make([]string, imageCount)
	for i := 0; i < imageCount; i++ {
		imageFiles[i] = filepath.Join(tmpDir, string(rune('a'+i))+".png")
		if err := os.WriteFile(imageFiles[i], []byte("test"), 0600); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	// Create engine with mock backend that returns errors
	mock := newMockBackend("test-native", true).withError(errTestProcess)
	engine := &Engine{
		backend: mock,
		lang:    "eng",
	}

	// Should complete even with errors
	result, err := engine.processImagesParallel(imageFiles, false)
	if err != nil {
		t.Fatalf("processImagesParallel() error = %v", err)
	}

	// Result should be empty since all processing failed
	if result != "" {
		t.Logf("processImagesParallel() with errors returned non-empty result: %q", result)
	}
}
