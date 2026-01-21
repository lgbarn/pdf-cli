package ocr

import (
	"slices"
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
