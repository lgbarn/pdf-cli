package ocr

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/danlock/gogosseract"
	"github.com/lgbarn/pdf-cli/internal/pdf"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/schollz/progressbar/v3"
)

const (
	// TessdataURL is the base URL for downloading tessdata files
	TessdataURL = "https://github.com/tesseract-ocr/tessdata_fast/raw/main"
)

// progressBarTheme is the default theme for progress bars
var progressBarTheme = progressbar.Theme{
	Saucer:        "=",
	SaucerHead:    ">",
	SaucerPadding: " ",
	BarStart:      "[",
	BarEnd:        "]",
}

// newProgressBar creates a progress bar for count-based operations
func newProgressBar(description string, total, threshold int) *progressbar.ProgressBar {
	if total <= threshold {
		return nil
	}
	return progressbar.NewOptions(total,
		progressbar.OptionSetDescription(description),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowCount(),
		progressbar.OptionSetTheme(progressBarTheme),
	)
}

// newBytesProgressBar creates a progress bar for byte-based downloads
func newBytesProgressBar(description string, total int64) *progressbar.ProgressBar {
	return progressbar.NewOptions64(total,
		progressbar.OptionSetDescription(description),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetTheme(progressBarTheme),
	)
}

// finishProgressBar prints a newline after the progress bar if it exists
func finishProgressBar(bar *progressbar.ProgressBar) {
	if bar != nil {
		fmt.Fprintln(os.Stderr)
	}
}

// Engine provides OCR capabilities
type Engine struct {
	dataDir string
	lang    string
}

// NewEngine creates a new OCR engine
func NewEngine(lang string) (*Engine, error) {
	dataDir, err := getDataDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get data directory: %w", err)
	}

	return &Engine{
		dataDir: dataDir,
		lang:    lang,
	}, nil
}

// getDataDir returns the directory for storing tessdata files
func getDataDir() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	dataDir := filepath.Join(configDir, "pdf-cli", "tessdata")
	if err := os.MkdirAll(dataDir, 0750); err != nil {
		return "", err
	}

	return dataDir, nil
}

// EnsureTessdata ensures the tessdata file for the language exists
func (e *Engine) EnsureTessdata() error {
	// Parse language(s) - can be comma or plus separated
	langs := strings.FieldsFunc(e.lang, func(r rune) bool {
		return r == '+' || r == ','
	})

	for _, lang := range langs {
		lang = strings.TrimSpace(lang)
		if lang == "" {
			continue
		}

		dataFile := filepath.Join(e.dataDir, lang+".traineddata")
		if _, err := os.Stat(dataFile); os.IsNotExist(err) {
			if err := e.downloadTessdata(lang); err != nil {
				return fmt.Errorf("failed to download tessdata for %s: %w", lang, err)
			}
		}
	}

	return nil
}

// downloadTessdata downloads the tessdata file for a language
func (e *Engine) downloadTessdata(lang string) error {
	url := fmt.Sprintf("%s/%s.traineddata", TessdataURL, lang)
	dataFile := filepath.Join(e.dataDir, lang+".traineddata")

	fmt.Fprintf(os.Stderr, "Downloading tessdata for '%s'...\n", lang)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download: HTTP %d", resp.StatusCode)
	}

	// Create temporary file first
	tmpFile, err := os.CreateTemp(e.dataDir, "tessdata-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	// Copy with progress bar
	bar := newBytesProgressBar(fmt.Sprintf("Downloading %s.traineddata", lang), resp.ContentLength)
	if _, err := io.Copy(io.MultiWriter(tmpFile, bar), resp.Body); err != nil {
		_ = tmpFile.Close()
		return err
	}
	_ = tmpFile.Close()
	finishProgressBar(bar)

	// Move to final location
	return os.Rename(tmpPath, dataFile)
}

// ExtractTextFromPDF extracts text from a PDF using OCR
func (e *Engine) ExtractTextFromPDF(pdfPath string, pages []int, password string, showProgress bool) (string, error) {
	// Ensure tessdata is available
	if err := e.EnsureTessdata(); err != nil {
		return "", err
	}

	// Create temp directory for extracted images
	tmpDir, err := os.MkdirTemp("", "pdf-ocr-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	// Get page count if no specific pages requested
	if len(pages) == 0 {
		pageCount, err := pdf.PageCount(pdfPath, password)
		if err != nil {
			return "", fmt.Errorf("failed to get page count: %w", err)
		}
		for i := 1; i <= pageCount; i++ {
			pages = append(pages, i)
		}
	}

	// Extract images from PDF
	conf := model.NewDefaultConfiguration()
	if password != "" {
		conf.UserPW = password
		conf.OwnerPW = password
	}

	pageStrs := make([]string, len(pages))
	for i, p := range pages {
		pageStrs[i] = fmt.Sprintf("%d", p)
	}

	if err := api.ExtractImagesFile(pdfPath, tmpDir, pageStrs, conf); err != nil {
		return "", fmt.Errorf("failed to extract images from PDF: %w", err)
	}

	// Find extracted images
	var imageFiles []string
	err = filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(path))
			if ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".tif" || ext == ".tiff" {
				imageFiles = append(imageFiles, path)
			}
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("failed to find extracted images: %w", err)
	}

	if len(imageFiles) == 0 {
		return "", fmt.Errorf("no images found in PDF - OCR requires image-based PDF")
	}

	// Initialize gogosseract
	ctx := context.Background()

	// Read tessdata file
	primaryLang := strings.Split(e.lang, "+")[0]
	primaryLang = strings.Split(primaryLang, ",")[0]
	tessDataPath := filepath.Join(e.dataDir, primaryLang+".traineddata")
	tessDataFile, err := os.Open(tessDataPath) // #nosec G304 -- path is within user config dir
	if err != nil {
		return "", fmt.Errorf("failed to read tessdata: %w", err)
	}
	defer tessDataFile.Close()

	// Create OCR engine
	tess, err := gogosseract.New(ctx, gogosseract.Config{
		Language:     primaryLang,
		TrainingData: tessDataFile,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create OCR engine: %w", err)
	}
	defer tess.Close(ctx)

	// Process each image with OCR
	var bar *progressbar.ProgressBar
	if showProgress {
		bar = newProgressBar("OCR processing", len(imageFiles), 1)
	}
	defer finishProgressBar(bar)

	var result strings.Builder
	for _, imgPath := range imageFiles {
		text := e.processImage(ctx, tess, imgPath)
		if text != "" {
			if result.Len() > 0 {
				result.WriteString("\n")
			}
			result.WriteString(text)
		}
		if bar != nil {
			_ = bar.Add(1)
		}
	}

	return result.String(), nil
}

// processImage performs OCR on a single image file
func (e *Engine) processImage(ctx context.Context, tess *gogosseract.Tesseract, imgPath string) string {
	imgFile, err := os.Open(imgPath) // #nosec G304 -- path is from temp directory we created
	if err != nil {
		return ""
	}
	defer imgFile.Close()

	if err := tess.LoadImage(ctx, imgFile, gogosseract.LoadImageOptions{}); err != nil {
		return ""
	}

	text, err := tess.GetText(ctx, nil)
	_ = tess.ClearImage(ctx)
	if err != nil {
		return ""
	}

	return strings.TrimSpace(text)
}
