# Phase 3 Research: Concurrency and Error Handling Fixes

## Overview
This document analyzes the codebase to prepare for implementing Phase 3 requirements focusing on goroutine context checking (R5), cleanup registry improvements (R7), debug logging for page extraction (R8), and password file validation (R9).

## R5: Goroutine ctx.Err() checks

### Current State Analysis

#### Text Extraction - `extractPagesParallel`
**Location**: `/Users/lgbarn/Personal/pdf-cli/internal/pdf/text.go:125-173`

**Current goroutine pattern**:
```go
for _, pageNum := range pages {
    if ctx.Err() != nil {
        // Context canceled, don't launch more work
        break
    }
    go func(pn int) {
        results <- pageResult{pageNum: pn, text: extractPageText(r, pn, totalPages)}
    }(pageNum)
}
```

**Problem**: The goroutine body (line 146) immediately calls `extractPageText` without checking if the context was cancelled. This means:
1. Goroutines launched before cancellation continue running
2. The expensive operation (`extractPageText`) executes even though the context is cancelled
3. Resources are wasted processing pages that will be discarded

**Expensive operation**: `extractPageText(r, pn, totalPages)` - calls `p.GetPlainText(nil)` which parses PDF page structure

**Fix location**: Line 146, before the `extractPageText` call

#### OCR Processing - `processImagesParallel`
**Location**: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr.go:472-530`

**Current goroutine pattern**:
```go
for i, imgPath := range imageFiles {
    if ctx.Err() != nil {
        // Context canceled, don't launch more work
        break
    }

    wg.Add(1)
    sem <- struct{}{} // Acquire semaphore
    go func(idx int, path string) {
        defer wg.Done()
        defer func() { <-sem }() // Release semaphore
        text, err := e.backend.ProcessImage(ctx, path, e.lang)
        results <- imageResult{index: idx, text: text, err: err}
    }(i, imgPath)
}
```

**Problem**: Similar issue - goroutine body (lines 498-503) immediately calls `e.backend.ProcessImage` without checking context status first.

**Expensive operation**: `e.backend.ProcessImage(ctx, path, e.lang)` - performs OCR on image, CPU-intensive

**Fix location**: Line 501, before the `ProcessImage` call

### Test Files
- `/Users/lgbarn/Personal/pdf-cli/internal/pdf/text_test.go` - Tests `extractPagesParallel` (lines 143-223)
- `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr_test.go` - Tests `processImagesParallel` (lines 212-248)
- `/Users/lgbarn/Personal/pdf-cli/internal/ocr/process_test.go` - More parallel processing tests (lines 112-224, 395-437)

**Test coverage**: Existing tests don't verify context cancellation behavior inside goroutine bodies. They test that parallel functions work but not that they respect context cancellation optimally.

### Implementation Strategy
For both functions, add `ctx.Err()` check at the start of goroutine body:

```go
go func(...) {
    defer ...
    if ctx.Err() != nil {
        return // Early exit before expensive operation
    }
    // expensive operation...
}(...)
```

---

## R7: Cleanup registry map conversion

### Current State Analysis

**Location**: `/Users/lgbarn/Personal/pdf-cli/internal/cleanup/cleanup.go:1-69`

**Current implementation**:
```go
var (
    mu     sync.Mutex
    paths  []string
    hasRun bool
)

func Register(path string) func() {
    mu.Lock()
    defer mu.Unlock()

    idx := len(paths)
    paths = append(paths, path)

    return func() {
        mu.Lock()
        defer mu.Unlock()
        if idx < len(paths) {
            paths[idx] = "" // mark as unregistered
        }
    }
}

func Run() error {
    mu.Lock()
    defer mu.Unlock()

    if hasRun {
        return nil
    }
    hasRun = true

    var firstErr error
    for i := len(paths) - 1; i >= 0; i-- {
        p := paths[i]
        if p == "" {
            continue
        }
        if err := os.RemoveAll(p); err != nil && firstErr == nil {
            firstErr = err
        }
    }
    paths = nil
    return firstErr
}
```

### Race Condition Analysis

**The Window**: Between Register and Unregister calls
1. Thread A: `Register("/tmp/foo")` → gets `idx = 0`, appends path
2. Thread B: `Register("/tmp/bar")` → gets `idx = 1`, appends path
3. Thread C: Calls `Run()` → sets `hasRun = true`, clears `paths = nil`
4. Thread A: Calls unregister function → checks `if idx < len(paths)` where `len(paths) = 0` (paths was nil'ed)
   - The check `idx < len(paths)` prevents panic, but the unregister is silently ignored
5. Now if `Reset()` is called and we try to use the registry again, state is inconsistent

**Issue**: The slice-index approach is fragile because:
- Indices become invalid after `Run()` clears the slice
- Unregister after Run() is a no-op (silently fails)
- Even with mutex protection, the semantic contract is broken

### Proposed Map-Based Approach

Replace slice with map tracking:
```go
var (
    mu     sync.Mutex
    paths  map[string]struct{} // set of paths to clean up
    hasRun bool
)
```

Benefits:
- No index invalidation issues
- Unregister is idempotent (deleting non-existent key is safe)
- Clearer semantic: "is this path tracked?" vs "what's at index N?"

### Callers of Register/Unregister

**Register callers** (found via grep):
1. `/Users/lgbarn/Personal/pdf-cli/internal/pdf/text.go:185` - `unregisterDir := cleanup.Register(tmpDir)`
2. `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr.go:236` - `unregisterTmp := cleanup.Register(tmpPath)`
3. `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr.go:350` - `unregisterDir := cleanup.Register(tmpDir)`

All follow pattern: `defer unregister()` immediately after registration.

### Test File
**Location**: `/Users/lgbarn/Personal/pdf-cli/internal/cleanup/cleanup_test.go:1-124`

**Existing tests**:
- `TestRegisterAndRun` - Basic registration and cleanup
- `TestUnregister` - Verify unregister prevents cleanup
- `TestConcurrentRegister` - 100 goroutines registering concurrently
- `TestRunIdempotent` - Verify Run() can be called multiple times

**Missing coverage**: No test for "unregister after Run()" scenario

---

## R8: Debug logging for page extraction errors

### Current State Analysis

**Location**: `/Users/lgbarn/Personal/pdf-cli/internal/pdf/text.go:109-123`

```go
// extractPageText extracts text from a single page, returning empty string on any error
func extractPageText(r *pdf.Reader, pageNum, totalPages int) string {
    if pageNum < 1 || pageNum > totalPages {
        return ""
    }
    p := r.Page(pageNum)
    if p.V.IsNull() {
        return ""
    }
    text, err := p.GetPlainText(nil)
    if err != nil {
        return ""
    }
    return text
}
```

**Error paths that silently return ""**:
1. Line 112: `pageNum < 1 || pageNum > totalPages` - out of bounds
2. Line 116: `p.V.IsNull()` - null page object
3. Line 120: `err != nil` from `GetPlainText()` - text extraction failure

**Impact**: From CONCERNS.md line 136:
> Users may not realize pages failed to extract; no indication in output or logs

### Logging API

**Location**: `/Users/lgbarn/Personal/pdf-cli/internal/logging/logger.go:160-163`

```go
// Debug logs at debug level.
func Debug(msg string, args ...any) {
    Get().Debug(msg, args...)
}
```

Uses structured logging via `slog`. Arguments are key-value pairs:
```go
logging.Debug("failed to extract page", "page", pageNum, "error", err)
```

**Import required**: `"github.com/lgbarn/pdf-cli/internal/logging"`

### Implementation Strategy

Add debug logging before each `return ""`:

```go
if pageNum < 1 || pageNum > totalPages {
    logging.Debug("page number out of range", "page", pageNum, "total", totalPages)
    return ""
}
// ...
if p.V.IsNull() {
    logging.Debug("page object is null", "page", pageNum)
    return ""
}
// ...
if err != nil {
    logging.Debug("failed to extract text from page", "page", pageNum, "error", err)
    return ""
}
```

### Verification

**Default log level**: From CONCERNS.md line 226:
> Log level defaults to "silent"

Debug logs will only appear when user runs with `--log-level debug`.

---

## R9: Password file printable character validation

### Current State Analysis

**Location**: `/Users/lgbarn/Personal/pdf-cli/internal/cli/password.go:19-84`

**Current password file reading** (lines 24-39):
```go
if passwordFile != "" {
    // Sanitize password file path against directory traversal
    for _, part := range strings.Split(passwordFile, "/") {
        if part == ".." {
            return "", fmt.Errorf("invalid password file path: contains directory traversal")
        }
    }
    passwordFile = filepath.Clean(passwordFile)
    data, err := os.ReadFile(passwordFile) // #nosec G304 -- path sanitized above
    if err != nil {
        return "", fmt.Errorf("failed to read password file: %w", err)
    }
    if len(data) > 1024 {
        return "", fmt.Errorf("password file exceeds 1KB size limit")
    }
    return strings.TrimSpace(string(data)), nil
}
```

**Current behavior**: Reads raw bytes, converts to string, trims whitespace - no content validation.

**Issue from CONCERNS.md** (line 108):
> Password file size limited to 1KB but no validation of content (e.g., binary data, non-printable chars)

### What are "Printable Characters"?

In Go's `unicode` package:
- `unicode.IsPrint(r rune)` - returns true for printable characters
  - Includes: letters, numbers, punctuation, spaces, symbols
  - Excludes: control characters (0x00-0x1F except space/tab), DEL (0x7F), most non-ASCII control codes

**Common printable ranges**:
- ASCII printable: 0x20-0x7E (space through tilde)
- Tabs/newlines: 0x09 (tab), 0x0A (LF), 0x0D (CR) - typically considered "acceptable" in passwords
- Unicode printable: Defined by Unicode character categories

### Context Decision (CONTEXT-3.md)

**Decision**: Warning only — print warning to stderr but still return the password content
**Rationale**: Avoids breaking users who legitimately use binary-looking passwords

### Implementation Strategy

**Validation logic**:
1. Iterate through runes in the file content
2. Count non-printable characters (excluding common whitespace: space, tab, \n, \r)
3. If non-printable characters detected, print warning to stderr
4. Return the password content regardless (warning only)

**Warning threshold**: Even a single non-printable character suggests wrong file

**Suggested implementation**:
```go
// After reading file
data, err := os.ReadFile(passwordFile)
if err != nil {
    return "", fmt.Errorf("failed to read password file: %w", err)
}
if len(data) > 1024 {
    return "", fmt.Errorf("password file exceeds 1KB size limit")
}

// Validate printable content
content := string(data)
nonPrintableCount := 0
for _, r := range content {
    // Allow common whitespace
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

**Import required**: `"unicode"`

### Test File

**Location**: `/Users/lgbarn/Personal/pdf-cli/internal/cli/password_test.go:1-203`

**Existing tests**:
- `TestReadPassword_PasswordFile` - Valid password file
- `TestReadPassword_PasswordFileTooLarge` - Size limit enforcement
- `TestReadPassword_PasswordFileMissing` - Missing file handling
- Priority tests (lines 96-132)

**Missing coverage**: No test for binary content detection

**New test needed**: `TestReadPassword_BinaryContentWarning`

---

## Summary of Changes Required

### R5: Goroutine ctx.Err() checks
- **Files to modify**:
  - `/Users/lgbarn/Personal/pdf-cli/internal/pdf/text.go` (line 146)
  - `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr.go` (line 501)
- **Change**: Add `if ctx.Err() != nil { return }` at start of goroutine body
- **Test impact**: Existing tests pass; new tests for context cancellation behavior recommended

### R7: Cleanup registry map conversion
- **Files to modify**:
  - `/Users/lgbarn/Personal/pdf-cli/internal/cleanup/cleanup.go` (lines 11-34, 38-59)
- **Change**: Convert `paths []string` to `paths map[string]struct{}`, update Register/Unregister/Run logic
- **Test impact**: Existing tests should continue to pass; add test for "unregister after Run()" scenario
- **Callers**: No changes needed - API remains identical

### R8: Debug logging for page extraction errors
- **Files to modify**:
  - `/Users/lgbarn/Personal/pdf-cli/internal/pdf/text.go` (lines 110-123)
- **Change**: Add `logging.Debug(...)` calls before each `return ""`
- **Imports**: Add `"github.com/lgbarn/pdf-cli/internal/logging"`
- **Test impact**: Existing tests pass; no behavior change unless `--log-level debug` is used

### R9: Password file printable character validation
- **Files to modify**:
  - `/Users/lgbarn/Personal/pdf-cli/internal/cli/password.go` (lines 31-38)
- **Change**: Add validation loop after reading file, print warning if non-printable chars found
- **Imports**: Add `"unicode"`
- **Test impact**: Existing tests pass; add new test for binary content warning

---

## Risk Assessment

| Requirement | Risk Level | Mitigation |
|------------|-----------|------------|
| R5 | Low | Context check is defensive - doesn't change correctness, only improves efficiency |
| R7 | Medium | Map-based tracking changes internal data structure - thorough testing needed |
| R8 | Low | Logging is append-only - doesn't change execution flow or return values |
| R9 | Low | Warning-only approach per CONTEXT-3.md - doesn't break existing workflows |

---

## Open Questions

1. **R5**: Should we add tests that verify context cancellation prevents expensive operations? (Recommended for completeness)

2. **R7**: Should `Run()` set `paths = nil` or `paths = make(map[string]struct{})` after cleanup? (nil is simpler, but empty map is more consistent)

3. **R8**: Should we also log at debug level in `extractPageText` when successful? (Probably not - would be very noisy)

4. **R9**: Should the warning specify which bytes are non-printable? (Could help debugging but may expose partial password content - avoid)

---

## References

- Primary source files analyzed:
  - `/Users/lgbarn/Personal/pdf-cli/internal/pdf/text.go`
  - `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr.go`
  - `/Users/lgbarn/Personal/pdf-cli/internal/cleanup/cleanup.go`
  - `/Users/lgbarn/Personal/pdf-cli/internal/cli/password.go`
  - `/Users/lgbarn/Personal/pdf-cli/internal/logging/logger.go`

- Test files analyzed:
  - `/Users/lgbarn/Personal/pdf-cli/internal/pdf/text_test.go`
  - `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr_test.go`
  - `/Users/lgbarn/Personal/pdf-cli/internal/ocr/process_test.go`
  - `/Users/lgbarn/Personal/pdf-cli/internal/cleanup/cleanup_test.go`
  - `/Users/lgbarn/Personal/pdf-cli/internal/cli/password_test.go`

- Context documents:
  - `.shipyard/phases/3/CONTEXT-3.md`
  - `.shipyard/codebase/CONCERNS.md` (lines 71-161)
