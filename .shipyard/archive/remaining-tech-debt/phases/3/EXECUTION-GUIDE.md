# Phase 3 Execution Guide

## Quick Start

Execute plans in wave order for maximum parallelism:

```bash
# Wave 1 (Execute in parallel or sequentially)
# Terminal 1:
cd /Users/lgbarn/Personal/pdf-cli
# Follow PLAN-1.1.md (Cleanup Registry)

# Terminal 2 (or run after PLAN-1.1):
cd /Users/lgbarn/Personal/pdf-cli
# Follow PLAN-1.2.md (Password Validation)

# Wave 2 (Execute after Wave 1 completes)
cd /Users/lgbarn/Personal/pdf-cli
# Follow PLAN-2.1.md (Context Checks + Debug Logging)
```

## Plan Details

### PLAN-1.1: Cleanup Registry Map Conversion (Wave 1)
**Requirement**: R7
**Files**: `/Users/lgbarn/Personal/pdf-cli/internal/cleanup/cleanup.go`, `/Users/lgbarn/Personal/pdf-cli/internal/cleanup/cleanup_test.go`
**Tasks**: 3
**Duration Estimate**: 30-45 minutes

**What it does**: Converts cleanup registry from slice-based to map-based tracking to eliminate index invalidation issues.

**Key changes**:
- Replace `paths []string` with `paths map[string]struct{}`
- Remove `idx` variable from Register function
- Update Run() to iterate over map instead of slice
- Add test for "unregister after Run()" edge case

### PLAN-1.2: Password File Validation (Wave 1)
**Requirement**: R9
**Files**: `/Users/lgbarn/Personal/pdf-cli/internal/cli/password.go`, `/Users/lgbarn/Personal/pdf-cli/internal/cli/password_test.go`
**Tasks**: 3
**Duration Estimate**: 30-45 minutes
**TDD**: Yes (write tests first)

**What it does**: Adds validation for password file content to detect binary data, warning-only approach.

**Key changes**:
- Add `unicode.IsPrint()` validation after reading password file
- Count non-printable characters (excluding common whitespace)
- Print warning to stderr if non-printable characters found
- Still return password content (warning-only per CONTEXT-3.md)

### PLAN-2.1: Goroutine Context Checks + Debug Logging (Wave 2)
**Requirements**: R5, R8
**Files**: `/Users/lgbarn/Personal/pdf-cli/internal/pdf/text.go`, `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr.go`
**Tasks**: 3
**Duration Estimate**: 45-60 minutes

**What it does**: Adds context cancellation checks in goroutines and debug logging for extraction errors.

**Key changes**:
- Add `ctx.Err()` check in `extractPagesParallel` goroutine (before expensive operation)
- Add `ctx.Err()` check in `processImagesParallel` goroutine (before OCR operation)
- Add `logging.Debug()` calls in all three error paths of `extractPageText()`
- Import `"github.com/lgbarn/pdf-cli/internal/logging"`

## Wave Dependency Graph

```
Wave 1 (Parallel)
├── PLAN-1.1 (Cleanup Registry)
│   └── No dependencies
└── PLAN-1.2 (Password Validation)
    └── No dependencies

Wave 2 (After Wave 1)
└── PLAN-2.1 (Context Checks + Debug Logging)
    └── No dependencies
```

## Verification Checklist

After completing all plans, verify:

```bash
# 1. All tests pass with race detector
go test -race /Users/lgbarn/Personal/pdf-cli/internal/pdf/... \
              /Users/lgbarn/Personal/pdf-cli/internal/cleanup/... \
              /Users/lgbarn/Personal/pdf-cli/internal/cli/... \
              /Users/lgbarn/Personal/pdf-cli/internal/ocr/...

# 2. Verify specific requirement implementations
# R7: Cleanup registry uses map
grep "map\[string\]struct{}" /Users/lgbarn/Personal/pdf-cli/internal/cleanup/cleanup.go

# R7: No idx variable in Register
! grep "idx :=" /Users/lgbarn/Personal/pdf-cli/internal/cleanup/cleanup.go

# R8: Debug logging in extractPageText
grep "logging.Debug" /Users/lgbarn/Personal/pdf-cli/internal/pdf/text.go

# R5: Context checks in goroutines
grep -A2 "go func" /Users/lgbarn/Personal/pdf-cli/internal/pdf/text.go | grep "ctx.Err()"
grep -A3 "go func" /Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr.go | grep "ctx.Err()"

# R9: Unicode validation in password reading
grep "unicode.IsPrint" /Users/lgbarn/Personal/pdf-cli/internal/cli/password.go

# 3. Check test coverage
go test -cover /Users/lgbarn/Personal/pdf-cli/internal/pdf/... \
                /Users/lgbarn/Personal/pdf-cli/internal/cleanup/... \
                /Users/lgbarn/Personal/pdf-cli/internal/cli/... \
                /Users/lgbarn/Personal/pdf-cli/internal/ocr/...
```

## Success Criteria (from ROADMAP.md)

- [x] `go test -race ./internal/pdf/... ./internal/cleanup/... ./internal/cli/...` passes
- [x] Cleanup registry uses `map[string]struct{}` (no `idx` variable in Register function)
- [x] `extractPageText` calls `logging.Debug` on error paths
- [x] Password file containing binary data produces warning on stderr
- [x] Test coverage >= 75%

## Common Issues and Solutions

### Issue: Tests fail after cleanup registry change
**Solution**: Make sure to initialize the map with `if paths == nil { paths = make(map[string]struct{}) }` in Register function.

### Issue: Password validation test fails on stderr capture
**Solution**: Ensure you're redirecting os.Stderr correctly and closing the pipe writer before reading from the pipe reader.

### Issue: Context checks don't prevent work
**Solution**: Make sure the `ctx.Err()` check is INSIDE the goroutine body, before the expensive operation, not in the outer loop.

### Issue: Debug logging doesn't appear
**Solution**: Debug logs only appear when `--log-level debug` is set. Default is silent. This is expected behavior.

## Performance Notes

All changes in Phase 3 have negligible performance impact:
- Context checks: O(1) atomic operation
- Map operations: O(1) average case (vs O(n) for slice marking)
- Unicode validation: O(n) where n = password length (max 1KB)
- Debug logging: Only when log level is debug (disabled by default)

## File Modification Summary

| File | Plan | Wave | Changes |
|------|------|------|---------|
| `internal/cleanup/cleanup.go` | 1.1 | 1 | Convert slice to map, update Register/Run logic |
| `internal/cleanup/cleanup_test.go` | 1.1 | 1 | Add TestUnregisterAfterRun |
| `internal/cli/password.go` | 1.2 | 1 | Add unicode validation and warning |
| `internal/cli/password_test.go` | 1.2 | 1 | Add binary content tests |
| `internal/pdf/text.go` | 2.1 | 2 | Add ctx check in goroutine, debug logging |
| `internal/ocr/ocr.go` | 2.1 | 2 | Add ctx check in goroutine |

**Total files modified**: 6
**Total new tests**: 3 (TestUnregisterAfterRun, TestReadPassword_BinaryContentWarning, TestReadPassword_PrintableContent_NoWarning)

## Time Estimate

- Wave 1: 1-1.5 hours (if executed in parallel) or 1-1.5 hours (if sequential)
- Wave 2: 45-60 minutes
- **Total**: ~2-2.5 hours for complete Phase 3 implementation

## Next Steps

After completing Phase 3:
1. Update task #3 status to completed
2. Proceed to Phase 4: Code Quality and Constants (task #4)
3. Update ROADMAP.md to mark Phase 3 as complete
