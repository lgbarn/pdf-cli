# Phase 4: Error Handling and Reliability - Research

**Date:** 2026-01-31
**Phase:** 4
**Researcher:** Domain Researcher
**Status:** Complete

---

## Executive Summary

Phase 4 addresses three critical reliability issues in pdf-cli:

1. **Silent error swallowing in parallel image processing** - Errors from individual OCR operations are discarded instead of being reported
2. **Unchecked file close errors on write paths** - Deferred `Close()` calls ignore errors, risking data loss
3. **No cleanup mechanism for temp files on crash/interrupt** - Temp directories persist when the process receives SIGINT/SIGTERM

The research shows these issues are widespread but well-understood in the Go ecosystem. Standard library solutions exist for all three (errors.Join, named returns with defer closures, and signal.NotifyContext). Implementation risk is low-to-medium, with the main challenge being comprehensive test coverage of error paths.

**Key Finding:** The codebase already uses `signal.NotifyContext` in main.go (from Phase 2), which provides a foundation for cleanup integration. The project uses Go 1.25, so `errors.Join` (available since 1.20) is available.

---

## 1. Technology Options

### 1.1 R6: Parallel Error Collection

#### Option A: errors.Join (Recommended)
- **Description:** Standard library function (since Go 1.20) that combines multiple errors into a single error
- **Maturity:** Stable, part of stdlib since Go 1.20
- **Pros:**
  - Zero dependencies
  - Standard Go idiom as of 2026
  - Works with errors.Is and errors.As for error inspection
  - Discards nil errors automatically
- **Cons:**
  - Returns all errors concatenated (may be verbose)
  - No custom formatting options
- **Example:**
  ```go
  var errs []error
  for _, item := range items {
      if err := process(item); err != nil {
          errs = append(errs, err)
      }
  }
  return errors.Join(errs...)
  ```

#### Option B: hashicorp/go-multierror
- **Description:** Third-party library for collecting multiple errors
- **Maturity:** Mature (2014+), well-maintained by HashiCorp
- **Pros:**
  - Custom formatting (bullet points)
  - Can append errors to existing multierror
  - Well-tested in production
- **Cons:**
  - External dependency (violates project's minimal dependency philosophy)
  - errors.Join is now standard approach
  - Unnecessary for this use case

#### Option C: uber-go/multierr
- **Description:** Uber's multi-error library
- **Maturity:** Mature, widely used
- **Pros:**
  - Allows combining errors with existing error
  - Advanced features for error composition
- **Cons:**
  - External dependency
  - Overkill for simple error collection
  - errors.Join suffices for this project

### 1.2 R8: File Close Error Propagation

#### Option A: Named Return + Defer Closure (Recommended)
- **Description:** Use named return value with deferred function that checks close error
- **Maturity:** Idiomatic Go pattern, documented extensively
- **Pros:**
  - Zero abstraction overhead
  - Preserves original error if one exists
  - Compiler-checked via named returns
  - Standard practice for write operations
- **Cons:**
  - Requires named return values
  - Slightly more verbose than ignoring errors
- **Example:**
  ```go
  func writeData(path string, data []byte) (err error) {
      f, err := os.Create(path)
      if err != nil {
          return err
      }
      defer func() {
          if cerr := f.Close(); cerr != nil && err == nil {
              err = cerr
          }
      }()
      _, err = f.Write(data)
      return err
  }
  ```

#### Option B: Sync() + Ignored Close
- **Description:** Call `f.Sync()` explicitly, then defer close without checking error
- **Maturity:** Common pattern for write-heavy operations
- **Pros:**
  - Ensures data is flushed to disk
  - Simpler than named returns
  - Close errors after Sync are usually not actionable
- **Cons:**
  - Still ignores close errors (metadata updates, etc.)
  - Not suitable for all write patterns
  - May mask resource leaks

#### Option C: errcheck Linter Exception (Current Approach)
- **Description:** Continue ignoring close errors via .golangci.yaml exclusion
- **Maturity:** Used in many projects
- **Pros:**
  - No code changes needed
  - Simpler code
- **Cons:**
  - Violates requirement R8
  - Risk of data loss on write paths
  - Not considered best practice

### 1.3 R11: Temp File Cleanup on Signals

#### Option A: Global Cleanup Registry with Defer (Recommended)
- **Description:** Create `internal/cleanup` package with global registry that tracks temp paths and removes them on exit
- **Maturity:** Common pattern in CLI tools
- **Pros:**
  - Centralized cleanup logic
  - Works with existing signal.NotifyContext
  - Simple API: Register(path), defer cleanup.Run()
  - Thread-safe with mutex
- **Cons:**
  - Global state (acceptable for CLI)
  - Must remember to register temp paths
- **Example:**
  ```go
  // In internal/cleanup/cleanup.go
  var (
      mu    sync.Mutex
      paths []string
  )

  func Register(path string) {
      mu.Lock()
      paths = append(paths, path)
      mu.Unlock()
  }

  func Run() {
      mu.Lock()
      defer mu.Unlock()
      for _, p := range paths {
          _ = os.RemoveAll(p)
      }
      paths = nil
  }
  ```

#### Option B: Context-Based Cleanup (context.AfterFunc)
- **Description:** Use context.AfterFunc (Go 1.21+) to register cleanup on context cancellation
- **Maturity:** Newer pattern (Go 1.21+)
- **Pros:**
  - Tied to context lifecycle
  - No global state
  - Automatic cleanup on context cancellation
- **Cons:**
  - Requires passing context to cleanup registration points
  - More complex to retrofit into existing code
  - Cleanup may run too early if context is canceled before operation completes

#### Option C: Per-Function Defer (Current Approach)
- **Description:** Continue using `defer os.RemoveAll(tmpDir)` in each function
- **Maturity:** Standard Go pattern
- **Pros:**
  - Simple, localized
  - No additional abstraction
- **Cons:**
  - Doesn't run on SIGINT/SIGTERM (process killed before defer executes)
  - Violates requirement R11
  - Leaves temp files on crash/interrupt

---

## 2. Recommended Approach

### 2.1 R6: Use errors.Join for Parallel Error Collection

**Decision:** Adopt `errors.Join` from the standard library.

**Rationale:**
- Project uses Go 1.25, errors.Join is available (since 1.20)
- Zero dependencies aligns with project philosophy
- Standard idiom as of 2026
- Already used in `internal/commands/helpers.go:73` for batch processing
- Integrates seamlessly with existing error handling

**Why alternatives were not chosen:**
- Third-party libraries (go-multierror, multierr) add unnecessary dependencies
- Custom error collection would reinvent the wheel
- errors.Join is sufficient for reporting multiple OCR failures

### 2.2 R8: Use Named Returns + Defer Closure

**Decision:** Implement named return values with deferred close error checking for write operations.

**Rationale:**
- Industry best practice for write operations ([Don't defer Close() on writable files](https://www.joeshaw.org/dont-defer-close-on-writable-files/))
- Prevents data loss from buffered writes
- Compiler helps via named return syntax
- Existing precedent in `internal/fileio/files.go` (AtomicWrite uses explicit close checking)

**Why alternatives were not chosen:**
- Sync() + ignored close still ignores potential errors (metadata, resource cleanup)
- Current approach (ignored errors) violates requirement R8 and best practices
- Named returns provide best balance of safety and clarity

### 2.3 R11: Global Cleanup Registry

**Decision:** Create `internal/cleanup` package with thread-safe global registry.

**Rationale:**
- Works naturally with existing `signal.NotifyContext` from Phase 2
- Simple API fits CLI usage pattern
- Thread-safe design prevents race conditions
- Cleanup runs on normal exit AND signal interruption
- Minimal refactoring required (just add Register calls where temp paths are created)

**Why alternatives were not chosen:**
- Context-based cleanup requires extensive refactoring to thread context through all temp file creation points
- Per-function defer doesn't handle signals (violates R11)
- Global registry is acceptable for single-process CLI tools

---

## 3. Potential Risks and Mitigations

### Risk 1: Breaking Existing Error Handling Tests

**Description:** Changing processImagesParallel to return all errors instead of first error may break tests that expect specific error messages.

**Likelihood:** Medium
**Impact:** Low (test-only)

**Mitigation:**
- Review all OCR-related tests before implementation
- Update tests to use errors.Is or check for substring matches
- Use go test -v to verify error message formats
- Consider adding helper function to extract first error from joined errors if needed for compatibility

### Risk 2: Performance Impact of Named Returns

**Description:** Named return values with defer closures add minimal overhead.

**Likelihood:** Low
**Impact:** Negligible

**Mitigation:**
- Benchmark if concerned (unlikely to be measurable)
- Only apply to write operations (not read-only file opens)
- Pattern is already used in fileio.AtomicWrite with no observed issues

### Risk 3: Cleanup Registry Race Conditions

**Description:** Concurrent registration and cleanup could cause data races.

**Likelihood:** Low
**Impact:** Medium (crash or incomplete cleanup)

**Mitigation:**
- Use sync.Mutex in cleanup package
- Run go test -race ./... to verify thread safety
- Document that cleanup.Run() should only be called once (from main)
- Consider using sync.Once for Run() to ensure single execution

### Risk 4: Double-Free of Temp Files

**Description:** Both local defer and cleanup registry might try to remove the same temp file.

**Likelihood:** Medium
**Impact:** Low (os.RemoveAll is idempotent)

**Mitigation:**
- Document that os.RemoveAll ignores "not found" errors
- Keep local defers for normal path cleanup
- Cleanup registry acts as safety net for interrupt path
- Alternative: Remove local defer for registered paths (not recommended - complicates normal flow)

### Risk 5: Coverage Drop from Additional Error Paths

**Description:** New error handling paths may not be fully covered by tests.

**Likelihood:** Medium
**Impact:** Medium (coverage requirement is >=81%)

**Mitigation:**
- Add tests for close errors (mock file that fails on close)
- Add tests for signal handling (send SIGTERM to test process)
- Add tests for errors.Join with multiple errors
- Run coverage report after changes: `go test -cover ./...`
- Current coverage is 75-100% across packages, so there's headroom

---

## 4. Relevant Documentation Links

### errors.Join
- [Go Package Documentation - errors.Join](https://pkg.go.dev/errors#Join)
- [Joining Errors in Golang - GeeksforGeeks](https://www.geeksforgeeks.org/go-language/joining-errors-in-golang/)
- [New in Go 1.20: wrapping multiple errors](https://lukas.zapletalovi.com/posts/2022/wrapping-multiple-errors/)
- [Wrapping Multiple Errors in Golang](https://www.tiredsg.dev/blog/wrapping-multiple-errors-golang/)

### Defer and Close Error Handling
- [Don't defer Close() on writable files - Joe Shaw](https://www.joeshaw.org/dont-defer-close-on-writable-files/)
- [Handling Errors from Deferred Functions in Go - Thomas Stringer](https://trstringer.com/golang-deferred-function-error-handling/)
- [Understanding Go's defer: Usage, Evaluation, and Error Handling](https://claudiuconstantinbogdan.me/articles/go-defer)
- [Using defer in Go: Best Practices and Common Use Cases](https://dev.to/zakariachahboun/common-use-cases-for-defer-in-go-1071)

### Signal Handling
- [Go Package Documentation - os/signal](https://pkg.go.dev/os/signal)
- [Signal Handling in Go Applications - Medium](https://medium.com/@AlexanderObregon/signal-handling-in-go-applications-b96eb61ecb69)
- [Graceful Shutdown in Go: Practical Patterns](https://victoriametrics.com/blog/go-graceful-shutdown/)
- [Go by Example: Signals](https://gobyexample.com/signals)
- [GO capture ctrl+c signal & run cleanup function](https://www.golinuxcloud.com/go-capture-ctrl-c/)

### Go Error Handling Best Practices
- [A practical guide to error handling in Go - Datadog](https://www.datadoghq.com/blog/go-error-handling/)
- [Popular Error Handling Techniques in Go - JetBrains Guide](https://www.jetbrains.com/guide/go/tutorials/handle_errors_in_go/error_technique/)

---

## 5. Implementation Considerations

### 5.1 Integration Points

#### 5.1.1 R6: Parallel Error Collection

**Primary Files:**
- `internal/ocr/ocr.go` (lines 334-380: processImagesParallel, lines 309-332: processImagesSequential)

**Current Behavior:**
```go
// processImagesParallel (line 359)
text, _ := e.backend.ProcessImage(ctx, path, e.lang)  // Error discarded!
results <- imageResult{index: idx, text: text}

// processImagesSequential (line 322-325)
text, err := e.backend.ProcessImage(ctx, imgPath, e.lang)
if err == nil {  // Only includes text if no error
    texts = append(texts, text)
}
// Error is silently discarded
```

**Issues Found:**
1. `processImagesParallel` (line 359): Errors from `ProcessImage` are completely ignored with `_` discard
2. `processImagesSequential` (lines 322-325): Errors are checked but not propagated; failed pages are silently skipped
3. `imageResult` struct (lines 293-296) only carries `text`, not error

**Required Changes:**
1. Add `err error` field to `imageResult` struct
2. Modify processImagesParallel to capture errors in results channel
3. Collect all errors and use `errors.Join` to return them
4. Modify processImagesSequential to collect errors in slice and join
5. Update callers to handle possibility of partial results with errors

**Example Fix for processImagesParallel:**
```go
type imageResult struct {
    index int
    text  string
    err   error  // NEW
}

func (e *Engine) processImagesParallel(...) (string, error) {
    // ... existing setup ...

    go func(idx int, path string) {
        defer wg.Done()
        defer func() { <-sem }()
        text, err := e.backend.ProcessImage(ctx, path, e.lang)
        results <- imageResult{index: idx, text: text, err: err}  // Capture error
    }(i, imgPath)

    // ... collect results ...

    var errs []error
    texts := make([]string, len(imageFiles))
    for res := range results {
        texts[res.index] = res.text
        if res.err != nil {
            errs = append(errs, fmt.Errorf("image %d: %w", res.index, res.err))
        }
        // ...
    }

    text := joinNonEmpty(texts, "\n")
    if len(errs) > 0 {
        return text, errors.Join(errs...)  // Return partial results + errors
    }
    return text, nil
}
```

**Testing Strategy:**
- Mock backend that fails on specific images
- Verify all errors are returned via errors.Join
- Test partial success (some images succeed, some fail)
- Verify error messages include image index/path

#### 5.1.2 R8: File Close Error Propagation

**Files with Write Operations:**

Based on `defer.*Close()` grep analysis, the following files have write paths:

1. **internal/fileio/files.go** (lines 94, 104):
   - `CopyFile`: Both srcFile and dstFile closes (line 104 is write operation)
   - Already has good pattern for dstFile (line 110: `return dstFile.Sync()`)
   - **Action:** Add named return and check close error after Sync

2. **internal/fileio/stdio.go** (line 63):
   - `WriteToStdout`: Reading file and writing to stdout
   - **Action:** Read-only, no change needed

3. **internal/commands/patterns/stdio.go** (line 43):
   - `Setup`: Creates temp file for output
   - **Action:** File is closed immediately, no defer needed (line 43: `_ = tmpFile.Close()`)
   - This is acceptable - file is empty placeholder

4. **internal/ocr/ocr.go** (lines 187, 202, 205):
   - Line 187: `defer resp.Body.Close()` - HTTP response body (read-only)
   - Line 202: `_ = tmpFile.Close()` after io.Copy (WRITE operation)
   - Line 205: `_ = tmpFile.Close()` in error path
   - **Action:** Line 202 needs error checking (downloading tessdata file)

5. **internal/ocr/native.go** (line 54):
   - `ProcessImage`: Creates temp file for tesseract output
   - **Action:** File is immediately closed (empty placeholder for tesseract), acceptable

6. **internal/ocr/wasm.go** (lines 82, 106):
   - Line 82: `defer tessDataFile.Close()` - read-only
   - Line 106: `defer imgFile.Close()` - read-only

7. **internal/pdf/text.go** (line 40):
   - `extractTextPrimary`: `defer f.Close()` on pdf.Open
   - **Action:** Read-only, no change needed

8. **internal/pdf/transform.go** (line 40):
   - `MergeWithProgress`: `_ = tmpFile.Close()` immediately after create
   - **Action:** Empty placeholder file, acceptable

**Summary of Changes Needed:**
- **internal/fileio/files.go:CopyFile** - Add named return and defer close check for dstFile
- **internal/ocr/ocr.go:downloadTessdata** - Check error from tmpFile.Close() on line 202

**Example Fix for fileio.CopyFile:**
```go
func CopyFile(src, dst string) (err error) {  // Named return
    cleanSrc := filepath.Clean(src)
    cleanDst := filepath.Clean(dst)

    srcFile, err := os.Open(cleanSrc)
    if err != nil {
        return fmt.Errorf("failed to open source file: %w", err)
    }
    defer func() { _ = srcFile.Close() }()  // Read-only, can ignore

    if err := EnsureParentDir(cleanDst); err != nil {
        return err
    }

    dstFile, err := os.Create(cleanDst)
    if err != nil {
        return fmt.Errorf("failed to create destination file: %w", err)
    }
    defer func() {
        if cerr := dstFile.Close(); cerr != nil && err == nil {
            err = cerr  // Capture close error
        }
    }()

    if _, err = io.Copy(dstFile, srcFile); err != nil {
        return fmt.Errorf("failed to copy file: %w", err)
    }

    if err = dstFile.Sync(); err != nil {
        return fmt.Errorf("failed to sync file: %w", err)
    }

    return nil  // err is set by defer if close fails
}
```

#### 5.1.3 R11: Temp File Cleanup on Signals

**Temp File Creation Points:**

Based on MkdirTemp/CreateTemp grep analysis:

**Production Code:**
1. **internal/ocr/ocr.go:219** - `os.MkdirTemp("", "pdf-ocr-*")` in ExtractTextFromPDF
   - Already has `defer os.RemoveAll(tmpDir)` on line 223
   - **Action:** Add `cleanup.Register(tmpDir)` after line 219

2. **internal/ocr/ocr.go:193** - `os.CreateTemp(dataDir, "tessdata-*.tmp")` in downloadTessdata
   - Already has `defer os.Remove(tmpPath)` on line 198
   - **Action:** Add `cleanup.Register(tmpPath)` after line 197

3. **internal/pdf/text.go:166** - `os.MkdirTemp("", "pdf-cli-text-*")` in extractTextFallback
   - Already has `defer os.RemoveAll(tmpDir)` on line 170
   - **Action:** Add `cleanup.Register(tmpDir)` after line 166

4. **internal/pdf/transform.go:35** - `os.CreateTemp("", "pdf-merge-*.pdf")` in MergeWithProgress
   - Already has `defer os.Remove(tmpPath)` on line 41
   - **Action:** Add `cleanup.Register(tmpPath)` after line 39

5. **internal/commands/patterns/stdio.go:37** - `os.CreateTemp("", "pdf-cli-"+h.Operation+"-*.pdf")`
   - Has cleanup in handler.outputCleanup (line 44)
   - **Action:** Add `cleanup.Register(h.outputPath)` after line 42

6. **internal/fileio/stdio.go:32** - `os.CreateTemp("", "pdf-cli-stdin-*.pdf")` in ReadFromStdin
   - Has cleanup function returned to caller (line 39)
   - **Action:** Add `cleanup.Register(tmpPath)` after line 37

7. **internal/ocr/native.go:49** - `os.CreateTemp("", "ocr-output-*.txt")` in ProcessImage
   - Already has `defer os.Remove(tmpPath)` on line 55
   - **Action:** Add `cleanup.Register(tmpPath)` after line 53

8. **internal/fileio/files.go:48** - `os.CreateTemp(dir, ".pdf-cli-tmp-*")` in AtomicWrite
   - Has cleanup in defer (line 58)
   - **Action:** Add `cleanup.Register(tmpPath)` after line 52

**Test Code:** (104 instances in *_test.go files)
- **Action:** No changes needed for test code

**New Package Structure:**
```
internal/cleanup/
├── cleanup.go          # Register, Run, Clear functions
└── cleanup_test.go     # Tests for thread safety, signal handling
```

**Integration with main.go:**
```go
func run() int {
    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer stop()
    defer cleanup.Run()  // NEW: Run cleanup on exit (normal or signal)

    cli.SetVersion(version, commit, date)
    if err := cli.ExecuteContext(ctx); err != nil {
        return 1
    }
    return 0
}
```

### 5.2 Migration Concerns

**Backwards Compatibility:**
- Error message format may change (joined errors use newline-separated format)
- Functions that previously succeeded partially may now return errors
- Callers should be reviewed to handle new error cases

**Gradual Rollout:**
1. Implement cleanup package first (isolated, low risk)
2. Add close error checking to write paths (affects 2 files)
3. Update OCR error collection last (most complex, affects error semantics)

**Rollback Plan:**
- Each change is independent and can be reverted separately
- Tests should catch regressions before merging
- Monitor error logs after deployment for unexpected error patterns

### 5.3 Performance Implications

**errors.Join:**
- Minimal overhead (just slice allocation and string concatenation)
- Only happens on error path (not hot path)
- No measurable performance impact expected

**Named Returns + Defer:**
- Defer has tiny overhead (~50ns per defer call)
- Only applies to file write operations (not hot path)
- Existing code already uses defer extensively

**Cleanup Registry:**
- Mutex acquisition on Register and Run
- Register happens O(10) times per execution (not a bottleneck)
- Run happens once at exit (no performance concern)

**Overall Assessment:** No significant performance impact expected.

### 5.4 Testing Strategies

#### 5.4.1 Error Collection Tests

**Test Cases:**
```go
// TestProcessImagesParallelCollectsAllErrors
// - Setup: Mock backend that fails on images 1, 3, 5
// - Verify: errors.Join returns error mentioning all three failures
// - Verify: Successful images (0, 2, 4) still have text in result

// TestProcessImagesSequentialCollectsAllErrors
// - Setup: Mock backend that fails on every other image
// - Verify: All errors are returned
// - Verify: Partial text results are included
```

#### 5.4.2 Close Error Tests

**Test Cases:**
```go
// TestCopyFileCloseError
// - Setup: Mock filesystem that fails on Close (requires test file wrapper)
// - Verify: CopyFile returns error from Close
// - Alternative: Check that defer close pattern is syntactically correct

// Note: Testing actual close errors is difficult without mocking OS layer
// May need to use integration test with full disk scenario
```

#### 5.4.3 Cleanup Registry Tests

**Test Cases:**
```go
// TestCleanupRegisterAndRun
// - Create temp files
// - Register with cleanup
// - Call cleanup.Run()
// - Verify files are removed

// TestCleanupThreadSafety
// - Run with -race flag
// - Concurrent Register calls from multiple goroutines
// - Verify no data races

// TestCleanupIdempotence
// - Register same path twice
// - Verify single removal (os.RemoveAll is idempotent)
// - Verify no errors from removing non-existent files

// TestCleanupWithSignal
// - Start subprocess running pdf-cli
// - Send SIGINT
// - Verify temp files are cleaned up
// - This may need to be integration test or documented manual test
```

#### 5.4.4 Coverage Maintenance

**Current Coverage:**
```
internal/cli:              93.7%
internal/commands:         82.8%
internal/fileio:           78.2%
internal/ocr:              75.0%  <- Focus here
internal/pdf:              84.5%
```

**Target:** >= 81% overall

**Strategy:**
- Add error path tests to increase ocr package coverage (currently 75.0%)
- New cleanup package should achieve 90%+ coverage (simple logic)
- File close error checks may be hard to test (manual verification acceptable)

---

## 6. Open Questions and Future Investigation

### 6.1 Should cleanup.Run() be called via defer or atexit pattern?

**Current Recommendation:** Use defer in main.go's run() function.

**Alternative:** Register cleanup.Run() with signal handler directly.

**Trade-offs:**
- Defer is simpler and runs on normal exit too
- Signal handler would be more explicit but requires coordinating with existing signal.NotifyContext

**Decision:** Defer is sufficient, aligns with Go idioms.

### 6.2 Should partial results be returned when errors occur?

**Current Recommendation:** Yes, return both partial results and joined errors.

**Rationale:**
- User may still find partial OCR text useful
- Errors indicate which pages failed
- Caller can decide to use partial results or fail completely

**Alternative:** Return empty string on any error (fail-fast).

**Decision:** Return partial results + errors (more information to caller).

### 6.3 Should cleanup registry support priority or ordering?

**Current Recommendation:** No, simple FIFO cleanup is sufficient.

**Rationale:**
- All temp files/dirs can be removed in any order
- No dependencies between temp artifacts
- Keeping implementation simple reduces risk

**Future Enhancement:** If nested temp directories become an issue, could add reverse-order cleanup.

---

## 7. Estimated Complexity

**Overall Phase Complexity:** M (Medium)

**Breakdown:**
- R6 (Error collection): S-M (simple concept, moderate testing)
- R8 (Close errors): S (straightforward pattern, 2 files)
- R11 (Cleanup registry): M (new package, signal integration, comprehensive testing)

**Estimated LOC Changes:**
- New code: ~100 lines (cleanup package + tests)
- Modified code: ~50 lines (error collection, close checks)
- Test code: ~150 lines (error path coverage)
- Total: ~300 lines

**Time Estimate:**
- Implementation: 4-6 hours
- Testing: 3-4 hours
- Documentation: 1-2 hours
- Total: 8-12 hours (1-2 days)

---

## 8. Conclusion

Phase 4 addresses three well-understood reliability gaps using standard Go patterns. The recommended approach uses stdlib solutions (errors.Join, named returns, sync.Mutex) with no new dependencies. Integration with existing Phase 2 signal handling is straightforward.

**Key Success Factors:**
1. Comprehensive error path testing (especially OCR error collection)
2. Clear documentation of error message format changes
3. Verification with `go test -race ./...` for cleanup thread safety
4. Integration test or manual test of signal-based cleanup

**Risks are low** given the maturity of the patterns and existing project infrastructure (signal.NotifyContext, processBatch error joining precedent).

**Recommendation:** Proceed with implementation using recommended approaches.

---

## Appendix A: Current Error Handling Inventory

### Files with Ignored Close Errors (Relevant to R8)

| File | Line | Context | Write Operation? | Action Required |
|------|------|---------|------------------|-----------------|
| internal/fileio/files.go | 94 | CopyFile srcFile | No (read) | None |
| internal/fileio/files.go | 104 | CopyFile dstFile | Yes (write) | Check close error |
| internal/fileio/stdio.go | 63 | WriteToStdout | No (read) | None |
| internal/ocr/ocr.go | 187 | HTTP resp.Body | No (read) | None |
| internal/ocr/ocr.go | 202 | downloadTessdata tmpFile | Yes (write) | Check close error |
| internal/ocr/ocr.go | 205 | downloadTessdata error path | Yes (write) | Check close error |
| internal/ocr/native.go | 54 | ProcessImage tmpFile | No (placeholder) | None |
| internal/ocr/wasm.go | 82 | tessDataFile | No (read) | None |
| internal/ocr/wasm.go | 106 | imgFile | No (read) | None |
| internal/pdf/text.go | 40 | PDF file | No (read) | None |
| internal/pdf/transform.go | 40 | tmpFile | No (placeholder) | None |

**Total Write Operations with Ignored Close:** 2 (fileio/files.go, ocr/ocr.go)

### Files with Temp File Creation (Relevant to R11)

| File | Line | Pattern | Current Cleanup | Action Required |
|------|------|---------|-----------------|-----------------|
| internal/ocr/ocr.go | 219 | MkdirTemp pdf-ocr-* | defer RemoveAll | Add cleanup.Register |
| internal/ocr/ocr.go | 193 | CreateTemp tessdata-*.tmp | defer Remove | Add cleanup.Register |
| internal/pdf/text.go | 166 | MkdirTemp pdf-cli-text-* | defer RemoveAll | Add cleanup.Register |
| internal/pdf/transform.go | 35 | CreateTemp pdf-merge-*.pdf | defer Remove | Add cleanup.Register |
| internal/commands/patterns/stdio.go | 37 | CreateTemp pdf-cli-*-*.pdf | handler cleanup | Add cleanup.Register |
| internal/fileio/stdio.go | 32 | CreateTemp pdf-cli-stdin-*.pdf | returned cleanup func | Add cleanup.Register |
| internal/ocr/native.go | 49 | CreateTemp ocr-output-*.txt | defer Remove | Add cleanup.Register |
| internal/fileio/files.go | 48 | CreateTemp .pdf-cli-tmp-* | defer Remove | Add cleanup.Register |

**Total Temp Files in Production Code:** 8 locations

---

## Appendix B: Example Cleanup Package Implementation

```go
// internal/cleanup/cleanup.go
package cleanup

import (
    "os"
    "sync"
)

var (
    mu    sync.Mutex
    paths []string
)

// Register adds a path to be cleaned up on exit.
// Thread-safe, can be called from multiple goroutines.
func Register(path string) {
    if path == "" {
        return
    }
    mu.Lock()
    paths = append(paths, path)
    mu.Unlock()
}

// Run removes all registered paths.
// Should only be called once, typically via defer in main.
// Errors are ignored (cleanup is best-effort).
func Run() {
    mu.Lock()
    defer mu.Unlock()

    for _, p := range paths {
        _ = os.RemoveAll(p)
    }
    paths = nil
}

// Clear removes all registered paths without cleaning them.
// Used for testing.
func Clear() {
    mu.Lock()
    paths = nil
    mu.Unlock()
}
```

```go
// internal/cleanup/cleanup_test.go
package cleanup

import (
    "os"
    "path/filepath"
    "sync"
    "testing"
)

func TestRegisterAndRun(t *testing.T) {
    Clear() // Start with clean state

    tmpDir, err := os.MkdirTemp("", "cleanup-test-*")
    if err != nil {
        t.Fatal(err)
    }

    testFile := filepath.Join(tmpDir, "test.txt")
    if err := os.WriteFile(testFile, []byte("test"), 0600); err != nil {
        t.Fatal(err)
    }

    Register(tmpDir)
    Run()

    if _, err := os.Stat(tmpDir); !os.IsNotExist(err) {
        t.Errorf("directory still exists after cleanup: %s", tmpDir)
    }
}

func TestThreadSafety(t *testing.T) {
    Clear()

    var wg sync.WaitGroup
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func(n int) {
            defer wg.Done()
            tmpDir, _ := os.MkdirTemp("", "cleanup-concurrent-*")
            Register(tmpDir)
        }(i)
    }
    wg.Wait()

    Run() // Should not panic or race
}

func TestIdempotence(t *testing.T) {
    Clear()

    tmpDir, err := os.MkdirTemp("", "cleanup-idem-*")
    if err != nil {
        t.Fatal(err)
    }

    Register(tmpDir)
    Register(tmpDir) // Register twice

    Run() // Should not error on double-removal
    Run() // Should not error when already removed
}
```

---

## Sources

- [Go Package Documentation - errors.Join](https://pkg.go.dev/errors#Join)
- [Joining Errors in Golang - GeeksforGeeks](https://www.geeksforgeeks.org/go-language/joining-errors-in-golang/)
- [New in Go 1.20: wrapping multiple errors](https://lukas.zapletalovi.com/posts/2022/wrapping-multiple-errors/)
- [Wrapping Multiple Errors in Golang](https://www.tiredsg.dev/blog/wrapping-multiple-errors-golang/)
- [Don't defer Close() on writable files - Joe Shaw](https://www.joeshaw.org/dont-defer-close-on-writable-files/)
- [Handling Errors from Deferred Functions in Go - Thomas Stringer](https://trstringer.com/golang-deferred-function-error-handling/)
- [Understanding Go's defer: Usage, Evaluation, and Error Handling](https://claudiuconstantinbogdan.me/articles/go-defer)
- [Using defer in Go: Best Practices and Common Use Cases](https://dev.to/zakariachahboun/common-use-cases-for-defer-in-go-1071)
- [Go Package Documentation - os/signal](https://pkg.go.dev/os/signal)
- [Signal Handling in Go Applications - Medium](https://medium.com/@AlexanderObregon/signal-handling-in-go-applications-b96eb61ecb69)
- [Graceful Shutdown in Go: Practical Patterns](https://victoriametrics.com/blog/go-graceful-shutdown/)
- [Go by Example: Signals](https://gobyexample.com/signals)
- [GO capture ctrl+c signal & run cleanup function](https://www.golinuxcloud.com/go-capture-ctrl-c/)
- [A practical guide to error handling in Go - Datadog](https://www.datadoghq.com/blog/go-error-handling/)
- [Popular Error Handling Techniques in Go - JetBrains Guide](https://www.jetbrains.com/guide/go/tutorials/handle_errors_in_go/error_technique/)
