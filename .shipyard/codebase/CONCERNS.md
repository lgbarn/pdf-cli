# CONCERNS.md

## Overview
The pdf-cli codebase demonstrates strong engineering practices overall with comprehensive testing (75% coverage threshold enforced), security scanning via gosec, and modern Go 1.25. However, several areas warrant attention: context.TODO usage in critical paths, limited language support for OCR checksum verification, deprecated password flag still in use, and potential goroutine leaks in parallel processing paths.

## Findings

### P0 - Critical Issues

**None Identified**

The codebase has no critical security vulnerabilities or show-stopping issues. Security scanning is active in CI, paths are sanitized, and password handling follows secure patterns.

---

### P1 - High Priority Issues

#### **Context Management: context.TODO in Production Code**
- **Issue**: Using `context.TODO()` instead of propagating parent context in OCR download operations
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/wasm.go` (line 53)
    ```go
    if err := downloadTessdata(context.TODO(), w.dataDir, l); err != nil {
    ```
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr.go` (line 175)
    ```go
    if err := downloadTessdata(context.TODO(), e.dataDir, lang); err != nil {
    ```
- **Impact**: Downloads cannot be cancelled via parent context; user interrupts (Ctrl+C) won't cancel in-flight tessdata downloads
- **Severity**: High - affects user experience and resource management
- **Recommendation**: Pass `ctx` from caller instead of `context.TODO()`

#### **OCR Security: Limited Checksum Coverage**
- **Issue**: Only English language has SHA256 checksum verification; other languages download without integrity checks
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/checksums.go` (lines 9-11)
    ```go
    var KnownChecksums = map[string]string{
        "eng": "7d4322bd2a7749724879683fc3912cb542f19906c83bcc1a52132556427170b2",
    }
    ```
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr.go` (lines 316-321)
    ```go
    } else {
        fmt.Fprintf(os.Stderr,
            "WARNING: No checksum available for language '%s'. Computed SHA256: %s\n",
            lang, computedHash,
        )
    }
    ```
- **Impact**: Users of non-English OCR are vulnerable to corrupted downloads or supply chain attacks
- **Severity**: High - security vulnerability for multi-language OCR users
- **Recommendation**: Add checksums for commonly used languages (fra, deu, spa, etc.) or implement automated checksum generation

#### **Deprecated Password Flag Still Active**
- **Issue**: The `--password` flag is deprecated but still functional, exposing passwords in process listings
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/cli/password.go` (lines 47-52)
    ```go
    // 3. Check --password flag (deprecated)
    if cmd.Flags().Lookup("password") != nil {
        password, _ := cmd.Flags().GetString("password")
        if password != "" {
            fmt.Fprintln(os.Stderr, "WARNING: --password flag is deprecated...")
    ```
- **Impact**: Users may continue using insecure password passing despite warnings
- **Severity**: High - security concern (password exposure in `ps` output)
- **Recommendation**: Remove flag in next major version (v3.0.0) or disable it by default with explicit opt-in

---

### P2 - Medium Priority Issues

#### **Parallel Processing: Potential Goroutine Leak on Context Cancel**
- **Issue**: Goroutines may continue running after context cancellation in parallel text extraction
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/pdf/text.go` (lines 140-148)
    ```go
    for _, pageNum := range pages {
        if ctx.Err() != nil {
            // Context canceled, don't launch more work
            break
        }
        go func(pn int) {
            results <- pageResult{pageNum: pn, text: extractPageText(r, pn, totalPages)}
        }(pageNum)
    }
    ```
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr.go` (lines 478-492)
    ```go
    for i, imgPath := range imageFiles {
        if ctx.Err() != nil {
            // Context canceled, don't launch more work
            break
        }
        // Goroutines launched but may not check context
    ```
- **Impact**: Goroutines launched before cancellation continue processing, wasting resources
- **Severity**: Medium - resource leak on user cancellation
- **Recommendation**: Check `ctx.Err()` inside goroutines before expensive operations

#### **HTTP Client: No Timeout Configuration**
- **Issue**: Using `http.DefaultClient` without custom timeout configuration for tessdata downloads
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr.go` (line 260)
    ```go
    resp, doErr := http.DefaultClient.Do(req)
    ```
- **Impact**: Network requests may hang indefinitely despite context timeout being set on request
- **Severity**: Medium - DefaultClient has no timeout by default, though request context provides some protection
- **Recommendation**: Create custom `http.Client` with `Timeout` field set for defense-in-depth

#### **Password File: Limited Size Validation**
- **Issue**: Password file size limited to 1KB but no validation of content (e.g., binary data, non-printable chars)
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/cli/password.go` (lines 35-38)
    ```go
    if len(data) > 1024 {
        return "", fmt.Errorf("password file exceeds 1KB size limit")
    }
    return strings.TrimSpace(string(data)), nil
    ```
- **Impact**: Potential confusion if user points to wrong file; binary data silently converted to string
- **Severity**: Medium - usability issue more than security
- **Recommendation**: Validate that file contains only printable characters or warn if suspicious content detected

#### **Error Handling: Silent Failures in Text Extraction**
- **Issue**: Text extraction from individual pages silently returns empty string on errors
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/pdf/text.go` (lines 110-123)
    ```go
    func extractPageText(r *pdf.Reader, pageNum, totalPages int) string {
        if pageNum < 1 || pageNum > totalPages {
            return ""
        }
        // ...
        if err != nil {
            return ""
        }
        return text
    }
    ```
- **Impact**: Users may not realize pages failed to extract; no indication in output or logs
- **Severity**: Medium - silent data loss
- **Recommendation**: Collect errors and return as warning or log at debug level

#### **Temporary File Cleanup: Race Condition Window**
- **Issue**: Cleanup registry uses slice index tracking which can fail if cleanup runs concurrently during registration
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/cleanup/cleanup.go` (lines 20-34)
    ```go
    func Register(path string) func() {
        mu.Lock()
        defer mu.Unlock()
        idx := len(paths)
        paths = append(paths, path)
        return func() {
            mu.Lock()
            defer mu.Unlock()
            if idx < len(paths) {
                paths[idx] = "" // mark as unregistered
            }
        }
    }
    ```
- **Impact**: Index may be out of bounds if Run() clears slice while unregister executes; though mutex protects against this, the idempotent Run() using `hasRun` flag could cause issues if reset during execution
- **Severity**: Medium - edge case in cleanup logic
- **Recommendation**: Use map[string]bool instead of slice with indices

---

### P3 - Low Priority Issues

#### **Testing: Panic Usage in Test Helpers**
- **Issue**: Test helper functions use `panic()` which can make debugging harder
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/testing/fixtures.go` (lines 14, 36, 46, 52)
    ```go
    panic("failed to get caller information")
    panic("failed to create temp dir: " + err.Error())
    panic("failed to create temp file: " + err.Error())
    panic("failed to write temp file: " + err.Error())
    ```
- **Impact**: Test failures less informative; panics bypass deferred cleanup
- **Severity**: Low - test-only code
- **Recommendation**: Return errors from test helpers instead of panicking

#### **Performance: Sequential Downloads During Retry**
- **Issue**: Progress bar created once for retry loop but may not display correctly on retries
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr.go` (lines 239-241)
    ```go
    // Progress bar created once; may not display perfectly on retries
    var bar *progressbar.ProgressBar
    ```
- **Impact**: Users see confusing progress output during download retries
- **Severity**: Low - cosmetic issue
- **Recommendation**: Reset/recreate progress bar on each retry attempt

#### **Dependency: Pseudo-versions in Production**
- **Issue**: Using pseudo-versioned dependencies (not proper semver tags)
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/go.mod` (lines 6-7, 18)
    ```go
    github.com/danlock/gogosseract v0.0.11-0ad3421
    github.com/ledongthuc/pdf v0.0.0-20250511090121-5959a4027728
    github.com/danlock/pkg v0.0.46-2e8eb6d
    ```
- **Impact**: Harder to track dependency updates; no semantic versioning guarantees
- **Severity**: Low - common in Go ecosystem but not ideal
- **Recommendation**: Request maintainers to publish proper releases or fork and maintain tagged versions

#### **Code Duplication: Repeated Output Filename Generation**
- **Issue**: Batch operations duplicate suffix patterns across commands
  - Evidence: Multiple commands use `outputOrDefault(output, inputFile, "_compressed.pdf")` pattern
  - Files: `/Users/lgbarn/Personal/pdf-cli/internal/commands/compress.go`, `encrypt.go`, `decrypt.go`, etc.
- **Impact**: Inconsistent suffix naming if someone forgets to update all locations
- **Severity**: Low - technical debt
- **Recommendation**: Define suffix constants in a central location

#### **Security: Directory Permissions Too Permissive**
- **Issue**: Directories created with 0750 permissions; group read/execute enabled
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/fileio/files.go` (line 15)
    ```go
    DefaultDirPerm = 0750
    ```
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr.go` (line 41)
    ```go
    DefaultDataDirPerm = 0750
    ```
- **Impact**: Users in same group can read tessdata/config directories
- **Severity**: Low - acceptable for most use cases; sensitive files use 0600
- **Recommendation**: Consider 0700 for user-only access, though current permissions are acceptable

#### **Logging: Silent Mode is Default**
- **Issue**: Log level defaults to "silent" which hides debugging information
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/cli/flags.go` (line 93)
    ```go
    cmd.PersistentFlags().StringVar(&logLevel, "log-level", "silent", "Log level (debug, info, warn, error, silent)")
    ```
- **Impact**: Users may struggle to debug issues without realizing logging is available
- **Severity**: Low - design decision; documented in help
- **Recommendation**: Consider "error" as default instead of "silent"

---

### P3 - Technical Debt

#### **Time.After in Retry Logic**
- **Issue**: Using `time.After` in select can cause timer leak if context cancelled first
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/retry/retry.go` (line 76)
    ```go
    case <-time.After(delay):
    ```
- **Impact**: Timer goroutine remains until timer expires even if context cancelled
- **Severity**: Low - minor resource leak during retries
- **Recommendation**: Use `time.NewTimer` with explicit Stop() call

#### **WASM Backend: Not Thread-Safe**
- **Issue**: WASM backend documented as not thread-safe; only native backend supports parallelism
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr.go` (line 422)
    ```go
    if len(imageFiles) <= threshold || e.backend.Name() == "wasm" {
        return e.processImagesSequential(ctx, imageFiles, showProgress)
    }
    ```
- **Impact**: WASM OCR is slower than native for large documents
- **Severity**: Low - known limitation of underlying library
- **Recommendation**: Document in README; consider using worker pool with single WASM instance

#### **Merge Progress: Inefficient for Large File Sets**
- **Issue**: Incremental merge creates intermediate files for each step
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/pdf/transform.go` (lines 58-68)
    ```go
    for i := 1; i < len(inputs); i++ {
        err := api.MergeCreateFile([]string{tmpPath, inputs[i]}, tmpPath+".new", false, NewConfig(password))
        // ...
        if err := os.Rename(tmpPath+".new", tmpPath); err != nil {
    ```
- **Impact**: O(n²) I/O operations for n files; slow for large merges
- **Severity**: Low - only affects >3 files with --progress
- **Recommendation**: Consider pdfcpu's batch merge API if available

---

## Summary Table

| Concern | Category | Severity | Affected Area | Confidence |
|---------|----------|----------|---------------|------------|
| context.TODO in downloads | Context Management | P1 High | OCR downloads | Observed |
| Limited OCR checksum coverage | Security | P1 High | Non-English OCR | Observed |
| Deprecated password flag active | Security | P1 High | Password handling | Observed |
| Goroutine leak on cancel | Concurrency | P2 Medium | Parallel processing | Observed |
| No HTTP client timeout | Network | P2 Medium | OCR downloads | Observed |
| Password file validation | Input Validation | P2 Medium | Password handling | Observed |
| Silent page extraction errors | Error Handling | P2 Medium | Text extraction | Observed |
| Cleanup race condition | Concurrency | P2 Medium | Temp file cleanup | Inferred |
| Panic in test helpers | Testing | P3 Low | Test infrastructure | Observed |
| Progress bar during retries | UX | P3 Low | OCR downloads | Observed |
| Pseudo-version dependencies | Dependencies | P3 Low | Build system | Observed |
| Code duplication (suffixes) | Technical Debt | P3 Low | Commands | Observed |
| Directory permissions | Security | P3 Low | File system | Observed |
| Silent logging default | UX | P3 Low | Logging | Observed |
| time.After leak | Performance | P3 Low | Retry logic | Observed |
| WASM not thread-safe | Performance | P3 Low | OCR backend | Observed |
| Inefficient merge progress | Performance | P3 Low | PDF merge | Observed |

---

## Open Questions

1. **OCR Checksum Strategy**: Should all tessdata languages have checksums, or is warning sufficient? Consider automated checksum generation script.

2. **Password Flag Removal**: Is v2.0.0 the right time to remove `--password` entirely, or wait for v3.0.0? Breaking change policy?

3. **Parallel Processing**: Should native Tesseract backend parallelism be configurable or always enabled? Current threshold is 5 images.

4. **Error Aggregation**: Should text extraction return partial results with warnings, or fail fast on first error? Current behavior is silent continuation.

5. **Dependency Management**: Should the project fork `gogosseract` and `ledongthuc/pdf` to maintain stable tagged versions?

6. **Security Policy**: SECURITY.md lists v1.3.x as supported, but current release is v2.0.0. Update needed?
   - Evidence: `/Users/lgbarn/Personal/pdf-cli/SECURITY.md` (lines 7-9)

7. **Golangci-lint Version**: CI uses golangci-lint v2.8.0 which is outdated (current is v1.62+). Intentional or needs update?
   - Evidence: `/Users/lgbarn/Personal/pdf-cli/.github/workflows/ci.yaml` (line 29)

8. **Resource Limits**: Should there be limits on parallel workers, temp file sizes, or memory usage for very large PDFs?

---

## Positive Observations

- **Strong Security Practices**: Path sanitization, gosec scanning, password file limits, #nosec annotations with justification
- **Comprehensive Testing**: 75% coverage threshold enforced, race detection in CI, both unit and integration tests
- **Good Error Handling**: Custom error types with context, password-specific error detection, retry with exponential backoff
- **Signal Handling**: Proper cleanup on SIGTERM/SIGINT via cleanup registry
- **Modern Go**: Using Go 1.25, no deprecated ioutil usage, proper context propagation in most places
- **Documentation**: README is thorough, SECURITY.md exists, inline comments explain non-obvious code
