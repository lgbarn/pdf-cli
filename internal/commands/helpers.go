package commands

import (
	"fmt"

	"github.com/lgbarn/pdf-cli/internal/cli"
	"github.com/lgbarn/pdf-cli/internal/pdf"
	"github.com/lgbarn/pdf-cli/internal/util"
)

// checkOutputFile verifies the output file can be written.
// Returns an error if the file exists and force mode is not enabled.
func checkOutputFile(output string) error {
	if util.FileExists(output) && !cli.Force() {
		return fmt.Errorf("output file already exists: %s (use -f to overwrite)", output)
	}
	return nil
}

// parseAndValidatePages parses the pages string and validates against the PDF.
// Returns nil slice if pagesStr is empty (meaning "all pages").
func parseAndValidatePages(pagesStr, inputFile, password string) ([]int, error) {
	if pagesStr == "" {
		return nil, nil
	}

	pages, err := util.ParseAndExpandPages(pagesStr)
	if err != nil {
		return nil, fmt.Errorf("invalid page specification: %w", err)
	}

	pageCount, err := pdf.PageCount(inputFile, password)
	if err != nil {
		return nil, util.WrapError("reading file", inputFile, err)
	}

	if err := util.ValidatePageNumbers(pages, pageCount); err != nil {
		return nil, err
	}

	return pages, nil
}

// outputOrDefault returns output if non-empty, otherwise generates a default filename.
func outputOrDefault(output, inputFile, suffix string) string {
	if output != "" {
		return output
	}
	return util.GenerateOutputFilename(inputFile, suffix)
}
