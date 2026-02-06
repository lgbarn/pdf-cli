# External Integrations

## External Services

### GitHub

#### Tesseract Training Data Repository
- **URL**: `https://github.com/tesseract-ocr/tessdata_fast`
- **Purpose**: Download OCR language training data files
- **Usage**: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr.go:24`
- **Constant**: `TessdataURL = "https://github.com/tesseract-ocr/tessdata_fast/raw/main"`
- **Access Pattern**: HTTP GET requests for `{lang}.traineddata` files
- **Languages Supported**: English (eng), German (deu), French (fra), and others
- **When Used**: On-demand when WASM OCR backend requires language data not already cached
- **Cache Location**: User config directory (`~/.config/pdf-cli/tessdata/` or `XDG_CONFIG_HOME`)
- **Network Dependency**: Required for first-time OCR usage with WASM backend (after that, cached locally)
- **Fallback**: No fallback - operation fails if download fails and file not cached

#### GitHub Releases (Distribution)
- **Repository**: `lgbarn/pdf-cli`
- **Purpose**: Binary distribution and version management
- **Automated via**: GoReleaser in `.github/workflows/release.yaml`
- **Artifacts**: Cross-platform binaries, checksums, deb/rpm packages
- **Trigger**: Git tags matching `v*` pattern

#### GitHub Actions
- **Purpose**: CI/CD automation
- **Workflows**: 2 workflows (CI, Release)
- **Permissions**: `contents: read` (CI), `contents: write` (Release)
- **External Actions Used**:
  - `actions/checkout@v4` - Repository checkout
  - `actions/setup-go@v5` - Go toolchain installation
  - `actions/upload-artifact@v4` - Build artifact storage
  - `golangci/golangci-lint-action@v8` - Linting
  - `goreleaser/goreleaser-action@v6` - Release automation
  - `codecov/codecov-action@v4` - Coverage reporting

#### Codecov
- **Purpose**: Code coverage tracking and visualization
- **Integration**: Via `codecov/codecov-action@v4` in CI workflow
- **Upload**: Automatic on PR and main branch commits
- **File**: `coverage.out` (atomic mode)
- **Fail Behavior**: `fail_ci_if_error: false` (non-blocking)
- **Dashboard**: Likely at `https://codecov.io/gh/lgbarn/pdf-cli`

#### Dependabot
- **Purpose**: Automated dependency updates
- **Configuration**: `/Users/lgbarn/Personal/pdf-cli/.github/dependabot.yaml`
- **Update Frequency**: Weekly
- **Monitored Ecosystems**:
  - Go modules (`go.mod`, `go.sum`)
  - GitHub Actions (workflow files)
- **PR Limits**: 5 per ecosystem
- **Commit Convention**: Prefixed with `deps:` or `ci:`

## External System Dependencies

### Tesseract OCR (Optional Native Backend)
- **Type**: Optional system dependency
- **Purpose**: Native OCR processing (alternative to WASM backend)
- **Detection**: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/detect.go`
- **Discovery Method**: `exec.LookPath("tesseract")`
- **Version Detection**: `tesseract --version` command
- **Supported Platforms**:
  - **macOS**: Homebrew (`/opt/homebrew/share/tessdata`, `/usr/local/share/tessdata`)
  - **Linux**: Package managers (`/usr/share/tesseract-ocr/*/tessdata`, `/usr/share/tessdata`)
  - **Windows**: Program Files (`C:\Program Files\Tesseract-OCR\tessdata`)
- **Backend Selection**: Auto-selected if available, otherwise falls back to WASM
- **Configuration**: Can be forced via `--ocr-backend native` flag or `PDF_CLI_OCR_BACKEND=native`
- **Tessdata Location**: Detected via:
  1. `TESSDATA_PREFIX` environment variable
  2. `tesseract --print-parameters` output parsing
  3. Common OS-specific paths
- **No Installation Required**: WASM backend works without Tesseract installed

## Data Storage

### Local File System

#### User Configuration
- **Location**: `~/.config/pdf-cli/config.yaml`
- **Standard**: XDG Base Directory Specification
- **Override**: `XDG_CONFIG_HOME` environment variable
- **Format**: YAML
- **Permissions**: 0600 (read/write owner only)
- **Directory Permissions**: 0750
- **Structure**:
  ```yaml
  defaults:
    output_format: human|json|csv|tsv
    verbose: bool
    show_progress: bool
  compress:
    quality: low|medium|high
  encrypt:
    algorithm: aes128|aes256
  ocr:
    language: eng|deu|fra|...
    backend: auto|native|wasm
  ```
- **Default Behavior**: Falls back to hardcoded defaults if file missing

#### OCR Training Data Cache
- **Location**: Same as config directory (`~/.config/pdf-cli/tessdata/`)
- **Files**: `{language}.traineddata` (e.g., `eng.traineddata`)
- **Size**: Varies by language (~10-30MB per language)
- **Persistence**: Permanent cache (not cleaned automatically)
- **Download**: Automatic on first use of WASM OCR with specific language
- **Reuse**: Shared between all WASM OCR operations

#### Temporary Files
- **Base**: System temp directory (via `os.MkdirTemp`)
- **Patterns**:
  - `pdf-cli-text-*` - Text extraction working directory
  - `pdf-merge-*.pdf` - Merge operation intermediate file
  - Image files during OCR processing
- **Cleanup**: Deferred cleanup via `defer os.RemoveAll(tmpDir)`
- **Permissions**: Default system temp permissions

## I/O Interfaces

### Standard Input/Output
- **stdin Support**: Yes - use `-` as input filename
- **stdout Support**: Yes - via `--stdout` flag for binary PDF output
- **stderr**: Used for progress bars and logging
- **Use Cases**:
  - `cat input.pdf | pdf text -`
  - `curl -s https://example.com/doc.pdf | pdf info -`
  - `pdf compress - --stdout > output.pdf`
- **Implementation**: `/Users/lgbarn/Personal/pdf-cli/internal/fileio/stdio.go`
- **Pattern Handling**: `/Users/lgbarn/Personal/pdf-cli/internal/commands/patterns/stdio.go`

### Terminal Detection
- **Library**: `golang.org/x/term`
- **Purpose**: Determine if running in interactive terminal
- **Impact**: Affects progress bar display and password prompting
- **Detection Points**:
  - Interactive mode for confirmation prompts
  - Progress bar suppression in non-TTY contexts
  - Password input masking

## No Database Systems

**Note**: This application does not integrate with any database systems. All data is:
- Processed in-memory during operations
- Read from and written to PDF files
- Configured via local YAML file
- Cached in local filesystem (OCR training data)

## No Authentication Services

**Note**: No OAuth, API keys, or authentication services are used except:
- GitHub tokens (only in CI/CD via `GITHUB_TOKEN` secret, managed by GitHub Actions)
- PDF user/owner passwords (provided by user, used locally only)

## Network Communication Summary

### Outbound Connections
1. **Tesseract Training Data Downloads**
   - Host: `github.com`
   - Path: `/tesseract-ocr/tessdata_fast/raw/main/{lang}.traineddata`
   - Protocol: HTTPS
   - Frequency: Once per language (cached thereafter)
   - Required: Only for WASM OCR with new language

### No Inbound Connections
- Application does not listen on any network ports
- No web server or API endpoints
- Pure CLI tool

## External Command Execution

### Tesseract (Native Backend)
- **Command**: `tesseract` (from PATH)
- **Detection**: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/detect.go:26`
  - `exec.LookPath("tesseract")` finds executable
- **Version Check**: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/detect.go:45`
  - `tesseract --version`
- **Data Directory Detection**: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/detect.go:72`
  - `tesseract --print-parameters`
- **OCR Processing**: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/native.go`
  - `tesseract {input} {output} -l {lang}`
  - Input: Temporary image files
  - Output: Text files
  - Context-aware: Supports cancellation via `context.Context`
- **Security**: All commands use `#nosec G204` with justification - paths validated via `exec.LookPath`

### No Other External Commands
- Build system uses standard Go toolchain
- No shell scripts executed at runtime
- No system utilities called (other than Tesseract)

## Package Repositories

### Go Module Proxy
- **Default**: `proxy.golang.org`
- **Purpose**: Dependency resolution and download
- **Dependencies**: 13 direct, 22 total (including indirect)
- **Checksum Database**: `sum.golang.org`
- **Files**: `go.mod`, `go.sum`

### Pre-commit Hook Repository
- **Host**: `github.com/pre-commit/pre-commit-hooks`
- **Version**: v5.0.0
- **Purpose**: Developer tooling (not runtime dependency)

## Monitoring & Observability

### No Telemetry
- No analytics or telemetry data collected
- No phone-home functionality
- No crash reporting services
- No usage tracking

### Logging
- **Destination**: stderr (via `/Users/lgbarn/Personal/pdf-cli/internal/logging/logger.go`)
- **Levels**: Error, Warning, Info, Debug, Verbose
- **Format**: Text-based, human-readable
- **Configuration**: Controlled by `--verbose`, `--log-level`, `--log-format` flags
- **Output**: Local only (never sent to external services)

## Security Considerations

### Security Scanning
- **Tool**: gosec v2.21.4
- **Frequency**: Every PR and commit to main
- **Scope**: All Go source files (excluding `testdata/`)
- **Integration**: GitHub Actions security job

### Dependency Scanning
- **Tool**: Dependabot (GitHub native)
- **Scope**: Go modules and GitHub Actions
- **Alerts**: Via GitHub Security tab
- **Updates**: Automated PRs weekly

### HTTPS Enforcement
- **Tesseract Data**: Always downloaded via HTTPS
- **No Insecure Protocols**: No HTTP, FTP, or other insecure protocols used

### No Secrets Storage
- PDF passwords: Provided via flags/prompts, used in-memory only
- No credentials persisted to disk
- No API keys or tokens (except CI/CD environment)

## Platform-Specific Integrations

### macOS
- **Homebrew Paths**: `/opt/homebrew`, `/usr/local` (for Tesseract)
- **Code Signing**: Not implemented in current build
- **Notarization**: Not implemented

### Linux
- **Package Managers**: deb/rpm packages generated via GoReleaser + nfpm
- **Install Locations**: `/usr/bin/pdf`
- **Standards**: FHS (Filesystem Hierarchy Standard) compliant

### Windows
- **Program Files**: Detection for Tesseract installation
- **Path Separators**: Handled via `filepath` package
- **Binary Extension**: `.exe` suffix
- **Archive Format**: ZIP (instead of tar.gz)

## Future Integration Considerations

Based on code structure and patterns, potential future integrations might include:
- Cloud storage providers (for PDF source/destination)
- Additional OCR backends
- PDF/A validation services
- Digital signature services

However, **none of these are currently implemented**.
