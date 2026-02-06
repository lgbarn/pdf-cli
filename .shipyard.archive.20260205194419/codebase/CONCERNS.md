# Codebase Concerns Analysis

This document identifies technical debt, security concerns, performance issues, and upgrade needs in the pdf-cli codebase.

## Priority Legend

- **P0 (Critical)**: Security vulnerabilities or data loss risks requiring immediate attention
- **P1 (High)**: Issues affecting reliability, performance, or maintenance significantly
- **P2 (Medium)**: Technical debt that should be addressed in upcoming releases
- **P3 (Low)**: Nice-to-have improvements with minimal impact

---

## P0 - Critical Security Concerns

### 1. Password Handling in Command Line

**Location**: `/internal/cli/flags.go:30-36`, `/internal/commands/*.go` (multiple files)

**Issue**: Passwords are passed via command-line flags (`--password`), which exposes them in:
- Process listings (`ps aux`)
- Shell history files
- System logs
- Parent process environments

**Evidence**:
```go
// internal/cli/flags.go
func AddPasswordFlag(cmd *cobra.Command, usage string) {
    cmd.Flags().String("password", "", usage)
}

// Usage throughout commands:
pdf decrypt secure.pdf --password mysecret -o unlocked.pdf
```

**Impact**: Credentials can be logged or viewed by other users on multi-user systems.

**Recommendation**:
- Add stdin password reading support (e.g., `--password-stdin` or prompt interactively)
- Support reading passwords from environment variables
- Warn users in documentation about command-line password risks
- Consider adding a `--password-file` option

**References**: SECURITY.md mentions this issue but implementation doesn't mitigate it.

---

### 2. HTTP Download Security - No Certificate Validation Controls

**Location**: `/internal/ocr/ocr.go:169-209`

**Issue**: Tessdata downloads use `http.DefaultClient` without explicit TLS configuration or checksum verification. While HTTPS is used, there's no integrity verification of downloaded files.

**Evidence**:
```go
// internal/ocr/ocr.go
func downloadTessdata(dataDir, lang string) error {
    url := fmt.Sprintf("%s/%s.traineddata", TessdataURL, lang)
    // TessdataURL = "https://github.com/tesseract-ocr/tessdata_fast/raw/main"

    resp, err := http.DefaultClient.Do(req)  // No checksum verification
    if err != nil {
        return err
    }
    // ... writes directly to file without verification
}
```

**Impact**:
- Potential supply chain attack vector
- Downloaded files could be corrupted or malicious
- No way to verify file authenticity

**Recommendation**:
- Add SHA256 checksum verification for downloaded tessdata files
- Implement retry logic with exponential backoff
- Add download size limits to prevent DoS
- Consider caching checksums or using signed releases

---

### 3. Lack of Input Sanitization for File Paths

**Location**: Multiple locations, especially `/internal/fileio/files.go`

**Issue**: While `filepath.Clean()` is used in some places, there's inconsistent path sanitization across the codebase. Several operations use user-provided paths directly with only basic validation.

**Evidence**:
```go
// internal/fileio/files.go:86-88
func CopyFile(src, dst string) error {
    cleanSrc := filepath.Clean(src)
    cleanDst := filepath.Clean(dst)
    // Good practice, but not consistently applied everywhere
}

// But in other places:
// internal/pdf/text.go:176
filePath := filepath.Join(tmpDir, filepath.Base(file.Name()))
data, err := os.ReadFile(filePath) // #nosec G304 -- path is within controlled tmpDir
```

**Impact**: Potential path traversal vulnerabilities, though mitigated by nosec annotations showing manual review.

**Recommendation**:
- Create a centralized path sanitization function
- Validate all user-provided paths against expected directories
- Add explicit directory traversal prevention checks
- Document security assumptions in code comments

---

## P1 - High Priority Issues

### 4. Global Mutable State (Race Condition Risk)

**Location**: `/internal/config/config.go:142-155`, `/internal/logging/logger.go:82-130`

**Issue**: Both config and logging packages use global mutable state without synchronization. While single-threaded CLI usage makes this safe currently, the codebase uses goroutines for parallel processing.

**Evidence**:
```go
// internal/config/config.go
var global *Config  // No mutex protection

func Get() *Config {
    if global == nil {
        var err error
        global, err = Load()  // Race if called from multiple goroutines
    }
    return global
}

// internal/logging/logger.go
var global *Logger  // No mutex protection

func Get() *Logger {
    if global == nil {
        Init(LevelSilent, FormatText)  // Race condition possible
    }
    return global
}
```

**Impact**:
- Potential race conditions if concurrent operations access config/logging during initialization
- Tests using `Reset()` could interfere with parallel test execution

**Recommendation**:
- Use `sync.Once` for lazy initialization
- Add mutex protection for global state access
- Consider dependency injection instead of global state

---

### 5. Context Usage - Missing Context Propagation

**Location**: `/internal/ocr/ocr.go:316`, `/internal/ocr/wasm.go:123`

**Issue**: Multiple places use `context.Background()` instead of accepting context from callers. This prevents proper cancellation propagation and timeout handling.

**Evidence**:
```go
// internal/ocr/ocr.go:316
ctx := context.Background()  // Should accept ctx from caller
for _, imgPath := range imageFiles {
    text, err := e.backend.ProcessImage(ctx, imgPath, e.lang)
}

// internal/ocr/wasm.go:123
func (w *WASMBackend) Close() error {
    if w.tess != nil {
        return w.tess.Close(context.Background())  // Hardcoded context
    }
    return nil
}
```

**Impact**:
- Cannot cancel long-running operations
- No timeout support for OCR operations
- Resource leaks if operations hang

**Recommendation**:
- Add `context.Context` parameter to all long-running functions
- Propagate context from command layer through all operations
- Add timeout handling for HTTP downloads and OCR processing

---

### 6. Error Handling - Silent Errors in Parallel Processing

**Location**: `/internal/pdf/text.go:106-150`, `/internal/fileio/files.go:143-167`

**Issue**: Parallel processing code silently ignores errors in some cases, which could lead to incomplete results without user notification.

**Evidence**:
```go
// internal/pdf/text.go:122-124
go func(pn int) {
    results <- pageResult{pageNum: pn, text: extractPageText(r, pn, totalPages)}
}(pageNum)  // extractPageText swallows errors, returns empty string

// internal/fileio/files.go:159-165
for range paths {
    r := <-results
    if r.err != nil && firstErr == nil {
        firstErr = r.err  // Only reports FIRST error, rest are silent
    }
}
```

**Impact**:
- Users may not know that some pages/files failed to process
- Silent failures lead to incomplete output
- Debugging is difficult when errors are swallowed

**Recommendation**:
- Collect all errors and report them comprehensively
- Add verbose logging for skipped/failed items
- Return partial success indication with error details

---

### 7. Dependency Version Management

**Location**: `go.mod`

**Issue**: Using non-semantic versioned dependencies and outdated packages. At least 21 dependencies have newer versions available.

**Evidence**:
```
// From go.mod
github.com/danlock/gogosseract v0.0.11-0ad3421  // Commit hash, not semantic version
github.com/danlock/pkg v0.0.17-a9828f2          // Commit hash, not semantic version

// Outdated dependencies (from go list -m -u all):
github.com/clipperhouse/uax29/v2 v2.2.0 [v2.4.0]
github.com/danlock/pkg v0.0.17-a9828f2 [v0.0.46-2e8eb6d]
github.com/google/pprof v0.0.0-20210407192527 [v0.0.0-20260115054156]
github.com/jerbob92/wazero-emscripten-embind v1.3.0 [v1.5.2]
... 17 more outdated packages
```

**Impact**:
- Missing security patches
- Missing bug fixes and performance improvements
- Harder to track dependency changes
- Potential compatibility issues

**Recommendation**:
- Update all dependencies to latest stable versions
- Switch from commit hashes to tagged releases
- Add Dependabot or similar tool for automated dependency updates
- Document dependency update policy in CONTRIBUTING.md

---

### 8. Missing Resource Cleanup - Deferred Close Errors Ignored

**Location**: Throughout codebase

**Issue**: Deferred `Close()` calls ignore errors, which could lead to data loss or incomplete writes. The `.golangci.yaml` explicitly excludes these from errcheck linting.

**Evidence**:
```go
// Pattern appears in multiple files:
defer f.Close()  // Error ignored
defer resp.Body.Close()  // Error ignored

// .golangci.yaml:60-64
exclusions:
  rules:
    - linters:
        - errcheck
      text: "Error return value of .*(Close|Remove|RemoveAll).*is not checked"
```

**Impact**:
- File write operations may not complete (buffered data not flushed)
- Network connections may not close cleanly
- Resource leaks in error paths

**Recommendation**:
- Check `Close()` errors for file writes
- Use named return values and defer error checking
- At minimum, log ignored close errors in verbose mode

---

## P2 - Medium Priority Technical Debt

### 9. Go Version Requirement Mismatch

**Location**: `go.mod:3`, `CONTRIBUTING.md:8`, `README.md:71`

**Issue**: Documentation claims Go 1.21+ compatibility, but go.mod specifies Go 1.24.1 (bleeding edge at time of analysis).

**Evidence**:
```go
// go.mod
go 1.24.1

// CONTRIBUTING.md
Prerequisites
- Go 1.21 or later

// README.md
Prerequisites
- Go 1.24 or later
```

**Impact**:
- Users with Go 1.21-1.23 may encounter unexpected issues
- Documentation inconsistency confuses users
- May exclude users on stable Go versions

**Recommendation**:
- Use the minimum supported Go version in go.mod (likely 1.21 or 1.22)
- Align all documentation
- Test against multiple Go versions in CI

---

### 10. Large Test Files and Coverage Gaps

**Location**: `/internal/pdf/pdf_test.go` (2340 lines)

**Issue**: Some test files are very large and difficult to maintain. The largest test file has 2,340 lines, making it hard to navigate and maintain.

**Evidence**:
```bash
# Largest test files:
2340 ./internal/pdf/pdf_test.go
882 ./internal/commands/commands_integration_test.go
620 ./internal/commands/additional_coverage_test.go
```

**Impact**:
- Harder to maintain and understand tests
- Slower test execution
- Difficult to identify which specific functionality is being tested

**Recommendation**:
- Split large test files by functionality
- Use table-driven tests more consistently
- Consider sub-test organization with `t.Run()`

---

### 11. Temporary File Management Risk

**Location**: `/internal/ocr/ocr.go:219-223`, `/internal/pdf/text.go:154-158`

**Issue**: Temporary directories are created and cleaned up with `defer os.RemoveAll()`, but if the process crashes or is killed, temp files may persist.

**Evidence**:
```go
// internal/ocr/ocr.go
tmpDir, err := os.MkdirTemp("", "pdf-ocr-*")
if err != nil {
    return "", fmt.Errorf("failed to create temp directory: %w", err)
}
defer os.RemoveAll(tmpDir)  // Won't run if process killed
```

**Impact**:
- Disk space consumption from orphaned temp files
- Potential sensitive data exposure in temp directories
- No cleanup mechanism for stale temp files

**Recommendation**:
- Document temp file location for manual cleanup
- Consider adding a cleanup command
- Use OS temp directory cleaning (relies on OS-level cleanup)
- Add process exit handlers where appropriate

---

### 12. Limited PDF/A Support

**Location**: Documentation and `/internal/pdf/validation.go`

**Issue**: PDF/A validation is explicitly documented as basic/limited, but users may not realize the limitations until after attempting validation.

**Evidence**:
```markdown
// README.md:419-432
> **⚠️ PDF/A Limitations**
> This tool provides **basic** PDF/A validation and optimization, not full ISO compliance:
> | Feature | Status |
> | Font embedding check | ✗ Limited |
> | Color profile validation | ✗ Not supported |
> | Full ISO 19005 compliance | ✗ Not supported |
```

**Impact**:
- Users may rely on incomplete validation
- False sense of compliance
- Potential issues when submitting to systems requiring full PDF/A

**Recommendation**:
- Make limitations more prominent in command output
- Add warning when validation is run
- Consider removing feature if not fully supported or integrate with veraPDF

---

### 13. No Rate Limiting on External Downloads

**Location**: `/internal/ocr/ocr.go:169-209`

**Issue**: No rate limiting or retry logic for tessdata downloads from GitHub.

**Evidence**:
```go
func downloadTessdata(dataDir, lang string) error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
    defer cancel()

    req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
    // No retry logic, no rate limiting
    resp, err := http.DefaultClient.Do(req)
}
```

**Impact**:
- Single failure means user must retry manually
- Could hit GitHub rate limits with multiple parallel downloads
- No graceful degradation

**Recommendation**:
- Add exponential backoff retry logic
- Implement rate limiting for multiple language downloads
- Cache downloaded files more aggressively
- Consider bundling common languages

---

### 14. WASM Backend Performance Concerns

**Location**: `/internal/ocr/ocr.go:302-307`

**Issue**: WASM OCR backend is forced to single-threaded execution even for large batches, significantly slower than native Tesseract.

**Evidence**:
```go
// internal/ocr/ocr.go:302-307
func (e *Engine) processImages(imageFiles []string, showProgress bool) (string, error) {
    // Use sequential processing for small batches or WASM backend (not thread-safe)
    if len(imageFiles) <= parallelThreshold || e.backend.Name() == "wasm" {
        return e.processImagesSequential(imageFiles, showProgress)
    }
    return e.processImagesParallel(imageFiles, showProgress)
}
```

**Impact**:
- WASM OCR is significantly slower for multi-page documents
- No way to parallelize WASM operations
- Poor user experience for large documents without native Tesseract

**Recommendation**:
- Document performance difference in README
- Encourage native Tesseract installation where possible
- Consider worker pool with multiple WASM instances if library supports it
- Add estimated time warnings for large WASM operations

---

## P3 - Low Priority Improvements

### 15. Hardcoded Magic Numbers

**Location**: Various files

**Issue**: Magic numbers scattered throughout code without named constants.

**Evidence**:
```go
// internal/ocr/ocr.go:299
const parallelThreshold = 5

// internal/pdf/text.go:58
if len(pages) > 5 {  // Hardcoded threshold

// internal/fileio/files.go:28
return os.MkdirAll(path, 0750)  // Permission constant
```

**Impact**:
- Harder to understand code intent
- Difficult to tune performance
- Inconsistent thresholds

**Recommendation**:
- Extract magic numbers to named constants
- Document reasoning for threshold values
- Consider making some values configurable

---

### 16. Inconsistent Logging

**Location**: Throughout codebase

**Issue**: Mix of `fmt.Fprintf(os.Stderr, ...)`, structured logging with `slog`, and `cli.PrintVerbose()`. No consistent logging strategy.

**Evidence**:
```go
// internal/ocr/ocr.go:173
fmt.Fprintf(os.Stderr, "Downloading tessdata for '%s'...\n", lang)

// internal/commands/meta.go:91
cli.PrintVerbose("Reading metadata from %s", inputFile)

// Both logging packages exist but not used consistently
```

**Impact**:
- Hard to filter/search logs
- Inconsistent output format
- Can't disable certain log types easily

**Recommendation**:
- Standardize on structured logging (slog)
- Use consistent log levels
- Route all output through logging package

---

### 17. Test Coverage Dependency

**Location**: `Makefile:112-121`, `.github/workflows/ci.yaml:52-61`

**Issue**: 75% coverage threshold is enforced but uses shell scripting with `bc` or `awk`, which may not be portable.

**Evidence**:
```makefile
# Makefile requires bc which isn't standard on all systems
coverage-check:
    if [ $$(echo "$$coverage < 75" | bc -l) -eq 1 ]; then
```

**Impact**:
- Build may fail on systems without `bc`
- CI/CD dependency on specific tools

**Recommendation**:
- Use Go-based coverage tools
- Consider `gocover-cobertura` or similar
- Document required system tools

---

### 18. Documentation Sync Challenges

**Location**: `go.mod` vs `CONTRIBUTING.md`, `SECURITY.md`

**Issue**: Multiple places specify version support, security policies, etc. Easy to get out of sync.

**Evidence**:
- Go version: 1.24.1 in go.mod, 1.21+ in CONTRIBUTING.md, 1.24+ in README.md
- Security policy mentions versions 1.2.x and 1.3.x but current version is 1.5.0

**Impact**:
- Confusing for users
- Incorrect security expectations

**Recommendation**:
- Single source of truth for versions (generate docs from code)
- Automated checks for documentation consistency
- Update security policy to match current release

---

## Performance Optimization Opportunities

### 19. Parallel Processing Thresholds

**Location**: `/internal/fileio/files.go:134`, `/internal/pdf/text.go:58`, `/internal/ocr/ocr.go:299`

**Issue**: Hardcoded thresholds (3, 5) for parallel processing may not be optimal for all systems.

**Evidence**:
```go
// internal/fileio/files.go:134
if len(paths) <= 3 {  // Why 3?
    // sequential
}

// internal/pdf/text.go:58
if len(pages) > 5 {  // Why 5?
    // parallel
}
```

**Impact**:
- May not utilize available CPU cores efficiently
- Could spawn too many goroutines on small systems

**Recommendation**:
- Base thresholds on `runtime.NumCPU()`
- Make configurable for advanced users
- Benchmark different thresholds

---

### 20. Memory Usage - Unbuffered Channels

**Location**: `/internal/pdf/text.go:119`, `/internal/ocr/ocr.go:340`

**Issue**: Buffered channels with capacity = len(items) could consume significant memory for large batches.

**Evidence**:
```go
// internal/pdf/text.go:119
results := make(chan pageResult, len(pages))  // Could be 1000+ pages

// internal/ocr/ocr.go:340
results := make(chan imageResult, len(imageFiles))  // Could be 100+ images
```

**Impact**:
- High memory usage for large documents
- Could cause OOM on constrained systems

**Recommendation**:
- Use fixed buffer size (e.g., NumCPU * 2)
- Implement worker pool pattern
- Stream results instead of buffering all

---

## Summary Statistics

**Total Issues Identified**: 20

| Priority | Count | Focus Area |
|----------|-------|------------|
| P0 (Critical) | 3 | Security (password handling, download integrity, path sanitization) |
| P1 (High) | 5 | Race conditions, context handling, error handling, dependencies, resource cleanup |
| P2 (Medium) | 6 | Version management, test organization, temp files, limitations disclosure, downloads |
| P3 (Low) | 4 | Code quality (magic numbers, logging consistency, documentation) |
| Performance | 2 | Parallel processing tuning, memory optimization |

**Code Metrics**:
- Total Go files: 87
- Production code: ~5,644 lines
- Largest test file: 2,340 lines
- Direct dependencies: 12
- Total dependencies (including transitive): 43 (from go.sum)
- Outdated dependencies: 21

**Security Posture**:
- Gosec scanning enabled in CI
- Multiple `#nosec` annotations (14+ instances) - all appear justified with comments
- SECURITY.md exists and is maintained
- No obvious SQL injection, XSS, or remote code execution vulnerabilities
- Main concerns: credential exposure and supply chain security

**Maintenance Health**:
- Active CI/CD pipeline with testing, linting, and security scanning
- 75% test coverage threshold enforced
- Recent refactoring (v1.5.0) shows active maintenance
- Clear contributing guidelines
- Architecture documentation exists

## Recommended Action Plan

### Immediate (Next Sprint)
1. **Fix password handling** - Add stdin/env var support (P0)
2. **Add download integrity verification** - SHA256 checksums for tessdata (P0)
3. **Update critical dependencies** - Security patches (P1)

### Short-term (Next Release)
4. **Fix race conditions** - Add sync.Once to global state (P1)
5. **Add context propagation** - Proper cancellation support (P1)
6. **Improve error reporting** - Collect and report all errors from parallel operations (P1)

### Medium-term (Next Quarter)
7. **Dependency modernization** - Update all 21 outdated packages (P1)
8. **Resource cleanup** - Check close errors for file writes (P1)
9. **Documentation consistency** - Align version requirements (P2)
10. **Test refactoring** - Split large test files (P2)

### Long-term (Backlog)
11. **Standardize logging** - Unified structured logging (P3)
12. **Performance tuning** - Adaptive parallelism based on CPU cores (Perf)
13. **Memory optimization** - Worker pool pattern for large batches (Perf)
