# pdf-cli Technical Debt Remediation

## Description

Systematically address all 20 identified technical concerns in the pdf-cli codebase, prioritized by severity (P0 critical through P3 minor). The project covers security vulnerabilities, reliability issues, code quality improvements, and dependency updates. Breaking changes are acceptable where they improve security or correctness.

pdf-cli is a Go CLI tool for PDF manipulation (merge, split, encrypt, OCR, watermark, metadata, validation) built with Cobra, pdfcpu, and a dual OCR backend (native Tesseract + WASM fallback). The codebase is ~87 Go files with 81.5% test coverage.

## Goals

1. Fix all P0 critical security issues (password exposure in CLI flags, no download integrity verification, inconsistent path sanitization)
2. Fix all P1 high-priority reliability issues (race conditions in global state, missing context propagation, silent errors in parallel processing, outdated dependencies, ignored close errors)
3. Address P2 medium-priority technical debt (Go version documentation mismatches, large test files, orphaned temp files, missing retry logic, documentation drift)
4. Clean up P3 code quality issues (magic numbers as named constants, inconsistent logging, coverage tooling portability, adaptive parallelism thresholds)
5. Update all 21 outdated dependencies to latest compatible versions

## Non-Goals

- Adding new PDF features or CLI commands
- Changing the overall architecture or package structure
- Achieving test coverage beyond current 81% level
- Migrating away from core dependencies (pdfcpu, Cobra, gogosseract)
- Performance optimization beyond fixing existing issues
- UI/UX changes to the CLI interface (beyond security-driven flag changes)

## Requirements

### Security (P0)
- R1: Passwords must not be visible in process listings or shell history — use stdin, env vars, or file-based input instead of CLI flags
- R2: Downloaded tessdata files must be verified with SHA256 checksums
- R3: All file path inputs must be sanitized against path traversal attacks consistently

### Reliability (P1)
- R4: Global config and logging state must be thread-safe (mutex or sync.Once)
- R5: All long-running operations must accept and propagate context.Context for cancellation
- R6: Parallel processing must surface all errors, not silently drop failures
- R7: All 21 outdated dependencies updated to latest compatible versions
- R8: File close errors must be checked and propagated, especially for write operations

### Technical Debt (P2)
- R9: Go version requirements consistent across go.mod, README, and CI config
- R10: Test files over 500 lines should be split into focused files
- R11: Temp file cleanup on crash/interrupt via signal handlers
- R12: Network operations (tessdata download) should have retry logic with backoff
- R13: Documentation (README, architecture.md) aligned with current code

### Code Quality (P3)
- R14: Magic numbers replaced with named constants
- R15: Logging consolidated to a single approach (slog)
- R16: Coverage tooling made portable (no bc/awk dependency)
- R17: Parallelism thresholds made configurable or adaptive to system resources

## Non-Functional Requirements

- All changes must pass `go test -race ./...`
- Test coverage must remain >= 81%
- CI pipeline must pass cleanly (lint, test, build, security scan)
- No new `#nosec` directives added without documented justification
- All changes follow existing Conventional Commits format

## Success Criteria

1. `gosec` runs clean — all warnings resolved without `#nosec` where possible
2. `go test -race ./...` passes with zero race conditions
3. `go mod tidy` shows no outdated dependencies
4. CI pipeline (lint, test, build, security) passes on all platforms
5. Test coverage >= 81% as reported by Codecov
6. No passwords visible in `ps aux` output during encrypt/decrypt operations
7. All file downloads verified with checksums

## Constraints

- Must maintain Go 1.24.1+ compatibility
- Cross-platform support (linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64)
- No new external dependencies unless strictly necessary
- Existing CI/CD pipeline (GitHub Actions + GoReleaser) must continue to work
