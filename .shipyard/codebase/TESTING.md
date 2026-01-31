# Testing Infrastructure

This document describes the test framework, coverage requirements, test patterns, and testing utilities used in pdf-cli.

## Test Framework

### Core Testing Tools

- **Framework**: Go standard library `testing` package
- **Test runner**: `go test` with race detection support
- **Coverage**: Built-in Go coverage tools
- **CI**: GitHub Actions with golangci-lint integration
- **Mock utilities**: Custom mocks in `internal/testing/`

### Test Execution

**Via Makefile:**
```bash
make test              # Run all tests
make test-coverage     # Generate HTML coverage report
make test-race         # Run with race detector
make coverage          # Show coverage percentage
make coverage-check    # Enforce 75% threshold
```

**Direct commands:**
```bash
go test ./...                                    # All tests
go test -v ./...                                 # Verbose
go test -race ./...                              # Race detection
go test -coverprofile=coverage.out ./...         # Coverage
go tool cover -html=coverage.out                 # HTML report
```

## Coverage Requirements

### Coverage Metrics

**Overall project coverage:** 81.5%

**Per-package coverage:**
```
internal/logging       100.0%  ✓ Full coverage
internal/progress      100.0%  ✓ Full coverage
internal/pdferrors      97.1%  ✓ Excellent
internal/output         96.5%  ✓ Excellent
internal/cli            95.2%  ✓ Excellent
internal/pages          94.4%  ✓ Excellent
internal/commands/patterns 89.3%  ✓ Good
internal/config         86.4%  ✓ Good
internal/pdf            85.6%  ✓ Good
internal/commands       82.8%  ✓ Good
internal/fileio         78.2%  ✓ Adequate
internal/ocr            75.5%  ✓ Meets threshold

cmd/pdf                  0.0%  (main package, expected)
internal/testing         0.0%  (test utilities, expected)
```

### Coverage Enforcement

**Threshold:** 75% minimum coverage enforced in CI

**Makefile target:**
```makefile
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

**CI enforcement (`.github/workflows/ci.yaml`):**
```yaml
- name: Run tests with coverage
  run: go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

- name: Check coverage threshold (75%)
  run: |
    COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | tr -d '%')
    echo "Total coverage: ${COVERAGE}%"
    if awk "BEGIN {exit !($COVERAGE < 75)}"; then
      echo "::error::Coverage ${COVERAGE}% is below 75% threshold"
      exit 1
    fi
```

## Test Organization

### File Placement

Tests are co-located with source files:
```
internal/fileio/
├── files.go
├── files_test.go
├── stdio.go
├── stdio_test.go
└── doc.go
```

**Statistics:**
- 31 test files total
- 405 test functions
- 56 source files (non-test)
- ~7.2 test functions per source file

### Test File Types

1. **Unit tests**: `*_test.go` alongside source
   - Example: `/Users/lgbarn/Personal/pdf-cli/internal/fileio/files_test.go`

2. **Integration tests**: `*_integration_test.go`
   - Example: `/Users/lgbarn/Personal/pdf-cli/internal/commands/commands_integration_test.go`

3. **Additional coverage tests**: `additional_coverage_test.go`
   - Example: `/Users/lgbarn/Personal/pdf-cli/internal/commands/additional_coverage_test.go`

4. **Extended tests**: `*_extended_test.go`
   - Example: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/engine_extended_test.go`

5. **Helper tests**: `helpers_test.go` for test utilities
   - Example: `/Users/lgbarn/Personal/pdf-cli/internal/commands/helpers_test.go`

## Test Patterns

### Table-Driven Tests

**Standard pattern used throughout:**

```go
func TestValidatePDFFile(t *testing.T) {
    tmpDir, err := os.MkdirTemp("", "test-*")
    if err != nil {
        t.Fatal(err)
    }
    defer os.RemoveAll(tmpDir)

    // Setup test files
    pdfFile := filepath.Join(tmpDir, "test.pdf")
    txtFile := filepath.Join(tmpDir, "test.txt")
    for _, f := range []string{pdfFile, txtFile} {
        if err := os.WriteFile(f, []byte("dummy"), 0644); err != nil {
            t.Fatal(err)
        }
    }

    tests := []struct {
        path    string
        wantErr bool
    }{
        {pdfFile, false},
        {txtFile, true},
        {"/nonexistent/file.pdf", true},
    }

    for _, tt := range tests {
        t.Run(filepath.Base(tt.path), func(t *testing.T) {
            if err := ValidatePDFFile(tt.path); (err != nil) != tt.wantErr {
                t.Errorf("ValidatePDFFile(%q) error = %v, wantErr %v",
                    tt.path, err, tt.wantErr)
            }
        })
    }
}
```

**Key characteristics:**
- Anonymous struct for test cases
- Fields: input values + expected outputs + `wantErr` bool
- `t.Run()` for subtests with descriptive names
- `defer` cleanup for temp files/directories

### Subtests

Consistently uses `t.Run()` for logical test grouping:

```go
func TestCopyFile(t *testing.T) {
    // Setup
    tmpDir, _ := os.MkdirTemp("", "test-*")
    defer os.RemoveAll(tmpDir)

    t.Run("success", func(t *testing.T) {
        // Test successful copy
    })

    t.Run("nested destination", func(t *testing.T) {
        // Test creating nested directories
    })

    t.Run("non-existent source", func(t *testing.T) {
        // Test error handling
    })

    t.Run("overwrite", func(t *testing.T) {
        // Test overwriting existing file
    })
}
```

### Test Fixtures & Helpers

**Centralized test utilities** in `internal/testing/`:

**1. Fixture helpers (`fixtures.go`):**

```go
// TestdataDir returns the path to the testdata directory
func TestdataDir() string {
    _, filename, _, ok := runtime.Caller(0)
    if !ok {
        panic("failed to get caller information")
    }
    projectRoot := filepath.Dir(filepath.Dir(filepath.Dir(filename)))
    return filepath.Join(projectRoot, "testdata")
}

// SamplePDF returns the path to sample.pdf in testdata
func SamplePDF() string {
    return filepath.Join(TestdataDir(), "sample.pdf")
}

// TempDir creates a temporary directory for test artifacts
func TempDir(prefix string) (string, func()) {
    dir, err := os.MkdirTemp("", "pdf-cli-test-"+prefix+"-")
    if err != nil {
        panic("failed to create temp dir: " + err.Error())
    }
    return dir, func() { _ = os.RemoveAll(dir) }
}

// TempFile creates a temporary file with the given content
func TempFile(prefix, content string) (string, func()) {
    f, err := os.CreateTemp("", "pdf-cli-test-"+prefix+"-*.pdf")
    if err != nil {
        panic("failed to create temp file: " + err.Error())
    }
    if content != "" {
        if _, err := f.WriteString(content); err != nil {
            _ = f.Close()
            _ = os.Remove(f.Name())
            panic("failed to write temp file: " + err.Error())
        }
    }
    _ = f.Close()
    return f.Name(), func() { _ = os.Remove(f.Name()) }
}
```

**2. Mock OCR backend (`mock_ocr.go`):**

```go
type MockOCRBackend struct {
    NameResult         string
    AvailableResult    bool
    ProcessImageResult string
    ProcessImageError  error
    Calls              []string
}

func NewMockOCRBackend() *MockOCRBackend {
    return &MockOCRBackend{
        NameResult:         "mock",
        AvailableResult:    true,
        ProcessImageResult: "Mock OCR extracted text",
        Calls:              make([]string, 0),
    }
}

func (m *MockOCRBackend) ProcessImage(_ context.Context, imagePath, _ string) (string, error) {
    m.Calls = append(m.Calls, "ProcessImage:"+imagePath)
    if m.ProcessImageError != nil {
        return "", m.ProcessImageError
    }
    return m.ProcessImageResult, nil
}
```

**3. Mock PDF operations (`mock_pdf.go`):**

Custom mock implementations for PDF testing without real files.

### Temporary File Management

**Pattern 1: Inline cleanup with defer**

```go
func TestAtomicWrite(t *testing.T) {
    tmpDir, err := os.MkdirTemp("", "test-*")
    if err != nil {
        t.Fatal(err)
    }
    defer os.RemoveAll(tmpDir)

    testPath := filepath.Join(tmpDir, "test.txt")
    testData := []byte("Hello, World!")

    if err := AtomicWrite(testPath, testData); err != nil {
        t.Errorf("error = %v", err)
    }
}
```

**Pattern 2: Helper-based cleanup**

```go
func TestWithHelpers(t *testing.T) {
    dir, cleanup := testing.TempDir("mytest")
    defer cleanup()

    // Test logic
}
```

### Command Testing

**Integration test pattern for CLI commands:**

```go
func resetFlags(t *testing.T) {
    t.Helper()
    rootCmd := cli.GetRootCmd()
    _ = rootCmd.PersistentFlags().Set("verbose", "false")
    _ = rootCmd.PersistentFlags().Set("force", "false")
    // ... reset all flags
}

func executeCommand(args ...string) error {
    rootCmd := cli.GetRootCmd()
    rootCmd.SetArgs(args)
    rootCmd.SetOut(&bytes.Buffer{})
    rootCmd.SetErr(&bytes.Buffer{})
    return rootCmd.Execute()
}

func TestCompressCommand_WithOutput(t *testing.T) {
    resetFlags(t)
    if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
        t.Skip("sample.pdf not found in testdata")
    }

    tmpDir, err := os.MkdirTemp("", "pdf-test-*")
    if err != nil {
        t.Fatalf("Failed to create temp dir: %v", err)
    }
    defer os.RemoveAll(tmpDir)

    output := filepath.Join(tmpDir, "compressed.pdf")
    err = executeCommand("compress", samplePDF(), "-o", output)
    if err != nil {
        t.Errorf("compress command failed: %v", err)
    }
}
```

**Key practices:**
- Reset flags between tests (Cobra persists state)
- Capture stdout/stderr to avoid test pollution
- Use `t.Skip()` for missing fixtures
- Helper function `t.Helper()` for test utilities

### Error Path Testing

Tests cover both success and failure cases:

```go
func TestCopyFile(t *testing.T) {
    t.Run("success", func(t *testing.T) {
        // Happy path
    })

    t.Run("non-existent source", func(t *testing.T) {
        if err := CopyFile("/nonexistent/file.txt", dst); err == nil {
            t.Error("should return error for non-existent source")
        }
    })
}
```

### Test Helpers

**t.Helper() usage:**

```go
func resetFlags(t *testing.T) {
    t.Helper()  // Marks this as helper, cleaner stack traces
    rootCmd := cli.GetRootCmd()
    // ... reset logic
}
```

### Parallel Tests

Large test suites can benefit from parallelization:

```go
func TestValidatePDFFiles(t *testing.T) {
    // Tests parallel validation with 10 files
    pdfs := make([]string, 10)
    for i := 0; i < 10; i++ {
        pdfs[i] = filepath.Join(tmpDir, "file"+string(rune('a'+i))+".pdf")
        os.WriteFile(pdfs[i], []byte("dummy"), 0644)
    }

    if err := ValidatePDFFiles(pdfs); err != nil {
        t.Errorf("error = %v", err)
    }
}
```

Tests validate that functions handle both sequential (≤3 files) and parallel (>3 files) processing correctly.

## Test Data

### Testdata Directory

Project maintains a `testdata/` directory:
```
testdata/
├── sample.pdf
├── test_image.png
└── (other test fixtures)
```

**Accessed via helpers:**
```go
testing.TestdataDir()    // Returns absolute path to testdata/
testing.SamplePDF()      // Returns path to sample.pdf
testing.TestImage()      // Returns path to test_image.png
```

**Exclusions in CI:**
- `.pre-commit-config.yaml`: excludes `testdata/` from large file checks

### Security Annotations in Tests

Test files use `#nosec` for intentional patterns:

```go
data, err := os.ReadFile(src) // #nosec G304 - test fixture, paths are controlled
os.WriteFile(dst, data, 0644) // #nosec G306 - test fixture, permissive permissions OK
```

## CI/CD Testing

### GitHub Actions Workflow

**Lint job:**
```yaml
- name: Run golangci-lint
  uses: golangci/golangci-lint-action@v8
  with:
    version: v2.8.0
    args: --timeout=5m
```

**Test job:**
```yaml
- name: Run tests with coverage
  run: go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

- name: Check coverage threshold (75%)
  run: |
    COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | tr -d '%')
    if awk "BEGIN {exit !($COVERAGE < 75)}"; then
      echo "::error::Coverage ${COVERAGE}% is below 75% threshold"
      exit 1
    fi

- name: Upload coverage to Codecov
  uses: codecov/codecov-action@v4
  with:
    files: ./coverage.out
    fail_ci_if_error: false
```

**Security scan job:**
```yaml
- name: Install Gosec
  run: go install github.com/securego/gosec/v2/cmd/gosec@v2.21.4

- name: Run Gosec
  run: gosec -exclude-dir=testdata ./...
```

### Pre-commit Hooks

Automated testing on every commit:

```yaml
- id: go-test
  name: go test
  entry: bash -c 'GO111MODULE=on go test ./...'
  language: system
  types: [go]
  pass_filenames: false
```

## Test Execution Performance

Tests complete in under 15 seconds total:

```
internal/cli             0.234s
internal/commands        1.170s
internal/commands/patterns 0.250s
internal/config          0.990s
internal/fileio          0.507s
internal/logging         1.247s
internal/ocr             2.915s  (slowest - OCR operations)
internal/output          1.033s
internal/pages           1.236s
internal/pdf             1.881s
internal/pdferrors       1.957s
internal/progress        2.166s
```

**Optimization:**
- Race detection enabled without significant slowdown
- Parallel processing tested but not overused
- Fast unit tests, heavier integration tests isolated

## Writing New Tests

### Checklist

1. **Co-locate**: Place `*_test.go` alongside source file
2. **Table-driven**: Use table-driven tests for multiple cases
3. **Subtests**: Use `t.Run()` for logical grouping
4. **Cleanup**: Always defer cleanup of temp resources
5. **Both paths**: Test success and error cases
6. **Helpers**: Extract common setup to helper functions with `t.Helper()`
7. **Fixtures**: Use `internal/testing/` helpers for test data
8. **Mocks**: Use provided mocks for external dependencies
9. **Coverage**: Aim for >75% coverage on new code
10. **Race detection**: Ensure `go test -race` passes

### Example Template

```go
package mypackage

import (
    "os"
    "testing"
)

func TestMyFunction(t *testing.T) {
    // Setup
    tmpDir, err := os.MkdirTemp("", "test-*")
    if err != nil {
        t.Fatal(err)
    }
    defer os.RemoveAll(tmpDir)

    // Table-driven test cases
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {"valid case", "input1", "output1", false},
        {"error case", "bad", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := MyFunction(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("got %v, want %v", got, tt.want)
            }
        })
    }
}
```

## Documentation Testing

Currently no doc examples (`Example*` functions), but package docs are maintained in `doc.go` files.

## Benchmark Tests

No benchmark tests (`Benchmark*` functions) currently present, but can be added for performance-critical paths.
