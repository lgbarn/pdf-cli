# Roadmap: Remaining Tech Debt

**Milestone**: Remaining Tech Debt
**Status**: Not Started
**Target**: Address all 17 technical concerns from codebase analysis (18 requirements total including R18)
**Constraint**: No breaking CLI changes, no new dependencies, maintain >= 75% coverage, `go test -race ./...` clean

---

## Success Criteria (Milestone-Level)

1. Zero `context.TODO()` calls in production (non-test) Go files
2. OCR checksum map contains entries for >= 20 languages
3. `--password` flag without `--allow-insecure-password` produces a clear error and does not return a password
4. All directory permission constants set to `0700`
5. Custom `http.Client` with explicit timeout used for tessdata downloads
6. Cleanup registry uses map-based tracking (no slice-index approach)
7. `go test -race ./...` passes on all platforms
8. Test coverage >= 75%
9. SECURITY.md references v2.0.0 as the current supported version
10. No P1 or P2 items remaining in CONCERNS.md

---

## Phase 1: OCR Download Path Hardening — COMPLETE

**Description**: Fix the OCR tessdata download path end-to-end. This is the highest-risk area because it touches context propagation, HTTP client configuration, and progress bar behavior in a single code path. Fixing it first de-risks the rest of the milestone and establishes patterns other phases depend on (the `ctx` propagation through `EnsureTessdata`).

**Requirements**: R4, R6, R12

**Complexity**: M (3 files, moderate refactoring, existing test coverage to update)

**Dependencies**: None (Wave 1)

**Risk**: Medium -- changing function signatures for context propagation could break callers. Mitigated by the fact that `EnsureTessdata` and `downloadTessdata` are internal functions with a small call surface.

**Key Files**:
- `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr.go` -- Replace `context.TODO()` with caller-provided `ctx`; replace `http.DefaultClient` with custom client; recreate progress bar per retry attempt
- `/Users/lgbarn/Personal/pdf-cli/internal/ocr/wasm.go` -- Replace `context.TODO()` with caller-provided `ctx`; update `EnsureTessdata` and `initializeTesseract` signatures
- `/Users/lgbarn/Personal/pdf-cli/internal/retry/retry.go` -- Replace `time.After` with `time.NewTimer` + explicit `Stop()`

**Success Criteria**:
- `grep -rn 'context.TODO' internal/ --include='*.go' | grep -v _test.go` returns zero results
- `grep -rn 'http.DefaultClient' internal/ --include='*.go'` returns zero results
- `grep -rn 'time.After' internal/retry/ --include='*.go'` returns zero results
- `go test -race ./internal/ocr/... ./internal/retry/...` passes
- Test coverage for `internal/ocr` and `internal/retry` >= 75%

---

## Phase 2: Security Hardening — COMPLETE

**Description**: Lock down the three remaining P1 security concerns: expand OCR checksum coverage to ~20 languages, make `--password` flag require explicit opt-in via `--allow-insecure-password`, and tighten directory permissions from 0750 to 0700. These are independent of each other but all security-scoped, so grouping them reduces context-switching.

**Requirements**: R1, R2, R3

**Complexity**: M (4 files, straightforward changes, but R2 requires careful flag interaction logic and R1 requires computing checksums for 20 languages)

**Dependencies**: None (Wave 1 -- can run in parallel with Phase 1, but sequencing after Phase 1 is preferred because R1 checksum changes touch `checksums.go` which Phase 1's download tests exercise)

**Risk**: Low-Medium -- R2 changes CLI behavior (password flag becomes non-functional without opt-in). This is technically backwards-compatible since `--password` was already deprecated, but users relying on it will need to add `--allow-insecure-password`.

**Key Files**:
- `/Users/lgbarn/Personal/pdf-cli/internal/ocr/checksums.go` -- Add SHA256 checksums for top ~20 languages (fra, deu, spa, ita, por, nld, pol, rus, jpn, chi_sim, chi_tra, kor, ara, hin, tur, vie, ukr, ces, swe, nor)
- `/Users/lgbarn/Personal/pdf-cli/internal/cli/password.go` -- Make `--password` return error unless `--allow-insecure-password` is also set; clear error message with alternatives
- `/Users/lgbarn/Personal/pdf-cli/internal/cli/flags.go` -- Add `--allow-insecure-password` flag definition
- `/Users/lgbarn/Personal/pdf-cli/internal/fileio/files.go` -- Change `DefaultDirPerm` from `0750` to `0700`
- `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr.go` -- Change `DefaultDataDirPerm` from `0750` to `0700`

**Success Criteria**:
- `go test -race ./internal/ocr/... ./internal/cli/...` passes
- `grep -c 'eng\|fra\|deu\|spa' internal/ocr/checksums.go` shows >= 20 entries in the map
- Running `pdf encrypt --password secret test.pdf` without `--allow-insecure-password` produces an error message mentioning `--password-file`, `PDF_CLI_PASSWORD`, or interactive prompt
- `grep -rn '0750' internal/ --include='*.go'` returns zero results
- Test coverage for affected packages >= 75%

---

## Phase 3: Concurrency and Error Handling Fixes

**Description**: Fix the goroutine leak pattern in parallel text extraction and OCR processing, convert cleanup registry from slice-index to map-based tracking, and add debug logging to silent error paths. These are all reliability and correctness fixes that reduce the risk of subtle bugs in production.

**Requirements**: R5, R7, R8, R9

**Complexity**: M (4 files, requires careful concurrency reasoning for R5 and R7)

**Dependencies**: Phase 1 (context propagation patterns established there are used here)

**Risk**: Medium -- concurrency changes (R5, R7) require careful reasoning about goroutine lifecycles. The cleanup registry change (R7) affects signal handling. Mitigated by existing test coverage and `-race` detection.

**Key Files**:
- `/Users/lgbarn/Personal/pdf-cli/internal/pdf/text.go` -- Add `ctx.Err()` check inside goroutine body in `extractPagesParallel` before calling `extractPageText`; add debug-level logging in `extractPageText` when errors are silently swallowed
- `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr.go` -- Add `ctx.Err()` check inside goroutine body in `processImagesParallel` before calling `ProcessImage`
- `/Users/lgbarn/Personal/pdf-cli/internal/cleanup/cleanup.go` -- Replace `[]string` + index tracking with `map[string]struct{}`; unregister by key instead of index
- `/Users/lgbarn/Personal/pdf-cli/internal/cli/password.go` -- Add printable character validation for password file content; warn on binary content detection

**Success Criteria**:
- `go test -race ./internal/pdf/... ./internal/cleanup/... ./internal/cli/...` passes
- Cleanup registry uses `map[string]struct{}` (no `idx` variable in Register function)
- `extractPageText` calls `logging.Debug` on error paths instead of returning bare `""`
- Password file containing binary data (e.g., `\x00\x01\x02`) produces a warning on stderr
- Test coverage for affected packages >= 75%

---

## Phase 4: Code Quality and Constants

**Description**: Clean up code quality issues: replace panic in test helpers with `testing.TB` + `t.Fatal()`, consolidate output filename suffix constants, change default log level from "silent" to "error", and replace `time.After` with `time.NewTimer` (if not already done in Phase 1). These are low-risk, high-confidence changes.

**Requirements**: R10, R11, R13, R14

**Complexity**: S (4 files, mechanical changes, low risk)

**Dependencies**: Phase 1 (R10 may already be done there; if so, this phase skips it)

**Risk**: Low -- R14 (default log level change) is a behavioral change but "error" is a reasonable default that only surfaces actual errors. R11 changes test helper signatures which requires updating all callers, but the test helper API surface is small.

**Key Files**:
- `/Users/lgbarn/Personal/pdf-cli/internal/testing/fixtures.go` -- Change `TempDir` and `TempFile` to accept `testing.TB` and call `t.Fatal()` instead of `panic()`; `TestdataDir` can remain as-is (no `testing.TB` available at init time) or accept `testing.TB`
- `/Users/lgbarn/Personal/pdf-cli/internal/commands/helpers.go` -- Define suffix constants (`SuffixEncrypted`, `SuffixDecrypted`, `SuffixCompressed`, `SuffixRotated`, `SuffixWatermarked`, `SuffixReordered`)
- `/Users/lgbarn/Personal/pdf-cli/internal/commands/encrypt.go` -- Use suffix constant
- `/Users/lgbarn/Personal/pdf-cli/internal/commands/decrypt.go` -- Use suffix constant
- `/Users/lgbarn/Personal/pdf-cli/internal/commands/compress.go` -- Use suffix constant
- `/Users/lgbarn/Personal/pdf-cli/internal/commands/rotate.go` -- Use suffix constant
- `/Users/lgbarn/Personal/pdf-cli/internal/commands/watermark.go` -- Use suffix constant
- `/Users/lgbarn/Personal/pdf-cli/internal/commands/reorder.go` -- Use suffix constant
- `/Users/lgbarn/Personal/pdf-cli/internal/cli/flags.go` -- Change default log level from `"silent"` to `"error"`

**Success Criteria**:
- `grep -rn 'panic(' internal/testing/fixtures.go` returns zero results (except `TestdataDir` if kept)
- `grep -rn '"_encrypted"\|"_decrypted"\|"_compressed"\|"_rotated"\|"_watermarked"\|"_reordered"' internal/commands/*.go | grep -v 'Suffix'` returns zero results (all string literals replaced with constants)
- Default log level is "error" in `flags.go`
- `go test -race ./internal/...` passes
- Test coverage >= 75%

---

## Phase 5: Performance, Documentation, and Finalization

**Description**: Address the remaining items: improve merge progress efficiency for large file sets, document WASM thread-safety limitation in README troubleshooting, update SECURITY.md for v2.0.0, and update any README sections affected by earlier phases (R18). This is the final phase -- it produces no code changes that other phases depend on.

**Requirements**: R15, R16, R17, R18

**Complexity**: S-M (R16 merge improvement requires evaluating pdfcpu's batch API; the rest are documentation)

**Risk**: Low -- documentation changes carry no regression risk. R16 merge optimization is best-effort and can be skipped if pdfcpu does not expose a suitable batch API.

**Key Files**:
- `/Users/lgbarn/Personal/pdf-cli/internal/pdf/transform.go` -- Improve `MergeWithProgress` to use batch merge with progress callback instead of N sequential merges (if pdfcpu API supports it); alternatively, use a binary-tree merge strategy to reduce from O(n^2) to O(n log n) I/O
- `/Users/lgbarn/Personal/pdf-cli/README.md` -- Add WASM thread-safety note in Troubleshooting section; update password flag documentation to mention `--allow-insecure-password`; update default log level documentation
- `/Users/lgbarn/Personal/pdf-cli/SECURITY.md` -- Update supported versions table to show v2.0.x as current, deprecate v1.x

**Success Criteria**:
- SECURITY.md contains `2.0` in the supported versions table
- README Troubleshooting section contains a subsection about WASM thread-safety
- README reflects `--allow-insecure-password` flag and "error" as default log level
- `MergeWithProgress` does not create N intermediate files for N inputs (or documents why the current approach is acceptable)
- `go test -race ./internal/pdf/...` passes
- Full CI pipeline passes: `go test -race ./...`

---

## Phase Summary

| Phase | Title | Requirements | Complexity | Wave | Risk |
|-------|-------|-------------|------------|------|------|
| 1 | OCR Download Path Hardening | R4, R6, R12 | M | 1 | Medium |
| 2 | Security Hardening | R1, R2, R3 | M | 1* | Low-Medium |
| 3 | Concurrency and Error Handling | R5, R7, R8, R9 | M | 2 | Medium |
| 4 | Code Quality and Constants | R10, R11, R13, R14 | S | 2 | Low |
| 5 | Performance, Docs, Finalization | R15, R16, R17, R18 | S-M | 3 | Low |

*Phase 2 has no hard dependency on Phase 1 but is sequenced after it because both touch `internal/ocr/ocr.go`. Running them in parallel would create merge conflicts.

## Requirement Traceability

| Req | Description | Phase |
|-----|-------------|-------|
| R1 | OCR checksum coverage (~20 languages) | 2 |
| R2 | Password flag requires --allow-insecure-password | 2 |
| R3 | Directory permissions 0700 | 2 |
| R4 | Replace context.TODO() with proper propagation | 1 |
| R5 | Goroutine ctx.Err() checks before expensive ops | 3 |
| R6 | Custom http.Client with explicit timeout | 1 |
| R7 | Map-based cleanup registry | 3 |
| R8 | Debug logging for page extraction errors | 3 |
| R9 | Password file printable character validation | 3 |
| R10 | time.NewTimer replaces time.After in retry | 1 |
| R11 | Test helpers use testing.TB + t.Fatal() | 4 |
| R12 | Progress bar recreated per retry attempt | 1 |
| R13 | Output suffix constants | 4 |
| R14 | Default log level "error" | 4 |
| R15 | WASM thread-safety documented in README | 5 |
| R16 | Merge progress efficiency improvement | 5 |
| R17 | SECURITY.md updated for v2.0.0 | 5 |
| R18 | README updated for code changes | 5 |
