package output

import (
	"bytes"
	"strings"
	"testing"
)

func TestParseOutputFormat(t *testing.T) {
	tests := []struct {
		input string
		want  OutputFormat
	}{
		{"json", FormatJSON},
		{"JSON", FormatJSON},
		{"Json", FormatJSON},
		{"csv", FormatCSV},
		{"CSV", FormatCSV},
		{"tsv", FormatTSV},
		{"TSV", FormatTSV},
		{"", FormatHuman},
		{"invalid", FormatHuman},
		{"xml", FormatHuman},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := ParseOutputFormat(tt.input); got != tt.want {
				t.Errorf("ParseOutputFormat(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestNewOutputFormatter(t *testing.T) {
	tests := []struct {
		format     string
		wantFormat OutputFormat
	}{
		{"json", FormatJSON},
		{"csv", FormatCSV},
		{"tsv", FormatTSV},
		{"", FormatHuman},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			formatter := NewOutputFormatter(tt.format)
			if formatter == nil {
				t.Fatal("NewOutputFormatter() returned nil")
			}
			if formatter.Format != tt.wantFormat {
				t.Errorf("Format = %v, want %v", formatter.Format, tt.wantFormat)
			}
			if formatter.Writer == nil {
				t.Error("Writer is nil")
			}
		})
	}
}

func TestOutputFormatterIsStructured(t *testing.T) {
	tests := []struct {
		format string
		want   bool
	}{
		{"json", true},
		{"csv", true},
		{"tsv", true},
		{"", false},
		{"invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			formatter := NewOutputFormatter(tt.format)
			if got := formatter.IsStructured(); got != tt.want {
				t.Errorf("IsStructured() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOutputFormatterPrint(t *testing.T) {
	tests := []struct {
		name    string
		data    interface{}
		wantStr string
	}{
		{"struct", struct{ Name string }{Name: "test"}, "Name"},
		{"map", map[string]string{"key": "value"}, "key"},
		{"slice", []int{1, 2, 3}, "1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			formatter := NewOutputFormatter("json")
			formatter.Writer = &buf

			if err := formatter.Print(tt.data); err != nil {
				t.Fatalf("Print() error = %v", err)
			}
			if !strings.Contains(buf.String(), tt.wantStr) {
				t.Errorf("output = %q, want containing %q", buf.String(), tt.wantStr)
			}
		})
	}
}

func TestOutputFormatterPrintUnsupportedFormat(t *testing.T) {
	var buf bytes.Buffer
	formatter := NewOutputFormatter("")
	formatter.Writer = &buf

	if err := formatter.Print(map[string]string{"key": "value"}); err == nil {
		t.Error("Print() with human format should return error")
	}
}

func TestPrintTableFormats(t *testing.T) {
	headers := []string{"Name", "Age"}
	rows := [][]string{{"Alice", "30"}, {"Bob", "25"}}

	tests := []struct {
		format      string
		wantHeader  string
		wantPattern string
	}{
		{"json", "", "["},
		{"csv", "Name,Age", "Alice,30"},
		{"tsv", "Name\tAge", "Alice\t30"},
		{"", "Name", "-"},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			var buf bytes.Buffer
			formatter := NewOutputFormatter(tt.format)
			formatter.Writer = &buf

			if err := formatter.PrintTable(headers, rows); err != nil {
				t.Fatalf("PrintTable() error = %v", err)
			}

			output := buf.String()
			if tt.wantHeader != "" && !strings.Contains(output, tt.wantHeader) {
				t.Errorf("output missing header %q", tt.wantHeader)
			}
			if !strings.Contains(output, tt.wantPattern) {
				t.Errorf("output missing pattern %q", tt.wantPattern)
			}
			if !strings.Contains(output, "Alice") {
				t.Error("output missing data 'Alice'")
			}
		})
	}
}

func TestPrintTableEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		format      string
		headers     []string
		rows        [][]string
		wantContain string
	}{
		{"empty rows", "", []string{"Name", "Age"}, [][]string{}, "Name"},
		{"mismatched row length", "json", []string{"Name", "Age", "City"}, [][]string{{"Alice", "30"}}, "Alice"},
		{"no headers", "json", []string{}, [][]string{{"Alice", "30"}}, "["},
		{"column widths", "", []string{"Short", "VeryLongHeaderName"}, [][]string{{"A", "B"}}, "VeryLongHeaderName"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			formatter := NewOutputFormatter(tt.format)
			formatter.Writer = &buf

			if err := formatter.PrintTable(tt.headers, tt.rows); err != nil {
				t.Fatalf("PrintTable() error = %v", err)
			}
			if !strings.Contains(buf.String(), tt.wantContain) {
				t.Errorf("output missing %q", tt.wantContain)
			}
		})
	}
}

func TestPrintTableSpecialCharsCSV(t *testing.T) {
	var buf bytes.Buffer
	formatter := NewOutputFormatter("csv")
	formatter.Writer = &buf

	headers := []string{"Name", "Description"}
	rows := [][]string{{"Test", "Contains, comma"}, {"Quote", "Has \"quotes\""}}

	if err := formatter.PrintTable(headers, rows); err != nil {
		t.Fatalf("PrintTable() error = %v", err)
	}
	if !strings.Contains(buf.String(), "\"Contains, comma\"") {
		t.Error("CSV should escape commas with quotes")
	}
}

func TestOutputFormatConstants(t *testing.T) {
	tests := []struct {
		got  OutputFormat
		want string
	}{
		{FormatHuman, ""},
		{FormatJSON, "json"},
		{FormatCSV, "csv"},
		{FormatTSV, "tsv"},
	}
	for _, tt := range tests {
		if string(tt.got) != tt.want {
			t.Errorf("Format constant = %q, want %q", tt.got, tt.want)
		}
	}
}
