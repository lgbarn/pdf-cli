// Package pdf provides operations for PDF file manipulation including merging,
// splitting, encryption, text extraction, watermarking, and validation.
package pdf

import (
	"strconv"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// NewConfig creates a pdfcpu configuration with optional password.
func NewConfig(password string) *model.Configuration {
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
