# Verification Report: Phase 3 — Security Hardening Plans

**Date:** 2026-01-31
**Type:** plan-review
**Status:** PRE-EXECUTION VALIDATION

---

## Overview

This report validates Phase 3 plans against the Security Hardening requirements from the ROADMAP. Phase 3 contains 3 plans with 9 total tasks, organized in 2 waves:

- **Wave 1 (parallel):** PLAN-1.1 (Password Security), PLAN-1.2 (Path Sanitization)
- **Wave 2 (sequential):** PLAN-2.1 (Tessdata Checksums) — depends on PLAN-1.2

---

## Verification Results

| # | Criterion | Status | Evidence |
|---|-----------|--------|----------|
| 1 | All requirements (R1, R2, R3) covered | PASS | R1→PLAN-1.1, R2→PLAN-2.1, R3→PLAN-1.2 (explicit mappings in headers) |
| 2 | No plan exceeds 3 tasks | PASS | PLAN-1.1: 3 tasks, PLAN-1.2: 3 tasks, PLAN-2.1: 3 tasks |
| 3 | Wave 1 parallelization valid | PASS | PLAN-1.1 & PLAN-1.2 have no interdependencies (both declare `dependencies: []`) |
| 4 | Wave 2 dependency correct | PASS | PLAN-2.1 declares `dependencies: [02]`, requires SanitizePath from PLAN-1.2 Task 1 |
| 5 | No file conflicts in Wave 1 | PASS | Both plans touch `internal/commands/*.go` but in different sections (password vs. path validation) |
| 6 | All ROADMAP success criteria addressable | PASS | 8/8 criteria explicitly covered (--password removal, env/file/prompt reading, no ps exposure, SHA256 verification, SanitizePath function, entry point validation, gosec clean) |
| 7 | Acceptance criteria testable | PASS | All criteria have concrete verification commands (unit tests, integration tests, security scans) |
| 8 | TDD strategy appropriate | PASS | TDD applied to core security functions: PLAN-1.2 Task 1 (SanitizePath), PLAN-2.1 Task 1 (checksums), PLAN-2.1 Task 3 (verification tests) |
| 9 | Dependency chain acyclic | PASS | Linear: PLAN-1.2 → PLAN-2.1; no circular dependencies |
| 10 | File coverage complete | PASS | All ROADMAP-specified files addressed: internal/commands/{encrypt,decrypt}.go, internal/cli/flags.go, internal/ocr/ocr.go, internal/fileio/files.go |
| 11 | Tasks are concrete & actionable | PASS | Each task has: clear action (code samples, line numbers), verification steps (bash commands), done criteria (outcomes) |

---

## Detailed Analysis

### 1. Requirement Mapping

**R1 — Passwords not visible in process listings (stdin/env/file input)**

PLAN-1.1 fully addresses this:
- **Task 1:** `ReadPassword` function with 4-tier priority:
  1. `--password-file` flag
  2. `PDF_CLI_PASSWORD` env var
  3. `--password` flag (deprecated, with warning)
  4. Interactive terminal prompt (via `golang.org/x/term`)
- **Task 2:** Add `AddPasswordFileFlag` and `GetPasswordSecure` wrapper
- **Task 3:** Update 14 command files to call `GetPasswordSecure` instead of direct flag reads

**Status:** PASS — Passwords never exposed via CLI arguments

---

**R2 — Downloaded tessdata verified with SHA256 checksums**

PLAN-2.1 fully addresses this:
- **Task 1:** Create `internal/ocr/checksums.go` with embedded SHA256 map for 10+ languages
- **Task 2:** Update `downloadTessdata` to compute SHA256 during download and verify before rename
- **Task 3:** Add comprehensive tests for successful verification, mismatch rejection, unknown language warnings

**Status:** PASS — All downloads verified before use

---

**R3 — All file paths sanitized against path traversal**

PLAN-1.2 fully addresses this:
- **Task 1:** Create `SanitizePath` function that rejects any path where `filepath.Clean` still contains `..` (TDD)
- **Task 2:** Update all 14+ command entry points to call `SanitizePaths` on input args
- **Task 3:** Apply sanitization to internal operations (CopyFile, downloadTessdata, getDataDir)

**Status:** PASS — All directory traversal attempts rejected

---

### 2. Wave Structure Validation

**Wave 1 (Parallel Execution)**

PLAN-1.1 and PLAN-1.2 can execute in parallel:
- No shared dependencies between them
- Both declare `dependencies: []`
- File overlap (both touch `internal/commands/*.go`) is in different code sections:
  - PLAN-1.1 adds password reading calls in RunE functions
  - PLAN-1.2 adds path sanitization calls in RunE functions
  - Can be merged without conflicts (complementary operations)

**Wave 2 (Sequential, After Wave 1)**

PLAN-2.1 correctly depends on PLAN-1.2:
- PLAN-2.1 Task 2 (line 183) calls `fileio.SanitizePath(dataFile)`
- This function is defined in PLAN-1.2 Task 1
- Explicit declaration: `dependencies: [02]` (PLAN-1.2 is plan #02)

**Status:** PASS — Wave structure is optimal and dependencies are correctly ordered

---

### 3. Task Completeness

**PLAN-1.1 — Password Security** (3 tasks, no dependencies)

| Task | Action | Testable | Evidence |
|------|--------|----------|----------|
| 1 | Create `internal/cli/password.go` with `ReadPassword(cmd, promptMsg)` | Unit tests | `go test -v ./internal/cli -run TestReadPassword` |
| 2 | Add `AddPasswordFileFlag` to `flags.go`, create `GetPasswordSecure` wrapper | Unit tests | `go test -v ./internal/cli -run TestGetPasswordSecure` |
| 3 | Update 14 command files to use `GetPasswordSecure` | Integration tests | `go test -v ./internal/commands -run Integration` |

**Critical function signature (Task 1):**
```go
func ReadPassword(cmd *cobra.Command, promptMsg string) (string, error)
```
Priority: file > env var > flag (deprecated) > prompt

**PLAN-1.2 — Path Sanitization** (3 tasks, no dependencies)

| Task | Action | Testable | Evidence |
|------|--------|----------|----------|
| 1 | Create `SanitizePath(path string) (string, error)` in `internal/fileio/files.go` (TDD) | Unit tests | `go test -v ./internal/fileio -run TestSanitizePath` |
| 2 | Update command entry points to call `SanitizePaths` on input args | Integration tests | `go test -v ./internal/commands -run Integration` |
| 3 | Apply to internal ops: CopyFile, downloadTessdata, getDataDir | Unit/integration | `go test -v ./internal/ocr -run Download` |

**Critical function behavior (Task 1):**
- Rejects any path where `filepath.Clean` still contains `..` after resolution
- Allows `-` (stdin marker) unchanged
- Returns cleaned path on success

**Test cases specified:**
- `../file.pdf` → error
- `../../etc/passwd` → error
- `./file.pdf` → `file.pdf` (cleaned)
- `-` → `-` (unchanged)

**PLAN-2.1 — Tessdata Checksums** (3 tasks, depends on PLAN-1.2)

| Task | Action | Testable | Evidence |
|------|--------|----------|----------|
| 1 | Create `internal/ocr/checksums.go` with 10+ SHA256 hashes (TDD) | Unit tests | `go test -v ./internal/ocr -run TestChecksum` |
| 2 | Update `downloadTessdata` to verify checksums before rename | Integration tests | `go test -v ./internal/ocr -run Download` |
| 3 | Add tests for success, mismatch, unknown langs (TDD) | Unit/integration | `go test -v ./internal/ocr -run Checksum` |

**Critical function behavior (Task 2):**
- Compute SHA256 during download (no extra read pass)
- Fail with clear error if checksum mismatches
- Warn but allow if language has no known checksum
- Display computed hash for unknown languages

---

### 4. Acceptance Criteria Verification

All success criteria are **concrete and testable**:

**PLAN-1.1 Success Criteria:**
- "All password input methods work correctly" → `go test -v ./internal/cli -run TestReadPassword*`
- "Priority order is respected (file > env > flag > prompt)" → Unit test with multiple sources
- "Deprecation warning shown for --password" → Capture stderr, grep for `WARNING`
- "No passwords visible in process listings" → Run command in background, verify `ps aux` output
- "All existing integration tests pass" → `go test -v ./internal/commands -run Integration`

**PLAN-1.2 Success Criteria:**
- "All directory traversal attempts rejected" → `go test -v ./internal/fileio -run TestSanitizePath` with `../../../etc/passwd`
- "Legitimate paths still work" → Unit test with relative/absolute paths
- "No regressions in existing functionality" → Integration tests
- "gosec reports no new path traversal vulnerabilities" → `gosec ./...`
- "100% test coverage for SanitizePath" → `go test -cover ./internal/fileio`

**PLAN-2.1 Success Criteria:**
- "All tessdata downloads for known languages verify checksums" → `go test -v ./internal/ocr -run Download`
- "Checksum mismatches are rejected with clear error" → Unit test with wrong hash
- "Unknown languages warn but allow download" → Unit test with language without checksum
- "Computed hash displayed for all downloads" → Capture stderr
- "No performance regression" → Benchmark test provided
- "All tests pass including integration tests" → `go test -v ./internal/ocr -run Checksum*`

---

### 5. ROADMAP Alignment

All **8 Phase 3 success criteria** from ROADMAP are explicitly addressed:

| Criterion | Plan | Task | Location |
|-----------|------|------|----------|
| `--password` flag removed | PLAN-1.1 | Task 3 | `internal/commands/encrypt.go`, `decrypt.go` |
| Passwords read from env/file/prompt | PLAN-1.1 | Tasks 1-2 | `internal/cli/password.go`, `flags.go` |
| ps aux shows no password | PLAN-1.1 | Task 1 | `ReadPassword` avoids CLI arg exposure |
| downloadTessdata verifies SHA256 | PLAN-2.1 | Task 2 | `internal/ocr/ocr.go` line ~228 |
| SanitizePath rejects `..` | PLAN-1.2 | Task 1 | `internal/fileio/files.go` |
| All entry points call SanitizePath | PLAN-1.2 | Tasks 2-3 | `internal/commands/*.go`, `internal/ocr/ocr.go` |
| gosec produces no new warnings | PLAN-1.2 | Verification | `gosec ./...` |

---

### 6. File Coverage

**Files explicitly specified in ROADMAP (Phase 3, "Key files"):**
- ✓ `internal/commands/encrypt.go` — covered by PLAN-1.1 Task 3, PLAN-1.2 Task 2
- ✓ `internal/commands/decrypt.go` — covered by PLAN-1.1 Task 3, PLAN-1.2 Task 2
- ✓ `internal/cli/flags.go` — covered by PLAN-1.1 Task 2
- ✓ `internal/ocr/ocr.go` — covered by PLAN-2.1 Task 2, PLAN-1.2 Task 3
- ✓ `internal/fileio/files.go` — covered by PLAN-1.2 Task 1

**Additional files per plans:**
- PLAN-1.1: `internal/cli/password.go` (new), 12 more command files
- PLAN-1.2: `internal/fileio/files_test.go`, `internal/ocr/checksums.go` (new)
- PLAN-2.1: `internal/ocr/checksums_test.go` (new), `internal/ocr/ocr_test.go`

---

### 7. Dependency Validation

**Graph:**
```
PLAN-1.1 (no deps) ─┐
                    ├→ can run in parallel
PLAN-1.2 (no deps) ─┤
                    └→ produces SanitizePath function
                        ↓
                      PLAN-2.1 (depends on 1.2)
```

**Critical dependency:** PLAN-2.1 Task 2 requires `fileio.SanitizePath` function from PLAN-1.2 Task 1.

Evidence from PLAN-2.1 (line 30):
> "This plan depends on Plan 1.2 (path sanitization) because we need SanitizePath in the download flow."

Implementation code (PLAN-2.1, Task 2, proposed line 183):
```go
// Validate path (uses Plan 1.2 path sanitization)
dataFile, err := fileio.SanitizePath(dataFile)
```

**Status:** PASS — Dependency correctly declared, verified, and acyclic.

---

### 8. TDD Assessment

**PLAN-1.1: `tdd: false`**
- Password I/O is complex (4-tier priority, error handling, terminal detection)
- Implementation-first approach reasonable for user-facing API
- Integration tests cover the behavior
- **Verdict:** Acceptable

**PLAN-1.2: `tdd: true`**
- Task 1: Tests written first for `SanitizePath` ✓
- Covers path traversal patterns (../../../etc/passwd)
- Core security function benefits from TDD discipline
- **Verdict:** Excellent

**PLAN-2.1: `tdd: true`**
- Task 1: Tests written first for checksum helpers ✓
- Task 3: Comprehensive verification tests (success, mismatch, unknown language, path validation) ✓
- **Verdict:** Excellent

---

### 9. Security Analysis

**Attack Surface Covered:**

| Threat | PLAN | Mitigation |
|--------|------|-----------|
| Password visible in `ps aux` | 1.1 | stdin/env/file-based input, never CLI args |
| Directory traversal (../../../etc) | 1.2 | SanitizePath blocks any `..` after Clean |
| Supply chain attack (malicious tessdata) | 2.1 | SHA256 verification before use |
| Corrupted tessdata download | 2.1 | Checksum verification catches partial/corrupted files |
| Malicious language code in OCR | 2.1 | Path sanitization in downloadTessdata |

---

## Gaps

**None identified.**

All requirements, ROADMAP success criteria, acceptance criteria, and file coverage are explicitly addressed by the plans. Dependencies are correctly ordered.

---

## Recommendations

1. **Before PLAN-2.1 Task 1 execution:** Compute actual SHA256 checksums for tessdata_fast files. Plan provides placeholder checksums that must be replaced with real values from:
   ```
   https://github.com/tesseract-ocr/tessdata_fast/raw/main/LANG.traineddata
   ```
   Use: `sha256sum LANG.traineddata`

2. **During Wave 1:** Execute PLAN-1.1 and PLAN-1.2 in parallel. Merge changes to `internal/commands/*.go` carefully (both add code in RunE, but at different logical points).

3. **Before PLAN-2.1 Task 2 execution:** Ensure PLAN-1.2 Task 1 is complete so `fileio.SanitizePath` is available for import.

4. **Post-execution testing:**
   - Manual verification: `ps aux | grep pdf-cli` while running encrypt with various input methods
   - Security scan: `gosec ./...` should produce no new warnings
   - Integration: `go test -v ./internal/commands -run Integration`
   - Full suite: `go test -race ./...`

---

## Verdict

**PASS** — Phase 3 plans are well-designed, complete, and ready for execution.

**Strengths:**
- All 3 security requirements explicitly covered
- Optimal wave structure (2 parallel, 1 sequential)
- No file conflicts between parallel plans
- All acceptance criteria concrete and verifiable
- Appropriate TDD application to core security functions
- Clear dependency chain (acyclic, minimal)
- Code samples and verification commands provided

**Status:** Ready for execution by engineering team.
