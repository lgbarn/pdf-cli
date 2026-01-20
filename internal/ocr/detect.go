package ocr

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
)

// ErrNativeNotFound is returned when native Tesseract is not installed.
var ErrNativeNotFound = errors.New("native Tesseract not found")

// NativeInfo contains information about the native Tesseract installation.
type NativeInfo struct {
	Path     string
	Version  string
	Tessdata string
}

// DetectNativeTesseract checks if native Tesseract is installed and returns its info.
func DetectNativeTesseract() (*NativeInfo, error) {
	path, err := exec.LookPath("tesseract")
	if err != nil {
		return nil, ErrNativeNotFound
	}

	version, err := getTesseractVersion(path)
	if err != nil {
		return nil, err
	}

	tessdata, _ := findTessdataDir(path) // Not fatal if tessdata not found

	return &NativeInfo{
		Path:     path,
		Version:  version,
		Tessdata: tessdata,
	}, nil
}

func getTesseractVersion(tesseractPath string) (string, error) {
	cmd := exec.Command(tesseractPath, "--version") // #nosec G204 -- tesseractPath comes from exec.LookPath
	output, err := cmd.CombinedOutput()
	// tesseract --version may exit non-zero on some systems but still produce output
	if err != nil && len(output) == 0 {
		return "", err
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) == 0 {
		return "unknown", nil
	}

	re := regexp.MustCompile(`tesseract\s+(\d+\.\d+(?:\.\d+)?)`)
	if matches := re.FindStringSubmatch(lines[0]); len(matches) >= 2 {
		return matches[1], nil
	}

	return "unknown", nil
}

func findTessdataDir(tesseractPath string) (string, error) {
	// Check TESSDATA_PREFIX environment variable first
	if envPath := os.Getenv("TESSDATA_PREFIX"); isValidTessdataDir(envPath) {
		return envPath, nil
	}

	// Try to get from tesseract --print-parameters
	cmd := exec.Command(tesseractPath, "--print-parameters") // #nosec G204 -- tesseractPath comes from exec.LookPath
	output, _ := cmd.CombinedOutput()
	if dataPath := extractTessdataFromParams(string(output)); isValidTessdataDir(dataPath) {
		return dataPath, nil
	}

	// Try common locations based on OS
	for _, path := range getCommonTessdataPaths() {
		if isValidTessdataDir(path) {
			return path, nil
		}
	}

	return "", errors.New("tessdata directory not found")
}

func extractTessdataFromParams(output string) string {
	re := regexp.MustCompile(`tessdata_dir\s+(\S+)`)
	if matches := re.FindStringSubmatch(output); len(matches) >= 2 {
		return matches[1]
	}
	return ""
}

func getCommonTessdataPaths() []string {
	switch runtime.GOOS {
	case "darwin":
		return []string{
			"/opt/homebrew/share/tessdata",
			"/usr/local/share/tessdata",
			"/opt/local/share/tessdata",
		}
	case "linux":
		return []string{
			"/usr/share/tesseract-ocr/5/tessdata",
			"/usr/share/tesseract-ocr/4.00/tessdata",
			"/usr/share/tessdata",
			"/usr/local/share/tessdata",
		}
	case "windows":
		return []string{
			filepath.Join(os.Getenv("PROGRAMFILES"), "Tesseract-OCR", "tessdata"),
			filepath.Join(os.Getenv("PROGRAMFILES(X86)"), "Tesseract-OCR", "tessdata"),
		}
	default:
		return nil
	}
}

func isValidTessdataDir(path string) bool {
	if path == "" {
		return false
	}

	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return false
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return false
	}

	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".traineddata") {
			return true
		}
	}
	return false
}
