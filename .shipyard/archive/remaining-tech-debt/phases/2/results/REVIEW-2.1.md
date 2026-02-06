# Review: PLAN-2.1 — Password Flag Security Lockdown (R2)

## Stage 1: Spec Compliance
**Verdict:** PASS

### Task 1: Add flag infrastructure and update password validation logic
- **Status:** PASS
- **Evidence:**
  - **flags.go** (lines 46-55): `AddAllowInsecurePasswordFlag()` and `GetAllowInsecurePassword()` functions added exactly as specified in the plan. Function signature, flag name, and default value match the plan.
  - **password.go** (lines 47-70): `ReadPassword()` correctly implements the opt-in check. When `--password` is used without `--allow-insecure-password`, it returns an error with the required message.
  - Error message (lines 58-64) lists all three secure alternatives as specified:
    1. `--password-file <path>` (recommended for automation)
    2. `PDF_CLI_PASSWORD env var` (recommended for CI/scripts)
    3. `Interactive prompt` (recommended for manual use)
  - Error message includes the escape hatch: `--allow-insecure-password`
  - When opt-in flag is present (line 67), warning is shown (not deprecated).
- **Notes:** Implementation exactly matches plan specification. Error message formatting is clean and helpful. The code properly handles the case where the opt-in flag hasn't been registered (lines 52-55).

### Task 2: Register flag in all 14 commands and update tests
- **Status:** PASS
- **Evidence:**
  - **Command files (14):** Grep search confirms all 14 command files contain `AddAllowInsecurePasswordFlag`:
    - compress.go (line 20)
    - decrypt.go (line 20)
    - encrypt.go (line 20)
    - extract.go (line 18)
    - images.go (line 18)
    - info.go (line 17)
    - merge.go (line 18)
    - meta.go (line 17)
    - pdfa.go (lines for both subcommands)
    - reorder.go (line 18)
    - rotate.go (line 18)
    - split.go (line 18)
    - text.go (line 18)
    - watermark.go (line 18)
  - All flags registered after `AddPasswordFileFlag` as specified.
  - **password_test.go** updates verified:
    - `newTestCmd()` helper updated (line 15) to include `allow-insecure-password` flag
    - `TestReadPassword_DeprecatedFlag` updated (line 85) to set opt-in flag
    - `TestReadPassword_Priority_EnvOverFlag` updated (line 123) to set opt-in flag
    - New test `TestReadPassword_PasswordFlagWithoutOptIn` (lines 149-173) verifies error when missing opt-in and checks all four required strings in error message
    - New test `TestReadPassword_PasswordFlagWithOptIn` (lines 175-189) verifies success with opt-in
  - **stdio_commands_test.go** updated: 4 test cases now include `--allow-insecure-password` (lines 49, 57, 73)
  - **dryrun_test.go** updated: 5 test cases now include `--allow-insecure-password` (lines 70, 85, 99, 115, 131)
- **Notes:** All command files updated consistently. Test coverage is comprehensive, covering both positive and negative cases. Error message validation in tests matches CONTEXT-2.md requirement.

### Task 3: Verify implementation end-to-end
- **Status:** PASS
- **Evidence:**
  - All password tests pass with `-race` flag: `go test -race -v -run TestReadPassword ./internal/cli/...` — 10 tests pass (includes 2 new tests)
  - CLI package tests pass: `go test -race ./internal/cli/...` — PASS
  - Command package tests pass: `go test -race ./internal/commands/...` — PASS
  - Full test suite passes: `go test -race ./...` — All 15 packages pass
  - Test coverage: 82.7% for cli package (exceeds 75% requirement)
- **Notes:** All verification commands pass. No regressions detected. Race detector found no issues.

---

## Stage 2: Code Quality
(Stage 1 passed — proceeding with code quality review)

### Critical
None.

### Important
None.

### Suggestions
None.

---

## Additional Observations

### Security Model
The implementation follows secure-by-default principles:
1. The `--password` flag is **blocked by default** — requires explicit opt-in
2. Error message is instructive and actionable — lists all alternatives
3. Priority order preserved: `--password-file` > `PDF_CLI_PASSWORD` > `--password` (with opt-in) > interactive
4. Warning message (not deprecated) shown when opt-in is used

### Backward Compatibility
This is an **intentional breaking change** for security:
- Scripts using `--password` without opt-in will fail with clear error
- Users are guided to migrate to secure alternatives
- The `--allow-insecure-password` opt-in provides escape hatch
- This aligns with the deprecation warnings present since v1.x

### Code Quality
- Import groups properly ordered (stdlib, then external packages)
- Error handling uses `fmt.Errorf` with `%w` for wrapping
- Flag registration follows consistent pattern across all commands
- Test helpers (`newTestCmd`) properly updated to match production flag setup
- Test naming follows Go conventions (`TestReadPassword_*`)

### Spec Compliance
- All tasks completed exactly as specified in PLAN-2.1.md
- Error message matches CONTEXT-2.md decision (lists all 3 alternatives)
- SUMMARY-2.1.md accurately reflects the implementation
- No deviations from plan

### Testing
- New tests verify both positive (with opt-in) and negative (without opt-in) cases
- Error message validation is thorough (checks all 4 required strings)
- Existing tests updated to maintain functionality
- Race detection passes (no concurrency issues)
- Test coverage exceeds requirements (82.7% vs 75% target)

---

## Summary
**Verdict:** APPROVE

Plan 2.1 (Password Flag Security Lockdown) has been implemented flawlessly. All tasks completed as specified, all verification steps pass, and code quality is excellent. The implementation correctly makes `--password` non-functional without `--allow-insecure-password`, provides clear error messages listing all secure alternatives, and maintains backward compatibility through an escape hatch. Test coverage is comprehensive and exceeds requirements.

**Findings:** Critical: 0 | Important: 0 | Suggestions: 0

**Recommendation:** Ready to proceed with next plan (PLAN-2.2 or Wave 3).
