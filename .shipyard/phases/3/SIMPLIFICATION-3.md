# Simplification Report
**Phase:** Phase 3 - Security Hardening
**Date:** 2026-01-31
**Files analyzed:** 32 files modified
**Findings:** 4 total (1 high priority, 2 medium priority, 1 low priority)

## High Priority

### H1: Repeated Input Path Sanitization Pattern (9 occurrences)
- **Type:** Consolidate
- **Locations:**
  - internal/commands/compress.go:47-51
  - internal/commands/decrypt.go:45-49
  - internal/commands/encrypt.go:46-50
  - internal/commands/merge.go:39-43
  - internal/commands/rotate.go:47-51
  - internal/commands/info.go:47-51
  - internal/commands/meta.go:50-54
  - internal/commands/combine_images.go:38-42
  - internal/commands/watermark.go:48-52
- **Description:** Nine command files contain identical 5-line blocks for sanitizing input paths:
  ```go
  // Sanitize input paths
  sanitizedArgs, err := fileio.SanitizePaths(args)
  if err != nil {
      return fmt.Errorf("invalid file path: %w", err)
  }
  args = sanitizedArgs
  ```
  This is exact duplication that violates the Rule of Three (3+ occurrences warrant extraction).
- **Suggestion:** Extract to `internal/commands/helpers.go`:
  ```go
  // sanitizeInputArgs validates and sanitizes input file paths.
  func sanitizeInputArgs(args []string) ([]string, error) {
      sanitized, err := fileio.SanitizePaths(args)
      if err != nil {
          return nil, fmt.Errorf("invalid file path: %w", err)
      }
      return sanitized, nil
  }
  ```
  Then replace all occurrences with single-line: `args, err := sanitizeInputArgs(args)`
- **Impact:**
  - Lines saved: ~36 lines (4 lines saved per occurrence × 9 occurrences)
  - Complexity reduced: Centralizes error message consistency
  - Maintainability: Single location to update validation logic

### H2: Repeated Output Path Sanitization Pattern (8 occurrences)
- **Type:** Consolidate
- **Locations:**
  - internal/commands/compress.go:61-66
  - internal/commands/decrypt.go:63-68
  - internal/commands/encrypt.go:65-70
  - internal/commands/merge.go:46-50
  - internal/commands/rotate.go:62-67
  - internal/commands/text.go:60-65
  - internal/commands/combine_images.go:45-49
  - internal/commands/watermark.go:64-69
- **Description:** Eight command files contain near-identical blocks for sanitizing output paths with conditional logic:
  ```go
  // Sanitize output path
  if output != "" && output != "-" {
      output, err = fileio.SanitizePath(output)
      if err != nil {
          return fmt.Errorf("invalid output path: %w", err)
      }
  }
  ```
  Some files check only `output != ""`, others check both `output != ""` and `output != "-"`.
- **Suggestion:** Extract to `internal/commands/helpers.go`:
  ```go
  // sanitizeOutputPath validates and sanitizes an output file path.
  // Returns the path unchanged if it's empty or "-" (stdin/stdout marker).
  func sanitizeOutputPath(output string) (string, error) {
      if output == "" || output == "-" {
          return output, nil
      }
      cleaned, err := fileio.SanitizePath(output)
      if err != nil {
          return "", fmt.Errorf("invalid output path: %w", err)
      }
      return cleaned, nil
  }
  ```
  Then replace all occurrences with: `output, err = sanitizeOutputPath(output)`
- **Impact:**
  - Lines saved: ~32 lines (4 lines saved per occurrence × 8 occurrences)
  - Clarity gained: Makes stdin/stdout special case handling consistent
  - Bug prevention: Eliminates inconsistent "-" handling across commands

## Medium Priority

### M1: GetPasswordSecure is Unnecessary Wrapper
- **Type:** Remove
- **Locations:**
  - internal/cli/flags.go:64-67 (definition)
  - 14 command files (all calls)
- **Description:** The `GetPasswordSecure` function is a simple one-line wrapper that just delegates to `ReadPassword`:
  ```go
  func GetPasswordSecure(cmd *cobra.Command, promptMsg string) (string, error) {
      return ReadPassword(cmd, promptMsg)
  }
  ```
  This adds no value. All callers could call `ReadPassword` directly.
- **Suggestion:**
  1. Delete `GetPasswordSecure` from flags.go
  2. Replace all calls in command files:
     - FROM: `password, err := cli.GetPasswordSecure(cmd, "Enter PDF password: ")`
     - TO: `password, err := cli.ReadPassword(cmd, "Enter PDF password: ")`
  3. Update imports if needed (ReadPassword should already be exported)
- **Impact:**
  - Lines removed: ~15 lines (function definition + simplicity)
  - Clarity gained: One less layer of indirection
  - API simplification: Fewer public functions in cli package
- **Counterpoint:** If `GetPasswordSecure` was intended as a stable public API while `ReadPassword` is an implementation detail, keep it. However, both are in `internal/cli`, so there's no external API concern.

### M2: HasChecksum Function is Rarely Needed
- **Type:** Refactor (optional)
- **Locations:**
  - internal/ocr/checksums.go:18-22 (definition)
  - internal/ocr/ocr_test.go:176-182 (used only in tests)
  - NOT used in production code
- **Description:** The `HasChecksum` function provides a `bool` check for checksum existence:
  ```go
  func HasChecksum(lang string) bool {
      _, ok := KnownChecksums[lang]
      return ok
  }
  ```
  However, production code only uses `GetChecksum`, which returns empty string for unknown languages. The `HasChecksum` function appears to be unused in production code, only in tests.
- **Suggestion:** Consider these options:
  1. **Keep as is** if it improves test readability (acceptable)
  2. **Remove** and replace test calls with `GetChecksum(lang) != ""`
  3. **Document** as test-only helper if keeping
- **Impact:**
  - Lines saved: ~5 lines (if removed)
  - Minimal impact: Only used in tests, so low risk either way
- **Recommendation:** Keep it. The clarity in tests (`HasChecksum("eng")` vs `GetChecksum("eng") != ""`) is valuable, and 5 lines is a small cost.

## Low Priority

### L1: Repeated Password Error Message Pattern
- **Type:** Note (stylistic)
- **Locations:**
  - internal/commands/decrypt.go:55-57
  - internal/commands/encrypt.go:56-58
  - Both: `return fmt.Errorf("password is required (use --password-file, PDF_CLI_PASSWORD env var, or interactive prompt)")`
- **Description:** Two commands (decrypt and encrypt) have identical error messages for missing passwords. This is only 2 occurrences, below the Rule of Three threshold.
- **Suggestion:** Extract to constant if more commands add this check in future:
  ```go
  const errPasswordRequired = "password is required (use --password-file, PDF_CLI_PASSWORD env var, or interactive prompt)"
  ```
- **Impact:** Minimal at 2 occurrences. Note for future consideration if count increases.

## Summary

- **Duplication found:** 17 instances of repeated patterns across 17 files
  - High-impact duplication: 2 patterns (input sanitization: 9 files, output sanitization: 8 files)
  - Low-impact duplication: 1 pattern (password error: 2 files)
- **Dead code found:** 0 unused definitions (all new code is called)
- **Unnecessary abstraction:** 1 instance (GetPasswordSecure wrapper)
- **Complexity hotspots:** 0 functions exceed thresholds
- **AI bloat patterns:** 0 instances (code is clean and purposeful)
- **Estimated cleanup impact:**
  - Lines removable: ~83 lines total
    - 36 lines from input sanitization consolidation
    - 32 lines from output sanitization consolidation
    - 15 lines from GetPasswordSecure removal
  - Abstractions eliminable: 1 (GetPasswordSecure wrapper)

## Code Quality Assessment

**Positive findings:**
- The new `ReadPassword` function in password.go is well-designed with clear priority ordering
- The `SanitizePath` implementation is thorough with good security checks
- Test coverage for new functionality is comprehensive
- No verbose error handling or defensive bloat detected
- Comments are appropriate and not excessive

**Architecture observation:**
The security hardening changes are conceptually sound but suffer from implementation-level duplication. The core abstractions (`ReadPassword`, `SanitizePath`) are good, but the command-level integration follows a copy-paste pattern instead of using helper functions. This is typical of multi-builder AI work where each builder adds the same pattern independently.

## Recommendation

**Action: Implement High Priority findings before shipping.**

The two high-priority consolidations (H1 and H2) are mechanical refactorings that will:
1. Reduce code size by ~68 lines
2. Improve maintainability by centralizing repeated logic
3. Eliminate inconsistencies in stdin/stdout handling
4. Make future security updates easier (one place to change)

The medium-priority items (M1 and M2) are acceptable to defer:
- M1 (GetPasswordSecure): Wrapper adds minimal cost, removing it is a nice-to-have
- M2 (HasChecksum): Improves test readability, keep it

**Estimated refactoring effort:** 30-45 minutes
- Add 2 helper functions to helpers.go
- Replace 17 call sites across command files
- Run tests to verify no behavior change
- Optionally remove GetPasswordSecure wrapper

This is a high-ROI refactoring that should be done before the phase is considered complete.
