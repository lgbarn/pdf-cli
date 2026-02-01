# Phase 6: Network Resilience and Retry Logic - Research Document

**Date**: 2026-01-31
**Phase**: Phase 6 - Network Resilience and Retry Logic
**Priority**: P2 (Medium Priority Technical Debt)

## Executive Summary

This phase adds retry logic with exponential backoff to the tessdata download path to improve resilience against transient network failures. The implementation will use a small, generic retry helper (no new external dependencies) that respects `context.Context` cancellation.

**Key Findings**:
- Current `downloadTessdata` function has no retry logic - any network failure requires manual retry
- Downloads from GitHub raw content (`https://github.com/tesseract-ocr/tessdata_fast/raw/main`)
- Context propagation already in place (Phase 2), download uses `http.NewRequestWithContext`
- Checksum verification already implemented (Phase 3)
- No existing retry patterns in codebase - need to implement from scratch
- Standard library provides sufficient primitives - no external dependencies needed
- Testing can use `net/http/httptest` for mock server with retry scenarios

---

## 1. Current Download Implementation Analysis

### Location and Code Structure

**File**: `/Users/lgbarn/Personal/pdf-cli/.worktrees/phase-2-concurrency/internal/ocr/ocr.go`

**Function**: `downloadTessdata` (Lines 201-277)

```go
func downloadTessdata(ctx context.Context, dataDir, lang string) (err error) {
	url := fmt.Sprintf("%s/%s.traineddata", TessdataURL, lang)
	dataFile := filepath.Join(dataDir, lang+".traineddata")

	// Sanitize the data file path to prevent directory traversal
	dataFile, err = fileio.SanitizePath(dataFile)
	if err != nil {
		return fmt.Errorf("invalid data file path: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Downloading tessdata for '%s'...\n", lang)

	ctx, cancel := context.WithTimeout(ctx, DefaultDownloadTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)  // NO RETRY LOGIC
	if err != nil {
		return err  // Immediate failure on network error
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download: HTTP %d", resp.StatusCode)
	}

	// ... temp file creation, download, checksum verification, rename
}
```

### Current Error Handling

**No Retry on Failure**:
- Network errors (timeout, connection reset, DNS failure) → immediate failure
- HTTP 5xx errors (server error, service unavailable) → immediate failure
- HTTP 4xx errors (not found, forbidden) → immediate failure (correct behavior)

**Context Integration**:
- Uses `context.WithTimeout(ctx, DefaultDownloadTimeout)` (5 minutes)
- HTTP request created with `http.NewRequestWithContext(ctx, ...)`
- Cancellation support already in place

**Constants** (Lines 26-41):
```go
const (
	TessdataURL             = "https://github.com/tesseract-ocr/tessdata_fast/raw/main"
	DefaultDownloadTimeout  = 5 * time.Minute
	DefaultDataDirPerm      = 0750
)
```

### Download Characteristics

**Source**: GitHub raw content CDN
- Base URL: `https://github.com/tesseract-ocr/tessdata_fast/raw/main`
- Example: `https://github.com/tesseract-ocr/tessdata_fast/raw/main/eng.traineddata`
- File sizes: ~10-30 MB per language
- Transport: HTTPS (TLS verified by default)

**Call Sites**:
1. `Engine.EnsureTessdata()` (Line 168) - Checks if file exists, downloads if missing
2. `wasm.go:ensureLangData()` (via same pattern)
3. Only called during engine initialization, not in hot paths

**Existing Safeguards**:
- Timeout: 5 minutes (sufficient for slow connections)
- Checksum verification: SHA256 hash checked after download (Phase 3)
- Path sanitization: Prevents directory traversal (Phase 3)
- Cleanup: Temp files registered with cleanup package (Phase 4)

---

## 2. Network Failure Scenarios

### Transient Errors (Should Retry)

Based on GitHub's CDN and general network behavior:

**5xx Server Errors**:
- `503 Service Unavailable` - GitHub CDN maintenance or overload
- `502 Bad Gateway` - Upstream CDN failure
- `504 Gateway Timeout` - CDN timeout
- `500 Internal Server Error` - Rare but transient

**Network-Level Errors**:
- `context.DeadlineExceeded` - Request timeout (could be transient)
- `net.DNSError` (temporary) - DNS resolution failure
- `net.OpError` with `syscall.ECONNRESET` - Connection reset by peer
- `net.OpError` with `syscall.ECONNREFUSED` - Connection refused
- `io.EOF` during download - Broken connection

**Rate Limiting**:
- `429 Too Many Requests` - Should retry with backoff (respect `Retry-After` header)

### Permanent Errors (Should NOT Retry)

**4xx Client Errors**:
- `404 Not Found` - Language file doesn't exist (e.g., typo in language code)
- `403 Forbidden` - Access denied (unlikely for public repo)
- `401 Unauthorized` - Authentication required (shouldn't happen for public URLs)
- `400 Bad Request` - Malformed request

**Context Cancellation**:
- `context.Canceled` - User pressed Ctrl+C, don't retry

**File System Errors**:
- Path sanitization failures - Invalid path, don't retry
- Temp file creation errors - Disk full, permissions issue
- Checksum verification failure - Corrupted download, should retry

---

## 3. Retry Strategy Options

### Option 1: Inline Retry Logic in downloadTessdata

**Approach**: Add retry loop directly in the `downloadTessdata` function.

**Pros**:
- No new files or abstractions
- Simple and direct
- All logic visible in one place

**Cons**:
- Not reusable for future network operations
- Harder to test in isolation
- Mixes retry logic with download logic

**Maturity**: N/A (custom implementation)
**Complexity**: Low

### Option 2: Generic Retry Helper in internal/ocr/retry.go

**Approach**: Create a small, focused retry helper in the OCR package.

**Pros**:
- Reusable within OCR package
- Testable in isolation
- Clean separation of concerns
- Stays within existing package structure

**Cons**:
- Scoped to OCR package only
- Potential duplication if other packages need retry

**Maturity**: N/A (custom implementation)
**Complexity**: Low

**Example API**:
```go
// internal/ocr/retry.go
func retryWithBackoff(ctx context.Context, maxAttempts int, operation func() error) error
```

### Option 3: Generic Retry Package in internal/retry/retry.go

**Approach**: Create a standalone retry package for project-wide use.

**Pros**:
- Reusable across all packages
- Clear abstraction boundary
- Easier to evolve and test
- Follows Go package organization best practices
- Aligns with ROADMAP.md suggestion (line 232: "or generic `internal/retry/retry.go`")

**Cons**:
- Slightly more upfront design work
- Need to ensure it doesn't become over-engineered

**Maturity**: N/A (custom implementation)
**Complexity**: Low-Medium

**Example API**:
```go
// internal/retry/retry.go
package retry

type Policy struct {
	MaxAttempts   int
	InitialDelay  time.Duration
	MaxDelay      time.Duration
	Multiplier    float64
	ShouldRetry   func(error) bool
}

func Do(ctx context.Context, policy Policy, operation func() error) error
```

### Option 4: External Library (NOT RECOMMENDED)

**Popular Libraries**:
- [`cenkalti/backoff/v4`](https://github.com/cenkalti/backoff) - 2.4k stars, well-maintained
- [`avast/retry-go`](https://github.com/avast/retry-go) - 2.4k stars
- [`sethvargo/go-retry`](https://github.com/sethvargo/go-retry) - highly extensible

**Pros**:
- Battle-tested implementations
- Feature-rich (jitter, context support, etc.)
- Less code to maintain

**Cons**:
- **ROADMAP explicitly forbids**: "no new external deps" (line 209)
- Adds dependency for a simple use case
- Go standard library is sufficient
- Project philosophy favors minimal dependencies

**Maturity**: High
**Recommendation**: **DO NOT USE** - violates project requirements

---

## 4. Recommended Approach

### Selection: Option 3 - Generic Retry Package

**Rationale**:
1. **Reusability**: PDF operations might need retry logic in the future (network watermarks, remote file access)
2. **Testability**: Isolated package with clear boundaries
3. **Alignment**: ROADMAP.md line 232 suggests "or generic `internal/retry/retry.go`"
4. **Maintainability**: Clean separation makes future enhancements easier
5. **No Dependencies**: Uses only Go standard library

### Implementation Design

**Package**: `internal/retry`

**File**: `internal/retry/retry.go`

**API**:
```go
package retry

import (
	"context"
	"errors"
	"math"
	"time"
)

// Config holds retry policy configuration.
type Config struct {
	MaxAttempts  int           // Maximum number of attempts (including first try)
	InitialDelay time.Duration // Delay after first failure
	MaxDelay     time.Duration // Cap on exponential backoff
	Multiplier   float64       // Backoff multiplier (2.0 for doubling)
}

// DefaultConfig returns sensible defaults for HTTP retries.
func DefaultConfig() Config {
	return Config{
		MaxAttempts:  3,
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
	}
}

// Do executes operation with exponential backoff retry.
// Returns the result of the last attempt if all retries fail.
func Do(ctx context.Context, cfg Config, operation func() error) error {
	var lastErr error

	for attempt := 1; attempt <= cfg.MaxAttempts; attempt++ {
		// Check context before each attempt
		if ctx.Err() != nil {
			return ctx.Err()
		}

		lastErr = operation()

		// Success - return immediately
		if lastErr == nil {
			return nil
		}

		// Don't retry if error is permanent
		if !IsRetryable(lastErr) {
			return lastErr
		}

		// Last attempt - don't sleep
		if attempt == cfg.MaxAttempts {
			return lastErr
		}

		// Calculate backoff delay
		delay := calculateBackoff(attempt, cfg)

		// Sleep with context cancellation support
		select {
		case <-time.After(delay):
			// Continue to next attempt
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return lastErr
}

// IsRetryable returns true if the error should trigger a retry.
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	// Never retry context cancellation
	if errors.Is(err, context.Canceled) {
		return false
	}

	// Timeout might be transient
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	// Check for HTTP status code errors
	if IsHTTPError(err) {
		return IsRetryableHTTPError(err)
	}

	// Network errors are generally retryable
	return IsNetworkError(err)
}

func calculateBackoff(attempt int, cfg Config) time.Duration {
	delay := float64(cfg.InitialDelay) * math.Pow(cfg.Multiplier, float64(attempt-1))
	if delay > float64(cfg.MaxDelay) {
		delay = float64(cfg.MaxDelay)
	}
	return time.Duration(delay)
}
```

**Helper Functions**:
```go
// IsHTTPError checks if error contains HTTP status information
func IsHTTPError(err error) bool {
	// Check error message for "HTTP XXX" pattern
	// Alternative: wrap HTTP errors in custom type
}

// IsRetryableHTTPError determines if HTTP status code warrants retry
func IsRetryableHTTPError(err error) bool {
	// Extract status code from error message
	// Retry on: 429, 500, 502, 503, 504
	// Don't retry on: 4xx (except 429)
}

// IsNetworkError checks for network-level errors
func IsNetworkError(err error) bool {
	// Check for net.OpError, net.DNSError, io.EOF, etc.
}
```

### HTTP Status Code Strategy

Based on [best practices research](https://www.baeldung.com/cs/http-error-status-codes-retry):

**Retry on**:
- `429 Too Many Requests` - Rate limiting, definitely retry
- `500 Internal Server Error` - Server issue, likely transient
- `502 Bad Gateway` - Upstream issue, retry
- `503 Service Unavailable` - Temporary unavailability
- `504 Gateway Timeout` - Timeout, might succeed on retry

**Don't Retry on**:
- `400 Bad Request` - Client error, won't change
- `401 Unauthorized` - Auth required, won't work without credentials
- `403 Forbidden` - Access denied, permanent
- `404 Not Found` - Resource doesn't exist (e.g., invalid language code)
- `405 Method Not Allowed` - Wrong HTTP method
- All other 4xx codes

**Implementation Approach**:

For clean error handling, wrap HTTP errors in a custom type:

```go
// internal/ocr/ocr.go - updated downloadTessdata
type HTTPError struct {
	StatusCode int
	Message    string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Message)
}

// In downloadTessdata:
if resp.StatusCode != http.StatusOK {
	return &HTTPError{
		StatusCode: resp.StatusCode,
		Message:    fmt.Sprintf("failed to download %s.traineddata", lang),
	}
}

// In retry package:
func IsRetryableHTTPError(err error) bool {
	var httpErr *HTTPError
	if !errors.As(err, &httpErr) {
		return false
	}

	switch httpErr.StatusCode {
	case 429, 500, 502, 503, 504:
		return true
	default:
		return false
	}
}
```

### Integration with downloadTessdata

**Modified Function**:
```go
func downloadTessdata(ctx context.Context, dataDir, lang string) error {
	retryConfig := retry.Config{
		MaxAttempts:  3,
		InitialDelay: 1 * time.Second,
		MaxDelay:     10 * time.Second,
		Multiplier:   2.0,
	}

	return retry.Do(ctx, retryConfig, func() error {
		return downloadTessdataAttempt(ctx, dataDir, lang)
	})
}

func downloadTessdataAttempt(ctx context.Context, dataDir, lang string) error {
	// Current downloadTessdata implementation
	// Return HTTPError for status code errors
	// Return network errors as-is (they'll be checked by IsNetworkError)
}
```

**Key Changes**:
1. Rename current `downloadTessdata` → `downloadTessdataAttempt` (single attempt)
2. New `downloadTessdata` wraps attempt in `retry.Do`
3. Wrap HTTP status errors in `HTTPError` type
4. Checksum failures should trigger retry (could be corrupt download)

---

## 5. Potential Risks and Mitigations

### Risk 1: Excessive Retry Delays

**Description**: With 3 attempts and exponential backoff (1s, 2s), total delay could be ~3 seconds on persistent failures.

**Mitigation**:
- Configure reasonable limits: max 3 attempts, max 10s delay
- Total worst case: 1s + 2s + download attempt time ≈ 3-5 seconds
- For 5xx errors, this is acceptable (better than immediate failure)
- Context timeout (5 minutes) provides ultimate backstop

### Risk 2: Checksum Failure Infinite Loop

**Description**: If GitHub serves corrupted data consistently, retry might loop indefinitely.

**Mitigation**:
- Max attempts limit (3) prevents infinite retry
- Checksum failure should be logged with clear message
- After 3 attempts, fail with detailed error including computed hash

### Risk 3: Non-Retryable Errors Delayed

**Description**: 404 errors (invalid language) shouldn't wait for backoff.

**Mitigation**:
- `IsRetryable` checks HTTP status code
- 4xx errors (except 429) return immediately without retry
- No delay on permanent errors

### Risk 4: Context Cancellation Not Respected

**Description**: Long backoff could ignore Ctrl+C.

**Mitigation**:
- Check `ctx.Err()` before each attempt
- Use `select` with `ctx.Done()` during sleep
- Return `ctx.Err()` immediately on cancellation

### Risk 5: Testing Flakiness

**Description**: Real network calls in tests could be flaky.

**Mitigation**:
- Use `httptest.Server` for unit tests (no real network)
- Mock server can simulate:
  - Success on first attempt
  - Failure then success (retry works)
  - Persistent failure (exhaust retries)
  - Different status codes (429, 503, 404)
- Integration tests marked with `testing.Short()` guard

---

## 6. Testing Strategy

### Unit Tests for Retry Package

**File**: `internal/retry/retry_test.go`

**Test Cases**:

1. **TestDoSuccess** - Operation succeeds on first attempt
   - Verify: No retries, no delay

2. **TestDoRetryThenSuccess** - Fails twice, succeeds on third
   - Verify: Backoff delays applied, total attempts = 3

3. **TestDoExhaustRetries** - Fails all attempts
   - Verify: All attempts exhausted, returns last error

4. **TestDoContextCancellation** - Context canceled mid-retry
   - Verify: Returns context.Canceled, stops immediately

5. **TestDoNonRetryableError** - Returns 404 error
   - Verify: No retries, immediate return

6. **TestBackoffCalculation** - Verify exponential backoff math
   - Attempt 1: 1s
   - Attempt 2: 2s
   - Attempt 3: 4s (capped at MaxDelay)

7. **TestIsRetryable** - Test error classification
   - Retryable: 429, 500, 503, timeout, network errors
   - Non-retryable: 404, 400, context.Canceled

### Integration Tests for downloadTessdata

**File**: `internal/ocr/ocr_test.go`

**Test Cases Using httptest.Server**:

1. **TestDownloadTessdataSuccess** - Mock server returns 200 OK
   - Verify: Download succeeds, file written, checksum verified

2. **TestDownloadTessdataRetry503** - Server returns 503 twice, then 200
   - Mock server with counter: attempts 1-2 return 503, attempt 3 returns 200
   - Verify: Download eventually succeeds, 3 HTTP requests made

3. **TestDownloadTessdata404** - Server returns 404
   - Verify: No retry, immediate failure, error message mentions HTTP 404

4. **TestDownloadTessdataExhaustRetries** - Server always returns 503
   - Verify: Fails after 3 attempts, error indicates exhaustion

5. **TestDownloadTessdataContextCancel** - Cancel context during retry
   - Server delays response, cancel context mid-retry
   - Verify: Returns context.Canceled, doesn't exhaust retries

6. **TestDownloadTessdataChecksumRetry** - Corrupt data then good data
   - Server returns data with wrong checksum on first attempt
   - Returns good data on second attempt
   - Verify: Retry triggered by checksum failure, eventually succeeds

**Example Test Pattern**:
```go
func TestDownloadTessdataRetry503(t *testing.T) {
	tmpDir := t.TempDir()

	attemptCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		if attemptCount < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		// Return valid tessdata file content
		w.WriteHeader(http.StatusOK)
		w.Write(validTessdataContent)
	}))
	defer server.Close()

	// Temporarily replace TessdataURL for testing
	originalURL := TessdataURL
	TessdataURL = server.URL
	defer func() { TessdataURL = originalURL }()

	ctx := context.Background()
	err := downloadTessdata(ctx, tmpDir, "eng")

	if err != nil {
		t.Fatalf("Expected success after retry, got: %v", err)
	}

	if attemptCount != 3 {
		t.Errorf("Expected 3 attempts, got %d", attemptCount)
	}
}
```

### Counter-Based Mock Pattern

Based on [testing retry logic with httptest](https://medium.com/@siddharthuncc/go-use-case-mocking-and-testing-http-retries-4783bf9b7e3d):

```go
type MockServer struct {
	attemptCount int
	mu           sync.Mutex
}

func (m *MockServer) Handler(w http.ResponseWriter, r *http.Request) {
	m.mu.Lock()
	m.attemptCount++
	attempt := m.attemptCount
	m.mu.Unlock()

	// Fail first 2 attempts, succeed on 3rd
	if attempt < 3 {
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write(testData)
}
```

---

## 7. Implementation Considerations

### 1. File Organization

**New Files**:
- `internal/retry/retry.go` - Retry logic implementation
- `internal/retry/retry_test.go` - Unit tests for retry logic

**Modified Files**:
- `internal/ocr/ocr.go` - Integrate retry into downloadTessdata
- `internal/ocr/ocr_test.go` - Add integration tests with httptest

### 2. Backward Compatibility

**No Breaking Changes**:
- `downloadTessdata` signature unchanged: `func(ctx context.Context, dataDir, lang string) error`
- Behavior change is additive: failures now retry instead of immediate failure
- UX improvement: users see "retrying download..." messages

### 3. Logging and UX

**User Feedback**:
```go
// In retry.Do, before sleeping:
if attempt < cfg.MaxAttempts {
	fmt.Fprintf(os.Stderr, "Download failed (attempt %d/%d), retrying in %v...\n",
		attempt, cfg.MaxAttempts, delay)
}
```

**Structured Logging**:
```go
// Use internal/logging package for detailed logs
logging.Get().Debug("Retry attempt %d failed: %v", attempt, err)
```

### 4. Performance Impact

**Negligible**:
- Retry logic only runs on failure
- Success path (common case): one extra function call (`retry.Do` wrapper)
- Failure path: adds backoff delays (acceptable for recovery)
- No CPU-intensive operations

### 5. Configuration

**Current Approach**: Hard-coded retry policy in `downloadTessdata`

**Future Enhancement** (out of scope for Phase 6):
- Add retry config to `internal/config/config.go`
- Allow users to customize max attempts, delays
- Environment variable override: `PDF_CLI_RETRY_ATTEMPTS=5`

### 6. Error Messages

**Current**: `"failed to download: HTTP 503"`

**Improved**:
```go
return fmt.Errorf("failed to download %s.traineddata after %d attempts: %w",
	lang, cfg.MaxAttempts, lastErr)
```

**On Retry Exhaustion**:
```
Error: failed to download eng.traineddata after 3 attempts: HTTP 503: Service Unavailable
Download failed (attempt 1/3), retrying in 1s...
Download failed (attempt 2/3), retrying in 2s...
```

---

## 8. Relevant Documentation Links

### Go Standard Library
- [context package](https://pkg.go.dev/context) - Context cancellation and timeouts
- [net/http/httptest](https://pkg.go.dev/net/http/httptest) - HTTP testing utilities
- [time.After](https://pkg.go.dev/time#After) - Backoff delay implementation
- [errors.Is/As](https://pkg.go.dev/errors) - Error type checking

### HTTP Best Practices
- [HTTP Status Codes and Retry](https://www.baeldung.com/cs/http-error-status-codes-retry) - Which codes to retry
- [REST API Retry Best Practices](https://www.restapitutorial.com/advanced/responses/retries) - Retry strategies
- [MDN HTTP Status Codes](https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Status) - Complete reference

### Retry Patterns
- [Exponential Backoff in Go](https://oneuptime.com/blog/post/2026-01-07-go-retry-exponential-backoff/view) - Implementation guide
- [Testing HTTP Retries in Go](https://medium.com/@siddharthuncc/go-use-case-mocking-and-testing-http-retries-4783bf9b7e3d) - httptest patterns
- [httptest Package Documentation](https://pkg.go.dev/net/http/httptest) - Testing tools

### Project-Specific
- `ROADMAP.md` - Phase 6 requirements (lines 202-233)
- `.shipyard/codebase/CONCERNS.md` - Original issue #14 (no retry logic)
- `.shipyard/PROJECT.md` - R12 requirement

---

## 9. Summary and Next Steps

### Recommended Implementation Plan

**Task 1**: Create retry package
- File: `internal/retry/retry.go`
- Implement: `Config`, `Do`, `IsRetryable`, helper functions
- Unit tests: `internal/retry/retry_test.go` (7 test cases)

**Task 2**: Define HTTPError type
- Location: `internal/ocr/ocr.go`
- Purpose: Wrap HTTP status codes for retry logic

**Task 3**: Refactor downloadTessdata
- Split into `downloadTessdata` (retry wrapper) and `downloadTessdataAttempt` (single try)
- Integrate `retry.Do` with configured policy
- Add user-facing retry messages

**Task 4**: Add integration tests
- File: `internal/ocr/ocr_test.go`
- Use `httptest.Server` for mock scenarios
- Test: success, retry-then-success, 404, exhaustion, cancellation, checksum retry

**Task 5**: Update documentation
- Add retry behavior to README
- Update `architecture.md` with retry package

### Success Criteria (from ROADMAP.md)

- [x] Research complete
- [ ] `downloadTessdata` retries up to 3 times on transient HTTP errors (5xx, timeout, connection reset)
- [ ] Non-retryable errors (4xx) fail immediately
- [ ] Retry respects `context.Context` cancellation
- [ ] Unit tests cover retry success on 2nd attempt, exhaustion, and non-retryable errors
- [ ] `go test -race ./...` passes

### Estimated Effort

**Complexity**: S (Small)
**Files Created**: 2 (`retry.go`, `retry_test.go`)
**Files Modified**: 2 (`ocr.go`, `ocr_test.go`)
**Lines of Code**: ~300 total
- Retry package: ~150 LOC
- Tests: ~150 LOC
- Integration: ~50 LOC modifications

### Dependencies

**Requires**:
- Phase 2 (context propagation) ✓ Complete
- Phase 3 (download security) ✓ Complete

**Blocks**:
- None (Phase 7 documentation can proceed in parallel)

---

## Sources

- [HTTP Status Codes and Retry - Baeldung](https://www.baeldung.com/cs/http-error-status-codes-retry)
- [REST API Tutorial - Retry Best Practices](https://www.restapitutorial.com/advanced/responses/retries)
- [How to Implement Retry Logic in Go with Exponential Backoff](https://oneuptime.com/blog/post/2026-01-07-go-retry-exponential-backoff/view)
- [Testing HTTP Retries in Go](https://medium.com/@siddharthuncc/go-use-case-mocking-and-testing-http-retries-4783bf9b7e3d)
- [httptest package documentation](https://pkg.go.dev/net/http/httptest)
- [MDN HTTP Response Status Codes](https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Status)
