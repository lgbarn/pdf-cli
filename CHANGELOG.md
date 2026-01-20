# Changelog

All notable changes to pdf-cli are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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

[1.3.1]: https://github.com/lgbarn/pdf-cli/compare/v1.3.0...v1.3.1
[1.3.0]: https://github.com/lgbarn/pdf-cli/compare/v1.2.0...v1.3.0
[1.2.0]: https://github.com/lgbarn/pdf-cli/compare/v1.1.0...v1.2.0
[1.1.0]: https://github.com/lgbarn/pdf-cli/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/lgbarn/pdf-cli/releases/tag/v1.0.0
