package pdf

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/ledongthuc/pdf"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

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
	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

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
	return api.MergeCreateFile(inputs, output, false, newConfig(password))
}

// Split splits a PDF into individual pages
func Split(input, outputDir, password string) error {
	return api.SplitFile(input, outputDir, 1, newConfig(password))
}

// SplitByPageCount splits a PDF into chunks of n pages
func SplitByPageCount(input, outputDir string, pageCount int, password string) error {
	return api.SplitFile(input, outputDir, pageCount, newConfig(password))
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
	// Try using ledongthuc/pdf first for better text extraction
	text, err := extractTextPrimary(input, pages)
	if err == nil && strings.TrimSpace(text) != "" {
		return text, nil
	}

	// Fall back to parsing pdfcpu content extraction
	return extractTextFallback(input, pages, password)
}

// extractTextPrimary uses the ledongthuc/pdf library for text extraction
func extractTextPrimary(input string, pages []int) (string, error) {
	f, r, err := pdf.Open(input)
	if err != nil {
		return "", err
	}
	defer f.Close()

	// If no specific pages requested, extract from all pages
	if len(pages) == 0 {
		var buf bytes.Buffer
		b, err := r.GetPlainText()
		if err != nil {
			return "", err
		}
		buf.ReadFrom(b)
		return buf.String(), nil
	}

	// Extract from specific pages
	sortedPages := make([]int, len(pages))
	copy(sortedPages, pages)
	sort.Ints(sortedPages)

	var result strings.Builder
	totalPages := r.NumPage()

	for _, pageNum := range sortedPages {
		if pageNum < 1 || pageNum > totalPages {
			continue
		}
		p := r.Page(pageNum)
		if p.V.IsNull() {
			continue
		}
		text, err := p.GetPlainText(nil)
		if err != nil {
			continue
		}
		if result.Len() > 0 {
			result.WriteString("\n")
		}
		result.WriteString(text)
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
		data, err := os.ReadFile(tmpDir + "/" + file.Name())
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
	inText := false
	i := 0

	for i < len(content) {
		// Look for text show operators: Tj, TJ, ', "
		if i+1 < len(content) {
			op := content[i : i+2]
			if op == "Tj" || op == "TJ" {
				inText = false
			}
		}

		// Look for string literals in parentheses
		if content[i] == '(' {
			// Find matching close parenthesis, handling escapes and nesting
			text, endIdx := extractParenString(content, i)
			if text != "" {
				if result.Len() > 0 && !strings.HasSuffix(result.String(), " ") && !strings.HasSuffix(result.String(), "\n") {
					result.WriteString(" ")
				}
				result.WriteString(text)
				inText = true
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

	_ = inText // suppress unused warning
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
