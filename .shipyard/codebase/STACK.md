# Technology Stack

## Overview
pdf-cli is a single-binary Go CLI application for PDF manipulation operations. Built on Go 1.25, it leverages native Go libraries for PDF processing, OCR capabilities (with dual backend support), and cross-platform compilation. The project follows modern Go practices with comprehensive tooling for build, test, lint, and release automation.

## Findings

### Primary Language

- **Language**: Go 1.25
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/go.mod` (line 3)
  - Module path: `github.com/lgbarn/pdf-cli`
  - Supports cross-compilation for Linux, macOS (amd64/arm64), and Windows (amd64)

### Core PDF Libraries

- **pdfcpu v0.11.1** -- Primary PDF manipulation library
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/go.mod` (line 8)
  - Purpose: PDF operations including merge, split, encrypt, compress, watermark, metadata
  - Used throughout `/Users/lgbarn/Personal/pdf-cli/internal/pdf/` package
  - Configuration abstraction: `/Users/lgbarn/Personal/pdf-cli/internal/pdf/pdf.go`

- **ledongthuc/pdf v0.0.0-20250511090121-5959a4027728** -- Text extraction library
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/go.mod` (line 7)
  - Purpose: Primary text extraction from PDFs (better quality than pdfcpu fallback)
  - Used in: `/Users/lgbarn/Personal/pdf-cli/internal/pdf/text.go` (lines 11, 46)
  - Supports parallel text extraction for large documents

### OCR Dependencies

- **danlock/gogosseract v0.0.11-0ad3421** -- WASM-based Tesseract OCR
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/go.mod` (line 6)
  - Purpose: Built-in OCR fallback when native Tesseract is unavailable
  - Implementation: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/wasm.go` (lines 10, 84-90)
  - Requires downloading tessdata files from GitHub on first use
  - Uses wazero WebAssembly runtime (see below)

- **tetratelabs/wazero v1.11.0** -- WebAssembly runtime
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/go.mod` (line 29)
  - Purpose: Runtime for WASM-based Tesseract OCR engine
  - Transitive dependency via gogosseract

- **jerbob92/wazero-emscripten-embind v1.5.2** -- Emscripten bindings
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/go.mod` (line 23)
  - Purpose: Emscripten-to-WASM bridge for Tesseract
  - Transitive dependency via gogosseract

- **Native Tesseract** -- Optional system-installed OCR
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/detect.go` (lines 24-42)
  - Detection via `exec.LookPath("tesseract")`
  - Preferred backend when available; graceful fallback to WASM
  - Backend selection: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr.go` (lines 120-138)

### CLI Framework

- **spf13/cobra v1.10.2** -- Command-line interface framework
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/go.mod` (line 10)
  - Usage: `/Users/lgbarn/Personal/pdf-cli/internal/cli/cli.go` (line 9)
  - Provides root command structure, subcommand registration, flag parsing
  - Supports shell completion generation (bash, zsh, fish)

- **spf13/pflag v1.0.10** -- POSIX-compliant flag parsing
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/go.mod` (line 28)
  - Transitive dependency of cobra

### User Interface Libraries

- **schollz/progressbar/v3 v3.19.0** -- Progress bar display
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/go.mod` (line 9)
  - Usage: `/Users/lgbarn/Personal/pdf-cli/internal/progress/progress.go`
  - Provides progress tracking for long-running operations (OCR, text extraction, downloads)
  - Used in: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr.go` (line 279)

- **mitchellh/colorstring v0.0.0-20190213212951-d06e56a500db** -- Terminal color support
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/go.mod` (line 25)
  - Transitive dependency via progressbar

- **mattn/go-runewidth v0.0.19** -- Unicode display width calculation
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/go.mod` (line 24)
  - Purpose: Correct terminal width calculation for progress bars

- **rivo/uniseg v0.4.7** -- Unicode text segmentation
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/go.mod` (line 27)
  - Transitive dependency for Unicode handling

### Configuration Management

- **gopkg.in/yaml.v3 v3.0.1** -- YAML parsing
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/go.mod` (line 12)
  - Usage: `/Users/lgbarn/Personal/pdf-cli/internal/config/config.go` (line 11)
  - Purpose: Parse user configuration file at `~/.config/pdf-cli/config.yaml`
  - Supports defaults for output format, OCR language, compression quality, performance tuning

### Standard Library Extensions

- **golang.org/x/term v0.39.0** -- Terminal interaction
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/go.mod` (line 11)
  - Purpose: Password input from terminal (no echo)
  - Used for secure password prompts

- **golang.org/x/crypto v0.47.0** -- Cryptographic functions
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/go.mod` (line 30)
  - Purpose: PDF encryption/decryption operations
  - Transitive dependency via pdfcpu

- **golang.org/x/image v0.35.0** -- Image processing
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/go.mod` (line 32)
  - Purpose: Image extraction and manipulation from PDFs
  - Transitive dependency via pdfcpu

- **golang.org/x/text v0.33.0** -- Text processing and encoding
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/go.mod` (line 34)
  - Purpose: Character encoding and internationalization
  - Transitive dependency

- **golang.org/x/sys v0.40.0** -- System calls
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/go.mod` (line 33)
  - Purpose: Low-level OS interaction
  - Transitive dependency

- **golang.org/x/exp v0.0.0-20260112195511-716be5621a96** -- Experimental packages
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/go.mod` (line 31)
  - Transitive dependency

### Supporting Libraries

- **pkg/errors v0.9.1** -- Error handling with stack traces
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/go.mod` (line 26)
  - Transitive dependency via pdfcpu

- **hhrutter/lzw v1.0.0** -- LZW compression
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/go.mod` (line 19)
  - Purpose: PDF compression algorithm support
  - Transitive dependency via pdfcpu

- **hhrutter/tiff v1.0.2** -- TIFF image handling
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/go.mod` (line 21)
  - Purpose: TIFF extraction/embedding in PDFs
  - Transitive dependency via pdfcpu

- **hhrutter/pkcs7 v0.2.0** -- PKCS#7 cryptographic message syntax
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/go.mod` (line 20)
  - Purpose: PDF digital signatures
  - Transitive dependency via pdfcpu

- **inconshreveable/mousetrap v1.1.0** -- Windows CLI behavior
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/go.mod` (line 22)
  - Purpose: Windows-specific command execution
  - Transitive dependency via cobra

- **clipperhouse/stringish v0.1.1** and **clipperhouse/uax29/v2 v2.4.0** -- Text tokenization
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/go.mod` (lines 16-17)
  - Purpose: Unicode text segmentation for OCR
  - Transitive dependencies via gogosseract

- **danlock/pkg v0.0.46-2e8eb6d** -- Utility library
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/go.mod` (line 18)
  - Transitive dependency via gogosseract

### Build Tools

- **make** -- Build automation
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/Makefile` (all lines)
  - Targets: build, test, lint, coverage, cross-compilation
  - Version information injected via ldflags

- **GoReleaser v2** -- Release automation
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/.goreleaser.yaml` (line 4)
  - Purpose: Multi-platform binary builds, archives, checksums, GitHub releases
  - Supports deb/rpm package generation (lines 77-88)
  - Configuration: `/Users/lgbarn/Personal/pdf-cli/.goreleaser.yaml`

- **golangci-lint v2.8.0** -- Linting aggregator
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/.github/workflows/ci.yaml` (line 29)
  - Configuration: `/Users/lgbarn/Personal/pdf-cli/.golangci.yaml`
  - Enabled linters: govet, ineffassign, staticcheck, unused, misspell, gocritic, revive, errcheck
  - Timeout: 5 minutes

- **gosec v2.22.11** -- Security scanner
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/.github/workflows/ci.yaml` (line 114)
  - Purpose: Static security analysis for Go code
  - Updated for Go 1.24+ compatibility

- **pre-commit v5.0.0** -- Git hook framework
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/.pre-commit-config.yaml` (line 3)
  - Hooks: trailing-whitespace, end-of-file-fixer, check-yaml, check-merge-conflict
  - Custom Go hooks: fmt, vet, mod tidy, build, test, golangci-lint

### Testing Framework

- **Go standard testing** -- Unit and integration tests
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/backend_test.go` (grep results)
  - Test files throughout codebase (`*_test.go`)
  - Coverage threshold: 75% (enforced in CI)
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/.github/workflows/ci.yaml` (line 54)

- **Coverage tool** -- Custom coverage checker
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/scripts/coverage-check.go`
  - Purpose: Parse and validate test coverage against threshold
  - Runs in CI pipeline

### CI/CD

- **GitHub Actions** -- Continuous integration
  - Workflows: `/Users/lgbarn/Personal/pdf-cli/.github/workflows/ci.yaml`, `/Users/lgbarn/Personal/pdf-cli/.github/workflows/release.yaml`
  - CI jobs: lint, test (with race detection and coverage), build (matrix for 5 OS/arch combinations), security scan
  - Go version: Read from `go.mod` (line 22-23 in ci.yaml)
  - Actions versions: actions/checkout@v4, actions/setup-go@v5, actions/upload-artifact@v4
  - Coverage upload: codecov/codecov-action@v4 (line 57)

- **Dependabot v2** -- Dependency updates
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/.github/dependabot.yaml`
  - Monitors: Go modules (weekly), GitHub Actions (weekly)
  - Pull request limit: 5 per ecosystem

### Build Scripts

- **build.sh** -- Cross-compilation script
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/scripts/build.sh`
  - Platforms: linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64
  - Sets ldflags for version, commit, date injection

## Summary Table

| Category | Item | Version | Source | Confidence |
|----------|------|---------|--------|------------|
| Language | Go | 1.25 | go.mod | Observed |
| PDF Core | pdfcpu | v0.11.1 | go.mod | Observed |
| PDF Text | ledongthuc/pdf | 2025-05-11 snapshot | go.mod | Observed |
| OCR WASM | gogosseract | v0.0.11-0ad3421 | go.mod | Observed |
| OCR Native | Tesseract | system-dependent | detect.go | Observed |
| CLI Framework | cobra | v1.10.2 | go.mod | Observed |
| Progress UI | progressbar | v3.19.0 | go.mod | Observed |
| Config | yaml.v3 | v3.0.1 | go.mod | Observed |
| Terminal | golang.org/x/term | v0.39.0 | go.mod | Observed |
| WASM Runtime | wazero | v1.11.0 | go.mod | Observed |
| Linter | golangci-lint | v2.8.0 | ci.yaml | Observed |
| Security | gosec | v2.22.11 | ci.yaml | Observed |
| Release | GoReleaser | v2 | release.yaml | Observed |
| Pre-commit | pre-commit-hooks | v5.0.0 | pre-commit-config.yaml | Observed |
| CI Platform | GitHub Actions | N/A | .github/workflows/ | Observed |
| Dependency Bot | Dependabot | v2 | dependabot.yaml | Observed |

## Version Management

- **Version injection**: Build-time ldflags set `main.version`, `main.commit`, `main.date`
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/Makefile` (line 8)
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/cmd/pdf/main.go` (lines 15-18)
- **Git tags**: Version derived from `git describe --tags` when available
- **Release workflow**: Triggered on `v*` tags
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/.github/workflows/release.yaml` (lines 5-6)

## Runtime Requirements

- **No external dependencies** for core functionality (merge, split, encrypt, etc.)
- **Optional Tesseract** for native OCR backend (falls back to WASM)
  - Detection: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/detect.go`
  - Common paths: Homebrew (/opt/homebrew), apt (/usr/share/tesseract-ocr), Windows Program Files
- **Internet access** required only for tessdata downloads (first-time OCR use)
  - URL: `https://github.com/tesseract-ocr/tessdata_fast/raw/main`
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr.go` (line 29)

## Build Environment

- **CGO**: Disabled (CGO_ENABLED=0) for static binary compilation
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/.goreleaser.yaml` (line 18)
- **Module mode**: GO111MODULE=on (enforced in Makefile)
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/Makefile` (line 28)
- **Supported platforms**:
  - Linux: amd64, arm64
  - macOS: amd64, arm64
  - Windows: amd64 only (arm64 excluded)
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/.goreleaser.yaml` (lines 19-28)

## Open Questions

- Are there plans to support additional OCR engines beyond Tesseract?
- Is there a minimum Tesseract version requirement for the native backend?
- Are there any plans to add GPU acceleration for image/OCR processing?
- What is the deprecation timeline for Go 1.24 support (currently 1.25 minimum)?
