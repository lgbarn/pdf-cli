# Phase 4 Implementation Plans: Error Handling and Reliability

## Overview

Phase 4 addresses critical error handling and reliability issues across the pdf-cli codebase. The phase is split into 2 waves with 2 total plans.

## Wave Structure

### Wave 1: Error Propagation Fixes
**Plan 1.1** — Error handling fixes (R6 + R8)
- Fix parallel OCR error swallowing
- Propagate file close errors on write paths
- Dependencies: None
- Files: `internal/ocr/ocr.go`, `internal/fileio/files.go`

### Wave 2: Temp File Cleanup
**Plan 2.1** — Signal-based temp file cleanup (R11)
- Create `internal/cleanup` package
- Integrate with existing signal.NotifyContext
- Register all 8 temp file creation sites
- Dependencies: Plan 1.1
- Files: 9 files across internal/ and cmd/

## Requirements Mapping

| Requirement | Plan | Status |
|-------------|------|--------|
| R6: Parallel processing must surface all errors | 1.1 | Wave 1 |
| R8: File close errors must be checked for writes | 1.1 | Wave 1 |
| R11: Temp file cleanup on crash/interrupt | 2.1 | Wave 2 |

## Key Decisions

1. **errors.Join for error collection**: Standard library solution (Go 1.20+) for combining multiple errors from parallel operations

2. **Named return + defer closure pattern**: Industry-standard approach for checking close errors on write paths without obscuring main error

3. **Centralized cleanup package**: Single registry better than scattered signal handlers; integrates naturally with existing signal.NotifyContext from Phase 2

4. **LIFO cleanup order**: Reverse registration order ensures dependencies cleaned up correctly

5. **Unregister on normal path**: Prevents double cleanup attempts (both cleanup.Run and defer os.Remove)

## Files Modified

### Wave 1 (Plan 1.1)
- `internal/ocr/ocr.go` — Add error field to imageResult, collect all errors with errors.Join
- `internal/fileio/files.go` — Fix CopyFile close error handling
- Test files created/modified

### Wave 2 (Plan 2.1)
- `internal/cleanup/cleanup.go` — New cleanup registry package
- `cmd/pdf/main.go` — Integrate cleanup.Run with signal handler
- `internal/ocr/ocr.go` — Register 2 temp creation sites
- `internal/ocr/native.go` — Register 1 temp creation site
- `internal/pdf/text.go` — Register 1 temp creation site
- `internal/pdf/transform.go` — Register 1 temp creation site
- `internal/commands/patterns/stdio.go` — Register 1 temp creation site
- `internal/fileio/stdio.go` — Register 1 temp creation site
- `internal/fileio/files.go` — Register 1 temp creation site
- Test files created/modified

## Testing Strategy

### Unit Tests
- Error propagation in processImagesParallel and processImagesSequential
- File close error handling in CopyFile and downloadTessdata
- Concurrent cleanup registry access with race detector
- Unregister functionality
- Idempotent cleanup

### Integration Tests
- End-to-end error collection in OCR pipeline
- Close errors on read-only destinations
- Signal-based cleanup with real process signals
- Normal exit cleanup

### Manual Tests
- OCR with SIGINT during processing
- Merge with SIGTERM during operation
- Stdin operations with signal interruption
- Verification that no temp files remain

## Success Metrics

1. **Error Visibility**: All parallel processing errors surfaced to user
2. **Write Integrity**: File close errors on write paths never silently ignored
3. **Resource Cleanup**: Zero temp file leaks on signal interruption
4. **Backward Compatibility**: No breaking changes to public APIs
5. **Test Coverage**: All error paths covered with unit/integration tests

## Execution Order

1. **Plan 1.1** (Wave 1) — 3 tasks, TDD approach
   - Task 1: Fix imageResult and error collection
   - Task 2: Fix file close error handling
   - Task 3: Add integration tests

2. **Plan 2.1** (Wave 2) — 3 tasks, depends on 1.1
   - Task 1: Create cleanup package with tests
   - Task 2: Integrate at all temp file sites
   - Task 3: Add signal integration tests

Total estimated effort: 2-3 days
Complexity: Medium
Risk: Low (standard patterns, well-tested solutions)

## Notes

- Read-only file close errors intentionally left as-is (8 locations) — not actionable
- cleanup package uses mutex for thread safety but avoids complexity of channels
- Existing signal.NotifyContext from Phase 2 provides natural integration point
- Unregister pattern prevents issues with early cleanup in normal flow
- errors.Join available since Go 1.20 (codebase uses Go 1.25)
