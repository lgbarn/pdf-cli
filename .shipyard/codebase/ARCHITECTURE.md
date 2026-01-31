# Architecture

This document describes the architectural patterns, layers, and data flow of pdf-cli.

## Overview

pdf-cli is a CLI application built in Go following a **layered architecture** with clear separation of concerns. The application is organized around:

- **Command-driven architecture** using Cobra for CLI framework
- **Modular domain layer** for PDF operations
- **Dependency injection** through package initialization
- **Strategy pattern** for OCR backends
- **Pipeline pattern** for stdin/stdout support

## Architectural Patterns

### 1. Layered Architecture

```
┌─────────────────────────────────────────┐
│         Presentation Layer              │
│  (cmd/pdf/main.go, internal/cli)        │
└─────────────────┬───────────────────────┘
                  │
┌─────────────────▼───────────────────────┐
│       Command/Application Layer         │
│        (internal/commands/*)            │
└─────────────────┬───────────────────────┘
                  │
┌─────────────────▼───────────────────────┐
│         Domain/Business Layer           │
│  (internal/pdf, internal/ocr, etc.)     │
└─────────────────┬───────────────────────┘
                  │
┌─────────────────▼───────────────────────┐
│      Infrastructure/Utility Layer       │
│  (internal/fileio, internal/output)     │
└─────────────────────────────────────────┘
```

#### Presentation Layer
- **Entry Point**: `/cmd/pdf/main.go` - Minimal bootstrap that sets version info and delegates to CLI framework
- **CLI Framework**: `/internal/cli/` - Cobra root command setup, flag management, global options
- **Pattern**: Thin presentation layer that delegates all logic to command handlers

#### Application Layer
- **Command Handlers**: `/internal/commands/` - One file per command (merge.go, split.go, etc.)
- **Command Patterns**: `/internal/commands/patterns/` - Reusable patterns like StdioHandler
- **Pattern**: Each command is self-contained with its own RunE handler
- **Registration**: Commands register themselves via `init()` functions calling `cli.AddCommand()`

#### Domain Layer
- **PDF Operations**: `/internal/pdf/` - Core PDF business logic (modular by responsibility)
- **OCR Engine**: `/internal/ocr/` - Text extraction from scanned PDFs
- **Page Parsing**: `/internal/pages/` - Page range and reorder sequence parsing
- **Pattern**: Pure business logic with no CLI or I/O concerns

#### Infrastructure Layer
- **File I/O**: `/internal/fileio/` - File operations, stdin/stdout handling, validation
- **Output Formatting**: `/internal/output/` - JSON/CSV/TSV formatting
- **Configuration**: `/internal/config/` - YAML config and environment variables
- **Progress**: `/internal/progress/` - Progress bar utilities
- **Error Handling**: `/internal/pdferrors/` - Domain-specific error types
- **Logging**: `/internal/logging/` - Structured logging with slog

### 2. Command Registration Pattern

Commands use **self-registration** via `init()` functions:

```go
// internal/commands/merge.go
func init() {
    cli.AddCommand(mergeCmd)
    cli.AddOutputFlag(mergeCmd, "Output file path (required)")
    cli.AddPasswordFlag(mergeCmd, "Password for encrypted input PDFs")
    _ = mergeCmd.MarkFlagRequired("output")
}

var mergeCmd = &cobra.Command{
    Use:   "merge <file1.pdf> <file2.pdf> [file3.pdf...]",
    Short: "Merge multiple PDFs into one",
    RunE:  runMerge,
}
```

The main package imports commands with a **blank import**:
```go
// cmd/pdf/main.go
import _ "github.com/lgbarn/pdf-cli/internal/commands"
```

This enables:
- Automatic command discovery
- No central registration list
- Easy addition of new commands
- Clear ownership boundaries

### 3. Strategy Pattern for OCR Backends

The OCR engine uses the **strategy pattern** to support multiple backends:

```
┌──────────────┐
│  OCR Engine  │
└──────┬───────┘
       │ uses
       ▼
┌──────────────┐        ┌─────────────────┐
│   Backend    │◄───────│ NativeBackend   │
│  (interface) │        │ (Tesseract CLI) │
└──────────────┘        └─────────────────┘
       ▲
       │                ┌─────────────────┐
       └────────────────│   WASMBackend   │
                        │  (gogosseract)  │
                        └─────────────────┘
```

**Backend Interface**:
```go
type Backend interface {
    Name() string
    Available() bool
    ProcessImage(ctx context.Context, imagePath, lang string) (string, error)
    Close() error
}
```

**Selection Logic** (`internal/ocr/ocr.go`):
- `BackendAuto`: Try native first, fall back to WASM
- `BackendNative`: Force system Tesseract
- `BackendWASM`: Force built-in WASM version

### 4. Pipeline Pattern for stdin/stdout

Commands supporting Unix pipelines use the **StdioHandler** pattern:

```go
// internal/commands/patterns/stdio.go
type StdioHandler struct {
    InputArg       string
    ExplicitOutput string
    ToStdout       bool
    DefaultSuffix  string
    Operation      string
}
```

**Lifecycle**:
1. `Setup()` - Resolves input (stdin → temp file) and output paths
2. **Operation** - Command performs work on resolved paths
3. `Finalize()` - Writes output to stdout if needed
4. `Cleanup()` - Removes temp files (via defer)

**Example Usage**:
```go
handler := &patterns.StdioHandler{
    InputArg:       args[0],
    ExplicitOutput: cli.GetOutput(cmd),
    ToStdout:       cli.GetStdout(cmd),
    DefaultSuffix:  "_compressed",
    Operation:      "compress",
}
defer handler.Cleanup()

input, output, err := handler.Setup()
// ... perform operation ...
return handler.Finalize()
```

### 5. Dependency Flow

```
cmd/pdf/main.go
    ↓ (blank import)
internal/commands/*.go (init functions)
    ↓ (registers to)
internal/cli/cli.go (root command)
    ↓ (uses)
internal/pdf/*.go (business logic)
    ↓ (calls)
external libraries (pdfcpu, ledongthuc/pdf)
```

**Key principle**: Dependencies flow **downward only**. Lower layers never import higher layers.

## Data Flow

### Command Execution Flow

```
1. User executes: pdf merge -o out.pdf file1.pdf file2.pdf
                     ↓
2. Cobra routes to mergeCmd.RunE handler
                     ↓
3. runMerge validates input files (fileio.ValidatePDFFiles)
                     ↓
4. runMerge calls pdf.MergeWithProgress
                     ↓
5. pdf.MergeWithProgress uses pdfcpu API
                     ↓
6. Success message printed to stdout
```

### stdin/stdout Pipeline Flow

```
1. User: cat input.pdf | pdf compress - --stdout > out.pdf
                     ↓
2. fileio.ResolveInputPath detects "-" → reads stdin to temp file
                     ↓
3. Command operates on temp file
                     ↓
4. fileio.WriteToStdout writes result to stdout
                     ↓
5. Cleanup removes temp files
```

### OCR Processing Flow

```
1. User: pdf text scanned.pdf --ocr
                     ↓
2. ocr.NewEngine selects backend (native or WASM)
                     ↓
3. Engine.ExtractTextFromPDF extracts images from PDF
                     ↓
4. For each image: backend.ProcessImage (parallel for native, sequential for WASM)
                     ↓
5. Results aggregated and returned as text
```

### Configuration Loading Flow

```
1. config.Get() called on first access
                     ↓
2. Load default values (config.DefaultConfig)
                     ↓
3. Override with YAML from ~/.config/pdf-cli/config.yaml (if exists)
                     ↓
4. Override with environment variables (PDF_CLI_*)
                     ↓
5. Return merged configuration
```

## Module Boundaries

### Core Domain Modules

**internal/pdf/** - PDF Operations (modular by responsibility)
- `pdf.go` - Base configuration and utilities
- `metadata.go` - Info, page count, metadata operations
- `transform.go` - Merge, split, extract, rotate, compress
- `encryption.go` - Encrypt, decrypt operations
- `text.go` - Text extraction (with fallback strategy)
- `watermark.go` - Watermark operations
- `validation.go` - PDF/A validation

**Boundaries**:
- No CLI dependencies
- No direct I/O (paths provided by callers)
- Pure business logic
- Wraps pdfcpu and ledongthuc/pdf libraries

**internal/ocr/** - OCR Engine
- `backend.go` - Backend interface and types
- `ocr.go` - Main engine with backend selection
- `native.go` - System Tesseract backend
- `wasm.go` - WASM Tesseract backend
- `detect.go` - Native Tesseract detection
- `process.go` - Image processing logic

**Boundaries**:
- Self-contained OCR logic
- Backend abstraction via interface
- Downloads tessdata on-demand for WASM
- Parallel processing for performance

### Infrastructure Modules

**internal/fileio/** - File Operations
- `files.go` - File validation, copy, atomic write
- `stdio.go` - stdin/stdout handling with temp files

**Responsibilities**:
- File existence and extension validation
- stdin → temp file conversion
- stdout writing from temp files
- Path resolution

**internal/pages/** - Page Parsing
- `parser.go` - Parse ranges like "1-5,7,10-12"
- `reorder.go` - Parse reorder sequences like "end-1" or "1,5,2,3,4"
- `validator.go` - Validate pages against PDF page count

**Boundaries**:
- Domain-specific language for page specifications
- No PDF operations, just parsing
- Used by commands to convert user input to page numbers

**internal/output/** - Output Formatting
- `formatter.go` - JSON/CSV/TSV formatting
- `table.go` - Tabular output utilities

**Responsibilities**:
- Convert Go structs to JSON
- Convert table data to CSV/TSV
- Support for batch command output

### Cross-Cutting Modules

**internal/config/** - Configuration Management
- Singleton pattern for global config
- YAML file loading from XDG_CONFIG_HOME
- Environment variable overrides
- Default values

**internal/progress/** - Progress Bars
- Consistent progress bar theming
- Threshold-based display (only for large operations)
- Byte-based progress for downloads

**internal/pdferrors/** - Error Handling
- Custom error types with context
- Error wrapping with user-friendly messages
- Hint system for common errors

**internal/logging/** - Structured Logging
- slog-based structured logging
- Text and JSON output formats
- Debug/info/warn/error/silent levels

## Concurrency Patterns

### Parallel File Validation

When merging >3 files, validation runs in parallel:
```go
// internal/fileio/files.go
for _, path := range paths {
    go func(p string) {
        results <- result{path: p, err: ValidatePDFFile(p)}
    }(path)
}
```

### Parallel Text Extraction

When extracting >5 pages, text extraction is parallelized:
```go
// internal/pdf/text.go
for _, pageNum := range pages {
    go func(pn int) {
        results <- pageResult{pageNum: pn, text: extractPageText(r, pn, totalPages)}
    }(pageNum)
}
```

### Parallel OCR Processing

Native OCR backend uses worker pool pattern:
```go
// internal/ocr/ocr.go
workers := min(runtime.NumCPU(), 8)
sem := make(chan struct{}, workers)

for i, imgPath := range imageFiles {
    sem <- struct{}{}  // Acquire
    go func(idx int, path string) {
        defer func() { <-sem }()  // Release
        text, _ := e.backend.ProcessImage(ctx, path, e.lang)
        results <- imageResult{index: idx, text: text}
    }(i, imgPath)
}
```

**Note**: WASM backend uses sequential processing (not thread-safe).

## Error Handling Strategy

### Layered Error Context

Errors are wrapped at each layer to add context:

```go
// Command layer
if err := pdf.MergeWithProgress(args, output, password, cli.Progress()); err != nil {
    return pdferrors.WrapError("merging files", output, err)
}

// Error wrapper adds user-friendly context
type PDFError struct {
    Operation string  // "merging files"
    File      string  // "output.pdf"
    Cause     error   // underlying error
    Hint      string  // suggestion for user
}
```

### Common Error Patterns

**File Not Found**:
```go
ErrFileNotFound = errors.New("file not found")
```

**Password Required**:
```go
return &PDFError{
    Operation: operation,
    File:      file,
    Cause:     ErrPasswordRequired,
    Hint:      "Use --password to provide the document password",
}
```

### Error Detection

The `WrapError` function detects common error patterns:
```go
switch {
case strings.Contains(errStr, "encrypted") || strings.Contains(errStr, "password"):
    return passwordRequiredError
case strings.Contains(errStr, "invalid PDF") || strings.Contains(errStr, "malformed"):
    return corruptPDFError
default:
    return genericError
}
```

## Testing Strategy

### Test Organization

- Unit tests alongside source files (`*_test.go`)
- Integration tests in `commands_integration_test.go`
- Test helpers in `internal/testing/`
- Test data in `/testdata/`

### Test Patterns

**Table-driven tests** for parsing logic:
```go
tests := []struct {
    name    string
    input   string
    want    []PageRange
    wantErr bool
}{
    // test cases
}
```

**Mock backends** for OCR testing:
```go
type MockBackend struct {
    name      string
    available bool
    results   map[string]string
}
```

## Build and Deployment

### Build Tags

No build tags currently used. All functionality compiled in by default.

### Version Injection

Version info injected at build time:
```go
// cmd/pdf/main.go
var (
    version = "dev"
    commit  = "none"
    date    = "unknown"
)
```

Set via linker flags:
```bash
go build -ldflags="-X main.version=1.5.0 -X main.commit=$(git rev-parse HEAD)"
```

### Single Binary Distribution

All dependencies compiled into single static binary:
- Native libraries wrapped (no CGo)
- WASM runtime embedded
- No external runtime dependencies

## Performance Characteristics

### Optimization Strategies

1. **Threshold-based parallelization**: Small operations run sequentially to avoid overhead
2. **Progress bar gating**: Only shown for operations with >threshold items
3. **Incremental merging**: Large merges done incrementally with temp files
4. **Worker pool limiting**: OCR parallelism capped at `min(NumCPU, 8)`
5. **Lazy backend selection**: OCR backend only initialized when needed

### Memory Management

- Temp files used for stdin/stdout to avoid loading entire PDFs in memory
- Streaming where possible (file I/O, image extraction)
- Cleanup via defer to ensure temp file removal

## Future Extensibility Points

### Adding New Commands

1. Create `/internal/commands/newcommand.go`
2. Define cobra.Command with `RunE` handler
3. Register in `init()` function
4. Implement business logic in appropriate domain package

### Adding New OCR Backends

1. Implement `Backend` interface
2. Add backend detection logic
3. Update `selectBackend()` in ocr.go
4. Add new BackendType constant

### Adding New Output Formats

1. Add format constant to `internal/output/formatter.go`
2. Implement format-specific print method
3. Update `Print()` or `PrintTable()` switch statement
4. Update CLI flag help text

## Key Design Decisions

### Why Cobra for CLI?

- Industry-standard Go CLI framework
- Built-in flag parsing and help generation
- Command composition and subcommands
- Easy shell completion support

### Why Self-Registration Pattern?

- Decouples commands from main package
- Enables easy addition/removal of commands
- Clear ownership boundaries
- No central "registry" to maintain

### Why Separate fileio Package?

- Centralizes I/O concerns
- Makes commands testable without filesystem
- Consistent stdin/stdout handling
- Platform-specific I/O could be isolated here

### Why Multiple PDF Libraries?

- **pdfcpu**: Full-featured for manipulation (merge, split, etc.)
- **ledongthuc/pdf**: Better text extraction quality
- Fallback strategy ensures best results

### Why Both Native and WASM OCR?

- **Native**: Better quality, faster, but requires system dependencies
- **WASM**: Zero dependencies, portable, auto-downloads data
- Auto-selection provides best UX
