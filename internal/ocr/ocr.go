package ocr

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/lgbarn/pdf-cli/internal/fileio"
	"github.com/lgbarn/pdf-cli/internal/pdf"
	"github.com/lgbarn/pdf-cli/internal/progress"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/schollz/progressbar/v3"
)

const (
	// TessdataURL is the base URL for downloading tessdata files.
	TessdataURL = "https://github.com/tesseract-ocr/tessdata_fast/raw/main"
)

// EngineOptions contains options for creating an OCR engine.
type EngineOptions struct {
	Lang        string
	DataDir     string
	BackendType BackendType
}

// Engine provides OCR capabilities with configurable backend.
type Engine struct {
	dataDir     string
	lang        string
	backendType BackendType
	backend     Backend
}

// NewEngine creates a new OCR engine with auto backend selection.
func NewEngine(lang string) (*Engine, error) {
	return NewEngineWithOptions(EngineOptions{
		Lang:        lang,
		BackendType: BackendAuto,
	})
}

// NewEngineWithOptions creates a new OCR engine with specified options.
func NewEngineWithOptions(opts EngineOptions) (*Engine, error) {
	dataDir := opts.DataDir
	if dataDir == "" {
		var err error
		dataDir, err = getDataDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get data directory: %w", err)
		}
	}

	lang := opts.Lang
	if lang == "" {
		lang = "eng"
	}

	engine := &Engine{
		dataDir:     dataDir,
		lang:        lang,
		backendType: opts.BackendType,
	}

	backend, err := engine.selectBackend()
	if err != nil {
		return nil, err
	}
	engine.backend = backend

	return engine, nil
}

func (e *Engine) selectBackend() (Backend, error) {
	switch e.backendType {
	case BackendNative:
		backend, err := NewNativeBackend(e.lang, e.dataDir)
		if err != nil {
			return nil, fmt.Errorf("native backend requested but not available: %w", err)
		}
		return backend, nil

	case BackendWASM:
		return NewWASMBackend(e.lang, e.dataDir)

	default: // BackendAuto - try native first, fall back to WASM
		if backend, err := NewNativeBackend(e.lang, ""); err == nil {
			return backend, nil
		}
		return NewWASMBackend(e.lang, e.dataDir)
	}
}

// BackendName returns the name of the currently active backend.
func (e *Engine) BackendName() string {
	if e.backend != nil {
		return e.backend.Name()
	}
	return "none"
}

// Close releases resources held by the engine.
func (e *Engine) Close() error {
	if e.backend != nil {
		return e.backend.Close()
	}
	return nil
}

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

// EnsureTessdata ensures the tessdata file for the language exists.
func (e *Engine) EnsureTessdata() error {
	for _, lang := range parseLanguages(e.lang) {
		dataFile := filepath.Join(e.dataDir, lang+".traineddata")
		if _, err := os.Stat(dataFile); os.IsNotExist(err) {
			if err := downloadTessdata(context.TODO(), e.dataDir, lang); err != nil {
				return fmt.Errorf("failed to download tessdata for %s: %w", lang, err)
			}
		}
	}
	return nil
}

// parseLanguages splits a language string (e.g., "eng+fra" or "eng,fra") into individual languages.
func parseLanguages(lang string) []string {
	parts := strings.FieldsFunc(lang, func(r rune) bool {
		return r == '+' || r == ','
	})

	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// primaryLanguage returns the first language from a language string.
func primaryLanguage(lang string) string {
	langs := parseLanguages(lang)
	if len(langs) > 0 {
		return langs[0]
	}
	return lang
}

func downloadTessdata(ctx context.Context, dataDir, lang string) error {
	url := fmt.Sprintf("%s/%s.traineddata", TessdataURL, lang)
	dataFile := filepath.Join(dataDir, lang+".traineddata")

	fmt.Fprintf(os.Stderr, "Downloading tessdata for '%s'...\n", lang)

	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
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

	tmpFile, err := os.CreateTemp(dataDir, "tessdata-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	bar := progress.NewBytesProgressBar(fmt.Sprintf("Downloading %s.traineddata", lang), resp.ContentLength)
	if _, err := io.Copy(io.MultiWriter(tmpFile, bar), resp.Body); err != nil {
		_ = tmpFile.Close()
		return err
	}
	_ = tmpFile.Close()
	progress.FinishProgressBar(bar)

	return os.Rename(tmpPath, dataFile)
}

// ExtractTextFromPDF extracts text from a PDF using OCR.
func (e *Engine) ExtractTextFromPDF(ctx context.Context, pdfPath string, pages []int, password string, showProgress bool) (string, error) {
	if e.backend.Name() == "wasm" {
		if err := e.EnsureTessdata(); err != nil {
			return "", err
		}
	}

	tmpDir, err := os.MkdirTemp("", "pdf-ocr-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	pages, err = e.resolvePages(pdfPath, pages, password)
	if err != nil {
		return "", err
	}

	if err := e.extractImagesToDir(pdfPath, tmpDir, pages, password); err != nil {
		return "", err
	}

	imageFiles, err := findImageFiles(tmpDir)
	if err != nil {
		return "", err
	}

	if len(imageFiles) == 0 {
		return "", fmt.Errorf("no images found in PDF - OCR requires image-based PDF")
	}

	return e.processImages(ctx, imageFiles, showProgress)
}

func (e *Engine) resolvePages(pdfPath string, pages []int, password string) ([]int, error) {
	if len(pages) > 0 {
		return pages, nil
	}

	pageCount, err := pdf.PageCount(pdfPath, password)
	if err != nil {
		return nil, fmt.Errorf("failed to get page count: %w", err)
	}

	result := make([]int, pageCount)
	for i := range result {
		result[i] = i + 1
	}
	return result, nil
}

func (e *Engine) extractImagesToDir(pdfPath, tmpDir string, pages []int, password string) error {
	pageStrs := make([]string, len(pages))
	for i, p := range pages {
		pageStrs[i] = fmt.Sprintf("%d", p)
	}

	if err := api.ExtractImagesFile(pdfPath, tmpDir, pageStrs, pdf.NewConfig(password)); err != nil {
		return fmt.Errorf("failed to extract images from PDF: %w", err)
	}
	return nil
}

func findImageFiles(dir string) ([]string, error) {
	var imageFiles []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && fileio.IsImageFile(path) {
			imageFiles = append(imageFiles, path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to find extracted images: %w", err)
	}
	return imageFiles, nil
}

// imageResult holds the result of processing a single image.
type imageResult struct {
	index int
	text  string
}

// parallelThreshold is the minimum number of images to trigger parallel processing.
const parallelThreshold = 5

func (e *Engine) processImages(ctx context.Context, imageFiles []string, showProgress bool) (string, error) {
	// Use sequential processing for small batches or WASM backend (not thread-safe)
	if len(imageFiles) <= parallelThreshold || e.backend.Name() == "wasm" {
		return e.processImagesSequential(ctx, imageFiles, showProgress)
	}
	return e.processImagesParallel(ctx, imageFiles, showProgress)
}

func (e *Engine) processImagesSequential(ctx context.Context, imageFiles []string, showProgress bool) (string, error) {
	var bar *progressbar.ProgressBar
	if showProgress {
		bar = progress.NewProgressBar("OCR processing", len(imageFiles), 1)
	}
	defer progress.FinishProgressBar(bar)

	texts := make([]string, 0, len(imageFiles))

	for _, imgPath := range imageFiles {
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		text, err := e.backend.ProcessImage(ctx, imgPath, e.lang)
		if err == nil {
			texts = append(texts, text)
		}
		if bar != nil {
			_ = bar.Add(1)
		}
	}

	return joinNonEmpty(texts, "\n"), nil
}

func (e *Engine) processImagesParallel(ctx context.Context, imageFiles []string, showProgress bool) (string, error) {
	var bar *progressbar.ProgressBar
	if showProgress {
		bar = progress.NewProgressBar("OCR processing", len(imageFiles), 1)
	}
	defer progress.FinishProgressBar(bar)

	results := make(chan imageResult, len(imageFiles))

	// Limit concurrent workers to avoid resource exhaustion
	workers := min(runtime.NumCPU(), 8)
	sem := make(chan struct{}, workers)

	var wg sync.WaitGroup
	for i, imgPath := range imageFiles {
		if ctx.Err() != nil {
			// Context canceled, don't launch more work
			break
		}

		wg.Add(1)
		sem <- struct{}{} // Acquire semaphore
		go func(idx int, path string) {
			defer wg.Done()
			defer func() { <-sem }() // Release semaphore
			text, _ := e.backend.ProcessImage(ctx, path, e.lang)
			results <- imageResult{index: idx, text: text}
		}(i, imgPath)
	}

	// Close results channel when all workers complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results in order using a slice
	texts := make([]string, len(imageFiles))
	for res := range results {
		texts[res.index] = res.text
		if bar != nil {
			_ = bar.Add(1)
		}
	}

	return joinNonEmpty(texts, "\n"), nil
}

// joinNonEmpty joins non-empty strings with the given separator.
func joinNonEmpty(strs []string, sep string) string {
	var result strings.Builder
	for _, s := range strs {
		if s == "" {
			continue
		}
		if result.Len() > 0 {
			result.WriteString(sep)
		}
		result.WriteString(s)
	}
	return result.String()
}
