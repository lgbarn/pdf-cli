# Testing Infrastructure

## Overview
pdf-cli has comprehensive test coverage with 45 test files, achieving 75%+ coverage threshold enforced in CI. Tests use Go's standard testing framework with table-driven patterns, mock implementations, and both unit and integration test suites. Coverage is tracked with go test -coverprofile and verified in every CI run.

## Findings

### Test Framework and Tools

- **Test Framework**: Go standard `testing` package
  - Evidence: All test files import `"testing"` -- e.g., `/Users/lgbarn/Personal/pdf-cli/internal/pages/pages_test.go` (line 6)
  - No third-party test frameworks (e.g., testify, ginkgo) used

- **Test Runner**: `go test` with various flags
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/Makefile` (lines 56-71)
  - Basic: `go test -v ./...`
  - Coverage: `go test -v -coverprofile=coverage.out ./...`
  - Race detection: `go test -race -v ./...`

- **Coverage Tools**: Built-in `go tool cover`
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/Makefile` (lines 62-66, 105-121)
  - Generates HTML reports: `go tool cover -html=coverage.out -o coverage.html`
  - Function-level coverage: `go tool cover -func=coverage.out`

- **Coverage Threshold**: 75% minimum enforced
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/Makefile` (lines 112-121)
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/.github/workflows/ci.yaml` (lines 51-54)
  - Custom script: `/Users/lgbarn/Personal/pdf-cli/scripts/coverage-check.go` parses coverage output and validates threshold

### Test File Organization

- **Test File Naming**: `*_test.go` suffix alongside source files
  - Evidence: 45 test files found in `/Users/lgbarn/Personal/pdf-cli/internal/`
  - Pattern: Tests live in same package as source (not separate `_test` package)
  - Examples:
    - `/Users/lgbarn/Personal/pdf-cli/internal/pages/pages_test.go` tests `parser.go`, `validator.go`, `reorder.go`
    - `/Users/lgbarn/Personal/pdf-cli/internal/pdferrors/errors_test.go` tests `errors.go`
    - `/Users/lgbarn/Personal/pdf-cli/internal/cli/flags_test.go` tests `flags.go`

- **Integration Tests**: Separate files with `integration` in name
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/commands/commands_integration_test.go`
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/commands/integration_batch_test.go`
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/commands/integration_content_test.go`
  - Pattern: Integration tests use real PDF files from testdata

- **Coverage-focused Tests**: Explicit coverage test files
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/commands/coverage_batch_test.go`
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/commands/coverage_images_test.go`
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/commands/additional_coverage_test.go`
  - Pattern: Separate files targeting specific coverage gaps

- **Mock Files**: Dedicated mock implementations
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/testing/mock_pdf.go`
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/testing/mock_ocr.go`
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/mock_test.go` (test-specific mock)
  - Pattern: Reusable mocks in `/internal/testing/`, test-specific mocks colocated

### Test Fixtures and Data

- **Testdata Directory**: Centralized test fixtures
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/testdata/` contains `sample.pdf` and `test_image.png`
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/testing/fixtures.go` (lines 9-29)
  - Helper functions: `TestdataDir()`, `SamplePDF()`, `TestImage()`

- **Fixture Access Pattern**: Runtime discovery of testdata location
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/testing/fixtures.go` (lines 9-19)
  ```go
  func TestdataDir() string {
      _, filename, _, ok := runtime.Caller(0)
      if !ok {
          panic("failed to get caller information")
      }
      // Go up from internal/testing to project root
      projectRoot := filepath.Dir(filepath.Dir(filepath.Dir(filename)))
      return filepath.Join(projectRoot, "testdata")
  }
  ```

- **Temporary File Helpers**: Utilities for test isolation
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/testing/fixtures.go` (lines 31-67)
  - `TempDir(prefix string)` -- creates temp directory with cleanup function
  - `TempFile(prefix, content string)` -- creates temp file with cleanup
  - `CopyFile(src, dst string)` -- copies fixtures for mutation tests

- **Test Skipping**: Conditional test execution
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/commands/commands_integration_test.go` (lines 11-13)
  ```go
  if _, err := os.Stat(samplePDF()); os.IsNotExist(err) {
      t.Skip("sample.pdf not found in testdata")
  }
  ```
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr_test.go` (lines 139-141)
  ```go
  if testing.Short() {
      t.Skip("Skipping in short mode")
  }
  ```

### Test Patterns

- **Table-Driven Tests**: Primary test pattern throughout codebase
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/pages/pages_test.go` (lines 9-46)
  ```go
  func TestParsePageRanges(t *testing.T) {
      tests := []struct {
          input   string
          want    []PageRange
          wantErr bool
      }{
          {"", nil, false},
          {"1", []PageRange{{1, 1}}, false},
          {"1,3,5", []PageRange{{1, 1}, {3, 3}, {5, 5}}, false},
          // ... more cases
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
  ```
  - Pattern: Struct with `name`, inputs, `want`, and `wantErr` fields

- **Subtests**: `t.Run()` for each test case
  - Evidence: All table-driven tests use `t.Run(tt.name, func(t *testing.T) { ... })`
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/pages/pages_test.go` (line 36)
  - Enables running individual cases: `go test -run TestParsePageRanges/single`

- **Test Naming Convention**: Descriptive test case names
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/pages/pages_test.go` (lines 117-147)
  - Examples: `"single page"`, `"multiple pages"`, `"simple range"`, `"reverse range"`, `"end keyword"`, `"move page 3 to front"`
  - Pattern: Human-readable strings describing the scenario being tested

- **Error Message Assertions**: String matching for error validation
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/pages/pages_test.go` (lines 124, 153-158)
  ```go
  {
      name:        "empty spec",
      spec:        "",
      want:        nil,
      wantErr:     true,
      errContains: "empty",
  },
  ```
  - Pattern: `errContains` field used with `strings.Contains(err.Error(), tt.errContains)`

### Mock Implementation Patterns

- **Mock Struct**: Configurable mock with call tracking
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/testing/mock_pdf.go` (lines 7-24)
  ```go
  type MockPDFOps struct {
      PageCountResult int
      PageCountError error
      CompressError error
      MergeError error
      SplitError error
      ExtractTextResult string
      ExtractTextError error
      Calls []string
  }
  ```

- **Call Tracking**: Records method invocations
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/testing/mock_pdf.go` (lines 36-43)
  ```go
  func (m *MockPDFOps) PageCount(file, _ string) (int, error) {
      m.Calls = append(m.Calls, "PageCount:"+file)
      if m.PageCountError != nil {
          return 0, m.PageCountError
      }
      return m.PageCountResult, nil
  }
  ```

- **Assertion Helpers**: Methods to verify mock calls
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/testing/mock_pdf.go` (lines 76-84)
  ```go
  func (m *MockPDFOps) AssertCalled(prefix string) bool {
      for _, call := range m.Calls {
          if len(call) >= len(prefix) && call[:len(prefix)] == prefix {
              return true
          }
      }
      return false
  }
  ```

- **Builder Pattern for Mocks**: Fluent mock configuration
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/mock_test.go` (lines 213-218)
  ```go
  backend := newMockBackend("mock", true).
      withOutput("test text").
      withErrorIndices(map[string]error{
          "img1.png": context.DeadlineExceeded,
          "img3.png": context.DeadlineExceeded,
      })
  ```

### Test Utilities and Helpers

- **Test Helper Functions**: Shared setup/teardown
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/testing/fixtures.go`
  - Pattern: Return cleanup functions for defer
  ```go
  dir, cleanup := TempDir("test")
  defer cleanup()
  ```

- **Test Reset Functions**: Clean up global state
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/logging/logger.go` (lines 143-148)
  ```go
  // Reset resets the global logger (for testing).
  func Reset() {
      globalMu.Lock()
      defer globalMu.Unlock()
      global = nil
  }
  ```

- **Test Server Mocking**: httptest for HTTP clients
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr_test.go` (lines 288-331)
  ```go
  server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
      n := requestCount.Add(1)
      if n == 1 {
          w.WriteHeader(http.StatusServiceUnavailable)
          return
      }
      w.WriteHeader(http.StatusOK)
      _, _ = w.Write(body)
  }))
  defer server.Close()
  ```

### Integration Test Patterns

- **Real File Operations**: Uses actual PDF files
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/commands/commands_integration_test.go` (lines 9-29)
  ```go
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
      if err := executeCommand("compress", samplePDF(), "-o", output); err != nil {
          t.Fatalf("compress command failed: %v", err)
      }

      if _, err := os.Stat(output); os.IsNotExist(err) {
          t.Error("compress did not create output file")
      }
  }
  ```

- **Command Execution**: Tests CLI commands end-to-end
  - Evidence: Integration tests call `executeCommand()` helper
  - Pattern: Sets up real files, executes command, verifies output files exist

- **Temporary Directories**: Isolated test environments
  - Evidence: All integration tests use `os.MkdirTemp` with `defer os.RemoveAll`
  - Pattern: `"pdf-test-*"` prefix for temp directories

### Coverage Strategy

- **Coverage Measurement**: Per-package and total coverage
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/Makefile` (lines 107-121)
  ```makefile
  coverage:
      @GO111MODULE=on $(GOTEST) -coverprofile=coverage.out ./... 2>/dev/null
      @go tool cover -func=coverage.out | grep total | awk '{print "Total coverage: " $$3}'

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

- **CI Coverage Enforcement**: Runs on every pull request
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/.github/workflows/ci.yaml` (lines 48-54)
  ```yaml
  - name: Run tests with coverage
    run: go test -v -race -coverprofile=coverage.out -covermode=atomic ./...

  - name: Check coverage threshold
    run: |
      go tool cover -func=coverage.out > coverage-summary.txt
      go run scripts/coverage-check.go coverage-summary.txt 75
  ```

- **Coverage Artifacts**: Multiple coverage output files
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/` root contains `cover.out`, `cover_cmd.out`, `cover_final.out`, `cover_verify.out`, `coverage.out`, `ocr_cover.out`
  - [Inferred] Different coverage files likely represent coverage from different test runs or phases

### Test Execution Modes

- **Race Detection**: Concurrent safety testing
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/Makefile` (lines 68-71)
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/.github/workflows/ci.yaml` (line 49)
  ```makefile
  test-race:
      GO111MODULE=on $(GOTEST) -race -v ./...
  ```

- **Short Mode**: Skip long-running tests
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr_test.go` (lines 139-141)
  - Use: `go test -short` to skip integration/network tests

- **Verbose Mode**: Detailed test output
  - Evidence: All Makefile test targets use `-v` flag
  - Shows individual test results and timing

### Test Documentation

- **Test Examples in Contributing Guide**: Documentation includes test patterns
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/CONTRIBUTING.md` (lines 63-106)
  - Shows table-driven test structure
  - Emphasizes testing both success and error paths

- **Test Location Guidance**: Tests colocated with source
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/CONTRIBUTING.md` (line 76)
  - "Place tests in `*_test.go` files alongside source"

### Parallel Testing

- **Parallel Test Execution**: Selected tests run in parallel
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr_test.go` (lines 288, 333, 361)
  ```go
  func TestDownloadTessdataRetryOnServerError(t *testing.T) {
      t.Parallel()
      // ...
  }
  ```
  - Pattern: Use `t.Parallel()` for tests that don't share global state

- **Concurrency Testing**: Atomic operations for thread safety
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr_test.go` (lines 291-304)
  ```go
  var requestCount atomic.Int32
  server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
      n := requestCount.Add(1)
      // ...
  }))
  ```

### Error Path Testing

- **Error Assertions**: Both presence and content validated
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/pdferrors/errors_test.go` (lines 119-140)
  ```go
  if tt.wantNil {
      if result != nil {
          t.Errorf("WrapError() = %v, want nil", result)
      }
      return
  }

  var pdfErr *PDFError
  if !errors.As(result, &pdfErr) {
      t.Errorf("WrapError() did not return PDFError")
      return
  }

  if tt.wantCause != nil && !errors.Is(pdfErr.Cause, tt.wantCause) {
      t.Errorf("WrapError() cause = %v, want %v", pdfErr.Cause, tt.wantCause)
  }
  ```

- **Error Unwrapping Tests**: Validates error chain
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/pdferrors/errors_test.go` (lines 246-261)
  ```go
  func TestPDFError_Unwrap(t *testing.T) {
      cause := errors.New("original error")
      err := &PDFError{
          Operation: "test",
          Cause:     cause,
      }

      if unwrapped := err.Unwrap(); unwrapped != cause {
          t.Errorf("Unwrap() = %v, want %v", unwrapped, cause)
      }

      // Test with errors.Unwrap
      if unwrapped := errors.Unwrap(err); unwrapped != cause {
          t.Errorf("errors.Unwrap() = %v, want %v", unwrapped, cause)
      }
  }
  ```

## Summary Table

| Component | Detail | Confidence |
|-----------|--------|------------|
| Test framework | Go standard testing package | Observed |
| Test count | 45 test files in internal/ | Observed |
| Coverage tool | go test -coverprofile + go tool cover | Observed |
| Coverage threshold | 75% enforced in CI and Makefile | Observed |
| Test pattern | Table-driven tests with subtests | Observed |
| Test location | Colocated with source in same package | Observed |
| Integration tests | Separate `*_integration_test.go` files | Observed |
| Mock pattern | Struct-based mocks with call tracking | Observed |
| Fixtures | Centralized testdata/ with helper functions | Observed |
| Temp files | TempDir/TempFile with cleanup functions | Observed |
| Race detection | `go test -race` in Makefile and CI | Observed |
| Short mode | `testing.Short()` for skipping slow tests | Observed |
| Parallel tests | `t.Parallel()` for independent tests | Observed |
| CI integration | Coverage check + race detector on every PR | Observed |
| Coverage artifacts | 6+ coverage files in project root | Observed |

## Open Questions

- **Coverage file proliferation**: Multiple `cover*.out` files in root suggest manual test runs or different coverage phases. Should these be in .gitignore or consolidated?
- **Mock reusability**: Some packages have test-specific mocks (`ocr/mock_test.go`) while others use shared mocks (`internal/testing/mock_*.go`). Is there a guideline for when to share vs. colocate mocks?
