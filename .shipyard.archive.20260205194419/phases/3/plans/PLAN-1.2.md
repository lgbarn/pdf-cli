---
phase: security-hardening
plan: 02
wave: 1
dependencies: []
must_haves:
  - Centralized path sanitization function in internal/fileio
  - Reject paths containing ".." after cleaning
  - Path validation at all file input entry points
  - Prevent directory traversal attacks
files_touched:
  - internal/fileio/files.go
  - internal/fileio/files_test.go
  - internal/commands/*.go (all command entry points)
  - internal/ocr/ocr.go
tdd: true
---

# Plan 1.2: Path Sanitization (R3)

## Goal
Implement centralized, consistent path traversal sanitization across all file input entry points. This prevents directory traversal attacks where malicious input like `../../etc/passwd` could access files outside intended directories.

## Context
Currently, path handling is inconsistent across 18+ files. Some use filepath.Clean (e.g., fileio/files.go lines 87-88), others trust paths implicitly. We need a single SanitizePath function in internal/fileio that all entry points must call before accepting user-provided file paths.

Path traversal is a OWASP Top 10 vulnerability. Even with filepath.Clean, paths like `/tmp/../../etc/passwd` clean to `/etc/passwd`, which may be unintended. We must reject any path where filepath.Clean still contains `..` components.

## Tasks

<task id="1" files="internal/fileio/files.go,internal/fileio/files_test.go" tdd="true">
  <action>
Add SanitizePath function to internal/fileio/files.go with comprehensive tests.

Add after IsImageFile function (after line 217):

```go
// SanitizePath cleans a file path and validates it against directory traversal attacks.
// It returns an error if the cleaned path still contains ".." components.
// This prevents attacks like "../../etc/passwd" from accessing unintended files.
//
// Special cases:
// - stdin marker "-" is always allowed and returned as-is
// - Absolute paths are allowed after validation
// - Relative paths are allowed if they don't escape current directory
//
// Returns the cleaned path and nil error if valid, or error if path is unsafe.
func SanitizePath(path string) (string, error) {
  // Allow stdin marker
  if path == "-" {
    return path, nil
  }

  // Clean the path (resolves . and .. components, removes duplicates)
  cleaned := filepath.Clean(path)

  // Check if cleaned path still contains .. (indicates traversal attempt)
  parts := strings.Split(cleaned, string(filepath.Separator))
  for _, part := range parts {
    if part == ".." {
      return "", fmt.Errorf("path contains directory traversal: %s", path)
    }
  }

  return cleaned, nil
}

// SanitizePaths validates multiple paths and returns cleaned versions.
// If any path is invalid, returns error immediately.
func SanitizePaths(paths []string) ([]string, error) {
  cleaned := make([]string, len(paths))
  for i, path := range paths {
    clean, err := SanitizePath(path)
    if err != nil {
      return nil, err
    }
    cleaned[i] = clean
  }
  return cleaned, nil
}
```

Create internal/fileio/files_test.go or add to existing test file:

```go
func TestSanitizePath(t *testing.T) {
  tests := []struct {
    name    string
    input   string
    want    string
    wantErr bool
  }{
    // Valid paths
    {name: "simple file", input: "file.pdf", want: "file.pdf", wantErr: false},
    {name: "subdirectory", input: "docs/file.pdf", want: "docs/file.pdf", wantErr: false},
    {name: "absolute path", input: "/tmp/file.pdf", want: "/tmp/file.pdf", wantErr: false},
    {name: "stdin marker", input: "-", want: "-", wantErr: false},
    {name: "current dir", input: "./file.pdf", want: "file.pdf", wantErr: false},
    {name: "redundant slashes", input: "docs//file.pdf", want: "docs/file.pdf", wantErr: false},

    // Invalid paths (directory traversal)
    {name: "parent traversal", input: "../file.pdf", want: "", wantErr: true},
    {name: "deep traversal", input: "../../etc/passwd", want: "", wantErr: true},
    {name: "traversal in middle", input: "docs/../../etc/passwd", want: "", wantErr: true},
    {name: "absolute traversal", input: "/tmp/../../etc/passwd", want: "", wantErr: true},
    {name: "mixed traversal", input: "./docs/../../../etc/passwd", want: "", wantErr: true},
  }

  for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
      got, err := SanitizePath(tt.input)
      if (err != nil) != tt.wantErr {
        t.Errorf("SanitizePath() error = %v, wantErr %v", err, tt.wantErr)
        return
      }
      if got != tt.want {
        t.Errorf("SanitizePath() = %v, want %v", got, tt.want)
      }
    })
  }
}

func TestSanitizePaths(t *testing.T) {
  // Test valid batch
  valid := []string{"file1.pdf", "docs/file2.pdf", "/tmp/file3.pdf"}
  got, err := SanitizePaths(valid)
  if err != nil {
    t.Errorf("SanitizePaths() unexpected error: %v", err)
  }
  if len(got) != 3 {
    t.Errorf("SanitizePaths() returned %d paths, want 3", len(got))
  }

  // Test batch with one invalid
  invalid := []string{"file1.pdf", "../etc/passwd", "file3.pdf"}
  _, err = SanitizePaths(invalid)
  if err == nil {
    t.Error("SanitizePaths() expected error for traversal, got nil")
  }
}
```
  </action>
  <verify>
Run tests first (TDD - should fail):
```bash
cd /Users/lgbarn/Personal/pdf-cli/.worktrees/phase-2-concurrency
go test -v ./internal/fileio -run TestSanitizePath
```

Implement function, then verify tests pass:
```bash
go test -v ./internal/fileio -run TestSanitizePath
```

Check coverage:
```bash
go test -cover ./internal/fileio
```
  </verify>
  <done>
- SanitizePath function implemented in internal/fileio/files.go
- SanitizePaths batch function implemented
- Comprehensive test coverage including edge cases
- All tests pass with >90% coverage
- Correctly rejects all directory traversal patterns
  </done>
</task>

<task id="2" files="internal/commands/*.go" tdd="false">
  <action>
Update all command entry points to sanitize file paths before processing.

For each command file that accepts file path arguments (18+ files), add path sanitization at the beginning of the RunE function, right after args are received and before any file operations.

Pattern to apply:

```go
func runCommandName(cmd *cobra.Command, args []string) error {
  // Sanitize input paths first (before any other operations)
  sanitizedArgs, err := fileio.SanitizePaths(args)
  if err != nil {
    return fmt.Errorf("invalid file path: %w", err)
  }

  // Continue with sanitized paths
  // Replace all uses of args with sanitizedArgs below this point
  ...
}
```

Specific files to update:
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

Special handling for single-file commands that also accept stdin ("-"):
The SanitizePath function already handles "-" as a special case, so no extra logic needed.

Also sanitize output paths from flags:
```go
output := cli.GetOutput(cmd)
if output != "" {
  output, err = fileio.SanitizePath(output)
  if err != nil {
    return fmt.Errorf("invalid output path: %w", err)
  }
}
```
  </action>
  <verify>
Build succeeds:
```bash
cd /Users/lgbarn/Personal/pdf-cli/.worktrees/phase-2-concurrency
go build -o /tmp/pdf-cli ./cmd/pdf
```

Test path traversal rejection:
```bash
# Should fail with sanitization error
/tmp/pdf-cli info "../../../etc/passwd" 2>&1 | grep "directory traversal"

# Should succeed with normal path
/tmp/pdf-cli info testdata/sample.pdf
```

Run integration tests:
```bash
go test -v ./internal/commands -run Integration
```
  </verify>
  <done>
- All 14+ command files updated with path sanitization
- Both input args and output flag paths are sanitized
- Directory traversal attempts are rejected with clear error
- Stdin marker "-" still works correctly
- Integration tests pass
  </done>
</task>

<task id="3" files="internal/ocr/ocr.go,internal/fileio/files.go" tdd="false">
  <action>
Add path sanitization to internal file operations and OCR data directory.

Update internal/fileio/files.go CopyFile function (lines 85-111):
Replace lines 87-88 with:
```go
// Sanitize paths to prevent directory traversal
cleanSrc, err := SanitizePath(src)
if err != nil {
  return fmt.Errorf("invalid source path: %w", err)
}
cleanDst, err := SanitizePath(dst)
if err != nil {
  return fmt.Errorf("invalid destination path: %w", err)
}
```

Update internal/ocr/ocr.go downloadTessdata function (line 169):
Add sanitization after filepath.Join on line 171:
```go
dataFile := filepath.Join(dataDir, lang+".traineddata")
// Validate path to prevent traversal if lang is malicious
dataFile, err := fileio.SanitizePath(dataFile)
if err != nil {
  return fmt.Errorf("invalid tessdata path for language %s: %w", lang, err)
}
```

Update internal/ocr/ocr.go getDataDir function (line 123):
Add validation after filepath.Join:
```go
dataDir := filepath.Join(configDir, "pdf-cli", "tessdata")
// Validate constructed path
dataDir, err = fileio.SanitizePath(dataDir)
if err != nil {
  return "", fmt.Errorf("invalid data directory path: %w", err)
}
```

These prevent malicious language codes like `../../etc/passwd` from creating files outside intended directories.
  </action>
  <verify>
Build succeeds:
```bash
cd /Users/lgbarn/Personal/pdf-cli/.worktrees/phase-2-concurrency
go build ./internal/fileio ./internal/ocr
```

Test malicious language code rejection:
```bash
cd /Users/lgbarn/Personal/pdf-cli/.worktrees/phase-2-concurrency
go test -v ./internal/ocr -run TestDownloadTessdata
```

Run full test suite:
```bash
go test -v ./...
```
  </verify>
  <done>
- CopyFile function updated with path sanitization
- OCR downloadTessdata function sanitizes language-based paths
- OCR getDataDir validates constructed directory paths
- Malicious input rejected at all internal file operations
- All tests pass
  </done>
</task>

## Verification Strategy

Security testing checklist:
1. Test directory traversal in input files: `pdf info "../../etc/passwd"`
2. Test directory traversal in output: `pdf split file.pdf -o "../../tmp/out.pdf"`
3. Test traversal in OCR language: manual unit test with malicious lang code
4. Verify stdin marker "-" still works
5. Verify absolute paths still work
6. Verify relative paths within working dir work
7. Run gosec security scanner to confirm no new issues

Automated tests:
```bash
# Unit tests
go test -v ./internal/fileio -run TestSanitizePath

# Integration tests with malicious inputs
go test -v ./internal/commands -run Integration

# Security scanner
gosec ./...
```

Success criteria:
- All directory traversal attempts rejected with clear error message
- Legitimate paths (absolute, relative, stdin) still work
- No regressions in existing functionality
- gosec reports no new path traversal vulnerabilities
- 100% test coverage for SanitizePath function

## Breaking Changes

**None** - This is purely defensive. Legitimate file paths continue to work. Only malicious directory traversal attempts are rejected.

## Security Impact

**HIGH IMPACT**: Prevents directory traversal attacks across all file operations. Centralizes path validation in a single, well-tested function that can be audited. Applies defense-in-depth by validating at both entry points (commands) and internal operations (fileio, ocr).
