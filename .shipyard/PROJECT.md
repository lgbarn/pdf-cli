# pdf-cli Remaining Tech Debt

## Description

Address all 17 technical concerns identified in the fresh codebase analysis following the v2.0.0 release. These span security hardening (OCR checksum coverage, password flag lockdown), reliability improvements (context propagation, goroutine leak prevention, HTTP client hardening), error handling (silent failures, cleanup race conditions), and code quality cleanup (timer leaks, test helper quality, logging defaults, dependency hygiene).

This milestone completes the technical debt remediation started in "Clean Baseline" and brings the codebase to a fully clean state with no known P1-P3 concerns remaining.

## Goals

1. Eliminate all P1 high-priority concerns (context.TODO removal, OCR checksum expansion, password flag lockdown)
2. Fix all P2 medium-priority concerns (goroutine leaks, HTTP client timeout, password file validation, silent errors, cleanup race condition)
3. Clean up all P3 low-priority concerns (timer leaks, test helper panics, progress bar retries, dependency pseudo-versions, output suffix constants, directory permissions, logging defaults, WASM documentation, merge efficiency)
4. Update stale documentation (SECURITY.md supported versions, README where applicable)

## Non-Goals

- Adding new PDF features or CLI commands
- Changing the overall architecture or package structure
- Bumping to v3.0.0 (all changes are backwards-compatible)
- Achieving test coverage beyond current levels (maintain >= 75%)
- Forking upstream dependencies (pseudo-version cleanup is best-effort)

## Requirements

### Security (P1)
- R1: OCR tessdata downloads must have SHA256 checksum verification for the top ~20 most common languages. Unknown languages continue with a warning.
- R2: The `--password` flag must be non-functional unless `--allow-insecure-password` is also passed. Clear error message directing users to secure alternatives.
- R3: Directory permissions for tessdata and config directories should use 0700 instead of 0750.

### Reliability (P1/P2)
- R4: All `context.TODO()` calls in production code replaced with proper context propagation from callers.
- R5: Goroutines in parallel text extraction and OCR processing must check `ctx.Err()` before performing expensive operations.
- R6: HTTP client for tessdata downloads must use a custom `http.Client` with explicit timeout (not `http.DefaultClient`).
- R7: Cleanup registry must use map-based tracking instead of slice-index approach to eliminate race condition window.

### Error Handling (P2)
- R8: Text extraction from individual pages must log errors at debug level instead of silently returning empty strings.
- R9: Password file reader must validate content contains only printable characters and warn if suspicious binary content detected.

### Code Quality (P3)
- R10: `time.After` in retry logic replaced with `time.NewTimer` and explicit `Stop()` call.
- R11: Test helper functions must accept `testing.TB` and call `t.Fatal()` instead of `panic()`.
- R12: Progress bar must be recreated/reset on each retry attempt during tessdata downloads.
- R13: Output filename suffixes consolidated into named constants in a central location.
- R14: Default log level changed from "silent" to "error".
- R15: WASM backend thread-safety limitation documented in README troubleshooting section.
- R16: Merge progress implementation improved to reduce I/O overhead for large file sets where possible.

### Documentation
- R17: SECURITY.md updated to reflect v2.0.0 as current supported version.
- R18: Any README sections affected by code changes updated accordingly.

## Non-Functional Requirements

- All changes must pass `go test -race ./...`
- Test coverage must remain >= 75%
- CI pipeline must pass cleanly (lint, test, build, security scan)
- No new `#nosec` directives without documented justification
- All changes follow existing Conventional Commits format
- No major version bump (all changes backwards-compatible or additive)

## Success Criteria

1. `go test -race ./...` passes with zero data races
2. Zero `context.TODO()` calls in production code (non-test files)
3. OCR checksum map contains entries for >= 20 languages
4. `--password flag` without `--allow-insecure-password` produces clear error
5. CI pipeline passes on all platforms
6. Test coverage >= 75%
7. SECURITY.md references v2.0.0
8. No items remaining in CONCERNS.md at P1 or P2 severity

## Constraints

- Must maintain Go 1.25+ compatibility
- Cross-platform support (linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64)
- No new external dependencies
- Existing CI/CD pipeline must continue to work
- No breaking changes to public CLI interface (v2.x compatibility)
