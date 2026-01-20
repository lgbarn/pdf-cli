package util

import (
	"reflect"
	"strings"
	"testing"
)

func TestParsePageRanges(t *testing.T) {
	tests := []struct {
		input   string
		want    []PageRange
		wantErr bool
	}{
		{"", nil, false},
		{"1", []PageRange{{1, 1}}, false},
		{"1,3,5", []PageRange{{1, 1}, {3, 3}, {5, 5}}, false},
		{"1-5", []PageRange{{1, 5}}, false},
		{"1-3,7-10", []PageRange{{1, 3}, {7, 10}}, false},
		{"1-3,5,7-10,15", []PageRange{{1, 3}, {5, 5}, {7, 10}, {15, 15}}, false},
		{"1 - 3, 5, 7 - 10", []PageRange{{1, 3}, {5, 5}, {7, 10}}, false},
		{"01", []PageRange{{1, 1}}, false},
		{"01-05", []PageRange{{1, 5}}, false},
		{"001,002,003", []PageRange{{1, 1}, {2, 2}, {3, 3}}, false},
		{"100", []PageRange{{100, 100}}, false},
		{"100-200", []PageRange{{100, 200}}, false},
		{"1,,2", []PageRange{{1, 1}, {2, 2}}, false},
		{",1,2,", []PageRange{{1, 1}, {2, 2}}, false},
		{"abc", nil, true},
		{"5-3", nil, true},
		{"0", nil, true},
		{"-1", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParsePageRanges(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExpandPageRanges(t *testing.T) {
	tests := []struct {
		ranges []PageRange
		want   []int
	}{
		{nil, nil},
		{[]PageRange{{1, 1}}, []int{1}},
		{[]PageRange{{1, 5}}, []int{1, 2, 3, 4, 5}},
		{[]PageRange{{1, 3}, {7, 9}}, []int{1, 2, 3, 7, 8, 9}},
		{[]PageRange{{1, 5}, {3, 7}}, []int{1, 2, 3, 4, 5, 6, 7}},
		{[]PageRange{{1, 3}, {2, 4}}, []int{1, 2, 3, 4}},
	}

	for _, tt := range tests {
		got := ExpandPageRanges(tt.ranges)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("ExpandPageRanges(%v) = %v, want %v", tt.ranges, got, tt.want)
		}
	}
}

func TestValidatePageNumbers(t *testing.T) {
	tests := []struct {
		pages      []int
		totalPages int
		wantErr    bool
	}{
		{[]int{1, 2, 3}, 10, false},
		{[]int{}, 10, false},
		{[]int{1}, 1, false},
		{[]int{100}, 100, false},
		{[]int{1, 11}, 10, true},
		{[]int{0, 1}, 10, true},
		{[]int{101}, 100, true},
		{[]int{-1}, 10, true},
	}

	for _, tt := range tests {
		if err := ValidatePageNumbers(tt.pages, tt.totalPages); (err != nil) != tt.wantErr {
			t.Errorf("ValidatePageNumbers(%v, %d) error = %v, wantErr %v", tt.pages, tt.totalPages, err, tt.wantErr)
		}
	}
}

func TestFormatPageRanges(t *testing.T) {
	tests := []struct {
		pages []int
		want  string
	}{
		{[]int{}, ""},
		{[]int{1}, "1"},
		{[]int{1, 2, 3, 4, 5}, "1-5"},
		{[]int{1, 3, 5}, "1,3,5"},
		{[]int{1, 2, 3, 5, 7, 8, 9}, "1-3,5,7-9"},
		{[]int{5, 1, 3, 2, 4}, "1-5"},
		{[]int{1, 1, 2, 2, 3}, "1-3"},
		{[]int{1, 2}, "1-2"},
		{[]int{1, 3}, "1,3"},
		{[]int{1, 2, 5, 6, 7}, "1-2,5-7"},
	}

	for _, tt := range tests {
		if got := FormatPageRanges(tt.pages); got != tt.want {
			t.Errorf("FormatPageRanges(%v) = %q, want %q", tt.pages, got, tt.want)
		}
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
		{"single page", "1", 5, []int{1}, false, ""},
		{"multiple pages", "1,3,5", 5, []int{1, 3, 5}, false, ""},
		{"simple range", "1-3", 5, []int{1, 2, 3}, false, ""},
		{"reverse range", "3-1", 5, []int{3, 2, 1}, false, ""},
		{"end keyword", "end", 5, []int{5}, false, ""},
		{"range to end", "3-end", 5, []int{3, 4, 5}, false, ""},
		{"end to start", "end-1", 5, []int{5, 4, 3, 2, 1}, false, ""},
		{"move page 3 to front", "3,1,2,4,5", 5, []int{3, 1, 2, 4, 5}, false, ""},
		{"reverse all", "5-1", 5, []int{5, 4, 3, 2, 1}, false, ""},
		{"duplicate page", "1,2,1", 5, []int{1, 2, 1}, false, ""},
		{"single page document", "1", 1, []int{1}, false, ""},
		{"whitespace handling", " 1 , 2 , 3 ", 5, []int{1, 2, 3}, false, ""},
		{"last page only", "end", 1, []int{1}, false, ""},
		{"end-1 with 10 pages", "end-1", 10, []int{10, 9, 8, 7, 6, 5, 4, 3, 2, 1}, false, ""},
		{"empty spec", "", 5, nil, true, "empty"},
		{"invalid page zero", "0", 5, nil, true, "out of range"},
		{"page exceeds total", "10", 5, nil, true, "out of range"},
		{"invalid character", "abc", 5, nil, true, "invalid"},
		{"negative page in range", "-1-5", 5, nil, true, "invalid"},
		{"invalid total pages zero", "1", 0, nil, true, "invalid total"},
		{"invalid total pages negative", "1", -1, nil, true, "invalid total"},
		{"only whitespace", "   ", 5, nil, true, "no pages"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseReorderSequence(tt.spec, tt.totalPages)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q", tt.errContains)
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

func TestParseAndExpandPages(t *testing.T) {
	tests := []struct {
		input   string
		want    []int
		wantErr bool
	}{
		{"", nil, false},
		{"1", []int{1}, false},
		{"1-3", []int{1, 2, 3}, false},
		{"1,3-5,7", []int{1, 3, 4, 5, 7}, false},
		{"abc", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseAndExpandPages(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPageRangeStruct(t *testing.T) {
	pr := PageRange{Start: 1, End: 5}
	if pr.Start != 1 || pr.End != 5 {
		t.Errorf("PageRange = {%d, %d}, want {1, 5}", pr.Start, pr.End)
	}
}
