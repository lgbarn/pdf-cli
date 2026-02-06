# Phase 2 Research: Security Hardening

**Phase**: 2
**Requirements**: R1, R2, R3
**Date**: 2026-02-05
**Status**: Research Complete

---

## Context

Phase 2 addresses three independent security hardening requirements:

1. **R1**: Expand OCR tessdata checksum verification from 1 language (eng) to ~20 most common languages
2. **R2**: Make `--password` flag non-functional unless `--allow-insecure-password` is also passed, with clear error directing users to secure alternatives
3. **R3**: Tighten directory permissions from 0750 to 0700 for tessdata and config directories

This phase was identified as Wave 1 (no dependencies on other phases) but sequenced after Phase 1 because both touch `internal/ocr/ocr.go` and R1 checksum changes are exercised by Phase 1's download tests.

---

## R1: OCR Checksum Expansion

### Current Implementation

**File**: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/checksums.go` (lines 9-22)

The current checksum map contains only 1 entry:

```go
var KnownChecksums = map[string]string{
    "eng": "7d4322bd2a7749724879683fc3912cb542f19906c83bcc1a52132556427170b2",
}
```

Two helper functions exist:
- `GetChecksum(lang string) string` - returns checksum or empty string
- `HasChecksum(lang string) bool` - returns true if checksum exists

**File**: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr.go` (lines 317-333)

Checksum verification occurs in `downloadTessdataWithBaseURL`:

```go
// Verify checksum if known
computedHash := hex.EncodeToString(hasher.Sum(nil))
if expectedHash := GetChecksum(lang); expectedHash != "" {
    if computedHash != expectedHash {
        return fmt.Errorf(
            "checksum verification failed for %s.traineddata\n  Expected: %s\n  Got:      %s\n"+
                "This may indicate a corrupted download or supply chain attack",
            lang, expectedHash, computedHash,
        )
    }
    fmt.Fprintf(os.Stderr, "Checksum verified for %s.traineddata\n", lang)
} else {
    fmt.Fprintf(os.Stderr,
        "WARNING: No checksum available for language '%s'. Computed SHA256: %s\n",
        lang, computedHash,
    )
}
```

Unknown languages print a warning with the computed SHA256 but continue. This is the correct behavior.

**Download URL Pattern**: `https://github.com/tesseract-ocr/tessdata_fast/raw/main/<LANG>.traineddata`

Example verified on 2026-02-05:
```bash
curl -sL "https://github.com/tesseract-ocr/tessdata_fast/raw/main/eng.traineddata" | sha256sum
# Returns: 7d4322bd2a7749724879683fc3912cb542f19906c83bcc1a52132556427170b2
```

### Test Coverage

**File**: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/checksums_test.go` (lines 1-38)

Three test cases exist:
1. `TestGetChecksum` - verifies known language returns checksum, unknown returns empty
2. `TestHasChecksum` - verifies boolean check works
3. `TestAllChecksumsValidFormat` - validates all checksums are 64-character lowercase hex

The format validation test will automatically cover new entries as they're added.

### Languages to Add (Top 20)

Based on the ROADMAP.md specification (line 64) and language usage statistics:

1. **fra** (French)
2. **deu** (German)
3. **spa** (Spanish)
4. **ita** (Italian)
5. **por** (Portuguese)
6. **nld** (Dutch)
7. **pol** (Polish)
8. **rus** (Russian)
9. **jpn** (Japanese)
10. **chi_sim** (Chinese Simplified)
11. **chi_tra** (Chinese Traditional)
12. **kor** (Korean)
13. **ara** (Arabic)
14. **hin** (Hindi)
15. **tur** (Turkish)
16. **vie** (Vietnamese)
17. **ukr** (Ukrainian)
18. **ces** (Czech)
19. **swe** (Swedish)
20. **nor** (Norwegian)

**Total**: 20 languages + existing "eng" = 21 entries in the map

### Implementation Plan for R1

1. Download each .traineddata file locally:
   ```bash
   curl -sL "https://github.com/tesseract-ocr/tessdata_fast/raw/main/<LANG>.traineddata" -o /tmp/<LANG>.traineddata
   ```
2. Compute SHA256 checksum:
   ```bash
   shasum -a 256 /tmp/<LANG>.traineddata
   ```
3. Add entry to `KnownChecksums` map in `internal/ocr/checksums.go`

**Files to Modify**:
- `/Users/lgbarn/Personal/pdf-cli/internal/ocr/checksums.go` (lines 9-11) - Add 20 new map entries

**Tests Affected**:
- `/Users/lgbarn/Personal/pdf-cli/internal/ocr/checksums_test.go` - Existing tests will automatically validate new entries

**Call Sites**: No changes needed - `GetChecksum()` and `HasChecksum()` remain unchanged

---

## R2: Password Flag Lockdown

### Current Implementation

**File**: `/Users/lgbarn/Personal/pdf-cli/internal/cli/password.go` (lines 13-68)

Current password reading priority (in `ReadPassword` function):

1. `--password-file` flag (if present)
2. `PDF_CLI_PASSWORD` env var (if set)
3. `--password` flag (deprecated, shows warning at line 51)
4. Interactive terminal prompt (if terminal and not CI/batch)

**Current Warning Message** (line 51):
```go
fmt.Fprintln(os.Stderr, "WARNING: --password flag is deprecated and exposes passwords in process listings. Use --password-file, PDF_CLI_PASSWORD, or interactive prompt instead.")
```

The `--password` flag currently works but prints a deprecation warning. **R2 requires making it non-functional** unless `--allow-insecure-password` is also passed.

**File**: `/Users/lgbarn/Personal/pdf-cli/internal/cli/flags.go`

Flag definition functions (lines 30-44):
```go
// AddPasswordFlag adds the --password flag to a command
func AddPasswordFlag(cmd *cobra.Command, usage string) {
    if usage == "" {
        usage = "Password for encryption/decryption"
    }
    cmd.Flags().String("password", "", usage)
}

// AddPasswordFileFlag adds the --password-file flag to a command.
func AddPasswordFileFlag(cmd *cobra.Command, usage string) {
    if usage == "" {
        usage = "Read password from file (more secure than --password)"
    }
    cmd.Flags().String("password-file", "", usage)
}
```

Two accessor functions exist (lines 58-67):
- `GetPassword(cmd *cobra.Command) string` - simple flag getter (used in 2 test files)
- `GetPasswordSecure(cmd *cobra.Command, promptMsg string) (string, error)` - secure multi-source reader (used by all commands)

### Commands Using Password Flags

All 14 commands that accept passwords use `cli.GetPasswordSecure`:

1. `/Users/lgbarn/Personal/pdf-cli/internal/commands/decrypt.go` (line 49)
2. `/Users/lgbarn/Personal/pdf-cli/internal/commands/encrypt.go` (line 50)
3. `/Users/lgbarn/Personal/pdf-cli/internal/commands/text.go` (line 66)
4. `/Users/lgbarn/Personal/pdf-cli/internal/commands/merge.go` (line 49)
5. `/Users/lgbarn/Personal/pdf-cli/internal/commands/compress.go` (line 51)
6. `/Users/lgbarn/Personal/pdf-cli/internal/commands/extract.go` (line 53)
7. `/Users/lgbarn/Personal/pdf-cli/internal/commands/images.go` (line 46)
8. `/Users/lgbarn/Personal/pdf-cli/internal/commands/info.go` (line 51)
9. `/Users/lgbarn/Personal/pdf-cli/internal/commands/meta.go` (line 60)
10. `/Users/lgbarn/Personal/pdf-cli/internal/commands/pdfa.go` (lines 109, 184)
11. `/Users/lgbarn/Personal/pdf-cli/internal/commands/reorder.go` (line 64)
12. `/Users/lgbarn/Personal/pdf-cli/internal/commands/rotate.go` (line 52)
13. `/Users/lgbarn/Personal/pdf-cli/internal/commands/split.go` (line 49)
14. `/Users/lgbarn/Personal/pdf-cli/internal/commands/watermark.go` (line 53)

All commands call `cli.GetPasswordSecure(cmd, "Enter PDF password: ")`. No command calls `cli.GetPassword()` directly.

### Test Coverage

**File**: `/Users/lgbarn/Personal/pdf-cli/internal/cli/password_test.go` (lines 1-145)

Eight test cases exist:
1. `TestReadPassword_PasswordFile` - reads from file successfully
2. `TestReadPassword_PasswordFileTooLarge` - rejects files > 1KB
3. `TestReadPassword_PasswordFileMissing` - returns error for missing file
4. `TestReadPassword_EnvVar` - reads from environment variable
5. `TestReadPassword_DeprecatedFlag` - currently succeeds with warning
6. `TestReadPassword_Priority_FileOverEnv` - verifies precedence order
7. `TestReadPassword_Priority_EnvOverFlag` - verifies precedence order
8. `TestReadPassword_NoSource` - returns empty string in CI mode

**Test #5 will need updating** - it currently expects success, but after R2 it should expect an error unless `--allow-insecure-password` is set.

**Test #6 and #7** - will need updating to include `--allow-insecure-password` if testing the flag path.

### Required Error Message (from CONTEXT-2.md)

Error message must list ALL three secure alternatives:
1. `--password-file <path>`
2. `PDF_CLI_PASSWORD` environment variable
3. Interactive prompt

Example error message:
```
Error: --password flag is insecure and disabled by default.
Use one of these secure alternatives:
  1. --password-file <path>        (recommended for automation)
  2. PDF_CLI_PASSWORD env var      (recommended for CI/scripts)
  3. Interactive prompt            (recommended for manual use)

To use --password anyway (not recommended), add --allow-insecure-password
```

### Implementation Plan for R2

**Files to Modify**:

1. `/Users/lgbarn/Personal/pdf-cli/internal/cli/flags.go`
   - Add `AddAllowInsecurePasswordFlag(cmd *cobra.Command)` function
   - Add `GetAllowInsecurePassword(cmd *cobra.Command) bool` accessor

2. `/Users/lgbarn/Personal/pdf-cli/internal/cli/password.go`
   - Modify `ReadPassword` function (lines 19-68)
   - Change step 3 (--password flag handling) to check for `--allow-insecure-password`
   - Return error with full alternatives message if `--password` used without opt-in flag
   - Update existing warning message if opt-in flag is present

3. Update all 14 command init() functions to add the new flag:
   ```go
   cli.AddAllowInsecurePasswordFlag(cmdName)
   ```

**Commands to Update** (add single line to each `init()` function):
- `/Users/lgbarn/Personal/pdf-cli/internal/commands/compress.go`
- `/Users/lgbarn/Personal/pdf-cli/internal/commands/decrypt.go`
- `/Users/lgbarn/Personal/pdf-cli/internal/commands/encrypt.go`
- `/Users/lgbarn/Personal/pdf-cli/internal/commands/extract.go`
- `/Users/lgbarn/Personal/pdf-cli/internal/commands/images.go`
- `/Users/lgbarn/Personal/pdf-cli/internal/commands/info.go`
- `/Users/lgbarn/Personal/pdf-cli/internal/commands/merge.go`
- `/Users/lgbarn/Personal/pdf-cli/internal/commands/meta.go`
- `/Users/lgbarn/Personal/pdf-cli/internal/commands/pdfa.go`
- `/Users/lgbarn/Personal/pdf-cli/internal/commands/reorder.go`
- `/Users/lgbarn/Personal/pdf-cli/internal/commands/rotate.go`
- `/Users/lgbarn/Personal/pdf-cli/internal/commands/split.go`
- `/Users/lgbarn/Personal/pdf-cli/internal/commands/text.go`
- `/Users/lgbarn/Personal/pdf-cli/internal/commands/watermark.go`

**Tests to Update**:
- `/Users/lgbarn/Personal/pdf-cli/internal/cli/password_test.go`
  - Update `newTestCmd()` helper to add `--allow-insecure-password` flag
  - Update `TestReadPassword_DeprecatedFlag` to set the opt-in flag
  - Update `TestReadPassword_Priority_*` tests that check flag precedence
  - Add new test: `TestReadPassword_PasswordFlagWithoutOptIn` - verifies error is returned
  - Add new test: `TestReadPassword_PasswordFlagWithOptIn` - verifies it works with opt-in

**Call Sites**: No other call sites - all commands use `GetPasswordSecure()` which internally calls `ReadPassword()`

### Backward Compatibility

This is a **breaking change** for users relying on `--password` without the opt-in flag. However:
- The flag was already marked as deprecated with a prominent warning
- The ROADMAP.md (line 61) acknowledges this: "R2 changes CLI behavior... technically backwards-compatible since `--password` was already deprecated"
- Users have three clear alternatives that are more secure
- Users who cannot migrate immediately can add `--allow-insecure-password` to restore old behavior

---

## R3: Directory Permissions

### Current Implementation

**File**: `/Users/lgbarn/Personal/pdf-cli/internal/fileio/files.go` (lines 13-15)

```go
const (
    // DefaultDirPerm is the default permission for creating directories.
    DefaultDirPerm = 0750
    ...
)
```

**Usage Sites** (from grep results):

1. `/Users/lgbarn/Personal/pdf-cli/internal/fileio/files.go` (line 41) - `EnsureDir()` function
   ```go
   return os.MkdirAll(path, DefaultDirPerm)
   ```

2. `/Users/lgbarn/Personal/pdf-cli/internal/config/config.go` (line 175) - config directory creation
   ```go
   if err := os.MkdirAll(dir, fileio.DefaultDirPerm); err != nil {
   ```

**File**: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr.go` (lines 44-45)

```go
const (
    ...
    // DefaultDataDirPerm is the default permission for tessdata directory.
    DefaultDataDirPerm = 0750
    ...
)
```

**Usage Sites**:

1. `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr.go` (line 167) - tessdata directory creation
   ```go
   if err := os.MkdirAll(dataDir, DefaultDataDirPerm); err != nil {
   ```

### Test Files Using 0750 (from grep results)

Five test files use hardcoded `0750` permissions:

1. `/Users/lgbarn/Personal/pdf-cli/internal/config/config_test.go` (line 179)
2. `/Users/lgbarn/Personal/pdf-cli/internal/ocr/detect_test.go` (lines 19, 28, 34)
3. `/Users/lgbarn/Personal/pdf-cli/internal/ocr/filesystem_test.go` (line 21)

These test files directly call `os.MkdirAll(path, 0750)` for test fixture setup. After changing the constants, these tests should also be updated to use `0700` for consistency.

### Security Rationale

Current permission `0750` (rwxr-x---):
- Owner: read, write, execute
- Group: read, execute
- Other: no access

Target permission `0700` (rwx------):
- Owner: read, write, execute
- Group: no access
- Other: no access

This prevents group members from reading tessdata files or config files, which may contain sensitive information (e.g., password files referenced by config).

### Implementation Plan for R3

**Files to Modify**:

1. `/Users/lgbarn/Personal/pdf-cli/internal/fileio/files.go` (line 15)
   - Change `DefaultDirPerm = 0750` to `DefaultDirPerm = 0700`

2. `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr.go` (line 45)
   - Change `DefaultDataDirPerm = 0750` to `DefaultDataDirPerm = 0700`

3. Test files (for consistency - optional but recommended):
   - `/Users/lgbarn/Personal/pdf-cli/internal/config/config_test.go` (line 179)
   - `/Users/lgbarn/Personal/pdf-cli/internal/ocr/detect_test.go` (lines 19, 28, 34)
   - `/Users/lgbarn/Personal/pdf-cli/internal/ocr/filesystem_test.go` (line 21)

**Call Sites**:
- `/Users/lgbarn/Personal/pdf-cli/internal/fileio/files.go` (line 41) - uses constant
- `/Users/lgbarn/Personal/pdf-cli/internal/config/config.go` (line 175) - uses constant
- `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr.go` (line 167) - uses constant

No changes needed at call sites - they reference the constants.

**Tests Affected**: None directly - the permission change doesn't affect test logic, only directory permissions on disk during test runs.

**Verification Command**:
```bash
grep -rn '0750' internal/ --include='*.go'
# Should return zero results after implementation
```

---

## Implementation Order Recommendation

### Wave 1 (Parallel)
- **R1** (OCR checksums) - independent, data-only change
- **R3** (Permissions) - independent, constant-only change

### Wave 2 (After Wave 1)
- **R2** (Password flag) - requires more careful testing, affects user behavior

**Rationale**: R1 and R3 are simple, mechanical changes with no behavioral impact on users. R2 changes CLI behavior and requires careful error message design and test coverage. Completing R1/R3 first reduces scope for R2 implementation and testing.

---

## Risk Assessment

| Requirement | Risk Level | Impact | Likelihood | Mitigation |
|-------------|-----------|--------|------------|------------|
| R1 - OCR checksums | **Low** | Low | Low | Incorrect checksums would cause download failures, but existing test `TestAllChecksumsValidFormat` validates hex format. Manual verification of 2-3 checksums recommended. |
| R2 - Password flag | **Medium** | High | Medium | Users relying on `--password` without `--allow-insecure-password` will see errors. Clear error message with alternatives mitigates. Comprehensive test coverage (8 existing tests + 2 new) validates behavior. |
| R3 - Permissions | **Low** | Low | Low | Tightening permissions from 0750 to 0700 cannot break existing users (more restrictive is safer). Only affects new directory creation. |

### R2 Specific Risks

**Breaking Change Risk**: Users who have `--password` in scripts will need to update.

**Mitigation**:
1. Clear error message listing all three alternatives
2. Escape hatch via `--allow-insecure-password` for users who cannot migrate immediately
3. Flag was already marked deprecated with warning message since v1.x
4. Document in CHANGELOG.md and README.md

**Test Strategy for R2**:
1. Unit tests: `internal/cli/password_test.go` (8 existing + 2 new)
2. Integration tests: Verify error message appears for all 14 commands
3. Manual testing: Test interactive prompt, env var, password file, and opt-in flag

---

## Success Criteria (from ROADMAP.md)

### R1 Success Criteria
- `grep -c 'eng\|fra\|deu\|spa' internal/ocr/checksums.go` shows >= 20 entries in the map
- Existing test `TestAllChecksumsValidFormat` passes
- `go test -race ./internal/ocr/...` passes

### R2 Success Criteria
- Running `pdf encrypt --password secret test.pdf` without `--allow-insecure-password` produces error
- Error message mentions `--password-file`, `PDF_CLI_PASSWORD`, and interactive prompt
- Running `pdf encrypt --password secret --allow-insecure-password test.pdf` succeeds
- `go test -race ./internal/cli/...` passes
- All 14 commands accept both `--password` and `--allow-insecure-password` flags

### R3 Success Criteria
- `grep -rn '0750' internal/ --include='*.go'` returns zero results
- Directories created with `fileio.EnsureDir()` have 0700 permissions
- Directories created by `ocr.getDataDir()` have 0700 permissions
- `go test -race ./internal/fileio/... ./internal/ocr/...` passes

### Phase-Level Success Criteria
- Test coverage for affected packages >= 75%
- Full CI pipeline passes: `go test -race ./...`

---

## Files Summary

### Files to Modify (R1)
1. `/Users/lgbarn/Personal/pdf-cli/internal/ocr/checksums.go` - Add 20 language checksums

### Files to Modify (R2)
1. `/Users/lgbarn/Personal/pdf-cli/internal/cli/flags.go` - Add `--allow-insecure-password` flag helpers
2. `/Users/lgbarn/Personal/pdf-cli/internal/cli/password.go` - Add opt-in check and error message
3. `/Users/lgbarn/Personal/pdf-cli/internal/cli/password_test.go` - Update tests for new behavior
4. 14 command files - Add `cli.AddAllowInsecurePasswordFlag()` to each `init()` function

### Files to Modify (R3)
1. `/Users/lgbarn/Personal/pdf-cli/internal/fileio/files.go` - Change constant from 0750 to 0700
2. `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr.go` - Change constant from 0750 to 0700
3. 3 test files (optional) - Update hardcoded 0750 to 0700 for consistency

### Total Files
- **Core changes**: 3 files (R1: 1, R2: 2, R3: 2)
- **Command updates**: 14 files (R2 only)
- **Test updates**: 4 files (R2: 1, R3: 3 optional)
- **Grand total**: 21 files

---

## Uncertainty Flags

### R1 Uncertainties
- None - implementation approach is clear and documented in `checksums.go` comments

### R2 Uncertainties
- **Error message wording**: Draft provided above, may need refinement based on user feedback
- **Flag name**: `--allow-insecure-password` is explicit but verbose. Alternative: `--insecure-password-ok` or `--unsafe-password`
- **Opt-in vs opt-out**: Current design is opt-in (must explicitly pass flag). Could be reversed with env var like `PDF_CLI_ALLOW_INSECURE_PASSWORD=1`, but opt-in is more secure.

### R3 Uncertainties
- None - permission change is straightforward and cannot break existing users

---

## References

- ROADMAP.md: `/Users/lgbarn/Personal/pdf-cli/.shipyard/ROADMAP.md` (lines 50-76)
- CONTEXT-2.md: `/Users/lgbarn/Personal/pdf-cli/.shipyard/phases/2/CONTEXT-2.md`
- tessdata_fast repository: https://github.com/tesseract-ocr/tessdata_fast
- Tesseract language documentation: https://tesseract-ocr.github.io/tessdoc/Data-Files-in-different-versions.html

---

## Next Steps

1. Review this research document with stakeholders
2. Confirm error message wording for R2
3. Confirm flag name for R2 (`--allow-insecure-password` vs alternatives)
4. Proceed to implementation:
   - Start with R1 and R3 (low risk, parallel)
   - Complete R2 after R1/R3 tests pass
5. Update CHANGELOG.md to document R2 breaking change
6. Update README.md to document `--allow-insecure-password` flag

---

**Research Complete**: 2026-02-05
**Estimated Implementation Time**: 4-6 hours (R1: 1-2h, R2: 2-3h, R3: 0.5h, testing: 0.5-1h)
