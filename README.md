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
- **Scriptable**: Perfect for automation and batch processing

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

# Get PDF info
pdf info document.pdf
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

| Command | Description |
|---------|-------------|
| `info` | Display PDF information (pages, metadata, encryption status) |
| `merge` | Combine multiple PDFs into a single file |
| `split` | Split a PDF into individual pages or chunks |
| `extract` | Extract specific pages into a new PDF |
| `rotate` | Rotate pages by 90, 180, or 270 degrees |
| `compress` | Optimize and reduce PDF file size |
| `encrypt` | Add password protection to a PDF |
| `decrypt` | Remove password protection from a PDF |
| `text` | Extract text content from a PDF |
| `images` | Extract embedded images from a PDF |
| `meta` | View or modify PDF metadata (title, author, etc.) |
| `watermark` | Add text or image watermarks |

## Usage Examples

### Get PDF Information

```bash
pdf info document.pdf
```

Output:
```
File:       document.pdf
Size:       2.45 MB
Pages:      42
Version:    1.7
Title:      Annual Report
Author:     John Doe
Encrypted:  No
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

### Rotate Pages

```bash
# Rotate all pages 90 degrees clockwise
pdf rotate document.pdf -a 90 -o rotated.pdf

# Rotate only pages 1-5 by 180 degrees
pdf rotate document.pdf -a 180 -p 1-5 -o rotated.pdf
```

### Compress a PDF

```bash
pdf compress large.pdf -o smaller.pdf
```

### Encrypt a PDF

```bash
# Add password protection
pdf encrypt document.pdf --password mysecret -o secure.pdf

# Set separate user and owner passwords
pdf encrypt document.pdf --password userpass --owner-password ownerpass -o secure.pdf
```

### Decrypt a PDF

```bash
pdf decrypt secure.pdf --password mysecret -o unlocked.pdf
```

### Extract Text

```bash
# Print text to terminal
pdf text document.pdf

# Save to a file
pdf text document.pdf -o content.txt

# Extract text from specific pages
pdf text document.pdf -p 1-5 -o chapter1.txt
```

### Extract Images

```bash
# Extract all images
pdf images document.pdf -o images/

# Extract images from specific pages
pdf images document.pdf -p 1-10 -o images/
```

### View and Modify Metadata

```bash
# View metadata
pdf meta document.pdf

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
```

## Global Options

These options work with all commands:

| Option | Short | Description |
|--------|-------|-------------|
| `--verbose` | `-v` | Show detailed output during operations |
| `--force` | `-f` | Overwrite existing files without prompting |
| `--password` | `-P` | Password for encrypted input PDFs |
| `--help` | `-h` | Show help for any command |
| `--version` | | Display version information |

### Working with Encrypted PDFs

Most commands accept a `--password` flag for reading encrypted PDFs:

```bash
pdf info secure.pdf --password mysecret
pdf extract secure.pdf -p 1-5 -o pages.pdf --password mysecret
```

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
├── cmd/pdf/           # Application entry point
├── internal/
│   ├── cli/           # CLI framework and flags
│   ├── commands/      # Individual command implementations
│   ├── pdf/           # PDF processing wrapper
│   └── util/          # Utilities (errors, files, page parsing)
├── testdata/          # Test PDF files
├── .github/workflows/ # CI/CD pipelines
├── Makefile           # Build automation
└── README.md
```

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

Some PDFs contain scanned images instead of actual text. The `text` command only extracts embedded text, not OCR. For image-based PDFs, you'll need an OCR tool.

### Large PDF processing is slow

For very large PDFs (hundreds of pages), operations may take time. Use `--verbose` to see progress:

```bash
pdf compress large.pdf -o smaller.pdf --verbose
```

Note: pdf-cli automatically uses parallel processing for:
- File validation when merging more than 3 files
- Text extraction when processing more than 5 pages

This significantly improves performance for batch operations.

## Contributing

Contributions are welcome! Here's how to get started:

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Make your changes and add tests
4. Run the test suite: `make test`
5. Commit your changes: `git commit -m 'Add amazing feature'`
6. Push to your fork: `git push origin feature/amazing-feature`
7. Open a Pull Request

Please ensure your code:
- Passes all existing tests
- Includes tests for new functionality
- Follows the existing code style
- Updates documentation as needed

## Dependencies

This project uses the following open-source libraries:

- [pdfcpu](https://github.com/pdfcpu/pdfcpu) - PDF processing library
- [ledongthuc/pdf](https://github.com/ledongthuc/pdf) - PDF text extraction
- [cobra](https://github.com/spf13/cobra) - CLI framework

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [pdfcpu](https://github.com/pdfcpu/pdfcpu) for the excellent PDF processing library
- [ledongthuc/pdf](https://github.com/ledongthuc/pdf) for reliable text extraction
- The Go community for great tooling and libraries
