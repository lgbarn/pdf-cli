package util

import (
	"path/filepath"
	"strings"
)

// SupportedImageExtensions contains all supported image file extensions.
var SupportedImageExtensions = []string{".png", ".jpg", ".jpeg", ".tif", ".tiff"}

// IsImageFile checks if a file has a supported image extension.
func IsImageFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	for _, supported := range SupportedImageExtensions {
		if ext == supported {
			return true
		}
	}
	return false
}
