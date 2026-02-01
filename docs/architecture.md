# pdf-cli Architecture

## Overview

pdf-cli is a command-line tool for PDF manipulation built with Go. It follows a layered architecture with clear separation between CLI handling, business logic, and external dependencies.

## Package Structure

```
cmd/pdf/              Entry point
internal/
├── cli/              CLI framework (Cobra wrapper, flags, output)
├── commands/         Command implementations (14 commands)
│   └── patterns/     Reusable command patterns (StdioHandler)
├── cleanup/          Signal-based temp file cleanup registry
├── config/           Configuration file support
├── fileio/           File operations and stdio utilities
├── logging/          Structured logging with slog
├── ocr/              OCR engine with pluggable backends
├── output/           Output formatting (JSON, CSV, TSV, human)
├── pages/            Page range parsing and validation
├── pdf/              PDF operations (wrapper around pdfcpu)
├── pdferrors/        Error handling with context and hints
├── progress/         Progress bar utilities
├── retry/            Exponential backoff retry logic
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
           commands/     config/    logging/
                │
    ┌───────────┼───────────┬───────────┐
    ▼           ▼           ▼           ▼
  pdf/        ocr/      fileio/     pages/
    │           │           │           │
    ▼           ▼           ▼           ▼
 pdfcpu   gogosseract   pdferrors/  output/
```

**Key principles:**
- No circular dependencies
- Commands depend on core packages, not vice versa
- Leaf packages (fileio, pages, output, pdferrors, progress) have minimal dependencies
- External dependencies isolated in pdf/ and ocr/
- config/ and logging/ integrate with cli/ for global state

## Package Responsibilities

### cli/
- Root command setup and version info
- Shared flag definitions (output, password, verbose, force)
- Secure password reading (ReadPassword) with 4-tier priority: password-file, env var, flag (deprecated), interactive prompt
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
- Language data management with retry and checksum verification
- Image-to-text conversion
- Configurable parallelism via PerformanceConfig
- Error collection with errors.Join for parallel operations

### fileio/
- File operations and validation
- Path sanitization (SanitizePath) against directory traversal
- Stdin/stdout utilities
- File size formatting
- Temporary file management
- AtomicWrite with cleanup registration
- CopyFile with close error propagation

### pages/
- Page range parsing (supports "1-5,7,end-1")
- Page number validation
- Reorder sequence parsing

### output/
- Output formatting (JSON, CSV, TSV, human)
- Table formatting utilities

### pdferrors/
- PDFError type with operation, file, cause, and hint
- Error wrapping with context
- User-friendly error messages

### progress/
- Progress bar utilities for long operations

### config/
- YAML configuration file support (~/.config/pdf-cli/config.yaml)
- Environment variable overrides (PDF_CLI_*)
- Default values per command
- PerformanceConfig with adaptive defaults based on runtime.NumCPU()
- Thread-safe singleton initialization with sync.Once

### logging/
- Structured logging with slog (Go 1.21+)
- Multiple formats (text, JSON)
- Multiple levels (debug, info, warn, error, silent)
- Thread-safe singleton initialization with sync.Once

### cleanup/
- Thread-safe cleanup registry for temporary files
- Register/Run API for deferred cleanup
- Integrated with signal handler in main.go for SIGINT/SIGTERM cleanup
- Prevents resource leaks on abnormal termination

### retry/
- Generic retry helper with exponential backoff
- PermanentError type for non-retryable errors
- Used by tessdata downloads for network resilience
- Configurable max attempts and backoff intervals

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

### Why adaptive parallelism?
- PerformanceConfig defaults based on `runtime.NumCPU()`
- Configurable via environment variables for fine-tuning
- Balances performance with resource usage
- Prevents overwhelming the system on machines with many cores

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

All errors use `pdferrors.WrapError()` for consistent formatting:
- Operation context (what was being done)
- File context (which file)
- Underlying error
- User-friendly hints for common issues (e.g., password hints for encrypted PDFs)

**Error Propagation:**
- Close errors are now properly propagated with named returns
- Parallel processing collects all errors via `errors.Join`
- Retry logic distinguishes between transient and permanent errors

## Signal Handling and Lifecycle

**Signal handling flow:**
1. `main.go` creates context with `signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)`
2. Temporary files registered with `cleanup.Register(path)`
3. On SIGINT/SIGTERM, context is cancelled
4. `defer cleanup.Run()` ensures all temp files are removed
5. Graceful shutdown with proper resource cleanup

**Context propagation:**
- All PDF and OCR operations accept `context.Context`
- Enables cancellation of long-running operations
- Supports timeout and deadline propagation

## Testing Strategy

- Unit tests co-located with source files
- Integration tests in `commands_integration_test.go`
- Mocks in `internal/testing/` for isolation
- Table-driven tests for comprehensive coverage
