---
phase: phase-3-concurrency-error-handling
plan: 1.2
wave: 1
dependencies: []
must_haves:
  - R9: Password file containing binary data produces warning on stderr
  - Warning-only approach (no errors, password content still returned)
  - Count non-printable characters excluding common whitespace (space, tab, \n, \r)
files_touched:
  - internal/cli/password.go
  - internal/cli/password_test.go
tdd: true
---

# Plan 1.2: Password File Printable Character Validation (R9)

## Overview
Add validation for password file content to detect binary data or non-printable characters. Per CONTEXT-3.md decision, this is warning-only: print to stderr but still return the password content. This avoids breaking users who legitimately use binary-looking passwords.

## Tasks

<task id="1" files="internal/cli/password_test.go" tdd="true">
  <action>
Add test case for binary content detection before implementing the feature (TDD approach):

Add new test function after TestReadPassword_PasswordFlagWithOptIn (after line 189):
```go
func TestReadPassword_BinaryContentWarning(t *testing.T) {
    tmpDir := t.TempDir()
    pwdFile := filepath.Join(tmpDir, "binary.txt")

    // Create file with non-printable characters (binary data)
    // Include some printable chars mixed with binary to simulate real wrong-file scenario
    binaryData := []byte{0x00, 0x01, 0x02, 'p', 'a', 's', 's', 0xFF, 0xFE}
    if err := os.WriteFile(pwdFile, binaryData, 0600); err != nil {
        t.Fatal(err)
    }

    cmd := newTestCmd()
    cmd.Flags().Set("password-file", pwdFile)

    // Capture stderr to verify warning is printed
    oldStderr := os.Stderr
    r, w, _ := os.Pipe()
    os.Stderr = w
    defer func() { os.Stderr = oldStderr }()

    // Run in goroutine to capture stderr
    stderrChan := make(chan string)
    go func() {
        var buf strings.Builder
        io.Copy(&buf, r)
        stderrChan <- buf.String()
    }()

    // Should still return password content despite binary data
    got, err := ReadPassword(cmd, "")

    w.Close()
    stderrOutput := <-stderrChan

    if err != nil {
        t.Fatalf("unexpected error: %v (should succeed with warning)", err)
    }

    // Password content should be returned (warning-only approach)
    expected := strings.TrimSpace(string(binaryData))
    if got != expected {
        t.Errorf("got %q, want %q (content should be returned despite warning)", got, expected)
    }

    // Verify warning was printed to stderr
    if !strings.Contains(stderrOutput, "WARNING") {
        t.Errorf("expected WARNING in stderr, got: %q", stderrOutput)
    }
    if !strings.Contains(stderrOutput, "non-printable") {
        t.Errorf("expected 'non-printable' in stderr, got: %q", stderrOutput)
    }
    // Should mention the count (4 non-printable chars: 0x00, 0x01, 0x02, 0xFF, 0xFE)
    if !strings.Contains(stderrOutput, "4") && !strings.Contains(stderrOutput, "5") {
        t.Errorf("expected count of non-printable characters in stderr, got: %q", stderrOutput)
    }
}

func TestReadPassword_PrintableContent_NoWarning(t *testing.T) {
    tmpDir := t.TempDir()
    pwdFile := filepath.Join(tmpDir, "printable.txt")

    // Create file with only printable characters and common whitespace
    printableData := []byte("MyP@ssw0rd!\nwith newline\tand tab  ")
    if err := os.WriteFile(pwdFile, printableData, 0600); err != nil {
        t.Fatal(err)
    }

    cmd := newTestCmd()
    cmd.Flags().Set("password-file", pwdFile)

    // Capture stderr to verify NO warning is printed
    oldStderr := os.Stderr
    r, w, _ := os.Pipe()
    os.Stderr = w
    defer func() { os.Stderr = oldStderr }()

    stderrChan := make(chan string)
    go func() {
        var buf strings.Builder
        io.Copy(&buf, r)
        stderrChan <- buf.String()
    }()

    got, err := ReadPassword(cmd, "")

    w.Close()
    stderrOutput := <-stderrChan

    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    expected := strings.TrimSpace(string(printableData))
    if got != expected {
        t.Errorf("got %q, want %q", got, expected)
    }

    // Verify NO warning for printable content
    if strings.Contains(stderrOutput, "WARNING") {
        t.Errorf("unexpected WARNING in stderr for printable content: %q", stderrOutput)
    }
}
```

Note: Add `import "io"` and `import "strings"` to the imports if not already present.
  </action>
  <verify>go test -v -run "TestReadPassword_BinaryContent|TestReadPassword_PrintableContent" /Users/lgbarn/Personal/pdf-cli/internal/cli/... 2>&1 | grep -E "(FAIL|--- FAIL)"</verify>
  <done>Tests compile and fail as expected (validation not yet implemented). Test output shows "FAIL" for both new tests, confirming TDD approach is working.</done>
</task>

<task id="2" files="internal/cli/password.go" tdd="false">
  <action>
Implement printable character validation for password file content:

1. Add `"unicode"` import at top of file (around line 3)

2. In ReadPassword function, after reading file content and checking size (between line 38 and the return statement), add validation logic:

```go
// After line 37: if len(data) > 1024 { ... }
// Before line 38: return strings.TrimSpace(string(data)), nil

// Validate printable content
content := string(data)
nonPrintableCount := 0
for _, r := range content {
    // Allow common whitespace characters
    if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
        continue
    }
    if !unicode.IsPrint(r) {
        nonPrintableCount++
    }
}

if nonPrintableCount > 0 {
    fmt.Fprintf(os.Stderr, "WARNING: Password file contains %d non-printable character(s). "+
        "This may indicate you're reading the wrong file.\n", nonPrintableCount)
}

return strings.TrimSpace(content), nil
```

This implementation:
- Uses `unicode.IsPrint()` to check for printable characters
- Explicitly allows common whitespace (space, tab, \n, \r) which are valid in passwords
- Counts and reports non-printable characters
- Prints warning to stderr if any non-printable characters found
- Returns password content regardless (warning-only approach per CONTEXT-3.md)
  </action>
  <verify>go test -v -run "TestReadPassword_BinaryContent|TestReadPassword_PrintableContent" /Users/lgbarn/Personal/pdf-cli/internal/cli/...</verify>
  <done>Both new tests pass. TestReadPassword_BinaryContentWarning confirms warning is printed for binary data and content is still returned. TestReadPassword_PrintableContent_NoWarning confirms no warning for printable content. All existing password tests still pass.</done>
</task>

<task id="3" files="internal/cli/password.go, internal/cli/password_test.go" tdd="false">
  <action>
Verify complete password file validation implementation:

1. Run all password tests to ensure no regressions
2. Verify warning behavior with a manual test case (if needed)
3. Confirm import of "unicode" package is present
4. Verify performance impact is negligible (validation is O(n) where n = password length, max 1KB)

The implementation should:
- Print warning to stderr when non-printable characters detected
- Still return password content (not an error condition)
- Not break any existing password reading functionality
- Work correctly with all password sources (file, env var, flag, prompt)
  </action>
  <verify>go test -v /Users/lgbarn/Personal/pdf-cli/internal/cli/... && go test -race /Users/lgbarn/Personal/pdf-cli/internal/cli/...</verify>
  <done>All password tests pass including new binary content tests. Race detector shows no issues. Password file validation correctly warns on binary content while maintaining backward compatibility with existing workflows.</done>
</task>

## Success Criteria
- ✓ Password file with binary data produces warning on stderr
- ✓ Warning message includes count of non-printable characters
- ✓ Password content is still returned (warning-only, not error)
- ✓ Printable content (including common whitespace) produces no warning
- ✓ All existing password tests continue to pass
- ✓ New tests: TestReadPassword_BinaryContentWarning and TestReadPassword_PrintableContent_NoWarning pass
- ✓ Race detector shows no issues

## Notes
- Per CONTEXT-3.md: Warning-only approach to avoid breaking legitimate use cases
- Validation checks for non-printable characters using `unicode.IsPrint()`
- Common whitespace (space, tab, \n, \r) explicitly allowed
- Warning printed to stderr, not stdout (doesn't interfere with command output)
- Maximum password file size remains 1KB (existing limit)
