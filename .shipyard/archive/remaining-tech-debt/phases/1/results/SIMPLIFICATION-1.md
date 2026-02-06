# Simplification Report: Phase 1
**Phase:** OCR Download Path Hardening
**Date:** 2026-02-05
**Files analyzed:** 6
**Findings:** 1 medium priority, 1 low priority

## Overall Assessment: CLEAN

Phase 1 changes are well-focused and appropriate. Both plans (1.1 and 1.2) addressed distinct concerns without introducing duplication or unnecessary complexity. The changes follow Go idioms and best practices.

## High Priority

None.

## Medium Priority

### Progress bar lifecycle management could be simplified
- **Type:** Refactor
- **Locations:** `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr.go:260-263`
- **Description:** The progress bar reset logic requires explicit nil assignment and conditional check on every retry iteration. The pattern is:
  ```go
  if bar != nil {
      progress.FinishProgressBar(bar)
      bar = nil
  }
  ```
  This is correct but slightly verbose. Inspection of `/Users/lgbarn/Personal/pdf-cli/internal/progress/progress.go:44-48` confirms that `FinishProgressBar` already handles nil gracefully with an internal nil check.
- **Suggestion:** Simplify to:
  ```go
  progress.FinishProgressBar(bar)
  bar = nil
  ```
  This removes the redundant conditional (since `FinishProgressBar` already checks for nil internally) and makes the code more readable.
- **Impact:** 2 lines saved (removing the outer if condition), slight readability improvement. Safe refactoring with no behavior change. Very low priority given the trivial nature of the improvement.

## Low Priority

### Function length: downloadTessdataWithBaseURL
- **Type:** Observation
- **Locations:** `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr.go:216-336` (120 lines)
- **Description:** The `downloadTessdataWithBaseURL` function is 120 lines long, which exceeds the typical 40-line threshold for single-purpose functions. However, the function has clear logical sections:
  1. Setup (path sanitization, context timeout, temp file creation)
  2. Retry loop with download logic
  3. Checksum verification
  4. File rename

  Each section is well-commented and the flow is sequential.
- **Suggestion:** This function could theoretically be split into helper functions like `downloadAttempt()` or `verifyChecksum()`, but the current implementation is readable and the sections are tightly coupled (they share variables like `tmpFile`, `hasher`, `bar`). Splitting would require passing many parameters or restructuring state.
- **Impact:** The current implementation is maintainable. Refactoring would not provide clear benefits and could reduce locality of reference. This is flagged only for awareness, not action.

## Summary

- **Duplication found:** 0 instances
- **Dead code found:** 0 unused definitions
- **Complexity hotspots:** 1 function (120 lines, but well-structured)
- **AI bloat patterns:** 0 instances
- **Estimated cleanup impact:** Minimal. Progress bar simplification could save 2 lines.

## Detailed Analysis

### Changes Review

**Plan 1.1 (HTTP client + timer leak fix):**
- Added `tessdataHTTPClient` with explicit timeout — appropriate hardening
- Replaced `time.After` with `time.NewTimer` in retry logic — correct resource leak fix
- Both changes are minimal, targeted, and follow best practices

**Plan 2.1 (Context propagation + progress bar per retry):**
- Propagated `context.Context` through `EnsureTessdata` methods — standard Go pattern, eliminates `context.TODO()` usage
- Added progress bar reset logic per retry — necessary for user experience, correctly implemented
- Test file updates to pass `context.Background()` — mechanical and correct

### Patterns Verified

✓ No duplicate error handling patterns across files
✓ No redundant type checks or defensive nil guards beyond what's necessary
✓ No wrapper functions with single callers
✓ No over-verbose comments explaining self-evident code
✓ All imports are used
✓ All defined functions are called

### Context Propagation Analysis

The context propagation through `EnsureTessdata` is correct and follows Go conventions:
- `Engine.EnsureTessdata(ctx)` accepts context and passes it to `downloadTessdata(ctx, ...)`
- `WASMBackend.EnsureTessdata(ctx, lang)` accepts context and passes it to `downloadTessdata(ctx, ...)`
- Both paths ultimately reach `downloadTessdataWithBaseURL(ctx, ...)` which uses the context for HTTP requests and timeout

No duplication or unnecessary indirection was introduced.

### Timer Management Analysis

The retry package's timer management is correct:
```go
timer := time.NewTimer(delay)
select {
case <-timer.C:
case <-ctx.Done():
    timer.Stop()
    return ctx.Err()
}
```

This properly prevents timer leaks by:
1. Calling `timer.Stop()` only in the cancellation path (the timer is consumed in the success path)
2. Not using `defer timer.Stop()` which would be redundant

This is the recommended pattern from the Go documentation.

## Recommendation

**No action required.** The Phase 1 changes are clean and appropriate for their purpose. The medium-priority finding about progress bar lifecycle is purely optional and only beneficial if the progress bar library already handles nil inputs gracefully. The long function observation is informational only — the current structure is maintainable.

This phase demonstrates good engineering discipline: changes are minimal, focused, and well-tested. No simplification work is needed before shipping.
