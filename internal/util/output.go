package util

import (
	"encoding/csv"
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

// PrintJSON outputs data as JSON.
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

// printTableJSON outputs tabular data as a JSON array of objects.
func (f *OutputFormatter) printTableJSON(headers []string, rows [][]string) error {
	result := make([]map[string]string, 0, len(rows))
	for _, row := range rows {
		obj := make(map[string]string, len(headers))
		for i, header := range headers {
			if i < len(row) {
				obj[header] = row[i]
			} else {
				obj[header] = ""
			}
		}
		result = append(result, obj)
	}
	return f.printJSON(result)
}

// printTableCSV outputs tabular data as CSV or TSV.
func (f *OutputFormatter) printTableCSV(headers []string, rows [][]string, delimiter rune) error {
	w := csv.NewWriter(f.Writer)
	w.Comma = delimiter

	if err := w.Write(headers); err != nil {
		return err
	}
	for _, row := range rows {
		if err := w.Write(row); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}

// printTableHuman outputs tabular data in a human-readable format.
func (f *OutputFormatter) printTableHuman(headers []string, rows [][]string) error {
	// Calculate column widths
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	// Print header
	for i, h := range headers {
		fmt.Fprintf(f.Writer, "%-*s", widths[i]+2, h)
	}
	fmt.Fprintln(f.Writer)

	// Print separator
	for _, w := range widths {
		fmt.Fprint(f.Writer, strings.Repeat("-", w+2))
	}
	fmt.Fprintln(f.Writer)

	// Print rows
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) {
				fmt.Fprintf(f.Writer, "%-*s", widths[i]+2, cell)
			}
		}
		fmt.Fprintln(f.Writer)
	}
	return nil
}

// IsStructured returns true if the format is a structured format (JSON, CSV, TSV).
func (f *OutputFormatter) IsStructured() bool {
	return f.Format == FormatJSON || f.Format == FormatCSV || f.Format == FormatTSV
}
