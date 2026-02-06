---
phase: code-quality-constants
plan: 1.1
wave: 1
dependencies: []
must_haves:
  - R13: Define output suffix constants in internal/commands/helpers.go
  - R13: Replace string literals with constants in 6 command files
  - R13: Update commands_test.go to use constants
files_touched:
  - internal/commands/helpers.go
  - internal/commands/encrypt.go
  - internal/commands/decrypt.go
  - internal/commands/compress.go
  - internal/commands/rotate.go
  - internal/commands/watermark.go
  - internal/commands/reorder.go
  - internal/commands/commands_test.go
tdd: false
---

# Plan 1.1: Output Suffix Constants (R13)

## Objective
Replace hardcoded output suffix string literals with named constants to improve maintainability and reduce typo risk.

## Tasks

<task id="1" files="internal/commands/helpers.go" tdd="false">
  <action>Define six public constants at the top of helpers.go: SuffixEncrypted = "_encrypted", SuffixDecrypted = "_decrypted", SuffixCompressed = "_compressed", SuffixRotated = "_rotated", SuffixWatermarked = "_watermarked", SuffixReordered = "_reordered". Add a comment block explaining these are default output filename suffixes.</action>
  <verify>grep -E "const Suffix(Encrypted|Decrypted|Compressed|Rotated|Watermarked|Reordered)" /Users/lgbarn/Personal/pdf-cli/internal/commands/helpers.go</verify>
  <done>All six constants are defined in helpers.go with correct values. Comment block is present explaining their purpose.</done>
</task>

<task id="2" files="internal/commands/encrypt.go, internal/commands/decrypt.go, internal/commands/compress.go, internal/commands/rotate.go, internal/commands/watermark.go, internal/commands/reorder.go" tdd="false">
  <action>Replace double-quoted string literals "_encrypted", "_decrypted", "_compressed", "_rotated", "_watermarked", "_reordered" with corresponding constants in: outputOrDefault() calls, validateBatchOutput() calls, and struct field initializations (e.g., DefaultSuffix fields). Do NOT change single-quoted literals in command Long description strings as these are user-facing documentation. Expected replacements: encrypt.go (2 occurrences in code), decrypt.go (2), compress.go (2), rotate.go (2), watermark.go (2), reorder.go (2).</action>
  <verify>cd /Users/lgbarn/Personal/pdf-cli && go build ./cmd/pdf && grep -E 'outputOrDefault.*"_(encrypted|decrypted|compressed|rotated|watermarked|reordered)"' internal/commands/*.go | wc -l | grep -q "^0$"</verify>
  <done>Build succeeds. No double-quoted suffix literals remain in function calls. Command descriptions still contain single-quoted literals for user documentation. All 12 programmatic occurrences replaced with constants.</done>
</task>

<task id="3" files="internal/commands/commands_test.go" tdd="false">
  <action>Replace string literals "_encrypted" and "_decrypted" in commands_test.go with SuffixEncrypted and SuffixDecrypted constants. Verify the test assertions still reference the correct expected values.</action>
  <verify>cd /Users/lgbarn/Personal/pdf-cli && go test -v ./internal/commands -run TestEncryptDecryptIntegration</verify>
  <done>Test passes. commands_test.go uses constants instead of literals. Test assertions correctly validate encrypted/decrypted output filenames.</done>
</task>

## Success Criteria
- All six suffix constants defined in helpers.go
- 14 total string literal replacements (12 in command files, 2 in test file)
- Command descriptions retain single-quoted literals for documentation
- files_test.go unchanged (cannot import commands package)
- All commands build successfully
- Integration test passes

## Verification
```bash
cd /Users/lgbarn/Personal/pdf-cli
go build ./cmd/pdf
go test ./internal/commands
```
