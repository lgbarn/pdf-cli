# Plan 1.2: Directory Permissions Hardening

## Context
This plan implements R3 from Phase 2: Security Hardening. It tightens directory permissions from 0750 (rwxr-x---) to 0700 (rwx------) for tessdata and config directories.

Current permission 0750 allows group members to read and execute. Target permission 0700 removes group access entirely, preventing group members from reading potentially sensitive files like config files or tessdata files.

This is a simple constant change in two files with optional test file updates for consistency.

## Dependencies
None. This plan can execute independently in Wave 1.

## Tasks

### Task 1: Update directory permission constants
**Files:**
- `/Users/lgbarn/Personal/pdf-cli/internal/fileio/files.go`
- `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr.go`

**Action:** Modify
**Description:**
Change two directory permission constants from 0750 to 0700:

1. In `/Users/lgbarn/Personal/pdf-cli/internal/fileio/files.go` at line 15:
   ```go
   const (
       // DefaultDirPerm is the default permission for creating directories.
       DefaultDirPerm = 0700  // Changed from 0750
       ...
   )
   ```

2. In `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr.go` at line 45:
   ```go
   const (
       ...
       // DefaultDataDirPerm is the default permission for tessdata directory.
       DefaultDataDirPerm = 0700  // Changed from 0750
       ...
   )
   ```

These constants are used at:
- `internal/fileio/files.go:41` - EnsureDir() function
- `internal/config/config.go:175` - config directory creation
- `internal/ocr/ocr.go:167` - tessdata directory creation

No changes needed at call sites since they reference the constants.

**Acceptance Criteria:**
- DefaultDirPerm changed from 0750 to 0700
- DefaultDataDirPerm changed from 0750 to 0700
- No other code changes required (constants are referenced, not hardcoded)
- Code compiles successfully

### Task 2: Update hardcoded permissions in test files
**Files:**
- `/Users/lgbarn/Personal/pdf-cli/internal/config/config_test.go`
- `/Users/lgbarn/Personal/pdf-cli/internal/ocr/detect_test.go`
- `/Users/lgbarn/Personal/pdf-cli/internal/ocr/filesystem_test.go`

**Action:** Modify
**Description:**
Update hardcoded 0750 permissions to 0700 in test files for consistency. These test files directly call `os.MkdirAll(path, 0750)` for test fixture setup.

Changes needed:
1. `/Users/lgbarn/Personal/pdf-cli/internal/config/config_test.go` line 179:
   ```go
   os.MkdirAll(filepath.Dir(tempFile), 0700)  // Changed from 0750
   ```

2. `/Users/lgbarn/Personal/pdf-cli/internal/ocr/detect_test.go` lines 19, 28, 34:
   ```go
   os.MkdirAll(dataDir, 0700)  // Changed from 0750
   ```

3. `/Users/lgbarn/Personal/pdf-cli/internal/ocr/filesystem_test.go` line 21:
   ```go
   os.MkdirAll(tessdataDir, 0700)  // Changed from 0750
   ```

**Acceptance Criteria:**
- All hardcoded 0750 permissions changed to 0700 in test files
- Tests compile successfully
- No functional test changes (permission difference doesn't affect test logic)

### Task 3: Verify no remaining 0750 references
**Files:** All Go files in `internal/`
**Action:** Test
**Description:**
Verify that no remaining references to 0750 permissions exist in the codebase, then run the full test suite to ensure no regressions.

Run verification commands:
```bash
# Verify no 0750 references remain
grep -rn '0750' /Users/lgbarn/Personal/pdf-cli/internal/ --include='*.go'

# Run affected test packages
go test -race ./internal/fileio/...
go test -race ./internal/ocr/...
go test -race ./internal/config/...

# Verify directories created with new permissions work correctly
go test -race ./...
```

**Acceptance Criteria:**
- `grep -rn '0750' internal/ --include='*.go'` returns zero results
- All tests in `./internal/fileio/...` pass with `-race` flag
- All tests in `./internal/ocr/...` pass with `-race` flag
- All tests in `./internal/config/...` pass with `-race` flag
- Full test suite passes: `go test -race ./...`
- No regressions in existing functionality

## Verification

Run all verification commands:

```bash
# Primary success criterion: no 0750 permissions remain
grep -rn '0750' /Users/lgbarn/Personal/pdf-cli/internal/ --include='*.go'
# Expected output: (empty)

# Verify constants are correctly set
grep 'DefaultDirPerm.*=' /Users/lgbarn/Personal/pdf-cli/internal/fileio/files.go
grep 'DefaultDataDirPerm.*=' /Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr.go
# Expected output: both show 0700

# Run affected package tests
go test -race ./internal/fileio/... ./internal/ocr/... ./internal/config/...

# Full test suite
go test -race ./...
```

## Success Criteria
- `grep -rn '0750' internal/ --include='*.go'` returns zero results
- Constants DefaultDirPerm and DefaultDataDirPerm both set to 0700
- All tests pass with `-race` flag
- No regressions in directory creation or file I/O operations
- Test coverage >= 75% maintained for affected packages
