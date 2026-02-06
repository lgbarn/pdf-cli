# Structure

## Overview

pdf-cli uses a hierarchical package structure with clear separation of concerns: a single entry point in `cmd/`, domain logic and commands in `internal/`, test fixtures in `testdata/`, and supporting files at the root. The `internal/` tree enforces encapsulation with 16 packages organized into functional layers. Documentation, CI/CD, and build tooling reside at the project root.

## Findings

### Root Directory Layout

```
/Users/lgbarn/Personal/pdf-cli/
├── cmd/                      # Entry points (1 binary)
├── internal/                 # Private packages (16 packages)
├── testdata/                 # Test fixtures (PDF files)
├── docs/                     # Documentation
├── .github/workflows/        # CI/CD pipelines
├── scripts/                  # Build/maintenance scripts
├── go.mod                    # Go module definition
├── go.sum                    # Dependency checksums
├── Makefile                  # Build automation
├── README.md                 # User documentation
├── CHANGELOG.md              # Release history
├── CONTRIBUTING.md           # Contribution guidelines
├── LICENSE                   # MIT license
├── SECURITY.md               # Security policy
└── .shipyard/                # Codebase analysis output
```

**Purpose**: Standard Go project layout with clear separation of public API (none), entry points, internal implementation, and supporting files.

### Entry Point: `cmd/`

**Path**: `/Users/lgbarn/Personal/pdf-cli/cmd/pdf/`
- **Purpose**: Single binary entry point
- **Files**:
  - `main.go` -- Program entry, signal handling, cleanup orchestration, version injection
    - Evidence: Lines 21-35 define `run()` with context creation, signal handling, cleanup defer
    - Creates `signal.NotifyContext` for SIGINT/SIGTERM
    - Calls `cli.ExecuteContext(ctx)` to run commands
    - Version variables (`version`, `commit`, `date`) set via build flags

**Key Characteristics**:
- No business logic -- delegates to `internal/cli`
- Imports only `internal/cli`, `internal/cleanup`, and `internal/commands` (for side-effect registration)
- `internal/commands` import is blank (`_ "..."`) to trigger `init()` functions

### Internal Packages: `internal/`

**Path**: `/Users/lgbarn/Personal/pdf-cli/internal/`
- **Purpose**: Private implementation, not importable by external projects
- **Package Count**: 16 packages (14 documented + 2 inferred from structure)

#### CLI Layer: `internal/cli/`

**Purpose**: Cobra wrapper, global flags, password handling, logging initialization
- **Files**:
  - `cli.go` -- Root command setup, version templating, command registration API
    - Evidence: Lines 25-48 define root command with Use, Short, Long, Version
    - Evidence: Lines 50-67 define `init()` with config loading, flag registration, logging initialization
  - `flags.go` -- Flag registration helpers (`AddOutputFlag`, `AddPasswordFlag`, `AddStdoutFlag`, etc.)
  - `password.go` -- Secure password reading with 4-tier priority: file > env > flag > interactive
  - `completion.go` -- Shell completion generation (bash, zsh, fish, powershell)
  - `cli_test.go`, `flags_test.go`, `password_test.go`, `completion_test.go` -- Tests

**Exports**:
- `AddCommand(cmd *cobra.Command)` -- Register subcommands
- `Verbose()`, `Force()`, `Progress()`, `IsDryRun()` -- Flag accessors
- `GetPasswordSecure()` -- Secure password retrieval
- `PrintVerbose()`, `PrintStatus()`, `DryRunPrint()` -- Output helpers

**Dependencies**: `internal/config`, `internal/logging`, `github.com/spf13/cobra`, `golang.org/x/term`

#### Commands Layer: `internal/commands/`

**Purpose**: Command implementations, batch processing, stdin/stdout handling
- **Command Files** (14 commands):
  - `info.go` -- Display PDF information (single + batch modes)
  - `merge.go` -- Combine multiple PDFs
  - `split.go` -- Split PDF into pages or chunks
  - `extract.go` -- Extract specific pages
  - `reorder.go` -- Reorder, reverse, duplicate pages
  - `rotate.go` -- Rotate pages by angle
  - `compress.go` -- Optimize PDF size
  - `encrypt.go` -- Add password protection
  - `decrypt.go` -- Remove password protection
  - `text.go` -- Extract text (with OCR support)
  - `images.go` -- Extract embedded images
  - `combine_images.go` -- Create PDF from images
  - `meta.go` -- View/modify metadata
  - `watermark.go` -- Add text/image watermarks
  - `pdfa.go` -- PDF/A validation and conversion

- **Helper Files**:
  - `helpers.go` -- Shared helpers: `checkOutputFile()`, `parseAndValidatePages()`, `outputOrDefault()`, `sanitizeInputArgs()`, `processBatch()`
    - Evidence: Lines 14-96 define reusable command utilities

- **Test Files**: 13 test files covering commands, integration tests, batch operations, dry-run mode
  - `commands_test.go`, `commands_integration_test.go`, `integration_batch_test.go`, `integration_content_test.go`, `dryrun_test.go`, etc.

**Sub-package**: `internal/commands/patterns/`
- **Purpose**: Reusable command patterns
- **Files**:
  - `stdio.go` -- StdioHandler for stdin/stdout pipeline support
    - Evidence: Lines 11-92 define Setup/Finalize/Cleanup lifecycle
  - `doc.go` -- Package documentation
  - `stdio_test.go` -- Tests

**Dependencies**: `internal/cli`, `internal/pdf`, `internal/ocr`, `internal/fileio`, `internal/pages`, `internal/output`, `internal/pdferrors`, `internal/progress`, `internal/cleanup`, `github.com/spf13/cobra`

#### PDF Domain: `internal/pdf/`

**Purpose**: PDF manipulation via pdfcpu and ledongthuc/pdf
- **Files**:
  - `pdf.go` -- Shared utilities: `NewConfig()`, `pagesToStrings()`
    - Evidence: Lines 11-31 define pdfcpu config creation with password
  - `metadata.go` -- Info, page count, metadata get/set
    - Evidence: Lines 13-27 define `Info` struct with file path, size, pages, version, encryption, metadata fields
    - Evidence: Lines 29-68 define `GetInfo()` using pdfcpu API
  - `transform.go` -- Merge, split, extract, rotate, compress, extract images, create PDF from images
    - Evidence: Lines 18-74 define `MergeWithProgress()` with incremental merge for large file counts
    - Evidence: Lines 76-138 define `SplitWithProgress()` with progress bar
  - `text.go` -- Text extraction with dual library strategy (ledongthuc/pdf primary, pdfcpu fallback)
    - Evidence: Lines 34-42 show primary/fallback pattern
    - Evidence: Lines 125-173 show parallel text extraction for large page counts
  - `encryption.go` -- Encrypt, decrypt
  - `watermark.go` -- Watermarking
  - `validation.go` -- PDF/A validation
  - Test files: `pdf_test.go`, `metadata_test.go`, `encrypt_test.go`, `text_test.go`, `images_test.go`, `content_parsing_test.go`

**Exports**:
- Types: `Info`, `Metadata`
- Functions: `GetInfo()`, `PageCount()`, `GetMetadata()`, `SetMetadata()`, `Merge()`, `Split()`, `ExtractPages()`, `Rotate()`, `Compress()`, `Encrypt()`, `Decrypt()`, `ExtractText()`, `ExtractImages()`, `CreatePDFFromImages()`, `AddWatermark()`, `ValidatePDFA()`

**Dependencies**: `github.com/pdfcpu/pdfcpu/pkg/api`, `github.com/ledongthuc/pdf`, `internal/cleanup`, `internal/config`, `internal/fileio`, `internal/progress`, `context`

#### OCR Domain: `internal/ocr/`

**Purpose**: OCR text extraction with dual backend support (native Tesseract + WASM)
- **Files**:
  - `ocr.go` -- Engine with backend selection, tessdata download, image processing (sequential + parallel)
    - Evidence: Lines 50-118 define `Engine` struct and `NewEngine()` constructor
    - Evidence: Lines 120-137 define `selectBackend()` with auto/native/wasm selection
    - Evidence: Lines 208-323 define `downloadTessdata()` with retry, checksum verification, progress bar
    - Evidence: Lines 460-518 define parallel image processing with semaphore-limited workers
  - `backend.go` -- Backend interface definition
    - Evidence: Lines 8-13 define `Backend` interface with `Name()`, `Available()`, `ProcessImage()`, `Close()`
  - `native.go` -- Native Tesseract backend (exec.Command wrapper)
  - `wasm.go` -- WASM Tesseract backend (gogosseract wrapper)
  - `detect.go` -- Native Tesseract detection (checks PATH for `tesseract` binary)
  - `checksums.go` -- SHA256 checksums for tessdata files (supply chain security)
  - Test files: `ocr_test.go`, `backend_test.go`, `native_test.go`, `wasm_test.go`, `detect_test.go`, `engine_test.go`, `engine_extended_test.go`, `process_test.go`, `checksums_test.go`, `filesystem_test.go`, `mock_test.go`

**Exports**:
- Types: `Engine`, `Backend` interface, `EngineOptions`, `BackendType` enum
- Functions: `NewEngine()`, `NewEngineWithOptions()`, `ExtractTextFromPDF()`

**Dependencies**: `github.com/danlock/gogosseract`, `github.com/pdfcpu/pdfcpu/pkg/api`, `internal/cleanup`, `internal/fileio`, `internal/pdf`, `internal/progress`, `internal/retry`, `context`, `net/http`, `crypto/sha256`

#### File I/O Utilities: `internal/fileio/`

**Purpose**: File operations, stdin/stdout, path sanitization, validation
- **Files**:
  - `files.go` -- File validation (`ValidatePDFFile`), path sanitization (`SanitizePath`), size formatting, atomic writes
  - `stdio.go` -- Stdin/stdout utilities
    - Evidence: Lines 15-83 define `IsStdinInput()`, `ReadFromStdin()`, `WriteToStdout()`, `ResolveInputPath()`
    - stdin reads to temp file (pdfcpu requires file path, not stream)
  - `doc.go` -- Package documentation
  - Test files: `files_test.go`, `stdio_test.go`

**Exports**:
- Constants: `StdinIndicator = "-"`
- Functions: `ValidatePDFFile()`, `SanitizePath()`, `SanitizePaths()`, `IsStdinInput()`, `ReadFromStdin()`, `WriteToStdout()`, `ResolveInputPath()`, `FileExists()`, `GetFileSize()`, `FormatFileSize()`, `GenerateOutputFilename()`, `AtomicWrite()`, `CopyFile()`, `IsImageFile()`

**Dependencies**: `golang.org/x/term`, `internal/cleanup`, minimal external dependencies

#### Page Utilities: `internal/pages/`

**Purpose**: Page range parsing and validation
- **Files**:
  - `parser.go` -- Parse "1-5,7,10-12" format
    - Evidence: Lines 16-155 define `ParsePageRanges()`, `ExpandPageRanges()`, `ParseAndExpandPages()`, `FormatPageRanges()`
  - `validator.go` -- Validate page numbers against PDF page count
  - `reorder.go` -- Parse reorder sequences with "end" keyword support
  - `doc.go` -- Package documentation
  - `pages_test.go` -- Tests

**Exports**:
- Types: `PageRange` struct
- Functions: `ParsePageRanges()`, `ExpandPageRanges()`, `ParseAndExpandPages()`, `FormatPageRanges()`, `ValidatePageNumbers()`, `ParseReorderSequence()`

**Dependencies**: None (pure Go standard library)

#### Output Formatting: `internal/output/`

**Purpose**: Structured output (JSON, CSV, TSV) and table formatting
- **Files**:
  - `formatter.go` -- OutputFormatter with format detection
    - Evidence: Lines 21-86 define `OutputFormat` enum, `ParseOutputFormat()`, `OutputFormatter` with `Print()`, `PrintTable()`, `IsStructured()`
  - `table.go` -- Table rendering utilities (human-readable format)
  - `doc.go` -- Package documentation
  - `output_test.go` -- Tests

**Exports**:
- Types: `OutputFormat` enum (FormatHuman, FormatJSON, FormatCSV, FormatTSV), `OutputFormatter`
- Functions: `NewOutputFormatter()`, `ParseOutputFormat()`

**Dependencies**: Standard library `encoding/json`, `io`, `os`

#### Error Handling: `internal/pdferrors/`

**Purpose**: Context-aware PDF error wrapping
- **Files**:
  - `errors.go` -- PDFError type with operation, file, cause, hint
    - Evidence: Lines 9-97 define `PDFError` struct, `NewPDFError()`, `WrapError()` with pattern matching for common errors
    - Detects encryption, file not found, corruption and adds helpful hints
  - `doc.go` -- Package documentation
  - `errors_test.go` -- Tests

**Exports**:
- Types: `PDFError` struct
- Functions: `NewPDFError()`, `WrapError()`, `FormatError()`, `IsFileNotFound()`, `IsPasswordRequired()`
- Variables: `ErrFileNotFound`, `ErrNotPDF`, `ErrInvalidPages`, `ErrPasswordRequired`, `ErrWrongPassword`, `ErrCorruptPDF`, `ErrOutputExists`

**Dependencies**: Standard library `errors`, `fmt`, `strings`

#### Progress Bars: `internal/progress/`

**Purpose**: Progress bar utilities
- **Files**:
  - `progress.go` -- Progress bar creation with consistent styling
  - `doc.go` -- Package documentation
  - `progress_test.go` -- Tests

**Exports**:
- Functions: `NewProgressBar()`, `NewBytesProgressBar()`, `FinishProgressBar()`

**Dependencies**: `github.com/schollz/progressbar/v3`

#### Cleanup Registry: `internal/cleanup/`

**Purpose**: Signal-based temp file cleanup
- **Files**:
  - `cleanup.go` -- Thread-safe temp file registry
    - Evidence: Lines 10-68 define mutex-protected path slice, `Register()`, `Run()`, `Reset()`
    - `Register()` returns unregister function
    - `Run()` is idempotent, removes files in LIFO order
  - `cleanup_test.go` -- Tests

**Exports**:
- Functions: `Register(path string) func()`, `Run() error`, `Reset()` (test-only)

**Dependencies**: Standard library `os`, `sync`

#### Configuration: `internal/config/`

**Purpose**: YAML config loading, environment variable overrides
- **Files**:
  - `config.go` -- Config struct, loader, singleton
    - Evidence: Lines 14-220 define nested config structs, `DefaultConfig()`, `Load()`, `Save()`, `Get()` singleton
    - Evidence: Lines 53-69 define `DefaultPerformanceConfig()` with CPU-adaptive defaults
    - Evidence: Lines 192-213 define thread-safe singleton with double-checked locking
  - `doc.go` -- Package documentation
  - `config_test.go` -- Tests

**Exports**:
- Types: `Config`, `DefaultsConfig`, `CompressConfig`, `EncryptConfig`, `OCRConfig`, `PerformanceConfig`
- Functions: `DefaultConfig()`, `DefaultPerformanceConfig()`, `ConfigPath()`, `Load()`, `Save()`, `Get()`, `Reset()`

**Dependencies**: `gopkg.in/yaml.v3`, `internal/fileio`, `sync`, `runtime`

#### Logging: `internal/logging/`

**Purpose**: Structured logging with slog
- **Files**:
  - `logger.go` -- Logger initialization, singleton, level/format management
  - `doc.go` -- Package documentation
  - `logger_test.go` -- Tests

**Exports**:
- Functions: `Init()`, `Get()`, `SetLevel()`, `SetFormat()`

**Dependencies**: Standard library `log/slog`, `sync`

#### Retry Logic: `internal/retry/`

**Purpose**: Exponential backoff retry
- **Files**: (inferred from references in ocr.go)
  - `retry.go` -- Generic retry with exponential backoff (likely)
  - `errors.go` -- PermanentError type for non-retryable errors (likely)

**Exports** (inferred):
- Types: `Options`, `PermanentError`
- Functions: `Do()`, `Permanent()`

**Dependencies**: Standard library `context`, `time`

#### Test Infrastructure: `internal/testing/`

**Purpose**: Shared test utilities and mocks
- **Files**:
  - `fixtures.go` -- Test fixture helpers
  - `mock_pdf.go` -- Mock PDF operations
  - `mock_ocr.go` -- Mock OCR backend
  - `doc.go` -- Package documentation

**Exports**:
- Mock implementations for testing

**Dependencies**: Test-only, not used in production code

### Test Data: `testdata/`

**Path**: `/Users/lgbarn/Personal/pdf-cli/testdata/`
- **Purpose**: Test PDF files for integration tests
- **Files**: Sample PDFs with various characteristics (encrypted, multi-page, etc.)
- **Usage**: Referenced by tests in `internal/pdf/*_test.go`, `internal/commands/*_test.go`

### Documentation: `docs/`

**Path**: `/Users/lgbarn/Personal/pdf-cli/docs/`
- **Purpose**: Architecture and planning documentation
- **Files**:
  - `architecture.md` -- Existing architecture documentation (source of truth for this analysis)
    - Evidence: Lines 1-209 document package structure, dependency graph, design decisions
  - `plans/` subdirectory -- Planning documents (inferred from directory listing)

### CI/CD: `.github/workflows/`

**Path**: `/Users/lgbarn/Personal/pdf-cli/.github/workflows/`
- **Purpose**: GitHub Actions CI/CD pipelines
- **Files** (inferred):
  - `ci.yaml` -- Continuous integration (tests, linting, coverage)
  - Build/release workflows (likely)

### Scripts: `scripts/`

**Path**: `/Users/lgbarn/Personal/pdf-cli/scripts/`
- **Purpose**: Build and maintenance automation
- **Files** (inferred from directory listing): Build scripts, release scripts, etc.

### Build and Configuration Files (Root)

- **`go.mod`**: Go module definition
  - Evidence: Lines 1-36 define module name `github.com/lgbarn/pdf-cli`, Go version `1.25`, dependencies
  - Direct dependencies: pdfcpu, ledongthuc/pdf, gogosseract, progressbar, cobra, yaml.v3, term

- **`go.sum`**: Dependency checksums for reproducible builds

- **`Makefile`**: Build automation
  - Evidence: Referenced in README.md (lines 651-671) with targets: `build`, `test`, `test-coverage`, `build-all`, `clean`, `lint`, `coverage-check`, `check-all`

- **`README.md`**: User-facing documentation
  - Evidence: Lines 1-841 document all 14 commands, installation, usage examples, configuration, troubleshooting

- **`CHANGELOG.md`**: Release history
  - Evidence: Referenced in git commits (v2.0.0 release)

- **`CONTRIBUTING.md`**: Contribution guidelines

- **`LICENSE`**: MIT license

- **`SECURITY.md`**: Security policy

- **`.golangci.yaml`**: golangci-lint configuration

- **`.goreleaser.yaml`**: GoReleaser configuration for multi-platform builds

- **`.pre-commit-config.yaml`**: Pre-commit hook configuration

- **`.gitignore`**: Git ignore rules

### Hidden/Build Artifacts (Not Source Controlled)

- **`.serena/`**: Local tooling artifacts (6 files)
- **`.shipyard/`**: Codebase analysis output (this document's target)
- **`cover*.out`**: Coverage report files (5 files)
- **`pdf` binary**: Compiled executable (34.5 MB)

### Module Boundaries Summary

**Public vs Internal**:
- **Public API**: None -- `internal/` enforces private packages
- **Entry point**: `cmd/pdf` is the only public surface area (builds to `pdf` binary)
- **Internal encapsulation**: All implementation details in `internal/`, not importable by external projects

**Layer Dependencies** (bottom-up):
1. **Leaf utilities**: fileio, pages, output, pdferrors, progress, cleanup, retry (no internal dependencies)
2. **Infrastructure**: config (depends on fileio), logging (no internal dependencies)
3. **Domain**: pdf (depends on cleanup, config, fileio, progress), ocr (depends on cleanup, fileio, pdf, progress, retry)
4. **Orchestration**: commands (depends on cli, pdf, ocr, fileio, pages, output, pdferrors, progress, cleanup)
5. **Framework**: cli (depends on config, logging)
6. **Entry**: cmd/pdf (depends on cli, cleanup, commands)

**Shared Code Locations**:
- **Command utilities**: `/Users/lgbarn/Personal/pdf-cli/internal/commands/helpers.go`
- **Command patterns**: `/Users/lgbarn/Personal/pdf-cli/internal/commands/patterns/stdio.go`
- **Test utilities**: `/Users/lgbarn/Personal/pdf-cli/internal/testing/`

**Configuration Hierarchy**:
- **Config file**: `~/.config/pdf-cli/config.yaml` or `$XDG_CONFIG_HOME/pdf-cli/config.yaml`
- **Tessdata cache**: `~/.config/pdf-cli/tessdata/` (for WASM OCR)
- **Environment variables**: `PDF_CLI_*` prefix overrides config file
- **CLI flags**: Override environment variables (except password, which has special 4-tier priority)

## Summary Table

| Component | Location | Files | Purpose | Confidence |
|-----------|----------|-------|---------|------------|
| Entry Point | `/Users/lgbarn/Personal/pdf-cli/cmd/pdf/` | 1 | Main binary, signal handling, cleanup | Observed |
| CLI Framework | `/Users/lgbarn/Personal/pdf-cli/internal/cli/` | 8 | Cobra wrapper, flags, password, completion | Observed |
| Commands | `/Users/lgbarn/Personal/pdf-cli/internal/commands/` | 28 | 14 commands + helpers + tests | Observed |
| Command Patterns | `/Users/lgbarn/Personal/pdf-cli/internal/commands/patterns/` | 3 | StdioHandler for pipelines | Observed |
| PDF Domain | `/Users/lgbarn/Personal/pdf-cli/internal/pdf/` | 13 | PDF operations via pdfcpu + ledongthuc/pdf | Observed |
| OCR Domain | `/Users/lgbarn/Personal/pdf-cli/internal/ocr/` | 15+ | Dual backend OCR (native + WASM) | Observed |
| File I/O | `/Users/lgbarn/Personal/pdf-cli/internal/fileio/` | 5 | Stdin/stdout, validation, sanitization | Observed |
| Page Utilities | `/Users/lgbarn/Personal/pdf-cli/internal/pages/` | 4 | Page range parsing and validation | Observed |
| Output Formatting | `/Users/lgbarn/Personal/pdf-cli/internal/output/` | 4 | JSON/CSV/TSV/human formatting | Observed |
| Error Handling | `/Users/lgbarn/Personal/pdf-cli/internal/pdferrors/` | 3 | Context-aware PDF errors | Observed |
| Progress Bars | `/Users/lgbarn/Personal/pdf-cli/internal/progress/` | 3 | Progress bar utilities | Observed |
| Cleanup Registry | `/Users/lgbarn/Personal/pdf-cli/internal/cleanup/` | 2 | Signal-based temp file cleanup | Observed |
| Configuration | `/Users/lgbarn/Personal/pdf-cli/internal/config/` | 3 | YAML config + env vars, singleton | Observed |
| Logging | `/Users/lgbarn/Personal/pdf-cli/internal/logging/` | 3 | Structured logging with slog | Observed |
| Retry Logic | `/Users/lgbarn/Personal/pdf-cli/internal/retry/` | ? | Exponential backoff | Inferred |
| Test Infrastructure | `/Users/lgbarn/Personal/pdf-cli/internal/testing/` | 4 | Mocks and fixtures | Observed |
| Test Data | `/Users/lgbarn/Personal/pdf-cli/testdata/` | ? | Sample PDFs | Observed |
| Documentation | `/Users/lgbarn/Personal/pdf-cli/docs/` | 2+ | Architecture docs | Observed |
| CI/CD | `/Users/lgbarn/Personal/pdf-cli/.github/workflows/` | ? | GitHub Actions | Observed |
| Scripts | `/Users/lgbarn/Personal/pdf-cli/scripts/` | ? | Build automation | Observed |
| Config Files | Root | 10 | Build, lint, git, release configs | Observed |

## Open Questions

- **Retry package details**: No direct file listing for `internal/retry/`, but heavily referenced in ocr.go. Location and API inferred but not verified by reading source files.
- **Scripts directory contents**: Not examined in detail. Likely contains release automation, version bumping, etc.
- **CI workflow details**: Workflow files not examined. Likely includes test, lint, coverage, security scanning, release builds.
- **Testdata contents**: Directory exists but individual test PDF characteristics not documented.
- **Hidden directories**: `.serena/` purpose unclear (6 files). Local tooling artifact, not relevant to architecture but may indicate development workflow.
