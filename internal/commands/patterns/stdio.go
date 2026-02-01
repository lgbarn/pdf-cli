package patterns

import (
	"fmt"
	"os"

	"github.com/lgbarn/pdf-cli/internal/cleanup"
	"github.com/lgbarn/pdf-cli/internal/fileio"
)

// StdioHandler manages stdin/stdout for commands that support pipelines.
type StdioHandler struct {
	InputArg       string
	ExplicitOutput string
	ToStdout       bool
	DefaultSuffix  string
	Operation      string

	inputPath     string
	outputPath    string
	inputCleanup  func()
	outputCleanup func()
}

// Setup prepares input and output paths, handling stdin/stdout as needed.
// Returns input path, output path, and error.
// Call Cleanup() when done, regardless of success or failure.
func (h *StdioHandler) Setup() (input, output string, err error) {
	// Resolve input (may be stdin)
	h.inputPath, h.inputCleanup, err = fileio.ResolveInputPath(h.InputArg)
	if err != nil {
		return "", "", fmt.Errorf("resolving input: %w", err)
	}

	// Resolve output
	switch {
	case h.ToStdout:
		tmpFile, err := os.CreateTemp("", "pdf-cli-"+h.Operation+"-*.pdf")
		if err != nil {
			h.inputCleanup()
			return "", "", fmt.Errorf("creating temp output: %w", err)
		}
		h.outputPath = tmpFile.Name()
		_ = tmpFile.Close()
		unregisterTmp := cleanup.Register(h.outputPath)
		h.outputCleanup = func() {
			unregisterTmp()
			_ = os.Remove(h.outputPath)
		}
	case h.ExplicitOutput != "":
		h.outputPath = h.ExplicitOutput
		h.outputCleanup = func() {}
	default:
		h.outputPath = fileio.GenerateOutputFilename(h.InputArg, h.DefaultSuffix)
		h.outputCleanup = func() {}
	}

	return h.inputPath, h.outputPath, nil
}

// Finalize writes output to stdout if needed.
// Call this after the operation succeeds.
func (h *StdioHandler) Finalize() error {
	if h.ToStdout {
		return fileio.WriteToStdout(h.outputPath)
	}
	return nil
}

// Cleanup releases all resources.
// Safe to call multiple times.
func (h *StdioHandler) Cleanup() {
	if h.inputCleanup != nil {
		h.inputCleanup()
		h.inputCleanup = nil
	}
	if h.outputCleanup != nil {
		h.outputCleanup()
		h.outputCleanup = nil
	}
}

// OutputPath returns the resolved output path.
func (h *StdioHandler) OutputPath() string {
	return h.outputPath
}

// InputPath returns the resolved input path.
func (h *StdioHandler) InputPath() string {
	return h.inputPath
}
