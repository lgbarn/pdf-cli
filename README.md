# pdf-cli

A fast, single-binary CLI tool for common PDF operations.

## Features

- **info** - Display PDF information (page count, metadata, encryption status)
- **merge** - Combine multiple PDFs into one
- **split** - Split a PDF into multiple files
- **extract** - Extract specific pages from a PDF
- **rotate** - Rotate pages in a PDF
- **compress** - Optimize and reduce PDF file size
- **encrypt** - Add password protection to a PDF
- **decrypt** - Remove password protection from a PDF
- **text** - Extract text content from a PDF
- **images** - Extract images from a PDF
- **meta** - View or modify PDF metadata
- **watermark** - Add text or image watermarks

## Installation

### Using Go

```bash
go install github.com/lgbarn/pdf-cli/cmd/pdf@latest
```

### From Source

```bash
git clone https://github.com/lgbarn/pdf-cli.git
cd pdf-cli
make build
```

### Pre-built Binaries

Download the latest release from the [releases page](https://github.com/lgbarn/pdf-cli/releases).

## Usage

### Display PDF Information

```bash
pdf info document.pdf
```

### Merge PDFs

```bash
pdf merge -o combined.pdf file1.pdf file2.pdf file3.pdf
```

### Split PDF into Individual Pages

```bash
pdf split document.pdf -o output/
```

### Split PDF into Chunks

```bash
pdf split document.pdf -n 5 -o chunks/  # 5 pages per file
```

### Extract Specific Pages

```bash
pdf extract document.pdf -p 1-5,10,15-20 -o selected.pdf
```

### Rotate Pages

```bash
pdf rotate document.pdf -a 90 -o rotated.pdf
pdf rotate document.pdf -a 180 -p 1-5 -o rotated.pdf  # Rotate specific pages
```

### Compress PDF

```bash
pdf compress large.pdf -o smaller.pdf
```

### Encrypt PDF

```bash
pdf encrypt document.pdf --password secret -o secure.pdf
pdf encrypt document.pdf --password user123 --owner-password admin456 -o secure.pdf
```

### Decrypt PDF

```bash
pdf decrypt secure.pdf --password secret -o unlocked.pdf
```

### Extract Text

```bash
pdf text document.pdf                    # Print to stdout
pdf text document.pdf -o content.txt     # Save to file
pdf text document.pdf -p 1-5 -o ch1.txt  # Specific pages
```

### Extract Images

```bash
pdf images document.pdf -o images/
pdf images document.pdf -p 1-5 -o images/
```

### View/Set Metadata

```bash
pdf meta document.pdf                                        # View metadata
pdf meta document.pdf --title "My Doc" -o updated.pdf        # Set metadata
pdf meta document.pdf --author "John" --subject "Report" -o updated.pdf
```

### Add Watermark

```bash
pdf watermark document.pdf -t "CONFIDENTIAL" -o marked.pdf   # Text watermark
pdf watermark document.pdf -i logo.png -o branded.pdf        # Image watermark
pdf watermark document.pdf -t "DRAFT" -p 1-5 -o draft.pdf    # Specific pages
```

## Global Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--verbose` | `-v` | Enable verbose output |
| `--force` | `-f` | Overwrite existing files without prompting |
| `--help` | `-h` | Show help for any command |
| `--version` | | Show version information |

## Shell Completion

Generate shell completion scripts:

```bash
# Bash
pdf completion bash > /etc/bash_completion.d/pdf

# Zsh
pdf completion zsh > "${fpath[1]}/_pdf"

# Fish
pdf completion fish > ~/.config/fish/completions/pdf.fish

# PowerShell
pdf completion powershell > pdf.ps1
```

## Building

### Prerequisites

- Go 1.21 or later

### Build Commands

```bash
make build          # Build for current platform
make build-all      # Build for all platforms
make test           # Run tests
make test-coverage  # Run tests with coverage report
make clean          # Clean build artifacts
```

## Project Structure

```
pdf-cli/
├── cmd/
│   └── pdf/
│       └── main.go           # Entry point
├── internal/
│   ├── cli/
│   │   ├── cli.go            # Root command setup
│   │   ├── completion.go     # Shell completion
│   │   └── flags.go          # Common flags
│   ├── commands/
│   │   ├── info.go           # info command
│   │   ├── merge.go          # merge command
│   │   ├── split.go          # split command
│   │   ├── extract.go        # extract command
│   │   ├── rotate.go         # rotate command
│   │   ├── compress.go       # compress command
│   │   ├── encrypt.go        # encrypt command
│   │   ├── decrypt.go        # decrypt command
│   │   ├── text.go           # text command
│   │   ├── images.go         # images command
│   │   ├── meta.go           # meta command
│   │   └── watermark.go      # watermark command
│   ├── pdf/
│   │   └── pdf.go            # pdfcpu wrapper
│   └── util/
│       ├── errors.go         # Error handling
│       ├── files.go          # File utilities
│       └── pages.go          # Page range parsing
├── scripts/
│   └── build.sh              # Cross-compilation script
├── Makefile
├── go.mod
├── go.sum
├── .goreleaser.yaml
├── LICENSE
└── README.md
```

## Dependencies

- [pdfcpu](https://github.com/pdfcpu/pdfcpu) - PDF processing library
- [cobra](https://github.com/spf13/cobra) - CLI framework
- [pflag](https://github.com/spf13/pflag) - POSIX-compliant flags

## License

MIT License - see [LICENSE](LICENSE) for details.

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Acknowledgments

- [pdfcpu](https://github.com/pdfcpu/pdfcpu) for the excellent PDF processing library
