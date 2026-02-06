# Phase 3 Implementation Plans Summary

## Overview
Phase 3 focuses on concurrency improvements and error handling enhancements. The work is organized into 2 waves with 3 plans total, ensuring maximum parallelism where possible.

## Wave Structure

### Wave 1 (Parallel Execution)
These plans touch separate files and can be executed in parallel:

- **PLAN-1.1**: Cleanup Registry Map Conversion (R7)
  - Files: `internal/cleanup/cleanup.go`, `internal/cleanup/cleanup_test.go`
  - No dependencies on other plans
  - 3 tasks: Map conversion, new test, race verification

- **PLAN-1.2**: Password File Validation (R9)
  - Files: `internal/cli/password.go`, `internal/cli/password_test.go`
  - No dependencies on other plans
  - 3 tasks: TDD test creation, validation implementation, verification

### Wave 2 (Sequential, depends on Wave 1 completion)
This plan touches files that are also modified in earlier requirements:

- **PLAN-2.1**: Goroutine Context Checks + Debug Logging (R5 + R8)
  - Files: `internal/pdf/text.go`, `internal/ocr/ocr.go`
  - Combined because both touch the same parallel processing code
  - 3 tasks: PDF context check + logging, OCR context check, comprehensive verification

## Requirement Mapping

| Requirement | Plan | Wave | Description |
|------------|------|------|-------------|
| R5 | 2.1 | 2 | Goroutines check ctx.Err() before expensive ops |
| R7 | 1.1 | 1 | Cleanup registry uses map-based tracking |
| R8 | 2.1 | 2 | Debug logging for page extraction errors |
| R9 | 1.2 | 1 | Password file printable character validation |

## Execution Order

### Recommended Sequence
1. Execute PLAN-1.1 and PLAN-1.2 in parallel (Wave 1)
2. After both Wave 1 plans complete, execute PLAN-2.1 (Wave 2)

### Alternative Sequential Order
If parallel execution is not desired, execute in plan ID order:
1. PLAN-1.1 (Cleanup Registry)
2. PLAN-1.2 (Password Validation)
3. PLAN-2.1 (Context Checks + Debug Logging)

## Task Count Summary
- Total Plans: 3
- Total Tasks: 9 (3 tasks per plan, adhering to max 3 rule)
- Wave 1: 6 tasks (2 plans × 3 tasks)
- Wave 2: 3 tasks (1 plan × 3 tasks)

## Files Modified by Wave

### Wave 1 Files
- `internal/cleanup/cleanup.go`
- `internal/cleanup/cleanup_test.go`
- `internal/cli/password.go`
- `internal/cli/password_test.go`

### Wave 2 Files
- `internal/pdf/text.go`
- `internal/ocr/ocr.go`

### No Overlaps
All files are modified by exactly one plan, ensuring clean separation of concerns and no merge conflicts if executed in parallel.

## Success Criteria (Phase 3)

From ROADMAP.md:
- ✓ `go test -race ./internal/pdf/... ./internal/cleanup/... ./internal/cli/...` passes
- ✓ Cleanup registry uses `map[string]struct{}` (no `idx` variable in Register function)
- ✓ `extractPageText` calls `logging.Debug` on error paths instead of returning bare `""`
- ✓ Password file containing binary data produces a warning on stderr
- ✓ Test coverage >= 75%

## Testing Strategy

Each plan includes comprehensive verification:
1. **Unit tests**: All existing tests must pass
2. **Race detection**: Run with `-race` flag on modified packages
3. **Integration**: Cross-package race tests in final verification tasks
4. **TDD where applicable**: PLAN-1.2 uses test-first approach for password validation

## Risk Assessment

| Plan | Risk Level | Mitigation |
|------|-----------|------------|
| 1.1 | Medium | Map-based tracking changes internal data structure; thorough testing with race detector |
| 1.2 | Low | Warning-only approach per CONTEXT-3.md; doesn't break existing workflows |
| 2.1 | Low | Context check is defensive (doesn't change correctness); logging is append-only |

## Dependencies
- PLAN-1.1 and PLAN-1.2 have no dependencies (Wave 1)
- PLAN-2.1 has no strict dependencies but placed in Wave 2 for clarity

## Verification Commands

After completing all plans, run:
```bash
# Comprehensive race detection
go test -race ./internal/pdf/... ./internal/cleanup/... ./internal/cli/... ./internal/ocr/...

# Coverage check
go test -cover ./internal/pdf/... ./internal/cleanup/... ./internal/cli/... ./internal/ocr/...

# Verify specific changes
grep "map\[string\]struct{}" internal/cleanup/cleanup.go
grep "logging.Debug" internal/pdf/text.go
grep "ctx.Err()" internal/pdf/text.go internal/ocr/ocr.go
grep "unicode.IsPrint" internal/cli/password.go
```

## Notes
- All plans follow max 3 tasks per plan rule
- Each task has concrete verification commands
- Done criteria are observable facts, not opinions
- File paths in verify commands are absolute per requirement
- Wave 1 plans are truly independent and can run in parallel
- Wave 2 combines R5+R8 because they touch the same parallel processing logic
