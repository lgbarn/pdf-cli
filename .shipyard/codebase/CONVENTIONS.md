# Code Conventions

This document describes the coding standards, naming patterns, and style conventions used in pdf-cli.

## Language & Tooling

- **Language**: Go 1.24.1
- **Linter**: golangci-lint v2.8.0 with custom configuration
- **Formatter**: gofmt + goimports (auto-enabled via golangci-lint)
- **Pre-commit**: Automated checks via `.pre-commit-config.yaml`
- **Build**: Make-based workflow

## Linting Configuration

### Enabled Linters

The project uses golangci-lint v2 with the following linters (`.golangci.yaml`):

**Core linters:**
- `govet` - Standard Go vet checks
- `ineffassign` - Detects ineffectual assignments
- `staticcheck` - Advanced static analysis
- `unused` - Detects unused constants, variables, functions

**Additional linters:**
- `misspell` - Catches spelling errors (US locale)
- `gocritic` - Comprehensive style and diagnostics
- `revive` - Fast, configurable, extensible Go linter
- `errcheck` - Ensures error return values are checked

### Linter Settings

**gocritic configuration:**
- Enabled tags: `diagnostic`, `style`
- Disabled checks: `ifElseChain`, `whyNoLint`, `octalLiteral`, `importShadow`, `deferInLoop`, `httpNoBody`, `unnamedResult`, `paramTypeCombine`

**revive configuration:**
- Enforces: blank-imports, context-as-argument, error-return, error-strings, error-naming, exported, var-naming, receiver-naming, time-naming, indent-error-flow, errorf
- Disabled: `package-comments` (too noisy for CLI tools)

**Common exclusions:**
- `errcheck` ignored for: `Close()`, `Remove()`, `RemoveAll()`, `fmt.Fprint*()` functions
- `errcheck` ignored in test files
- `revive` exported comments rule disabled (too noisy for CLI)
- `revive` unused-parameter ignored in test files

### Formatter Configuration

- `gofmt` - Standard Go formatting
- `goimports` - Auto-manages imports

## Code Organization

### Package Structure

```
internal/
├── cli/              # CLI framework and root command
├── commands/         # One file per command
│   └── patterns/     # Reusable command patterns
├── config/           # Configuration management
├── fileio/           # File I/O and stdio utilities
├── logging/          # Structured logging
├── ocr/              # OCR engine with backends
├── output/           # Output formatting
├── pages/            # Page range parsing
├── pdf/              # PDF operations (pdfcpu wrapper)
├── pdferrors/        # Error handling
├── progress/         # Progress bars
└── testing/          # Test helpers and mocks
```

**Key principles:**
- No circular dependencies
- Leaf packages have minimal dependencies
- External dependencies isolated in dedicated packages
- One command per file in `commands/`
- All packages in `internal/` (not exported as library)

### Package Documentation

Every package includes a `doc.go` file with package-level documentation:

```go
// Package name provides description of what it does.
package name
```

**Examples:**
- `/Users/lgbarn/Personal/pdf-cli/internal/pdferrors/doc.go`
- `/Users/lgbarn/Personal/pdf-cli/internal/fileio/doc.go`
- `/Users/lgbarn/Personal/pdf-cli/internal/progress/doc.go`

## Naming Conventions

### Files

- **Source files**: lowercase with underscores: `pdf_operations.go`, `file_utils.go`
- **Test files**: `*_test.go` (placed alongside source)
- **Package docs**: `doc.go` (one per package)
- **Commands**: Named after command: `compress.go`, `encrypt.go`, `rotate.go`

### Functions & Methods

- **Exported functions**: PascalCase with descriptive names
  - `ValidatePDFFile()`, `GenerateOutputFilename()`, `FormatFileSize()`
- **Unexported functions**: camelCase
  - `outputOrDefault()`, `compressFile()`, `checkOutputFile()`
- **Test functions**: `TestFunctionName` or `TestFunctionName_Scenario`
  - `TestFileExists()`, `TestCompressCommand_WithOutput()`
- **Helper functions**: Named by purpose, often with `Helper` suffix
  - `resetFlags(t)`, `executeCommand()`, `samplePDF()`

### Variables & Constants

- **Exported constants**: PascalCase
  - `LevelDebug`, `FormatJSON`, `BackendNative`
- **Unexported variables**: camelCase
  - `global`, `version`, `commit`, `buildDate`
- **Error variables**: Prefixed with `Err`
  - `ErrFileNotFound`, `ErrNotPDF`, `ErrPasswordRequired`
- **Package-level variables**: Exported when needed for configuration
  - `SupportedImageExtensions` (slice of strings)

### Types

- **Structs**: PascalCase
  - `PDFError`, `Engine`, `Config`, `Logger`
- **Interfaces**: PascalCase, often with "-er" suffix
  - `Backend` (OCR backend interface)
- **Type aliases**: PascalCase
  - `Level` (string), `Format` (string)

### Receivers

- **Consistent naming**: 1-2 letter abbreviation of type name
  - `(e *PDFError)` for PDFError methods
  - `(l *Logger)` for Logger methods
  - `(m *MockOCRBackend)` for mock methods
- **Pointer vs value**: Use pointer receivers for structs that maintain state

## Code Style Patterns

### Error Handling

**Custom error type with context:**

```go
type PDFError struct {
    Operation string
    File      string
    Cause     error
    Hint      string
}

func (e *PDFError) Error() string {
    var sb strings.Builder
    sb.WriteString(e.Operation)
    if e.File != "" {
        sb.WriteString(fmt.Sprintf(" '%s'", e.File))
    }
    sb.WriteString(": ")
    sb.WriteString(e.Cause.Error())
    if e.Hint != "" {
        sb.WriteString(fmt.Sprintf("\nHint: %s", e.Hint))
    }
    return sb.String()
}

func (e *PDFError) Unwrap() error {
    return e.Cause
}
```

**Error wrapping pattern:**

```go
if err := pdf.Compress(input, output, password); err != nil {
    return pdferrors.WrapError("compressing file", inputArg, err)
}
```

**Predefined errors:**

```go
var (
    ErrFileNotFound     = errors.New("file not found")
    ErrNotPDF           = errors.New("not a valid PDF file")
    ErrPasswordRequired = errors.New("password required")
)
```

### File Operations

**Atomic writes with cleanup:**

```go
tmpFile, err := os.CreateTemp(dir, ".pdf-cli-tmp-*")
if err != nil {
    return fmt.Errorf("failed to create temp file: %w", err)
}

defer func() {
    if tmpFile != nil {
        _ = tmpFile.Close()
        _ = os.Remove(tmpPath)
    }
}()
```

**Path validation before use:**

```go
func CopyFile(src, dst string) error {
    // Clean paths to prevent directory traversal
    cleanSrc := filepath.Clean(src)
    cleanDst := filepath.Clean(dst)

    srcFile, err := os.Open(cleanSrc) // #nosec G304 -- path is cleaned
    // ...
}
```

### Security Annotations

Uses `#nosec` comments to suppress false positive security warnings from gosec:

```go
srcFile, err := os.Open(cleanSrc) // #nosec G304 -- path is cleaned
if err := os.WriteFile(dst, data, 0644) // #nosec G306 -- test fixture, permissive OK
```

**Pattern**: Always include justification after `#nosec` directive

### Configuration Management

**Global singleton with lazy initialization:**

```go
var global *Config

func Get() *Config {
    if global == nil {
        var err error
        global, err = Load()
        if err != nil {
            global = DefaultConfig()
        }
    }
    return global
}

func Reset() {
    global = nil
}
```

**Environment variable overrides:**

```go
func applyEnvOverrides(cfg *Config) {
    if env := os.Getenv("PDF_CLI_OUTPUT_FORMAT"); env != "" {
        cfg.Defaults.OutputFormat = env
    }
    if env := os.Getenv("PDF_CLI_VERBOSE"); env == "true" || env == "1" {
        cfg.Defaults.Verbose = true
    }
}
```

### Table-Driven Tests

Standard pattern across the codebase:

```go
func TestFormatFileSize(t *testing.T) {
    tests := []struct {
        bytes int64
        want  string
    }{
        {0, "0 B"},
        {1024, "1.00 KB"},
        {1024 * 1024, "1.00 MB"},
        {1024 * 1024 * 1024, "1.00 GB"},
    }
    for _, tt := range tests {
        if got := FormatFileSize(tt.bytes); got != tt.want {
            t.Errorf("FormatFileSize(%d) = %q, want %q", tt.bytes, got, tt.want)
        }
    }
}
```

### CLI Patterns

**Command initialization:**

```go
func init() {
    cli.AddCommand(compressCmd)
    cli.AddOutputFlag(compressCmd, "Output file path (only with single file)")
    cli.AddPasswordFlag(compressCmd, "Password for encrypted PDFs")
    cli.AddStdoutFlag(compressCmd)
}

var compressCmd = &cobra.Command{
    Use:   "compress <file.pdf> [file2.pdf...]",
    Short: "Compress and optimize PDF(s)",
    Long:  `...`,
    Args:  cobra.MinimumNArgs(1),
    RunE:  runCompress,
}
```

**Stdin/stdout handling via patterns:**

```go
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
```

## File Permissions

**Consistent permission model:**
- **Directories**: `0750` (owner rwx, group r-x)
- **Config files**: `0600` (owner rw only)
- **Temporary files**: Created with `os.CreateTemp()` (secure by default)
- **Test files**: `0644` (permissive OK for test fixtures)

**Examples:**
```go
os.MkdirAll(dir, 0750)           // Directories
os.WriteFile(path, data, 0600)   // Sensitive files like config
os.WriteFile(testFile, data, 0644) // Test fixtures
```

## Commit Message Convention

Follows Conventional Commits:

- `feat:` New feature
- `fix:` Bug fix
- `docs:` Documentation only
- `test:` Test changes
- `chore:` Maintenance tasks
- `refactor:` Code restructuring without behavior change

**Examples from git history:**
- `docs: update README and CHANGELOG for v1.5.0 release`
- `fix: add nosec directives for gosec false positives`
- `fix: update golangci-lint config for v2 and fix CI gosec`
- `docs: update architecture.md to reflect new package structure`

## Comments & Documentation

**Exported symbols**: Should have documentation comments (revive `exported` rule relaxed for CLI)

**Inline comments**: Used for:
- Security directives (`#nosec`)
- Complex logic explanations
- TODOs (if any)

**Function documentation**: Describes what function does, not how

```go
// FileExists checks if a file exists
func FileExists(path string) bool

// AtomicWrite writes data to a file atomically by writing to a temp file first
func AtomicWrite(path string, data []byte) error
```

## Build & Version Info

**Build-time injection via ldflags:**

```makefile
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"
```

**Variables in main.go:**

```go
var (
    version = "dev"
    commit  = "none"
    date    = "unknown"
)
```

## Pre-commit Workflow

Automated via `.pre-commit-config.yaml`:

1. **File hygiene**: trailing whitespace, EOF fixer, YAML validation
2. **Go formatting**: `go fmt`
3. **Go vetting**: `go vet`
4. **Dependency management**: `go mod tidy`
5. **Build verification**: `go build ./...`
6. **Test execution**: `go test ./...`
7. **Linting**: `golangci-lint run`

All hooks use `GO111MODULE=on` and `pass_filenames: false` for consistency.
