# External Integrations

## Overview
pdf-cli is designed as a self-contained, offline-first CLI tool with minimal external dependencies. The only external integration is for OCR language data downloads from GitHub. No databases, authentication services, or third-party APIs are used. All PDF operations occur locally using embedded Go libraries.

## Findings

### External Services

#### Tesseract OCR Training Data (GitHub)

- **Service**: GitHub raw file hosting
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr.go` (line 29)
  - Base URL: `https://github.com/tesseract-ocr/tessdata_fast/raw/main`
  - Purpose: Download language training data files (`.traineddata`) for OCR operations

- **When triggered**:
  - First use of OCR functionality with WASM backend
  - When requested language data not found in local cache
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr.go` (lines 171-181)

- **Download pattern**:
  - URL format: `{baseURL}/{lang}.traineddata`
  - Examples:
    - English: `https://github.com/tesseract-ocr/tessdata_fast/raw/main/eng.traineddata`
    - German: `https://github.com/tesseract-ocr/tessdata_fast/raw/main/deu.traineddata`
    - French: `https://github.com/tesseract-ocr/tessdata_fast/raw/main/fra.traineddata`
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr.go` (line 213)

- **HTTP client configuration**:
  - Client: `http.DefaultClient` (standard library)
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr.go` (line 260)
  - Method: HTTP GET with context
  - Timeout: 5 minutes (DefaultDownloadTimeout)
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr.go` (line 38)

- **Retry logic**:
  - Max attempts: 3 (DefaultRetryAttempts)
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr.go` (line 44)
  - Base delay: 1 second with exponential backoff
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr.go` (line 47)
  - Retry package: `/Users/lgbarn/Personal/pdf-cli/internal/retry/retry.go`
  - Retryable: Network errors, HTTP 429 (Too Many Requests), HTTP 5xx
  - Permanent failure: HTTP 4xx (except 429)
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr.go` (lines 268-277)

- **Security measures**:
  - SHA256 checksum verification for known languages
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr.go` (lines 305-315)
  - Checksum database: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/checksums.go`
  - Supply chain attack detection: Aborts on checksum mismatch with clear error message
  - Path sanitization: Prevents directory traversal
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr.go` (lines 216-220)

- **Local storage**:
  - Directory: `~/.config/pdf-cli/tessdata/` (Linux/macOS)
  - Directory: `%APPDATA%\pdf-cli\tessdata\` (Windows, inferred)
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr.go` (line 162)
  - Permissions: 0750 (owner rwx, group rx, no world access)
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr.go` (line 41)
  - Files persist across runs (cache)

- **Progress tracking**:
  - Real-time progress bar with byte count
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr.go` (lines 279-286)
  - Output: stderr (non-blocking for pipelines)

- **User notification**:
  - Download initiation message: "Downloading tessdata for '{lang}'..."
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr.go` (line 222)
  - Checksum verification: "Checksum verified for {lang}.traineddata"
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr.go` (line 315)
  - Warning for missing checksum: Displays computed SHA256
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr.go` (lines 317-320)

#### Network Error Handling

- **No external connectivity required** for non-OCR operations
- **Graceful degradation**: If tessdata download fails, OCR operation fails with clear error
- **Offline mode**: Native Tesseract backend uses system-installed language data (no download)
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/native.go` (lines 108-114)
  - System paths checked: `/opt/homebrew/share/tessdata`, `/usr/local/share/tessdata`, `/usr/share/tesseract-ocr/*/tessdata`
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/detect.go` (lines 96-119)

### External System Dependencies (Optional)

#### Native Tesseract OCR

- **Type**: Optional system-installed binary
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/detect.go` (lines 24-42)
  - Detection: `exec.LookPath("tesseract")`
  - Execution: Subprocess via `exec.CommandContext`
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/native.go` (line 70)

- **Why external**:
  - Native backend preferred for performance (faster than WASM)
  - Provides access to system-installed language data
  - Reduces binary size (WASM engine embedded in pdf-cli binary)

- **Communication protocol**:
  - Input: Image file path
  - Output: Text written to temporary file
  - Command pattern: `tesseract [--tessdata-dir DIR] IMAGE_PATH OUTPUT_BASE -l LANG`
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/native.go` (lines 86-92)
  - Temporary files cleaned up via cleanup registry
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/native.go` (lines 51-66)

- **Version detection**:
  - Command: `tesseract --version`
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/detect.go` (lines 44-62)
  - Regex: `tesseract\s+(\d+\.\d+(?:\.\d+)?)`
  - Version stored in NativeInfo struct

- **Language availability**:
  - Checks for `.traineddata` files in system tessdata directory
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/native.go` (lines 108-114)
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/detect.go` (lines 121-142)

- **Fallback behavior**:
  - If native Tesseract not found, automatically falls back to WASM backend
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr.go` (lines 132-136)
  - User can force backend via config or flag
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/config/config.go` (line 43)

### Configuration Sources

#### Environment Variables

- **Configuration overrides** (no external service calls):
  - `PDF_CLI_OUTPUT_FORMAT`: Output format (json, csv, tsv, human)
  - `PDF_CLI_VERBOSE`: Enable verbose logging (true/1)
  - `PDF_CLI_OCR_LANGUAGE`: Default OCR language
  - `PDF_CLI_OCR_BACKEND`: OCR backend selection (auto, native, wasm)
  - `PDF_CLI_PERF_OCR_THRESHOLD`: Parallel OCR threshold
  - `PDF_CLI_PERF_TEXT_THRESHOLD`: Parallel text extraction threshold
  - `PDF_CLI_PERF_MAX_WORKERS`: Max parallel workers
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/config/config.go` (lines 136-164)

- **XDG Base Directory Spec** (Linux/macOS):
  - `XDG_CONFIG_HOME`: Custom config directory location
  - Default: `~/.config`
  - Config file: `${XDG_CONFIG_HOME}/pdf-cli/config.yaml`
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/config/config.go` (lines 95-106)

- **System environment** (native Tesseract):
  - `TESSDATA_PREFIX`: Override tessdata directory location
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/detect.go` (lines 66-69)

### File System Integrations

#### Local Configuration File

- **Type**: YAML configuration file
  - Path: `~/.config/pdf-cli/config.yaml` (XDG-compliant)
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/config/config.go` (line 105)
  - Format: YAML (gopkg.in/yaml.v3)
  - Permissions: 0644 (DefaultFilePerm, inferred from fileio package)

- **Configuration sections**:
  - defaults: output_format, verbose, show_progress
  - compress: quality (low, medium, high)
  - encrypt: algorithm (aes128, aes256)
  - ocr: language, backend
  - performance: ocr_parallel_threshold, text_parallel_threshold, max_workers
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/config/config.go` (lines 14-51)

- **Load behavior**:
  - Non-existence not treated as error (uses defaults)
  - Environment variables override config file values
  - Singleton pattern with lazy initialization
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/config/config.go` (lines 187-213)

#### Temporary File Storage

- **System temp directory**: Used for intermediate processing
  - PDF image extraction: `os.MkdirTemp("", "pdf-ocr-*")`
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr.go` (line 334)
  - OCR output files: `os.CreateTemp("", "ocr-output-*.txt")`
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/native.go` (line 51)
  - Tessdata downloads: `os.CreateTemp(dataDir, "tessdata-*.tmp")`
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr.go` (line 227)

- **Cleanup mechanism**:
  - Cleanup registry tracks all temporary files
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/cleanup/cleanup.go` (inferred from imports)
  - Deferred cleanup on function exit
  - Global cleanup on program termination
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/cmd/pdf/main.go` (line 28)

### Standard I/O Integration

#### stdin/stdout Support

- **stdin reading**: Supports PDF input from pipe
  - Pattern: `pdf text -` or `pdf info -`
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/README.md` (lines 63-64)
  - Implementation: `/Users/lgbarn/Personal/pdf-cli/internal/fileio/stdio.go` (inferred)
  - Allows pipeline operations: `cat doc.pdf | pdf text -`

- **stdout writing**: Supports PDF output to pipe
  - Pattern: `pdf merge file1.pdf file2.pdf > output.pdf`
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/README.md` (table line 98-100)
  - Enables Unix pipeline workflows

- **stderr usage**:
  - Progress bars written to stderr (non-blocking)
  - Error messages written to stderr
  - Allows piping output while displaying progress

### No Database Integration

- **No database servers**: All operations in-memory or file-based
- **No embedded databases**: No SQLite, BoltDB, or similar
- **State persistence**: Only via configuration file and tessdata cache

### No Authentication Services

- **No OAuth/OIDC**: No external authentication
- **No API keys**: No service-specific credentials required
- **Password handling**: Only for PDF encryption/decryption (local operation)
  - Passwords passed via CLI flags or stdin
  - Not logged or persisted
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/SECURITY.md` (lines 36-39)

### No Cloud Services

- **No cloud storage**: S3, GCS, Azure Blob, etc. not supported
- **No cloud APIs**: No AWS, GCP, Azure service integrations
- **No telemetry**: No usage tracking or analytics
- **No update checks**: No phone-home behavior

### CI/CD Integrations (Development Only)

#### GitHub Actions

- **Purpose**: Automated testing and releases
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/.github/workflows/ci.yaml`
  - Triggers: Push to main/master, pull requests
  - Jobs: lint, test, build, security scan

- **External actions used**:
  - actions/checkout@v4 -- Repository checkout
  - actions/setup-go@v5 -- Go installation
  - actions/upload-artifact@v4 -- Build artifact storage
  - golangci/golangci-lint-action@v8 -- Linting
  - codecov/codecov-action@v4 -- Coverage upload
  - goreleaser/goreleaser-action@v6 -- Release automation
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/.github/workflows/ci.yaml` and `release.yaml`

#### Codecov

- **Service**: Code coverage tracking
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/.github/workflows/ci.yaml` (lines 56-60)
  - Upload: codecov-action uploads `coverage.out`
  - Fail behavior: `fail_ci_if_error: false` (non-blocking)
  - Purpose: Code quality metrics (not required for operation)

#### Dependabot

- **Service**: GitHub-native dependency updates
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/.github/dependabot.yaml`
  - Monitors: Go modules, GitHub Actions
  - Frequency: Weekly
  - No external API calls (GitHub-integrated)

## Summary Table

| Integration | Type | Purpose | Network Required | Confidence |
|------------|------|---------|-----------------|------------|
| GitHub tessdata | HTTP Download | OCR language data | Yes (first use) | Observed |
| Native Tesseract | System Binary | OCR processing | No | Observed |
| XDG Config | File System | User configuration | No | Observed |
| Temp Directory | File System | Intermediate files | No | Observed |
| stdin/stdout | Standard I/O | Pipeline support | No | Observed |
| GitHub Actions | CI/CD | Automated builds | N/A (dev only) | Observed |
| Codecov | CI/CD | Coverage reporting | N/A (dev only) | Observed |
| Dependabot | CI/CD | Dependency updates | N/A (dev only) | Observed |

## Security Considerations

### Download Security

- **Checksum verification**: SHA256 validation for tessdata downloads
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr.go` (lines 305-315)
  - Known checksums: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/checksums.go`
  - Failure mode: Explicit error message warning of potential supply chain attack

- **Path sanitization**: Prevents directory traversal in file operations
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr.go` (lines 216-220)
  - Function: `fileio.SanitizePath()`

- **HTTPS enforcement**: URL upgrades HTTP to HTTPS (if supported)
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/SECURITY.md` (line 42)
  - Base URL already HTTPS

### Privacy

- **No telemetry**: No usage data collected
- **No external logging**: Logs stay local
- **No credential storage**: Passwords not persisted
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/SECURITY.md` (lines 36-39)

### Offline Operation

- **Core features**: Work completely offline (merge, split, encrypt, text, images, compress, etc.)
- **OCR fallback**: Native Tesseract uses system data (offline)
- **WASM OCR**: Requires one-time download per language, then offline

## Integration Failure Modes

### Tessdata Download Failure

- **Symptom**: OCR operation fails with download error
- **Causes**: Network unavailable, GitHub down, rate limiting, firewall
- **Mitigation**: Install native Tesseract with system package manager
- **User control**: Can pre-download tessdata files manually

### Native Tesseract Unavailable

- **Symptom**: Fallback to WASM backend (slower)
- **Causes**: Tesseract not installed or not in PATH
- **Mitigation**: Automatic fallback (transparent to user)
- **User control**: Set `PDF_CLI_OCR_BACKEND=wasm` to bypass detection

### Filesystem Errors

- **Symptom**: Cannot create temp files or read config
- **Causes**: Disk full, permission denied, invalid paths
- **Mitigation**: Detailed error messages guide user to issue
- **User control**: Set TMPDIR or XDG_CONFIG_HOME to writable location

## Open Questions

- Is there a plan to support direct HTTP(S) URL input for PDFs (e.g., `pdf info https://example.com/doc.pdf`)?
- Are there any plans to integrate with cloud storage providers (S3, Dropbox, etc.)?
- Should there be an option to use a local mirror for tessdata downloads (e.g., corporate networks)?
- Is there consideration for optional telemetry (opt-in) to track feature usage for prioritization?
