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
├── config/           Configuration file support
├── fileio/           File operations and stdio utilities
├── logging/          Structured logging with slog
├── ocr/              OCR engine with pluggable backends
├── output/           Output formatting (JSON, CSV, TSV, human)
├── pages/            Page range parsing and validation
├── pdf/              PDF operations (wrapper around pdfcpu)
├── pdferrors/        Error handling with context and hints
├── progress/         Progress bar utilities
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

### fileio/
- File operations and validation
- Stdin/stdout utilities
- File size formatting
- Temporary file management

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

### logging/
- Structured logging with slog (Go 1.21+)
- Multiple formats (text, JSON)
- Multiple levels (debug, info, warn, error, silent)

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

All errors use `pdferrors.WrapError()` for consistent formatting:
- Operation context (what was being done)
- File context (which file)
- Underlying error
- User-friendly hints for common issues (e.g., password hints for encrypted PDFs)

## Testing Strategy

- Unit tests co-located with source files
- Integration tests in `commands_integration_test.go`
- Mocks in `internal/testing/` for isolation
- Table-driven tests for comprehensive coverage
