# Security Audit Report: Phase 1 - OCR Download Path Hardening

**Phase:** Phase 1 - OCR Download Path Hardening
**Date:** 2026-02-05
**Scope:** 4 commits, 3 core files modified (ocr.go, wasm.go, retry.go)
**Auditor:** Security and Compliance Review
**Lines Changed:** ~150 lines across 3 files

## Executive Summary

**Overall Risk Assessment:** ✅ **LOW**

Phase 1 focused on hardening the OCR download path by:
1. Replacing `http.DefaultClient` with a custom timeout-configured client
2. Propagating context through `EnsureTessdata` methods to enable cancellation
3. Replacing `time.After` with `time.NewTimer` to prevent resource leaks
4. Recreating progress bars per retry attempt for cleaner UX

**Security Posture:** The phase successfully addresses resource exhaustion vectors and improves defensive programming without introducing new vulnerabilities.

**Findings Summary:**
- **Critical:** 0
- **High:** 0
- **Medium:** 1 (redirect following behavior)
- **Low:** 2 (informational improvements)

---

## Detailed Security Analysis

### 1. Code Security (OWASP Top 10)

#### ✅ A01:2021 - Broken Access Control
**Status:** PASS
**Analysis:** No access control changes in this phase. Path sanitization (`SanitizePath`) was already implemented and is correctly used at line 221 of `ocr.go`.

#### ✅ A02:2021 - Cryptographic Failures
**Status:** PASS
**Analysis:**
- TLS/HTTPS enforced: Downloads use `https://github.com/tesseract-ocr/tessdata_fast/raw/main` (line 33)
- SHA256 checksum verification present (lines 241, 318-327)
- Checksum database exists in `checksums.go` with known-good hashes
- Supply chain integrity protected via checksum validation with clear error message referencing "supply chain attack" (line 323)

#### ✅ A03:2021 - Injection
**Status:** PASS
**Analysis:**
- Download URLs constructed safely using `fmt.Sprintf` with controlled baseURL and validated language parameter (line 217)
- Language parameter sanitized through `parseLanguages` (line 188) which splits on delimiters and trims whitespace
- No command injection vectors: no shell execution in changed code
- File paths sanitized via `SanitizePath` (line 221) which prevents directory traversal

#### ✅ A04:2021 - Insecure Design
**Status:** PASS
**Analysis:**
- **Context propagation:** Properly implemented throughout the call chain (`EnsureTessdata(ctx)`, `downloadTessdata(ctx, ...)`, `downloadTessdataWithBaseURL(ctx, ...)`)
- **Timeout defense:** Custom HTTP client with 5-minute timeout prevents indefinite hangs (lines 27-29)
- **Retry strategy:** Exponential backoff with context awareness (lines 246-300)
- **Permanent vs. transient error classification:** Correctly distinguishes permanent errors (4xx client errors) from retryable errors (network failures, 429, 5xx)

#### ✅ A05:2021 - Security Misconfiguration
**Status:** PASS
**Analysis:**
- HTTP client configured with explicit timeout (5 minutes - reasonable for ~15MB downloads)
- Default HTTP client replaced to prevent infinite timeout vulnerability
- No debug/verbose logging of sensitive data observed
- Directory permissions set to 0750 (read/execute for owner+group only) - secure default

#### ✅ A08:2021 - Software and Data Integrity Failures
**Status:** PASS (with enhancement recommendation)
**Analysis:**
- **Checksum verification:** SHA256 hashes verified after download (lines 318-327)
- **Warning on missing checksums:** Tool warns when no checksum available (lines 329-332)
- **Atomic file operations:** Temporary file used during download, renamed on success (lines 231-335)
- **Corruption detection:** Download failure triggers retry, checksum mismatch aborts with clear error

---

### 2. Secrets Scanning

**Status:** ✅ PASS - No secrets detected

Scanned files:
- `internal/ocr/ocr.go` - No hardcoded credentials
- `internal/ocr/wasm.go` - No hardcoded credentials
- `internal/retry/retry.go` - No hardcoded credentials

Verified:
- No API keys, tokens, or passwords in code or comments
- No base64-encoded credentials
- GitHub URL is public (no authentication required)
- No `.env` files committed

---

### 3. Dependency Audit

**Status:** ✅ PASS - No new dependencies added

Phase 1 modified existing code without adding new dependencies. Current dependency status:

| Dependency | Version | Known CVEs | Status |
|-----------|---------|------------|---------|
| pdfcpu/pdfcpu | v0.11.1 | None known | ✅ OK |
| danlock/gogosseract | v0.0.11-0ad3421 | None known | ✅ OK |
| schollz/progressbar/v3 | v3.19.0 | None known | ✅ OK |
| ledongthuc/pdf | v0.0.0-20250511090121 | None known | ✅ OK |
| golang.org/x/crypto | v0.47.0 | None known | ✅ OK |

**Note:** All dependencies are pinned to specific versions in `go.mod`. No version ranges used.

---

### 4. HTTP Client Security

**Status:** ✅ PASS (with advisory recommendation)

#### Changes Made:
```go
// Before: Used http.DefaultClient (no timeout - vulnerability)
resp, doErr := http.DefaultClient.Do(req)

// After: Custom client with explicit timeout
var tessdataHTTPClient = &http.Client{
    Timeout: DefaultDownloadTimeout,
}
resp, doErr := tessdataHTTPClient.Do(req)
```

#### Security Analysis:

✅ **Timeout Configuration:** 5 minutes (`DefaultDownloadTimeout = 5 * time.Minute`) is appropriate for downloading ~15MB tessdata files over slow connections.

✅ **Context Cancellation:** HTTP requests use `http.NewRequestWithContext(retryCtx, ...)` (line 265) enabling proper cancellation propagation.

✅ **TLS Verification:** Default Go TLS configuration used (validates certificates, rejects invalid certs).

✅ **Request Method:** Uses `http.MethodGet` (line 265) - read-only operation.

⚠️ **Redirect Following:** Default redirect policy (up to 10 redirects) is used. See **Medium Finding #1** below.

---

### 5. Context and Cancellation

**Status:** ✅ PASS

#### Changes Made:
1. `Engine.EnsureTessdata()` → `Engine.EnsureTessdata(ctx context.Context)` (line 175)
2. `WASMBackend.EnsureTessdata(lang)` → `WASMBackend.EnsureTessdata(ctx context.Context, lang)` (line 45)
3. `context.TODO()` replaced with propagated context in both call sites

#### Security Benefits:
- **DoS Prevention:** User cancellation (Ctrl+C) now properly cancels in-flight HTTP requests
- **Resource Cleanup:** Goroutines and network connections released on cancellation
- **Timeout Enforcement:** Context timeout at line 228 ensures downloads can't hang indefinitely

#### Context Layering:
```go
// Parent context from caller
func (e *Engine) EnsureTessdata(ctx context.Context) error {
    ...
    downloadTessdata(ctx, e.dataDir, lang)  // Context passed through
}

// Additional timeout layer added for download operation
ctx, cancel := context.WithTimeout(ctx, DefaultDownloadTimeout)
defer cancel()
```

**Analysis:** Context timeout is additive (whichever expires first cancels the operation). This is correct defensive behavior.

---

### 6. Timer Management

**Status:** ✅ PASS

#### Changes Made (internal/retry/retry.go):
```go
// Before: Resource leak - time.After creates non-stoppable timer
select {
case <-time.After(delay):
case <-ctx.Done():
    return ctx.Err()
}

// After: Proper cleanup with timer.Stop()
timer := time.NewTimer(delay)
select {
case <-timer.C:
case <-ctx.Done():
    timer.Stop()
    return ctx.Err()
}
```

#### Security Impact:
- **Resource Leak Prevention:** `time.After` creates a goroutine and timer that can't be stopped. If context cancels early, the timer persists until expiry (CWE-404: Improper Resource Shutdown).
- **Memory Exhaustion:** Under high retry load with frequent cancellations, leaked timers accumulate until GC.
- **Fix Verification:** `timer.Stop()` called on early exit, preventing leak.

**CWE Reference:** CWE-404 (Improper Resource Shutdown or Release)

---

## Findings

### Medium Findings

#### M-1: HTTP Redirect Following Without Host Validation
- **Location:** `internal/ocr/ocr.go:27-29`
- **Category:** OWASP A05:2021 - Security Misconfiguration / CWE-601 (URL Redirection to Untrusted Site)
- **Description:** The custom HTTP client `tessdataHTTPClient` uses Go's default redirect policy, which follows up to 10 redirects without host validation. An attacker with control over `github.com` DNS or a MITM position could redirect downloads to a malicious host.
- **Risk:**
  - **Likelihood:** Low (requires MITM or DNS compromise)
  - **Impact:** Medium (malicious tessdata could be downloaded and loaded into WASM runtime, though checksum verification provides defense-in-depth)
- **Existing Mitigations:**
  - TLS certificate validation prevents most MITM attacks
  - SHA256 checksum verification (lines 318-327) would detect altered files
  - Only English (`eng`) has a known checksum currently; other languages show warning but proceed
- **Remediation:**
  ```go
  var tessdataHTTPClient = &http.Client{
      Timeout: DefaultDownloadTimeout,
      CheckRedirect: func(req *http.Request, via []*http.Request) error {
          // Only allow redirects within github.com
          if req.URL.Host != "github.com" && req.URL.Host != "raw.githubusercontent.com" {
              return fmt.Errorf("redirect to untrusted host blocked: %s", req.URL.Host)
          }
          if len(via) >= 5 {
              return fmt.Errorf("too many redirects")
          }
          return nil
      },
  }
  ```
- **Reference:** CWE-601 (URL Redirection to Untrusted Site)

---

### Low / Informational Findings

#### L-1: Limited Checksum Coverage for Tessdata Languages
- **Location:** `internal/ocr/checksums.go:9-11`
- **Description:** Only English (`eng`) tessdata has a known SHA256 checksum. Other languages (fra, deu, spa, etc.) download without integrity verification, only showing a warning.
- **Risk:** Supply chain integrity cannot be verified for non-English languages. A compromised GitHub repository or MITM could serve malicious tessdata.
- **Current Behavior:**
  ```
  WARNING: No checksum available for language 'fra'. Computed SHA256: <hash>
  ```
- **Recommendation:** Add checksums for commonly used languages (fra, deu, spa, ita, por, rus, chi_sim, jpn). The `checksums.go` file includes instructions for generating them.
- **Priority:** Low (GitHub repo is trusted source with HTTPS, and WASM runtime provides some sandboxing)

#### L-2: Password Parameter Logging Risk
- **Location:** `internal/ocr/ocr.go:339, 375, 392`
- **Description:** Several functions accept a `password` string parameter for encrypted PDFs. While the phase doesn't log passwords, future logging additions could accidentally expose them.
- **Current State:** No logging of password parameter observed in changed code.
- **Recommendation:** Consider using a dedicated `Password` type (not just `string`) that can be marked as sensitive for logging frameworks:
  ```go
  type Password string

  func (p Password) MarshalJSON() ([]byte, error) {
      return []byte(`"[REDACTED]"`), nil
  }
  ```
- **Reference:** CWE-532 (Insertion of Sensitive Information into Log File)

---

## Cross-Task Security Observations

### ✅ Defense in Depth
Phase 1 demonstrates multiple layers of security controls:
1. **HTTPS transport** (TLS encryption and certificate validation)
2. **Path sanitization** (directory traversal prevention at line 221)
3. **Checksum verification** (supply chain integrity)
4. **Context cancellation** (DoS prevention via timeout and user cancellation)
5. **Retry logic with exponential backoff** (prevents overwhelming remote server)
6. **Atomic file operations** (temp file + rename pattern)

### ✅ Error Handling Consistency
- Network errors return clear, non-leaking error messages
- Permanent vs. transient error classification is consistent
- HTTP status codes handled appropriately:
  - `200 OK` → proceed
  - `429 Too Many Requests` → retry
  - `5xx Server Errors` → retry
  - `4xx Client Errors (except 429)` → permanent failure

### ✅ Resource Management
- Temporary files registered with cleanup handler (line 236)
- Timers properly stopped on early exit (retry.go:79)
- File descriptors closed via defer with error checking
- Progress bars finished on all exit paths

### ✅ Secure Defaults
- Directory permissions: `0750` (owner+group only)
- HTTP timeout: 5 minutes (reasonable for ~15MB downloads)
- Retry attempts: 3 (prevents excessive load)
- Base delay: 1 second with exponential backoff

---

## Compliance and Standards

### ✅ CWE Coverage
- **CWE-22 (Path Traversal):** Protected via `SanitizePath` (lines 221-224)
- **CWE-89 (SQL Injection):** N/A (no database operations)
- **CWE-94 (Code Injection):** N/A (no dynamic code execution)
- **CWE-319 (Cleartext Transmission):** Mitigated (HTTPS enforced)
- **CWE-404 (Improper Resource Shutdown):** Fixed in Phase 1 (timer cleanup)
- **CWE-502 (Deserialization):** Low risk (binary tessdata loaded into sandboxed WASM)
- **CWE-601 (URL Redirection):** See Medium Finding M-1

### ✅ OWASP Top 10 Compliance
All OWASP 2021 Top 10 categories reviewed. No critical or high-risk issues identified.

---

## Recommendations

### Immediate (Pre-Ship)
None. All critical issues resolved. Phase 1 may proceed to shipping.

### Short-Term (Next Phase)
1. **Implement redirect host validation** (Medium Finding M-1) - Recommended for Phase 2
2. **Add checksums for top 10 languages** (Low Finding L-1) - Recommended for Phase 2

### Long-Term (Future Phases)
1. Consider implementing a `Password` type with redaction for logging safety
2. Evaluate WASM runtime sandboxing for tessdata (though current risk is low given checksum verification)
3. Consider rate limiting for download retries (currently 3 attempts with exponential backoff is reasonable)

---

## Test Coverage Analysis

**Files Modified:**
- `internal/ocr/ocr.go` - Core download logic
- `internal/ocr/wasm.go` - WASM backend context propagation
- `internal/retry/retry.go` - Timer management fix

**Test Files Updated:**
- `internal/ocr/ocr_test.go` - Updated for context parameter
- `internal/ocr/engine_extended_test.go` - Updated for context parameter
- `internal/ocr/wasm_test.go` - Updated for context parameter

**Security Test Gaps:**
- ❌ No test for redirect following behavior (M-1)
- ❌ No test for timer cleanup on context cancellation (though code inspection confirms correct implementation)
- ✅ Checksum verification tested (existing tests)
- ✅ Path sanitization tested (existing `fileio/files_test.go`)

**Recommendation:** Add integration test verifying HTTP redirects are limited/blocked for Phase 2.

---

## Approval

**Security Review Status:** ✅ **APPROVED FOR SHIPPING**

**Rationale:**
- No critical or high-severity vulnerabilities introduced
- One medium-severity finding (redirect validation) is mitigated by existing checksum verification and TLS
- Two low-severity findings are informational and can be addressed in Phase 2
- Phase successfully improves security posture by fixing resource leaks and enabling cancellation
- Defense-in-depth controls present throughout the codebase

**Conditions:**
- None. Phase may proceed to `/shipyard:ship`.

**Signed:** Security and Compliance Auditor
**Date:** 2026-02-05

---

## Appendix: Phase 1 Commit History

```
d47d78f fix(ocr): recreate progress bar per retry attempt
5e6e82d refactor(ocr): propagate context through EnsureTessdata methods
2ebeedb fix(retry): replace time.After with time.NewTimer to prevent leaks
d10814e refactor(ocr): replace http.DefaultClient with custom timeout client
```

**Total Changes:**
- 3 core files modified
- ~150 lines changed (additions + deletions)
- 4 focused commits addressing specific security/quality issues
- All changes backward-compatible (context parameter addition)

---

## Audit Methodology

This audit followed the security protocol defined in the `shipyard:security-audit` skill:

1. **Code Security Analysis:** Manual review of all changed lines against OWASP Top 10
2. **Secrets Scanning:** Regex-based scan for API keys, tokens, passwords, private keys
3. **Dependency Audit:** Review of `go.mod` for new/updated dependencies and CVE check
4. **HTTP Client Security:** Analysis of timeout, TLS, redirect, and request configuration
5. **Context/Cancellation:** Verification of context propagation and resource cleanup
6. **Timer Management:** Verification of timer lifecycle and cleanup on cancellation
7. **Cross-Task Analysis:** Evaluation of security coherence across the entire phase

**Tools Used:**
- Manual code review
- Git diff analysis
- Grep-based secrets scanning
- Go module dependency listing
- CWE/OWASP standards reference

---

**END OF AUDIT REPORT**
