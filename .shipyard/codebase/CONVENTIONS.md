# Code Conventions

## Overview
pdf-cli is a Go 1.25 CLI project following standard Go conventions with explicit linting rules enforced via golangci-lint, pre-commit hooks, and GitHub Actions CI. The codebase emphasizes clean structure, comprehensive error handling, and user-friendly messaging patterns.

## Findings

### Language and Build

- **Language**: Go 1.25
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/go.mod` (line 3)
  - Go modules enabled with `GO111MODULE=on` in all Makefile targets

- **Code Formatting**: Standard `go fmt` and `goimports`
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/.golangci.yaml` (lines 92-94)
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/.pre-commit-config.yaml` (lines 17-22) -- `go fmt` runs on every commit
  - All `.go` files formatted with tabs for indentation, standard Go style

### Linting Configuration

- **Primary Linter**: golangci-lint v2.8.0 with 5-minute timeout
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/.golangci.yaml` (lines 3-4)
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/.github/workflows/ci.yaml` (lines 26-30)

- **Enabled Linters**: 8 linters active
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/.golangci.yaml` (lines 6-18)
  - `govet`, `ineffassign`, `staticcheck`, `unused` (original set)
  - `misspell` (US locale), `gocritic`, `revive`, `errcheck` (added later)

- **gocritic Settings**: Diagnostic and style tags enabled with 8 checks disabled
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/.golangci.yaml` (lines 21-33)
  - Disabled: `ifElseChain`, `whyNoLint`, `octalLiteral`, `importShadow`, `deferInLoop`, `httpNoBody`, `unnamedResult`, `paramTypeCombine`

- **revive Settings**: 13 rules enabled, package-comments disabled
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/.golangci.yaml` (lines 35-54)
  - Key rules: `error-strings`, `error-naming`, `exported`, `var-naming`, `receiver-naming`, `indent-error-flow`, `errorf`

- **Linter Exclusions**: Strategic exclusions for common patterns
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/.golangci.yaml` (lines 59-89)
  - errcheck: Excludes `Close`, `Remove`, `RemoveAll`, `fmt.Fprint*` errors
  - Test files: Exempt from `errcheck` and `unused-parameter` checks
  - Exported comments: Excluded from `revive` (too noisy for CLI tools)

### Naming Conventions

- **Package Names**: Single-word lowercase, descriptive
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/pages/parser.go` (line 1) -- `package pages`
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/pdferrors/errors.go` (line 1) -- `package pdferrors`
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/progress/progress.go` (line 1) -- `package progress`
  - Pattern: Package names match directory names exactly

- **File Names**: Snake_case with descriptive suffixes
  - Evidence: `commands_test.go`, `commands_integration_test.go`, `stdio_test.go`, `mock_pdf.go`, `mock_ocr.go`
  - Test files: `*_test.go`
  - Integration tests: `*_integration_test.go`
  - Mock implementations: `mock_*.go`
  - Coverage tests: `coverage_*_test.go`, `additional_coverage_test.go`

- **Type Names**: PascalCase, descriptive
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/pages/parser.go` (line 11) -- `type PageRange struct`
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/pdferrors/errors.go` (line 10) -- `type PDFError struct`
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/output/formatter.go` (line 12) -- `type OutputFormat string`
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/logging/logger.go` (line 12) -- `type Level string`

- **Function Names**: MixedCaps, descriptive verbs
  - Exported: PascalCase -- `ParsePageRanges`, `ExpandPageRanges`, `ValidatePageNumbers`, `FormatPageRanges`
  - Unexported: camelCase -- `parseRangePart`, `parseSinglePage`, `formatRange`, `checkOutputFile`
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/pages/parser.go` (lines 16, 50, 78, 92, 118)

- **Variable Names**: camelCase, concise but clear
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/pages/parser.go` (lines 23, 30, 51-52, 57, 62) -- `ranges`, `parts`, `part`, `bounds`, `start`, `end`
  - Short names for common patterns: `err`, `ctx`, `cmd`, `cfg`, `f` (file)
  - Longer names for less obvious items: `pageNums`, `inputFile`, `totalPages`, `password`

- **Constant Names**: PascalCase for exported, camelCase for unexported
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/logging/logger.go` (lines 14-19) -- `LevelDebug`, `LevelInfo`, `LevelWarn`, `LevelError`, `LevelSilent`
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/output/formatter.go` (lines 15-18) -- `FormatHuman`, `FormatJSON`, `FormatCSV`, `FormatTSV`

- **Error Variables**: `Err` prefix for sentinel errors
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/pdferrors/errors.go` (lines 51-59)
  - Examples: `ErrFileNotFound`, `ErrNotPDF`, `ErrInvalidPages`, `ErrPasswordRequired`, `ErrWrongPassword`, `ErrCorruptPDF`, `ErrOutputExists`
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/testing/mock_pdf.go` (lines 86-90) -- `ErrMockPasswordRequired`, `ErrMockCorrupted`

### Import Organization

- **Import Grouping**: Standard library first, blank line, then external packages
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/pages/parser.go` (lines 3-8)
  ```go
  import (
      "fmt"
      "sort"
      "strconv"
      "strings"
  )
  ```
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/commands/helpers.go` (lines 3-12)
  ```go
  import (
      "errors"
      "fmt"

      "github.com/lgbarn/pdf-cli/internal/cli"
      "github.com/lgbarn/pdf-cli/internal/fileio"
      "github.com/lgbarn/pdf-cli/internal/pages"
      "github.com/lgbarn/pdf-cli/internal/pdf"
      "github.com/lgbarn/pdf-cli/internal/pdferrors"
  )
  ```

- **Blank Imports**: Used for side-effect initialization
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/cmd/pdf/main.go` (line 11)
  ```go
  _ "github.com/lgbarn/pdf-cli/internal/commands" // Register all commands
  ```

### Comment Style

- **Package Comments**: Single-line descriptive comments in `doc.go` files
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/pages/doc.go` (lines 1-2)
  ```go
  // Package pages provides page range parsing and validation for pdf-cli.
  package pages
  ```
  - Pattern: Every internal package has a `doc.go` file with package-level documentation

- **Function Comments**: Complete sentences starting with function name
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/pages/parser.go` (lines 16-17, 50, 78, 92, 118)
  ```go
  // ParsePageRanges parses a page range string like "1-5,7,10-12".
  // Returns a slice of PageRange structs.
  ```
  - Pattern: Multi-sentence comments use separate lines for each sentence

- **Struct Field Comments**: Inline comments when brief, above when detailed
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/pdferrors/errors.go` (lines 10-15)
  ```go
  type PDFError struct {
      Operation string
      File      string
      Cause     error
      Hint      string
  }
  ```
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/testing/mock_pdf.go` (lines 8-23)
  ```go
  // PageCountResult is returned by PageCount
  PageCountResult int
  // PageCountError is returned by PageCount if set
  PageCountError error
  ```

- **Inline Comments**: Used sparingly for non-obvious logic
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/pages/parser.go` (line 141) -- `// Skip duplicates (sorted[i] == end)`
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/commands/helpers.go` (lines 14-15, 23-24)

- **Security Comments**: `#nosec` annotations with justifications
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/testing/fixtures.go` (lines 61, 65)
  ```go
  data, err := os.ReadFile(src) // #nosec G304 - test fixture, paths are controlled
  return os.WriteFile(dst, data, 0644) // #nosec G306 - test fixture, permissive permissions OK
  ```

### Error Handling

- **Error Wrapping**: Extensive use of `fmt.Errorf` with `%w` verb
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/commands/helpers.go` (lines 32, 37, 68, 92)
  ```go
  return nil, fmt.Errorf("invalid page specification: %w", err)
  return nil, fmt.Errorf("invalid file path: %w", err)
  ```

- **Custom Error Types**: Structured errors with context
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/pdferrors/errors.go` (lines 9-33)
  - `PDFError` implements `Error()` and `Unwrap()` methods
  - Builder pattern with `WithHint()` method

- **Error Checking**: Checked immediately after call
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/commands/merge.go` (lines 38-41, 49-52, 54-56)
  ```go
  args, err := sanitizeInputArgs(args)
  if err != nil {
      return err
  }
  ```

- **Error Joining**: Multiple errors combined with `errors.Join`
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/commands/helpers.go` (lines 88-96)
  ```go
  func processBatch(files []string, processor func(file string) error) error {
      var errs []error
      for _, file := range files {
          if err := processor(file); err != nil {
              errs = append(errs, fmt.Errorf("%s: %w", file, err))
          }
      }
      return errors.Join(errs...)
  }
  ```

### Code Organization Patterns

- **init() Functions**: Used for command registration
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/commands/merge.go` (lines 13-19)
  ```go
  func init() {
      cli.AddCommand(mergeCmd)
      cli.AddOutputFlag(mergeCmd, "Output file path (required)")
      cli.AddPasswordFlag(mergeCmd, "Password for encrypted input PDFs")
      cli.AddPasswordFileFlag(mergeCmd, "")
      _ = mergeCmd.MarkFlagRequired("output")
  }
  ```
  - Pattern: Commands self-register via side-effect imports

- **Factory Functions**: Constructor functions with `New` prefix
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/pdferrors/errors.go` (line 36) -- `func NewPDFError(...)`
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/output/formatter.go` (line 42) -- `func NewOutputFormatter(...)`
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/testing/mock_pdf.go` (line 27) -- `func NewMockPDFOps()`

- **Helper Functions**: Unexported helpers grouped logically
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/commands/helpers.go`
  - Pattern: Small, focused functions with clear single responsibility
  - Examples: `checkOutputFile`, `parseAndValidatePages`, `outputOrDefault`, `validateBatchOutput`, `sanitizeInputArgs`, `processBatch`

- **Singleton Pattern**: Global logger with lazy initialization
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/logging/logger.go` (lines 84-92, 123-141)
  - Uses `sync.RWMutex` for thread-safe access

### String Formatting

- **Format Verbs**: Consistent use of standard format verbs
  - `%s` for strings, `%d` for integers, `%v` for default format, `%w` for errors, `%q` for quoted strings
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/pdferrors/errors.go` (lines 21, 26)
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/commands/helpers.go` (line 18)

- **String Building**: `strings.Builder` for efficient concatenation
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/pdferrors/errors.go` (lines 18-28)
  ```go
  var sb strings.Builder
  sb.WriteString(e.Operation)
  if e.File != "" {
      sb.WriteString(fmt.Sprintf(" '%s'", e.File))
  }
  ```

### Build Tags

- **Ignore Tag**: Used for helper scripts
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/scripts/coverage-check.go` (line 1)
  ```go
  //go:build ignore
  ```

### Pre-commit and CI Integration

- **Pre-commit Hooks**: 11 hooks enforcing code quality
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/.pre-commit-config.yaml`
  - General: `trailing-whitespace`, `end-of-file-fixer`, `check-yaml`, `check-added-large-files`, `check-merge-conflict`
  - Go-specific: `go-fmt`, `go-vet`, `go-mod-tidy`, `go-build`, `go-test`, `golangci-lint`

- **CI Pipeline**: 4 jobs (lint, test, build, security)
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/.github/workflows/ci.yaml`
  - Runs golangci-lint with same config as local
  - Security scanning with Gosec v2.22.11
  - Multi-platform builds (Linux, Darwin, Windows) × (amd64, arm64)

### Conventional Commits

- **Commit Prefixes**: Structured commit messages required
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/CONTRIBUTING.md` (lines 51-61)
  - Prefixes: `feat:`, `fix:`, `docs:`, `test:`, `chore:`, `refactor:`
  - Example: `feat: add --dry-run flag to merge command`

## Summary Table

| Convention | Detail | Confidence |
|------------|--------|------------|
| Language | Go 1.25 | Observed |
| Formatting | go fmt + goimports | Observed |
| Linter | golangci-lint v2.8.0, 8 linters | Observed |
| Package names | Single-word lowercase | Observed |
| File names | Snake_case, `*_test.go`, `mock_*.go` | Observed |
| Type names | PascalCase | Observed |
| Function names | PascalCase (exported), camelCase (unexported) | Observed |
| Variable names | camelCase, descriptive | Observed |
| Error vars | `Err` prefix for sentinels | Observed |
| Import order | stdlib first, then external with blank line | Observed |
| Error wrapping | `fmt.Errorf` with `%w`, custom `PDFError` type | Observed |
| Comments | Complete sentences, package docs in `doc.go` | Observed |
| Pre-commit | 11 hooks including fmt, vet, lint, test | Observed |
| CI | golangci-lint, tests with race detector, gosec | Observed |
| Commit style | Conventional commits (feat:, fix:, etc.) | Observed |

## Open Questions

None. The project has comprehensive linting configuration, clear conventions documented in CONTRIBUTING.md, and enforcement via pre-commit hooks and CI. All conventions are consistently applied throughout the codebase.
