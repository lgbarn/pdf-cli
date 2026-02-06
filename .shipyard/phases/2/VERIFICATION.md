# Verification Report
**Phase:** Phase 2: Security Hardening
**Date:** 2026-02-05
**Type:** plan-review

## Results

| # | Criterion | Status | Evidence |
|---|-----------|--------|----------|
| 1 | Coverage: All Phase 2 requirements (R1, R2, R3) covered | PASS | PLAN-1.1 covers R1 (OCR checksums). PLAN-1.2 covers R3 (directory permissions). PLAN-2.1 covers R2 (password flag). All three requirements mapped to exactly one plan each. |
| 2 | Task limit: Each plan has AT MOST 3 tasks | **FAIL** | PLAN-1.1: 3 tasks (PASS). PLAN-1.2: 3 tasks (PASS). PLAN-2.1: **5 tasks (FAIL)** -- exceeds maximum of 3 tasks. CRITICAL VIOLATION. |
| 3 | Correctness: Approaches are technically sound | PASS | PLAN-1.1: Downloading tessdata files and computing checksums, then adding to map is standard practice. PLAN-1.2: Changing permission constants is straightforward. PLAN-2.1: Adding opt-in flag and modifying password logic follows established flag patterns in the codebase. All approaches are valid. |
| 4a | File accuracy: PLAN-1.1 references | PASS | `/Users/lgbarn/Personal/pdf-cli/internal/ocr/checksums.go` line 9-11 matches actual file (KnownChecksums map at lines 9-11). `/Users/lgbarn/Personal/pdf-cli/internal/ocr/checksums_test.go` lines 25-38 not verified but test name TestAllChecksumsValidFormat is plausible. File references are accurate. |
| 4b | File accuracy: PLAN-1.2 references | PASS | `/Users/lgbarn/Personal/pdf-cli/internal/fileio/files.go` line 15 matches actual file (DefaultDirPerm at line 15). `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr.go` line 45 matches actual file (DefaultDataDirPerm at line 45). All file references accurate. |
| 4c | File accuracy: PLAN-2.1 references | PASS | `/Users/lgbarn/Personal/pdf-cli/internal/cli/flags.go` line 44 is approximately correct (AddPasswordFileFlag ends at 44). `/Users/lgbarn/Personal/pdf-cli/internal/cli/password.go` lines 47-54 matches actual file (password flag check at lines 47-54). `/Users/lgbarn/Personal/pdf-cli/internal/commands/encrypt.go` shows pattern at lines 18-19 (AddPasswordFlag and AddPasswordFileFlag). All 14 command files verified to have AddPasswordFlag. File references accurate. |
| 5 | Wave ordering: Plans in same wave don't conflict | PASS | PLAN-1.1 (Wave 1) touches `internal/ocr/checksums.go`. PLAN-1.2 (Wave 1) touches `internal/fileio/files.go` and `internal/ocr/ocr.go` (permission constants only). PLAN-2.1 (Wave 2) touches `internal/cli/flags.go`, `internal/cli/password.go`, `internal/commands/*.go`. No file conflicts between Wave 1 plans. Wave 2 depends on Wave 1 completing first. Ordering is correct. |
| 6 | Verification commands: Are they concrete and runnable? | PASS | PLAN-1.1: `go test -race ./internal/ocr/...` is concrete and runnable. `grep -c` command is concrete. PLAN-1.2: `grep -rn '0750'` is concrete and runnable. `go test -race` commands are concrete. PLAN-2.1: `go test -race -v -run TestReadPassword` is concrete. `go run ... | grep -c` commands are concrete. All verification commands are executable. |
| 7 | Success criteria: Are they measurable and objective? | PASS | PLAN-1.1: "21 entries" is measurable. "grep -c returns >= 20" is measurable. "tests pass" is binary. PLAN-1.2: "zero results" is measurable. "both set to 0700" is measurable. PLAN-2.1: "error message mentions X" is measurable via grep. "tests pass" is binary. All criteria are objective and verifiable. |
| 8 | Dependency ordering: Are plan dependencies correct? | PASS | PLAN-1.1: No dependencies (Wave 1). PLAN-1.2: No dependencies (Wave 1). PLAN-2.1: Depends on Wave 1 completing to avoid merge conflicts in `internal/ocr/ocr.go` (both PLAN-1.2 and PLAN-2.1 touch this file, but PLAN-1.2 only changes constants while PLAN-2.1 doesn't touch this file -- actually, PLAN-2.1 doesn't touch ocr.go). No circular dependencies. Ordering is safe. |
| 9 | File conflicts: Do multiple plans touch the same files in conflicting ways? | PASS | PLAN-1.1 touches `internal/ocr/checksums.go` only. PLAN-1.2 touches `internal/fileio/files.go` and `internal/ocr/ocr.go` (permission constants). PLAN-2.1 touches `internal/cli/*.go` and `internal/commands/*.go`. PLAN-1.2 and PLAN-2.1 both run in sequence, avoiding conflicts. No conflicting writes. |
| 10 | Roadmap alignment: Phase 2 success criteria can be verified | PASS | Criterion 1: `go test -race ./internal/ocr/... ./internal/cli/...` -- verified runnable. Criterion 2: `grep -c 'eng\|fra\|deu\|spa' internal/ocr/checksums.go` shows >= 20 -- concrete and measurable. Criterion 3: Running command without opt-in flag produces error -- verifiable via manual test or integration test. Criterion 4: `grep -rn '0750' internal/ --include='*.go'` returns zero -- concrete and measurable. Criterion 5: Test coverage >= 75% -- verifiable via `go test -cover`. All roadmap criteria are verifiable. |

## Gaps

### CRITICAL: PLAN-2.1 Task Count Violation
- **Issue**: PLAN-2.1 has 5 tasks, exceeding the maximum of 3 tasks per plan.
- **Impact**: Violates planning constraints. Plan is too complex and should be split or consolidated.
- **Tasks in PLAN-2.1**:
  1. Add --allow-insecure-password flag infrastructure
  2. Update password reading logic with opt-in check
  3. Add --allow-insecure-password flag to all 14 password commands
  4. Update password_test.go for new behavior
  5. Verify implementation across all commands
- **Root Cause**: Task 3 involves modifying 14 separate command files, which could be considered a single task ("Update all command files") rather than 14 micro-tasks. Task 5 is a verification task that could be merged with Task 4.
- **Recommendation**: Consolidate tasks to meet 3-task limit:
  - **Option A (Recommended)**: Merge Task 4 and Task 5 into a single "Update tests and verify" task. This reduces the plan to 4 tasks, which is still over the limit.
  - **Option B**: Split PLAN-2.1 into two plans:
    - PLAN-2.1a (Wave 2): Tasks 1-3 (add flag infrastructure, update password logic, update command files)
    - PLAN-2.1b (Wave 2, depends on 2.1a): Tasks 4-5 (update tests and verify)
  - **Option C (Most Conservative)**: Consolidate to 3 tasks:
    - Task 1: Add flag infrastructure and update password logic (merge current Tasks 1 and 2)
    - Task 2: Update all command files and test files (merge current Tasks 3 and 4)
    - Task 3: Verify implementation (current Task 5)

### Minor: Test File References Not Fully Verified
- **Issue**: PLAN-1.1 references `checksums_test.go` lines 25-38 for `TestAllChecksumsValidFormat`, but this was not fully verified by reading the test file.
- **Impact**: Low. Test name is plausible and the verification command in Task 3 will catch any issues.
- **Recommendation**: No action required. Verification step will catch any discrepancies.

### Minor: PLAN-2.1 Wave Dependency Justification Unclear
- **Issue**: PLAN-2.1 states it "must execute after Wave 1" to "reduce merge conflicts" because "both touch internal/ocr/ocr.go". However, PLAN-2.1 does not actually touch `internal/ocr/ocr.go` based on the task descriptions.
- **Impact**: Low. The wave ordering is still safe (running in Wave 2 after Wave 1 is not harmful), but the justification is inaccurate.
- **Recommendation**: Clarify dependency justification. The real reason to sequence PLAN-2.1 after Wave 1 might be to keep Phase 2 focused and avoid running all 3 plans in parallel, not file conflicts.

## Recommendations

### 1. CRITICAL: Revise PLAN-2.1 to Meet 3-Task Limit
**Action Required**: Choose one of the consolidation options above and revise PLAN-2.1 before execution.

**Recommended Approach (Option C)**: Consolidate to 3 tasks as follows:

**Task 1: Add flag infrastructure and password validation**
- Add `AddAllowInsecurePasswordFlag` and `GetAllowInsecurePassword` to `flags.go`
- Update `ReadPassword` in `password.go` to check for opt-in flag
- Files: `internal/cli/flags.go`, `internal/cli/password.go`
- Acceptance: Flag functions added, password logic updated, code compiles

**Task 2: Update all commands and tests**
- Add `cli.AddAllowInsecurePasswordFlag(cmd)` to all 14 command files
- Update `newTestCmd()` helper in `password_test.go`
- Update existing tests to set opt-in flag where needed
- Add two new tests: `TestReadPassword_PasswordFlagWithoutOptIn` and `TestReadPassword_PasswordFlagWithOptIn`
- Files: 14 command files + `internal/cli/password_test.go`
- Acceptance: All commands have flag, all tests updated and pass

**Task 3: Verify implementation**
- Run full test suite with race detection
- Verify error message behavior via manual command test
- Verify test coverage >= 75%
- Files: All test files
- Acceptance: All tests pass, error message verified, coverage maintained

This consolidation reduces PLAN-2.1 from 5 tasks to 3 tasks while preserving all the work and acceptance criteria.

### 2. OPTIONAL: Clarify PLAN-2.1 Wave Dependency
**Action**: Update the "Dependencies" section of PLAN-2.1 to clarify why it runs in Wave 2 after Wave 1. If the real reason is to keep Phase 2 focused and avoid parallel overhead (not file conflicts), state that explicitly.

### 3. OPTIONAL: Verify Test File Line Numbers
**Action**: Before executing PLAN-1.1, read `checksums_test.go` to verify line numbers for `TestAllChecksumsValidFormat`. This is a nice-to-have for accuracy but not critical since the verification commands will catch any issues.

## Verdict
**NEEDS_REVISION** -- PLAN-2.1 exceeds the 3-task limit with 5 tasks. This is a CRITICAL violation of planning constraints. The plan must be revised to consolidate tasks to 3 or fewer before execution. All other checks pass.

**Summary**:
- Coverage: PASS (all 3 requirements covered)
- Task limit: FAIL (PLAN-2.1 has 5 tasks, max is 3)
- Correctness: PASS (all approaches are sound)
- File accuracy: PASS (all references verified)
- Wave ordering: PASS (no conflicts)
- Verification commands: PASS (all concrete and runnable)
- Success criteria: PASS (all measurable)
- Dependencies: PASS (no circular dependencies)
- File conflicts: PASS (no conflicting writes)
- Roadmap alignment: PASS (all criteria verifiable)

**Action Required**: Revise PLAN-2.1 to consolidate from 5 tasks to 3 tasks using Option C recommended above. After revision, all plans will be APPROVED for execution.
