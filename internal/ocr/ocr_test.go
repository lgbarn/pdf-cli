package ocr

import (
	"slices"
	"testing"

	"github.com/lgbarn/pdf-cli/internal/util"
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
	tests := []struct {
		path string
		want bool
	}{
		{"image.png", true},
		{"image.PNG", true},
		{"photo.jpg", true},
		{"photo.jpeg", true},
		{"photo.JPEG", true},
		{"scan.tif", true},
		{"scan.tiff", true},
		{"scan.TIFF", true},
		{"document.pdf", false},
		{"file.txt", false},
		{"noext", false},
		{"/path/to/image.png", true},
		{"/path/to/file.doc", false},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := util.IsImageFile(tt.path); got != tt.want {
				t.Errorf("isImageFile(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
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
