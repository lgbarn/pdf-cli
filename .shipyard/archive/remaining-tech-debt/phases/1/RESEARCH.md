# Phase 1: OCR Download Path Hardening - Research

## Executive Summary

This document analyzes the current implementation of OCR download functionality and context propagation in pdf-cli to inform the implementation of Phase 1 requirements (R4, R6, R10, R12) from the "Remaining Tech Debt" milestone.

**Key Findings:**
- 2 `context.TODO()` calls in production code (ocr.go:175, wasm.go:53)
- 1 usage of `http.DefaultClient` without timeout (ocr.go:260)
- 1 usage of `time.After` in retry logic creating potential timer leak (retry.go:76)
- Progress bar created once outside retry loop, causing poor UX on retry attempts (ocr.go:240-289)
- All signature changes are isolated to the `internal/ocr` package with one external caller in `internal/commands/text.go`

## Call Chain Analysis

### Context Flow

**Entry Point to OCR:**
```
cmd/pdf/main.go (line 26)
  └─> signal.NotifyContext creates root context
      └─> cobra.Command.Execute(ctx)
          └─> internal/commands/text.go:runText (line 112)
              └─> engine.ExtractTextFromPDF(cmd.Context(), ...)
                  └─> internal/ocr/ocr.go:327
```

**ExtractTextFromPDF Flow:**
```
ocr.go:327 ExtractTextFromPDF(ctx, ...)
  └─> ocr.go:329: e.EnsureTessdata() [NO CONTEXT PASSED]
      └─> ocr.go:171: loops over languages
          └─> ocr.go:175: downloadTessdata(context.TODO(), ...) [VIOLATION R4]
              └─> ocr.go:208: delegates to downloadTessdataWithBaseURL
                  └─> ocr.go:224: creates child context with timeout
                  └─> ocr.go:242: retry.Do accepts context
                      └─> ocr.go:255: http.NewRequestWithContext uses retry context
                      └─> ocr.go:260: http.DefaultClient.Do(req) [VIOLATION R6]
```

**WASM Backend Flow:**
```
ocr.go:327 ExtractTextFromPDF(ctx, ...)
  └─> ocr.go:328: checks if backend is WASM
      └─> ocr.go:329: e.EnsureTessdata() [NO CONTEXT PASSED]
          └─> ocr.go:171: Engine.EnsureTessdata loops languages
              └─> ocr.go:175: downloadTessdata(context.TODO(), ...) [VIOLATION R4]

wasm.go:98 ProcessImage(ctx, ...)
  └─> wasm.go:98: w.initializeTesseract(ctx, ...)
      └─> wasm.go:62: initializeTesseract accepts context
          └─> wasm.go:71: w.EnsureTessdata(lang) [NO CONTEXT PASSED]
              └─> wasm.go:45: WASMBackend.EnsureTessdata loops languages
                  └─> wasm.go:53: downloadTessdata(context.TODO(), ...) [VIOLATION R4]
```

### Functions Requiring Signature Changes

**Signature Chain (from top to bottom):**

1. **Engine.EnsureTessdata** (ocr.go:171)
   - Current: `func (e *Engine) EnsureTessdata() error`
   - New: `func (e *Engine) EnsureTessdata(ctx context.Context) error`
   - Called by: ExtractTextFromPDF (ocr.go:329)

2. **WASMBackend.EnsureTessdata** (wasm.go:45)
   - Current: `func (w *WASMBackend) EnsureTessdata(lang string) error`
   - New: `func (w *WASMBackend) EnsureTessdata(ctx context.Context, lang string) error`
   - Called by: initializeTesseract (wasm.go:71)

3. **WASMBackend.initializeTesseract** (wasm.go:62)
   - Current: Already accepts context
   - No change needed
   - Called by: ProcessImage (wasm.go:98) - already receives context

### All Call Sites

**downloadTessdata calls (2 total):**
- `internal/ocr/ocr.go:175` - Engine.EnsureTessdata
- `internal/ocr/wasm.go:53` - WASMBackend.EnsureTessdata

**Engine.EnsureTessdata calls (1 in production, 3 in tests):**
- `internal/ocr/ocr.go:329` - ExtractTextFromPDF (production)
- `internal/ocr/ocr_test.go:114` - TestEnsureTessdataDir
- `internal/ocr/engine_extended_test.go:31` - TestEnsureTessdata
- `internal/ocr/engine_extended_test.go:55` - TestEnsureTessdataMultipleLanguages

**WASMBackend.EnsureTessdata calls (4 in tests, 1 internal):**
- `internal/ocr/wasm.go:71` - initializeTesseract (production)
- `internal/ocr/wasm_test.go:83` - TestWASMBackendEnsureTessdata
- `internal/ocr/wasm_test.go:88` - TestWASMBackendEnsureTessdata

**http.DefaultClient usage (1 total):**
- `internal/ocr/ocr.go:260` - downloadTessdataWithBaseURL retry loop

**time.After usage (1 total):**
- `internal/retry/retry.go:76` - Do function retry loop

## R4: Context Propagation Analysis

### Current Violations

**Location 1: internal/ocr/ocr.go:175**
```go
func (e *Engine) EnsureTessdata() error {
    for _, lang := range parseLanguages(e.lang) {
        dataFile := filepath.Join(e.dataDir, lang+".traineddata")
        if _, err := os.Stat(dataFile); os.IsNotExist(err) {
            if err := downloadTessdata(context.TODO(), e.dataDir, lang); err != nil {
                return fmt.Errorf("failed to download tessdata for %s: %w", lang, err)
            }
        }
    }
    return nil
}
```

**Location 2: internal/ocr/wasm.go:53**
```go
func (w *WASMBackend) EnsureTessdata(lang string) error {
    if lang == "" {
        lang = w.lang
    }
    for _, l := range parseLanguages(lang) {
        dataFile := filepath.Join(w.dataDir, l+".traineddata")
        if _, err := os.Stat(dataFile); os.IsNotExist(err) {
            if err := downloadTessdata(context.TODO(), w.dataDir, l); err != nil {
                return fmt.Errorf("failed to download tessdata for %s: %w", l, err)
            }
        }
    }
    return nil
}
```

### Context Availability

**At ocr.go:175:**
- Context is available at caller (ExtractTextFromPDF line 327)
- ExtractTextFromPDF accepts `ctx context.Context` parameter
- Solution: Add context parameter to Engine.EnsureTessdata, thread through from ExtractTextFromPDF

**At wasm.go:53:**
- Context is available at caller (initializeTesseract line 62)
- initializeTesseract already accepts `ctx context.Context` parameter
- Solution: Add context parameter to WASMBackend.EnsureTessdata, thread through from initializeTesseract

### Impact Assessment

**Function Signatures to Change:**
1. `Engine.EnsureTessdata() error` → `Engine.EnsureTessdata(ctx context.Context) error`
2. `WASMBackend.EnsureTessdata(lang string) error` → `WASMBackend.EnsureTessdata(ctx context.Context, lang string) error`

**External Callers:**
- None outside internal/ocr package
- One production call in ExtractTextFromPDF (already has context)
- One internal call in initializeTesseract (already has context)

**Test Updates Required:**
- `internal/ocr/ocr_test.go:114` - TestEnsureTessdataDir
- `internal/ocr/engine_extended_test.go:31` - TestEnsureTessdata
- `internal/ocr/engine_extended_test.go:55` - TestEnsureTessdataMultipleLanguages
- `internal/ocr/wasm_test.go:83, 88` - TestWASMBackendEnsureTessdata

All test call sites can use `context.Background()` as tests don't need cancellation.

## R6: HTTP Client Timeout Analysis

### Current Violation

**Location: internal/ocr/ocr.go:260**
```go
resp, doErr := http.DefaultClient.Do(req)
```

**Context:**
- Within downloadTessdataWithBaseURL function
- Already has context timeout at line 224: `ctx, cancel := context.WithTimeout(ctx, DefaultDownloadTimeout)`
- Request created with context at line 255: `http.NewRequestWithContext(retryCtx, ...)`
- DefaultDownloadTimeout = 5 minutes (ocr.go:38)

### Existing Timeout Mechanisms

**Current Protection:**
1. Context timeout: 5 minutes (line 224)
2. Request context: Derived from timeout context (line 255)
3. HTTP client: No timeout (uses http.DefaultClient)

**Gap:**
The HTTP client has no timeout configured. While the request context provides cancellation, http.DefaultClient itself has no transport-level timeouts.

### Solution Requirements (from CONTEXT-1.md)

Decision: Set `http.Client.Timeout` to 5 minutes, matching the existing `DefaultDownloadTimeout` context timeout.

**Implementation:**
```go
var tessdataHTTPClient = &http.Client{
    Timeout: DefaultDownloadTimeout,
}
```

**Rationale:**
- Belt and suspenders approach - both context and HTTP client enforce same limit
- HTTP client timeout is a transport-level timeout (entire request/response cycle)
- Context timeout provides cancellation mechanism
- Having both ensures timeout even if context is not properly propagated

### Alternative Considered: Custom http.Transport

Could configure custom Transport with more granular timeouts:
```go
var tessdataHTTPClient = &http.Client{
    Timeout: DefaultDownloadTimeout,
    Transport: &http.Transport{
        DialContext: (&net.Dialer{
            Timeout:   30 * time.Second,
            KeepAlive: 30 * time.Second,
        }).DialContext,
        TLSHandshakeTimeout:   10 * time.Second,
        ResponseHeaderTimeout: 10 * time.Second,
        ExpectContinueTimeout: 1 * time.Second,
    },
}
```

**Rejected Rationale:**
- Over-engineered for a simple file download use case
- Simple `Client.Timeout` sufficient for single-shot downloads
- Additional complexity not justified by requirements

## R10: time.After Replacement Analysis

### Current Violation

**Location: internal/retry/retry.go:76**
```go
select {
case <-time.After(delay):
case <-ctx.Done():
    return ctx.Err()
}
```

**Context:**
- Within retry loop (lines 56-81)
- Runs up to maxAttempts times (default 3)
- Uses exponential backoff with delay calculation at line 71

### Why time.After is Problematic

From Go documentation:
> time.After creates a timer that will send on its channel after duration d. The timer cannot be stopped. If the function returns before the timer expires, the timer will leak until it fires.

**Memory Leak Scenario:**
1. Retry attempt N starts
2. `time.After(delay)` creates timer for, say, 2 seconds
3. Context is cancelled at 1 second
4. Function returns via `ctx.Done()` case
5. Timer continues running for another 1 second, holding memory
6. With many retries or long delays, this accumulates

### Solution Requirements

Replace with `time.NewTimer` and explicit `Stop()`:

**Before:**
```go
select {
case <-time.After(delay):
case <-ctx.Done():
    return ctx.Err()
}
```

**After:**
```go
timer := time.NewTimer(delay)
defer timer.Stop()
select {
case <-timer.C:
case <-ctx.Done():
    if !timer.Stop() {
        <-timer.C
    }
    return ctx.Err()
}
```

**Rationale:**
- `timer.Stop()` releases timer resources immediately on context cancellation
- `defer timer.Stop()` ensures cleanup even if loop exits early
- Drain pattern (`if !timer.Stop() { <-timer.C }`) prevents goroutine leak if timer already fired

### Test Coverage Impact

**Existing tests in retry_test.go:**
- TestDoSuccess (line 11) - No delay, no impact
- TestDoRetryThenSuccess (line 26) - Has delays, should verify behavior unchanged
- TestDoExhaustion (line 44) - Uses 1ms delays, low impact
- TestDoPermanentError (line 60) - No retries, no impact
- TestDoContextCancellation (line 76) - Critical test for timer cleanup verification
- TestDoBackoffTiming (line 92) - Tests delay timing, should verify behavior unchanged

**TestDoContextCancellation is key:**
This test cancels context during retry and verifies cancellation works. After R10 fix, this test will also implicitly verify that timers are cleaned up (no goroutine leak).

## R12: Progress Bar Recreation Analysis

### Current Implementation

**Location: internal/ocr/ocr.go:240-289**

```go
// Progress bar created once; may not display perfectly on retries
var bar *progressbar.ProgressBar

retryErr := retry.Do(ctx, retry.Options{
    MaxAttempts: DefaultRetryAttempts,
    BaseDelay:   DefaultRetryBaseDelay,
}, func(retryCtx context.Context) error {
    // Reset temp file and hasher for each attempt
    if _, seekErr := tmpFile.Seek(0, io.SeekStart); seekErr != nil {
        return retry.Permanent(fmt.Errorf("failed to seek temp file: %w", seekErr))
    }
    if truncErr := tmpFile.Truncate(0); truncErr != nil {
        return retry.Permanent(fmt.Errorf("failed to truncate temp file: %w", truncErr))
    }
    hasher.Reset()

    req, reqErr := http.NewRequestWithContext(retryCtx, http.MethodGet, dlURL, nil)
    if reqErr != nil {
        return retry.Permanent(reqErr)
    }

    resp, doErr := http.DefaultClient.Do(req)
    if doErr != nil {
        // Network error — retryable
        fmt.Fprintf(os.Stderr, "Download attempt failed: %v\n", doErr)
        return doErr
    }
    defer resp.Body.Close() //nolint:errcheck // response body

    switch {
    case resp.StatusCode == http.StatusOK:
        // Success — download the body
    case resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500:
        // Retryable server error
        return fmt.Errorf("failed to download: HTTP %d", resp.StatusCode)
    default:
        // Other client errors (4xx) — permanent
        return retry.Permanent(fmt.Errorf("failed to download: HTTP %d", resp.StatusCode))
    }

    bar = progress.NewBytesProgressBar(
        fmt.Sprintf("Downloading %s.traineddata", lang),
        resp.ContentLength,
    )
    if _, copyErr := io.Copy(io.MultiWriter(tmpFile, bar, hasher), resp.Body); copyErr != nil {
        fmt.Fprintf(os.Stderr, "Download attempt failed during copy: %v\n", copyErr)
        return copyErr
    }

    return nil
})
```

### Problem Analysis

**Current Behavior:**
1. Progress bar declared outside retry loop (line 240)
2. Progress bar created inside retry function (line 279) but ONLY on successful HTTP response
3. On retry after network error, bar is nil (never created)
4. On retry after server error (503), bar is recreated but previous bar not finished
5. Comment acknowledges problem: "may not display perfectly on retries"

**UX Issues:**
1. **First attempt fails before creating bar (network error)**: No progress shown
2. **First attempt creates bar, then fails during copy**: Bar left incomplete, new bar overlays on retry
3. **Multiple retries**: Multiple overlapping progress bars create messy output

**Example Bad UX:**
```
Downloading eng.traineddata...
Download attempt failed: connection timeout
Downloading eng.traineddata [=====>          ] 5MB/15MB
Download attempt failed during copy: unexpected EOF
Downloading eng.traineddata [=>              ] 1MB/15MB
[===========================>] 15MB/15MB
```

### Solution Requirements

**Goal:** Clean progress display on each retry attempt

**Approach 1: Recreate bar at start of each retry attempt**
```go
retryErr := retry.Do(ctx, retry.Options{...}, func(retryCtx context.Context) error {
    // Create fresh progress bar at start of each attempt
    attemptBar := progress.NewBytesProgressBar(
        fmt.Sprintf("Downloading %s.traineddata", lang),
        -1, // unknown size until we get response
    )
    defer progress.FinishProgressBar(attemptBar)

    // ... rest of download logic

    // Update bar with actual size once known
    if resp.StatusCode == http.StatusOK {
        attemptBar.ChangeMax64(resp.ContentLength)
        if _, copyErr := io.Copy(io.MultiWriter(tmpFile, attemptBar, hasher), resp.Body); copyErr != nil {
            return copyErr
        }
    }

    return nil
})
```

**Problem with Approach 1:** Progress bar should only show during actual download, not during HTTP errors that return immediately.

**Approach 2: Recreate bar only when download starts (current + reset)**
```go
var bar *progressbar.ProgressBar

retryErr := retry.Do(ctx, retry.Options{...}, func(retryCtx context.Context) error {
    // Reset resources for retry
    if _, seekErr := tmpFile.Seek(0, io.SeekStart); seekErr != nil {
        return retry.Permanent(fmt.Errorf("failed to seek temp file: %w", seekErr))
    }
    if truncErr := tmpFile.Truncate(0); truncErr != nil {
        return retry.Permanent(fmt.Errorf("failed to truncate temp file: %w", truncErr))
    }
    hasher.Reset()

    // Reset progress bar from previous attempt
    if bar != nil {
        progress.FinishProgressBar(bar)
        bar = nil
    }

    // ... HTTP request logic

    if resp.StatusCode == http.StatusOK {
        // Create new bar for this download attempt
        bar = progress.NewBytesProgressBar(
            fmt.Sprintf("Downloading %s.traineddata", lang),
            resp.ContentLength,
        )
        if _, copyErr := io.Copy(io.MultiWriter(tmpFile, bar, hasher), resp.Body); copyErr != nil {
            return copyErr
        }
    }

    return nil
})

// Finish final bar if download succeeded
progress.FinishProgressBar(bar)
```

**Approach 2 Analysis:**
- Pros: Progress bar only shown during actual download, not for quick errors
- Pros: Each retry gets fresh progress bar starting at 0
- Cons: Previous progress bar finished before retry starts (minor visual gap)
- Matches existing pattern: bar created only on HTTP 200

**Recommended: Approach 2** - Minimal change to existing logic, clean UX

### Progress Bar Lifecycle

**From internal/progress/progress.go:**
```go
func NewBytesProgressBar(description string, total int64) *progressbar.ProgressBar {
    return progressbar.NewOptions64(total,
        progressbar.OptionSetDescription(description),
        progressbar.OptionSetWriter(os.Stderr),
        progressbar.OptionShowBytes(true),
        progressbar.OptionSetTheme(ProgressBarTheme),
    )
}

func FinishProgressBar(bar *progressbar.ProgressBar) {
    if bar != nil {
        fmt.Fprintln(os.Stderr)
    }
}
```

**Behavior:**
- `NewBytesProgressBar`: Creates bar, prints to stderr
- `FinishProgressBar`: Prints newline to separate bar from subsequent output
- Bar automatically renders as bytes written via `io.Copy`

### Test Coverage Impact

**Existing Tests:**
- `TestDownloadTessdataRetryOnServerError` (ocr_test.go:288) - Tests retry with 503 then success
- `TestDownloadTessdataNoRetryOn404` (ocr_test.go:333) - Tests no retry on 404
- `TestDownloadTessdataRetryExhaustion` (ocr_test.go:361) - Tests retry exhaustion with 500s

**Testing Challenges:**
- Progress bar writes to stderr, hard to capture in tests
- Tests use httptest.Server which returns immediately (no actual progress)
- Tests don't verify progress bar behavior, only download logic

**No New Tests Required:**
Progress bar is a presentation concern. Existing tests verify retry logic works correctly. Manual testing needed to verify clean progress UX.

## Testing Strategy

### Tests Requiring Updates

**For R4 (context propagation):**

1. **internal/ocr/ocr_test.go:114** - TestEnsureTessdataDir
   - Change: `engine.EnsureTessdata()` → `engine.EnsureTessdata(context.Background())`

2. **internal/ocr/engine_extended_test.go:31** - TestEnsureTessdata
   - Change: `engine.EnsureTessdata()` → `engine.EnsureTessdata(context.Background())`

3. **internal/ocr/engine_extended_test.go:55** - TestEnsureTessdataMultipleLanguages
   - Change: `engine.EnsureTessdata()` → `engine.EnsureTessdata(context.Background())`

4. **internal/ocr/wasm_test.go:83** - TestWASMBackendEnsureTessdata
   - Change: `backend.EnsureTessdata("eng")` → `backend.EnsureTessdata(context.Background(), "eng")`

5. **internal/ocr/wasm_test.go:88** - TestWASMBackendEnsureTessdata (second call)
   - Change: `backend.EnsureTessdata("")` → `backend.EnsureTessdata(context.Background(), "")`

**For R6 (HTTP client):**

No test changes needed. Existing tests use httptest.Server which works with any http.Client.

**For R10 (time.After):**

No test changes needed. Existing tests in retry_test.go verify behavior, not implementation.

**For R12 (progress bar):**

No test changes needed. Progress bar is presentation layer, not tested.

### New Tests to Consider

**For R4 (context cancellation):**

Consider adding test that cancels context during tessdata download to verify cancellation propagates:

```go
func TestEnsureTessdataContextCancellation(t *testing.T) {
    ctx, cancel := context.WithCancel(context.Background())
    cancel() // Cancel immediately

    engine, err := NewEngineWithOptions(EngineOptions{
        BackendType: BackendWASM,
        Lang:        "nonexistent_lang_xyz",
    })
    if err != nil {
        t.Fatal(err)
    }
    defer engine.Close()

    err = engine.EnsureTessdata(ctx)
    if !errors.Is(err, context.Canceled) {
        t.Errorf("expected context.Canceled, got %v", err)
    }
}
```

**For R10 (timer cleanup):**

Existing TestDoContextCancellation (retry_test.go:76) already tests context cancellation during delay. After R10 fix, could add goroutine leak detection, but may be overkill for this codebase.

## Implementation Order

### Recommended Sequence

**1. R6: HTTP Client (Simplest, No Breaking Changes)**
- Create `tessdataHTTPClient` variable with 5-minute timeout
- Replace `http.DefaultClient.Do(req)` with `tessdataHTTPClient.Do(req)`
- No function signatures change
- No tests change
- Can be done independently

**2. R10: time.After Replacement (Independent)**
- Replace `time.After` with `time.NewTimer` in retry.Do
- No function signatures change
- No tests change (behavior identical)
- Can be done independently

**3. R4: Context Propagation (Breaking Changes)**
- Add context parameter to Engine.EnsureTessdata
- Add context parameter to WASMBackend.EnsureTessdata
- Update all call sites (1 production, 5 test)
- Must be done atomically (compile will fail if incomplete)

**4. R12: Progress Bar Recreation (Depends on R4)**
- Requires testing actual retry behavior with network delays
- Best done after R4 (context) and R6 (HTTP client) are complete
- Can be validated with manual testing

### Atomic Change Groups

**Group A (Independent):** R6 + R10
- No dependencies between them
- Can be done in any order
- Can be separate commits or combined

**Group B (Depends on A completion):** R4
- Changes function signatures
- All call sites must be updated atomically
- Single commit recommended

**Group C (Depends on B completion):** R12
- Modifies retry function behavior
- Easiest to test after context propagation is complete
- Single commit recommended

## Risk Assessment

### R4 Risks

**Risk: Missing a call site**
- Likelihood: Low
- Impact: High (compilation failure)
- Mitigation: Compiler will catch all call sites. Use `grep -r "EnsureTessdata" internal/` to verify.

**Risk: Context not available at some call site**
- Likelihood: None
- Impact: N/A
- Mitigation: Analysis shows context available at all call sites.

### R6 Risks

**Risk: HTTP client timeout too aggressive**
- Likelihood: Low
- Impact: Medium (download failures for large files on slow connections)
- Mitigation: Using same 5-minute timeout as existing context timeout. Tessdata files are ~15MB max.

**Risk: HTTP client conflicts with context timeout**
- Likelihood: None
- Impact: N/A
- Mitigation: Belt-and-suspenders approach. Context timeout = HTTP client timeout.

### R10 Risks

**Risk: Timer drain logic incorrect**
- Likelihood: Low
- Impact: Medium (goroutine leak or deadlock)
- Mitigation: Standard Go pattern from time package documentation. Existing tests verify behavior.

**Risk: Performance regression from timer allocation**
- Likelihood: None
- Impact: N/A
- Mitigation: time.NewTimer is same cost as time.After. No performance difference.

### R12 Risks

**Risk: Progress bar not properly finished on retry**
- Likelihood: Low
- Impact: Low (cosmetic issue, messy terminal output)
- Mitigation: Defer FinishProgressBar or explicit finish before recreation.

**Risk: Progress bar recreation breaks terminal state**
- Likelihood: Low
- Impact: Low (cosmetic)
- Mitigation: progressbar library handles multiple bars. Manual testing will verify.

## Open Questions

1. **Should context cancellation propagate to in-flight downloads?**
   - Current: Yes, via http.NewRequestWithContext
   - After R4: Yes, context will propagate properly
   - Decision: No change needed, current behavior is correct

2. **Should we add metrics/logging for retry attempts?**
   - Current: fmt.Fprintf to stderr for failures
   - Potential: Use internal/logging package for structured logging
   - Decision: Out of scope for Phase 1. Can be addressed in future phase.

3. **Should progress bar show retry attempt number?**
   - Current: Description is "Downloading {lang}.traineddata"
   - Potential: "Downloading {lang}.traineddata (attempt 2/3)"
   - Decision: Out of scope for Phase 1. Would require retry.Do to pass attempt number to retry function.

4. **Should http.Client be reused across downloads?**
   - Current: Will create package-level var (reused)
   - Alternative: Create per-download (not reused)
   - Decision: Package-level var is correct. HTTP client is designed to be reused.

## References

### Code Locations

**Primary Files:**
- `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr.go` - Main OCR engine, download logic
- `/Users/lgbarn/Personal/pdf-cli/internal/ocr/wasm.go` - WASM backend implementation
- `/Users/lgbarn/Personal/pdf-cli/internal/retry/retry.go` - Retry logic with exponential backoff
- `/Users/lgbarn/Personal/pdf-cli/internal/progress/progress.go` - Progress bar utilities
- `/Users/lgbarn/Personal/pdf-cli/internal/commands/text.go` - CLI command calling OCR

**Test Files:**
- `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr_test.go` - Download retry tests
- `/Users/lgbarn/Personal/pdf-cli/internal/ocr/wasm_test.go` - WASM backend tests
- `/Users/lgbarn/Personal/pdf-cli/internal/ocr/engine_extended_test.go` - Engine integration tests
- `/Users/lgbarn/Personal/pdf-cli/internal/retry/retry_test.go` - Retry logic unit tests

### Constants

**From internal/ocr/ocr.go:**
- `TessdataURL = "https://github.com/tesseract-ocr/tessdata_fast/raw/main"` (line 29)
- `DefaultDownloadTimeout = 5 * time.Minute` (line 38)
- `DefaultRetryAttempts = 3` (line 43)
- `DefaultRetryBaseDelay = 1 * time.Second` (line 46)

**From internal/retry/retry.go:**
- `DefaultMaxAttempts = 3` (line 11)
- `DefaultBaseDelay = 1 * time.Second` (line 12)
- `DefaultMaxDelay = 30 * time.Second` (line 13)

### Conventions (from .shipyard/codebase/CONVENTIONS.md)

**Testing:**
- Standard Go testing (no testify)
- Table-driven tests with subtests
- Test files: `*_test.go`
- Integration tests: `*_integration_test.go`

**Error Handling:**
- Error wrapping with `fmt.Errorf` + `%w`
- Custom error types implement `Error()` and `Unwrap()`
- Errors checked immediately after call

**Code Organization:**
- Import groups: stdlib first, blank line, then external
- Factory functions with `New` prefix
- Unexported helpers grouped logically

## Conclusion

Phase 1 requirements are well-scoped and straightforward to implement:

1. **R4 (context.TODO)**: 2 call sites, both in internal/ocr, context available at all callers
2. **R6 (HTTP client)**: 1 call site, simple replacement with custom client
3. **R10 (time.After)**: 1 call site in retry package, standard timer pattern
4. **R12 (progress bar)**: Reset logic in single function, minimal change

All changes are isolated to `internal/ocr` and `internal/retry` packages. No changes required in commands layer or external APIs. Test updates are minimal (5 test files, simple context.Background() additions).

**Estimated Complexity:** Low
**Estimated Risk:** Low
**Estimated Effort:** 2-4 hours for implementation + testing
