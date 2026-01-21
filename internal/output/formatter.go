package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// OutputFormat represents the output format type.
type OutputFormat string

const (
	FormatHuman OutputFormat = ""
	FormatJSON  OutputFormat = "json"
	FormatCSV   OutputFormat = "csv"
	FormatTSV   OutputFormat = "tsv"
)

// ParseOutputFormat parses a string into an OutputFormat.
func ParseOutputFormat(s string) OutputFormat {
	switch strings.ToLower(s) {
	case "json":
		return FormatJSON
	case "csv":
		return FormatCSV
	case "tsv":
		return FormatTSV
	default:
		return FormatHuman
	}
}

// OutputFormatter handles formatted output in various formats.
type OutputFormatter struct {
	Format OutputFormat
	Writer io.Writer
}

// NewOutputFormatter creates a new OutputFormatter with the given format.
func NewOutputFormatter(format string) *OutputFormatter {
	return &OutputFormatter{
		Format: ParseOutputFormat(format),
		Writer: os.Stdout,
	}
}

// Print outputs data in the configured format.
// For JSON, data is marshaled directly.
// For other formats, data should be a struct or map.
func (f *OutputFormatter) Print(data interface{}) error {
	switch f.Format {
	case FormatJSON:
		return f.printJSON(data)
	default:
		return fmt.Errorf("unsupported format for Print: %s", f.Format)
	}
}

// printJSON outputs data as JSON.
func (f *OutputFormatter) printJSON(data interface{}) error {
	encoder := json.NewEncoder(f.Writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

// PrintTable outputs tabular data in the configured format.
func (f *OutputFormatter) PrintTable(headers []string, rows [][]string) error {
	switch f.Format {
	case FormatJSON:
		return f.printTableJSON(headers, rows)
	case FormatCSV:
		return f.printTableCSV(headers, rows, ',')
	case FormatTSV:
		return f.printTableCSV(headers, rows, '\t')
	default:
		return f.printTableHuman(headers, rows)
	}
}

// IsStructured returns true if the format is a structured format (JSON, CSV, TSV).
func (f *OutputFormatter) IsStructured() bool {
	return f.Format == FormatJSON || f.Format == FormatCSV || f.Format == FormatTSV
}
