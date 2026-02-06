package pdf

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ledongthuc/pdf"
	"github.com/lgbarn/pdf-cli/internal/cleanup"
	"github.com/lgbarn/pdf-cli/internal/config"
	"github.com/lgbarn/pdf-cli/internal/logging"
	"github.com/lgbarn/pdf-cli/internal/progress"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/schollz/progressbar/v3"
)

const (
	// ParallelThreshold is the minimum number of pages to trigger parallel processing.
	ParallelThreshold = 5

	// ProgressUpdateInterval is the interval for updating the progress bar (number of pages).
	ProgressUpdateInterval = 5
)

// ExtractText extracts text content from a PDF
func ExtractText(ctx context.Context, input string, pages []int, password string) (string, error) {
	return ExtractTextWithProgress(ctx, input, pages, password, false)
}

// ExtractTextWithProgress extracts text content from a PDF with optional progress bar
func ExtractTextWithProgress(ctx context.Context, input string, pages []int, password string, showProgress bool) (string, error) {
	// Try using ledongthuc/pdf first for better text extraction
	text, err := extractTextPrimary(ctx, input, pages, showProgress)
	if err == nil && strings.TrimSpace(text) != "" {
		return text, nil
	}

	// Fall back to parsing pdfcpu content extraction
	return extractTextFallback(ctx, input, pages, password)
}

// extractTextPrimary uses the ledongthuc/pdf library for text extraction
func extractTextPrimary(ctx context.Context, input string, pages []int, showProgress bool) (string, error) {
	f, r, err := pdf.Open(input)
	if err != nil {
		return "", err
	}
	defer f.Close()

	totalPages := r.NumPage()

	// If no specific pages requested, extract from all pages
	if len(pages) == 0 {
		pages = make([]int, totalPages)
		for i := range pages {
			pages[i] = i + 1
		}
	} else {
		// Sort specific pages for consistent ordering
		sortedPages := make([]int, len(pages))
		copy(sortedPages, pages)
		sort.Ints(sortedPages)
		pages = sortedPages
	}

	// Use parallel extraction for larger page counts
	cfg := config.Get()
	threshold := cfg.Performance.TextParallelThreshold
	if threshold <= 0 {
		threshold = ParallelThreshold
	}
	if len(pages) > threshold {
		return extractPagesParallel(ctx, r, pages, totalPages, showProgress)
	}

	return extractPagesSequential(ctx, r, pages, totalPages, showProgress)
}

// extractPagesSequential extracts text from pages sequentially
func extractPagesSequential(ctx context.Context, r *pdf.Reader, pages []int, totalPages int, showProgress bool) (string, error) {
	var bar *progressbar.ProgressBar
	if showProgress {
		bar = progress.NewProgressBar("Extracting text", len(pages), ProgressUpdateInterval)
	}
	defer progress.FinishProgressBar(bar)

	var result strings.Builder
	for _, pageNum := range pages {
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		text := extractPageText(r, pageNum, totalPages)
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

// extractPageText extracts text from a single page, returning empty string on any error
func extractPageText(r *pdf.Reader, pageNum, totalPages int) string {
	if pageNum < 1 || pageNum > totalPages {
		logging.Debug("page number out of range", "page", pageNum, "total", totalPages)
		return ""
	}
	p := r.Page(pageNum)
	if p.V.IsNull() {
		logging.Debug("page object is null", "page", pageNum)
		return ""
	}
	text, err := p.GetPlainText(nil)
	if err != nil {
		logging.Debug("failed to extract text from page", "page", pageNum, "error", err)
		return ""
	}
	return text
}

// extractPagesParallel extracts text from pages in parallel
func extractPagesParallel(ctx context.Context, r *pdf.Reader, pages []int, totalPages int, showProgress bool) (string, error) {
	type pageResult struct {
		pageNum int
		text    string
	}

	var bar *progressbar.ProgressBar
	if showProgress {
		bar = progress.NewProgressBar("Extracting text", len(pages), ProgressUpdateInterval)
	}
	defer progress.FinishProgressBar(bar)

	results := make(chan pageResult, len(pages))

	for _, pageNum := range pages {
		if ctx.Err() != nil {
			// Context canceled, don't launch more work
			break
		}
		go func(pn int) {
			if ctx.Err() != nil {
				results <- pageResult{pageNum: pn, text: ""}
				return
			}
			results <- pageResult{pageNum: pn, text: extractPageText(r, pn, totalPages)}
		}(pageNum)
	}

	// Collect results into a map
	pageTexts := make(map[int]string)
	for range pages {
		res := <-results
		pageTexts[res.pageNum] = res.text
		if bar != nil {
			_ = bar.Add(1)
		}
	}

	// Build result in page order
	var result strings.Builder
	for _, pageNum := range pages {
		text := pageTexts[pageNum]
		if text != "" {
			if result.Len() > 0 {
				result.WriteString("\n")
			}
			result.WriteString(text)
		}
	}

	return result.String(), nil
}

// extractTextFallback parses text from pdfcpu's raw content extraction
func extractTextFallback(ctx context.Context, input string, pages []int, password string) (string, error) {
	if ctx.Err() != nil {
		return "", ctx.Err()
	}

	tmpDir, err := os.MkdirTemp("", "pdf-cli-text-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	unregisterDir := cleanup.Register(tmpDir)
	defer unregisterDir()
	defer os.RemoveAll(tmpDir)

	if err := api.ExtractContentFile(input, tmpDir, pagesToStrings(pages), NewConfig(password)); err != nil {
		return "", fmt.Errorf("failed to extract content: %w", err)
	}

	files, err := os.ReadDir(tmpDir)
	if err != nil {
		return "", err
	}

	var result strings.Builder
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".txt") {
			continue
		}
		// Use filepath.Join to safely construct path within tmpDir
		filePath := filepath.Join(tmpDir, filepath.Base(file.Name()))
		data, err := os.ReadFile(filePath) // #nosec G304 -- path is within controlled tmpDir
		if err != nil {
			continue
		}
		text := parseTextFromPDFContent(string(data))
		if text != "" {
			if result.Len() > 0 {
				result.WriteString("\n")
			}
			result.WriteString(text)
		}
	}

	return result.String(), nil
}

// parseTextFromPDFContent extracts readable text from raw PDF content stream
func parseTextFromPDFContent(content string) string {
	var result strings.Builder
	i := 0

	for i < len(content) {
		// Look for string literals in parentheses
		if content[i] == '(' {
			// Find matching close parenthesis, handling escapes and nesting
			text, endIdx := extractParenString(content, i)
			if text != "" {
				if result.Len() > 0 && !strings.HasSuffix(result.String(), " ") && !strings.HasSuffix(result.String(), "\n") {
					result.WriteString(" ")
				}
				result.WriteString(text)
			}
			if endIdx > i {
				i = endIdx
				continue
			}
		}

		// Handle newlines in content stream (usually after ET)
		if i+1 < len(content) && content[i:i+2] == "ET" {
			if result.Len() > 0 && !strings.HasSuffix(result.String(), "\n") {
				result.WriteString("\n")
			}
		}

		i++
	}

	return strings.TrimSpace(result.String())
}

// extractParenString extracts a string from parentheses, handling escapes
func extractParenString(content string, start int) (string, int) {
	if start >= len(content) || content[start] != '(' {
		return "", start
	}

	var result strings.Builder
	depth := 1
	i := start + 1

	for i < len(content) && depth > 0 {
		ch := content[i]

		if ch == '\\' && i+1 < len(content) {
			// Handle escape sequences
			next := content[i+1]
			switch next {
			case 'n':
				result.WriteByte('\n')
			case 'r':
				result.WriteByte('\r')
			case 't':
				result.WriteByte('\t')
			case '(', ')', '\\':
				result.WriteByte(next)
			default:
				// Octal escape or unknown - skip
			}
			i += 2
			continue
		}

		switch ch {
		case '(':
			depth++
			result.WriteByte(ch)
		case ')':
			depth--
			if depth > 0 {
				result.WriteByte(ch)
			}
		default:
			result.WriteByte(ch)
		}

		i++
	}

	return result.String(), i
}
