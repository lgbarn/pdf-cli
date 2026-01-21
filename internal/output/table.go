package output

import (
	"encoding/csv"
	"fmt"
	"strings"
)

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
	widths := columnWidths(headers, rows)

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

// columnWidths calculates the maximum width for each column.
func columnWidths(headers []string, rows [][]string) []int {
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
	return widths
}
