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
