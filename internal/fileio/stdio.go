package fileio

import (
	"fmt"
	"io"
	"os"

	signalcleanup "github.com/lgbarn/pdf-cli/internal/cleanup"
	"golang.org/x/term"
)

// StdinIndicator is the conventional indicator for stdin input.
const StdinIndicator = "-"

// IsStdinInput returns true if the input path indicates stdin.
func IsStdinInput(path string) bool {
	return path == StdinIndicator
}

// IsStdinPiped returns true if stdin has data piped to it (not a terminal).
func IsStdinPiped() bool {
	return !term.IsTerminal(int(os.Stdin.Fd()))
}

// ReadFromStdin reads stdin to a temporary file and returns the path and cleanup function.
// The caller is responsible for calling the cleanup function when done.
func ReadFromStdin() (path string, cleanup func(), err error) {
	if !IsStdinPiped() {
		return "", nil, fmt.Errorf("stdin is empty or not piped")
	}

	// Create temp file with .pdf extension (required by pdfcpu)
	tmpFile, err := os.CreateTemp("", "pdf-cli-stdin-*.pdf")
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temp file: %w", err)
	}

	tmpPath := tmpFile.Name()
	unregisterTmp := signalcleanup.Register(tmpPath)
	cleanup = func() {
		unregisterTmp()
		_ = os.Remove(tmpPath)
	}

	// Copy stdin to temp file
	if _, err := io.Copy(tmpFile, os.Stdin); err != nil {
		_ = tmpFile.Close()
		cleanup()
		return "", nil, fmt.Errorf("failed to read stdin: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		cleanup()
		return "", nil, fmt.Errorf("failed to close temp file: %w", err)
	}

	return tmpPath, cleanup, nil
}

// WriteToStdout writes a file's contents to stdout.
func WriteToStdout(path string) error {
	f, err := os.Open(path) // #nosec G304 -- path comes from temp files we control
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(os.Stdout, f); err != nil {
		return fmt.Errorf("failed to write to stdout: %w", err)
	}

	return nil
}

// ResolveInputPath resolves an input path, handling stdin indicator.
// Returns the actual path to use and a cleanup function.
// If the input is not stdin, cleanup is a no-op.
func ResolveInputPath(input string) (path string, cleanup func(), err error) {
	if !IsStdinInput(input) {
		return input, func() {}, nil
	}
	return ReadFromStdin()
}
