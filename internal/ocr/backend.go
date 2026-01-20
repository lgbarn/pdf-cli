package ocr

import (
	"context"
)

// Backend defines the interface for OCR backends.
type Backend interface {
	Name() string
	Available() bool
	ProcessImage(ctx context.Context, imagePath, lang string) (string, error)
	Close() error
}

// BackendType represents the type of OCR backend to use.
type BackendType int

const (
	BackendAuto   BackendType = iota // Auto-select best available backend
	BackendNative                    // System-installed Tesseract
	BackendWASM                      // WASM-based Tesseract (gogosseract)
)

func (b BackendType) String() string {
	switch b {
	case BackendNative:
		return "native"
	case BackendWASM:
		return "wasm"
	default:
		return "auto"
	}
}

// ParseBackendType converts a string to BackendType.
func ParseBackendType(s string) BackendType {
	switch s {
	case "native":
		return BackendNative
	case "wasm":
		return BackendWASM
	default:
		return BackendAuto
	}
}
