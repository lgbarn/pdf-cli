# Phase 4 Research: Code Quality and Constants

**Research Date:** 2026-02-05
**Project:** pdf-cli
**Phase:** 4 - Code Quality and Constants
**Requirements:** R11, R13, R14

## Executive Summary

This research identifies all code locations requiring changes for Phase 4:
- **R11**: 4 panic() calls in test helpers, NO current callers (functions not yet used)
- **R13**: 6 suffix constants across 5 command files + 1 test file
- **R14**: 1 default value change + 1 test assertion update

All changes are straightforward refactoring with minimal risk.

---

## R11: Test Helpers Use testing.TB + t.Fatal()

### Current State

**File:** `/Users/lgbarn/Personal/pdf-cli/internal/testing/fixtures.go`

#### Panic Locations

1. **Line 14** - `TestdataDir()`:
   ```go
   if !ok {
       panic("failed to get caller information")
   }
   ```

2. **Line 36** - `TempDir()`:
   ```go
   if err != nil {
       panic("failed to create temp dir: " + err.Error())
   }
   ```

3. **Line 46** - `TempFile()`:
   ```go
   if err != nil {
       panic("failed to create temp file: " + err.Error())
   }
   ```

4. **Line 52** - `TempFile()`:
   ```go
   if _, err := f.WriteString(content); err != nil {
       _ = f.Close()
       _ = os.Remove(f.Name())
       panic("failed to write temp file: " + err.Error())
   }
   ```

### Current Function Signatures

```go
// Line 31-39
func TempDir(prefix string) (string, func()) {
    dir, err := os.MkdirTemp("", "pdf-cli-test-"+prefix+"-")
    if err != nil {
        panic("failed to create temp dir: " + err.Error())
    }
    return dir, func() { _ = os.RemoveAll(dir) }
}

// Line 41-57
func TempFile(prefix, content string) (string, func()) {
    f, err := os.CreateTemp("", "pdf-cli-test-"+prefix+"-*.pdf")
    if err != nil {
        panic("failed to create temp file: " + err.Error())
    }
    if content != "" {
        if _, err := f.WriteString(content); err != nil {
            _ = f.Close()
            _ = os.Remove(f.Name())
            panic("failed to write temp file: " + err.Error())
        }
    }
    _ = f.Close()
    return f.Name(), func() { _ = os.Remove(f.Name()) }
}
```

### Callers Analysis

**Result:** NO CALLERS FOUND

Searched for:
- `testing.TempDir(`
- `testing.TempFile(`

These functions exist but are **not currently used** in the codebase. This means:
- No existing tests will break when signatures change
- No migration of call sites is needed
- Functions can be freely refactored

### Required Changes

1. **Add `testing.TB` parameter** to both functions
2. **Replace all `panic()` calls** with `t.Fatal()`
3. **Update function signatures** to:
   ```go
   func TempDir(t testing.TB, prefix string) (string, func())
   func TempFile(t testing.TB, prefix, content string) (string, func())
   ```

### TestdataDir() Function

**Status:** NOT IN SCOPE for R11

Line 14 contains a panic but:
- Does not need a `testing.TB` parameter
- This panic is for initialization failure, not test failure
- Should remain as panic() for now
- Can be addressed in future refactoring if needed

---

## R13: Output Suffix Constants

### Current State

**All suffix strings are hardcoded** across multiple files. Each suffix appears in 3 contexts:
1. User-facing documentation (command descriptions)
2. Error messages (`validateBatchOutput`)
3. Default output generation (`outputOrDefault`, `DefaultSuffix` field)

### Suffix Locations by File

#### `/Users/lgbarn/Personal/pdf-cli/internal/commands/encrypt.go`

| Line | Context | Value |
|------|---------|-------|
| 34 | Command description | `'_encrypted'` |
| 78 | validateBatchOutput call | `"_encrypted"` |
| 100 | outputOrDefault call | `"_encrypted"` |
| 115 | DefaultSuffix field | `"_encrypted"` |

#### `/Users/lgbarn/Personal/pdf-cli/internal/commands/decrypt.go`

| Line | Context | Value |
|------|---------|-------|
| 33 | Command description | `'_decrypted'` |
| 76 | validateBatchOutput call | `"_decrypted"` |
| 98 | outputOrDefault call | `"_decrypted"` |
| 111 | DefaultSuffix field | `"_decrypted"` |

#### `/Users/lgbarn/Personal/pdf-cli/internal/commands/compress.go`

| Line | Context | Value |
|------|---------|-------|
| 33 | Command description | `'_compressed'` |
| 74 | validateBatchOutput call | `"_compressed"` |
| 96 | outputOrDefault call | `"_compressed"` |
| 109 | DefaultSuffix field | `"_compressed"` |

#### `/Users/lgbarn/Personal/pdf-cli/internal/commands/rotate.go`

| Line | Context | Value |
|------|---------|-------|
| 35 | Command description | `'_rotated'` |
| 81 | validateBatchOutput call | `"_rotated"` |
| 103 | outputOrDefault call | `"_rotated"` |
| 122 | DefaultSuffix field | `"_rotated"` |

#### `/Users/lgbarn/Personal/pdf-cli/internal/commands/watermark.go`

| Line | Context | Value |
|------|---------|-------|
| 35 | Command description | `'_watermarked'` |
| 91 | validateBatchOutput call | `"_watermarked"` |
| 108 | outputOrDefault call | `"_watermarked"` |

**Missing:** No `DefaultSuffix` field found (may need addition or doesn't use stdio pattern)

#### `/Users/lgbarn/Personal/pdf-cli/internal/commands/reorder.go`

| Line | Context | Value |
|------|---------|-------|
| 81 | DefaultSuffix field | `"_reordered"` |
| 152 | outputOrDefault call | `"_reordered"` |

**Note:** No command description or validateBatchOutput (single-file command)

### Test Files Using Suffixes

#### `/Users/lgbarn/Personal/pdf-cli/internal/commands/commands_test.go`

| Line | Context | Value |
|------|---------|-------|
| 15 | Test case | `"_compressed"` |
| 16 | Test case | `"_rotated"` |

These test the `outputOrDefault` helper function.

#### `/Users/lgbarn/Personal/pdf-cli/internal/fileio/files_test.go`

| Line | Context | Value |
|------|---------|-------|
| 105 | Test case | `"_compressed"` |
| 106 | Test case | `"_rotated"` |

These test the `GenerateOutputFilename` function.

### Constant Definitions Needed

**File:** `/Users/lgbarn/Personal/pdf-cli/internal/commands/helpers.go`

Add these constants (suggested location: after imports, before functions):

```go
// Output filename suffixes for batch operations
const (
    SuffixEncrypted   = "_encrypted"
    SuffixDecrypted   = "_decrypted"
    SuffixCompressed  = "_compressed"
    SuffixRotated     = "_rotated"
    SuffixWatermarked = "_watermarked"
    SuffixReordered   = "_reordered"
)
```

### Total Replacements Required

| Suffix | Count | Files |
|--------|-------|-------|
| `_encrypted` | 4 | encrypt.go |
| `_decrypted` | 4 | decrypt.go |
| `_compressed` | 6 | compress.go (4) + 2 test files |
| `_rotated` | 6 | rotate.go (4) + 2 test files |
| `_watermarked` | 3 | watermark.go |
| `_reordered` | 2 | reorder.go |
| **TOTAL** | **25** | **8 files** |

---

## R14: Default Log Level

### Current State

**File:** `/Users/lgbarn/Personal/pdf-cli/internal/cli/flags.go`

**Line 104:**
```go
cmd.PersistentFlags().StringVar(&logLevel, "log-level", "silent", "Log level (debug, info, warn, error, silent)")
```

**Current Default:** `"silent"`
**Required Default:** `"error"`

### Test Dependencies

**File:** `/Users/lgbarn/Personal/pdf-cli/internal/logging/logger_test.go`

**Lines 196-199 - `TestGlobalLogger()`:**
```go
// Should be silent by default
if l.Level() != LevelSilent {
    t.Errorf("Default level should be silent, got: %v", l.Level())
}
```

**Issue:** This test asserts that the global logger defaults to `LevelSilent`. However:
- This test uses `logging.Get()` which initializes a logger internally
- It does NOT test the CLI flag default
- It tests the `logging` package's default, not the `cli` package's default

**Analysis:**
```go
// In logger.go (lines 87-95)
func Get() *Logger {
    once.Do(func() {
        if global == nil {
            global = New(LevelSilent, FormatText, os.Stderr)
        }
    })
    return global
}
```

The `logging.Get()` function has its own hardcoded default of `LevelSilent` for cases where `logging.Init()` is not called first.

### Required Changes

1. **Update CLI flag default** in `internal/cli/flags.go` line 104:
   ```go
   // OLD
   cmd.PersistentFlags().StringVar(&logLevel, "log-level", "silent", "Log level (debug, info, warn, error, silent)")

   // NEW
   cmd.PersistentFlags().StringVar(&logLevel, "log-level", "error", "Log level (debug, info, warn, error, silent)")
   ```

2. **Update test comment** in `internal/logging/logger_test.go` lines 196-198:
   ```go
   // OLD
   // Should be silent by default
   if l.Level() != LevelSilent {
       t.Errorf("Default level should be silent, got: %v", l.Level())
   }

   // NEW
   // Should be silent by default (logging package default, separate from CLI flag default)
   if l.Level() != LevelSilent {
       t.Errorf("Default level should be silent, got: %v", l.Level())
   }
   ```

**Note:** The test itself does NOT need to change, only the comment. The test is checking the `logging` package's internal default, which should remain `LevelSilent`. The CLI flag default is separate and will be changed to `"error"`.

### Impact Analysis

**No Breaking Changes Expected:**

1. **User Impact:**
   - Users currently see no logs by default (silent)
   - After change, users will see error logs by default
   - This is a UX improvement (users can see when errors occur)
   - Users can still use `--log-level silent` if they want the old behavior

2. **Test Impact:**
   - No tests explicitly rely on CLI flag default being "silent"
   - Tests that want silent logging already explicitly set it via `logging.Init(LevelSilent, ...)`
   - Integration tests that run commands may now see error logs, but this is expected behavior

3. **Behavior Change:**
   - CLI commands will now log errors to stderr by default
   - Library usage via `logging.Get()` still defaults to silent
   - This separation is intentional and correct

---

## Implementation Strategy

### Order of Operations

1. **R11 First** - No callers exist, zero risk
2. **R13 Second** - Define constants, then replace all usage (use IDE find/replace)
3. **R14 Last** - Single line change + comment update

### Verification Steps

**For R11:**
```bash
# Verify no callers before changes
grep -r "testing\.TempDir\|testing\.TempFile" --include="*.go" .

# After changes, verify tests still pass
go test ./internal/testing/...
```

**For R13:**
```bash
# Find all remaining string literals (should be zero after changes)
grep -r "_encrypted\|_decrypted\|_compressed\|_rotated\|_watermarked\|_reordered" \
  --include="*.go" internal/commands/ | grep -v "Suffix"

# Verify tests pass
go test ./internal/commands/...
go test ./internal/fileio/...
```

**For R14:**
```bash
# Verify flag default changed
grep "log-level.*silent" internal/cli/flags.go  # Should be zero results

# Verify test still passes
go test ./internal/logging/logger_test.go -v -run TestGlobalLogger

# Verify CLI behavior (manual test)
./pdf-cli compress nonexistent.pdf 2>&1 | grep -i error
```

---

## Risk Assessment

| Requirement | Risk Level | Reason |
|-------------|-----------|--------|
| R11 | **NONE** | Functions not currently used, no callers to break |
| R13 | **LOW** | Simple string replacement, compiler will catch any misses |
| R14 | **LOW** | User-facing change, but improves UX; old behavior available via flag |

---

## Open Questions

1. **R13:** Should `SuffixWatermarked` follow the pattern of other commands and also check for `DefaultSuffix` field usage? (May be missing or using different pattern)

2. **R14:** Should documentation (README, help text) be updated to reflect the new default? (Not mentioned in requirements)

3. **R11:** Should `TestdataDir()` panic be addressed in this phase, or left for future work?

---

## References

### Files Analyzed

- `/Users/lgbarn/Personal/pdf-cli/internal/testing/fixtures.go`
- `/Users/lgbarn/Personal/pdf-cli/internal/commands/helpers.go`
- `/Users/lgbarn/Personal/pdf-cli/internal/commands/encrypt.go`
- `/Users/lgbarn/Personal/pdf-cli/internal/commands/decrypt.go`
- `/Users/lgbarn/Personal/pdf-cli/internal/commands/compress.go`
- `/Users/lgbarn/Personal/pdf-cli/internal/commands/rotate.go`
- `/Users/lgbarn/Personal/pdf-cli/internal/commands/watermark.go`
- `/Users/lgbarn/Personal/pdf-cli/internal/commands/reorder.go`
- `/Users/lgbarn/Personal/pdf-cli/internal/commands/commands_test.go`
- `/Users/lgbarn/Personal/pdf-cli/internal/fileio/files_test.go`
- `/Users/lgbarn/Personal/pdf-cli/internal/cli/flags.go`
- `/Users/lgbarn/Personal/pdf-cli/internal/logging/logger.go`
- `/Users/lgbarn/Personal/pdf-cli/internal/logging/logger_test.go`

### Search Patterns Used

```bash
# R11 - Find panic calls
grep -n "panic" internal/testing/fixtures.go

# R11 - Find callers
grep -r "testing\.TempDir\|testing\.TempFile" --include="*_test.go"

# R13 - Find suffix literals
grep -r "_encrypted\|_decrypted\|_compressed\|_rotated\|_watermarked\|_reordered" \
  --include="*.go" internal/

# R14 - Find log level references
grep -r "silent\|log-level" --include="*.go" internal/cli/ internal/logging/
```

---

## Conclusion

Phase 4 changes are well-scoped and low-risk:
- All code locations identified
- No complex dependencies
- Clear implementation path
- Comprehensive verification strategy available

Ready for implementation.
