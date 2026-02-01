package ocr

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/lgbarn/pdf-cli/internal/cleanup"
	"github.com/lgbarn/pdf-cli/internal/fileio"
	"github.com/lgbarn/pdf-cli/internal/pdf"
	"github.com/lgbarn/pdf-cli/internal/progress"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/schollz/progressbar/v3"
)

const (
	// TessdataURL is the base URL for downloading tessdata files.
	TessdataURL = "https://github.com/tesseract-ocr/tessdata_fast/raw/main"

	// DefaultParallelThreshold is the minimum number of images to trigger parallel processing.
	DefaultParallelThreshold = 5

	// DefaultMaxWorkers is the maximum number of concurrent workers for parallel processing.
	DefaultMaxWorkers = 8

	// DefaultDownloadTimeout is the timeout for downloading tessdata files.
	DefaultDownloadTimeout = 5 * time.Minute

	// DefaultDataDirPerm is the default permission for tessdata directory.
	DefaultDataDirPerm = 0750
)

// EngineOptions contains options for creating an OCR engine.
type EngineOptions struct {
	Lang              string
	DataDir           string
	BackendType       BackendType
	ParallelThreshold int // minimum images to trigger parallel processing (0 = use default)
	MaxWorkers        int // maximum concurrent workers (0 = use default)
}

// Engine provides OCR capabilities with configurable backend.
type Engine struct {
	dataDir           string
	lang              string
	backendType       BackendType
	backend           Backend
	parallelThreshold int
	maxWorkers        int
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

	parallelThreshold := opts.ParallelThreshold
	if parallelThreshold <= 0 {
		parallelThreshold = DefaultParallelThreshold
	}

	maxWorkers := opts.MaxWorkers
	if maxWorkers <= 0 {
		maxWorkers = DefaultMaxWorkers
	}

	engine := &Engine{
		dataDir:           dataDir,
		lang:              lang,
		backendType:       opts.BackendType,
		parallelThreshold: parallelThreshold,
		maxWorkers:        maxWorkers,
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
	if err := os.MkdirAll(dataDir, DefaultDataDirPerm); err != nil {
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

func downloadTessdata(ctx context.Context, dataDir, lang string) (err error) {
	url := fmt.Sprintf("%s/%s.traineddata", TessdataURL, lang)
	dataFile := filepath.Join(dataDir, lang+".traineddata")

	// Sanitize the data file path to prevent directory traversal
	dataFile, err = fileio.SanitizePath(dataFile)
	if err != nil {
		return fmt.Errorf("invalid data file path: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Downloading tessdata for '%s'...\n", lang)

	ctx, cancel := context.WithTimeout(ctx, DefaultDownloadTimeout)
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
	unregisterTmp := cleanup.Register(tmpPath)
	defer unregisterTmp()
	defer os.Remove(tmpPath)

	// Create SHA256 hasher to verify download integrity
	hasher := sha256.New()

	bar := progress.NewBytesProgressBar(fmt.Sprintf("Downloading %s.traineddata", lang), resp.ContentLength)
	if _, err := io.Copy(io.MultiWriter(tmpFile, bar, hasher), resp.Body); err != nil {
		_ = tmpFile.Close()
		return err
	}

	// On success path, use defer with named return to catch close errors
	defer func() {
		if cerr := tmpFile.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("failed to close temp file: %w", cerr)
		}
	}()

	progress.FinishProgressBar(bar)

	// Verify checksum if known
	computedHash := hex.EncodeToString(hasher.Sum(nil))
	if expectedHash := GetChecksum(lang); expectedHash != "" {
		if computedHash != expectedHash {
			return fmt.Errorf(
				"checksum verification failed for %s.traineddata\n  Expected: %s\n  Got:      %s\n"+
					"This may indicate a corrupted download or supply chain attack",
				lang, expectedHash, computedHash,
			)
		}
		fmt.Fprintf(os.Stderr, "Checksum verified for %s.traineddata\n", lang)
	} else {
		fmt.Fprintf(os.Stderr,
			"WARNING: No checksum available for language '%s'. Computed SHA256: %s\n",
			lang, computedHash,
		)
	}

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
	unregisterDir := cleanup.Register(tmpDir)
	defer unregisterDir()
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
	err   error
}

func (e *Engine) processImages(ctx context.Context, imageFiles []string, showProgress bool) (string, error) {
	// Use sequential processing for small batches or WASM backend (not thread-safe)
	threshold := e.parallelThreshold
	if threshold <= 0 {
		threshold = DefaultParallelThreshold
	}
	if len(imageFiles) <= threshold || e.backend.Name() == "wasm" {
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
	var errs []error

	for i, imgPath := range imageFiles {
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		text, err := e.backend.ProcessImage(ctx, imgPath, e.lang)
		if err != nil {
			errs = append(errs, fmt.Errorf("image %d: %w", i, err))
		} else {
			texts = append(texts, text)
		}
		if bar != nil {
			_ = bar.Add(1)
		}
	}

	if len(errs) > 0 {
		return "", errors.Join(errs...)
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
	maxW := e.maxWorkers
	if maxW <= 0 {
		maxW = DefaultMaxWorkers
	}
	workers := min(runtime.NumCPU(), maxW)
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
			text, err := e.backend.ProcessImage(ctx, path, e.lang)
			results <- imageResult{index: idx, text: text, err: err}
		}(i, imgPath)
	}

	// Close results channel when all workers complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results in order using a slice
	texts := make([]string, len(imageFiles))
	var errs []error
	for res := range results {
		texts[res.index] = res.text
		if res.err != nil {
			errs = append(errs, fmt.Errorf("image %d: %w", res.index, res.err))
		}
		if bar != nil {
			_ = bar.Add(1)
		}
	}

	if len(errs) > 0 {
		return "", errors.Join(errs...)
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
