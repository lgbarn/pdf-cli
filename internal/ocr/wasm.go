package ocr

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/danlock/gogosseract"
)

// WASMBackend implements Backend using gogosseract (WASM-based Tesseract).
type WASMBackend struct {
	dataDir string
	lang    string
	tess    *gogosseract.Tesseract
}

// NewWASMBackend creates a new WASM-based Tesseract backend.
func NewWASMBackend(lang, dataDir string) (*WASMBackend, error) {
	if dataDir == "" {
		var err error
		dataDir, err = getDataDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get data directory: %w", err)
		}
	}

	return &WASMBackend{
		dataDir: dataDir,
		lang:    lang,
	}, nil
}

func (w *WASMBackend) Name() string {
	return "wasm"
}

func (w *WASMBackend) Available() bool {
	return true
}

// EnsureTessdata ensures the tessdata file for the language exists.
func (w *WASMBackend) EnsureTessdata(lang string) error {
	if lang == "" {
		lang = w.lang
	}

	for _, l := range parseLanguages(lang) {
		dataFile := filepath.Join(w.dataDir, l+".traineddata")
		if _, err := os.Stat(dataFile); os.IsNotExist(err) {
			if err := downloadTessdata(context.TODO(), w.dataDir, l); err != nil {
				return fmt.Errorf("failed to download tessdata for %s: %w", l, err)
			}
		}
	}

	return nil
}

func (w *WASMBackend) initializeTesseract(ctx context.Context, lang string) error {
	if w.tess != nil {
		return nil
	}

	if lang == "" {
		lang = w.lang
	}

	if err := w.EnsureTessdata(lang); err != nil {
		return err
	}

	primaryLang := primaryLanguage(lang)

	tessDataPath := filepath.Join(w.dataDir, primaryLang+".traineddata")
	tessDataFile, err := os.Open(tessDataPath) // #nosec G304 -- path is within user config dir
	if err != nil {
		return fmt.Errorf("failed to read tessdata: %w", err)
	}
	defer tessDataFile.Close()

	w.tess, err = gogosseract.New(ctx, gogosseract.Config{
		Language:     primaryLang,
		TrainingData: tessDataFile,
	})
	if err != nil {
		return fmt.Errorf("failed to create WASM OCR engine: %w", err)
	}

	return nil
}

func (w *WASMBackend) ProcessImage(ctx context.Context, imagePath, lang string) (string, error) {
	lang = defaultLang(lang, w.lang)

	if err := w.initializeTesseract(ctx, lang); err != nil {
		return "", err
	}

	imgFile, err := os.Open(imagePath) // #nosec G304 -- path from temp directory we created
	if err != nil {
		return "", fmt.Errorf("failed to open image: %w", err)
	}
	defer imgFile.Close()

	if err := w.tess.LoadImage(ctx, imgFile, gogosseract.LoadImageOptions{}); err != nil {
		return "", fmt.Errorf("failed to load image: %w", err)
	}

	text, err := w.tess.GetText(ctx, nil)
	_ = w.tess.ClearImage(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get text: %w", err)
	}

	return strings.TrimSpace(text), nil
}

func (w *WASMBackend) Close() error {
	if w.tess != nil {
		return w.tess.Close(context.Background())
	}
	return nil
}
