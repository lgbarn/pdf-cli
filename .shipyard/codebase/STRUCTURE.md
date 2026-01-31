# Project Structure

This document provides a detailed directory layout with purpose annotations for the pdf-cli codebase.

## Directory Tree

```
pdf-cli/
├── cmd/                          # Application entry points
│   └── pdf/                      # Main CLI application
│       └── main.go               # Minimal bootstrap: sets version, calls cli.Execute()
│
├── internal/                     # Private application code (not importable)
│   ├── cli/                      # CLI framework and global options
│   │   ├── cli.go                # Root command setup, global flags (verbose, force, progress)
│   │   ├── flags.go              # Reusable flag helpers (output, pages, password, format)
│   │   └── completion.go         # Shell completion generation
│   │
│   ├── commands/                 # Command implementations (self-registering)
│   │   ├── combine_images.go     # Create PDF from image files
│   │   ├── compress.go           # PDF optimization and compression
│   │   ├── decrypt.go            # Remove password protection
│   │   ├── encrypt.go            # Add password protection
│   │   ├── extract.go            # Extract specific pages to new PDF
│   │   ├── images.go             # Extract embedded images
│   │   ├── info.go               # Display PDF information (supports batch mode)
│   │   ├── merge.go              # Combine multiple PDFs
│   │   ├── meta.go               # View/modify metadata (title, author, etc.)
│   │   ├── pdfa.go               # PDF/A validation and conversion
│   │   ├── reorder.go            # Reorder, reverse, duplicate pages
│   │   ├── rotate.go             # Rotate pages by angle
│   │   ├── split.go              # Split PDF into pages or chunks
│   │   ├── text.go               # Extract text (supports OCR)
│   │   ├── watermark.go          # Add text/image watermarks
│   │   ├── helpers.go            # Shared command utilities (checkOutputFile, etc.)
│   │   └── patterns/             # Reusable command patterns
│   │       ├── doc.go            # Package documentation
│   │       └── stdio.go          # StdioHandler for stdin/stdout support
│   │
│   ├── config/                   # Configuration management
│   │   ├── config.go             # YAML config loading, env var overrides, singleton
│   │   └── doc.go                # Package documentation
│   │
│   ├── fileio/                   # File I/O operations
│   │   ├── files.go              # File validation, atomic write, copy, size formatting
│   │   ├── stdio.go              # stdin/stdout handling with temp file management
│   │   └── doc.go                # Package documentation
│   │
│   ├── logging/                  # Structured logging
│   │   ├── logger.go             # slog-based logging (text/JSON, debug/info/warn/error/silent)
│   │   └── doc.go                # Package documentation
│   │
│   ├── ocr/                      # OCR text extraction engine
│   │   ├── backend.go            # Backend interface definition and type constants
│   │   ├── ocr.go                # Main Engine with backend selection, image processing
│   │   ├── native.go             # Native Tesseract backend (exec)
│   │   ├── wasm.go               # WASM Tesseract backend (gogosseract)
│   │   ├── detect.go             # Native Tesseract detection logic
│   │   └── process.go            # Parallel/sequential image processing logic
│   │
│   ├── output/                   # Output formatting
│   │   ├── formatter.go          # JSON/CSV/TSV formatter with OutputFormatter type
│   │   ├── table.go              # Tabular output utilities for batch operations
│   │   └── doc.go                # Package documentation
│   │
│   ├── pages/                    # Page specification parsing
│   │   ├── parser.go             # Parse ranges like "1-5,7,10-12" → []PageRange
│   │   ├── reorder.go            # Parse reorder sequences like "end-1" or "1,5,2"
│   │   ├── validator.go          # Validate page numbers against PDF
│   │   └── doc.go                # Package documentation
│   │
│   ├── pdf/                      # Core PDF operations (modular by responsibility)
│   │   ├── pdf.go                # Base configuration and utility functions
│   │   ├── metadata.go           # Info, page count, metadata get/set
│   │   ├── transform.go          # Merge, split, extract, rotate, compress, images→PDF
│   │   ├── encryption.go         # Encrypt, decrypt operations
│   │   ├── text.go               # Text extraction (primary + fallback strategies)
│   │   ├── watermark.go          # Watermark operations
│   │   └── validation.go         # PDF/A validation and conversion
│   │
│   ├── pdferrors/                # Error handling
│   │   ├── errors.go             # PDFError type with operation context and hints
│   │   └── doc.go                # Package documentation
│   │
│   ├── progress/                 # Progress bar utilities
│   │   ├── progress.go           # Consistent progress bar creation (count/bytes)
│   │   └── doc.go                # Package documentation
│   │
│   └── testing/                  # Test infrastructure
│       ├── testhelpers.go        # Shared test utilities and helpers
│       └── mocks.go              # Mock implementations for testing
│
├── docs/                         # Documentation
│   ├── architecture.md           # Architecture patterns and design decisions
│   └── ...                       # Additional documentation
│
├── testdata/                     # Test fixtures
│   ├── sample.pdf                # Sample PDF for testing
│   ├── encrypted.pdf             # Password-protected PDF
│   └── ...                       # Other test PDFs and images
│
├── .github/                      # GitHub configuration
│   └── workflows/                # CI/CD pipelines
│       ├── ci.yaml               # Main CI: test, lint, build
│       └── release.yaml          # Release automation with goreleaser
│
├── scripts/                      # Utility scripts
│   └── ...                       # Build/release helper scripts
│
├── .shipyard/                    # Shipyard analysis output
│   └── codebase/                 # Codebase documentation
│       ├── ARCHITECTURE.md       # This architecture document
│       └── STRUCTURE.md          # This structure document
│
├── go.mod                        # Go module definition
├── go.sum                        # Dependency checksums
├── Makefile                      # Build automation (build, test, lint, coverage)
├── .golangci.yaml                # Linter configuration
├── .goreleaser.yaml              # Release configuration
├── .pre-commit-config.yaml       # Pre-commit hooks configuration
├── README.md                     # User documentation
├── CHANGELOG.md                  # Version history
├── CONTRIBUTING.md               # Contribution guidelines
├── SECURITY.md                   # Security policy
└── LICENSE                       # MIT license
```

## Package Organization

### Entry Point

**cmd/pdf/main.go** (17 lines)
- Sets version variables from build flags
- Calls `cli.SetVersion()` and `cli.Execute()`
- Minimal bootstrap - delegates everything to internal packages
- Exit code 1 on error

**Design rationale**: Keep main package as thin as possible to maximize testability.

### Presentation Layer

**internal/cli/** - CLI Framework
- **cli.go**: Root command definition, global flags, version template
- **flags.go**: Reusable flag functions (AddOutputFlag, GetPassword, etc.)
- **completion.go**: Shell completion for bash/zsh/fish/powershell

**Responsibilities**:
- Cobra root command setup and configuration
- Global flag definitions (verbose, force, progress, dry-run, logging)
- Utility functions for accessing flag values
- Version display formatting

**Dependencies**: cobra, internal/config, internal/logging

### Application Layer

**internal/commands/** - Command Handlers
Each command is self-contained in its own file:
- `init()` function registers command with CLI framework
- `cobra.Command` definition with flags and help text
- `runXxx()` handler function with business logic orchestration
- Batch mode support where applicable (compress, encrypt, rotate, etc.)

**Common patterns**:
- Validate inputs using `fileio.ValidatePDFFile()`
- Check output file with `checkOutputFile()`
- Handle dry-run mode early
- Call domain layer (internal/pdf) for actual operations
- Format output appropriately

**stdin/stdout support** (compress, extract, rotate, reorder, encrypt, decrypt):
- Use `patterns.StdioHandler` for temp file management
- Support `-` as stdin indicator
- Support `--stdout` flag for binary output

**Batch mode support** (compress, encrypt, decrypt, rotate, watermark, meta):
- Detect when multiple input files provided
- Generate output filenames with suffix (e.g., `_compressed.pdf`)
- Process files individually with error handling

**internal/commands/patterns/** - Reusable Patterns
- **stdio.go**: `StdioHandler` struct for stdin/stdout pipeline support
  - `Setup()`: Resolves input/output paths, handles temp files
  - `Finalize()`: Writes to stdout if needed
  - `Cleanup()`: Removes temp files

**Design rationale**: Extract common patterns to reduce duplication across commands.

### Domain Layer

**internal/pdf/** - PDF Operations (Modular by Responsibility)
- **pdf.go**: Base config creation, page number utilities
- **metadata.go**: Info, page count, metadata operations
- **transform.go**: Merge, split, extract, rotate, compress, image→PDF conversion
- **encryption.go**: Encrypt and decrypt with password support
- **text.go**: Text extraction with primary (ledongthuc/pdf) and fallback (pdfcpu) strategies
- **watermark.go**: Text and image watermark operations
- **validation.go**: PDF/A validation and optimization

**Pattern**: Each file has a focused responsibility. Functions are pure operations that:
- Accept paths and parameters
- Return results or errors
- Have no CLI dependencies
- Are easily testable

**Design rationale**:
- Modular organization by feature area (not monolithic pdf.go)
- Wraps pdfcpu and ledongthuc/pdf libraries with consistent API
- Progress bar support via optional `showProgress` parameters

**internal/ocr/** - OCR Engine
- **backend.go**: Backend interface definition
  ```go
  type Backend interface {
      Name() string
      Available() bool
      ProcessImage(ctx context.Context, imagePath, lang string) (string, error)
      Close() error
  }
  ```
- **ocr.go**: Engine with backend selection logic
  - `NewEngine()`: Auto-select best backend
  - `NewEngineWithOptions()`: Explicit backend choice
  - `ExtractTextFromPDF()`: Orchestrates PDF→images→OCR→text
  - `EnsureTessdata()`: Downloads language data for WASM
- **native.go**: System Tesseract backend (exec.Command)
- **wasm.go**: WASM Tesseract backend (gogosseract library)
- **detect.go**: Tesseract detection (PATH and TESSDATA_PREFIX)

**Design patterns**:
- Strategy pattern for backend selection
- Auto-selection: native if available, else WASM
- Parallel processing for native (thread-safe), sequential for WASM
- Worker pool with semaphore for concurrency limiting

**internal/pages/** - Page Specification Parsing
- **parser.go**: Parse page ranges
  - "1-5,7,10-12" → `[]PageRange{{1,5}, {7,7}, {10,12}}`
  - `ExpandPageRanges()` → `[]int{1,2,3,4,5,7,10,11,12}`
  - `FormatPageRanges()` - Reverse operation for display
- **reorder.go**: Parse reorder sequences
  - "1,5,2,3,4" - Move page 5 to position 2
  - "end-1" - Reverse all pages
  - "1-end,1" - Duplicate page 1 at end
- **validator.go**: Validate page numbers against PDF page count

**Design rationale**: Domain-specific language for page specifications. Parsing isolated from PDF operations.

### Infrastructure Layer

**internal/fileio/** - File I/O Operations
- **files.go**: Core file operations
  - `FileExists()`, `IsDir()`, `EnsureDir()`
  - `AtomicWrite()`: Temp file + rename for safe writes
  - `CopyFile()`: Safe file copying with path cleaning
  - `ValidatePDFFile()`: Extension and existence checking
  - `ValidatePDFFiles()`: Parallel validation for >3 files
  - `GenerateOutputFilename()`: Suffix-based name generation
  - `FormatFileSize()`: Human-readable sizes (KB, MB, GB)
  - `IsImageFile()`: Image extension checking
- **stdio.go**: stdin/stdout handling
  - `ReadFromStdin()`: stdin → temp PDF file + cleanup function
  - `WriteToStdout()`: File → stdout
  - `ResolveInputPath()`: Handle "-" as stdin indicator

**Design rationale**: Centralize all filesystem concerns. Atomic operations for safety. Cleanup via returned functions.

**internal/output/** - Output Formatting
- **formatter.go**: Format converter
  - `OutputFormatter` struct with format type
  - `Print()`: Output as JSON
  - `PrintTable()`: Output as JSON/CSV/TSV/human table
  - `IsStructured()`: Check if format is machine-readable
- **table.go**: Table formatting utilities
  - CSV writer with proper escaping
  - TSV writer
  - JSON array of objects
  - Human-readable aligned columns

**Design rationale**: Separate presentation concerns from business logic. Support scripting with structured formats.

**internal/config/** - Configuration Management
- **config.go**: Configuration loading and access
  - `Config` struct with nested sections (Defaults, Compress, Encrypt, OCR)
  - `Load()`: YAML + env var loading with overrides
  - `Get()`: Singleton access pattern
  - `Save()`: Write config to disk
  - `ConfigPath()`: XDG-compliant path resolution

**Pattern**: Singleton with lazy initialization
```go
var global *Config

func Get() *Config {
    if global == nil {
        global, _ = Load()
    }
    return global
}
```

**Environment variables**: `PDF_CLI_*` prefix overrides config file

**Design rationale**: Single source of truth for defaults. Config file is optional - app works without it.

### Cross-Cutting Concerns

**internal/progress/** - Progress Bars
- **progress.go**: Consistent progress bar creation
  - `NewProgressBar()`: Threshold-based progress (only if >threshold items)
  - `NewBytesProgressBar()`: For downloads with byte counts
  - `FinishProgressBar()`: Add newline after completion

**Design rationale**: Centralize progress bar styling and threshold logic.

**internal/pdferrors/** - Error Handling
- **errors.go**: Custom error types
  - `PDFError` struct with operation, file, cause, hint
  - `WrapError()`: Add context to errors with pattern detection
  - Error constants: `ErrFileNotFound`, `ErrPasswordRequired`, etc.
  - `FormatError()`: User-friendly error display

**Pattern**: Wrap errors with context at each layer
```go
if err := pdf.Merge(files, output, pass); err != nil {
    return pdferrors.WrapError("merging files", output, err)
}
```

**Design rationale**: Provide actionable error messages. Detect common patterns and suggest fixes.

**internal/logging/** - Structured Logging
- **logger.go**: slog wrapper
  - Configurable levels: debug, info, warn, error, silent
  - Configurable formats: text, JSON
  - Global logger initialization
  - Helper functions: `Debug()`, `Info()`, `Warn()`, `Error()`

**Design rationale**: Structured logging for debugging and monitoring without cluttering normal output.

**internal/testing/** - Test Infrastructure
- **testhelpers.go**: Shared test utilities
- **mocks.go**: Mock implementations for interfaces

**Design rationale**: Reduce test duplication. Make mocking easier.

## File Size Distribution

**Total**: 87 Go source files

**By package** (excluding tests):
- commands/: 16 files (one per command + helpers + patterns)
- pdf/: 7 files (modular by responsibility)
- ocr/: 5 files (engine + backends)
- cli/: 3 files (framework + flags + completion)
- pages/: 3 files (parser + reorder + validator)
- fileio/: 2 files (files + stdio)
- output/: 2 files (formatter + table)
- config/: 1 file
- logging/: 1 file
- progress/: 1 file
- pdferrors/: 1 file
- testing/: 2 files

**Line count range**: Most files are 100-300 lines, keeping them focused and maintainable.

## Import Rules

### Allowed Import Directions
```
cmd/pdf → internal/cli, internal/commands
internal/commands → internal/cli, internal/pdf, internal/ocr, internal/pages, internal/fileio, internal/output, internal/pdferrors
internal/pdf → external libraries (pdfcpu, ledongthuc/pdf)
internal/ocr → internal/fileio, internal/pdf, external libraries
internal/pages → (no internal dependencies)
internal/fileio → (no internal dependencies except golang.org/x/term)
internal/output → (no internal dependencies)
internal/config → (no internal dependencies)
```

### Forbidden Import Directions
- Domain packages (pdf, ocr) CANNOT import CLI packages
- Infrastructure packages (fileio, output) CANNOT import domain packages
- Lower layers CANNOT import higher layers

**Design rationale**: Enforce unidirectional dependency flow. Keep domain logic pure.

## Configuration Files

**go.mod** - Go module definition
- Module: `github.com/lgbarn/pdf-cli`
- Go version: 1.24.1
- Key dependencies: pdfcpu, ledongthuc/pdf, gogosseract, cobra, progressbar

**Makefile** - Build automation targets
- `build`: Compile binary
- `test`: Run tests
- `test-coverage`: Generate coverage report
- `lint`: Run golangci-lint
- `clean`: Remove build artifacts
- `check-all`: Full pre-commit check suite

**.golangci.yaml** - Linter configuration
- Enabled linters: staticcheck, gosec, errcheck, etc.
- Custom rules and exclusions
- Configured for v2 format

**.goreleaser.yaml** - Release configuration
- Multi-platform builds (Linux, macOS, Windows)
- Archive creation
- Checksums and signatures

**.pre-commit-config.yaml** - Pre-commit hooks
- Format checking
- Linting
- Test running

## Data Files

**testdata/** - Test fixtures
- Sample PDFs for various scenarios
- Encrypted PDFs
- Image files for combine-images testing
- Scanned PDFs for OCR testing

**Purpose**: Provide consistent test data without generating files in tests.

## Build Artifacts (gitignored)

- `pdf` - Main binary (build output)
- `*.out` - Coverage reports
- `dist/` - Goreleaser output
- `.worktrees/` - Git worktree checkouts

## Development Workflow

1. **Add new command**: Create file in `internal/commands/`, define command, add `init()` registration
2. **Add business logic**: Extend appropriate package in `internal/pdf/` or create new package
3. **Add tests**: Create `*_test.go` alongside source files
4. **Run checks**: `make check-all` (test + lint + coverage)
5. **Commit**: Pre-commit hooks run automatically
6. **Release**: Tag triggers GitHub Actions release workflow

## Key Design Principles

1. **Thin layers**: Each layer is minimal and focused
2. **Self-registration**: Commands register themselves via init()
3. **Pure domain logic**: PDF operations have no CLI dependencies
4. **Temp file safety**: Always use defer for cleanup
5. **Batch mode support**: Maximize user productivity
6. **Unix philosophy**: stdin/stdout for composability
7. **Progressive complexity**: Simple defaults, advanced options available
8. **Testability**: Clear module boundaries enable easy testing
