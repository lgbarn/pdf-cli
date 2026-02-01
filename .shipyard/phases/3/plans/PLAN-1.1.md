---
phase: security-hardening
plan: 01
wave: 1
dependencies: []
must_haves:
  - Remove password exposure from process listings
  - Support PDF_CLI_PASSWORD environment variable
  - Support --password-file flag
  - Support interactive terminal prompt
  - Deprecate --password flag with warning
files_touched:
  - internal/cli/password.go (new)
  - internal/cli/flags.go
  - internal/commands/encrypt.go
  - internal/commands/decrypt.go
  - internal/commands/info.go
  - internal/commands/text.go
  - internal/commands/pdfa.go
  - internal/commands/extract.go
  - internal/commands/images.go
  - internal/commands/meta.go
  - internal/commands/watermark.go
  - internal/commands/rotate.go
  - internal/commands/split.go
  - internal/commands/reorder.go
  - internal/commands/merge.go
  - internal/commands/compress.go
tdd: false
---

# Plan 1.1: Password Security (R1)

## Goal
Eliminate password exposure in process listings by implementing secure password input methods: environment variable, file-based input, and interactive terminal prompts. The --password flag will be deprecated but retained with a security warning for backward compatibility.

## Context
Currently, 14 command files accept passwords via the --password CLI flag, which exposes passwords in process listings (`ps aux`). This is a critical P0 security issue. We must provide three secure alternatives while maintaining backward compatibility with a deprecation warning.

The golang.org/x/term package is already available as a direct dependency in go.mod (line 11).

## Tasks

<task id="1" files="internal/cli/password.go" tdd="false">
  <action>
Create internal/cli/password.go with secure password reading logic.

Function signature:
```go
// ReadPassword reads a password securely from multiple sources with priority:
// 1. --password-file flag (if present)
// 2. PDF_CLI_PASSWORD environment variable (if set)
// 3. --password flag (deprecated, shows warning)
// 4. Interactive terminal prompt (if terminal and not CI/batch mode)
// Returns empty string if no password source available.
func ReadPassword(cmd *cobra.Command, promptMsg string) (string, error)
```

Implementation requirements:
- Check --password-file flag first. If present, read file contents and trim whitespace. Return error if file doesn't exist or can't be read. Use os.ReadFile with size limit of 1KB.
- Check PDF_CLI_PASSWORD env var second. If set and non-empty, return its value.
- Check --password flag third. If set, print deprecation warning to stderr: "WARNING: --password flag is deprecated and exposes passwords in process listings. Use --password-file, PDF_CLI_PASSWORD, or interactive prompt instead."
- If none of the above and promptMsg is non-empty and terminal is interactive (check with term.IsTerminal(int(os.Stdin.Fd()))), prompt user with promptMsg and read password using term.ReadPassword.
- Return empty string if no password available from any source.

Helper function:
```go
// isInteractiveTerminal returns true if stdin is an interactive terminal and not in CI/batch mode
func isInteractiveTerminal() bool {
  return term.IsTerminal(int(os.Stdin.Fd())) && os.Getenv("CI") == "" && os.Getenv("PDF_CLI_BATCH") == ""
}
```

Error handling:
- Return error if --password-file specified but file unreadable
- Return error if terminal read fails
- Never return error for missing/empty password sources
  </action>
  <verify>
Build succeeds:
```bash
cd /Users/lgbarn/Personal/pdf-cli/.worktrees/phase-2-concurrency
go build -o /tmp/pdf-cli ./cmd/pdf
```

Unit test the function:
```bash
go test -v ./internal/cli -run TestReadPassword
```
  </verify>
  <done>
- internal/cli/password.go created
- ReadPassword function implemented with 4-tier priority
- isInteractiveTerminal helper function implemented
- File compiles without errors
- Unit tests pass covering all input methods
  </done>
</task>

<task id="2" files="internal/cli/flags.go,internal/cli/password.go" tdd="false">
  <action>
Add --password-file flag support and update password reading in internal/cli/flags.go.

1. Add new flag function after AddPasswordFlag (line 36):
```go
// AddPasswordFileFlag adds the --password-file flag to a command
func AddPasswordFileFlag(cmd *cobra.Command, usage string) {
  if usage == "" {
    usage = "Read password from file (more secure than --password)"
  }
  cmd.Flags().String("password-file", "", usage)
}
```

2. Replace GetPassword function (lines 50-54) with:
```go
// GetPassword gets the password using secure methods (DEPRECATED: use ReadPassword instead)
// This is kept for backward compatibility but ReadPassword should be used going forward.
func GetPassword(cmd *cobra.Command) string {
  password, _ := cmd.Flags().GetString("password")
  return password
}

// GetPasswordSecure reads password securely from multiple sources
// Prefer this over GetPassword for all new code.
func GetPasswordSecure(cmd *cobra.Command, promptMsg string) (string, error) {
  return ReadPassword(cmd, promptMsg)
}
```

3. Add unit tests in internal/cli/password_test.go:
- Test --password-file flag reading
- Test PDF_CLI_PASSWORD env var
- Test --password flag (deprecated path)
- Test priority order (file > env > flag)
- Test error cases (missing file, unreadable file)
- Mock term.IsTerminal for interactive tests
  </action>
  <verify>
Build succeeds:
```bash
cd /Users/lgbarn/Personal/pdf-cli/.worktrees/phase-2-concurrency
go build ./internal/cli
```

Run tests:
```bash
go test -v ./internal/cli -run TestReadPassword
go test -v ./internal/cli -run TestGetPasswordSecure
```
  </verify>
  <done>
- AddPasswordFileFlag function added to flags.go
- GetPassword function retained for backward compatibility
- GetPasswordSecure function added as new preferred API
- Unit tests pass with 100% coverage of password reading paths
  </done>
</task>

<task id="3" files="internal/commands/*.go" tdd="false">
  <action>
Update all 14 password-accepting commands to use the new secure password reading.

For each command file (encrypt.go, decrypt.go, info.go, text.go, pdfa.go, extract.go, images.go, meta.go, watermark.go, rotate.go, split.go, reorder.go, merge.go, compress.go):

1. In init() function, add --password-file flag after existing AddPasswordFlag:
```go
cli.AddPasswordFileFlag(cmdName, "Read password from file")
```

2. In RunE function, replace direct cli.GetPassword(cmd) call with:
```go
password, err := cli.GetPasswordSecure(cmd, "Enter PDF password: ")
if err != nil {
  return fmt.Errorf("failed to read password: %w", err)
}
```

3. For commands where password is required (encrypt.go, decrypt.go), update error message:
```go
if password == "" {
  return fmt.Errorf("password is required (use --password-file, PDF_CLI_PASSWORD env var, or interactive prompt)")
}
```

4. Remove MarkFlagRequired("password") from init() functions in encrypt.go and decrypt.go since password can come from multiple sources now.

Specific files to update:
- internal/commands/encrypt.go (line 45: runEncrypt)
- internal/commands/decrypt.go (line 44: runDecrypt)
- internal/commands/info.go (line 45: runInfo)
- internal/commands/text.go (line 53: runText)
- internal/commands/pdfa.go (find password usage)
- internal/commands/extract.go (find password usage)
- internal/commands/images.go (find password usage)
- internal/commands/meta.go (find password usage)
- internal/commands/watermark.go (find password usage)
- internal/commands/rotate.go (find password usage)
- internal/commands/split.go (find password usage)
- internal/commands/reorder.go (find password usage)
- internal/commands/merge.go (find password usage)
- internal/commands/compress.go (find password usage)
  </action>
  <verify>
Build the CLI:
```bash
cd /Users/lgbarn/Personal/pdf-cli/.worktrees/phase-2-concurrency
go build -o /tmp/pdf-cli ./cmd/pdf
```

Test each password input method:
```bash
# Test --password-file
echo "testpass" > /tmp/pwd.txt
/tmp/pdf-cli encrypt testdata/sample.pdf -o /tmp/out.pdf --password-file /tmp/pwd.txt

# Test PDF_CLI_PASSWORD
PDF_CLI_PASSWORD=testpass /tmp/pdf-cli encrypt testdata/sample.pdf -o /tmp/out.pdf

# Test deprecated --password (should show warning)
/tmp/pdf-cli encrypt testdata/sample.pdf -o /tmp/out.pdf --password testpass 2>&1 | grep WARNING
```

Run integration tests:
```bash
go test -v ./internal/commands -run Integration
```
  </verify>
  <done>
- All 14 command files updated with --password-file flag
- All password reading switched to GetPasswordSecure
- MarkFlagRequired removed from encrypt/decrypt
- Deprecation warning displays for --password flag usage
- Integration tests pass
- Manual testing confirms all three input methods work
  </done>
</task>

## Verification Strategy

Manual testing checklist:
1. Test --password-file with encrypt command
2. Test PDF_CLI_PASSWORD with decrypt command
3. Test interactive prompt (when terminal is TTY)
4. Test --password still works but shows warning
5. Verify password not visible in `ps aux` output
6. Test error cases (missing file, unreadable file)

Success criteria:
- All password input methods work correctly
- Priority order is respected (file > env > flag > prompt)
- Deprecation warning shown for --password
- No passwords visible in process listings
- All existing integration tests pass
- Backward compatibility maintained

## Breaking Changes

**Minor breaking change**: Commands requiring password (encrypt, decrypt) will now fail if password not provided via any method, instead of showing a flag error. Error message will guide users to use --password-file, PDF_CLI_PASSWORD, or interactive prompt.

## Security Impact

**HIGH IMPACT**: Eliminates critical security vulnerability where passwords are exposed in process listings. Users can now use secure methods:
- File-based: `--password-file /path/to/secret.txt`
- Environment: `PDF_CLI_PASSWORD=secret pdf encrypt ...`
- Interactive: prompted when terminal is TTY
