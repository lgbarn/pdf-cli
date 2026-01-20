# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 1.3.x   | :white_check_mark: |
| 1.2.x   | :white_check_mark: |
| < 1.2   | :x:                |

## Reporting a Vulnerability

If you discover a security vulnerability in pdf-cli, please report it responsibly:

1. **Do not** open a public GitHub issue for security vulnerabilities
2. Email the maintainer directly or use [GitHub's private vulnerability reporting](https://github.com/lgbarn/pdf-cli/security/advisories/new)
3. Include:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if any)

## Response Timeline

- **Initial response**: Within 48 hours
- **Status update**: Within 7 days
- **Fix timeline**: Depends on severity (critical: ASAP, high: 14 days, medium: 30 days)

## Security Considerations

### File Handling
- pdf-cli processes files provided by the user
- Temporary files are created in system temp directories and cleaned up after use
- No files are sent to external services (except tessdata downloads for OCR)

### Password Handling
- Passwords for encrypted PDFs are passed via command-line flags
- Passwords are not logged or stored
- Consider using environment variables or stdin for sensitive passwords in scripts

### OCR Data Downloads
- WASM OCR backend downloads tessdata from `https://github.com/tesseract-ocr/tessdata_fast`
- Downloads occur only on first use of a language
- Files are stored in user config directory (`~/.config/pdf-cli/tessdata/`)

### Dependencies
- Dependencies are monitored via GitHub Dependabot
- Security updates are applied promptly
