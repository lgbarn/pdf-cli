# Comprehensive Refactoring Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Refactor pdf-cli to improve architecture, test coverage, developer experience, and add operational features while maintaining backward compatibility.

**Architecture:** Six-phase approach starting with low-risk foundation changes (linting, mocks, docs), then progressively refactoring packages (util → pdf → commands), followed by new operational features, and finally test coverage improvements.

**Tech Stack:** Go 1.24, Cobra CLI, pdfcpu, log/slog, YAML config

---

## Phase 1: Foundation (Low Risk)

### Task 1.1: Enhance Linting Configuration

**Files:**
- Modify: `.golangci.yaml`

**Step 1: Update linter configuration**

Replace the contents of `.golangci.yaml` with:

```yaml
version: "2"

run:
  timeout: 5m

linters:
  default: none
  enable:
    # Original linters
    - govet
    - ineffassign
    - staticcheck
    - unused
    # New linters
    - gofmt
    - goimports
    - misspell
    - gocritic
    - revive
    - errcheck
    - gosimple
    - typecheck

linters-settings:
  gocritic:
    enabled-tags:
      - diagnostic
      - style
    disabled-checks:
      - ifElseChain
      - whyNoLint

  revive:
    rules:
      - name: blank-imports
      - name: context-as-argument
      - name: context-keys-type
      - name: dot-imports
      - name: error-return
      - name: error-strings
      - name: error-naming
      - name: exported
      - name: increment-decrement
      - name: var-naming
      - name: package-comments
        disabled: true
      - name: range
      - name: receiver-naming
      - name: time-naming
      - name: unexported-return
      - name: indent-error-flow
      - name: errorf

  misspell:
    locale: US

issues:
  exclude-use-default: false
  max-issues-per-linter: 0
  max-same-issues: 0
```

**Step 2: Run linter to check current state**

Run: `cd /Users/lgbarn/Personal/pdf-cli/.worktrees/refactor && golangci-lint run 2>&1 | head -50`

Note any issues that need fixing before proceeding.

**Step 3: Fix any linting issues found**

Address issues reported by the new linters. Common fixes:
- `goimports`: Add missing imports or remove unused ones
- `misspell`: Fix typos in comments/strings
- `errcheck`: Handle unchecked errors
- `gocritic`: Address code style suggestions

**Step 4: Verify linting passes**

Run: `cd /Users/lgbarn/Personal/pdf-cli/.worktrees/refactor && golangci-lint run`

Expected: No output (all checks pass)

**Step 5: Commit**

```bash
git add .golangci.yaml
git commit -m "chore: enhance linting configuration

Add gofmt, goimports, misspell, gocritic, revive, errcheck,
gosimple, and typecheck linters for better code quality."
```

---

### Task 1.2: Add Makefile Improvements

**Files:**
- Modify: `Makefile`

**Step 1: Add new targets to Makefile**

Add these targets after the existing `lint` target:

```makefile
# Lint and fix issues
.PHONY: lint-fix
lint-fix:
	golangci-lint run --fix

# Run all checks (comprehensive pre-commit)
.PHONY: check-all
check-all: fmt vet lint test

# Show test coverage percentage
.PHONY: coverage
coverage:
	@GO111MODULE=on $(GOTEST) -coverprofile=coverage.out ./... 2>/dev/null
	@go tool cover -func=coverage.out | grep total | awk '{print "Total coverage: " $$3}'

# Check coverage meets threshold (75%)
.PHONY: coverage-check
coverage-check:
	@GO111MODULE=on $(GOTEST) -coverprofile=coverage.out ./... 2>/dev/null
	@coverage=$$(go tool cover -func=coverage.out | grep total | awk '{print $$3}' | tr -d '%'); \
	if [ $$(echo "$$coverage < 75" | bc -l) -eq 1 ]; then \
		echo "Coverage $$coverage% is below 75% threshold"; \
		exit 1; \
	else \
		echo "Coverage $$coverage% meets 75% threshold"; \
	fi
```

**Step 2: Update help target**

Add these lines to the help target output:

```makefile
	@echo "  make lint-fix     Run linter with auto-fix"
	@echo "  make check-all    Run all checks (fmt, vet, lint, test)"
	@echo "  make coverage     Show test coverage percentage"
	@echo "  make coverage-check Check coverage meets 75% threshold"
```

**Step 3: Verify new targets work**

Run: `cd /Users/lgbarn/Personal/pdf-cli/.worktrees/refactor && make coverage`

Expected: Shows current coverage percentage

**Step 4: Commit**

```bash
git add Makefile
git commit -m "chore: add lint-fix, check-all, coverage targets to Makefile"
```

---

### Task 1.3: Create Test Infrastructure Package

**Files:**
- Create: `internal/testing/doc.go`
- Create: `internal/testing/fixtures.go`
- Create: `internal/testing/mock_pdf.go`
- Create: `internal/testing/mock_ocr.go`

**Step 1: Create testing package doc**

Create `internal/testing/doc.go`:

```go
// Package testing provides test utilities, mocks, and fixtures for pdf-cli tests.
//
// This package should only be imported in test files (*_test.go).
package testing
```

**Step 2: Create fixtures helper**

Create `internal/testing/fixtures.go`:

```go
package testing

import (
	"os"
	"path/filepath"
	"runtime"
)

// TestdataDir returns the path to the testdata directory.
// It handles being called from any package by finding the project root.
func TestdataDir() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("failed to get caller information")
	}
	// Go up from internal/testing to project root
	projectRoot := filepath.Dir(filepath.Dir(filepath.Dir(filename)))
	return filepath.Join(projectRoot, "testdata")
}

// SamplePDF returns the path to sample.pdf in testdata.
func SamplePDF() string {
	return filepath.Join(TestdataDir(), "sample.pdf")
}

// EncryptedPDF returns the path to encrypted.pdf in testdata.
func EncryptedPDF() string {
	return filepath.Join(TestdataDir(), "encrypted.pdf")
}

// TempDir creates a temporary directory for test artifacts.
// Returns the path and a cleanup function.
func TempDir(prefix string) (string, func()) {
	dir, err := os.MkdirTemp("", "pdf-cli-test-"+prefix+"-")
	if err != nil {
		panic("failed to create temp dir: " + err.Error())
	}
	return dir, func() { os.RemoveAll(dir) }
}

// TempFile creates a temporary file with the given content.
// Returns the path and a cleanup function.
func TempFile(prefix, content string) (string, func()) {
	f, err := os.CreateTemp("", "pdf-cli-test-"+prefix+"-*.pdf")
	if err != nil {
		panic("failed to create temp file: " + err.Error())
	}
	if content != "" {
		if _, err := f.WriteString(content); err != nil {
			f.Close()
			os.Remove(f.Name())
			panic("failed to write temp file: " + err.Error())
		}
	}
	f.Close()
	return f.Name(), func() { os.Remove(f.Name()) }
}

// CopyFile copies a file from src to dst for test isolation.
func CopyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}
```

**Step 3: Create PDF operations mock**

Create `internal/testing/mock_pdf.go`:

```go
package testing

import "errors"

// MockPDFOps provides mock PDF operations for testing commands
// without requiring actual PDF files.
type MockPDFOps struct {
	// PageCountResult is returned by PageCount
	PageCountResult int
	// PageCountError is returned by PageCount if set
	PageCountError error
	// CompressError is returned by Compress if set
	CompressError error
	// MergeError is returned by Merge if set
	MergeError error
	// SplitError is returned by Split if set
	SplitError error
	// ExtractTextResult is returned by ExtractText
	ExtractTextResult string
	// ExtractTextError is returned by ExtractText if set
	ExtractTextError error
	// Calls tracks which methods were called
	Calls []string
}

// NewMockPDFOps creates a MockPDFOps with sensible defaults.
func NewMockPDFOps() *MockPDFOps {
	return &MockPDFOps{
		PageCountResult:   10,
		ExtractTextResult: "Sample extracted text",
		Calls:             make([]string, 0),
	}
}

// PageCount mocks pdf.PageCount.
func (m *MockPDFOps) PageCount(file, password string) (int, error) {
	m.Calls = append(m.Calls, "PageCount:"+file)
	if m.PageCountError != nil {
		return 0, m.PageCountError
	}
	return m.PageCountResult, nil
}

// Compress mocks pdf.Compress.
func (m *MockPDFOps) Compress(input, output, password string) error {
	m.Calls = append(m.Calls, "Compress:"+input+"->"+output)
	return m.CompressError
}

// Merge mocks pdf.Merge.
func (m *MockPDFOps) Merge(inputs []string, output, password string) error {
	m.Calls = append(m.Calls, "Merge:"+output)
	return m.MergeError
}

// Split mocks pdf.Split.
func (m *MockPDFOps) Split(input, outputDir, password string) error {
	m.Calls = append(m.Calls, "Split:"+input)
	return m.SplitError
}

// ExtractText mocks pdf.ExtractText.
func (m *MockPDFOps) ExtractText(input, password string, pages []int) (string, error) {
	m.Calls = append(m.Calls, "ExtractText:"+input)
	if m.ExtractTextError != nil {
		return "", m.ExtractTextError
	}
	return m.ExtractTextResult, nil
}

// Reset clears the call history.
func (m *MockPDFOps) Reset() {
	m.Calls = make([]string, 0)
}

// AssertCalled checks if a method was called with the given prefix.
func (m *MockPDFOps) AssertCalled(prefix string) bool {
	for _, call := range m.Calls {
		if len(call) >= len(prefix) && call[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}

// ErrMockPasswordRequired is a mock error for password-protected files.
var ErrMockPasswordRequired = errors.New("mock: password required")

// ErrMockCorrupted is a mock error for corrupted files.
var ErrMockCorrupted = errors.New("mock: file corrupted")
```

**Step 4: Create OCR backend mock**

Create `internal/testing/mock_ocr.go`:

```go
package testing

import "context"

// MockOCRBackend provides a mock OCR backend for testing.
type MockOCRBackend struct {
	// NameResult is returned by Name
	NameResult string
	// AvailableResult is returned by Available
	AvailableResult bool
	// ProcessImageResult is returned by ProcessImage
	ProcessImageResult string
	// ProcessImageError is returned by ProcessImage if set
	ProcessImageError error
	// Calls tracks which methods were called
	Calls []string
}

// NewMockOCRBackend creates a MockOCRBackend with sensible defaults.
func NewMockOCRBackend() *MockOCRBackend {
	return &MockOCRBackend{
		NameResult:         "mock",
		AvailableResult:    true,
		ProcessImageResult: "Mock OCR extracted text",
		Calls:              make([]string, 0),
	}
}

// Name returns the backend name.
func (m *MockOCRBackend) Name() string {
	m.Calls = append(m.Calls, "Name")
	return m.NameResult
}

// Available returns whether the backend is available.
func (m *MockOCRBackend) Available() bool {
	m.Calls = append(m.Calls, "Available")
	return m.AvailableResult
}

// ProcessImage mocks OCR processing.
func (m *MockOCRBackend) ProcessImage(ctx context.Context, imagePath, lang string) (string, error) {
	m.Calls = append(m.Calls, "ProcessImage:"+imagePath)
	if m.ProcessImageError != nil {
		return "", m.ProcessImageError
	}
	return m.ProcessImageResult, nil
}

// Close mocks backend cleanup.
func (m *MockOCRBackend) Close() error {
	m.Calls = append(m.Calls, "Close")
	return nil
}

// Reset clears the call history.
func (m *MockOCRBackend) Reset() {
	m.Calls = make([]string, 0)
}
```

**Step 5: Verify package compiles**

Run: `cd /Users/lgbarn/Personal/pdf-cli/.worktrees/refactor && GO111MODULE=on go build ./internal/testing/...`

Expected: No errors

**Step 6: Commit**

```bash
git add internal/testing/
git commit -m "test: add testing infrastructure package with mocks and fixtures

- Add fixtures.go with test file helpers
- Add mock_pdf.go for mocking PDF operations
- Add mock_ocr.go for mocking OCR backends"
```

---

### Task 1.4: Create Architecture Documentation

**Files:**
- Create: `docs/architecture.md`

**Step 1: Create architecture documentation**

Create `docs/architecture.md`:

```markdown
# pdf-cli Architecture

## Overview

pdf-cli is a command-line tool for PDF manipulation built with Go. It follows a layered architecture with clear separation between CLI handling, business logic, and external dependencies.

## Package Structure

```
cmd/pdf/              Entry point
internal/
├── cli/              CLI framework (Cobra wrapper, flags, output)
├── commands/         Command implementations (14 commands)
├── pdf/              PDF operations (wrapper around pdfcpu)
├── ocr/              OCR engine with pluggable backends
├── util/             Shared utilities (files, pages, errors, output)
└── testing/          Test infrastructure (mocks, fixtures)
```

## Dependency Graph

```
                    cmd/pdf/main.go
                          │
                          ▼
                    internal/cli
                          │
              ┌───────────┼───────────┐
              ▼           ▼           ▼
         commands/      (flags)    (output)
              │
    ┌─────────┼─────────┬─────────┐
    ▼         ▼         ▼         ▼
  pdf/      ocr/      util/      cli/
    │         │
    ▼         ▼
 pdfcpu   gogosseract
```

**Key principles:**
- No circular dependencies
- Commands depend on core packages, not vice versa
- util/ is a leaf package with no internal dependencies
- External dependencies isolated in pdf/ and ocr/

## Package Responsibilities

### cli/
- Root command setup and version info
- Shared flag definitions (output, password, verbose, force)
- Output formatting helpers
- Shell completion

### commands/
- One file per command (merge.go, split.go, etc.)
- Orchestrates pdf/ and ocr/ operations
- Handles stdin/stdout for pipelines
- Batch processing logic

### pdf/
- Wraps pdfcpu for PDF manipulation
- Wraps ledongthuc/pdf for text extraction fallback
- Provides unified API for all PDF operations
- Handles progress reporting

### ocr/
- Dual backend architecture (native Tesseract, WASM fallback)
- Backend interface for pluggability
- Language data management
- Image-to-text conversion

### util/
- File operations and validation
- Page range parsing (supports "1-5,7,end-1")
- Error wrapping with context
- Output formatting (JSON, CSV, TSV, human)
- Progress bar utilities

### testing/
- Mock implementations for pdf/ and ocr/
- Test fixtures and helpers
- Shared test utilities

## Design Decisions

### Why pdfcpu + ledongthuc/pdf?
- pdfcpu: Primary library for manipulation (merge, split, rotate, compress)
- ledongthuc/pdf: Fallback for text extraction when pdfcpu returns empty

### Why dual OCR backends?
- Native Tesseract: Faster, better quality (when installed)
- WASM Tesseract: Zero dependencies, works everywhere
- Auto-detection with manual override

### Why Cobra for CLI?
- Industry standard for Go CLIs
- Built-in completion, help, flags
- Subcommand support

## Extension Points

### Adding a new command
1. Create `internal/commands/newcmd.go`
2. Define cobra.Command with init() registration
3. Use helpers from helpers.go for common patterns
4. Add tests in `newcmd_test.go`

### Adding a new OCR backend
1. Implement `Backend` interface in `internal/ocr/`
2. Add detection logic in `detect.go`
3. Register in `Engine.selectBackend()`

## Error Handling

All errors use `util.WrapError()` for consistent formatting:
- Operation context (what was being done)
- File context (which file)
- Underlying error
- User-friendly hints for common issues

## Testing Strategy

- Unit tests co-located with source files
- Integration tests in `commands_integration_test.go`
- Mocks in `internal/testing/` for isolation
- Table-driven tests for comprehensive coverage
```

**Step 2: Commit**

```bash
git add docs/architecture.md
git commit -m "docs: add architecture documentation

Document package structure, dependencies, design decisions,
and extension points."
```

---

### Task 1.5: Create CONTRIBUTING.md

**Files:**
- Create: `CONTRIBUTING.md`

**Step 1: Create contributing guide**

Create `CONTRIBUTING.md`:

```markdown
# Contributing to pdf-cli

Thank you for your interest in contributing to pdf-cli!

## Development Setup

### Prerequisites
- Go 1.21 or later
- golangci-lint (`go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`)
- (Optional) Tesseract for native OCR testing

### Getting Started

```bash
# Clone the repository
git clone https://github.com/lgbarn/pdf-cli.git
cd pdf-cli

# Download dependencies
make deps

# Run tests
make test

# Build
make build
```

## Development Workflow

### Before Making Changes

1. Create a feature branch: `git checkout -b feature/your-feature`
2. Ensure tests pass: `make test`
3. Ensure linting passes: `make lint`

### Making Changes

1. Write tests first (TDD encouraged)
2. Implement the minimal code to pass tests
3. Run `make check-all` to verify everything passes
4. Commit with descriptive messages

### Code Style

- Follow standard Go conventions
- Run `make fmt` before committing
- Use `make lint-fix` to auto-fix linting issues
- Keep functions focused and under 50 lines when possible

### Commit Messages

Follow conventional commits:
- `feat:` New feature
- `fix:` Bug fix
- `docs:` Documentation
- `test:` Test changes
- `chore:` Maintenance
- `refactor:` Code restructuring

Example: `feat: add --dry-run flag to merge command`

## Testing

### Running Tests

```bash
make test           # Run all tests
make test-coverage  # Run with coverage report
make test-race      # Run with race detection
make coverage       # Show coverage percentage
```

### Writing Tests

- Place tests in `*_test.go` files alongside source
- Use table-driven tests for multiple cases
- Use mocks from `internal/testing/` for isolation
- Test both success and error paths

Example:
```go
func TestParsePages(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    []int
        wantErr bool
    }{
        {"single page", "1", []int{1}, false},
        {"range", "1-3", []int{1, 2, 3}, false},
        {"invalid", "abc", nil, true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := ParsePages(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
            }
            if !reflect.DeepEqual(got, tt.want) {
                t.Errorf("got %v, want %v", got, tt.want)
            }
        })
    }
}
```

## Pull Request Process

1. Ensure all checks pass: `make check-all`
2. Update documentation if needed
3. Add tests for new functionality
4. Keep PRs focused on a single change
5. Respond to review feedback promptly

## Project Structure

See [docs/architecture.md](docs/architecture.md) for detailed architecture documentation.

## Questions?

Open an issue for questions or discussions.
```

**Step 2: Commit**

```bash
git add CONTRIBUTING.md
git commit -m "docs: add CONTRIBUTING.md with development guidelines"
```

---

## Phase 2: Package Reorganization

### Task 2.1: Create fileio Package

**Files:**
- Create: `internal/fileio/doc.go`
- Create: `internal/fileio/files.go`
- Create: `internal/fileio/stdio.go`
- Create: `internal/fileio/files_test.go`
- Create: `internal/fileio/stdio_test.go`

**Step 1: Create package doc**

Create `internal/fileio/doc.go`:

```go
// Package fileio provides file I/O operations and stdio handling for pdf-cli.
package fileio
```

**Step 2: Copy and adapt files.go**

Create `internal/fileio/files.go` by copying from `internal/util/files.go` and changing the package declaration to `package fileio`.

**Step 3: Copy and adapt stdio.go**

Create `internal/fileio/stdio.go` by copying from `internal/util/stdio.go` and changing the package declaration to `package fileio`.

**Step 4: Copy and adapt test files**

Copy `internal/util/files_test.go` to `internal/fileio/files_test.go` and update package to `fileio`.
Copy `internal/util/stdio_test.go` to `internal/fileio/stdio_test.go` and update package to `fileio`.

**Step 5: Run tests**

Run: `cd /Users/lgbarn/Personal/pdf-cli/.worktrees/refactor && GO111MODULE=on go test ./internal/fileio/...`

Expected: All tests pass

**Step 6: Commit**

```bash
git add internal/fileio/
git commit -m "refactor: extract fileio package from util

Move file operations and stdio handling to dedicated package."
```

---

### Task 2.2: Create pages Package

**Files:**
- Create: `internal/pages/doc.go`
- Create: `internal/pages/parser.go`
- Create: `internal/pages/validator.go`
- Create: `internal/pages/reorder.go`
- Create: `internal/pages/pages_test.go`

**Step 1: Create package doc**

Create `internal/pages/doc.go`:

```go
// Package pages provides page range parsing and validation for pdf-cli.
package pages
```

**Step 2: Split pages.go into focused files**

Extract parsing functions to `internal/pages/parser.go`:
- ParsePageRanges
- ExpandPageRanges
- ParseAndExpandPages
- expandRange
- parseEndRelative

Extract validation functions to `internal/pages/validator.go`:
- ValidatePageNumbers
- validatePageInRange

Extract reorder functions to `internal/pages/reorder.go`:
- ParseReorderSequence
- parseReorderPart

**Step 3: Copy and adapt test file**

Copy `internal/util/pages_test.go` to `internal/pages/pages_test.go` and update package to `pages`.

**Step 4: Run tests**

Run: `cd /Users/lgbarn/Personal/pdf-cli/.worktrees/refactor && GO111MODULE=on go test ./internal/pages/...`

Expected: All tests pass

**Step 5: Commit**

```bash
git add internal/pages/
git commit -m "refactor: extract pages package from util

Split page parsing into parser.go, validator.go, and reorder.go."
```

---

### Task 2.3: Create output Package

**Files:**
- Create: `internal/output/doc.go`
- Create: `internal/output/formatter.go`
- Create: `internal/output/table.go`
- Create: `internal/output/output_test.go`

**Step 1: Create package doc**

Create `internal/output/doc.go`:

```go
// Package output provides output formatting (JSON, CSV, TSV, human) for pdf-cli.
package output
```

**Step 2: Extract formatter to output/formatter.go**

Move from `internal/util/output.go`:
- OutputFormat type and constants
- ParseOutputFormat
- OutputFormatter struct and methods (except table rendering)

**Step 3: Extract table rendering to output/table.go**

Move from `internal/util/output.go`:
- printTableHuman
- printTableJSON
- printTableCSV
- columnWidths helper

**Step 4: Copy and adapt test file**

Copy `internal/util/output_test.go` to `internal/output/output_test.go` and update package to `output`.

**Step 5: Run tests**

Run: `cd /Users/lgbarn/Personal/pdf-cli/.worktrees/refactor && GO111MODULE=on go test ./internal/output/...`

Expected: All tests pass

**Step 6: Commit**

```bash
git add internal/output/
git commit -m "refactor: extract output package from util

Separate formatter logic from table rendering."
```

---

### Task 2.4: Create errors Package

**Files:**
- Create: `internal/errors/doc.go`
- Create: `internal/errors/errors.go`
- Create: `internal/errors/errors_test.go`

**Step 1: Create package doc**

Create `internal/errors/doc.go`:

```go
// Package errors provides error types and wrapping for pdf-cli.
package errors
```

**Step 2: Copy and adapt errors.go**

Copy `internal/util/errors.go` to `internal/errors/errors.go` and update package to `errors`.

**Step 3: Copy and adapt test file**

Copy `internal/util/errors_test.go` to `internal/errors/errors_test.go` and update package to `errors`.

**Step 4: Run tests**

Run: `cd /Users/lgbarn/Personal/pdf-cli/.worktrees/refactor && GO111MODULE=on go test ./internal/errors/...`

Expected: All tests pass

**Step 5: Commit**

```bash
git add internal/errors/
git commit -m "refactor: extract errors package from util"
```

---

### Task 2.5: Create progress Package

**Files:**
- Create: `internal/progress/doc.go`
- Create: `internal/progress/progress.go`
- Create: `internal/progress/progress_test.go`

**Step 1: Create package doc**

Create `internal/progress/doc.go`:

```go
// Package progress provides progress bar utilities for pdf-cli.
package progress
```

**Step 2: Copy and adapt progress.go**

Copy `internal/util/progress.go` to `internal/progress/progress.go` and update package to `progress`.

**Step 3: Copy and adapt test file**

Copy `internal/util/progress_test.go` to `internal/progress/progress_test.go` and update package to `progress`.

**Step 4: Run tests**

Run: `cd /Users/lgbarn/Personal/pdf-cli/.worktrees/refactor && GO111MODULE=on go test ./internal/progress/...`

Expected: All tests pass

**Step 5: Commit**

```bash
git add internal/progress/
git commit -m "refactor: extract progress package from util"
```

---

### Task 2.6: Update All Imports

**Files:**
- Modify: All files in `internal/cli/`, `internal/commands/`, `internal/pdf/`, `internal/ocr/`

**Step 1: Update imports across codebase**

Replace all occurrences:
- `github.com/lgbarn/pdf-cli/internal/util` → appropriate new package

Import mapping:
- File operations → `github.com/lgbarn/pdf-cli/internal/fileio`
- Page parsing → `github.com/lgbarn/pdf-cli/internal/pages`
- Output formatting → `github.com/lgbarn/pdf-cli/internal/output`
- Error handling → `github.com/lgbarn/pdf-cli/internal/errors`
- Progress bars → `github.com/lgbarn/pdf-cli/internal/progress`

**Step 2: Run tests to verify**

Run: `cd /Users/lgbarn/Personal/pdf-cli/.worktrees/refactor && GO111MODULE=on go test ./...`

Expected: All 616 tests pass

**Step 3: Commit**

```bash
git add .
git commit -m "refactor: update imports to use new packages

Replace util imports with fileio, pages, output, errors, progress."
```

---

### Task 2.7: Remove Old util Package

**Files:**
- Delete: `internal/util/` (entire directory)

**Step 1: Remove util directory**

```bash
rm -rf internal/util/
```

**Step 2: Verify build and tests**

Run: `cd /Users/lgbarn/Personal/pdf-cli/.worktrees/refactor && GO111MODULE=on go build ./... && GO111MODULE=on go test ./...`

Expected: Build succeeds, all tests pass

**Step 3: Commit**

```bash
git add -A
git commit -m "refactor: remove old util package

All functionality moved to dedicated packages:
- fileio: file operations, stdio
- pages: page range parsing
- output: output formatting
- errors: error types and wrapping
- progress: progress bar utilities"
```

---

## Phase 3: PDF Package Refactoring

### Task 3.1: Split pdf.go into Modules

**Files:**
- Modify: `internal/pdf/pdf.go` (reduce to config and API)
- Create: `internal/pdf/metadata.go`
- Create: `internal/pdf/transform.go`
- Create: `internal/pdf/encryption.go`
- Create: `internal/pdf/text.go`
- Create: `internal/pdf/watermark.go`
- Create: `internal/pdf/validation.go`

**Step 1: Create metadata.go**

Move from pdf.go:
- GetInfo
- PageCount
- GetMetadata
- SetMetadata

**Step 2: Create transform.go**

Move from pdf.go:
- Merge, MergeWithProgress
- Split, SplitByPageCount, SplitWithProgress
- ExtractPages
- Rotate
- Compress

**Step 3: Create encryption.go**

Move from pdf.go:
- Encrypt
- Decrypt

**Step 4: Create text.go**

Move from pdf.go:
- ExtractText, ExtractTextWithProgress
- extractTextPrimary, extractTextFallback
- extractPagesSequential, extractPagesParallel
- extractPageText, parseTextFromPDFContent
- extractParenString

**Step 5: Create watermark.go**

Move from pdf.go:
- AddWatermark
- AddImageWatermark

**Step 6: Create validation.go**

Move from pdf.go:
- Validate
- ValidateToBuffer

**Step 7: Reduce pdf.go**

Keep in pdf.go:
- NewConfig
- pagesToStrings helper
- Package doc comments

**Step 8: Run tests**

Run: `cd /Users/lgbarn/Personal/pdf-cli/.worktrees/refactor && GO111MODULE=on go test ./internal/pdf/...`

Expected: All pdf tests pass

**Step 9: Commit**

```bash
git add internal/pdf/
git commit -m "refactor: split pdf.go into focused modules

- metadata.go: GetInfo, PageCount, GetMetadata, SetMetadata
- transform.go: Merge, Split, ExtractPages, Rotate, Compress
- encryption.go: Encrypt, Decrypt
- text.go: text extraction with fallback
- watermark.go: AddWatermark, AddImageWatermark
- validation.go: Validate, ValidateToBuffer"
```

---

## Phase 4: Command Layer Improvements

### Task 4.1: Create StdioHandler Pattern

**Files:**
- Create: `internal/commands/patterns/doc.go`
- Create: `internal/commands/patterns/stdio.go`
- Create: `internal/commands/patterns/stdio_test.go`

**Step 1: Create patterns package doc**

Create `internal/commands/patterns/doc.go`:

```go
// Package patterns provides reusable patterns for pdf-cli commands.
package patterns
```

**Step 2: Create StdioHandler**

Create `internal/commands/patterns/stdio.go`:

```go
package patterns

import (
	"fmt"
	"os"

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
	if h.ToStdout {
		tmpFile, err := os.CreateTemp("", "pdf-cli-"+h.Operation+"-*.pdf")
		if err != nil {
			h.inputCleanup()
			return "", "", fmt.Errorf("creating temp output: %w", err)
		}
		h.outputPath = tmpFile.Name()
		_ = tmpFile.Close()
		h.outputCleanup = func() { _ = os.Remove(h.outputPath) }
	} else if h.ExplicitOutput != "" {
		h.outputPath = h.ExplicitOutput
		h.outputCleanup = func() {}
	} else {
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
```

**Step 3: Add tests**

Create `internal/commands/patterns/stdio_test.go` with tests for Setup, Finalize, Cleanup.

**Step 4: Commit**

```bash
git add internal/commands/patterns/
git commit -m "feat: add StdioHandler pattern for command pipelines"
```

---

### Task 4.2: Refactor Commands to Use StdioHandler

**Files:**
- Modify: `internal/commands/compress.go`
- Modify: `internal/commands/decrypt.go`
- Modify: `internal/commands/encrypt.go`
- Modify: `internal/commands/extract.go`
- Modify: `internal/commands/pdfa.go`
- Modify: `internal/commands/reorder.go`
- Modify: `internal/commands/rotate.go`

**Step 1: Refactor compress.go**

Replace `compressWithStdio` function:

```go
func compressWithStdio(inputArg, explicitOutput, password string, toStdout bool) error {
	handler := &patterns.StdioHandler{
		InputArg:       inputArg,
		ExplicitOutput: explicitOutput,
		ToStdout:       toStdout,
		DefaultSuffix:  "_compressed",
		Operation:      "compress",
	}
	defer handler.Cleanup()

	input, output, err := handler.Setup()
	if err != nil {
		return err
	}

	if !toStdout {
		if err := checkOutputFile(output); err != nil {
			return err
		}
	}

	if err := pdf.Compress(input, output, password); err != nil {
		return errors.WrapError("compressing file", inputArg, err)
	}

	if err := handler.Finalize(); err != nil {
		return err
	}

	if !toStdout {
		newSize, _ := fileio.GetFileSize(output)
		fmt.Fprintf(os.Stderr, "Compressed to %s (%s)\n", output, fileio.FormatFileSize(newSize))
	}
	return nil
}
```

**Step 2-7: Repeat for other commands**

Apply same pattern to decrypt, encrypt, extract, pdfa, reorder, rotate.

**Step 8: Run tests**

Run: `cd /Users/lgbarn/Personal/pdf-cli/.worktrees/refactor && GO111MODULE=on go test ./internal/commands/...`

Expected: All command tests pass

**Step 9: Commit**

```bash
git add internal/commands/
git commit -m "refactor: use StdioHandler in pipeline commands

Reduces ~200 LOC of duplicated stdin/stdout handling."
```

---

## Phase 5: Operational Features

### Task 5.1: Add Config File Support

**Files:**
- Create: `internal/config/doc.go`
- Create: `internal/config/config.go`
- Create: `internal/config/config_test.go`
- Modify: `internal/cli/cli.go`

(Detailed implementation in separate task file)

---

### Task 5.2: Add Structured Logging

**Files:**
- Create: `internal/logging/doc.go`
- Create: `internal/logging/logger.go`
- Create: `internal/logging/logger_test.go`
- Modify: `internal/cli/flags.go`

(Detailed implementation in separate task file)

---

### Task 5.3: Add Dry-Run Mode

**Files:**
- Modify: `internal/cli/flags.go`
- Modify: All modifying commands

(Detailed implementation in separate task file)

---

## Phase 6: Test Coverage Push

### Task 6.1: Add Missing Command Tests

Target: Increase commands package from 60.7% to 75%+

Focus areas:
- Batch mode variations
- Stdin/stdout paths
- Error paths (invalid files, permissions)

---

### Task 6.2: Add Missing PDF Tests

Target: Increase pdf package from 57.9% to 75%+

Focus areas:
- Text extraction edge cases
- Error paths in each module

---

### Task 6.3: Add CI Coverage Enforcement

**Files:**
- Modify: `.github/workflows/ci.yaml`

Add step to fail CI if coverage drops below 75%.

---

## Execution Summary

| Phase | Tasks | Risk | Estimated Changes |
|-------|-------|------|-------------------|
| 1. Foundation | 5 tasks | Low | ~10 files |
| 2. Package Reorg | 7 tasks | Medium | ~40 files |
| 3. PDF Refactor | 1 task | Low | ~7 files |
| 4. Commands | 2 tasks | Medium | ~10 files |
| 5. Operational | 3 tasks | Low | ~15 files |
| 6. Coverage | 3 tasks | Low | ~20 files |

**Total: 21 tasks across 6 phases**
