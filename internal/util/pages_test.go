package util

import (
	"reflect"
	"strings"
	"testing"
)

func TestParsePageRanges(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []PageRange
		wantErr bool
	}{
		{
			name:  "empty input",
			input: "",
			want:  nil,
		},
		{
			name:  "single page",
			input: "1",
			want:  []PageRange{{Start: 1, End: 1}},
		},
		{
			name:  "multiple single pages",
			input: "1,3,5",
			want: []PageRange{
				{Start: 1, End: 1},
				{Start: 3, End: 3},
				{Start: 5, End: 5},
			},
		},
		{
			name:  "single range",
			input: "1-5",
			want:  []PageRange{{Start: 1, End: 5}},
		},
		{
			name:  "multiple ranges",
			input: "1-3,7-10",
			want: []PageRange{
				{Start: 1, End: 3},
				{Start: 7, End: 10},
			},
		},
		{
			name:  "mixed pages and ranges",
			input: "1-3,5,7-10,15",
			want: []PageRange{
				{Start: 1, End: 3},
				{Start: 5, End: 5},
				{Start: 7, End: 10},
				{Start: 15, End: 15},
			},
		},
		{
			name:  "with spaces",
			input: "1 - 3, 5, 7 - 10",
			want: []PageRange{
				{Start: 1, End: 3},
				{Start: 5, End: 5},
				{Start: 7, End: 10},
			},
		},
		{
			name:    "invalid page number",
			input:   "abc",
			wantErr: true,
		},
		{
			name:    "invalid range",
			input:   "5-3",
			wantErr: true,
		},
		{
			name:    "zero page",
			input:   "0",
			wantErr: true,
		},
		{
			name:    "negative page",
			input:   "-1",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePageRanges(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePageRanges() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParsePageRanges() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExpandPageRanges(t *testing.T) {
	tests := []struct {
		name   string
		ranges []PageRange
		want   []int
	}{
		{
			name:   "empty input",
			ranges: nil,
			want:   nil,
		},
		{
			name:   "single page",
			ranges: []PageRange{{Start: 1, End: 1}},
			want:   []int{1},
		},
		{
			name:   "single range",
			ranges: []PageRange{{Start: 1, End: 5}},
			want:   []int{1, 2, 3, 4, 5},
		},
		{
			name: "multiple ranges",
			ranges: []PageRange{
				{Start: 1, End: 3},
				{Start: 7, End: 9},
			},
			want: []int{1, 2, 3, 7, 8, 9},
		},
		{
			name: "overlapping ranges",
			ranges: []PageRange{
				{Start: 1, End: 5},
				{Start: 3, End: 7},
			},
			want: []int{1, 2, 3, 4, 5, 6, 7},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExpandPageRanges(tt.ranges)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ExpandPageRanges() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidatePageNumbers(t *testing.T) {
	tests := []struct {
		name       string
		pages      []int
		totalPages int
		wantErr    bool
	}{
		{
			name:       "valid pages",
			pages:      []int{1, 2, 3},
			totalPages: 10,
			wantErr:    false,
		},
		{
			name:       "page out of range",
			pages:      []int{1, 11},
			totalPages: 10,
			wantErr:    true,
		},
		{
			name:       "zero page",
			pages:      []int{0, 1},
			totalPages: 10,
			wantErr:    true,
		},
		{
			name:       "empty pages",
			pages:      []int{},
			totalPages: 10,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePageNumbers(tt.pages, tt.totalPages)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePageNumbers() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFormatPageRanges(t *testing.T) {
	tests := []struct {
		name  string
		pages []int
		want  string
	}{
		{
			name:  "empty",
			pages: []int{},
			want:  "",
		},
		{
			name:  "single page",
			pages: []int{1},
			want:  "1",
		},
		{
			name:  "consecutive pages",
			pages: []int{1, 2, 3, 4, 5},
			want:  "1-5",
		},
		{
			name:  "non-consecutive pages",
			pages: []int{1, 3, 5},
			want:  "1,3,5",
		},
		{
			name:  "mixed",
			pages: []int{1, 2, 3, 5, 7, 8, 9},
			want:  "1-3,5,7-9",
		},
		{
			name:  "unsorted",
			pages: []int{5, 1, 3, 2, 4},
			want:  "1-5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatPageRanges(tt.pages)
			if got != tt.want {
				t.Errorf("FormatPageRanges() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseReorderSequence(t *testing.T) {
	tests := []struct {
		name        string
		spec        string
		totalPages  int
		want        []int
		wantErr     bool
		errContains string
	}{
		// Basic cases
		{name: "single page", spec: "1", totalPages: 5, want: []int{1}},
		{name: "multiple pages", spec: "1,3,5", totalPages: 5, want: []int{1, 3, 5}},
		{name: "simple range", spec: "1-3", totalPages: 5, want: []int{1, 2, 3}},
		{name: "reverse range", spec: "3-1", totalPages: 5, want: []int{3, 2, 1}},

		// end keyword
		{name: "end keyword", spec: "end", totalPages: 5, want: []int{5}},
		{name: "range to end", spec: "3-end", totalPages: 5, want: []int{3, 4, 5}},
		{name: "end to start", spec: "end-1", totalPages: 5, want: []int{5, 4, 3, 2, 1}},

		// Reorder scenarios
		{name: "move page 3 to front", spec: "3,1,2,4,5", totalPages: 5, want: []int{3, 1, 2, 4, 5}},
		{name: "reverse all", spec: "5-1", totalPages: 5, want: []int{5, 4, 3, 2, 1}},
		{name: "duplicate page", spec: "1,2,1", totalPages: 5, want: []int{1, 2, 1}},

		// Edge cases
		{name: "single page document", spec: "1", totalPages: 1, want: []int{1}},
		{name: "whitespace handling", spec: " 1 , 2 , 3 ", totalPages: 5, want: []int{1, 2, 3}},
		{name: "last page only", spec: "end", totalPages: 1, want: []int{1}},

		// Error cases
		{name: "empty spec", spec: "", totalPages: 5, wantErr: true, errContains: "empty"},
		{name: "invalid page zero", spec: "0", totalPages: 5, wantErr: true, errContains: "out of range"},
		{name: "page exceeds total", spec: "10", totalPages: 5, wantErr: true, errContains: "out of range"},
		{name: "invalid character", spec: "abc", totalPages: 5, wantErr: true, errContains: "invalid"},
		{name: "negative page in range", spec: "-1-5", totalPages: 5, wantErr: true, errContains: "invalid"},
		{name: "invalid total pages zero", spec: "1", totalPages: 0, wantErr: true, errContains: "invalid total"},
		{name: "invalid total pages negative", spec: "1", totalPages: -1, wantErr: true, errContains: "invalid total"},
		{name: "only whitespace", spec: "   ", totalPages: 5, wantErr: true, errContains: "no pages"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseReorderSequence(tt.spec, tt.totalPages)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errContains)
				} else if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error = %q, want containing %q", err.Error(), tt.errContains)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
