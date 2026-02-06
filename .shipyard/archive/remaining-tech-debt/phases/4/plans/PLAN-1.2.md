---
phase: code-quality-constants
plan: 1.2
wave: 1
dependencies: []
must_haves:
  - R11: Refactor test helper functions to use testing.TB instead of panic()
  - R14: Change default CLI log level from "silent" to "error"
files_touched:
  - internal/testing/fixtures.go
  - internal/cli/flags.go
tdd: false
---

# Plan 1.2: Test Helpers and Default Log Level (R11, R14)

## Objective
Improve test helper ergonomics by using testing.TB for error reporting, and make CLI logging more useful by defaulting to "error" level instead of "silent".

## Tasks

<task id="1" files="internal/testing/fixtures.go" tdd="false">
  <action>Refactor TempDir() and TempFile() helper functions: (1) Add testing.TB parameter as first argument to both functions, (2) Replace all panic() calls with t.Fatal(), (3) Update function signatures and doc comments to reflect new parameter. TempDir has 1 panic call to replace. TempFile has 2 panic calls to replace. Do NOT modify TestdataDir() as its panic is out of scope.</action>
  <verify>cd /Users/lgbarn/Personal/pdf-cli && grep -c "func TempDir(t testing.TB" internal/testing/fixtures.go | grep -q "1" && grep -c "func TempFile(t testing.TB" internal/testing/fixtures.go | grep -q "1" && ! grep -E "panic\(" internal/testing/fixtures.go | grep -v TestdataDir</verify>
  <done>TempDir and TempFile both accept testing.TB as first parameter. All 3 panic() calls in these functions replaced with t.Fatal(). TestdataDir unchanged. No callers exist to update (zero current usage).</done>
</task>

<task id="2" files="internal/cli/flags.go" tdd="false">
  <action>In flags.go line 104, change the default log level from "silent" to "error" in the PersistentFlags().StringVarP call. The flag should now default to showing error-level messages unless user overrides with --log-level flag.</action>
  <verify>cd /Users/lgbarn/Personal/pdf-cli && grep 'PersistentFlags.*StringVarP.*logLevel.*"error"' internal/cli/flags.go</verify>
  <done>Default log level is "error". The flag definition in flags.go uses "error" as the default value string.</done>
</task>

<task id="3" files="internal/cli/flags.go, internal/testing/fixtures.go" tdd="false">
  <action>Run full test suite to verify: (1) test helpers refactor has no impact (zero callers), (2) logging tests still pass (they test package defaults, not CLI flag defaults), (3) CLI flag change doesn't break existing tests.</action>
  <verify>cd /Users/lgbarn/Personal/pdf-cli && go test ./... -v</verify>
  <done>All tests pass. Test helper changes are isolated (no callers). Logging package tests verify LevelSilent package default (unchanged). CLI integration tests pass with new "error" default.</done>
</task>

## Success Criteria
- TempDir and TempFile accept testing.TB parameter
- 3 panic() calls replaced with t.Fatal() in these two functions
- TestdataDir unchanged (panic out of scope)
- Default CLI log level is "error" instead of "silent"
- All existing tests pass
- Build succeeds

## Verification
```bash
cd /Users/lgbarn/Personal/pdf-cli
go build ./cmd/pdf
go test ./...
```

## Notes
- R11: These helpers have zero current callers, so the signature change has no ripple effects
- R14: The logging package's default (LevelSilent) is separate from the CLI flag default — only the flag changes
