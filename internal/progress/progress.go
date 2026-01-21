package progress

import (
	"fmt"
	"os"

	"github.com/schollz/progressbar/v3"
)

// ProgressBarTheme is the default theme for progress bars.
var ProgressBarTheme = progressbar.Theme{
	Saucer:        "=",
	SaucerHead:    ">",
	SaucerPadding: " ",
	BarStart:      "[",
	BarEnd:        "]",
}

// NewProgressBar creates a consistent progress bar with the given description and total count.
// Returns nil if total is at or below the threshold for showing progress.
func NewProgressBar(description string, total, threshold int) *progressbar.ProgressBar {
	if total <= threshold {
		return nil
	}
	return progressbar.NewOptions(total,
		progressbar.OptionSetDescription(description),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowCount(),
		progressbar.OptionSetTheme(ProgressBarTheme),
	)
}

// NewBytesProgressBar creates a progress bar for byte-based progress (e.g., downloads).
func NewBytesProgressBar(description string, total int64) *progressbar.ProgressBar {
	return progressbar.NewOptions64(total,
		progressbar.OptionSetDescription(description),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetTheme(ProgressBarTheme),
	)
}

// FinishProgressBar prints a newline after the progress bar if it exists.
func FinishProgressBar(bar *progressbar.ProgressBar) {
	if bar != nil {
		fmt.Fprintln(os.Stderr)
	}
}
