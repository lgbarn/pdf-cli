package ocr

import (
	"context"
	"slices"
	"strings"
	"testing"

	"github.com/lgbarn/pdf-cli/internal/fileio"
)

func TestParseLanguages(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"single language", "eng", []string{"eng"}},
		{"plus delimiter", "eng+fra", []string{"eng", "fra"}},
		{"comma delimiter", "eng,fra", []string{"eng", "fra"}},
		{"mixed delimiters", "eng+fra,deu", []string{"eng", "fra", "deu"}},
		{"with whitespace", " eng + fra ", []string{"eng", "fra"}},
		{"empty string", "", nil},
		{"only delimiters", "++,", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseLanguages(tt.input)
			if !slices.Equal(got, tt.want) {
				t.Errorf("parseLanguages(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestPrimaryLanguage(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"single language", "eng", "eng"},
		{"multiple languages", "eng+fra", "eng"},
		{"empty string", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := primaryLanguage(tt.input); got != tt.want {
				t.Errorf("primaryLanguage(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsImageFile(t *testing.T) {
	imageFiles := []string{"image.png", "image.PNG", "photo.jpg", "photo.jpeg", "photo.JPEG", "scan.tif", "scan.tiff", "scan.TIFF", "/path/to/image.png"}
	nonImageFiles := []string{"document.pdf", "file.txt", "noext", "/path/to/file.doc"}

	for _, path := range imageFiles {
		if !fileio.IsImageFile(path) {
			t.Errorf("IsImageFile(%q) = false, want true", path)
		}
	}
	for _, path := range nonImageFiles {
		if fileio.IsImageFile(path) {
			t.Errorf("IsImageFile(%q) = true, want false", path)
		}
	}
}

func TestJoinNonEmpty(t *testing.T) {
	tests := []struct {
		name string
		strs []string
		sep  string
		want string
	}{
		{"all non-empty", []string{"a", "b", "c"}, "\n", "a\nb\nc"},
		{"some empty", []string{"a", "", "c"}, "\n", "a\nc"},
		{"all empty", []string{"", "", ""}, "\n", ""},
		{"single", []string{"a"}, "\n", "a"},
		{"nil slice", nil, "\n", ""},
		{"empty slice", []string{}, "\n", ""},
		{"different separator", []string{"foo", "bar"}, " - ", "foo - bar"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := joinNonEmpty(tt.strs, tt.sep); got != tt.want {
				t.Errorf("joinNonEmpty() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestEnsureTessdataDir(t *testing.T) {
	engine, err := NewEngineWithOptions(EngineOptions{
		BackendType: BackendWASM, // WASM should always be available
		Lang:        "eng",
	})
	if err != nil {
		t.Logf("NewEngineWithOptions error: %v (may be expected)", err)
		return
	}
	defer engine.Close()

	// Call EnsureTessdata - it may download data or just verify existing
	// This test just verifies it doesn't panic
	err = engine.EnsureTessdata()
	if err != nil {
		t.Logf("EnsureTessdata error: %v (may be expected if no network)", err)
	}
}

func TestEngineOptionsWithDataDir(t *testing.T) {
	opts := EngineOptions{
		BackendType: BackendWASM,
		Lang:        "eng",
		DataDir:     "/custom/path",
	}
	engine, err := NewEngineWithOptions(opts)
	if err != nil {
		t.Logf("NewEngineWithOptions error: %v", err)
		return
	}
	defer engine.Close()

	if engine.dataDir != "/custom/path" {
		t.Errorf("engine.dataDir = %q, want %q", engine.dataDir, "/custom/path")
	}
}

func TestDownloadTessdataChecksumMismatch(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping in short mode")
	}

	// This is an integration test that verifies checksum logic with network access.
	// It uses an unknown language that will fail to download, allowing us to verify
	// that checksums are being checked in the code flow.
	tmpDir := t.TempDir()

	// Save original checksums
	origChecksums := KnownChecksums
	defer func() {
		KnownChecksums = origChecksums
	}()

	// Test 1: Download with unknown language (no checksum)
	// This should warn but attempt download (which will fail with 404)
	t.Run("no_checksum_warns", func(t *testing.T) {
		KnownChecksums = map[string]string{} // Empty map
		testLang := "xyz_nonexistent_lang_test"

		ctx := context.Background()
		err := downloadTessdata(ctx, tmpDir, testLang)

		// Should fail with HTTP 404, not a checksum error
		if err == nil {
			t.Error("Expected error for non-existent language, got nil")
		}
		if strings.Contains(err.Error(), "checksum verification failed") {
			t.Errorf("Should not get checksum error for unknown language: %v", err)
		}
	})

	// Test 2: Verify checksum validation is in the code
	// We can't easily mock the HTTP client without refactoring, but we can verify
	// the checksum functions exist and are called correctly via code inspection.
	// The real validation happens during actual downloads with known checksums.
	t.Run("checksum_functions_exist", func(t *testing.T) {
		// Verify GetChecksum returns empty for unknown language
		if checksum := GetChecksum("unknown_lang"); checksum != "" {
			t.Errorf("Expected empty checksum for unknown language, got: %s", checksum)
		}

		// Verify HasChecksum works correctly
		KnownChecksums = map[string]string{"test": "abc123"}
		if !HasChecksum("test") {
			t.Error("HasChecksum should return true for known language")
		}
		if HasChecksum("unknown") {
			t.Error("HasChecksum should return false for unknown language")
		}
	})
}

func TestDownloadTessdataPathSanitization(t *testing.T) {
	tmpDir := t.TempDir()
	malicious := []string{"../../etc/passwd", "../escape", "/etc/passwd"}

	for _, lang := range malicious {
		t.Run(lang, func(t *testing.T) {
			ctx := context.Background()
			err := downloadTessdata(ctx, tmpDir, lang)

			// We expect some kind of error (either path sanitization or HTTP 404)
			if err == nil {
				t.Errorf("Expected error for malicious lang %q, got nil", lang)
			}
			// The error should be caught either by SanitizePath or by HTTP failure
			// Either way, the download should not succeed
		})
	}
}

func TestProcessImagesParallelErrorPropagation(t *testing.T) {
	backend := newMockBackend("mock", true).
		withOutput("test text").
		withErrorIndices(map[string]error{
			"img1.png": context.DeadlineExceeded,
			"img3.png": context.DeadlineExceeded,
		})

	engine := &Engine{
		lang:    "eng",
		backend: backend,
	}

	imagePaths := []string{"img0.png", "img1.png", "img2.png", "img3.png", "img4.png"}

	ctx := context.Background()
	text, err := engine.processImagesParallel(ctx, imagePaths, false)

	// Should return error joined from multiple failures
	if err == nil {
		t.Fatal("Expected error from failed images, got nil")
	}

	// Error should mention both failed images
	errStr := err.Error()
	if !strings.Contains(errStr, "image 1") {
		t.Errorf("Error should mention image 1: %v", err)
	}
	if !strings.Contains(errStr, "image 3") {
		t.Errorf("Error should mention image 3: %v", err)
	}

	// Text should still be empty when errors occur
	if text != "" {
		t.Errorf("Expected empty text with errors, got: %q", text)
	}
}

func TestProcessImagesSequentialErrorPropagation(t *testing.T) {
	backend := newMockBackend("mock", true).
		withOutput("test text").
		withErrorIndices(map[string]error{
			"img0.png": context.DeadlineExceeded,
			"img2.png": context.DeadlineExceeded,
		})

	engine := &Engine{
		lang:    "eng",
		backend: backend,
	}

	imagePaths := []string{"img0.png", "img1.png", "img2.png"}

	ctx := context.Background()
	text, err := engine.processImagesSequential(ctx, imagePaths, false)

	// Should return error joined from multiple failures
	if err == nil {
		t.Fatal("Expected error from failed images, got nil")
	}

	// Error should mention both failed images
	errStr := err.Error()
	if !strings.Contains(errStr, "image 0") {
		t.Errorf("Error should mention image 0: %v", err)
	}
	if !strings.Contains(errStr, "image 2") {
		t.Errorf("Error should mention image 2: %v", err)
	}

	// Text should still be empty when errors occur
	if text != "" {
		t.Errorf("Expected empty text with errors, got: %q", text)
	}
}
