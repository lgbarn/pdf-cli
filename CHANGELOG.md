# Changelog

All notable changes to pdf-cli are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.0.0] - 2026-01-31

### Breaking Changes
- **Secure password input**: Passwords for `encrypt`/`decrypt` commands are no longer accepted
  via `--password` CLI flag (visible in `ps aux` and shell history). Use stdin prompt, `PDF_CLI_PASSWORD`
  environment variable, or `--password-file` instead.

### Security
- **Path traversal protection**: All file path inputs sanitized against directory traversal attacks (R3)
- **Download integrity verification**: SHA256 checksum verification for tessdata downloads (R2)
- **Password exposure eliminated**: Passwords no longer visible in process listings (R1)

### Added
- **Retry logic with exponential backoff** for tessdata network downloads (R12)
- **Signal-based temp file cleanup** on crash/interrupt via signal handlers (R11)
- **Portable coverage checking**: Go-based coverage script replaces `bc`/`awk` dependency (R16)
- **Configurable parallelism thresholds**: Adaptive to system resources (R17)

### Fixed
- **Thread-safe globals**: `config.Get()` and `logging.Get()` use `sync.Once`; safe under concurrent access (R4)
- **Context propagation**: All long-running operations accept `context.Context` for cancellation (R5)
- **Parallel error collection**: All errors surfaced from parallel processing, no silent drops (R6)
- **File close errors**: Checked and propagated, especially for write operations (R8)

### Changed
- **Go 1.25** required (up from 1.22)
- **All 21 dependencies updated** to latest compatible versions (R7)
- **Magic numbers replaced** with named constants throughout codebase (R14)
- **Logging consolidated** to `log/slog` (R15)
- **Documentation aligned**: README and `docs/architecture.md` reflect current code (R13)
- **Large test files split** into focused files under 500 lines (R10)
- **CI**: gosec updated to v2.22.11 for Go 1.24+ compatibility

## [1.5.0] - 2026-01-21

### Added
- **Configuration file support**: Optional YAML config at `~/.config/pdf-cli/config.yaml`
  - Set default values for verbose, force, progress flags
  - Configure OCR language and backend preferences
  - Environment variable overrides with `PDF_CLI_` prefix
- **Structured logging**: New `--log-level` and `--log-format` flags
  - Levels: `debug`, `info`, `warn`, `error`, `silent` (default)
  - Formats: `text` (default), `json`
  - Uses Go 1.21+ `log/slog` for structured output
- **Dry-run mode**: New `--dry-run` flag for all modifying commands
  - Preview operations without making changes
  - Shows what files would be created/modified
  - Useful for scripting and validation
- **CI coverage enforcement**: GitHub Actions fails if coverage drops below 75%

### Changed
- **Architecture refactoring**: Comprehensive reorganization for maintainability
  - Split monolithic `util/` package into focused packages:
    - `fileio/` - File operations and stdio utilities
    - `pages/` - Page range parsing and validation
    - `output/` - Output formatting (JSON, CSV, TSV)
    - `pdferrors/` - Error handling with context and hints
    - `progress/` - Progress bar utilities
  - Split `pdf.go` (689 LOC) into focused modules:
    - `metadata.go` - GetInfo, PageCount, GetMetadata, SetMetadata
    - `transform.go` - Merge, Split, Rotate, Compress, ExtractPages
    - `encryption.go` - Encrypt, Decrypt
    - `text.go` - ExtractText, ExtractTextWithProgress
    - `watermark.go` - AddWatermark, AddImageWatermark
    - `validation.go` - Validate, ValidatePDFA, ConvertToPDFA
  - Created `StdioHandler` pattern for consistent stdin/stdout handling
- **Test coverage improvements**:
  - Commands package: 60.7% → 82.8%
  - PDF package: 57.9% → 85.6%
  - Overall: 69.6% → 81.5%
  - All packages now meet 75%+ threshold
- **Enhanced linting**: Added 8 additional linters to `.golangci.yaml`
  - misspell, gocritic, revive, errcheck
  - Stricter code quality enforcement

### Documentation
- Added `docs/architecture.md` with dependency graphs and extension points
- Added `CONTRIBUTING.md` with development guidelines
- Updated README with new features and project structure

### Internal
- Added `internal/config/` for configuration file support
- Added `internal/logging/` for structured logging
- Added `internal/commands/patterns/` for reusable command patterns
- Added `internal/testing/` for mock infrastructure
- New Makefile targets: `lint-fix`, `check-all`, `coverage`, `coverage-check`

## [1.4.0] - 2026-01-20

### Added
- **Structured output formats**: New `--format` flag for machine-readable output
  - Supports `json`, `csv`, and `tsv` formats
  - Available on `info`, `meta`, and `pdfa validate` commands
  - Example: `pdf info document.pdf --format json | jq .pages`
- **stdin/stdout support**: Process PDFs from pipes (Unix philosophy)
  - Use `-` to read from stdin: `cat doc.pdf | pdf text -`
  - Use `--stdout` for binary output: `pdf compress input.pdf --stdout > out.pdf`
  - Supported on: text, info, compress, extract, rotate, reorder, encrypt, decrypt, pdfa convert
- **Comprehensive test coverage improvements**:
  - CLI: 98.0% coverage
  - Util: 90.7% coverage
  - OCR: 75.5% coverage (up from 9.3%)
  - Commands: 60.7% coverage
  - PDF: 57.9% coverage
  - Overall: 69.6% coverage
  - Comprehensive mock infrastructure for OCR testing
  - Table-driven test patterns for maintainability
  - Integration tests with real PDF files

### Changed
- **Code deduplication**: Extracted shared utilities to `internal/util/`
  - `util.IsImageFile()` - Shared image file detection
  - `util.NewProgressBar()` - Unified progress bar creation
  - `util.FinishProgressBar()` - Consistent progress bar cleanup
- Exported `pdf.NewConfig()` for use across packages

### Internal
- Added `internal/util/images.go` for shared image utilities
- Added `internal/util/progress.go` for shared progress bar utilities
- Added `internal/util/output.go` for structured output formatting
- Added `internal/util/stdio.go` for stdin/stdout handling
- Added format flag helpers to `internal/cli/flags.go`

## [1.3.2] - 2025-01-20

### Added
- **`combine-images` command**: Create PDFs from multiple images
  - Supports PNG, JPEG, and TIFF formats
  - Optional `--page-size` flag (A4, Letter, or auto-fit to image)
  - Each image becomes one page in the output PDF
- Unit tests for OCR package helper functions
  - `parseLanguages`, `primaryLanguage`, `isImageFile`, `joinNonEmpty`
  - `ParseBackendType`, `BackendType.String()`

### Changed
- Code simplification using Go 1.21+ `slices.Equal` in tests
- Improved naming consistency (`isImageFile` function)

## [1.3.1] - 2025-01-19

### Added
- Parallel OCR processing for native Tesseract backend (>5 images)
- Comprehensive unit tests for `ParseReorderSequence` function
- Unit tests for `reorder` and `pdfa` command structure

### Changed
- Improved error messages for page sequence parsing with helpful hints
- Code simplification and cleanup across OCR and test modules

### Documentation
- Added PDF/A limitations callout in README
- Updated parallel processing documentation to include OCR

## [1.3.0] - 2025-01-18

### Added
- **Native OCR support**: Auto-detects system Tesseract installation
- **`reorder` command**: Reorder, reverse, or duplicate PDF pages
  - Supports `end` keyword for last page reference
  - Reverse ranges (e.g., `5-1` for descending order)
  - Page duplication (e.g., `1,2,3,1` repeats page 1)
- **`pdfa` command**: PDF/A validation and conversion
  - `pdfa validate` - Check PDF/A compliance
  - `pdfa convert` - Optimize PDFs toward PDF/A format
- **Batch mode**: Process multiple files with pattern expansion
  - Commands supporting batch: info, rotate, compress, encrypt, decrypt, meta, watermark

### Changed
- OCR backend selection: `auto` (default), `native`, or `wasm`
- Updated README with new command documentation

## [1.2.0] - 2025-01-17

### Added
- **Batch mode** for processing multiple PDFs
- **Progress bars** for long-running operations (`--progress` flag)
- **OCR support** via WASM Tesseract (no external dependencies)
  - Multi-language support (`--ocr-lang eng+fra`)
  - Automatic tessdata download on first use

### Fixed
- Security findings from Gosec static analysis

### Changed
- Parallel processing documentation improvements

## [1.1.0] - 2025-01-16

### Added
- **Parallel processing** for improved performance
  - File validation when merging >3 files
  - Text extraction when processing >5 pages
- Comprehensive test suite with helpers

### Fixed
- Removed Homebrew tap (simplified distribution)
- GoReleaser deprecation warnings

## [1.0.0] - 2025-01-15

### Added
- Initial release of pdf-cli
- **Core commands**:
  - `info` - Display PDF information
  - `merge` - Combine multiple PDFs
  - `split` - Split PDF into pages or chunks
  - `extract` - Extract specific pages
  - `rotate` - Rotate pages
  - `compress` - Optimize file size
  - `encrypt` - Add password protection
  - `decrypt` - Remove password protection
  - `text` - Extract text content
  - `images` - Extract embedded images
  - `meta` - View/modify metadata
  - `watermark` - Add text or image watermarks
- Cross-platform binaries (Linux, macOS, Windows)
- Shell completion (bash, zsh, fish, PowerShell)
- CI/CD pipeline with GitHub Actions

### Security
- Addressed all Gosec static analysis findings
- Secure handling of encrypted PDFs

[2.0.0]: https://github.com/lgbarn/pdf-cli/compare/v1.4.0...v2.0.0
[1.5.0]: https://github.com/lgbarn/pdf-cli/compare/v1.4.0...v1.5.0
[1.4.0]: https://github.com/lgbarn/pdf-cli/compare/v1.3.2...v1.4.0
[1.3.2]: https://github.com/lgbarn/pdf-cli/compare/v1.3.1...v1.3.2
[1.3.1]: https://github.com/lgbarn/pdf-cli/compare/v1.3.0...v1.3.1
[1.3.0]: https://github.com/lgbarn/pdf-cli/compare/v1.2.0...v1.3.0
[1.2.0]: https://github.com/lgbarn/pdf-cli/compare/v1.1.0...v1.2.0
[1.1.0]: https://github.com/lgbarn/pdf-cli/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/lgbarn/pdf-cli/releases/tag/v1.0.0
