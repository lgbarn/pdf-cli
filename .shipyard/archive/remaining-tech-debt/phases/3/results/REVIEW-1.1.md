# Review: PLAN-1.1 - Cleanup Registry Map Conversion

**Reviewer:** Claude Code (Sonnet 4.5)
**Date:** 2026-02-05
**Commits Reviewed:** 29937eb, 03f7eff
**Branch:** main

---

## Stage 1: Spec Compliance

**Verdict:** PASS

### Task 1: Replace slice-based path tracking with map-based tracking
- **Status:** PASS
- **Evidence:**
  - `/Users/lgbarn/Personal/pdf-cli/internal/cleanup/cleanup.go:12` - `paths` changed from `[]string` to `map[string]struct{}`
  - Line 24-27: `idx` variable removed, replaced with map initialization and `paths[path] = struct{}{}`
  - Line 29-32: Unregister closure now uses `delete(paths, path)` instead of index-based marking
  - Line 48: Run function uses `for p := range paths` instead of reverse iteration
  - Removed empty string checks (line 48-50 in old code)
  - All 8 callers in codebase unchanged (internal/pdf/text.go:185, internal/ocr/ocr.go:236,350, etc.)
- **Verification:**
  - All tests pass with race detection: `ok github.com/lgbarn/pdf-cli/internal/cleanup 1.359s`
  - Dependent packages pass: internal/pdf (cached), internal/ocr (cached)
  - No DATA RACE warnings detected
- **Notes:**
  - Implementation matches expected diff in plan exactly
  - Map-based approach eliminates index invalidation bug where unregister after Run() would fail bounds check
  - LIFO ordering lost (map iteration is unordered), but plan explicitly notes this is acceptable

### Task 2: Add test for unregister-after-Run edge case
- **Status:** PASS
- **Evidence:**
  - `/Users/lgbarn/Personal/pdf-cli/internal/cleanup/cleanup_test.go:125-152` - New `TestUnregisterAfterRun` function added
  - Test creates file, registers it, runs cleanup, then calls unregister
  - Test verifies registry can be Reset and reused after this sequence
  - Test confirms file is removed in both cleanup cycles
- **Verification:**
  - Test passes: `--- PASS: TestUnregisterAfterRun (0.00s)`
  - All 5 cleanup tests pass (TestRegisterAndRun, TestUnregister, TestConcurrentRegister, TestRunIdempotent, TestUnregisterAfterRun)
- **Notes:**
  - Test correctly validates the edge case that was problematic with slice-based approach
  - With old code, calling unregister after Run() would be a silent no-op due to index invalidation
  - With new code, delete on a nil/empty map is safe and idempotent

### Task 3: Run race detection across all dependent packages
- **Status:** PASS
- **Evidence:**
  - Race detection output shows: `ok github.com/lgbarn/pdf-cli/internal/cleanup 1.359s`
  - Dependent packages: `ok github.com/lgbarn/pdf-cli/internal/pdf (cached)`, `ok github.com/lgbarn/pdf-cli/internal/ocr (cached)`
  - Zero DATA RACE warnings in output
- **Verification:**
  - `go test -v -race ./internal/cleanup/...` - PASS
  - `go test -race ./internal/pdf/... ./internal/ocr/...` - PASS
- **Notes:**
  - Map access correctly protected by existing mutex
  - Concurrent registration test (TestConcurrentRegister) validates thread safety

### Acceptance Criteria Verification
- ✓ `map[string]struct{}` used instead of `[]string` (cleanup.go:12)
- ✓ No `idx` variable in Register function (removed from line 24)
- ✓ All tests pass with race detection (5/5 tests pass, 1.359s)
- ✓ TestUnregisterAfterRun test added and passes
- ✓ No changes needed to callers (8 call sites verified unchanged)

---

## Stage 2: Code Quality

### Critical
None.

### Important

#### 1. Linter configuration change bundled with implementation
- **Location:** `.golangci.yaml:34` - Added `uncheckedInlineErr` to disabled-checks
- **Issue:** The golangci-lint configuration change is bundled in the same commit (29937eb) as the cleanup registry implementation. The commit message mentions "Fix gocritic false positive (uncheckedInlineErr) by disabling the check" but this affects files outside the scope of PLAN-1.1.
- **Analysis:** According to SUMMARY-1.1.md, the false positives were in `internal/cli/password_test.go` (lines 22, 42, 99), which is unrelated to cleanup registry changes and is the subject of PLAN-1.2 (password file validation).
- **Impact:** The linter change allows the commit to pass pre-commit hooks, but it's unclear if this disables a valuable check project-wide or if it's truly a false positive issue.
- **Remediation:** This doesn't block the cleanup registry work, but should be tracked separately. The uncheckedInlineErr disable should be validated in PLAN-1.2 review or reverted if it masks real issues.

### Suggestions

#### 1. Map iteration ordering documented
- **Location:** `internal/cleanup/cleanup.go:36` - Run function comment
- **Observation:** The comment on line 36 states "Run removes all registered paths in reverse order (LIFO)" but the implementation now uses `for p := range paths` which provides no ordering guarantee.
- **Remediation:** Update the comment to remove the LIFO claim:
  ```go
  // Run removes all registered paths. It is idempotent: subsequent calls
  // after the first are no-ops.
  ```

#### 2. Excellent test coverage for edge case
- **Location:** `internal/cleanup/cleanup_test.go:125-152`
- **Positive:** The TestUnregisterAfterRun test is well-designed. It validates:
  1. Calling unregister after Run() doesn't panic
  2. Files are properly cleaned up
  3. Registry can be Reset and reused
  4. Second cleanup cycle works correctly
- **Note:** This test explicitly validates the bug that motivated the refactor, making future regressions less likely.

#### 3. Code simplification achieved
- **Location:** `internal/cleanup/cleanup.go:48-52`
- **Positive:** The map-based approach eliminates:
  - Reverse iteration complexity
  - Empty string sentinel values
  - Index bounds checking in unregister
  - Fragile index capture in closures
- **Impact:** Reduced cognitive load, fewer edge cases, clearer semantics.

---

## Integration Analysis

### PLAN-1.2 Compatibility
- **PLAN-1.2 Scope:** Password file validation (internal/cli/password.go, internal/cli/password_test.go)
- **PLAN-1.1 Scope:** Cleanup registry (internal/cleanup/cleanup.go, internal/cleanup/cleanup_test.go)
- **Verdict:** No conflicts. The two plans touch completely separate subsystems with zero file overlap.
- **Note:** The .golangci.yaml change in PLAN-1.1 may have been added to fix PLAN-1.2 issues prematurely, but this doesn't create a functional conflict.

### Code Conventions
- ✓ `go fmt` - No output from `gofmt -l internal/cleanup/` (files properly formatted)
- ✓ `go vet` - No output from `go vet ./internal/cleanup/...` (passes static analysis)
- ✓ `golangci-lint` - Pre-commit hooks passed (per SUMMARY-1.1.md)
- ✓ Import ordering - Standard library only (os, sync), correctly sorted
- ✓ Test patterns - Follows existing test conventions (setup helper, t.TempDir(), proper cleanup)

### Regression Risk
- **Low:** All existing tests pass, callers unchanged, mutex protection maintained
- **Race conditions:** Zero DATA RACE warnings with `-race` flag across all dependent packages
- **Performance:** Map operations are O(1) average case vs O(n) for slice marking, no degradation
- **Backward compatibility:** API unchanged (Register/Unregister/Run/Reset signatures identical)

---

## Summary

**Verdict:** APPROVE with minor documentation fix recommended

PLAN-1.1 successfully converts the cleanup registry from slice-based to map-based tracking, eliminating the index invalidation bug identified in review R7. The implementation is clean, well-tested, and maintains backward compatibility with all callers.

**Strengths:**
- Exact match to plan specification
- Excellent test coverage including edge case validation
- Simplified code with clearer semantics
- Zero race conditions detected
- All acceptance criteria met

**Areas for improvement:**
- Update Run() comment to remove obsolete LIFO claim (suggestion)
- Track .golangci.yaml change separately if it affects unrelated code (important, non-blocking)

**Counts:**
- Critical: 0
- Important: 1 (linter config scope)
- Suggestions: 3 (1 documentation fix, 2 positive observations)

---

## Appendix: Verification Commands

```bash
# Tests with race detection
go test -v -race ./internal/cleanup/...
# Output: PASS (1.359s), 5/5 tests pass

# Dependent package tests
go test -race ./internal/pdf/... ./internal/ocr/...
# Output: ok (cached) for both packages

# Code quality checks
go vet ./internal/cleanup/...
# Output: (no issues)

gofmt -l internal/cleanup/
# Output: (no files listed)

# Caller verification
git diff 29937eb~1..03f7eff -- internal/pdf/ internal/ocr/
# Output: (no changes outside cleanup package except .golangci.yaml)
```

## Related Documents
- Plan: `.shipyard/phases/3/plans/PLAN-1.1.md`
- Summary: `.shipyard/phases/3/results/SUMMARY-1.1.md`
- Research: `.shipyard/phases/3/RESEARCH.md` (R7 section)
- Commits: 29937eb (implementation), 03f7eff (edge case test)
