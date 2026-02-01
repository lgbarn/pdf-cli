package fileio

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	// DefaultDirPerm is the default permission for creating directories.
	DefaultDirPerm = 0750

	// DefaultFilePerm is the default permission for creating files.
	DefaultFilePerm = 0600
)

// FileExists checks if a file exists
func FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// IsDir checks if a path is a directory
func IsDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// EnsureDir creates a directory if it doesn't exist
func EnsureDir(path string) error {
	return os.MkdirAll(path, DefaultDirPerm)
}

// EnsureParentDir creates the parent directory of a file path if it doesn't exist
func EnsureParentDir(path string) error {
	dir := filepath.Dir(path)
	if dir == "." || dir == "/" {
		return nil
	}
	return EnsureDir(dir)
}

// AtomicWrite writes data to a file atomically by writing to a temp file first
func AtomicWrite(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := EnsureDir(dir); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Create temp file in the same directory
	tmpFile, err := os.CreateTemp(dir, ".pdf-cli-tmp-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()

	// Clean up temp file on error
	defer func() {
		if tmpFile != nil {
			_ = tmpFile.Close()
			_ = os.Remove(tmpPath)
		}
	}()

	// Write data to temp file
	if _, err := tmpFile.Write(data); err != nil {
		return fmt.Errorf("failed to write data: %w", err)
	}

	if err := tmpFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync file: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}
	tmpFile = nil // Prevent defer from closing again

	// Rename temp file to target (atomic on most filesystems)
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

// CopyFile copies a file from src to dst
func CopyFile(src, dst string) error {
	// Sanitize paths to prevent directory traversal
	cleanSrc, err := SanitizePath(src)
	if err != nil {
		return fmt.Errorf("invalid source path: %w", err)
	}
	cleanDst, err := SanitizePath(dst)
	if err != nil {
		return fmt.Errorf("invalid destination path: %w", err)
	}

	srcFile, err := os.Open(cleanSrc) // #nosec G304 -- path is sanitized
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer func() { _ = srcFile.Close() }()

	if err := EnsureParentDir(cleanDst); err != nil {
		return err
	}

	dstFile, err := os.Create(cleanDst) // #nosec G304 -- path is sanitized
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer func() { _ = dstFile.Close() }()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	return dstFile.Sync()
}

// ValidatePDFFile checks if a file exists and has a .pdf extension
func ValidatePDFFile(path string) error {
	if !FileExists(path) {
		return fmt.Errorf("file not found: %s", path)
	}

	ext := filepath.Ext(path)
	if ext != ".pdf" && ext != ".PDF" {
		return fmt.Errorf("not a PDF file: %s", path)
	}

	return nil
}

// ValidatePDFFiles validates multiple PDF files in parallel
func ValidatePDFFiles(paths []string) error {
	if len(paths) == 0 {
		return nil
	}

	// For small number of files, use sequential validation
	if len(paths) <= 3 {
		for _, path := range paths {
			if err := ValidatePDFFile(path); err != nil {
				return err
			}
		}
		return nil
	}

	// For larger number of files, validate in parallel
	type result struct {
		path string
		err  error
	}

	results := make(chan result, len(paths))

	for _, path := range paths {
		go func(p string) {
			results <- result{path: p, err: ValidatePDFFile(p)}
		}(path)
	}

	// Collect results
	var firstErr error
	for range paths {
		r := <-results
		if r.err != nil && firstErr == nil {
			firstErr = r.err
		}
	}

	return firstErr
}

// GenerateOutputFilename generates an output filename based on the input and a suffix
func GenerateOutputFilename(input, suffix string) string {
	ext := filepath.Ext(input)
	base := input[:len(input)-len(ext)]
	return base + suffix + ext
}

// GetFileSize returns the size of a file in bytes
func GetFileSize(path string) (int64, error) {
	info, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// FormatFileSize formats a file size in bytes to a human-readable string
func FormatFileSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

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

// SanitizePath cleans a file path and validates it against directory traversal attacks.
// It returns an error if the cleaned path still contains ".." components.
// This prevents attacks like "../../etc/passwd" from accessing unintended files.
//
// Special cases:
//   - stdin marker "-" is always allowed and returned as-is
//   - Absolute paths are allowed after validation
//   - Relative paths are allowed if they don't contain ".." after cleaning
func SanitizePath(path string) (string, error) {
	if path == "-" {
		return path, nil
	}

	// Check for ".." components in the original path before cleaning.
	// filepath.Clean resolves ".." in absolute paths (e.g., /tmp/../../etc -> /etc),
	// so checking only the cleaned result would miss traversal attempts.
	for _, part := range strings.Split(path, "/") {
		if part == ".." {
			return "", fmt.Errorf("path contains directory traversal: %s", path)
		}
	}

	cleaned := filepath.Clean(path)
	return cleaned, nil
}

// SanitizePaths validates multiple paths and returns cleaned versions.
func SanitizePaths(paths []string) ([]string, error) {
	cleaned := make([]string, len(paths))
	for i, path := range paths {
		clean, err := SanitizePath(path)
		if err != nil {
			return nil, err
		}
		cleaned[i] = clean
	}
	return cleaned, nil
}
