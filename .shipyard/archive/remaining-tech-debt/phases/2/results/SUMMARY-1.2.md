# SUMMARY-1.2: Directory Permissions Hardening (R3)

## Overview
Successfully tightened directory permissions from 0750 to 0700 across the codebase, removing group read/execute access from tessdata and config directories. This implements security requirement R3.

## Tasks Completed

### Task 1: Update Directory Permission Constants ✓
**Files Modified:**
- `/Users/lgbarn/Personal/pdf-cli/internal/fileio/files.go`
  - Changed `DefaultDirPerm` from `0750` to `0700` (line 15)
- `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr.go`
  - Changed `DefaultDataDirPerm` from `0750` to `0700` (line 45)

**Verification:**
- Code compiles: `go build ./...` ✓

**Commit:**
```
fix(security): tighten directory permissions from 0750 to 0700

Remove group read/execute access from tessdata and config directories.
Only the owner can now access these directories. Addresses R3.
```
Commit hash: `6d598a9`

### Task 2: Update Hardcoded Permissions in Test Files ✓
**Files Modified:**
- `/Users/lgbarn/Personal/pdf-cli/internal/config/config_test.go`
  - Changed hardcoded `0750` to `0700` in `TestLoadInvalidYAML` (line 179)
- `/Users/lgbarn/Personal/pdf-cli/internal/ocr/detect_test.go`
  - Changed all 3 instances of `0750` to `0700` (lines 19, 28, 34)
- `/Users/lgbarn/Personal/pdf-cli/internal/ocr/filesystem_test.go`
  - Changed hardcoded `0750` to `0700` in `TestFindImageFiles` (line 21)

**Verification:**
- Code compiles: `go build ./...` ✓

**Commit:**
```
test: update hardcoded permissions to match new 0700 default

Align test file permissions with the updated DefaultDirPerm and
DefaultDataDirPerm constants for consistency.
```
Commit hash: `82eb858`

### Task 3: Verify No Remaining 0750 References ✓
**Verification Steps:**
1. `grep -rn '0750' internal/ --include='*.go'` → Zero results ✓
2. `go test -race ./internal/fileio/... ./internal/ocr/... ./internal/config/...` → All passed ✓
3. `go test -race ./...` → Full suite passed ✓

**Test Results:**
- `internal/fileio`: OK (1.309s)
- `internal/ocr`: OK (5.614s)
- `internal/config`: OK (1.327s)
- Full test suite: All 13 packages passed

No commit required for this verification task.

## Impact Analysis

### Security Improvement
- **Before:** Directories created with `0750` (owner: rwx, group: r-x, other: ---)
- **After:** Directories created with `0700` (owner: rwx, group: ---, other: ---)

This change removes all group permissions, ensuring only the file owner can:
- Read directory contents
- Execute (traverse) the directory
- Write to the directory

### Affected Directories
1. **Config directories:** Created by `fileio.EnsureDir()` using `DefaultDirPerm`
2. **Tessdata directory:** Created by `getDataDir()` using `DefaultDataDirPerm`
3. **Test directories:** All test fixtures now use consistent `0700` permissions

### Compatibility
- No breaking changes to public API
- Internal constant changes only
- Test suite passes without modifications to test logic
- Pre-commit hooks passed (including go fmt, go vet, golangci-lint)

## Deviations from Plan
None. All tasks executed exactly as specified in PLAN-1.2.

## Final State
- Branch: `main`
- Total commits: 2
- All tests passing
- No remaining `0750` references in codebase
- Security requirement R3 implemented successfully

## Next Steps
The directory permissions hardening is complete. The codebase now enforces strict owner-only access to sensitive directories containing configuration and OCR data files.
