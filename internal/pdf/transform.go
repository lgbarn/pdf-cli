package pdf

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lgbarn/pdf-cli/internal/cleanup"
	"github.com/lgbarn/pdf-cli/internal/progress"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

// Merge combines multiple PDF files into one
func Merge(inputs []string, output, password string) error {
	return MergeWithProgress(inputs, output, password, false)
}

// MergeWithProgress combines multiple PDF files into one with optional progress bar
func MergeWithProgress(inputs []string, output, password string, showProgress bool) error {
	if len(inputs) == 0 {
		return fmt.Errorf("no input files provided")
	}

	// For small number of files or no progress, use the standard merge
	if !showProgress || len(inputs) <= 3 {
		return api.MergeCreateFile(inputs, output, false, NewConfig(password))
	}

	// For larger number of files with progress, merge incrementally
	bar := progress.NewProgressBar("Merging PDFs", len(inputs), 0)

	// Create temp file for intermediate results
	tmpFile, err := os.CreateTemp("", "pdf-merge-*.pdf")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	_ = tmpFile.Close()
	unregisterTmp := cleanup.Register(tmpPath)
	defer unregisterTmp()
	defer os.Remove(tmpPath)

	// Copy first file to temp
	firstContent, err := os.ReadFile(inputs[0])
	if err != nil {
		return fmt.Errorf("failed to read first file: %w", err)
	}
	if err := os.WriteFile(tmpPath, firstContent, 0600); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	_ = bar.Add(1)

	// Merge remaining files one at a time
	for i := 1; i < len(inputs); i++ {
		err := api.MergeCreateFile([]string{tmpPath, inputs[i]}, tmpPath+".new", false, NewConfig(password))
		if err != nil {
			return fmt.Errorf("failed to merge file %s: %w", inputs[i], err)
		}
		// Replace temp with new merged result
		if err := os.Rename(tmpPath+".new", tmpPath); err != nil {
			return fmt.Errorf("failed to update temp file: %w", err)
		}
		_ = bar.Add(1)
	}

	progress.FinishProgressBar(bar)

	// Move final result to output
	return os.Rename(tmpPath, output)
}

// Split splits a PDF into individual pages
func Split(input, outputDir, password string) error {
	return SplitWithProgress(input, outputDir, 1, password, false)
}

// SplitByPageCount splits a PDF into chunks of n pages
func SplitByPageCount(input, outputDir string, pageCount int, password string) error {
	return SplitWithProgress(input, outputDir, pageCount, password, false)
}

// SplitWithProgress splits a PDF with optional progress bar
func SplitWithProgress(input, outputDir string, pageCount int, password string, showProgress bool) error {
	totalPages, err := PageCount(input, password)
	if err != nil {
		return fmt.Errorf("failed to get page count: %w", err)
	}

	// For small PDFs or no progress, use the standard split
	if !showProgress || totalPages <= 5 {
		return api.SplitFile(input, outputDir, pageCount, NewConfig(password))
	}

	// Calculate number of output files
	numOutputFiles := (totalPages + pageCount - 1) / pageCount
	bar := progress.NewProgressBar("Splitting PDF", numOutputFiles, 0)

	// Get base name for output files
	baseName := filepath.Base(input)
	ext := filepath.Ext(baseName)
	baseName = baseName[:len(baseName)-len(ext)]

	// Split page by page (or chunk by chunk)
	for i := 0; i < numOutputFiles; i++ {
		startPage := i*pageCount + 1
		endPage := startPage + pageCount - 1
		if endPage > totalPages {
			endPage = totalPages
		}

		// Build page list for this chunk
		var pages []int
		for p := startPage; p <= endPage; p++ {
			pages = append(pages, p)
		}

		// Create output filename
		var outputFile string
		if pageCount == 1 {
			outputFile = filepath.Join(outputDir, fmt.Sprintf("%s_%d%s", baseName, startPage, ext))
		} else {
			outputFile = filepath.Join(outputDir, fmt.Sprintf("%s_%d-%d%s", baseName, startPage, endPage, ext))
		}

		if err := ExtractPages(input, outputFile, pages, password); err != nil {
			return fmt.Errorf("failed to extract pages %d-%d: %w", startPage, endPage, err)
		}

		_ = bar.Add(1)
	}

	progress.FinishProgressBar(bar)
	return nil
}

// ExtractPages extracts specific pages from a PDF into a new file
func ExtractPages(input, output string, pages []int, password string) error {
	return api.CollectFile(input, output, pagesToStrings(pages), NewConfig(password))
}

// Rotate rotates pages in a PDF
func Rotate(input, output string, angle int, pages []int, password string) error {
	return api.RotateFile(input, output, angle, pagesToStrings(pages), NewConfig(password))
}

// Compress optimizes a PDF for file size
func Compress(input, output, password string) error {
	return api.OptimizeFile(input, output, NewConfig(password))
}

// ExtractImages extracts images from a PDF
func ExtractImages(input, outputDir string, pages []int, password string) error {
	return api.ExtractImagesFile(input, outputDir, pagesToStrings(pages), NewConfig(password))
}

// CreatePDFFromImages creates a PDF from multiple image files.
// Each image becomes one page in the output PDF.
func CreatePDFFromImages(images []string, output, pageSize string) error {
	imp := pdfcpu.DefaultImportConfig()
	imp.Pos = types.Full

	if pageSize != "" {
		imp.UserDim = true
		switch strings.ToUpper(pageSize) {
		case "A4":
			imp.PageSize = "A4"
		case "LETTER":
			imp.PageSize = "Letter"
		default:
			imp.PageSize = pageSize
		}
	}

	return api.ImportImagesFile(images, output, imp, nil)
}
