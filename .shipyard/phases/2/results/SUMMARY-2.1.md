# Summary: PLAN-2.1 — Password Flag Security Lockdown (R2)

## Overview
Successfully implemented security hardening for the `--password` flag by making it non-functional unless explicitly opted-in with `--allow-insecure-password`. This addresses R2 from the "Remaining Tech Debt" milestone.

## Completed Tasks

### Task 1: Add flag infrastructure and update password validation logic
**Commit:** ea5075c - `feat(cli): add --allow-insecure-password opt-in for --password flag`

**Changes:**
- Added `AddAllowInsecurePasswordFlag()` function to `/Users/lgbarn/Personal/pdf-cli/internal/cli/flags.go`
- Added `GetAllowInsecurePassword()` accessor function to `/Users/lgbarn/Personal/pdf-cli/internal/cli/flags.go`
- Modified `ReadPassword()` in `/Users/lgbarn/Personal/pdf-cli/internal/cli/password.go` to:
  - Return clear error when `--password` used without `--allow-insecure-password`
  - Error message lists all 3 secure alternatives:
    1. `--password-file <path>` (recommended for automation)
    2. `PDF_CLI_PASSWORD` env var (recommended for CI/scripts)
    3. Interactive prompt (recommended for manual use)
  - Show WARNING (not deprecated) when both flags are present
  - Handle case where opt-in flag hasn't been registered

### Task 2: Register flag in all 14 commands and update tests
**Commit:** 354e688 - `feat(cli): register --allow-insecure-password across all commands`

**Command Files Updated (14):**
1. `/Users/lgbarn/Personal/pdf-cli/internal/commands/compress.go`
2. `/Users/lgbarn/Personal/pdf-cli/internal/commands/decrypt.go`
3. `/Users/lgbarn/Personal/pdf-cli/internal/commands/encrypt.go`
4. `/Users/lgbarn/Personal/pdf-cli/internal/commands/extract.go`
5. `/Users/lgbarn/Personal/pdf-cli/internal/commands/images.go`
6. `/Users/lgbarn/Personal/pdf-cli/internal/commands/info.go`
7. `/Users/lgbarn/Personal/pdf-cli/internal/commands/merge.go`
8. `/Users/lgbarn/Personal/pdf-cli/internal/commands/meta.go`
9. `/Users/lgbarn/Personal/pdf-cli/internal/commands/pdfa.go` (2 subcommands)
10. `/Users/lgbarn/Personal/pdf-cli/internal/commands/reorder.go`
11. `/Users/lgbarn/Personal/pdf-cli/internal/commands/rotate.go`
12. `/Users/lgbarn/Personal/pdf-cli/internal/commands/split.go`
13. `/Users/lgbarn/Personal/pdf-cli/internal/commands/text.go`
14. `/Users/lgbarn/Personal/pdf-cli/internal/commands/watermark.go`

Each command now has `cli.AddAllowInsecurePasswordFlag(cmdVar)` registered in its `init()` function, placed right after `cli.AddPasswordFileFlag()`.

**Test Files Updated:**
- `/Users/lgbarn/Personal/pdf-cli/internal/cli/password_test.go`:
  - Updated `newTestCmd()` helper to include the new flag
  - Updated `TestReadPassword_DeprecatedFlag` to set opt-in flag
  - Updated `TestReadPassword_Priority_EnvOverFlag` to set opt-in flag
  - Added `TestReadPassword_PasswordFlagWithoutOptIn` - verifies error when missing opt-in
  - Added `TestReadPassword_PasswordFlagWithOptIn` - verifies success with opt-in

- `/Users/lgbarn/Personal/pdf-cli/internal/commands/stdio_commands_test.go`:
  - Updated 4 test cases using `--password` to include `--allow-insecure-password`

- `/Users/lgbarn/Personal/pdf-cli/internal/commands/dryrun_test.go`:
  - Updated 5 test cases using `--password` to include `--allow-insecure-password`

### Task 3: Verification (No commit)

**Test Results:**
```
go test -race -v ./internal/cli/...
  PASS - All 48 tests passed

go test -race ./internal/commands/...
  PASS - All tests passed in 1.945s

go test -race ./...
  PASS - Full suite passed (all 15 packages)

go test -cover ./internal/cli/...
  Coverage: 82.7% (exceeds 75% requirement)

golangci-lint run ./internal/cli/... ./internal/commands/...
  0 issues
```

## Implementation Details

### Security Model
The implementation follows a secure-by-default approach:
1. `--password` flag is BLOCKED unless `--allow-insecure-password` is present
2. Error message is instructive, listing all secure alternatives
3. When opt-in is used, a WARNING is shown (removed "deprecated" language)
4. Priority order unchanged: `--password-file` > `PDF_CLI_PASSWORD` > `--password` > interactive

### Backward Compatibility Break
This is an **intentional breaking change** for security:
- Scripts using `--password` without opt-in will now fail with a clear error
- Users are guided to migrate to secure alternatives
- The `--allow-insecure-password` opt-in provides an escape hatch for urgent cases

### Code Quality
- All changes follow Go 1.25 conventions
- Import groups properly ordered (stdlib first, then external)
- Error wrapping uses `fmt.Errorf` with `%w`
- No new linter warnings
- Test coverage maintained above 75%

## Deviations from Plan
None. All tasks completed exactly as specified.

## Files Changed

### Source Files (17):
- internal/cli/flags.go
- internal/cli/password.go
- internal/commands/compress.go
- internal/commands/decrypt.go
- internal/commands/encrypt.go
- internal/commands/extract.go
- internal/commands/images.go
- internal/commands/info.go
- internal/commands/merge.go
- internal/commands/meta.go
- internal/commands/pdfa.go
- internal/commands/reorder.go
- internal/commands/rotate.go
- internal/commands/split.go
- internal/commands/text.go
- internal/commands/watermark.go

### Test Files (3):
- internal/cli/password_test.go
- internal/commands/stdio_commands_test.go
- internal/commands/dryrun_test.go

## Verification Evidence

### New Test Cases Added
1. `TestReadPassword_PasswordFlagWithoutOptIn` - Ensures error is returned with helpful message
2. `TestReadPassword_PasswordFlagWithOptIn` - Ensures password is accepted with opt-in

### Error Message Validation
The error message includes all required strings:
- `--password-file`
- `PDF_CLI_PASSWORD`
- `Interactive prompt`
- `--allow-insecure-password`

### Race Detection
All tests pass with `-race` flag, confirming no concurrency issues.

## Status
✅ **COMPLETE** - All tasks executed successfully, all verification steps passed.
