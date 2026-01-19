package util

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// PageRange represents a range of pages
type PageRange struct {
	Start int
	End   int
}

// ParsePageRanges parses a page range string like "1-5,7,10-12"
// Returns a slice of PageRange structs
func ParsePageRanges(input string) ([]PageRange, error) {
	if input == "" {
		return nil, nil
	}

	var ranges []PageRange
	parts := strings.Split(input, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		if strings.Contains(part, "-") {
			// Range like "1-5"
			bounds := strings.SplitN(part, "-", 2)
			if len(bounds) != 2 {
				return nil, fmt.Errorf("invalid page range: %s", part)
			}

			start, err := strconv.Atoi(strings.TrimSpace(bounds[0]))
			if err != nil {
				return nil, fmt.Errorf("invalid page number in range %s: %v", part, err)
			}

			end, err := strconv.Atoi(strings.TrimSpace(bounds[1]))
			if err != nil {
				return nil, fmt.Errorf("invalid page number in range %s: %v", part, err)
			}

			if start < 1 || end < 1 {
				return nil, fmt.Errorf("page numbers must be positive: %s", part)
			}

			if start > end {
				return nil, fmt.Errorf("invalid range (start > end): %s", part)
			}

			ranges = append(ranges, PageRange{Start: start, End: end})
		} else {
			// Single page like "7"
			page, err := strconv.Atoi(part)
			if err != nil {
				return nil, fmt.Errorf("invalid page number: %s", part)
			}

			if page < 1 {
				return nil, fmt.Errorf("page numbers must be positive: %d", page)
			}

			ranges = append(ranges, PageRange{Start: page, End: page})
		}
	}

	return ranges, nil
}

// ExpandPageRanges expands a slice of PageRange to individual page numbers
func ExpandPageRanges(ranges []PageRange) []int {
	var pages []int
	seen := make(map[int]bool)

	for _, r := range ranges {
		for p := r.Start; p <= r.End; p++ {
			if !seen[p] {
				pages = append(pages, p)
				seen[p] = true
			}
		}
	}

	return pages
}

// ParseAndExpandPages parses a page range string and returns individual page numbers
func ParseAndExpandPages(input string) ([]int, error) {
	ranges, err := ParsePageRanges(input)
	if err != nil {
		return nil, err
	}
	return ExpandPageRanges(ranges), nil
}

// ValidatePageNumbers checks if all page numbers are within the valid range
func ValidatePageNumbers(pages []int, totalPages int) error {
	for _, p := range pages {
		if p < 1 || p > totalPages {
			return fmt.Errorf("page %d is out of range (document has %d pages)", p, totalPages)
		}
	}
	return nil
}

// FormatPageRanges converts a slice of page numbers back to a compact range string
func FormatPageRanges(pages []int) string {
	if len(pages) == 0 {
		return ""
	}

	// Sort and deduplicate
	sorted := make([]int, len(pages))
	copy(sorted, pages)
	sort.Ints(sorted)

	var result []string
	start := sorted[0]
	end := sorted[0]

	for i := 1; i < len(sorted); i++ {
		if sorted[i] == end+1 {
			end = sorted[i]
		} else if sorted[i] > end+1 {
			result = append(result, formatRange(start, end))
			start = sorted[i]
			end = sorted[i]
		}
		// Skip duplicates (sorted[i] == end)
	}

	result = append(result, formatRange(start, end))
	return strings.Join(result, ",")
}

// formatRange formats a single page range as a string
func formatRange(start, end int) string {
	if start == end {
		return strconv.Itoa(start)
	}
	return fmt.Sprintf("%d-%d", start, end)
}
