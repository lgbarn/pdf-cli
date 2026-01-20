package ocr

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// NativeBackend implements Backend using system-installed Tesseract.
type NativeBackend struct {
	tesseractPath string
	tessdataDir   string
	lang          string
}

// NewNativeBackend creates a new native Tesseract backend.
func NewNativeBackend(lang, dataDir string) (*NativeBackend, error) {
	info, err := DetectNativeTesseract()
	if err != nil {
		return nil, err
	}

	tessdata := dataDir
	if tessdata == "" {
		tessdata = info.Tessdata
	}

	return &NativeBackend{
		tesseractPath: info.Path,
		tessdataDir:   tessdata,
		lang:          lang,
	}, nil
}

func (n *NativeBackend) Name() string {
	return "native"
}

func (n *NativeBackend) Available() bool {
	return n.tesseractPath != ""
}

func (n *NativeBackend) ProcessImage(ctx context.Context, imagePath, lang string) (string, error) {
	lang = defaultLang(lang, n.lang)

	tmpFile, err := os.CreateTemp("", "ocr-output-*.txt")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	_ = tmpFile.Close()
	defer os.Remove(tmpPath)

	// Tesseract adds .txt extension automatically
	outputBase := strings.TrimSuffix(tmpPath, ".txt")
	resultPath := outputBase + ".txt"
	defer os.Remove(resultPath)

	args := n.buildArgs(imagePath, outputBase, lang)

	cmd := exec.CommandContext(ctx, n.tesseractPath, args...) // #nosec G204 -- tesseractPath from exec.LookPath, args are controlled
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("tesseract failed: %w (output: %s)", err, string(output))
	}

	text, err := os.ReadFile(resultPath) // #nosec G304 -- path is within temp directory we control
	if err != nil {
		return "", fmt.Errorf("failed to read OCR output: %w", err)
	}

	return strings.TrimSpace(string(text)), nil
}

func (n *NativeBackend) buildArgs(imagePath, outputBase, lang string) []string {
	var args []string
	if n.tessdataDir != "" {
		args = append(args, "--tessdata-dir", n.tessdataDir)
	}
	return append(args, imagePath, outputBase, "-l", lang)
}

func (n *NativeBackend) Close() error {
	return nil
}

// Version returns the version of the native Tesseract installation.
func (n *NativeBackend) Version() string {
	info, err := DetectNativeTesseract()
	if err != nil {
		return "unknown"
	}
	return info.Version
}

// HasLanguage checks if a language is available in the system tessdata.
func (n *NativeBackend) HasLanguage(lang string) bool {
	if n.tessdataDir == "" {
		return false
	}
	_, err := os.Stat(filepath.Join(n.tessdataDir, lang+".traineddata"))
	return err == nil
}

func defaultLang(lang, fallback string) string {
	if lang != "" {
		return lang
	}
	if fallback != "" {
		return fallback
	}
	return "eng"
}
