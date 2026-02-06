# Documentation Report: Phase 2 - Security Hardening
**Date:** 2026-02-05
**Phase:** Security Hardening (Wave 1 + Wave 2)

## Summary

Phase 2 implemented three security improvements:
- **R1**: OCR checksum expansion (20 new languages) - Internal only, no user-facing docs needed
- **R2**: Password flag lockdown (BREAKING CHANGE) - Requires immediate README update
- **R3**: Directory permissions hardening (0750→0700) - Internal only, minimal docs impact

**Documentation Status:**
- API/Code docs: 3 files need documentation
- User-facing docs: 1 critical update required (README.md)
- Architecture docs: 1 minor update recommended

## Critical Finding: README is Outdated (R2 Breaking Change)

### Issue
The `--password` flag is now **non-functional by default** and requires `--allow-insecure-password` to work (commit ea5075c, 354e688). However, the README still shows examples using `--password` without the new flag, which will cause errors for users.

**Affected README sections:**
- Line 463: Global Options table lists `--password` as "deprecated" but doesn't mention it's now blocked
- Line 482: Dry-run example uses `--password secret` without opt-in flag (will fail)
- Line 528-529: Example shows `--password mysecret` with warning but no opt-in flag (will fail)

### Recommendation
**Update README.md NOW** - This is a breaking change that affects user scripts and workflows. The documentation must accurately reflect the new behavior before release.

**Rationale:**
1. Breaking changes require immediate documentation updates
2. Users will copy examples from README and encounter errors
3. Clear migration guidance is essential for adoption
4. The opt-in flag is the escape hatch users need to know about

**Deferring to Phase 5 would be incorrect** because:
- Phase 5 is labeled "Code Quality Improvements" (see archived phase 5 README.md)
- Phase 5 does NOT mention R2 or password flag documentation
- Users need accurate docs for this release, not the next phase

## Required Changes to README.md

### 1. Update Global Options Table (Line 463)
**Current:**
```markdown
| `--password` | `-P` | Password for encrypted PDFs (deprecated, use --password-file) |
```

**Recommended:**
```markdown
| `--password` | `-P` | Password for encrypted PDFs (INSECURE, blocked by default - requires --allow-insecure-password) |
| `--allow-insecure-password` | | Allow using --password flag (not recommended, see security note below) |
```

### 2. Update "Working with Encrypted PDFs" Section (Lines 505-545)
**Current text (lines 526-532):**
```markdown
**4. Command-line flag (deprecated, shows warning):**
```bash
pdf info secure.pdf --password mysecret
# WARNING: --password flag exposes passwords in process listings
```

Password sources are checked in the above order. The first available source is used.
```

**Recommended replacement:**
```markdown
**4. Command-line flag (INSECURE, blocked by default):**

The `--password` flag is now blocked for security reasons. It requires explicit opt-in:
```bash
pdf info secure.pdf --password mysecret --allow-insecure-password
# WARNING: --password flag exposes passwords in process listings
```

If you try to use `--password` without `--allow-insecure-password`, you'll get an error with migration guidance.

Password sources are checked in this priority order:
1. `--password-file` (most secure for automation)
2. `PDF_CLI_PASSWORD` env var (secure for CI/scripts)
3. `--password` + `--allow-insecure-password` (insecure fallback)
4. Interactive prompt (most secure for manual use)
```

### 3. Fix Dry-Run Example (Line 482)
**Current:**
```bash
pdf encrypt document.pdf --password secret --dry-run
```

**Recommended:**
```bash
pdf encrypt document.pdf --password-file pass.txt --dry-run
```

**Rationale:** Show the secure method in examples, not the insecure one. Don't promote the `--allow-insecure-password` workaround.

### 4. Add Security Note After Global Options
Insert new subsection after line 469:

```markdown
### Password Security Note

**IMPORTANT:** The `--password` flag is now blocked by default because it:
- Exposes passwords in process listings (`ps aux`, shell history)
- Creates security vulnerabilities in multi-user systems
- Violates best practices for password handling

**Secure alternatives:**
1. `--password-file <path>` - Store password in a file with restricted permissions (0600)
2. `PDF_CLI_PASSWORD` env var - Set in shell environment or CI secrets
3. Interactive prompt - Let the tool prompt you securely

**Emergency use only:** If you absolutely must use `--password` (not recommended), add `--allow-insecure-password`:
```bash
pdf info secure.pdf --password secret --allow-insecure-password
```

This is provided as an escape hatch for urgent migration scenarios but should not be used in production or scripts.
```

## API/Code Documentation

### 1. internal/cli/flags.go
**Status:** Well-documented
**New functions added:**
- `AddAllowInsecurePasswordFlag()` - Properly documented with clear purpose
- `GetAllowInsecurePassword()` - Simple accessor, minimal docs needed

**Recommendation:** No changes needed. Function names and comments are self-explanatory.

### 2. internal/cli/password.go
**Status:** Implementation updated, docs adequate
**Changes:** `ReadPassword()` now returns error when `--password` used without opt-in

**Recommendation:** No changes needed. The error message itself serves as documentation:
```
--password flag is insecure and disabled by default.
Use one of these secure alternatives:
  1. --password-file <path>        (recommended for automation)
  2. PDF_CLI_PASSWORD env var      (recommended for CI/scripts)
  3. Interactive prompt            (recommended for manual use)

To use --password anyway (not recommended), add --allow-insecure-password
```

### 3. internal/ocr/checksums.go
**Status:** Well-documented
**Changes:** Added 20 new language checksums to `KnownChecksums` map

**Current documentation (lines 1-12):**
```go
// Package ocr provides optical character recognition functionality.
package ocr

import "fmt"

// KnownChecksums maps language codes to their expected SHA256 checksums.
// These checksums are used to verify the integrity of downloaded tessdata files.
var KnownChecksums = map[string]string{
    // Languages listed in alphabetical order
    ...
}
```

**Recommendation:** No changes needed. The package comment and variable comment clearly explain purpose. The alphabetical ordering comment is helpful for maintainers.

## Architecture Documentation

### docs/architecture.md
**Status:** Mostly accurate, one minor update recommended

**Current text (lines 60-63):**
```markdown
### cli/
- Root command setup and version info
- Shared flag definitions (output, password, verbose, force)
- Secure password reading (ReadPassword) with 4-tier priority: password-file, env var, flag (deprecated), interactive prompt
- Output formatting helpers
```

**Recommended update (line 62):**
```markdown
- Secure password reading (ReadPassword) with 4-tier priority: password-file, env var, flag (blocked by default, requires opt-in), interactive prompt
```

**Rationale:** Change "deprecated" to "blocked by default, requires opt-in" to accurately reflect the security model.

## Documentation Not Needed

### R1: OCR Checksum Expansion
**Files changed:** `internal/ocr/checksums.go`
**User impact:** None - internal data structure
**Rationale:** Checksum verification is transparent to users. The expanded language support doesn't change the API or user experience.

### R3: Directory Permissions Hardening
**Files changed:**
- `internal/fileio/files.go` (DefaultDirPerm constant)
- `internal/ocr/ocr.go` (DefaultDataDirPerm constant)
**User impact:** None - internal implementation detail
**Rationale:** Permission changes are internal security improvements. Users don't need to know or care about directory permission modes.

## Gaps

No critical gaps identified, but consider these future enhancements:

1. **Migration guide for v2.0**: Document the breaking change from v1.x where `--password` worked without opt-in
2. **Security best practices page**: Create `docs/security.md` with comprehensive password handling guidance
3. **CHANGELOG.md**: Ensure R2 breaking change is prominently documented in the release notes

## Recommendations

### Immediate (Block Release)
1. ✅ **UPDATE README.md** - Fix all `--password` examples and add security note (see sections above)
2. ✅ **Update docs/architecture.md** - Change "deprecated" to "blocked by default" (line 62)

### Before Release
3. Add migration note to CHANGELOG.md:
   ```markdown
   ## [2.0.0] - 2026-02-05

   ### BREAKING CHANGES
   - **Password flag now requires opt-in**: The `--password` flag is now blocked by default
     for security. Use `--password-file`, `PDF_CLI_PASSWORD` env var, or interactive prompt
     instead. If you must use `--password`, add `--allow-insecure-password` flag.
   ```

### Nice to Have (Can Defer)
4. Create `docs/security.md` with detailed password handling best practices
5. Add examples of secure password file creation to README:
   ```bash
   # Create password file with secure permissions
   echo "mysecret" > pass.txt
   chmod 600 pass.txt
   ```

## Verification

Run these checks before release:

```bash
# Verify all README examples are syntactically valid
grep -n "pdf.*--password[^-]" README.md  # Should return 0 results (all should use --password-file or --allow-insecure-password)

# Verify architecture docs are accurate
grep -n "deprecated" docs/architecture.md  # Should not find password flag references

# Test that examples work
pdf encrypt testdata/sample.pdf --password-file <(echo "test") -o /tmp/test.pdf
pdf decrypt /tmp/test.pdf --password-file <(echo "test") -o /tmp/unlocked.pdf
```

## Summary of Documentation Impact

| Category | Files | Status | Urgency |
|----------|-------|--------|---------|
| **User-facing** | README.md | ⚠️ **OUTDATED** | **CRITICAL - Block release** |
| **Architecture** | docs/architecture.md | Minor update | Before release |
| **API/Code** | internal/cli/*.go | Adequate | No action needed |
| **API/Code** | internal/ocr/checksums.go | Adequate | No action needed |
| **Release notes** | CHANGELOG.md | Missing | Before release |

## Conclusion

Phase 2 documentation status: **REQUIRES IMMEDIATE ACTION**

The R2 password flag lockdown is a **breaking security change** that requires README.md updates before release. The current README shows examples that will fail, which would create a poor user experience and support burden.

**DO NOT defer to Phase 5** - the README must be updated as part of Phase 2 completion. Phase 5 focuses on code quality improvements (magic numbers, adaptive parallelism, CI tooling) and does not include documentation work for R2.

**Estimated effort:** 30-45 minutes to update README.md, architecture.md, and add CHANGELOG entry.
