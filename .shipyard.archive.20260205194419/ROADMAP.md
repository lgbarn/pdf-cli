# pdf-cli Technical Debt Remediation -- Roadmap

## Milestone: Clean Baseline

Bring pdf-cli to a secure, reliable, and maintainable state by resolving all 20
identified technical concerns. The work is ordered to fail fast on high-risk
changes (dependency updates, concurrency fixes, security) before moving to
lower-risk cleanup. Every phase is independently verifiable and keeps CI green.

---

## Phase 1 -- Dependency Updates and Go Version Alignment

**Directory:** `phases/01-deps/`

**Description:**
Update all 21 outdated dependencies to latest compatible versions and align the
Go version string across go.mod, README, and CI config. This phase has no
functional code changes but establishes the foundation that every subsequent
phase builds on. Doing it first means all later work compiles against current
APIs and avoids rebasing pain.

**Requirements:** R7 (dependency updates), R9 (Go version consistency)

**Complexity:** S

**Dependencies:** none

**Risk:** Medium -- a transitive dependency bump could introduce breaking API
changes. Mitigated by running `go test -race ./...` and full CI after the
update.

**Success criteria:**
- `go mod tidy` produces no diff
- `go test -race ./...` passes
- Go version in go.mod, README, and `.github/workflows/ci.yaml` match
- CI pipeline (lint, test, build, security) passes on all platforms

**Key files:**
- `go.mod`, `go.sum`
- `README.md`
- `.github/workflows/ci.yaml`

---

## Phase 2 -- Thread Safety and Context Propagation

**Directory:** `phases/02-concurrency/`

**Description:**
Make global config and logging state thread-safe using `sync.Once`, and thread
`context.Context` through all long-running operations so they support
cancellation and timeouts. This is prerequisite work for the security and error
handling phases because those changes touch the same call paths.

**Requirements:** R4 (thread-safe globals), R5 (context propagation)

**Complexity:** M

**Dependencies:** Phase 1

**Risk:** High -- concurrency changes can introduce subtle bugs. Mitigated by
the `-race` detector and existing 81%+ coverage.

**Success criteria:**
- `config.Get()` and `logging.Get()` use `sync.Once` (no bare nil-check)
- `config.Reset()` and `logging.Reset()` are safe under concurrent access
- `ExtractTextFromPDF`, `downloadTessdata`, and batch processing functions
  accept `context.Context` as first parameter
- `go test -race ./...` passes with zero data races
- No change to public CLI behavior

**Key files:**
- `internal/config/config.go`
- `internal/logging/logger.go`
- `internal/ocr/ocr.go`
- `internal/commands/helpers.go`
- All command files that call long-running operations

---

## Phase 3 -- Security Hardening

**Directory:** `phases/03-security/`

**Description:**
Address all P0 critical security issues: remove password CLI flags in favor of
stdin/env/file-based input, add SHA256 checksum verification for downloaded
tessdata files, and apply consistent path traversal sanitization to all file
path inputs. This is the highest-priority functional change but depends on
context propagation being in place so the new password-reading code can respect
timeouts.

**Requirements:** R1 (password input), R2 (tessdata checksums), R3 (path sanitization)

**Complexity:** L

**Dependencies:** Phase 2

**Risk:** High -- R1 is a breaking CLI change. Mitigated by clear migration
notes and the project constraint that breaking changes are acceptable.

**Success criteria:**
- `--password` flag removed from encrypt/decrypt commands
- Passwords read from `PDF_CLI_PASSWORD` env var, `--password-file` flag, or
  interactive stdin prompt (via `golang.org/x/term`)
- `ps aux` during encrypt/decrypt shows no password in arguments
- `downloadTessdata` verifies SHA256 of downloaded file against a known-good
  checksum map before renaming into place
- A dedicated `SanitizePath()` function in `internal/fileio/` rejects paths
  containing `..` after cleaning, and all file-accepting entry points call it
- `gosec ./...` produces no new warnings

**Key files:**
- `internal/commands/encrypt.go`, `internal/commands/decrypt.go`
- `internal/cli/flags.go`
- `internal/ocr/ocr.go` (download + checksum)
- `internal/fileio/files.go` (path sanitization)

---

## Phase 4 -- Error Handling and Reliability

**Directory:** `phases/04-reliability/`

**Description:**
Fix silent error swallowing in parallel image processing, propagate file close
errors on write paths, and add signal-handler-based temp file cleanup. These
changes build on the context work from Phase 2 so that error propagation and
cleanup integrate with cancellation.

**Requirements:** R6 (parallel error surfacing), R8 (close error propagation),
R11 (temp file cleanup on crash)

**Complexity:** M

**Dependencies:** Phase 2

**Risk:** Medium -- changing error handling can surface previously hidden
failures that break tests. Mitigated by updating tests to expect the new
behavior.

**Success criteria:**
- `processImagesParallel` collects and returns all errors (not just first)
  using `errors.Join` or equivalent
- `processImagesSequential` does not silently discard errors from
  `ProcessImage`
- All `defer f.Close()` on write paths changed to check and return the error
- A `cleanup.Register(path)` / `cleanup.Run()` mechanism tracks temp dirs and
  removes them on SIGINT/SIGTERM
- `go test -race ./...` passes
- Coverage remains >= 81%

**Key files:**
- `internal/ocr/ocr.go` (parallel and sequential processing)
- `internal/fileio/files.go` (close error propagation)
- `internal/commands/patterns/stdio.go`
- New: `internal/cleanup/cleanup.go` (signal-based temp cleanup)

---

## Phase 5 -- Code Quality Improvements

**Directory:** `phases/05-quality/`

**Description:**
Replace magic numbers with named constants, consolidate any remaining non-slog
logging to slog, make the CI coverage check portable (replace awk/bc with a Go
test helper or `go tool cover` parsing), and make parallelism thresholds
configurable via config or adaptive to `runtime.NumCPU()`.

**Requirements:** R14 (named constants), R15 (slog consolidation), R16 (portable
coverage tooling), R17 (configurable parallelism)

**Complexity:** S

**Dependencies:** Phase 2 (logging changes depend on thread-safe logger)

**Risk:** Low -- these are mechanical refactors with no behavioral change.

**Success criteria:**
- No raw numeric literals used as thresholds or sizes (grep for common
  patterns)
- All logging goes through `internal/logging` (no direct `fmt.Fprintf` to
  stderr for log-style messages, no `log.` standard library usage)
- CI coverage step works without `bc` or `awk` (uses a Go-based check or
  `go tool cover` text parsing)
- `parallelThreshold` in `internal/ocr/ocr.go` reads from config with a
  sensible default; worker count adapts to `runtime.NumCPU()`
- `go test -race ./...` passes
- CI pipeline passes

**Key files:**
- `internal/ocr/ocr.go` (parallelThreshold, worker count)
- `internal/logging/logger.go`
- `internal/config/config.go` (new parallelism config fields)
- `.github/workflows/ci.yaml` (coverage step)
- Various command files (magic number and logging cleanup)

---

## Phase 6 -- Network Resilience and Retry Logic

**Directory:** `phases/06-retry/`

**Description:**
Add retry with exponential backoff to the tessdata download path. This is
isolated from other phases and can be done in parallel with Phase 5. The
implementation should use a small, generic retry helper (no new external deps)
that respects `context.Context` cancellation.

**Requirements:** R12 (network retry with backoff)

**Complexity:** S

**Dependencies:** Phase 2 (context propagation), Phase 3 (download function
changes)

**Risk:** Low -- additive change to an isolated code path.

**Success criteria:**
- `downloadTessdata` retries up to 3 times on transient HTTP errors (5xx,
  timeout, connection reset) with exponential backoff
- Non-retryable errors (4xx) fail immediately
- Retry respects `context.Context` cancellation
- Unit tests cover retry success on 2nd attempt, exhaustion, and
  non-retryable errors
- `go test -race ./...` passes

**Key files:**
- `internal/ocr/ocr.go`
- New: `internal/ocr/retry.go` (or generic `internal/retry/retry.go`)

---

## Phase 7 -- Documentation and Test Organization

**Directory:** `phases/07-docs/`

**Description:**
Split test files over 500 lines into focused files, and align all documentation
(README, architecture.md) with the current code after all prior changes. This
phase is intentionally last because earlier phases change code that
documentation describes.

**Requirements:** R10 (split large test files), R13 (documentation alignment)

**Complexity:** S

**Dependencies:** Phases 1-6 (documentation must reflect final state)

**Risk:** Low -- no functional changes.

**Success criteria:**
- No test file exceeds 500 lines (`commands_integration_test.go` at 882 lines
  and `additional_coverage_test.go` at 620 lines are split)
- Each split test file has a clear focus indicated by its filename
- README reflects current Go version, new password input method, and any other
  CLI changes from Phase 3
- `architecture.md` reflects any new packages (cleanup, retry) and updated
  function signatures
- `go test ./...` passes (test behavior unchanged)
- Coverage remains >= 81%

**Key files:**
- `internal/commands/commands_integration_test.go` (882 lines -- split)
- `internal/commands/additional_coverage_test.go` (620 lines -- split)
- `README.md`
- `docs/architecture.md`

---

## Phase Dependency Graph

```
Phase 1 (deps, go version)
   |
   v
Phase 2 (thread safety, context)
   |
   +--------+--------+
   |        |        |
   v        v        |
Phase 3  Phase 4     |
(security) (errors)  |
   |        |        |
   +--------+        |
   |                 |
   v                 v
Phase 6           Phase 5
(retry)           (quality)
   |                 |
   +--------+--------+
            |
            v
         Phase 7
      (docs, tests)
```

**Wave assignment for maximum parallelism:**

| Wave | Phases | Description |
|------|--------|-------------|
| 1 | Phase 1 | Foundation: deps and version alignment |
| 2 | Phase 2 | Foundation: concurrency primitives |
| 3 | Phase 3, Phase 4 | Security + error handling (parallel) |
| 4 | Phase 5, Phase 6 | Quality + retry (parallel) |
| 5 | Phase 7 | Final: docs and test organization |

---

## Estimated Effort

| Phase | Complexity | Estimated Plans |
|-------|-----------|-----------------|
| 1 -- Deps | S | 1 |
| 2 -- Concurrency | M | 2 |
| 3 -- Security | L | 3 |
| 4 -- Reliability | M | 2 |
| 5 -- Quality | S | 2 |
| 6 -- Retry | S | 1 |
| 7 -- Docs | S | 1 |
| **Total** | | **12 plans** |
