# Security Audit Report
**Phase:** Phase 3 - Security Hardening
**Date:** 2026-01-31
**Scope:** 185 lines changed across 4 new/modified files + 14 command files updated
**Auditor:** Security & Compliance Agent

## Summary
**Verdict:** PASS WITH ADVISORIES
**Critical findings:** 0
**Important findings:** 3
**Advisory findings:** 5

Phase 3 successfully addresses the three P0 security requirements (R1, R2, R3) with robust implementations. The password handling, path sanitization, and checksum verification are well-designed. However, several important hardening opportunities and edge cases should be addressed before production deployment.

---

## Critical Findings
**None.**

---

## Important Findings

### I1: Password File Path Not Sanitized (OWASP: Path Traversal)
**Location:** `/Users/lgbarn/Personal/pdf-cli/.worktrees/phase-2-concurrency/internal/cli/password.go:22-24`
**Category:** Path Traversal / Input Validation
**Description:**
The `ReadPassword` function reads the password file path directly from the `--password-file` flag without sanitizing it through `fileio.SanitizePath()`. While the file is read (not written), an attacker could use this to:
1. Read arbitrary files on the system (e.g., `--password-file=/etc/passwd`)
2. Bypass intended access controls
3. Potentially leak sensitive data if the command output is logged

```go
// Current code (line 22-24)
passwordFile, _ := cmd.Flags().GetString("password-file")
if passwordFile != "" {
    data, err := os.ReadFile(passwordFile)  // No sanitization!
```

**Risk:** An attacker with CLI access could read any file the process has permission to read by providing a malicious path like `../../etc/shadow` or `/root/.ssh/id_rsa`.

**Remediation:**
```go
passwordFile, _ := cmd.Flags().GetString("password-file")
if passwordFile != "" {
    // Sanitize path before reading
    sanitizedPath, err := fileio.SanitizePath(passwordFile)
    if err != nil {
        return "", fmt.Errorf("invalid password file path: %w", err)
    }
    data, err := os.ReadFile(sanitizedPath)
    if err != nil {
        return "", fmt.Errorf("failed to read password file: %w", err)
    }
    // ... rest of logic
}
```

**Reference:** CWE-22 (Improper Limitation of a Pathname to a Restricted Directory)

---

### I2: Symlink Attack Vector in Path Sanitization
**Location:** `/Users/lgbarn/Personal/pdf-cli/.worktrees/phase-2-concurrency/internal/fileio/files.go:225-249`
**Category:** Path Traversal / TOCTOU
**Description:**
The `SanitizePath` function checks for ".." components but does not resolve or check for symlinks. An attacker could create a symlink pointing outside the intended directory tree:

```bash
ln -s /etc/passwd ./innocent-looking.pdf
pdf-cli info innocent-looking.pdf
```

This bypasses the ".." check because the path itself contains no ".." components, but the symlink resolves to a restricted location.

**Risk:**
- Medium impact if combined with write operations (e.g., `--output` paths)
- Low impact for read-only operations
- Potential for data exfiltration or overwriting system files

**Remediation Options:**
1. **Best:** Use `filepath.EvalSymlinks()` to resolve symlinks and then check if the resolved path is within allowed directories
2. **Alternative:** Use `filepath.Abs()` + base directory validation
3. **Defense in depth:** Add file type validation (check magic bytes for PDFs)

Example:
```go
func SanitizePath(path string) (string, error) {
    if path == "-" {
        return path, nil
    }

    // Check for ".." before resolving symlinks
    for _, part := range strings.Split(path, "/") {
        if part == ".." {
            return "", fmt.Errorf("path contains directory traversal: %s", path)
        }
    }

    cleaned := filepath.Clean(path)

    // Resolve symlinks to detect indirect traversal
    resolved, err := filepath.EvalSymlinks(cleaned)
    if err != nil && !os.IsNotExist(err) {
        return "", fmt.Errorf("failed to resolve symlinks: %w", err)
    }
    if err == nil {
        // Re-check resolved path for traversal
        for _, part := range strings.Split(resolved, "/") {
            if part == ".." {
                return "", fmt.Errorf("symlink resolves to traversal path: %s", path)
            }
        }
        return resolved, nil
    }

    return cleaned, nil
}
```

**Note:** `EvalSymlinks` will fail for non-existent files (common for `--output` paths), so error handling must distinguish between "file not found" (acceptable) vs. "symlink resolution failed" (suspicious).

**Reference:** CWE-59 (Improper Link Resolution Before File Access)

---

### I3: Checksum Hardcoded in Source Code (Supply Chain Risk)
**Location:** `/Users/lgbarn/Personal/pdf-cli/.worktrees/phase-2-concurrency/internal/ocr/checksums.go:9-11`
**Category:** Supply Chain Security / Data Integrity
**Description:**
The SHA256 checksums for tessdata files are hardcoded directly in the source code. While this is better than no verification, it has weaknesses:

1. **No verification of checksum source:** The checksum itself is not signed or verified
2. **Build-time modification risk:** An attacker who compromises the build process can modify both the code and checksums
3. **No rotation mechanism:** Updating checksums requires code changes and redeployment
4. **Limited language support:** Only 1 language (eng) has a checksum

Current state:
```go
var KnownChecksums = map[string]string{
    "eng": "7d4322bd2a7749724879683fc3912cb542f19906c83bcc1a52132556427170b2",
}
```

**Risk:**
- **Supply chain attack:** If an attacker compromises the GitHub repository or tessdata_fast releases, they could serve malicious traineddata files
- **Limited protection:** Only protects against accidental corruption and opportunistic MITM, not sophisticated attacks
- **Incomplete coverage:** Users of non-English languages get no checksum protection (WARNING printed but download proceeds)

**Remediation:**
1. **Short-term:** Add checksums for all commonly used languages (fra, deu, spa, ita, por, rus, ara, zho, jpn, kor)
2. **Medium-term:** Fetch checksums from a signed manifest file (e.g., SHA256SUMS signed with GPG)
3. **Long-term:** Consider using official Tesseract releases with verified signatures, or bundle tessdata files in releases

**Reference:** CWE-494 (Download of Code Without Integrity Check), OWASP Top 10 2021: A08-Software and Data Integrity Failures

---

## Advisory Findings

### A1: Password Size Limit Could Reject Valid Passwords
**Location:** `/Users/lgbarn/Personal/pdf-cli/.worktrees/phase-2-concurrency/internal/cli/password.go:27-29`
**Description:** The 1KB size limit for password files may reject legitimate use cases like very long passphrases or files with metadata. While 1KB is reasonable for most passwords, some password managers or key derivation systems use longer values.

**Suggestion:** Document the 1KB limit in user-facing documentation and help text. Consider increasing to 4KB or 8KB if compatibility issues arise.

---

### A2: Password Memory Not Cleared After Use
**Location:** `/Users/lgbarn/Personal/pdf-cli/.worktrees/phase-2-concurrency/internal/cli/password.go:18-60`
**Description:** Passwords are stored in Go strings, which are immutable and not cleared from memory. Sensitive data could remain in memory until garbage collected, and may be swapped to disk or visible in core dumps.

**Suggestion:** For maximum security, use `[]byte` instead of `string` and explicitly zero the byte slice after use:
```go
defer func() {
    for i := range passwordBytes {
        passwordBytes[i] = 0
    }
}()
```
**Note:** This is defense-in-depth; Go's garbage collector makes this less critical than in C/C++, and modern OSes have memory protection.

---

### A3: No Rate Limiting on Interactive Password Prompts
**Location:** `/Users/lgbarn/Personal/pdf-cli/.worktrees/phase-2-concurrency/internal/cli/password.go:48-57`
**Description:** The interactive password prompt has no rate limiting or maximum retry attempts. An attacker with local access could attempt brute-force attacks if the CLI is used as part of a service.

**Suggestion:** If the CLI is used in any service/daemon context, add rate limiting or max retry logic. For typical CLI usage, this is low priority.

---

### A4: HTTP Without TLS for Tessdata Downloads
**Location:** `/Users/lgbarn/Personal/pdf-cli/.worktrees/phase-2-concurrency/internal/ocr/ocr.go:26`
**Description:** The TessdataURL uses HTTPS, which is good. However, the Go HTTP client configuration doesn't explicitly enforce TLS 1.2+ or certificate validation settings.

**Current:** `http.DefaultClient.Do(req)` (line 191)

**Suggestion:** Use a custom HTTP client with stricter TLS settings:
```go
client := &http.Client{
    Transport: &http.Transport{
        TLSClientConfig: &tls.Config{
            MinVersion: tls.VersionTLS12,
        },
    },
    Timeout: 5 * time.Minute,
}
resp, err := client.Do(req)
```

**Note:** `http.DefaultClient` already validates certificates by default, so this is defense-in-depth.

---

### A5: Windows Backslash Handling in Path Sanitization
**Location:** `/Users/lgbarn/Personal/pdf-cli/.worktrees/phase-2-concurrency/internal/fileio/files.go:241`
**Description:** The path sanitization splits on "/" but doesn't handle Windows backslash separators. An attacker on Windows could use "..\" to bypass the check.

**Current:**
```go
for _, part := range strings.Split(path, "/") {
    if part == ".." {
        return "", fmt.Errorf("path contains directory traversal: %s", path)
    }
}
```

**Suggestion:** Use `filepath.SplitList()` or check both separators:
```go
for _, part := range strings.FieldsFunc(path, func(r rune) bool {
    return r == '/' || r == '\\'
}) {
    if part == ".." {
        return "", fmt.Errorf("path contains directory traversal: %s", path)
    }
}
```

**Note:** `filepath.Clean()` normalizes separators, but the check happens before cleaning. The risk is mitigated by `filepath.Clean()` running afterward, but explicit handling is cleaner.

---

## Cross-Task Analysis

### Password Handling Coherence ✅
**Observation:** All 14 password-accepting commands (encrypt, decrypt, compress, rotate, extract, info, images, merge, meta, pdfa, reorder, split, text, watermark) correctly use `cli.GetPasswordSecure()` with consistent error handling.

**Strength:** The priority order (file > env > flag > prompt) is consistently applied across all commands.

**Gap:** No commands log or print password values - excellent. Error messages correctly wrap errors without exposing passwords.

---

### Path Sanitization Coverage ✅
**Observation:**
- All user-provided input paths are sanitized via `fileio.SanitizePaths()` at command entry points
- Output paths are sanitized individually via `fileio.SanitizePath()`
- The stdin marker "-" is correctly exempted from sanitization
- CopyFile() sanitizes both source and destination (line 86-94 of files.go)

**Strength:** Comprehensive application of sanitization with clear nosec directives for gosec false positives.

**Gap:** Password file paths (see I1) and potential symlink issues (see I2).

---

### Checksum Verification Flow ✅
**Observation:**
- Checksums are verified after download but before file rename (ocr.go:220-235)
- If checksum fails, the error message is clear and mentions supply chain attack
- Unknown languages warn but don't block (acceptable trade-off for usability)
- No TOCTOU issue: verification happens on the temp file before atomic rename

**Strength:** Atomic download pattern (temp file → verify → rename) prevents partial writes.

**Gap:** Limited language coverage (see I3) and no checksum source verification.

---

## Dependency Status

| Package | Version | Known CVEs | Status |
|---------|---------|-----------|--------|
| golang.org/x/term | v0.39.0 | None known | ✅ OK |
| golang.org/x/crypto | v0.47.0 | None known | ✅ OK |
| github.com/spf13/cobra | v1.10.2 | None known | ✅ OK |
| github.com/pdfcpu/pdfcpu | v0.11.1 | None known | ✅ OK |
| gopkg.in/yaml.v3 | v3.0.1 | CVE-2022-28948 (fixed in v3.0.0+) | ✅ OK |

**Note:** All dependencies are current as of January 2026. No known vulnerabilities in the security-critical dependencies.

---

## Test Coverage Analysis

### Path Sanitization Tests ✅
**File:** `internal/fileio/files_test.go:320-373`

Excellent coverage including:
- Valid paths (simple, subdirectory, absolute, stdin marker)
- Invalid paths (parent traversal, deep traversal, traversal in middle, absolute traversal)
- Batch operations

**Gap:** Missing tests for:
- Windows backslash separators (e.g., `docs\..\..\..\etc\passwd`)
- Null bytes (e.g., `docs/file.pdf\x00../../etc/passwd`)
- Unicode normalization (e.g., `docs/../` with Unicode equivalents)
- Symlink scenarios (see I2)

---

### Password Tests ✅
**File:** `internal/cli/password_test.go:1-145`

Excellent coverage including:
- All four password sources (file, env, flag, prompt)
- Priority ordering (file > env > flag)
- Edge cases (oversized file, missing file, no source)

**Gap:** Missing tests for:
- Password file with traversal path (e.g., `--password-file=../../etc/passwd`)
- Concurrent access (race conditions)
- Whitespace handling beyond newlines

---

### Checksum Tests ✅
**File:** `internal/ocr/ocr_test.go:132-204`

Good coverage including:
- Unknown language warning behavior
- Checksum function existence
- Path sanitization in downloadTessdata

**Gap:** Cannot easily test actual checksum verification without mocking HTTP client. Current tests verify the logic exists but not the full integration.

---

## OWASP Top 10 Assessment

| Category | Status | Notes |
|----------|--------|-------|
| A01: Broken Access Control | ✅ PASS | Path sanitization blocks traversal attacks. Minor gaps in I1, I2. |
| A02: Cryptographic Failures | ✅ PASS | Passwords protected in transit and storage (file permissions). SHA256 used for integrity. |
| A03: Injection | ✅ PASS | No SQL/command injection vectors. File paths validated. |
| A04: Insecure Design | ✅ PASS | Security-by-design: defense in depth, fail-secure defaults. |
| A05: Security Misconfiguration | ⚠️ ADVISORY | Default HTTP client (see A4). No secrets in code (good). |
| A06: Vulnerable Components | ✅ PASS | All dependencies current with no known CVEs. |
| A07: Identification/Auth | N/A | Not applicable (CLI tool, no authentication system). |
| A08: Data Integrity Failures | ⚠️ IMPORTANT | Checksum verification implemented but limited coverage (see I3). |
| A09: Security Logging Failures | ✅ PASS | No sensitive data logged. Errors don't leak passwords. |
| A10: Server-Side Request Forgery | ✅ PASS | Tessdata URL is hardcoded, not user-controlled. |

---

## Secrets Scanning Results

**Scope:** All modified files (password.go, files.go, checksums.go, ocr.go) + all 14 command files

✅ **No hardcoded secrets found.**

**Checked for:**
- API keys, tokens, passwords in code
- Private keys or certificates
- Base64-encoded credentials
- Connection strings
- Secrets in comments or TODOs
- Committed .env files (none in changes)

**Good practices observed:**
- Password handling code never hardcodes passwords
- Checksum is a hash, not a secret
- No credentials in test files
- Environment variable names documented but no values committed

---

## Recommendations

### Before Production (Priority Order)

1. **Fix I1 (Important):** Sanitize password file path to prevent arbitrary file reads
2. **Fix I2 (Important):** Add symlink resolution to path sanitization or document limitation
3. **Fix I3 (Important):** Add checksums for top 10 languages used globally
4. **Address A5 (Advisory):** Handle Windows backslash in path sanitization
5. **Add tests:** Symlink tests, Windows path tests, Unicode normalization tests

### Long-term Hardening

1. Consider pinning tessdata downloads to specific git commit SHAs
2. Add GPG signature verification for tessdata downloads
3. Implement password memory zeroing (A2) if used in long-running services
4. Add custom HTTP client with strict TLS settings (A4)
5. Consider bundling tessdata files in releases to eliminate download attack vector

---

## Compliance Notes

**Security Standards Addressed:**
- ✅ CWE-22 (Path Traversal) - Mitigated via SanitizePath
- ✅ CWE-257 (Storing Passwords in Recoverable Format) - Mitigated via secure input methods
- ✅ CWE-494 (Download Without Integrity Check) - Mitigated via SHA256 verification
- ⚠️ CWE-59 (Improper Link Resolution) - Partially addressed (see I2)

**OWASP ASVS Level 1 Compliance:** ✅ PASS (with noted advisories)

---

## Conclusion

Phase 3 successfully implements the three P0 security requirements with well-designed solutions:

1. **R1 (Password Security):** ✅ Excellent - multi-tier secure input with proper deprecation warnings
2. **R3 (Path Sanitization):** ⚠️ Good with gaps - effective against basic attacks, needs symlink hardening
3. **R2 (Tessdata Checksums):** ⚠️ Good with limitations - works for English, needs broader coverage

**Overall Security Posture:** Significantly improved. The application is ready for production deployment with the caveat that the three Important findings should be addressed in a near-term security patch (Phase 3.1 or Phase 4).

**Recommended Actions:**
1. Create issues for I1, I2, I3 to track remediation
2. Document limitations (1KB password limit, symlink behavior, checksum coverage) in security docs
3. Add the missing test cases identified in this audit
4. Schedule a follow-up audit after Phase 4 to verify cross-phase security coherence

---

**Report Signature:**
Security & Compliance Auditor
Date: 2026-01-31
Version: 1.0
