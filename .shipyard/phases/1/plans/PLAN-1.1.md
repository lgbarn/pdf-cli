# Plan 1.1: HTTP Client and Timer Hardening

## Context
This plan addresses R6 and R10 by replacing unsafe resource management patterns in the OCR download path with properly managed alternatives. These changes are independent of each other and introduce no breaking API changes.

- R6: Replace `http.DefaultClient` with a custom client with 5-minute timeout
- R10: Replace `time.After` with `time.NewTimer` + explicit `Stop()` to prevent resource leaks

## Dependencies
None. This is a Wave 1 plan with no dependencies.

## Tasks

### Task 1: Replace http.DefaultClient with custom client
**Files:**
- `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr.go`

**Action:** modify

**Description:**
1. Add a package-level variable at the top of `ocr.go` (after imports, before any functions):
   ```go
   var tessdataHTTPClient = &http.Client{
       Timeout: 5 * time.Minute,
   }
   ```
2. At line 260 in `downloadTessdata`, replace:
   ```go
   resp, err := http.DefaultClient.Do(req)
   ```
   with:
   ```go
   resp, err := tessdataHTTPClient.Do(req)
   ```

**Acceptance Criteria:**
- `http.DefaultClient` does not appear in `ocr.go`
- Package-level `tessdataHTTPClient` is defined with 5-minute timeout
- `downloadTessdata` uses `tessdataHTTPClient.Do(req)`
- Existing tests pass: `go test ./internal/ocr/...`

### Task 2: Replace time.After with time.NewTimer
**Files:**
- `/Users/lgbarn/Personal/pdf-cli/internal/retry/retry.go`

**Action:** modify

**Description:**
At line 76 in the `RetryWithBackoff` function, replace the existing `time.After` select case pattern:
```go
case <-time.After(delay):
```

With a properly managed timer:
```go
timer := time.NewTimer(delay)
select {
case <-ctx.Done():
    timer.Stop()
    return fmt.Errorf("context cancelled during retry backoff: %w", ctx.Err())
case <-timer.C:
}
```

Ensure the timer is stopped in the ctx.Done() case to prevent resource leaks.

**Acceptance Criteria:**
- `time.After` does not appear in `retry.go`
- Timer is created with `time.NewTimer(delay)`
- Timer is explicitly stopped with `timer.Stop()` in the context cancellation path
- Existing tests pass: `go test ./internal/ocr/...`
- No new goroutine leaks (verified by test suite)

### Task 3: Verify integrated behavior
**Files:**
- `/Users/lgbarn/Personal/pdf-cli/internal/ocr/`

**Action:** test

**Description:**
Run the full OCR test suite to verify both changes work correctly together:
1. Run all OCR tests: `go test -v ./internal/ocr/...`
2. Run with race detector: `go test -race ./internal/ocr/...`
3. Verify no regression in coverage: `go test -cover ./internal/ocr/...`

**Acceptance Criteria:**
- All OCR tests pass
- No race conditions detected
- Test coverage remains at or above baseline
- No new lint warnings from `golangci-lint run ./internal/ocr/...`

## Verification
```bash
# Verify http.DefaultClient is replaced
cd /Users/lgbarn/Personal/pdf-cli
grep -n "http.DefaultClient" internal/ocr/ocr.go || echo "✓ No http.DefaultClient found"
grep -n "tessdataHTTPClient" internal/ocr/ocr.go | grep -q "Timeout.*5.*time.Minute" && echo "✓ Custom client configured"

# Verify time.After is replaced
grep -n "time.After" internal/retry/retry.go || echo "✓ No time.After found"
grep -n "time.NewTimer" internal/retry/retry.go | grep -q "NewTimer" && echo "✓ NewTimer used"
grep -n "timer.Stop()" internal/retry/retry.go | grep -q "Stop" && echo "✓ Timer stopped"

# Run tests
go test -v -race ./internal/ocr/...
```
