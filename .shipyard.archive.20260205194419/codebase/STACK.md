# Technology Stack

## Language & Runtime

### Go 1.24.1
- **Specified in**: `/Users/lgbarn/Personal/pdf-cli/go.mod`
- **Module**: `github.com/lgbarn/pdf-cli`
- **Build Configuration**: `GO111MODULE=on` (enforced throughout build scripts and CI)
- **CGO**: Disabled (`CGO_ENABLED=0` in GoReleaser config for static binary compilation)
- **Target Platforms**:
  - Linux (amd64, arm64)
  - macOS/Darwin (amd64, arm64)
  - Windows (amd64)

## Core Dependencies

### PDF Processing Libraries

#### pdfcpu v0.11.1
- **Purpose**: Primary PDF manipulation library
- **Import**: `github.com/pdfcpu/pdfcpu/pkg/api` and subpackages
- **Capabilities**: Merge, split, extract, rotate, compress, encrypt/decrypt, watermark, metadata
- **Usage**: Used in `/Users/lgbarn/Personal/pdf-cli/internal/pdf/*.go`
- **Key Subpackages**:
  - `pkg/api` - High-level API functions
  - `pkg/pdfcpu` - Core PDF processing
  - `pkg/pdfcpu/model` - Configuration and data models
  - `pkg/pdfcpu/types` - Type definitions

#### ledongthuc/pdf v0.0.0-20250511090121-5959a4027728
- **Purpose**: Alternative/fallback PDF text extraction
- **Import**: `github.com/ledongthuc/pdf`
- **Usage**: Primary method for text extraction in `/Users/lgbarn/Personal/pdf-cli/internal/pdf/text.go`
- **Note**: Used as primary extractor with pdfcpu as fallback for better text quality

### OCR Libraries

#### gogosseract v0.0.11-0ad3421
- **Purpose**: WASM-based Tesseract OCR engine (no system dependencies)
- **Import**: `github.com/danlock/gogosseract`
- **Usage**: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/wasm.go`
- **Key Feature**: Embedded WASM Tesseract - works without native Tesseract installation
- **Dependencies**:
  - `github.com/danlock/pkg v0.0.17-a9828f2` (indirect)
  - `github.com/jerbob92/wazero-emscripten-embind v1.3.0` (indirect)
  - `github.com/tetratelabs/wazero v1.5.0` (WebAssembly runtime)

### CLI Framework

#### spf13/cobra v1.10.2
- **Purpose**: Command-line interface framework
- **Import**: `github.com/spf13/cobra`
- **Usage**: Root command and all subcommands in `/Users/lgbarn/Personal/pdf-cli/internal/cli/cli.go` and `/Users/lgbarn/Personal/pdf-cli/internal/commands/*.go`
- **Features Used**: Flags, subcommands, completion generation
- **Dependencies**:
  - `github.com/spf13/pflag v1.0.10` (POSIX/GNU-style flags)
  - `github.com/inconshreveable/mousetrap v1.1.0` (Windows support)

### UI/UX Libraries

#### progressbar v3.19.0
- **Purpose**: Terminal progress bars for long-running operations
- **Import**: `github.com/schollz/progressbar/v3`
- **Usage**: `/Users/lgbarn/Personal/pdf-cli/internal/progress/progress.go`
- **Operations**: Merging, extracting, OCR processing, text extraction
- **Dependencies**:
  - `github.com/mitchellh/colorstring v0.0.0-20190213212951-d06e56a500db`
  - `github.com/rivo/uniseg v0.4.7` (Unicode text segmentation)
  - `github.com/mattn/go-runewidth v0.0.19` (East Asian width)

### System Libraries

#### golang.org/x/term v0.39.0
- **Purpose**: Terminal state detection and handling
- **Usage**: Password input, terminal capabilities detection
- **Related**: `golang.org/x/sys v0.40.0` (system calls)

#### gopkg.in/yaml.v3 v3.0.1
- **Purpose**: Configuration file parsing
- **Usage**: `/Users/lgbarn/Personal/pdf-cli/internal/config/config.go`
- **Config Location**: `~/.config/pdf-cli/config.yaml` (XDG_CONFIG_HOME compliant)
- **Note**: Also uses `gopkg.in/yaml.v2 v2.4.0` (indirect dependency)

### Supporting Libraries

#### Image Processing
- `golang.org/x/image v0.32.0` - Image format support (PNG, JPEG, etc.)
- `github.com/hhrutter/tiff v1.0.2` - TIFF support for pdfcpu
- `github.com/clipperhouse/uax29/v2 v2.2.0` - Unicode text segmentation

#### Cryptography & Encoding
- `golang.org/x/crypto v0.43.0` - Cryptographic primitives
- `github.com/hhrutter/pkcs7 v0.2.0` - PKCS#7 support for PDF signatures
- `github.com/hhrutter/lzw v1.0.0` - LZW compression

#### Text Processing
- `golang.org/x/text v0.30.0` - Unicode and text processing
- `golang.org/x/exp v0.0.0-20231006140011-7918f672742d` - Experimental packages

#### Error Handling
- `github.com/pkg/errors v0.9.1` - Enhanced error handling

## Build Tools

### Make
- **Configuration**: `/Users/lgbarn/Personal/pdf-cli/Makefile`
- **Key Targets**:
  - `build` - Single binary for host platform
  - `build-all` - Cross-compile for Linux, macOS, Windows (amd64, arm64)
  - `test` - Run test suite
  - `test-coverage` - Generate coverage reports with HTML output
  - `test-race` - Run with race detector
  - `lint` - Run golangci-lint
  - `coverage-check` - Enforce 75% coverage threshold
  - `completions` - Generate shell completions (bash, zsh, fish)

### GoReleaser v2
- **Configuration**: `/Users/lgbarn/Personal/pdf-cli/.goreleaser.yaml`
- **Purpose**: Automated release builds and distribution
- **Build Flags**: `-s -w` (strip debug info and symbol table)
- **Output Formats**:
  - Archives: tar.gz (Linux/macOS), zip (Windows)
  - Package formats: deb, rpm (via nfpm)
  - Checksums file
- **Version Injection**: Sets `main.version`, `main.commit`, `main.date` via ldflags
- **GitHub Integration**: Automated releases to `lgbarn/pdf-cli`

### Build Script
- **Location**: `/Users/lgbarn/Personal/pdf-cli/scripts/build.sh`
- **Features**: Colored output, cross-compilation, file size reporting
- **Platforms**: Same as Makefile (5 platform/arch combinations)

## Code Quality Tools

### golangci-lint v2.8.0
- **Configuration**: `/Users/lgbarn/Personal/pdf-cli/.golangci.yaml`
- **Version**: 2 (config format)
- **Timeout**: 5 minutes
- **Enabled Linters**:
  - `govet` - Official Go static analyzer
  - `ineffassign` - Detect ineffectual assignments
  - `staticcheck` - Advanced Go linter
  - `unused` - Find unused code
  - `misspell` - Spell checker (US locale)
  - `gocritic` - Opinionated linter (diagnostic + style tags)
  - `revive` - Fast, configurable linter
  - `errcheck` - Check error handling
- **Formatters**: `gofmt`, `goimports`
- **Custom Exclusions**: Configured for common cleanup patterns, test files, package comments

### gosec v2.21.4
- **Purpose**: Security scanner for Go code
- **Usage**: CI security scan job
- **Exclusions**: `testdata/` directory
- **Installation**: `go install github.com/securego/gosec/v2/cmd/gosec@v2.21.4`

### Pre-commit Hooks
- **Configuration**: `/Users/lgbarn/Personal/pdf-cli/.pre-commit-config.yaml`
- **Framework**: pre-commit.com
- **Hooks**:
  - File hygiene (trailing whitespace, EOF fixer, YAML validation, merge conflict detection)
  - Large file detection (1MB limit, excluding testdata)
  - Go formatting (`go fmt`)
  - Go vetting (`go vet`)
  - Module tidying (`go mod tidy`)
  - Build verification (`go build`)
  - Test execution (`go test`)
  - Linting (`golangci-lint`)

## Testing Tools

### Go Testing Framework
- **Built-in**: `testing` package
- **Coverage Tools**: `go tool cover`
- **Flags Used**:
  - `-v` (verbose output)
  - `-race` (race detector in CI)
  - `-coverprofile=coverage.out` (coverage tracking)
  - `-covermode=atomic` (thread-safe coverage)
- **Coverage Threshold**: 75% (enforced in CI and Makefile)
- **Test Organization**: `*_test.go` files alongside implementation

### Coverage Reporting
- **Local**: HTML reports via `coverage.html`
- **CI**: Codecov integration (via `codecov/codecov-action@v4`)
- **Files Generated**: Multiple coverage files tracked (`cover.out`, `cover_final.out`, etc.)

## Version Control

### Git
- **Repository**: `https://github.com/lgbarn/pdf-cli`
- **Versioning**: Git tags (semantic versioning pattern `v*`)
- **Ignored Files**: `.gitignore` excludes build artifacts, coverage reports, binaries

### Dependabot
- **Configuration**: `/Users/lgbarn/Personal/pdf-cli/.github/dependabot.yaml`
- **Ecosystems**:
  - `gomod` - Go module dependencies (weekly updates)
  - `github-actions` - GitHub Actions versions (weekly updates)
- **Limits**: 5 open PRs per ecosystem
- **Commit Prefixes**: `deps` (Go), `ci` (Actions)

## CI/CD Platform

### GitHub Actions
- **Workflows**:
  - **CI** (`.github/workflows/ci.yaml`):
    - Lint job (golangci-lint v2.8.0)
    - Test job (with race detection and coverage)
    - Build job (matrix: 5 platform/arch combinations)
    - Security job (gosec scan)
  - **Release** (`.github/workflows/release.yaml`):
    - Triggered on version tags
    - Runs tests
    - Executes GoReleaser
- **Go Setup**: `actions/setup-go@v5` with version from `go.mod`
- **Caching**: Enabled for Go modules
- **Artifacts**: Cross-compiled binaries uploaded via `actions/upload-artifact@v4`

## Shell Completion

### Supported Shells
- **bash** - POSIX-compatible shell
- **zsh** - Z shell
- **fish** - Friendly interactive shell

### Generation
- **Method**: Via cobra's built-in completion command
- **Target**: `make completions` generates files in `completions/` directory
- **File Names**:
  - `pdf.bash`
  - `_pdf` (zsh format)
  - `pdf.fish`

## Binary Distribution

### Artifact Naming
- **Pattern**: `pdf-cli_{version}_{os}_{arch}.tar.gz` (or `.zip` for Windows)
- **Binary Name**: `pdf` (or `pdf.exe` on Windows)
- **Included Files**: `README.md`, `LICENSE`

### Package Formats
- **Debian**: `.deb` packages
- **Red Hat**: `.rpm` packages
- **Maintainer**: lgbarn
- **License**: MIT
- **Install Path**: `/usr/bin/pdf`

## Environment Variables

### Configuration Overrides
- `PDF_CLI_OUTPUT_FORMAT` - Override output format (json, csv, tsv, human)
- `PDF_CLI_VERBOSE` - Enable verbose mode (true/1)
- `PDF_CLI_OCR_LANGUAGE` - Default OCR language
- `PDF_CLI_OCR_BACKEND` - OCR backend preference (auto, native, wasm)
- `XDG_CONFIG_HOME` - Config directory location
- `TESSDATA_PREFIX` - Tesseract data directory (for native OCR)

### Build Variables
- `GOPATH` - Go workspace path
- `GOOS` - Target operating system
- `GOARCH` - Target architecture
- `GO111MODULE` - Go modules mode (always `on`)
- `CGO_ENABLED` - CGO compilation (set to `0`)

## Documentation Tools

### Markdown
- **Files**: `README.md`, `CHANGELOG.md`, `CONTRIBUTING.md`, `SECURITY.md`
- **Architecture Docs**: `docs/architecture.md`, `docs/plans/*.md`

### Package Documentation
- **Format**: Go doc comments
- **Package Docs**: `doc.go` files in each package
- **Examples**: Embedded in command help text via cobra
