package pdf

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/ledongthuc/pdf"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
	"github.com/schollz/progressbar/v3"
)

// progressBarTheme is the default theme for progress bars
var progressBarTheme = progressbar.Theme{
	Saucer:        "=",
	SaucerHead:    ">",
	SaucerPadding: " ",
	BarStart:      "[",
	BarEnd:        "]",
}

// newProgressBar creates a consistent progress bar with the given description and total count.
// Returns nil if total is below the threshold for showing progress.
func newProgressBar(description string, total int, threshold int) *progressbar.ProgressBar {
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

// finishProgressBar prints a newline after the progress bar if it exists
func finishProgressBar(bar *progressbar.ProgressBar) {
	if bar != nil {
		fmt.Fprintln(os.Stderr)
	}
}

// newConfig creates a pdfcpu configuration with optional password.
func newConfig(password string) *model.Configuration {
	conf := model.NewDefaultConfiguration()
	if password != "" {
		conf.UserPW = password
		conf.OwnerPW = password
	}
	return conf
}

// pagesToStrings converts page numbers to string format for pdfcpu API.
func pagesToStrings(pages []int) []string {
	if len(pages) == 0 {
		return nil
	}
	result := make([]string, len(pages))
	for i, p := range pages {
		result[i] = strconv.Itoa(p)
	}
	return result
}

// Info holds PDF document information
type Info struct {
	FilePath    string
	FileSize    int64
	Pages       int
	Version     string
	Title       string
	Author      string
	Subject     string
	Keywords    string
	Creator     string
	Producer    string
	CreatedDate string
	ModDate     string
	Encrypted   bool
}

// GetInfo returns information about a PDF file
func GetInfo(path, password string) (*Info, error) {
	// Clean path to prevent directory traversal
	cleanPath := filepath.Clean(path)

	fileInfo, err := os.Stat(cleanPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	f, err := os.Open(cleanPath) // #nosec G304 -- path is cleaned
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() { _ = f.Close() }()

	pdfInfoResult, err := api.PDFInfo(f, path, nil, false, newConfig(password))
	if err != nil {
		return nil, fmt.Errorf("failed to read PDF info: %w", err)
	}

	info := &Info{
		FilePath:  path,
		FileSize:  fileInfo.Size(),
		Pages:     pdfInfoResult.PageCount,
		Version:   pdfInfoResult.Version,
		Title:     pdfInfoResult.Title,
		Author:    pdfInfoResult.Author,
		Subject:   pdfInfoResult.Subject,
		Creator:   pdfInfoResult.Creator,
		Producer:  pdfInfoResult.Producer,
		Encrypted: pdfInfoResult.Encrypted,
	}

	if len(pdfInfoResult.Keywords) > 0 {
		info.Keywords = strings.Join(pdfInfoResult.Keywords, ", ")
	}

	return info, nil
}

// PageCount returns the number of pages in a PDF
func PageCount(path, password string) (int, error) {
	// Note: PageCountFile doesn't use config in newer pdfcpu versions
	_ = newConfig(password)
	return api.PageCountFile(path)
}

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
		return api.MergeCreateFile(inputs, output, false, newConfig(password))
	}

	// For larger number of files with progress, merge incrementally
	bar := newProgressBar("Merging PDFs", len(inputs), 0)

	// Create temp file for intermediate results
	tmpFile, err := os.CreateTemp("", "pdf-merge-*.pdf")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
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
		err := api.MergeCreateFile([]string{tmpPath, inputs[i]}, tmpPath+".new", false, newConfig(password))
		if err != nil {
			return fmt.Errorf("failed to merge file %s: %w", inputs[i], err)
		}
		// Replace temp with new merged result
		if err := os.Rename(tmpPath+".new", tmpPath); err != nil {
			return fmt.Errorf("failed to update temp file: %w", err)
		}
		_ = bar.Add(1)
	}

	finishProgressBar(bar)

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
		return api.SplitFile(input, outputDir, pageCount, newConfig(password))
	}

	// Calculate number of output files
	numOutputFiles := (totalPages + pageCount - 1) / pageCount
	bar := newProgressBar("Splitting PDF", numOutputFiles, 0)

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

	finishProgressBar(bar)
	return nil
}

// ExtractPages extracts specific pages from a PDF into a new file
func ExtractPages(input, output string, pages []int, password string) error {
	return api.CollectFile(input, output, pagesToStrings(pages), newConfig(password))
}

// Rotate rotates pages in a PDF
func Rotate(input, output string, angle int, pages []int, password string) error {
	return api.RotateFile(input, output, angle, pagesToStrings(pages), newConfig(password))
}

// Compress optimizes a PDF for file size
func Compress(input, output, password string) error {
	return api.OptimizeFile(input, output, newConfig(password))
}

// Encrypt adds password protection to a PDF
func Encrypt(input, output, userPW, ownerPW string) error {
	conf := model.NewDefaultConfiguration()
	conf.UserPW = userPW
	if ownerPW != "" {
		conf.OwnerPW = ownerPW
	} else {
		conf.OwnerPW = userPW
	}
	return api.EncryptFile(input, output, conf)
}

// Decrypt removes password protection from a PDF
func Decrypt(input, output, password string) error {
	return api.DecryptFile(input, output, newConfig(password))
}

// ExtractText extracts text content from a PDF
func ExtractText(input string, pages []int, password string) (string, error) {
	return ExtractTextWithProgress(input, pages, password, false)
}

// ExtractTextWithProgress extracts text content from a PDF with optional progress bar
func ExtractTextWithProgress(input string, pages []int, password string, showProgress bool) (string, error) {
	// Try using ledongthuc/pdf first for better text extraction
	text, err := extractTextPrimary(input, pages, showProgress)
	if err == nil && strings.TrimSpace(text) != "" {
		return text, nil
	}

	// Fall back to parsing pdfcpu content extraction
	return extractTextFallback(input, pages, password)
}

// extractTextPrimary uses the ledongthuc/pdf library for text extraction
func extractTextPrimary(input string, pages []int, showProgress bool) (string, error) {
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
	if len(pages) > 5 {
		return extractPagesParallel(r, pages, totalPages, showProgress)
	}

	return extractPagesSequential(r, pages, totalPages, showProgress)
}

// extractPagesSequential extracts text from pages sequentially
func extractPagesSequential(r *pdf.Reader, pages []int, totalPages int, showProgress bool) (string, error) {
	var bar *progressbar.ProgressBar
	if showProgress {
		bar = newProgressBar("Extracting text", len(pages), 5)
	}
	defer finishProgressBar(bar)

	var result strings.Builder
	for _, pageNum := range pages {
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
		return ""
	}
	p := r.Page(pageNum)
	if p.V.IsNull() {
		return ""
	}
	text, err := p.GetPlainText(nil)
	if err != nil {
		return ""
	}
	return text
}

// extractPagesParallel extracts text from pages in parallel
func extractPagesParallel(r *pdf.Reader, pages []int, totalPages int, showProgress bool) (string, error) {
	type pageResult struct {
		pageNum int
		text    string
	}

	var bar *progressbar.ProgressBar
	if showProgress {
		bar = newProgressBar("Extracting text", len(pages), 5)
	}
	defer finishProgressBar(bar)

	results := make(chan pageResult, len(pages))

	for _, pageNum := range pages {
		go func(pn int) {
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
func extractTextFallback(input string, pages []int, password string) (string, error) {
	tmpDir, err := os.MkdirTemp("", "pdf-cli-text-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	if err := api.ExtractContentFile(input, tmpDir, pagesToStrings(pages), newConfig(password)); err != nil {
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

// ExtractImages extracts images from a PDF
func ExtractImages(input, outputDir string, pages []int, password string) error {
	return api.ExtractImagesFile(input, outputDir, pagesToStrings(pages), newConfig(password))
}

// Metadata holds PDF metadata fields
type Metadata struct {
	Title       string
	Author      string
	Subject     string
	Keywords    string
	Creator     string
	Producer    string
	CreatedDate string
	ModDate     string
}

// GetMetadata returns the metadata of a PDF
func GetMetadata(input, password string) (*Metadata, error) {
	info, err := GetInfo(input, password)
	if err != nil {
		return nil, err
	}

	return &Metadata{
		Title:       info.Title,
		Author:      info.Author,
		Subject:     info.Subject,
		Keywords:    info.Keywords,
		Creator:     info.Creator,
		Producer:    info.Producer,
		CreatedDate: info.CreatedDate,
		ModDate:     info.ModDate,
	}, nil
}

// SetMetadata sets metadata on a PDF
func SetMetadata(input, output string, meta *Metadata, password string) error {
	properties := make(map[string]string)
	if meta.Title != "" {
		properties["Title"] = meta.Title
	}
	if meta.Author != "" {
		properties["Author"] = meta.Author
	}
	if meta.Subject != "" {
		properties["Subject"] = meta.Subject
	}
	if meta.Keywords != "" {
		properties["Keywords"] = meta.Keywords
	}
	if meta.Creator != "" {
		properties["Creator"] = meta.Creator
	}
	if meta.Producer != "" {
		properties["Producer"] = meta.Producer
	}

	return api.AddPropertiesFile(input, output, properties, newConfig(password))
}

// AddWatermark adds a text watermark to a PDF
func AddWatermark(input, output, text string, pages []int, password string) error {
	wm, err := pdfcpu.ParseTextWatermarkDetails(text, "scale:1.0, rotation:45, opacity:0.3, color:0.5 0.5 0.5", true, types.POINTS)
	if err != nil {
		return fmt.Errorf("failed to parse watermark: %w", err)
	}
	return api.AddWatermarksFile(input, output, pagesToStrings(pages), wm, newConfig(password))
}

// AddImageWatermark adds an image watermark to a PDF
func AddImageWatermark(input, output, imagePath string, pages []int, password string) error {
	wm, err := pdfcpu.ParseImageWatermarkDetails(imagePath, "scale:0.5, opacity:0.3", true, types.POINTS)
	if err != nil {
		return fmt.Errorf("failed to parse image watermark: %w", err)
	}
	return api.AddWatermarksFile(input, output, pagesToStrings(pages), wm, newConfig(password))
}

// Validate validates a PDF file
func Validate(path, password string) error {
	return api.ValidateFile(path, newConfig(password))
}

// ValidateToBuffer validates a PDF from bytes
func ValidateToBuffer(data []byte) error {
	return api.Validate(bytes.NewReader(data), newConfig(""))
}
