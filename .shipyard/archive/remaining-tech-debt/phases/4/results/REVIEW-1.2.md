# REVIEW-1.2: Test Helpers and Default Log Level

**Reviewer:** Claude Code (Senior Code Reviewer)
**Date:** 2026-02-05
**Plan:** PLAN-1.2 (Phase 4 - R11, R14)
**Status:** REQUEST CHANGES

---

## Stage 1: Spec Compliance

**Verdict:** FAIL

### Critical Issue: Commit Bundling Violation

The implementation violates commit separation principles. Commit `bc85124` was intended for PLAN-1.2 (test helper refactoring) but also contains changes from PLAN-1.1 (suffix constants in helpers.go). This creates ambiguity in the git history and makes it difficult to track which changes belong to which plan.

**Evidence:**
```bash
$ git show bc85124 --stat
commit bc85124f9142bb54bd0315aa126dcccd98249884
    shipyard(phase-4): refactor test helpers to use testing.TB instead of panic

 internal/commands/helpers.go | 10 ++++++++++  # PLAN-1.1 changes
 internal/testing/fixtures.go | 11 ++++++-----  # PLAN-1.2 changes
```

The commit message "refactor test helpers to use testing.TB instead of panic" correctly describes the fixtures.go changes but does not mention the helpers.go changes (suffix constants) at all.

**Root Cause (per SUMMARY-1.1):** The pre-commit hook automatically modified fixtures.go while working on PLAN-1.1. Both changes were staged together, creating a mixed commit.

**Impact:** This violates the principle that each commit should represent a single logical change tied to a specific plan. It makes code archaeology difficult and conflicts with the stated commits in SUMMARY-1.2.

---

### Task 1: Refactor TempDir() and TempFile() to use testing.TB

**Status:** PASS (implementation correct, commit bundling issue noted above)

**Evidence:**
- `/Users/lgbarn/Personal/pdf-cli/internal/testing/fixtures.go` line 3-8: `testing` package imported
- Line 34: Function signature changed to `func TempDir(t testing.TB, prefix string)`
- Line 44: Function signature changed to `func TempFile(t testing.TB, prefix, content string)`
- Line 37: `panic("failed to create temp dir: ...")` replaced with `t.Fatal("failed to create temp dir: ...")`
- Line 47: `panic("failed to create temp file: ...")` replaced with `t.Fatal("failed to create temp file: ...")`
- Line 53: `panic("failed to write temp file: ...")` replaced with `t.Fatal("failed to write temp file: ...")`
- Line 15: TestdataDir() panic unchanged as specified

**Verification:**
```bash
$ grep -n 'panic(' internal/testing/fixtures.go
15:		panic("failed to get caller information")
```
Only TestdataDir contains panic, as required.

**Caller verification:**
Grep search confirmed zero callers of `TempDir(` or `TempFile(` in actual Go test files. All instances found were:
- `t.TempDir()` (standard library method, not the custom helper)
- Documentation/plan references

**Notes:** The implementation is technically correct. All 3 panic() calls in TempDir and TempFile were replaced with t.Fatal(). Function signatures correctly accept testing.TB as the first parameter. The zero-caller assumption is verified, so there are no breaking changes to existing tests.

---

### Task 2: Change default CLI log level from "silent" to "error"

**Status:** PASS

**Evidence:**
- `/Users/lgbarn/Personal/pdf-cli/internal/cli/flags.go` line 104:
  ```go
  cmd.PersistentFlags().StringVar(&logLevel, "log-level", "error", "Log level (debug, info, warn, error, silent)")
  ```
  The default value is "error" (5th parameter in StringVar call).

**Verification:**
```bash
$ grep 'PersistentFlags.*StringVar.*logLevel.*"error"' internal/cli/flags.go
cmd.PersistentFlags().StringVar(&logLevel, "log-level", "error", "Log level (debug, info, warn, error, silent)")
```

**Notes:** The change is minimal and correct. Users will now see error-level messages by default unless they explicitly set `--log-level=silent`. The logging package's internal default (LevelSilent) remains unchanged, as documented in the plan.

---

### Task 3: Full test suite verification

**Status:** PASS

**Evidence:**
```bash
$ go build ./cmd/pdf
# Build succeeds with no output

$ go test ./...
ok  	github.com/lgbarn/pdf-cli/internal/cleanup	(cached)
ok  	github.com/lgbarn/pdf-cli/internal/cli	(cached)
ok  	github.com/lgbarn/pdf-cli/internal/commands	1.141s
ok  	github.com/lgbarn/pdf-cli/internal/config	(cached)
ok  	github.com/lgbarn/pdf-cli/internal/fileio	(cached)
ok  	github.com/lgbarn/pdf-cli/internal/logging	(cached)
ok  	github.com/lgbarn/pdf-cli/internal/ocr	5.592s
ok  	github.com/lgbarn/pdf-cli/internal/output	(cached)
ok  	github.com/lgbarn/pdf-cli/internal/pages	(cached)
ok  	github.com/lgbarn/pdf-cli/internal/pdf	(cached)
ok  	github.com/lgbarn/pdf-cli/internal/pdferrors	(cached)
ok  	github.com/lgbarn/pdf-cli/internal/progress	(cached)
ok  	github.com/lgbarn/pdf-cli/internal/retry	(cached)
```

All packages pass. No test failures. No regressions from either change.

**Notes:** The test helper changes have no impact (zero callers). The logging package tests verify package-level defaults (LevelSilent) which were not changed. The CLI flag tests accept the new "error" default without issues.

---

## Stage 1 Summary

**Verdict:** FAIL

**Blocking Issue:** Commit bundling violation. Commit `bc85124` mixes PLAN-1.1 and PLAN-1.2 changes, creating unclear git history and violating the one-plan-per-commit principle.

The actual code implementation is correct and all acceptance criteria are met:
- ✓ TempDir and TempFile accept testing.TB parameter
- ✓ 3 panic() calls replaced with t.Fatal()
- ✓ TestdataDir unchanged
- ✓ Default CLI log level is "error"
- ✓ All tests pass
- ✓ Build succeeds

However, the commit organization fails to meet the separation of concerns required by the two-plan structure.

---

## Stage 2: Code Quality

**Stage 2 skipped** due to Stage 1 failure. Will perform Stage 2 review once commit organization issue is resolved.

---

## Integration Review: Conflicts with PLAN-1.1

**Conflict detected:** PLAN-1.1 (commits `bc85124`, `aedda33`, `cc46f8c`) and PLAN-1.2 (commits `bc85124`, `945e9dc`) share commit `bc85124`.

**PLAN-1.1 files touched:**
- internal/commands/helpers.go (suffix constants)
- internal/commands/{encrypt,decrypt,compress,rotate,watermark,reorder}.go (constant usage)
- internal/commands/commands_test.go (constant usage in tests)
- internal/testing/fixtures.go (bundled accidentally)

**PLAN-1.2 files touched:**
- internal/testing/fixtures.go (test helper refactoring)
- internal/cli/flags.go (log level change)

**Overlap:** `internal/testing/fixtures.go` appears in both plans via commit `bc85124`.

**Assessment:** The plans are logically independent (no code dependencies between suffix constants and test helpers). The commit bundling is purely a process issue, not a technical conflict. However, this makes it impossible to cleanly attribute changes to their respective plans in the git history.

---

## Conventions Review

**Go conventions:** ✓ Followed
- testing.TB interface usage is idiomatic
- t.Fatal() for test helper errors is standard practice
- Const naming follows Go conventions (exported PascalCase)

**Commit message conventions:** ✗ Violated
- Commit `bc85124` message "refactor test helpers to use testing.TB instead of panic" does not mention the suffix constants added to helpers.go
- Per convention, commit messages should accurately describe ALL changes in the commit

**Plan execution conventions:** ✗ Violated
- Each plan should produce distinct, non-overlapping commits
- Mixing changes from multiple plans in a single commit defeats the purpose of having separate plans

---

## Critical Findings

### Critical-1: Commit Bundling Breaks Plan Traceability
**Location:** Commit `bc85124f9142bb54bd0315aa126dcccd98249884`
**Issue:** This commit contains changes from both PLAN-1.1 (suffix constants in helpers.go) and PLAN-1.2 (test helper refactoring in fixtures.go), making it impossible to trace which changes belong to which plan in the git history.

**Remediation:**
This issue requires a git history rewrite to properly separate the concerns. The options are:

1. **Interactive rebase** (cleanest but most disruptive):
   ```bash
   # Create backup branch first
   git branch backup-phase4

   # Interactive rebase to split bc85124
   git rebase -i bc85124^
   # Mark bc85124 as "edit"
   # When stopped:
   git reset HEAD^
   git add internal/commands/helpers.go
   git commit -m "shipyard(phase-4): define output suffix constants in helpers.go"
   git add internal/testing/fixtures.go
   git commit -m "shipyard(phase-4): refactor test helpers to use testing.TB instead of panic"
   git rebase --continue
   ```

2. **Accept as technical debt** (pragmatic if downstream work exists):
   - Document the bundling issue in `.shipyard/ISSUES.md`
   - Add a note to SUMMARY-1.2 explaining the root cause
   - Ensure future plans maintain strict commit separation
   - Consider this a lessons-learned for improving the pre-commit hook workflow

**Recommendation:** Given that this is already merged and subsequent commits (aedda33, cc46f8c, 945e9dc) have been added on top, **option 2 (document as technical debt) is recommended** unless the user specifically requests a history rewrite. The functional implementation is correct; only the commit organization is suboptimal.

---

## Important Findings

None. The code quality is good once the commit organization issue is resolved.

---

## Suggestions

### Suggestion-1: Add test coverage for new test helper signatures
**Location:** `/Users/lgbarn/Personal/pdf-cli/internal/testing/fixtures.go`
**Issue:** TempDir() and TempFile() have new signatures accepting testing.TB, but there are no tests verifying they correctly call t.Fatal() on errors.

**Remediation:**
Add a test file `/Users/lgbarn/Personal/pdf-cli/internal/testing/fixtures_test.go`:
```go
package testing

import (
	"testing"
)

type mockTB struct {
	testing.TB
	fatalCalled bool
	fatalMsg    string
}

func (m *mockTB) Fatal(args ...interface{}) {
	m.fatalCalled = true
	if len(args) > 0 {
		m.fatalMsg = args[0].(string)
	}
}

func (m *mockTB) Helper() {}

func TestTempDir_CallsFatalOnError(t *testing.T) {
	// Create TempDir in an impossible location to trigger error
	mock := &mockTB{}
	_, cleanup := TempDir(mock, "/impossible/prefix")
	defer cleanup()

	if !mock.fatalCalled {
		t.Error("Expected TempDir to call t.Fatal() on error")
	}
}

func TestTempFile_CallsFatalOnError(t *testing.T) {
	// Similar test for TempFile
	mock := &mockTB{}
	_, cleanup := TempFile(mock, "/impossible/prefix", "content")
	defer cleanup()

	if !mock.fatalCalled {
		t.Error("Expected TempFile to call t.Fatal() on error")
	}
}
```

**Note:** This is a suggestion only. Since these helpers currently have zero callers and are straightforward wrappers around os.MkdirTemp/os.CreateTemp, the risk is low. However, if they gain callers in the future, having test coverage would be valuable.

---

### Suggestion-2: Update pre-commit hook to prevent cross-plan bundling
**Location:** Pre-commit hook workflow (exact location TBD)
**Issue:** The pre-commit hook automatically modified fixtures.go while working on PLAN-1.1, causing commit bundling. This suggests the hook may be too aggressive in auto-fixing issues across the entire codebase.

**Remediation:**
Consider one of the following approaches:
1. Scope pre-commit hook fixes to only the files explicitly staged for commit
2. Add a flag to selectively disable auto-fixes when working on multi-plan phases
3. Run pre-commit checks but require manual intervention for fixes outside the current plan's scope

This would prevent similar bundling issues in future phases.

---

### Suggestion-3: Document the zero-caller assumption
**Location:** `/Users/lgbarn/Personal/pdf-cli/internal/testing/fixtures.go` (doc comment)
**Issue:** The refactored TempDir and TempFile functions now have a different signature, but there's no indication in the code that this was a breaking change with zero impact because there were no callers.

**Remediation:**
Add a note to the function doc comments:
```go
// TempDir creates a temporary directory for test artifacts.
// Returns the path and a cleanup function.
//
// Note: Signature changed in Phase 4 to accept testing.TB for proper
// error reporting. This is a breaking change but had zero impact as
// the function had no callers at the time of refactoring.
func TempDir(t testing.TB, prefix string) (string, func()) {
```

This helps future developers understand the history and that the signature change was intentional.

---

## Summary

**Verdict:** REQUEST CHANGES

The implementation of R11 and R14 is functionally correct - all code changes meet the spec requirements, tests pass, and the build succeeds. However, **commit organization violates the separation of concerns between PLAN-1.1 and PLAN-1.2**, creating ambiguous git history.

**Root Cause:** The pre-commit hook auto-modified `internal/testing/fixtures.go` while working on PLAN-1.1, bundling both plans' changes into commit `bc85124`.

**Recommended Action:**
1. Document this as technical debt in `.shipyard/ISSUES.md` with category "Process"
2. Update SUMMARY-1.2 to explicitly acknowledge the commit bundling issue
3. Implement pre-commit hook improvements to prevent cross-plan bundling in future phases
4. Consider this plan "functionally complete but with process debt" for tracking purposes

**Findings Count:**
- Critical: 1 (commit bundling breaks plan traceability)
- Important: 0
- Suggestions: 3 (test coverage, pre-commit hook improvements, documentation)

---

## Appendix: Verification Commands Run

```bash
# Spec compliance checks
grep -n 'panic(' internal/testing/fixtures.go
grep -E "TempDir\(" internal/cli/password_test.go
grep -E "TempDir\(" internal/cleanup/cleanup_test.go
grep -E "TempDir\(" internal/config/config_test.go
grep -E "TempDir\(" internal/ocr/ocr_test.go
grep 'PersistentFlags.*StringVar.*logLevel.*"error"' internal/cli/flags.go

# Build and test verification
go build ./cmd/pdf
go test ./...

# Commit history verification
git log --oneline -10
git show bc85124 --stat
git show 945e9dc --stat
git show bc85124 internal/commands/helpers.go
git show bc85124 internal/testing/fixtures.go
```

All verification commands executed successfully and confirmed the implementation correctness (with the noted commit organization issue).
