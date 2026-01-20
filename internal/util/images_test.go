package util

import "testing"

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
		{"", false},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := IsImageFile(tt.path); got != tt.want {
				t.Errorf("IsImageFile(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}
