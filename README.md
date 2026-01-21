# pdf-cli

[![CI](https://github.com/lgbarn/pdf-cli/actions/workflows/ci.yaml/badge.svg)](https://github.com/lgbarn/pdf-cli/actions/workflows/ci.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/lgbarn/pdf-cli)](https://goreportcard.com/report/github.com/lgbarn/pdf-cli)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/github/go-mod/go-version/lgbarn/pdf-cli)](https://go.dev/)

A fast, lightweight command-line tool for everyday PDF operations. No GUI needed, no complicated setup—just simple commands to merge, split, compress, encrypt, and manipulate PDF files.

## Table of Contents

- [Why pdf-cli?](#why-pdf-cli)
- [Quick Start](#quick-start)
- [Installation](#installation)
- [Commands](#commands)
- [Usage Examples](#usage-examples)
- [Global Options](#global-options)
- [Configuration](#configuration)
- [Shell Completion](#shell-completion)
- [Building from Source](#building-from-source)
- [Troubleshooting](#troubleshooting)
- [Contributing](#contributing)
- [License](#license)

## Why pdf-cli?

- **Fast**: Single binary with no external dependencies, parallel processing for large operations
- **Simple**: Intuitive commands that do one thing well
- **Secure**: Supports encrypted PDFs with password protection
- **Cross-platform**: Works on Linux, macOS, and Windows
- **Scriptable**: Perfect for automation and batch processing with JSON/CSV/TSV output
- **Unix-friendly**: Supports stdin/stdout for seamless pipelines
- **OCR Support**: Extract text from scanned PDFs using native Tesseract (when installed) or built-in WASM fallback

## Quick Start

```bash
# Install
go install github.com/lgbarn/pdf-cli/cmd/pdf@latest

# Merge two PDFs
pdf merge -o combined.pdf file1.pdf file2.pdf

# Extract pages 1-5 from a PDF
pdf extract document.pdf -p 1-5 -o pages.pdf

# Compress a large PDF
pdf compress large.pdf -o smaller.pdf

# Batch compress multiple PDFs
pdf compress *.pdf

# Get PDF info
pdf info document.pdf

# Get PDF info as JSON (for scripting)
pdf info document.pdf --format json

# Extract text from a scanned PDF using OCR
pdf text scanned.pdf --ocr

# Process PDF from stdin (Unix pipes)
cat document.pdf | pdf text -
curl -s https://example.com/doc.pdf | pdf info -
```

## Installation

### Prerequisites

- Go 1.24 or later (for installation via `go install`)

### Using Go (Recommended)

```bash
go install github.com/lgbarn/pdf-cli/cmd/pdf@latest
```

### Pre-built Binaries

Download the latest release for your platform from the [Releases page](https://github.com/lgbarn/pdf-cli/releases).

Available platforms:
- Linux (amd64, arm64)
- macOS (amd64, arm64)
- Windows (amd64)

### From Source

```bash
git clone https://github.com/lgbarn/pdf-cli.git
cd pdf-cli
make build
```

## Commands

| Command | Description | Batch | stdin | stdout |
|---------|-------------|:-----:|:-----:|:------:|
| `info` | Display PDF information (pages, metadata, encryption status) | ✓ | ✓ | - |
| `merge` | Combine multiple PDFs into a single file | - | - | - |
| `split` | Split a PDF into individual pages or chunks | - | - | - |
| `extract` | Extract specific pages into a new PDF | - | ✓ | ✓ |
| `reorder` | Reorder, reverse, or duplicate pages | - | ✓ | ✓ |
| `rotate` | Rotate pages by 90, 180, or 270 degrees | ✓ | ✓ | ✓ |
| `compress` | Optimize and reduce PDF file size | ✓ | ✓ | ✓ |
| `encrypt` | Add password protection to a PDF | ✓ | ✓ | ✓ |
| `decrypt` | Remove password protection from a PDF | ✓ | ✓ | ✓ |
| `text` | Extract text content (supports OCR for scanned PDFs) | - | ✓ | - |
| `images` | Extract embedded images from a PDF | - | - | - |
| `combine-images` | Create a PDF from multiple images | - | - | - |
| `meta` | View or modify PDF metadata (title, author, etc.) | ✓ | - | - |
| `watermark` | Add text or image watermarks | ✓ | - | - |
| `pdfa` | PDF/A validation and conversion | - | ✓ | ✓ |

## Usage Examples

### Get PDF Information

```bash
# Single file - detailed output
pdf info document.pdf

# Multiple files - summary table
pdf info *.pdf

# Machine-readable output (JSON, CSV, TSV)
pdf info document.pdf --format json
pdf info *.pdf --format csv > report.csv
pdf info *.pdf --format tsv

# Process via jq
pdf info document.pdf --format json | jq '.pages'
```

Single file output:
```
File:       document.pdf
Size:       2.45 MB
Pages:      42
Version:    1.7
Title:      Annual Report
Author:     John Doe
Encrypted:  No
```

JSON output (`--format json`):
```json
{
  "file": "document.pdf",
  "size": 2568192,
  "sizeHuman": "2.45 MB",
  "pages": 42,
  "version": "1.7",
  "title": "Annual Report",
  "author": "John Doe",
  "encrypted": false
}
```

Batch output:
```
FILE                                        PAGES    VER       SIZE
----------------------------------------------------------------------
document1.pdf                                  42    1.7    2.45 MB
document2.pdf                                  15    1.5  512.00 KB
report.pdf                                    128    1.7   10.23 MB
```

### Merge Multiple PDFs

```bash
# Merge two files
pdf merge -o combined.pdf file1.pdf file2.pdf

# Merge all PDFs in a directory
pdf merge -o combined.pdf *.pdf
```

### Split a PDF

```bash
# Split into individual pages (creates page_001.pdf, page_002.pdf, etc.)
pdf split document.pdf -o output/

# Split into chunks of 5 pages each
pdf split document.pdf -n 5 -o chunks/
```

### Extract Specific Pages

```bash
# Extract pages 1 through 5
pdf extract document.pdf -p 1-5 -o first-five.pdf

# Extract specific pages and ranges
pdf extract document.pdf -p 1,3,5,10-15 -o selected.pdf
```

### Reorder Pages

```bash
# Move page 5 to position 2
pdf reorder document.pdf -s "1,5,2,3,4" -o reordered.pdf

# Reverse all pages
pdf reorder document.pdf -s "end-1" -o reversed.pdf

# Duplicate page 1 at the end
pdf reorder document.pdf -s "1-end,1" -o with-copy.pdf

# Remove the first page
pdf reorder document.pdf -s "2-end" -o skip-first.pdf
```

### Rotate Pages

```bash
# Rotate all pages 90 degrees clockwise
pdf rotate document.pdf -a 90 -o rotated.pdf

# Rotate only pages 1-5 by 180 degrees
pdf rotate document.pdf -a 180 -p 1-5 -o rotated.pdf
```

### Compress a PDF

```bash
# Compress a single file
pdf compress large.pdf -o smaller.pdf

# Batch compress multiple PDFs (output: *_compressed.pdf)
pdf compress *.pdf

# With progress bar for large files
pdf compress large.pdf -o smaller.pdf --progress

# stdin/stdout support for pipelines
cat large.pdf | pdf compress - --stdout > compressed.pdf
curl -s https://example.com/doc.pdf | pdf compress - --stdout > local.pdf
```

### Encrypt a PDF

```bash
# Add password protection
pdf encrypt document.pdf --password mysecret -o secure.pdf

# Set separate user and owner passwords
pdf encrypt document.pdf --password userpass --owner-password ownerpass -o secure.pdf

# Batch encrypt multiple PDFs (output: *_encrypted.pdf)
pdf encrypt *.pdf --password mysecret
```

### Decrypt a PDF

```bash
# Decrypt a single file
pdf decrypt secure.pdf --password mysecret -o unlocked.pdf

# Batch decrypt multiple PDFs (output: *_decrypted.pdf)
pdf decrypt *.pdf --password mysecret
```

### Extract Text

```bash
# Print text to terminal
pdf text document.pdf

# Save to a file
pdf text document.pdf -o content.txt

# Extract text from specific pages
pdf text document.pdf -p 1-5 -o chapter1.txt

# With progress bar for large documents
pdf text large-document.pdf --progress

# Read from stdin
cat document.pdf | pdf text -
curl -s https://example.com/doc.pdf | pdf text -
```

### Extract Text with OCR (for scanned PDFs)

```bash
# Use OCR for scanned/image-based PDFs
pdf text scanned.pdf --ocr

# OCR with specific language (downloads tessdata on first use for WASM)
pdf text scanned.pdf --ocr --ocr-lang eng

# Multi-language OCR
pdf text scanned.pdf --ocr --ocr-lang eng+fra

# OCR specific pages and save to file
pdf text scanned.pdf --ocr -p 1-10 -o content.txt

# Force native Tesseract (if installed)
pdf text scanned.pdf --ocr --ocr-backend=native

# Force WASM Tesseract (no system dependencies)
pdf text scanned.pdf --ocr --ocr-backend=wasm

# Auto-select (native if available, else WASM) - this is the default
pdf text scanned.pdf --ocr --ocr-backend=auto
```

**OCR Backend Selection:**
- `auto` (default): Uses native Tesseract if installed, otherwise falls back to WASM
- `native`: Requires system Tesseract installation but provides better quality/speed
- `wasm`: Built-in, no external dependencies, downloads tessdata on first use (~15MB/language)

### Extract Images

```bash
# Extract all images
pdf images document.pdf -o images/

# Extract images from specific pages
pdf images document.pdf -p 1-10 -o images/
```

### Using stdin/stdout Pipelines

pdf-cli supports Unix-style pipelines for processing PDFs without intermediate files:

```bash
# Download and extract text in one command
curl -s https://example.com/document.pdf | pdf text -

# Download, compress, and save
curl -s https://example.com/large.pdf | pdf compress - --stdout > compressed.pdf

# Chain multiple operations
cat input.pdf | pdf extract - -p 1-5 --stdout | pdf rotate - -a 90 --stdout > output.pdf

# Process PDF from another command
generate-report | pdf compress - --stdout > report.pdf

# Get info from a remote PDF
curl -s https://example.com/doc.pdf | pdf info - --format json | jq '.pages'
```

**Notes:**
- Use `-` as the input file to read from stdin
- Use `--stdout` flag to write binary output to stdout
- When using stdin, pdfcpu requires the entire file, so the PDF is temporarily stored

### Combine Images into PDF

```bash
# Create PDF from multiple images
pdf combine-images photo1.jpg photo2.jpg -o album.pdf

# Create PDF from all PNG files in current directory
pdf combine-images *.png -o scans.pdf

# Create PDF with specific page size
pdf combine-images scan1.png scan2.png -o document.pdf --page-size A4
```

### View and Modify Metadata

```bash
# View metadata for a single file
pdf meta document.pdf

# View metadata for multiple files
pdf meta *.pdf

# Set metadata
pdf meta document.pdf --title "My Document" --author "Jane Doe" -o updated.pdf

# Set multiple fields
pdf meta document.pdf \
  --title "Annual Report" \
  --author "John Doe" \
  --subject "2024 Financial Summary" \
  -o updated.pdf
```

### Add Watermarks

```bash
# Add text watermark
pdf watermark document.pdf -t "CONFIDENTIAL" -o marked.pdf

# Add image watermark (logo)
pdf watermark document.pdf -i logo.png -o branded.pdf

# Watermark specific pages only
pdf watermark document.pdf -t "DRAFT" -p 1-5 -o draft.pdf

# Batch watermark multiple PDFs (output: *_watermarked.pdf)
pdf watermark *.pdf -t "CONFIDENTIAL"
```

### PDF/A Validation and Conversion

```bash
# Validate PDF/A compliance
pdf pdfa validate document.pdf

# Validate against specific PDF/A level
pdf pdfa validate document.pdf --level 1b

# Convert/optimize a PDF toward PDF/A format
pdf pdfa convert document.pdf -o archive.pdf

# Convert with specific target level
pdf pdfa convert document.pdf --level 2b -o archive.pdf
```

**Note:** Full PDF/A validation and conversion may require specialized tools. This tool provides basic validation and optimization that can help with PDF/A compliance. For comprehensive validation, consider using [veraPDF](https://verapdf.org/).

> **⚠️ PDF/A Limitations**
>
> This tool provides **basic** PDF/A validation and optimization, not full ISO compliance:
>
> | Feature | Status |
> |---------|--------|
> | Structure validation | ✓ Supported |
> | Encryption detection | ✓ Supported |
> | Font embedding check | ✗ Limited |
> | Color profile validation | ✗ Not supported |
> | Full ISO 19005 compliance | ✗ Not supported |
>
> For comprehensive PDF/A validation, use [veraPDF](https://verapdf.org/).
> For full PDF/A conversion, consider Ghostscript or Adobe Acrobat.

## Global Options

These options work with all commands:

| Option | Short | Description |
|--------|-------|-------------|
| `--verbose` | `-v` | Show detailed output during operations |
| `--force` | `-f` | Overwrite existing files without prompting |
| `--progress` | | Show progress bar for long operations |
| `--password` | `-P` | Password for encrypted input PDFs |
| `--dry-run` | | Preview what would happen without making changes |
| `--log-level` | | Set logging level: `debug`, `info`, `warn`, `error`, `silent` (default: silent) |
| `--log-format` | | Set log format: `text` or `json` (default: text) |
| `--help` | `-h` | Show help for any command |
| `--version` | | Display version information |

### Dry-Run Mode

Preview operations without making any changes:

```bash
# See what files would be created
pdf compress *.pdf --dry-run

# Preview merge operation
pdf merge -o combined.pdf *.pdf --dry-run

# Check encryption without modifying files
pdf encrypt document.pdf --password secret --dry-run
```

### Logging

Enable structured logging for debugging or monitoring:

```bash
# Debug logging to see detailed operations
pdf compress large.pdf --log-level debug

# JSON logging for log aggregation
pdf merge -o out.pdf *.pdf --log-level info --log-format json
```

### Command-Specific Options

| Option | Commands | Description |
|--------|----------|-------------|
| `--format` | info, meta, pdfa | Output format: `json`, `csv`, `tsv` (default: human-readable) |
| `--stdout` | compress, extract, rotate, reorder, encrypt, decrypt, pdfa convert | Write binary output to stdout |
| `-` (stdin) | text, info, compress, extract, rotate, reorder, encrypt, decrypt, pdfa convert | Read PDF from stdin |

### Working with Encrypted PDFs

Most commands accept a `--password` flag for reading encrypted PDFs:

```bash
pdf info secure.pdf --password mysecret
pdf extract secure.pdf -p 1-5 -o pages.pdf --password mysecret
```

## Configuration

pdf-cli supports an optional configuration file for setting default values.

### Config File Location

The config file is loaded from (in order of precedence):
1. `$XDG_CONFIG_HOME/pdf-cli/config.yaml`
2. `~/.config/pdf-cli/config.yaml`

### Example Configuration

```yaml
# ~/.config/pdf-cli/config.yaml

defaults:
  verbose: false
  force: false
  progress: true

compress:
  # No specific defaults

encrypt:
  # Default encryption settings

ocr:
  language: "eng"
  backend: "auto"  # auto, native, or wasm
```

### Environment Variables

All config options can be overridden with environment variables using the `PDF_CLI_` prefix:

```bash
# Override verbose mode
export PDF_CLI_VERBOSE=true

# Override OCR language
export PDF_CLI_OCR_LANGUAGE=eng+fra

# Override OCR backend
export PDF_CLI_OCR_BACKEND=native
```

Environment variables take precedence over config file values.

## Shell Completion

Enable tab completion for your shell:

### Bash

```bash
# Add to ~/.bashrc
echo 'source <(pdf completion bash)' >> ~/.bashrc

# Or install system-wide
pdf completion bash | sudo tee /etc/bash_completion.d/pdf > /dev/null
```

### Zsh

```bash
# Add to ~/.zshrc
echo 'source <(pdf completion zsh)' >> ~/.zshrc
```

### Fish

```bash
pdf completion fish > ~/.config/fish/completions/pdf.fish
```

### PowerShell

```powershell
pdf completion powershell | Out-String | Invoke-Expression
```

## Building from Source

### Prerequisites

- Go 1.24 or later
- Make (optional, for convenience commands)

### Build Commands

```bash
# Clone the repository
git clone https://github.com/lgbarn/pdf-cli.git
cd pdf-cli

# Build for your current platform
make build

# Run tests
make test

# Run tests with coverage report
make test-coverage

# Build for all platforms
make build-all

# Clean build artifacts
make clean
```

### Project Structure

```
pdf-cli/
├── cmd/pdf/              # Application entry point
├── internal/
│   ├── cli/              # CLI framework and flags
│   ├── commands/         # Individual command implementations
│   │   └── patterns/     # Reusable command patterns (StdioHandler)
│   ├── config/           # Configuration file support
│   ├── fileio/           # File operations and stdio utilities
│   ├── logging/          # Structured logging with slog
│   ├── ocr/              # OCR text extraction (native + WASM backends)
│   │   ├── backend.go    # Backend interface and types
│   │   ├── detect.go     # Native Tesseract detection
│   │   ├── native.go     # Native Tesseract backend
│   │   ├── wasm.go       # WASM Tesseract backend
│   │   └── ocr.go        # Engine with backend selection
│   ├── output/           # Output formatting (JSON, CSV, TSV)
│   ├── pages/            # Page range parsing and validation
│   ├── pdf/              # PDF processing (modular design)
│   │   ├── metadata.go   # Info, page count, metadata
│   │   ├── transform.go  # Merge, split, rotate, compress
│   │   ├── encryption.go # Encrypt, decrypt
│   │   ├── text.go       # Text extraction
│   │   ├── watermark.go  # Watermarking
│   │   └── validation.go # PDF/A validation
│   ├── pdferrors/        # Error handling with context
│   ├── progress/         # Progress bar utilities
│   └── testing/          # Test infrastructure and mocks
├── docs/
│   └── architecture.md   # Architecture documentation
├── testdata/             # Test PDF files
├── .github/workflows/    # CI/CD pipelines
├── Makefile              # Build automation
├── CONTRIBUTING.md       # Contribution guidelines
└── README.md
```

For detailed architecture information, see [docs/architecture.md](docs/architecture.md).

## Troubleshooting

### "command not found: pdf"

Make sure your Go bin directory is in your PATH:

```bash
export PATH=$PATH:$(go env GOPATH)/bin
```

Add this line to your `~/.bashrc`, `~/.zshrc`, or equivalent.

### "failed to open file: permission denied"

Check file permissions:

```bash
ls -la document.pdf
chmod 644 document.pdf  # Make readable
```

### "encrypted PDF requires password"

The PDF is password-protected. Use the `--password` flag:

```bash
pdf info document.pdf --password yourpassword
```

### "no text extracted" from a PDF

Some PDFs contain scanned images instead of actual text. Use the `--ocr` flag to extract text using OCR:

```bash
pdf text scanned.pdf --ocr
```

The OCR engine automatically uses native Tesseract if installed, or falls back to the built-in WASM version.

### Native Tesseract not detected

If you have Tesseract installed but pdf-cli doesn't detect it:

```bash
# Check if Tesseract is in PATH
tesseract --version

# Force native backend to see the error
pdf text scanned.pdf --ocr --ocr-backend=native -v
```

Common solutions:
- Ensure `tesseract` is in your PATH
- Set `TESSDATA_PREFIX` to your tessdata directory
- Install Tesseract: `brew install tesseract` (macOS) or `apt install tesseract-ocr` (Linux)

### WASM OCR tessdata download

The first time you use WASM OCR, pdf-cli will download the required language data (~15MB for English).

### Large PDF processing is slow

For very large PDFs (hundreds of pages), operations may take time. Use `--progress` to see a progress bar:

```bash
pdf text large.pdf --progress
pdf split large.pdf -o output/ --progress
pdf merge -o combined.pdf *.pdf --progress
```

Note: pdf-cli automatically uses parallel processing for:
- File validation when merging more than 3 files
- Text extraction when processing more than 5 pages
- OCR processing when using native Tesseract backend with more than 5 images

This significantly improves performance for batch operations.

## Contributing

Contributions are welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines.

Quick start:

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Make your changes and add tests
4. Run the full check suite: `make check-all`
5. Commit your changes: `git commit -m 'Add amazing feature'`
6. Push to your fork: `git push origin feature/amazing-feature`
7. Open a Pull Request

Code requirements:
- All tests pass (`make test`)
- Linter passes (`make lint`)
- Coverage meets 75% threshold (`make coverage-check`)
- Documentation updated as needed

## Dependencies

This project uses the following open-source libraries:

- [pdfcpu](https://github.com/pdfcpu/pdfcpu) - PDF processing library
- [ledongthuc/pdf](https://github.com/ledongthuc/pdf) - PDF text extraction
- [gogosseract](https://github.com/danlock/gogosseract) - WASM-based OCR (no external dependencies)
- [progressbar](https://github.com/schollz/progressbar) - Progress bar display
- [cobra](https://github.com/spf13/cobra) - CLI framework
- [yaml.v3](https://gopkg.in/yaml.v3) - YAML configuration parsing

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [pdfcpu](https://github.com/pdfcpu/pdfcpu) for the excellent PDF processing library
- [ledongthuc/pdf](https://github.com/ledongthuc/pdf) for reliable text extraction
- The Go community for great tooling and libraries
