---
phase: phase-3-concurrency-error-handling
plan: 2.1
wave: 2
dependencies: []
must_haves:
  - R5: Goroutines check ctx.Err() before expensive operations
  - R8: Debug logging for page extraction errors
  - Context cancellation prevents wasted work in parallel processing
  - Silent errors in extractPageText are logged at debug level
files_touched:
  - internal/pdf/text.go
  - internal/ocr/ocr.go
tdd: false
---

# Plan 2.1: Goroutine Context Checks and Debug Logging (R5 + R8)

## Overview
Add context cancellation checks inside goroutine bodies to prevent expensive operations from running after cancellation (R5), and add debug logging for silent error paths in text extraction (R8). These requirements touch the same files and are logically related (both improve observability and efficiency).

## Tasks

<task id="1" files="internal/pdf/text.go" tdd="false">
  <action>
Add context check inside extractPagesParallel goroutine and debug logging to extractPageText:

### Part A: Context Check in Goroutine (R5)
Location: Line 146 in extractPagesParallel function

Current code:
```go
go func(pn int) {
    results <- pageResult{pageNum: pn, text: extractPageText(r, pn, totalPages)}
}(pageNum)
```

Change to:
```go
go func(pn int) {
    if ctx.Err() != nil {
        results <- pageResult{pageNum: pn, text: ""}
        return
    }
    results <- pageResult{pageNum: pn, text: extractPageText(r, pn, totalPages)}
}(pageNum)
```

This prevents the expensive `extractPageText` operation from running if context is already cancelled.

### Part B: Debug Logging in extractPageText (R8)
Location: Lines 110-123 in extractPageText function

Add import at top of file (around line 4):
```go
"github.com/lgbarn/pdf-cli/internal/logging"
```

Update extractPageText function:
```go
// extractPageText extracts text from a single page, returning empty string on any error
func extractPageText(r *pdf.Reader, pageNum, totalPages int) string {
    if pageNum < 1 || pageNum > totalPages {
        logging.Debug("page number out of range", "page", pageNum, "total", totalPages)
        return ""
    }
    p := r.Page(pageNum)
    if p.V.IsNull() {
        logging.Debug("page object is null", "page", pageNum)
        return ""
    }
    text, err := p.GetPlainText(nil)
    if err != nil {
        logging.Debug("failed to extract text from page", "page", pageNum, "error", err)
        return ""
    }
    return text
}
```

This logs all silent error paths at debug level, helping users diagnose extraction issues when running with `--log-level debug`.
  </action>
  <verify>go test -v /Users/lgbarn/Personal/pdf-cli/internal/pdf/... && go test -race /Users/lgbarn/Personal/pdf-cli/internal/pdf/...</verify>
  <done>All pdf package tests pass including text extraction tests. Context check in goroutine prevents expensive operations after cancellation. Debug logging added to all three error paths in extractPageText (out of range, null page, extraction error). Race detector shows no issues.</done>
</task>

<task id="2" files="internal/ocr/ocr.go" tdd="false">
  <action>
Add context check inside processImagesParallel goroutine before expensive OCR operation:

Location: Line 498-503 in processImagesParallel function

Current code:
```go
go func(idx int, path string) {
    defer wg.Done()
    defer func() { <-sem }() // Release semaphore
    text, err := e.backend.ProcessImage(ctx, path, e.lang)
    results <- imageResult{index: idx, text: text, err: err}
}(i, imgPath)
```

Change to:
```go
go func(idx int, path string) {
    defer wg.Done()
    defer func() { <-sem }() // Release semaphore

    if ctx.Err() != nil {
        results <- imageResult{index: idx, text: "", err: ctx.Err()}
        return
    }

    text, err := e.backend.ProcessImage(ctx, path, e.lang)
    results <- imageResult{index: idx, text: text, err: err}
}(i, imgPath)
```

Add the context check immediately after defer statements (around line 500) and before the expensive `ProcessImage` call. This ensures:
1. Goroutines launched before cancellation don't waste resources on OCR
2. Context error is properly propagated in the result
3. Semaphore is still released (handled by defer)
  </action>
  <verify>go test -v /Users/lgbarn/Personal/pdf-cli/internal/ocr/... && go test -race /Users/lgbarn/Personal/pdf-cli/internal/ocr/...</verify>
  <done>All ocr package tests pass. Context check in goroutine prevents expensive OCR operations after cancellation. Semaphore is correctly released via defer regardless of early return. Race detector shows no issues.</done>
</task>

<task id="3" files="internal/pdf/text.go, internal/ocr/ocr.go" tdd="false">
  <action>
Verify goroutine context checks and debug logging implementation across all affected packages:

Run comprehensive test suite:
1. Test pdf package with race detector
2. Test ocr package with race detector
3. Run integration tests if available
4. Verify debug logging can be enabled with --log-level debug (manual verification if needed)

Verification points:
- Context cancellation prevents expensive operations in goroutines
- Debug logging appears in stderr when log level is set to debug
- No race conditions in context checking
- Existing parallel processing tests still pass
- Performance is not degraded (context check is O(1))

Success indicators:
- All tests pass with race detector
- grep for "ctx.Err() != nil" shows checks in both extractPagesParallel and processImagesParallel
- grep for "logging.Debug" shows calls in extractPageText error paths
- No "DATA RACE" warnings in test output
  </action>
  <verify>go test -race /Users/lgbarn/Personal/pdf-cli/internal/pdf/... /Users/lgbarn/Personal/pdf-cli/internal/ocr/... 2>&1 | tee /tmp/context-race-test.log && ! grep -i "DATA RACE" /tmp/context-race-test.log && grep -l "ctx.Err()" /Users/lgbarn/Personal/pdf-cli/internal/pdf/text.go /Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr.go</verify>
  <done>All race detector tests pass with zero DATA RACE warnings. Context checks are present in both parallel processing functions (extractPagesParallel and processImagesParallel). Debug logging is present in all three error paths of extractPageText. Goroutines now respect context cancellation before expensive operations.</done>
</task>

## Success Criteria

### R5: Goroutine Context Checks
- ✓ extractPagesParallel (internal/pdf/text.go:146) checks ctx.Err() before extractPageText
- ✓ processImagesParallel (internal/ocr/ocr.go:501) checks ctx.Err() before ProcessImage
- ✓ Early return prevents expensive operations after context cancellation
- ✓ Race detector shows no issues with context checking

### R8: Debug Logging for Page Extraction
- ✓ extractPageText logs at debug level for out-of-range page numbers
- ✓ extractPageText logs at debug level for null page objects
- ✓ extractPageText logs at debug level for GetPlainText errors
- ✓ logging.Debug calls include structured context (page number, error details)
- ✓ Logs appear in stderr when --log-level debug is set (default is silent)

### General
- ✓ All existing tests pass (pdf and ocr packages)
- ✓ Race detector shows zero warnings
- ✓ No behavior changes unless log level is debug or context is cancelled
- ✓ Performance impact negligible (ctx.Err() is O(1) atomic check)

## Notes
- Context checks are defensive programming: they don't change correctness, only improve efficiency
- Debug logging helps diagnose extraction failures without changing behavior
- Both changes are backward compatible (no API changes)
- The context check prevents wasted CPU on goroutines launched before cancellation
- Debug logs only appear with `--log-level debug`, so no noise in production use
