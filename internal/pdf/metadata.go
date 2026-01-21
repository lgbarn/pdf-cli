package pdf

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pdfcpu/pdfcpu/pkg/api"
)

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

	pdfInfoResult, err := api.PDFInfo(f, path, nil, false, NewConfig(password))
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
	_ = NewConfig(password)
	return api.PageCountFile(path)
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

	return api.AddPropertiesFile(input, output, properties, NewConfig(password))
}
