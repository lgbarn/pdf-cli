---
phase: network-resilience
plan: 1.1
wave: 1
dependencies: []
must_haves:
  - R12: Network operations (tessdata download) should have retry logic with backoff
files_touched:
  - internal/retry/retry.go
  - internal/retry/retry_test.go
  - internal/ocr/ocr.go
  - internal/ocr/ocr_test.go
tdd: true
---

# Plan 1.1: Network Retry with Exponential Backoff

## Context

The `downloadTessdata` function in `internal/ocr/ocr.go` has no retry logic. Any transient network failure (timeout, connection reset, HTTP 5xx) requires the user to manually re-run the command. This plan creates a generic retry helper and integrates it with the download path.

## Tasks

<task id="1" files="internal/retry/retry.go,internal/retry/retry_test.go" tdd="true">
  <action>
    Create a generic retry package at `internal/retry/retry.go`:

    ```go
    package retry

    import (
        "context"
        "errors"
        "math"
        "time"
    )

    const (
        // DefaultMaxAttempts is the default number of attempts before giving up.
        DefaultMaxAttempts = 3

        // DefaultBaseDelay is the initial delay between retries.
        DefaultBaseDelay = 1 * time.Second

        // DefaultMaxDelay caps the backoff delay.
        DefaultMaxDelay = 30 * time.Second
    )

    // PermanentError wraps an error to indicate it should not be retried.
    type PermanentError struct {
        Err error
    }

    func (e *PermanentError) Error() string { return e.Err.Error() }
    func (e *PermanentError) Unwrap() error { return e.Err }

    // Permanent wraps an error to mark it as non-retryable.
    func Permanent(err error) error { return &PermanentError{Err: err} }

    // IsPermanent checks if an error is marked as permanent.
    func IsPermanent(err error) bool {
        var pe *PermanentError
        return errors.As(err, &pe)
    }

    // Options configures retry behavior.
    type Options struct {
        MaxAttempts int           // Max number of attempts (0 = use default)
        BaseDelay   time.Duration // Initial delay (0 = use default)
        MaxDelay    time.Duration // Maximum delay cap (0 = use default)
    }

    // Do executes fn with retry logic. fn should return Permanent(err) for
    // non-retryable errors. Respects context cancellation.
    func Do(ctx context.Context, opts Options, fn func(ctx context.Context) error) error {
        maxAttempts := opts.MaxAttempts
        if maxAttempts <= 0 {
            maxAttempts = DefaultMaxAttempts
        }
        baseDelay := opts.BaseDelay
        if baseDelay <= 0 {
            baseDelay = DefaultBaseDelay
        }
        maxDelay := opts.MaxDelay
        if maxDelay <= 0 {
            maxDelay = DefaultMaxDelay
        }

        var lastErr error
        for attempt := 0; attempt < maxAttempts; attempt++ {
            if ctx.Err() != nil {
                return ctx.Err()
            }

            lastErr = fn(ctx)
            if lastErr == nil {
                return nil
            }

            if IsPermanent(lastErr) {
                return lastErr
            }

            // Don't sleep after the last attempt
            if attempt < maxAttempts-1 {
                delay := time.Duration(math.Pow(2, float64(attempt))) * baseDelay
                if delay > maxDelay {
                    delay = maxDelay
                }
                select {
                case <-time.After(delay):
                case <-ctx.Done():
                    return ctx.Err()
                }
            }
        }
        return lastErr
    }
    ```

    Create `internal/retry/retry_test.go` with these tests:

    1. **TestDoSuccess** — fn succeeds on first attempt, no retries
    2. **TestDoRetryThenSuccess** — fn fails once, succeeds on second attempt
    3. **TestDoExhaustion** — fn fails on all attempts, returns last error
    4. **TestDoPermanentError** — fn returns Permanent(err), no retry
    5. **TestDoContextCancellation** — context canceled mid-retry, returns ctx.Err()
    6. **TestDoBackoffTiming** — verify delays increase exponentially (use time.Since checks with tolerance)
    7. **TestIsPermanent** — verify IsPermanent detects wrapped permanent errors
  </action>
  <verify>
    ```bash
    cd /Users/lgbarn/Personal/pdf-cli/.worktrees/phase-2-concurrency
    go test -v -race ./internal/retry/...
    ```
  </verify>
  <done>
    - internal/retry package created with Do(), Permanent(), IsPermanent()
    - Exponential backoff with configurable base/max delay
    - Context cancellation respected between retries
    - PermanentError type for non-retryable errors
    - All 7 unit tests pass with race detector
  </done>
</task>

<task id="2" files="internal/ocr/ocr.go,internal/ocr/ocr_test.go" tdd="true">
  <action>
    Integrate retry logic into `downloadTessdata`:

    1. Add import `"github.com/lgbarn/pdf-cli/internal/retry"` to ocr.go

    2. Add constants for retry configuration:
       ```go
       const (
           DefaultRetryAttempts = 3
           DefaultRetryBaseDelay = 1 * time.Second
       )
       ```

    3. Wrap the HTTP request portion of `downloadTessdata` with `retry.Do`. The retry should cover:
       - Creating the HTTP request
       - Executing the request
       - Checking the status code
       - Reading the response body into the temp file

    4. Mark non-retryable errors with `retry.Permanent()`:
       - HTTP 4xx responses (client errors) — except 429 (Too Many Requests)
       - Path sanitization errors
       - Context errors

    5. Retryable errors (just return the error):
       - Network errors (connection refused, timeout, DNS)
       - HTTP 429 (rate limited)
       - HTTP 5xx (server errors)

    **Important implementation notes:**
    - The progress bar, checksum verification, and file rename should happen OUTSIDE the retry loop (after successful download)
    - Each retry attempt needs a fresh temp file write (truncate or recreate)
    - The overall context timeout (DefaultDownloadTimeout) should cover ALL attempts, not per-attempt
    - Log retry attempts to stderr: `fmt.Fprintf(os.Stderr, "Download failed (attempt %d/%d): %v. Retrying...\n", ...)`

    Refactored structure:
    ```go
    func downloadTessdata(ctx context.Context, dataDir, lang string) (err error) {
        // ... path setup, sanitization (unchanged) ...

        ctx, cancel := context.WithTimeout(ctx, DefaultDownloadTimeout)
        defer cancel()

        tmpFile, err := os.CreateTemp(dataDir, "tessdata-*.tmp")
        // ... temp file setup (unchanged) ...

        hasher := sha256.New()
        bar := progress.NewBytesProgressBar(...)

        var attempt int
        err = retry.Do(ctx, retry.Options{MaxAttempts: DefaultRetryAttempts, BaseDelay: DefaultRetryBaseDelay}, func(ctx context.Context) error {
            attempt++

            // Reset temp file and hasher for each attempt
            if _, err := tmpFile.Seek(0, io.SeekStart); err != nil {
                return retry.Permanent(err)
            }
            if err := tmpFile.Truncate(0); err != nil {
                return retry.Permanent(err)
            }
            hasher.Reset()

            req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
            if err != nil {
                return retry.Permanent(err)
            }

            resp, err := http.DefaultClient.Do(req)
            if err != nil {
                if attempt < DefaultRetryAttempts {
                    fmt.Fprintf(os.Stderr, "Download failed (attempt %d/%d): %v. Retrying...\n", attempt, DefaultRetryAttempts, err)
                }
                return err // retryable
            }
            defer resp.Body.Close()

            if resp.StatusCode == http.StatusOK {
                if _, err := io.Copy(io.MultiWriter(tmpFile, bar, hasher), resp.Body); err != nil {
                    return err // retryable (network error during body read)
                }
                return nil
            }

            // 429 Too Many Requests — retryable
            if resp.StatusCode == http.StatusTooManyRequests {
                return fmt.Errorf("rate limited: HTTP %d", resp.StatusCode)
            }

            // 5xx — retryable
            if resp.StatusCode >= 500 {
                return fmt.Errorf("server error: HTTP %d", resp.StatusCode)
            }

            // 4xx — permanent
            return retry.Permanent(fmt.Errorf("failed to download: HTTP %d", resp.StatusCode))
        })
        if err != nil {
            return err
        }

        // ... close, checksum verification, rename (unchanged) ...
    }
    ```

    6. Add/update tests in `internal/ocr/ocr_test.go`:
       - **TestDownloadTessdataRetryOnServerError** — httptest server returns 503 once then 200, verify download succeeds
       - **TestDownloadTessdataNoRetryOn404** — httptest server returns 404, verify immediate failure (no retry)
       - **TestDownloadTessdataRetryExhaustion** — httptest server always returns 500, verify failure after 3 attempts
  </action>
  <verify>
    ```bash
    cd /Users/lgbarn/Personal/pdf-cli/.worktrees/phase-2-concurrency
    go test -v -race -run "TestDownload|TestRetry" ./internal/ocr/...
    go test -race ./... -short -count=1
    ```
  </verify>
  <done>
    - downloadTessdata wrapped with retry.Do
    - HTTP 5xx and network errors trigger retry with exponential backoff
    - HTTP 4xx (except 429) fails immediately
    - Retry attempts logged to stderr
    - Progress bar and checksum verification outside retry loop
    - Unit tests verify retry, no-retry, and exhaustion scenarios
    - All tests pass with race detector
  </done>
</task>

## Verification

```bash
cd /Users/lgbarn/Personal/pdf-cli/.worktrees/phase-2-concurrency

# Unit tests
go test -v -race ./internal/retry/...
go test -v -race -run "TestDownload|TestRetry" ./internal/ocr/...

# Full test suite
go test -race ./... -short -count=1

# Lint
golangci-lint run ./...
```

## Success Criteria

- Generic retry.Do helper exists in internal/retry/ with exponential backoff
- downloadTessdata retries up to 3 times on transient HTTP errors (5xx, timeout, connection reset)
- Non-retryable errors (4xx except 429) fail immediately via retry.Permanent
- Retry respects context.Context cancellation
- Unit tests cover retry success on 2nd attempt, exhaustion, and non-retryable errors
- go test -race ./... passes
- No new external dependencies added
