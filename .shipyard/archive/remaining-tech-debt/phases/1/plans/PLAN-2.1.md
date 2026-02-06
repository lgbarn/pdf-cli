# Plan 2.1: Context Propagation and Progress Bar Fixes

## Context
This plan addresses R4 and R12 by propagating context through the OCR initialization path and fixing progress bar recreation during retry attempts. These changes require API signature updates and must come after the HTTP/timer hardening to avoid merge conflicts.

- R4: Replace `context.TODO()` with proper context propagation through `EnsureTessdata` methods
- R12: Recreate progress bar per retry attempt during tessdata downloads

## Dependencies
- Plan 1.1 (HTTP Client and Timer Hardening) must complete first to avoid conflicts in `ocr.go`

## Tasks

### Task 1: Propagate context through EnsureTessdata API
**Files:**
- `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr.go`
- `/Users/lgbarn/Personal/pdf-cli/internal/ocr/wasm.go`
- `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr_test.go`
- `/Users/lgbarn/Personal/pdf-cli/internal/ocr/engine_extended_test.go`
- `/Users/lgbarn/Personal/pdf-cli/internal/ocr/wasm_test.go`

**Action:** refactor

**Description:**
1. Update `Engine.EnsureTessdata` signature in `ocr.go` line ~175:
   - Change from: `func (e *Engine) EnsureTessdata() error`
   - Change to: `func (e *Engine) EnsureTessdata(ctx context.Context) error`
   - Replace `downloadTessdata(context.TODO(), ...)` with `downloadTessdata(ctx, ...)`

2. Update `WASMBackend.EnsureTessdata` signature in `wasm.go` line ~53:
   - Change from: `func (w *WASMBackend) EnsureTessdata(lang string) error`
   - Change to: `func (w *WASMBackend) EnsureTessdata(ctx context.Context, lang string) error`
   - Replace `downloadTessdata(context.TODO(), ...)` with `downloadTessdata(ctx, ...)`

3. Update production callers:
   - `ocr.go:329` in `ExtractTextFromPDF`: change `e.EnsureTessdata()` to `e.EnsureTessdata(ctx)`
   - `wasm.go:71` in `initializeTesseract`: change `w.EnsureTessdata(lang)` to `w.EnsureTessdata(ctx, lang)`

4. Update test callers to use `context.Background()`:
   - `ocr_test.go:114`
   - `engine_extended_test.go:31`
   - `engine_extended_test.go:55`
   - `wasm_test.go:83`
   - `wasm_test.go:88`

**Acceptance Criteria:**
- `context.TODO()` does not appear in `ocr.go` or `wasm.go`
- Both `EnsureTessdata` methods accept `context.Context` as first parameter
- All production callers pass their existing context
- All test callers use `context.Background()`
- Code compiles without errors: `go build ./...`
- All tests pass: `go test ./internal/ocr/...`

### Task 2: Recreate progress bar per retry attempt
**Files:**
- `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr.go`

**Action:** modify

**Description:**
Refactor `downloadTessdata` to recreate the progress bar on each retry attempt instead of once before the retry loop.

Current structure (lines 234-277):
```go
bar := progressbar.DefaultBytes(...)
err := RetryWithBackoff(ctx, ..., func() error {
    // download logic using bar
})
```

New structure:
1. Move progress bar creation inside the retry function
2. Finish the previous bar before creating a new one on retry
3. Pattern:
   ```go
   var bar *progressbar.ProgressBar
   err := RetryWithBackoff(ctx, ..., func() error {
       // Finish previous bar if it exists
       if bar != nil {
           _ = bar.Finish()
       }

       // Create new bar for this attempt
       bar = progressbar.DefaultBytes(
           resp.ContentLength,
           fmt.Sprintf("Downloading %s", filename),
       )

       // existing download logic...
   })
   ```

**Acceptance Criteria:**
- Progress bar is created inside the retry function, not outside
- Previous progress bar is finished before creating new one on retry
- Progress bar correctly shows download progress for each attempt
- Tests pass: `go test ./internal/ocr/...`
- Manual verification: trigger a retry (e.g., network flake) and observe new progress bar

### Task 3: Verify context cancellation and progress bar behavior
**Files:**
- `/Users/lgbarn/Personal/pdf-cli/internal/ocr/`

**Action:** test

**Description:**
Comprehensive verification of context propagation and progress bar behavior:

1. Run full test suite: `go test -v ./internal/ocr/...`
2. Run with race detector: `go test -race ./internal/ocr/...`
3. Verify context cancellation works:
   - Create a test that cancels context during download
   - Verify download stops promptly
4. Verify no context leaks:
   ```bash
   go test -v ./internal/ocr/... 2>&1 | grep -i "context"
   ```
5. Run golangci-lint: `golangci-lint run ./internal/ocr/...`

**Acceptance Criteria:**
- All tests pass with `-race` flag
- Context cancellation terminates download operations
- No "context leak" warnings
- No new lint violations
- `grep -r "context.TODO()" internal/ocr/` returns no matches in production code

## Verification
```bash
cd /Users/lgbarn/Personal/pdf-cli

# Verify context.TODO removal
echo "Checking for context.TODO in production code..."
grep -n "context.TODO()" internal/ocr/ocr.go && echo "✗ Found context.TODO" || echo "✓ No context.TODO in ocr.go"
grep -n "context.TODO()" internal/ocr/wasm.go && echo "✗ Found context.TODO" || echo "✓ No context.TODO in wasm.go"

# Verify signature changes
echo "Checking EnsureTessdata signatures..."
grep -n "func.*EnsureTessdata(ctx context.Context)" internal/ocr/ocr.go && echo "✓ Engine.EnsureTessdata has ctx param"
grep -n "func.*EnsureTessdata(ctx context.Context, lang string)" internal/ocr/wasm.go && echo "✓ WASMBackend.EnsureTessdata has ctx param"

# Verify progress bar is inside retry function
echo "Checking progress bar placement..."
grep -B5 -A5 "progressbar.DefaultBytes" internal/ocr/ocr.go | grep -q "func() error" && echo "✓ Progress bar inside retry function"

# Run tests
echo "Running tests..."
go test -v -race ./internal/ocr/...

# Final lint check
echo "Running linter..."
golangci-lint run ./internal/ocr/...
```
