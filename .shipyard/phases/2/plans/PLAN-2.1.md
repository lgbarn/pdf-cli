# Plan 2.1: Password Flag Security Lockdown

## Context
This plan implements R2 from Phase 2: Security Hardening. It makes the `--password` flag non-functional by default, requiring users to explicitly add `--allow-insecure-password` to use it.

The `--password` flag has been deprecated since v1.x with a warning message, but it still works. This plan makes it fail with a clear error listing three secure alternatives:
1. `--password-file <path>` (recommended for automation)
2. `PDF_CLI_PASSWORD` env var (recommended for CI/scripts)
3. Interactive prompt (recommended for manual use)

This is a breaking change for users relying on `--password`, but the error message provides clear migration paths and an escape hatch via `--allow-insecure-password`.

## Dependencies
Must execute after Wave 1 (PLAN-1.1 and PLAN-1.2) to keep Phase 2 focused and reduce merge conflicts. R2 affects CLI behavior and needs careful testing.

## Tasks

### Task 1: Add flag infrastructure and update password validation logic
**Files:**
- `/Users/lgbarn/Personal/pdf-cli/internal/cli/flags.go`
- `/Users/lgbarn/Personal/pdf-cli/internal/cli/password.go`

**Action:** Modify
**Description:**

**Part A — flags.go:** Add two new functions to support the `--allow-insecure-password` flag:

1. Add flag registration function after `AddPasswordFileFlag` (around line 44):
   ```go
   // AddAllowInsecurePasswordFlag adds the --allow-insecure-password flag to a command.
   func AddAllowInsecurePasswordFlag(cmd *cobra.Command) {
       cmd.Flags().Bool("allow-insecure-password", false, "Allow using --password flag (insecure, not recommended)")
   }
   ```

2. Add accessor function after `GetPasswordSecure` (around line 67):
   ```go
   // GetAllowInsecurePassword gets the allow-insecure-password flag value.
   func GetAllowInsecurePassword(cmd *cobra.Command) bool {
       allow, _ := cmd.Flags().GetBool("allow-insecure-password")
       return allow
   }
   ```

**Part B — password.go:** Modify the `ReadPassword` function to check for `--allow-insecure-password` before accepting `--password` flag.

Update step 3 (lines 47-54) from:
```go
// 3. Check --password flag (deprecated)
if cmd.Flags().Lookup("password") != nil {
    password, _ := cmd.Flags().GetString("password")
    if password != "" {
        fmt.Fprintln(os.Stderr, "WARNING: --password flag is deprecated and exposes passwords in process listings. Use --password-file, PDF_CLI_PASSWORD, or interactive prompt instead.")
        return password, nil
    }
}
```

To:
```go
// 3. Check --password flag (requires opt-in)
if cmd.Flags().Lookup("password") != nil {
    password, _ := cmd.Flags().GetString("password")
    if password != "" {
        // Check if opt-in flag is present
        allowInsecure := false
        if cmd.Flags().Lookup("allow-insecure-password") != nil {
            allowInsecure, _ = cmd.Flags().GetBool("allow-insecure-password")
        }

        if !allowInsecure {
            return "", fmt.Errorf(
                "--password flag is insecure and disabled by default.\n" +
                "Use one of these secure alternatives:\n" +
                "  1. --password-file <path>        (recommended for automation)\n" +
                "  2. PDF_CLI_PASSWORD env var      (recommended for CI/scripts)\n" +
                "  3. Interactive prompt            (recommended for manual use)\n\n" +
                "To use --password anyway (not recommended), add --allow-insecure-password",
            )
        }

        fmt.Fprintln(os.Stderr, "WARNING: --password flag exposes passwords in process listings. Use --password-file, PDF_CLI_PASSWORD, or interactive prompt instead.")
        return password, nil
    }
}
```

**Acceptance Criteria:**
- AddAllowInsecurePasswordFlag and GetAllowInsecurePassword functions added to flags.go
- ReadPassword returns error when --password used without --allow-insecure-password
- Error message lists all three secure alternatives and the escape hatch
- Warning still shown when opt-in is present
- Code compiles successfully

### Task 2: Register flag in all 14 commands and update tests
**Files:**
- `/Users/lgbarn/Personal/pdf-cli/internal/commands/compress.go`
- `/Users/lgbarn/Personal/pdf-cli/internal/commands/decrypt.go`
- `/Users/lgbarn/Personal/pdf-cli/internal/commands/encrypt.go`
- `/Users/lgbarn/Personal/pdf-cli/internal/commands/extract.go`
- `/Users/lgbarn/Personal/pdf-cli/internal/commands/images.go`
- `/Users/lgbarn/Personal/pdf-cli/internal/commands/info.go`
- `/Users/lgbarn/Personal/pdf-cli/internal/commands/merge.go`
- `/Users/lgbarn/Personal/pdf-cli/internal/commands/meta.go`
- `/Users/lgbarn/Personal/pdf-cli/internal/commands/pdfa.go`
- `/Users/lgbarn/Personal/pdf-cli/internal/commands/reorder.go`
- `/Users/lgbarn/Personal/pdf-cli/internal/commands/rotate.go`
- `/Users/lgbarn/Personal/pdf-cli/internal/commands/split.go`
- `/Users/lgbarn/Personal/pdf-cli/internal/commands/text.go`
- `/Users/lgbarn/Personal/pdf-cli/internal/commands/watermark.go`
- `/Users/lgbarn/Personal/pdf-cli/internal/cli/password_test.go`

**Action:** Modify
**Description:**

**Part A — Command files:** Add `cli.AddAllowInsecurePasswordFlag()` call to each command's `init()` function, right after the existing `cli.AddPasswordFlag()` and `cli.AddPasswordFileFlag()` calls.

For each file, find the `init()` function and add one line. Example pattern (from encrypt.go):
```go
func init() {
    ...
    cli.AddPasswordFlag(encryptCmd, "Password for encryption")
    cli.AddPasswordFileFlag(encryptCmd, "")
    cli.AddAllowInsecurePasswordFlag(encryptCmd)  // ADD THIS LINE
    ...
}
```

All 14 files follow the same pattern. Search for `AddPasswordFlag` to find the correct location in each file's init() function.

**Part B — password_test.go:** Update existing tests and add two new tests.

1. Update `newTestCmd()` helper to add the opt-in flag:
   ```go
   func newTestCmd() *cobra.Command {
       cmd := &cobra.Command{Use: "test"}
       cmd.Flags().String("password", "", "")
       cmd.Flags().String("password-file", "", "")
       cmd.Flags().Bool("allow-insecure-password", false, "")  // ADD THIS LINE
       return cmd
   }
   ```

2. Update `TestReadPassword_DeprecatedFlag` to set opt-in flag:
   ```go
   cmd.Flags().Set("allow-insecure-password", "true")  // ADD after cmd.Flags().Set("password", ...)
   ```

3. Update `TestReadPassword_Priority_EnvOverFlag` to set opt-in flag:
   ```go
   cmd.Flags().Set("allow-insecure-password", "true")  // ADD after cmd.Flags().Set("password", ...)
   ```

4. Add new test `TestReadPassword_PasswordFlagWithoutOptIn`:
   ```go
   func TestReadPassword_PasswordFlagWithoutOptIn(t *testing.T) {
       cmd := newTestCmd()
       cmd.Flags().Set("password", "secret123")

       _, err := ReadPassword(cmd, "")
       if err == nil {
           t.Fatal("Expected error when using --password without opt-in")
       }

       errMsg := err.Error()
       if !strings.Contains(errMsg, "--password-file") {
           t.Error("Error message should mention --password-file")
       }
       if !strings.Contains(errMsg, "PDF_CLI_PASSWORD") {
           t.Error("Error message should mention PDF_CLI_PASSWORD")
       }
       if !strings.Contains(errMsg, "Interactive prompt") {
           t.Error("Error message should mention interactive prompt")
       }
       if !strings.Contains(errMsg, "--allow-insecure-password") {
           t.Error("Error message should mention --allow-insecure-password")
       }
   }
   ```

5. Add new test `TestReadPassword_PasswordFlagWithOptIn`:
   ```go
   func TestReadPassword_PasswordFlagWithOptIn(t *testing.T) {
       cmd := newTestCmd()
       cmd.Flags().Set("password", "secret123")
       cmd.Flags().Set("allow-insecure-password", "true")

       password, err := ReadPassword(cmd, "")
       if err != nil {
           t.Fatalf("Unexpected error with opt-in flag: %v", err)
       }
       if password != "secret123" {
           t.Errorf("Expected 'secret123', got '%s'", password)
       }
   }
   ```

**Acceptance Criteria:**
- All 14 command files have `cli.AddAllowInsecurePasswordFlag(cmdName)` added to init()
- `newTestCmd()` helper includes --allow-insecure-password flag
- Existing tests updated to set opt-in flag where needed
- Two new tests verify reject-without-opt-in and accept-with-opt-in behavior
- Error message test verifies all three alternatives are mentioned
- All tests pass: `go test -race ./internal/cli/... ./internal/commands/...`

### Task 3: Verify implementation end-to-end
**Files:** All command and CLI test files
**Action:** Test
**Description:**
Run comprehensive verification:

```bash
# Run CLI package tests
go test -race -v ./internal/cli/...

# Run all command tests to ensure flag registration works
go test -race ./internal/commands/...

# Full test suite
go test -race ./...

# Check coverage
go test -cover ./internal/cli/...
```

**Acceptance Criteria:**
- All tests in `./internal/cli/...` pass with `-race` flag (including 2 new tests)
- All tests in `./internal/commands/...` pass with `-race` flag
- Full test suite passes: `go test -race ./...`
- Test coverage >= 75% for cli package
- No regressions in password reading functionality

## Verification

Run all verification commands:

```bash
# Run password tests
go test -race -v -run TestReadPassword ./internal/cli/...

# Run CLI tests
go test -race ./internal/cli/...

# Run command tests
go test -race ./internal/commands/...

# Full test suite
go test -race ./...
```

## Success Criteria
- Running `pdf encrypt --password secret test.pdf` without `--allow-insecure-password` produces error
- Error message mentions `--password-file`, `PDF_CLI_PASSWORD`, and interactive prompt
- Error message mentions `--allow-insecure-password` as escape hatch
- Running `pdf encrypt --password secret --allow-insecure-password test.pdf` succeeds with warning
- All 14 commands accept `--allow-insecure-password` flag
- All tests pass with `-race` flag
- Test coverage >= 75% for cli package
- No regressions in secure password reading (password-file, env var, interactive prompt)

## Notes
This is a breaking change for users currently using `--password`. The error message provides clear migration paths and an escape hatch for users who cannot migrate immediately. The flag was already deprecated with warnings since v1.x, so impact should be minimal.
