package pdf

import (
	"fmt"

	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/types"
)

// AddWatermark adds a text watermark to a PDF
func AddWatermark(input, output, text string, pages []int, password string) error {
	wm, err := pdfcpu.ParseTextWatermarkDetails(text, "scale:1.0, rotation:45, opacity:0.3, color:0.5 0.5 0.5", true, types.POINTS)
	if err != nil {
		return fmt.Errorf("failed to parse watermark: %w", err)
	}
	return api.AddWatermarksFile(input, output, pagesToStrings(pages), wm, NewConfig(password))
}

// AddImageWatermark adds an image watermark to a PDF
func AddImageWatermark(input, output, imagePath string, pages []int, password string) error {
	wm, err := pdfcpu.ParseImageWatermarkDetails(imagePath, "scale:0.5, opacity:0.3", true, types.POINTS)
	if err != nil {
		return fmt.Errorf("failed to parse image watermark: %w", err)
	}
	return api.AddWatermarksFile(input, output, pagesToStrings(pages), wm, NewConfig(password))
}
