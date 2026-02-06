# Milestone Report: Remaining Tech Debt

**Completed:** 2026-02-05
**Phases:** 5/5 complete
**Total Commits:** 33 (implementation + artifacts)
**Files Changed:** 39 source files (416 insertions, 85 deletions)
**Test Coverage:** 80.6% (threshold: 75%)
**Race Detection:** Clean (`go test -race ./...` passes)

---

## Phase Summaries

### Phase 1: OCR Download Path Hardening
**Requirements:** R4, R6, R10, R12
**Status:** COMPLETE

- Replaced all `context.TODO()` with proper context propagation through OCR download path
- Replaced `http.DefaultClient` with custom `http.Client` with explicit timeout
- Replaced `time.After` with `time.NewTimer` + explicit `Stop()` in retry logic
- Progress bar recreated on each retry attempt
- Coverage: OCR 78.4%, Retry 87.5%

### Phase 2: Security Hardening
**Requirements:** R1, R2, R3
**Status:** COMPLETE

- Added SHA256 checksums for 21 OCR language data files
- `--password` flag now requires `--allow-insecure-password` opt-in; clear error with alternatives
- Directory permissions tightened from 0750 to 0700
- Coverage: OCR 78.4%, CLI 82.7%

### Phase 3: Concurrency and Error Handling
**Requirements:** R5, R7, R8, R9
**Status:** COMPLETE

- Added `ctx.Err()` checks in goroutines before expensive operations (text extraction, OCR)
- Converted cleanup registry from slice-index to map-based tracking (eliminates race window)
- Added debug-level logging to silent error paths in text extraction
- Password file reader validates for binary content with user warning
- Coverage: Cleanup 95.8%, PDF 83.3%, CLI 84.1%, OCR 78.0%

### Phase 4: Code Quality and Constants
**Requirements:** R11, R13, R14
**Status:** COMPLETE

- Test helpers refactored to accept `testing.TB` and use `t.Fatal()` instead of `panic()`
- 21 string literals consolidated into 6 named suffix constants
- Default log level changed from "silent" to "error"
- R10 (time.After) already completed in Phase 1, skipped
- Coverage: Commands 80.9%, CLI 84.1%

### Phase 5: Performance, Documentation, and Finalization
**Requirements:** R15, R16, R17, R18
**Status:** COMPLETE

- Documented O(N²) merge trade-off in transform.go (pdfcpu lacks progress callbacks)
- SECURITY.md updated for v2.0.0 (2.0.x and 1.3.x supported)
- README: WASM OCR thread-safety limitation documented in troubleshooting
- README: Password flag docs updated (--allow-insecure-password, default log level "error", security warning)
- Coverage: 80.6% overall

---

## Key Decisions

1. **R16 Merge Optimization:** Research determined pdfcpu's `MergeCreateFile` lacks progress callbacks. The O(N²) incremental approach is necessary for progress reporting UX. Documented rather than rewritten.
2. **R2 Password Lockdown:** `--password` flag produces hard error (not warning) without `--allow-insecure-password`. Users directed to `--password-file`, `PDF_CLI_PASSWORD`, or interactive prompt.
3. **R7 Cleanup Registry:** Map-based tracking chosen over ordered alternatives because cleanup order is not critical (OS handles it). Eliminates race condition entirely.
4. **R14 Log Level:** Changed default from "silent" to "error" — surfaces actual errors without noise. Logging package internal default (LevelSilent) left unchanged.
5. **golangci.yaml:** `uncheckedInlineErr` disabled globally to fix false positives from gocritic.

---

## Documentation Status

- **README.md:** Updated with WASM troubleshooting, password security docs, log level default, security warning
- **SECURITY.md:** Updated version support table for v2.0.0
- **Code comments:** MergeWithProgress trade-off documented in transform.go
- **No separate docs/ directory:** Not needed for this CLI tool

---

## Quality Gates Summary

| Phase | Reviews | Verification | Security | Simplification | Documentation |
|-------|---------|-------------|----------|---------------|---------------|
| 1     | PASS    | PASS        | PASS     | CLEAN         | NO_ACTION     |
| 2     | PASS    | PASS        | PASS     | CLEAN         | NO_ACTION     |
| 3     | PASS    | PASS        | PASS     | CLEAN         | NO_ACTION     |
| 4     | PASS    | PASS        | PASS     | CLEAN         | NO_ACTION     |
| 5     | PASS    | PASS        | PASS     | CLEAN         | PASS          |

---

## Known Issues

- **golangci.yaml global disable:** `uncheckedInlineErr` disabled globally rather than with targeted exclusions. Low priority.
- **Test helper zero callers:** `internal/testing/fixtures.go` TempDir/TempFile refactored but currently unused. Preparatory for future tests.

---

## Metrics

- **Files created:** 3 (password_test.go, cleanup_test.go additions, helpers.go constants)
- **Files modified:** 39
- **Total commits:** 33
- **Test coverage:** 80.6%
- **Security issues:** 0 (gosec clean)
- **Lint issues:** 0

---

## Requirement Traceability

| Req | Description | Phase | Status |
|-----|-------------|-------|--------|
| R1  | OCR checksum coverage (~20 languages) | 2 | DONE |
| R2  | Password flag requires --allow-insecure-password | 2 | DONE |
| R3  | Directory permissions 0700 | 2 | DONE |
| R4  | Replace context.TODO() with proper propagation | 1 | DONE |
| R5  | Goroutine ctx.Err() checks | 3 | DONE |
| R6  | Custom http.Client with explicit timeout | 1 | DONE |
| R7  | Map-based cleanup registry | 3 | DONE |
| R8  | Debug logging for page extraction errors | 3 | DONE |
| R9  | Password file printable character validation | 3 | DONE |
| R10 | time.NewTimer replaces time.After in retry | 1 | DONE |
| R11 | Test helpers use testing.TB + t.Fatal() | 4 | DONE |
| R12 | Progress bar recreated per retry attempt | 1 | DONE |
| R13 | Output suffix constants | 4 | DONE |
| R14 | Default log level "error" | 4 | DONE |
| R15 | WASM thread-safety documented in README | 5 | DONE |
| R16 | Merge progress efficiency (documented) | 5 | DONE |
| R17 | SECURITY.md updated for v2.0.0 | 5 | DONE |
| R18 | README updated for code changes | 5 | DONE |
