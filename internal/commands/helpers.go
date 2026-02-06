package commands

import (
	"errors"
	"fmt"

	"github.com/lgbarn/pdf-cli/internal/cli"
	"github.com/lgbarn/pdf-cli/internal/fileio"
	"github.com/lgbarn/pdf-cli/internal/pages"
	"github.com/lgbarn/pdf-cli/internal/pdf"
	"github.com/lgbarn/pdf-cli/internal/pdferrors"
)

// Output filename suffixes for batch operations.
const (
	SuffixEncrypted   = "_encrypted"
	SuffixDecrypted   = "_decrypted"
	SuffixCompressed  = "_compressed"
	SuffixRotated     = "_rotated"
	SuffixWatermarked = "_watermarked"
	SuffixReordered   = "_reordered"
)

// checkOutputFile verifies the output file can be written.
// Returns an error if the file exists and force mode is not enabled.
func checkOutputFile(output string) error {
	if fileio.FileExists(output) && !cli.Force() {
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

	pageNums, err := pages.ParseAndExpandPages(pagesStr)
	if err != nil {
		return nil, fmt.Errorf("invalid page specification: %w", err)
	}

	pageCount, err := pdf.PageCount(inputFile, password)
	if err != nil {
		return nil, pdferrors.WrapError("reading file", inputFile, err)
	}

	if err := pages.ValidatePageNumbers(pageNums, pageCount); err != nil {
		return nil, err
	}

	return pageNums, nil
}

// outputOrDefault returns output if non-empty, otherwise generates a default filename.
func outputOrDefault(output, inputFile, suffix string) string {
	if output != "" {
		return output
	}
	return fileio.GenerateOutputFilename(inputFile, suffix)
}

// validateBatchOutput checks if -o flag is used with multiple files.
// Returns an error if so, since -o is only allowed with single file operations.
func validateBatchOutput(files []string, output, suffix string) error {
	if len(files) > 1 && output != "" {
		return fmt.Errorf("cannot use -o with multiple files; output files will be named with '%s' suffix", suffix)
	}
	return nil
}

// sanitizeInputArgs validates and cleans input file path arguments.
func sanitizeInputArgs(args []string) ([]string, error) {
	sanitized, err := fileio.SanitizePaths(args)
	if err != nil {
		return nil, fmt.Errorf("invalid file path: %w", err)
	}
	return sanitized, nil
}

// sanitizeOutputPath validates and cleans an output file path from flags.
// Returns the path unchanged if empty or stdin marker "-".
func sanitizeOutputPath(output string) (string, error) {
	if output == "" || output == "-" {
		return output, nil
	}
	cleaned, err := fileio.SanitizePath(output)
	if err != nil {
		return "", fmt.Errorf("invalid output path: %w", err)
	}
	return cleaned, nil
}

// processBatch processes multiple files with the given processor function.
// Each file is processed independently, and all errors are collected and joined.
func processBatch(files []string, processor func(file string) error) error {
	var errs []error
	for _, file := range files {
		if err := processor(file); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", file, err))
		}
	}
	return errors.Join(errs...)
}
