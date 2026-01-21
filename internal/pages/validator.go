package pages

import "fmt"

// ValidatePageNumbers checks if all page numbers are within the valid range.
func ValidatePageNumbers(pages []int, totalPages int) error {
	for _, p := range pages {
		if err := validatePageInRange(p, totalPages); err != nil {
			return fmt.Errorf("page %d is out of range (document has %d pages)", p, totalPages)
		}
	}
	return nil
}

// validatePageInRange checks if a page number is within valid bounds.
func validatePageInRange(page, totalPages int) error {
	if page < 1 || page > totalPages {
		return fmt.Errorf("page %d out of range: document has %d pages (valid: 1-%d, or 'end')",
			page, totalPages, totalPages)
	}
	return nil
}
