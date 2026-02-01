# Phase 3: Security Hardening - Research Document

**Date**: 2026-01-31
**Phase**: Phase 3 - Security Hardening
**Priority**: P0 (Critical Security Issues)

## Executive Summary

This phase addresses three critical security vulnerabilities in pdf-cli:

1. Password exposure via command-line flags (process listings, shell history)
2. Missing checksum verification for downloaded tessdata files (supply chain risk)
3. Inconsistent path sanitization for user-provided file paths (path traversal risk)

All three issues are classified as P0 in the codebase concerns analysis and require immediate remediation.

---

## R1: Password Handling Security

### Current Implementation

**Location**: `/internal/cli/flags.go`, `/internal/commands/encrypt.go`, `/internal/commands/decrypt.go`

**Current Code**:
```go
// internal/cli/flags.go:30-36
func AddPasswordFlag(cmd *cobra.Command, usage string) {
    if usage == "" {
        usage = "Password for encryption/decryption"
    }
    cmd.Flags().String("password", "", usage)
}

// internal/commands/encrypt.go:18-21
cli.AddPasswordFlag(encryptCmd, "User password (required)")
_ = encryptCmd.MarkFlagRequired("password")

// internal/commands/decrypt.go:18-20
cli.AddPasswordFlag(decryptCmd, "Password for the encrypted PDF (required)")
_ = decryptCmd.MarkFlagRequired("password")
```

**Current Usage Pattern**:
```bash
# INSECURE - Password visible in process list and shell history
pdf encrypt document.pdf --password mysecret -o secure.pdf
pdf decrypt secure.pdf --password mysecret -o unlocked.pdf
```

**Commands Accepting Passwords**:
1. `encrypt` - User password (required) and owner password (optional)
2. `decrypt` - Password (required)
3. `merge` - Password for encrypted input PDFs (optional)
4. `extract` - Password for encrypted PDFs (optional)
5. `watermark` - Password for encrypted PDFs (optional)
6. `info` - Password for encrypted PDFs (optional)
7. All other commands accept `--password` via `cli.AddPasswordFlag()`

### Security Vulnerability

**Exposure Vectors**:
- `ps aux` shows passwords in process arguments
- Shell history files (`.bash_history`, `.zsh_history`) log commands with passwords
- Parent process environment may expose passwords
- System audit logs may capture command-line arguments

**Risk Level**: P0 - Critical
**Impact**: Credential exposure on multi-user systems or shared environments

### Technology Options

#### Option 1: Environment Variable Input
**Library**: Standard library (`os.Getenv`)
**Maturity**: Stable (Go 1.0+)
**Community Support**: Universal pattern

**Pros**:
- No additional dependencies
- Works across all platforms
- Easy to integrate with CI/CD pipelines
- Environment variables not shown in `ps` output

**Cons**:
- Environment variables still visible in `/proc/<pid>/environ` on Linux
- May persist in shell history if set inline: `PASSWORD=secret pdf encrypt ...`
- Less secure than interactive prompt for local usage

**Implementation**:
```go
password := os.Getenv("PDF_CLI_PASSWORD")
if password == "" {
    password, _ = cmd.Flags().GetString("password")
}
```

#### Option 2: Interactive Stdin Prompt
**Library**: `golang.org/x/term` (already a dependency)
**Maturity**: Stable, official Go extended library
**Community Support**: Widely used (1.6k+ dependent projects)

**Pros**:
- Most secure for interactive usage (password never logged)
- Already have `golang.org/x/term v0.39.0` dependency
- No password in process list, history, or environment
- Standard UX pattern (similar to `sudo`, `ssh`, etc.)

**Cons**:
- Not suitable for non-interactive/CI environments
- Requires TTY availability check
- Slightly worse UX for scripting use cases

**Implementation**:
```go
// Already imported in internal/fileio/stdio.go:8
import "golang.org/x/term"

func ReadPassword() (string, error) {
    if !term.IsTerminal(int(os.Stdin.Fd())) {
        return "", fmt.Errorf("password prompt requires interactive terminal")
    }
    fmt.Fprint(os.Stderr, "Enter password: ")
    password, err := term.ReadPassword(int(os.Stdin.Fd()))
    fmt.Fprintln(os.Stderr)
    return string(password), err
}
```

#### Option 3: Password File Input
**Library**: Standard library (`os.ReadFile`)
**Maturity**: Stable (Go 1.16+)
**Community Support**: Common pattern in security tools

**Pros**:
- Secure for automation (file permissions protect password)
- Works in CI/CD pipelines
- No shell history exposure
- Can use secure storage backends (e.g., mounted secrets)

**Cons**:
- Requires file management
- Path to password file may still be visible
- Need to sanitize file paths (see R3)

**Implementation**:
```go
if passwordFile != "" {
    data, err := os.ReadFile(passwordFile) // Need path sanitization
    if err != nil {
        return fmt.Errorf("reading password file: %w", err)
    }
    password = strings.TrimSpace(string(data))
}
```

#### Option 4: Stdin Pipe (Non-Interactive)
**Library**: Standard library (`io`)
**Maturity**: Stable
**Community Support**: Common pattern (e.g., Docker, AWS CLI)

**Pros**:
- Secure for scripting: `echo -n "password" | pdf encrypt --password-stdin ...`
- No history or process list exposure
- No additional dependencies
- Works with secret managers: `vault read -field=password secret | pdf encrypt ...`

**Cons**:
- Requires pipe setup
- May conflict with PDF stdin input (`pdf encrypt - --password-stdin`)
- Need clear UX to distinguish from interactive stdin

**Implementation**:
```go
if passwordStdin {
    if fileio.IsStdinPiped() {
        data, err := io.ReadAll(os.Stdin)
        if err != nil {
            return fmt.Errorf("reading password from stdin: %w", err)
        }
        password = strings.TrimSpace(string(data))
    } else {
        return fmt.Errorf("--password-stdin requires piped input")
    }
}
```

### Recommended Approach

**Hybrid Solution**: Support all four methods with priority order:

1. `--password-stdin` flag (highest priority if set)
2. `PDF_CLI_PASSWORD` environment variable
3. `--password-file <path>` flag
4. Interactive prompt (if TTY available and none of above set)
5. Deprecated: `--password <value>` flag (warn user about security risk)

**Priority Order Justification**:
- Explicit flags (`--password-stdin`, `--password-file`) indicate user intent
- Environment variable as fallback for CI/CD
- Interactive prompt for best UX in terminal sessions
- Deprecate but don't remove `--password` for backward compatibility

**Migration Strategy**:
- Add new password input methods in v1.6.0
- Deprecate `--password` with warning in v1.6.0
- Remove `--password` in v2.0.0

**Example Usage**:
```bash
# Method 1: Stdin pipe (recommended for scripts)
echo -n "mysecret" | pdf encrypt file.pdf --password-stdin -o secure.pdf

# Method 2: Environment variable (CI/CD)
export PDF_CLI_PASSWORD=mysecret
pdf encrypt file.pdf -o secure.pdf

# Method 3: Password file (secure automation)
echo -n "mysecret" > /secure/password.txt
chmod 600 /secure/password.txt
pdf encrypt file.pdf --password-file /secure/password.txt -o secure.pdf

# Method 4: Interactive prompt (best for local use)
pdf encrypt file.pdf -o secure.pdf
# Prompts: Enter password: ****

# Method 5: Deprecated (backward compat with warning)
pdf encrypt file.pdf --password mysecret -o secure.pdf
# Warning: Using --password flag is insecure. Use --password-stdin, environment variable, or interactive prompt.
```

### Implementation Files to Modify

1. `/internal/cli/flags.go` - Add new password flag functions
2. `/internal/cli/password.go` (new) - Password reading logic
3. `/internal/commands/encrypt.go` - Update password retrieval
4. `/internal/commands/decrypt.go` - Update password retrieval
5. `/internal/commands/merge.go` - Update password retrieval
6. `/internal/commands/extract.go` - Update password retrieval
7. `/internal/commands/watermark.go` - Update password retrieval
8. `/internal/commands/info.go` - Update password retrieval (if applicable)
9. All other commands using `cli.AddPasswordFlag()`

### Dependencies

**Required**: None (all methods use existing dependencies)
- `golang.org/x/term v0.39.0` - Already present in `go.mod:11`
- Standard library: `os`, `io`, `strings`, `fmt`

**Verification**:
```go
// go.mod:11
golang.org/x/term v0.39.0
```

Already used in `/internal/fileio/stdio.go:8`:
```go
import "golang.org/x/term"
```

### Risks and Mitigations

| Risk | Mitigation |
|------|-----------|
| Breaking change for scripts using `--password` | Deprecate with warning, remove only in v2.0 |
| Stdin conflict with PDF stdin input | Use explicit `--password-stdin` flag; don't auto-detect |
| TTY not available for interactive prompt | Fall back to error with instructions to use env var or file |
| Password file path traversal | Apply R3 path sanitization before reading |
| Empty password from environment variable | Treat empty string as "not set", require explicit value |

---

## R2: Tessdata Checksum Verification

### Current Implementation

**Location**: `/internal/ocr/ocr.go:169-209`

**Current Code**:
```go
// internal/ocr/ocr.go:23-25
const (
    // TessdataURL is the base URL for downloading tessdata files.
    TessdataURL = "https://github.com/tesseract-ocr/tessdata_fast/raw/main"
)

// internal/ocr/ocr.go:169-209
func downloadTessdata(ctx context.Context, dataDir, lang string) error {
    url := fmt.Sprintf("%s/%s.traineddata", TessdataURL, lang)
    dataFile := filepath.Join(dataDir, lang+".traineddata")

    fmt.Fprintf(os.Stderr, "Downloading tessdata for '%s'...\n", lang)

    ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
    defer cancel()

    req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
    if err != nil {
        return err
    }

    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("failed to download: HTTP %d", resp.StatusCode)
    }

    tmpFile, err := os.CreateTemp(dataDir, "tessdata-*.tmp")
    if err != nil {
        return err
    }
    tmpPath := tmpFile.Name()
    defer os.Remove(tmpPath)

    bar := progress.NewBytesProgressBar(fmt.Sprintf("Downloading %s.traineddata", lang), resp.ContentLength)
    if _, err := io.Copy(io.MultiWriter(tmpFile, bar), resp.Body); err != nil {
        _ = tmpFile.Close()
        return err
    }
    _ = tmpFile.Close()
    progress.FinishProgressBar(bar)

    return os.Rename(tmpPath, dataFile)  // NO CHECKSUM VERIFICATION
}
```

**Security Vulnerability**:
- No integrity verification of downloaded files
- Vulnerable to man-in-the-middle attacks (even with HTTPS)
- Vulnerable to compromised GitHub repository
- Vulnerable to CDN corruption
- No way to detect tampered or corrupted downloads

**Risk Level**: P0 - Critical
**Impact**: Supply chain attack vector, potential arbitrary code execution via malicious traineddata files

### Technology Options

#### Option 1: Embedded SHA256 Checksums
**Library**: Standard library (`crypto/sha256`)
**Maturity**: Stable (Go 1.0+)
**Community Support**: Universal pattern

**Pros**:
- No external dependencies
- Works offline once checksums are embedded
- Fast verification (streaming hash calculation)
- No network calls for checksum retrieval

**Cons**:
- Checksums must be updated when tessdata files change
- Requires manual tracking of upstream changes
- Larger binary size (minimal - ~100 bytes per language)

**Implementation**:
```go
var tessdataChecksums = map[string]string{
    "eng": "sha256:...",
    "fra": "sha256:...",
    // ... more languages
}

func verifyChecksum(filePath string, expectedHash string) error {
    f, err := os.Open(filePath)
    if err != nil {
        return err
    }
    defer f.Close()

    h := sha256.New()
    if _, err := io.Copy(h, f); err != nil {
        return err
    }

    actualHash := hex.EncodeToString(h.Sum(nil))
    if actualHash != expectedHash {
        return fmt.Errorf("checksum mismatch: expected %s, got %s", expectedHash, actualHash)
    }
    return nil
}
```

#### Option 2: Fetch Checksums from GitHub API
**Library**: Standard library (`net/http`, `encoding/json`)
**Maturity**: Stable
**Community Support**: Common pattern

**Pros**:
- Always up-to-date with repository
- No manual checksum updates needed
- Can use GitHub's commit SHA for verification

**Cons**:
- Requires additional network call
- GitHub API rate limiting (60 req/hour unauthenticated)
- Depends on GitHub API availability
- More complex implementation

**Implementation**:
```go
// Fetch file SHA from GitHub API
type GitHubContent struct {
    SHA string `json:"sha"`
}

func fetchGitHubSHA(lang string) (string, error) {
    url := fmt.Sprintf("https://api.github.com/repos/tesseract-ocr/tessdata_fast/contents/%s.traineddata", lang)
    resp, err := http.Get(url)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    var content GitHubContent
    if err := json.NewDecoder(resp.Body).Decode(&content); err != nil {
        return "", err
    }
    return content.SHA, nil
}
```

#### Option 3: Checksum File in Repository
**Library**: Standard library
**Maturity**: Stable
**Community Support**: Common pattern (Linux distros use this)

**Pros**:
- Single source of truth
- Easy to update (just update checksum file)
- Can verify multiple files at once
- Standard format (SHA256SUMS)

**Cons**:
- Checksums file itself needs verification (chicken-and-egg)
- Requires parsing checksum file format
- Additional file to maintain

**Implementation**:
```go
// .shipyard/data/tessdata_checksums.txt format:
// <sha256>  <filename>
// Parse and verify against this embedded file
```

### Recommended Approach

**Embedded SHA256 Checksums** (Option 1) with the following features:

1. Embed checksums in source code for common languages
2. Calculate checksum after download and before file rename
3. Delete corrupted download if checksum fails
4. Provide clear error message with expected vs actual hash
5. Add `--skip-checksum-verification` flag for advanced users (with warning)

**Justification**:
- Simplest implementation (no additional network calls)
- Works offline after first download
- Most reliable (no external dependencies)
- Fast (streaming hash calculation during download)
- Security best practice (defense in depth)

**Checksum Strategy**:
- Manually compute checksums for common languages (eng, fra, deu, spa, etc.)
- Store in embedded map in source code
- Update checksums when tessdata_fast repository releases new versions
- Document checksum verification process in SECURITY.md

### Obtaining SHA256 Checksums

**Research Finding**: The tesseract-ocr/tessdata_fast repository does NOT publish official SHA256 checksums.

**Sources Consulted**:
- [GitHub tessdata_fast repository](https://github.com/tesseract-ocr/tessdata_fast)
- [Tesseract documentation](https://tesseract-ocr.github.io/tessdoc/Data-Files.html)
- [tessdata_fast releases page](https://github.com/tesseract-ocr/tessdata_fast/releases)

**Finding**: No official checksums, verification files, or hash documentation found.

**Manual Checksum Generation Required**:
```bash
# Download and compute checksums for common languages
cd /tmp
wget https://github.com/tesseract-ocr/tessdata_fast/raw/main/eng.traineddata
wget https://github.com/tesseract-ocr/tessdata_fast/raw/main/fra.traineddata
wget https://github.com/tesseract-ocr/tessdata_fast/raw/main/deu.traineddata
wget https://github.com/tesseract-ocr/tessdata_fast/raw/main/spa.traineddata

sha256sum *.traineddata
# Output format: <hash> <filename>
```

**Implementation Plan**:
1. Manually download top 10 most common languages from tessdata_fast
2. Compute SHA256 checksums locally
3. Embed checksums in source code with generation date and commit SHA
4. Add comment with verification command for reproducibility
5. Update checksums when tessdata_fast releases new versions

**Checksum Data Structure**:
```go
// internal/ocr/checksums.go (new file)
package ocr

// TessdataChecksums maps language codes to their SHA256 checksums.
// Generated from tesseract-ocr/tessdata_fast@main on 2026-01-31
// Verify with: sha256sum <file>
var TessdataChecksums = map[string]string{
    "eng": "7d4322bd2a7ca61d688c94066f7e0d978d8879ecea26eb71955d7e34fe56e098",  // eng.traineddata
    "fra": "...",  // To be computed
    "deu": "...",  // To be computed
    "spa": "...",  // To be computed
    "ita": "...",  // To be computed
    "por": "...",  // To be computed
    "rus": "...",  // To be computed
    "jpn": "...",  // To be computed
    "chi_sim": "...",  // To be computed
    "chi_tra": "...",  // To be computed
}

// NOTE: The above eng hash is an example. Real checksums must be computed
// by downloading files from the official repository.
```

### Implementation Files to Modify

1. `/internal/ocr/checksums.go` (new) - Embedded checksum map
2. `/internal/ocr/ocr.go` - Update `downloadTessdata()` to verify checksums
3. `/internal/ocr/verify.go` (new) - Checksum verification logic
4. `.shipyard/scripts/update_tessdata_checksums.sh` (new) - Script to update checksums

### Dependencies

**Required**: None (uses standard library)
- `crypto/sha256` - Standard library
- `encoding/hex` - Standard library
- `io` - Standard library

### Risks and Mitigations

| Risk | Mitigation |
|------|-----------|
| Checksums become outdated | Document update process, add CI check for new releases |
| Unknown language requested | Skip verification with warning, or fail and suggest contribution |
| Upstream file changes without notice | Add version/commit tracking, document expected version |
| False positive checksum mismatch | Provide clear error with instructions to report issue |
| User needs to skip verification | Add `--skip-checksum` flag with security warning |

### Implementation Considerations

**Download Flow**:
```
1. Check if file exists locally -> Skip download
2. Download to temp file with progress bar
3. Calculate SHA256 while downloading (streaming)
4. Compare against embedded checksum
5. If match: rename temp to final location
6. If mismatch: delete temp, return error
7. If no checksum for language: warn and continue (or fail based on flag)
```

**Error Messages**:
```
ERROR: Checksum verification failed for eng.traineddata
  Expected: 7d4322bd2a7ca61d688c94066f7e0d978d8879ecea26eb71955d7e34fe56e098
  Got:      1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef

This may indicate:
- Network corruption during download
- Man-in-the-middle attack
- Corrupted upstream file

Please retry the download. If the issue persists, file a bug report at:
https://github.com/lgbarn/pdf-cli/issues
```

**Performance Impact**:
- Minimal: SHA256 calculation is ~300 MB/s on modern CPUs
- Tessdata files are typically 1-50 MB
- Adds <0.2 seconds to download time
- Acceptable tradeoff for security

---

## R3: Path Traversal Sanitization

### Current Implementation

**Location**: Multiple files, inconsistent sanitization

**Partial Sanitization** (Good Example):
```go
// internal/fileio/files.go:86-94
func CopyFile(src, dst string) error {
    // Clean paths to prevent directory traversal
    cleanSrc := filepath.Clean(src)
    cleanDst := filepath.Clean(dst)

    srcFile, err := os.Open(cleanSrc) // #nosec G304 -- path is cleaned
    // ...
}
```

**Missing Sanitization** (Examples Found):
```go
// internal/config/config.go:86
data, err := os.ReadFile(path) // #nosec G304 - path is from XDG config, not user input
// Comment acknowledges risk but doesn't sanitize

// internal/fileio/stdio.go:59
f, err := os.Open(path) // #nosec G304 -- path comes from temp files we control
// Assumes path is safe but doesn't validate

// internal/commands/patterns/stdio.go:37
tmpFile, err := os.CreateTemp("", "pdf-cli-"+h.Operation+"-*.pdf")
// Operation field comes from string literal, but not validated
```

**File Operations Audit**:
Files using `os.Open`, `os.Create`, `os.ReadFile`, `os.WriteFile`:
1. `/internal/fileio/files.go` - Partially sanitized
2. `/internal/fileio/stdio.go` - Trusts temp file paths
3. `/internal/config/config.go` - Trusts config paths
4. `/internal/ocr/native.go` - File operations on tessdata
5. `/internal/ocr/wasm.go` - File operations on tessdata
6. `/internal/ocr/ocr.go` - Temp files and downloads
7. `/internal/pdf/metadata.go` - PDF operations
8. `/internal/pdf/text.go` - PDF operations
9. `/internal/pdf/transform.go` - PDF operations
10. `/internal/pdf/validation.go` - PDF operations
11. `/internal/testing/fixtures.go` - Test fixtures
12. `/internal/commands/patterns/stdio.go` - Temp files
13. `/internal/commands/text.go` - User-provided paths

### User-Controlled Path Entry Points

**Command Arguments** (Direct User Input):
1. All commands accept PDF file paths as arguments
2. `--output` / `-o` flag - Output file path
3. `--password-file` flag (R1) - Password file path
4. `--image` / `-i` flag (watermark) - Image file path
5. Image paths in `combine-images` command

**Indirect User Input**:
1. Config file path from `XDG_CONFIG_HOME` environment variable
2. Tessdata directory path (though user-controlled via config)
3. Temp file names (controlled by application, safe)

### Security Vulnerability

**Path Traversal Risks**:
```bash
# Potential attack vectors
pdf encrypt ../../../etc/passwd -o output.pdf  # Read sensitive files
pdf merge input.pdf -o ../../bin/malicious     # Write to system paths
pdf watermark doc.pdf -i ../../../etc/shadow   # Access restricted files
pdf decrypt file.pdf --password-file ../../.ssh/id_rsa  # Read SSH keys (R1)
```

**Current Protections**:
- `filepath.Clean()` used in some places (removes `..` and redundant separators)
- `#nosec G304` annotations acknowledge risk but rely on assumptions

**Gaps**:
- `filepath.Clean()` is NOT sufficient alone - it normalizes but doesn't validate
- No validation that paths stay within expected directories
- No checks for absolute paths when relative expected
- Inconsistent application across codebase

**Risk Level**: P0 - Critical
**Impact**: Arbitrary file read/write, potential privilege escalation

### Technology Options

#### Option 1: filepath.Clean + Validation
**Library**: Standard library (`path/filepath`)
**Maturity**: Stable (Go 1.0+)
**Community Support**: Universal

**Pros**:
- No additional dependencies
- Fast and simple
- Handles OS-specific path separators

**Cons**:
- `filepath.Clean()` alone is insufficient
- Need additional validation logic
- Doesn't prevent absolute paths

**Implementation**:
```go
func SanitizePath(userPath string) (string, error) {
    // Clean the path (removes .., ., redundant separators)
    cleaned := filepath.Clean(userPath)

    // Check for absolute paths (may or may not be desired)
    if filepath.IsAbs(cleaned) {
        return "", fmt.Errorf("absolute paths not allowed: %s", userPath)
    }

    // Check for path traversal attempts
    if strings.Contains(cleaned, "..") {
        return "", fmt.Errorf("path traversal detected: %s", userPath)
    }

    return cleaned, nil
}
```

**Note**: Checking for `..` after `filepath.Clean()` is redundant - Clean removes them.
Better approach is to validate against a base directory.

#### Option 2: Secure Join Pattern
**Library**: Standard library (`path/filepath`)
**Maturity**: Stable
**Community Support**: Recommended by OWASP

**Pros**:
- Prevents traversal outside base directory
- Works with relative paths
- Simple to implement
- Language-agnostic pattern

**Cons**:
- Requires defining base directory for each operation
- Doesn't work well for absolute paths
- May restrict legitimate use cases

**Implementation**:
```go
func SecureJoin(base, userPath string) (string, error) {
    // Join base and user path
    joined := filepath.Join(base, userPath)

    // Clean to resolve . and ..
    cleaned := filepath.Clean(joined)

    // Verify result is still under base directory
    if !strings.HasPrefix(cleaned, filepath.Clean(base)) {
        return "", fmt.Errorf("path traversal detected: %s", userPath)
    }

    return cleaned, nil
}
```

**Note**: This only works for paths that SHOULD be under a base directory (e.g., temp files, config files).
For user-provided PDF paths, we typically want to allow arbitrary locations.

#### Option 3: Symlink Resolution + Validation
**Library**: Standard library (`filepath.EvalSymlinks`)
**Maturity**: Stable (Go 1.0+)
**Community Support**: Used in security-sensitive code

**Pros**:
- Resolves symlinks to actual paths
- Prevents symlink-based traversal attacks
- Can validate against allowed directories after resolution

**Cons**:
- Requires file to exist (fails on output paths)
- Performance overhead (filesystem calls)
- May break legitimate symlink usage
- Different behavior on Windows

**Implementation**:
```go
func ValidatePath(userPath string, allowedPrefixes []string) (string, error) {
    // Resolve symlinks
    resolved, err := filepath.EvalSymlinks(userPath)
    if err != nil {
        if os.IsNotExist(err) {
            // File doesn't exist yet (e.g., output path)
            // Validate parent directory instead
            dir := filepath.Dir(userPath)
            resolved, err = filepath.EvalSymlinks(dir)
            if err != nil {
                return "", err
            }
            resolved = filepath.Join(resolved, filepath.Base(userPath))
        } else {
            return "", err
        }
    }

    // Validate against allowed prefixes
    for _, prefix := range allowedPrefixes {
        if strings.HasPrefix(resolved, prefix) {
            return resolved, nil
        }
    }

    return "", fmt.Errorf("path outside allowed directories: %s", userPath)
}
```

#### Option 4: Input Validation Only (No Restriction)
**Library**: Standard library
**Maturity**: Stable
**Community Support**: Common for CLI tools

**Pros**:
- Doesn't restrict legitimate use cases
- Simple to implement
- Works with any path user has access to
- Respects OS permissions

**Cons**:
- Relies on OS permissions for security
- Doesn't prevent user mistakes
- May allow access to sensitive files user owns

**Implementation**:
```go
func ValidateInputPath(path string) error {
    // Clean path
    cleaned := filepath.Clean(path)

    // Check if readable
    info, err := os.Stat(cleaned)
    if err != nil {
        return fmt.Errorf("cannot access path: %w", err)
    }

    // Validate it's a file, not a directory
    if info.IsDir() {
        return fmt.Errorf("path is a directory: %s", path)
    }

    return nil
}

func ValidateOutputPath(path string) error {
    // Clean path
    cleaned := filepath.Clean(path)

    // Check if directory exists and is writable
    dir := filepath.Dir(cleaned)
    if _, err := os.Stat(dir); err != nil {
        return fmt.Errorf("output directory does not exist: %w", err)
    }

    // Check if file already exists (let command decide if overwrite is OK)
    if _, err := os.Stat(cleaned); err == nil {
        // File exists - return info, caller decides
        return nil
    }

    return nil
}
```

### Recommended Approach

**Tiered Validation** based on path context:

1. **User-Provided Input/Output Paths** (PDF files, images):
   - Use `filepath.Clean()` for normalization
   - No restriction on location (respect OS permissions)
   - Validate file exists (input) or directory exists (output)
   - Validate file type (extension check)
   - Trust OS-level permissions for access control

2. **Internal Paths** (temp files, config files):
   - Use `SecureJoin()` pattern to enforce base directory
   - Validate paths stay within expected directories
   - Use `filepath.EvalSymlinks()` for sensitive operations

3. **Password Files** (R1 feature):
   - Use `filepath.Clean()` for normalization
   - Validate file permissions (should be 0600 or similar)
   - Warn if permissions are too permissive

**Justification**:
- PDF CLI is a user tool, not a service - users should be able to access their own files
- OS permissions provide security boundary (can't read files user doesn't have access to)
- Primary risk is user mistakes, not malicious input (CLI tools trust their users)
- Validation focuses on preventing errors, not security isolation
- Internal paths (temp, config) need stricter controls to prevent exploitation

**Centralized Validation**:
Create `/internal/fileio/validation.go` with standardized functions:
- `ValidateInputFile(path string) (string, error)` - For input PDFs/images
- `ValidateOutputFile(path string) (string, error)` - For output PDFs
- `SecureJoin(base, path string) (string, error)` - For internal paths
- `ValidatePasswordFile(path string) (string, error)` - For R1 password files

### Implementation Files to Modify

**New Files**:
1. `/internal/fileio/validation.go` - Centralized path validation

**Files to Update** (apply validation consistently):
1. `/internal/commands/encrypt.go` - Input/output paths
2. `/internal/commands/decrypt.go` - Input/output paths
3. `/internal/commands/merge.go` - Input/output paths
4. `/internal/commands/watermark.go` - Input/output/image paths
5. `/internal/commands/combine_images.go` - Image paths
6. `/internal/commands/extract.go` - Input/output paths
7. `/internal/commands/split.go` - Input/output paths
8. `/internal/commands/compress.go` - Input/output paths
9. `/internal/commands/rotate.go` - Input/output paths
10. `/internal/commands/reorder.go` - Input/output paths
11. `/internal/commands/meta.go` - Input paths
12. `/internal/commands/text.go` - Input paths
13. `/internal/commands/info.go` - Input paths
14. `/internal/commands/pdfa.go` - Input/output paths
15. `/internal/fileio/files.go` - Update CopyFile and other functions
16. `/internal/config/config.go` - Config file path validation
17. `/internal/ocr/ocr.go` - Tessdata directory validation
18. `/internal/cli/password.go` (R1) - Password file validation

### Dependencies

**Required**: None (uses standard library)
- `path/filepath` - Standard library
- `os` - Standard library
- `strings` - Standard library

### Risks and Mitigations

| Risk | Mitigation |
|------|-----------|
| Breaking legitimate use cases | Don't restrict user file access, only validate |
| Inconsistent application | Centralize in fileio package, enforce via linting |
| Symlink attacks | Use EvalSymlinks for sensitive paths only |
| Windows path handling | Use filepath package (OS-agnostic) |
| User mistakes (wrong path) | Provide clear error messages with suggestions |
| Password file permissions | Warn if permissions too permissive (R1) |

### Implementation Considerations

**Validation Functions**:
```go
// internal/fileio/validation.go
package fileio

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"
)

// ValidateInputFile validates a user-provided input file path.
// Returns cleaned path or error if file doesn't exist or isn't readable.
func ValidateInputFile(path string) (string, error) {
    cleaned := filepath.Clean(path)
    info, err := os.Stat(cleaned)
    if err != nil {
        if os.IsNotExist(err) {
            return "", fmt.Errorf("file not found: %s", path)
        }
        return "", fmt.Errorf("cannot access file: %w", err)
    }
    if info.IsDir() {
        return "", fmt.Errorf("path is a directory, not a file: %s", path)
    }
    return cleaned, nil
}

// ValidateOutputFile validates a user-provided output file path.
// Returns cleaned path or error if directory doesn't exist or isn't writable.
func ValidateOutputFile(path string) (string, error) {
    cleaned := filepath.Clean(path)
    dir := filepath.Dir(cleaned)

    // Check if directory exists
    if _, err := os.Stat(dir); err != nil {
        if os.IsNotExist(err) {
            return "", fmt.Errorf("output directory does not exist: %s", dir)
        }
        return "", fmt.Errorf("cannot access output directory: %w", err)
    }

    return cleaned, nil
}

// SecureJoin joins base and user path, ensuring result stays under base directory.
// Returns error if path traversal is detected.
func SecureJoin(base, userPath string) (string, error) {
    base = filepath.Clean(base)
    joined := filepath.Join(base, userPath)
    cleaned := filepath.Clean(joined)

    // Ensure cleaned path is still under base
    if !strings.HasPrefix(cleaned+string(filepath.Separator), base+string(filepath.Separator)) {
        return "", fmt.Errorf("path traversal detected: %s", userPath)
    }

    return cleaned, nil
}

// ValidatePasswordFile validates a password file path and checks permissions.
// Warns if file has insecure permissions.
func ValidatePasswordFile(path string) (string, error) {
    cleaned, err := ValidateInputFile(path)
    if err != nil {
        return "", err
    }

    // Check file permissions (warn if too permissive)
    info, _ := os.Stat(cleaned)
    mode := info.Mode().Perm()

    // On Unix, warn if readable by group or others
    if mode & 0077 != 0 {
        fmt.Fprintf(os.Stderr, "Warning: password file has insecure permissions: %o\n", mode)
        fmt.Fprintf(os.Stderr, "Recommended: chmod 600 %s\n", path)
    }

    return cleaned, nil
}
```

**Testing Strategy**:
```go
// internal/fileio/validation_test.go
func TestValidateInputFile(t *testing.T) {
    // Test cases:
    // - Normal file: should pass
    // - Non-existent file: should fail
    // - Directory: should fail
    // - Path with ..: should normalize but still validate
    // - Symlink: should follow and validate target
}

func TestSecureJoin(t *testing.T) {
    // Test cases:
    // - Normal path: should join
    // - Path with ..: should fail
    // - Absolute path: should fail
    // - Path with symlink outside base: should fail
}

func TestValidatePasswordFile(t *testing.T) {
    // Test cases:
    // - File with 0600 perms: should pass
    // - File with 0644 perms: should warn but pass
    // - Non-existent file: should fail
}
```

---

## Cross-Cutting Concerns

### Testing Strategy

**Unit Tests**:
- Password input methods (interactive, env, file, stdin)
- Checksum verification (valid, invalid, missing)
- Path validation (normal, traversal, symlinks)

**Integration Tests**:
- End-to-end encrypt/decrypt with all password methods
- Download and verify tessdata with checksums
- File operations with various path types

**Security Tests**:
- Attempt path traversal attacks
- Attempt checksum bypass
- Verify password not in process list

### Documentation Updates

**Files to Update**:
1. `README.md` - Document new password input methods
2. `SECURITY.md` - Document security improvements
3. `CHANGELOG.md` - Document breaking changes
4. `docs/` - Add security best practices guide

**Example Documentation**:
```markdown
## Secure Password Handling

pdf-cli supports multiple secure methods for providing passwords:

### Method 1: Interactive Prompt (Recommended for local use)
```bash
pdf encrypt file.pdf -o secure.pdf
# Prompts: Enter password:
```

### Method 2: Stdin Pipe (Recommended for scripts)
```bash
echo -n "mysecret" | pdf encrypt file.pdf --password-stdin -o secure.pdf
```

### Method 3: Environment Variable (CI/CD)
```bash
export PDF_CLI_PASSWORD=mysecret
pdf encrypt file.pdf -o secure.pdf
```

### Method 4: Password File (Secure automation)
```bash
echo -n "mysecret" > /secure/password.txt
chmod 600 /secure/password.txt
pdf encrypt file.pdf --password-file /secure/password.txt -o secure.pdf
```

### Deprecated: --password flag
Using `--password` on the command line is **insecure** and will be removed in v2.0.
```

### Performance Considerations

**Password Input**:
- Interactive prompt: ~0ms (user input time)
- Environment variable: ~0ms (instant)
- File read: <1ms for small password file
- Stdin pipe: <1ms for password input

**Checksum Verification**:
- SHA256 calculation: ~300 MB/s on modern CPUs
- Tessdata files: 1-50 MB
- Added time: <0.2 seconds per download
- Acceptable overhead for security

**Path Validation**:
- `filepath.Clean()`: <0.1ms per path
- `os.Stat()`: <1ms per file
- `EvalSymlinks()`: <5ms per path (filesystem dependent)
- Negligible impact on overall command execution

### Backward Compatibility

**Breaking Changes**:
- None in v1.6.0 (all changes are additions)
- `--password` flag deprecated but still works (with warning)

**Migration Path**:
- v1.6.0: Add new features, deprecate `--password`
- v1.7.0: Make deprecation warnings more prominent
- v2.0.0: Remove `--password` flag entirely

**Compatibility Matrix**:
| Feature | v1.5.x | v1.6.x | v2.0.x |
|---------|--------|--------|--------|
| `--password` | ✓ | ✓ (deprecated) | ✗ |
| `--password-stdin` | ✗ | ✓ | ✓ |
| `--password-file` | ✗ | ✓ | ✓ |
| `PDF_CLI_PASSWORD` | ✗ | ✓ | ✓ |
| Interactive prompt | ✗ | ✓ | ✓ |
| Checksum verification | ✗ | ✓ | ✓ |
| Path validation | Partial | ✓ | ✓ |

---

## Implementation Checklist

### R1: Password Handling
- [ ] Create `/internal/cli/password.go` with password input methods
- [ ] Add `--password-stdin` flag support
- [ ] Add `--password-file` flag support
- [ ] Add `PDF_CLI_PASSWORD` environment variable support
- [ ] Add interactive prompt with `golang.org/x/term`
- [ ] Add deprecation warning to `--password` flag
- [ ] Update all commands to use new password methods
- [ ] Add unit tests for each password input method
- [ ] Add integration tests for encrypt/decrypt
- [ ] Update README.md with password security documentation
- [ ] Update SECURITY.md with password best practices

### R2: Checksum Verification
- [ ] Create `/internal/ocr/checksums.go` with embedded checksums
- [ ] Manually compute SHA256 for common languages (eng, fra, deu, spa, etc.)
- [ ] Create `/internal/ocr/verify.go` with verification logic
- [ ] Update `downloadTessdata()` to verify checksums after download
- [ ] Add `--skip-checksum-verification` flag (with warning)
- [ ] Add checksum mismatch error message with instructions
- [ ] Add unit tests for checksum verification
- [ ] Add integration test for download and verify
- [ ] Create `.shipyard/scripts/update_tessdata_checksums.sh`
- [ ] Document checksum update process in CONTRIBUTING.md
- [ ] Update SECURITY.md with supply chain security info

### R3: Path Sanitization
- [ ] Create `/internal/fileio/validation.go` with validation functions
- [ ] Implement `ValidateInputFile()` function
- [ ] Implement `ValidateOutputFile()` function
- [ ] Implement `SecureJoin()` function
- [ ] Implement `ValidatePasswordFile()` function (for R1)
- [ ] Update all commands to use validation functions
- [ ] Update `CopyFile()` in `/internal/fileio/files.go`
- [ ] Update config loading in `/internal/config/config.go`
- [ ] Update tessdata handling in `/internal/ocr/ocr.go`
- [ ] Add unit tests for validation functions
- [ ] Add integration tests with path traversal attempts
- [ ] Add security tests for symlink attacks
- [ ] Update SECURITY.md with path validation info

### Testing
- [ ] Add unit tests for all new functions (>80% coverage)
- [ ] Add integration tests for end-to-end workflows
- [ ] Add security tests for attack scenarios
- [ ] Verify no regressions with existing test suite
- [ ] Test on Linux, macOS, Windows

### Documentation
- [ ] Update README.md with security features
- [ ] Update SECURITY.md with threat model and mitigations
- [ ] Update CHANGELOG.md with v1.6.0 changes
- [ ] Create security best practices guide in `docs/`
- [ ] Update command help text for new flags

### CI/CD
- [ ] Add gosec security scanning (already present)
- [ ] Add test for password not in process list
- [ ] Add test for checksum verification
- [ ] Add test for path traversal prevention

---

## Estimated Effort

| Component | Files | LOC | Complexity | Effort |
|-----------|-------|-----|------------|--------|
| R1: Password Handling | 10 | ~300 | Medium | 2-3 days |
| R2: Checksum Verification | 3 | ~200 | Low | 1-2 days |
| R3: Path Sanitization | 18 | ~400 | Medium | 2-3 days |
| Testing (all) | 15 | ~600 | Medium | 2-3 days |
| Documentation | 5 | ~200 | Low | 1 day |
| **Total** | **51** | **~1700** | **Medium** | **8-12 days** |

**Notes**:
- Assumes experienced Go developer
- Includes time for manual checksum generation (R2)
- Includes comprehensive testing and documentation
- Conservative estimate with buffer for edge cases

---

## References

### Password Security
- [OWASP Password Storage Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html)
- [CWE-214: Invocation of Process Using Visible Sensitive Information](https://cwe.mitre.org/data/definitions/214.html)
- golang.org/x/term documentation

### Checksum Verification
- [NIST FIPS 180-4: Secure Hash Standard](https://csrc.nist.gov/publications/detail/fips/180/4/final)
- [Supply Chain Security Best Practices](https://slsa.dev/)
- crypto/sha256 Go package documentation

### Path Traversal
- [OWASP Path Traversal](https://owasp.org/www-community/attacks/Path_Traversal)
- [CWE-22: Improper Limitation of a Pathname](https://cwe.mitre.org/data/definitions/22.html)
- filepath Go package documentation

### External Research
- [GitHub tesseract-ocr/tessdata_fast repository](https://github.com/tesseract-ocr/tessdata_fast)
- [Tesseract documentation](https://tesseract-ocr.github.io/tessdoc/Data-Files.html)
- [tessdata_fast releases](https://github.com/tesseract-ocr/tessdata_fast/releases)

---

## Conclusion

Phase 3 addresses three critical P0 security vulnerabilities with comprehensive, battle-tested solutions:

1. **Password Handling**: Multi-method approach (stdin, env, file, interactive) eliminates exposure in process lists and shell history
2. **Checksum Verification**: Embedded SHA256 checksums protect against supply chain attacks and file corruption
3. **Path Sanitization**: Centralized validation prevents path traversal while respecting user access rights

All solutions use standard library features (except already-present golang.org/x/term), ensuring stability, portability, and maintainability. Implementation is estimated at 8-12 days with comprehensive testing and documentation.

**Next Steps**: Review this research document, then proceed to implementation planning in `PLAN.md`.
