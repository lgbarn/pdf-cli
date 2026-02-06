# Plan 2.1: Context Propagation

## Context
This plan threads `context.Context` through all long-running operations to support cancellation and timeouts. Currently, OCR processing, PDF text extraction, and batch operations create their own `context.Background()` instances without accepting caller context, preventing graceful shutdown and timeout handling.

The context flow is: `main` → `signal.NotifyContext` → `cli.ExecuteContext` → `cmd.Context()` → domain functions → backends.

This plan depends on Plan 1.1 completing first, since race testing with context propagation requires stable global state.

## Dependencies
- Plan 1.1: Thread-Safe Singletons (must be completed first)

## Tasks

### Task 1: Add context to OCR package
**Files:** `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr.go`

**Action:** Modify

**Description:**
Thread context through all OCR functions that perform long-running operations:

1. **Update downloadTessdata** (lines 169-209):
   - Change signature: `func downloadTessdata(ctx context.Context, dataDir, lang string) error`
   - Remove lines 175-176 (local context creation)
   - Use `ctx` parameter in `http.NewRequestWithContext(ctx, ...)` at line 178
   - Update call site at line 136: `downloadTessdata(context.TODO(), e.dataDir, lang)`
     - Use TODO() temporarily; will be replaced when EnsureTessdata gets context

2. **Update ExtractTextFromPDF** (lines 211-244):
   - Change signature: `func (e *Engine) ExtractTextFromPDF(ctx context.Context, pdfPath string, pages []int, password string, showProgress bool) (string, error)`
   - Pass `ctx` to `e.processImages(ctx, imageFiles, showProgress)` at line 243

3. **Update processImages** (lines 301-307):
   - Change signature: `func (e *Engine) processImages(ctx context.Context, imageFiles []string, showProgress bool) (string, error)`
   - Pass `ctx` to both branches:
     - `e.processImagesSequential(ctx, imageFiles, showProgress)` at line 304
     - `e.processImagesParallel(ctx, imageFiles, showProgress)` at line 306

4. **Update processImagesSequential** (lines 309-330):
   - Change signature: `func (e *Engine) processImagesSequential(ctx context.Context, imageFiles []string, showProgress bool) (string, error)`
   - Remove line 316 (context.Background() creation)
   - Use `ctx` parameter in backend call at line 320
   - Add context cancellation check in loop after line 319:
     ```go
     if ctx.Err() != nil {
         return "", ctx.Err()
     }
     ```

5. **Update processImagesParallel** (lines 332-374):
   - Change signature: `func (e *Engine) processImagesParallel(ctx context.Context, imageFiles []string, showProgress bool) (string, error)`
   - Remove line 339 (context.Background() creation)
   - Use `ctx` parameter in goroutine at line 353
   - Add context cancellation check in goroutine before processing:
     ```go
     select {
     case <-ctx.Done():
         results <- imageResult{index: idx, text: ""}
         return
     default:
     }
     ```

**Acceptance Criteria:**
- All OCR functions accept context as first parameter
- No internal context.Background() or context.TODO() creation (except temporary EnsureTessdata)
- Context is propagated to backend.ProcessImage() calls
- Cancellation is checked between image processing iterations
- Parallel processing respects context cancellation in goroutines

### Task 2: Add context to PDF text extraction
**Files:** `/Users/lgbarn/Personal/pdf-cli/internal/pdf/text.go`

**Action:** Modify

**Description:**
Thread context through PDF text extraction functions:

1. **Update ExtractTextWithProgress** (lines 21-31):
   - Change signature: `func ExtractTextWithProgress(ctx context.Context, input string, pages []int, password string, showProgress bool) (string, error)`
   - Pass `ctx` to `extractTextPrimary(ctx, input, pages, showProgress)` at line 24
   - Pass `ctx` to `extractTextFallback(ctx, input, pages, password)` at line 30

2. **Update ExtractText** (lines 16-19):
   - Change signature: `func ExtractText(ctx context.Context, input string, pages []int, password string) (string, error)`
   - Pass `ctx` to `ExtractTextWithProgress(ctx, input, pages, password, false)` at line 18

3. **Update extractTextPrimary** (lines 33-63):
   - Change signature: `func extractTextPrimary(ctx context.Context, input string, pages []int, showProgress bool) (string, error)`
   - Pass `ctx` to both extraction paths:
     - `extractPagesParallel(ctx, r, pages, totalPages, showProgress)` at line 59
     - `extractPagesSequential(ctx, r, pages, totalPages, showProgress)` at line 62

4. **Update extractPagesSequential** (lines 65-88):
   - Change signature: `func extractPagesSequential(ctx context.Context, r *pdf.Reader, pages []int, totalPages int, showProgress bool) (string, error)`
   - Add context check in loop after line 74:
     ```go
     if ctx.Err() != nil {
         return "", ctx.Err()
     }
     ```

5. **Update extractPagesParallel** (lines 106-150):
   - Change signature: `func extractPagesParallel(ctx context.Context, r *pdf.Reader, pages []int, totalPages int, showProgress bool) (string, error)`
   - Add context check in goroutine at line 122:
     ```go
     go func(pn int) {
         select {
         case <-ctx.Done():
             results <- pageResult{pageNum: pn, text: ""}
             return
         default:
         }
         results <- pageResult{pageNum: pn, text: extractPageText(r, pn, totalPages)}
     }(pageNum)
     ```

6. **Update extractTextFallback** (lines 152-190):
   - Change signature: `func extractTextFallback(ctx context.Context, input string, pages []int, password string) (string, error)`
   - Add context check before pdfcpu call at line 160:
     ```go
     if ctx.Err() != nil {
         return "", ctx.Err()
     }
     ```

**Note:** pdfcpu does not support context, so we add checks between operations as a best-effort cancellation mechanism.

**Acceptance Criteria:**
- All text extraction functions accept context as first parameter
- Context is propagated through all extraction code paths
- Cancellation is checked between page processing iterations
- Parallel goroutines respect context cancellation
- Fallback path checks context before calling pdfcpu

### Task 3: Wire context from CLI to domain layer
**Files:**
- `/Users/lgbarn/Personal/pdf-cli/cmd/pdf/main.go`
- `/Users/lgbarn/Personal/pdf-cli/internal/cli/cli.go`
- `/Users/lgbarn/Personal/pdf-cli/internal/commands/text.go`
- `/Users/lgbarn/Personal/pdf-cli/internal/commands/helpers.go`

**Action:** Modify

**Description:**
Connect OS signals to context and thread through command layer:

**1. Update main.go** (lines 17-22):
```go
import (
    "context"
    "os"
    "os/signal"
    "syscall"

    "github.com/lgbarn/pdf-cli/internal/cli"
    _ "github.com/lgbarn/pdf-cli/internal/commands"
)

func main() {
    // Create context that cancels on interrupt/termination signals
    ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
    defer stop()

    cli.SetVersion(version, commit, date)
    if err := cli.ExecuteContext(ctx); err != nil {
        os.Exit(1)
    }
}
```

**2. Update cli.go** (lines 68-71):
```go
// Execute runs the root command with background context (for testing)
func Execute() error {
    return rootCmd.Execute()
}

// ExecuteContext runs the root command with the given context
func ExecuteContext(ctx context.Context) error {
    return rootCmd.ExecuteContext(ctx)
}
```

**3. Update commands/text.go** (line 93 and 100):
- Change OCR call: `text, err = engine.ExtractTextFromPDF(cmd.Context(), inputFile, pages, password, cli.Progress())`
- Change PDF call: `text, err = pdf.ExtractTextWithProgress(cmd.Context(), inputFile, pages, password, cli.Progress())`

**4. Update commands/helpers.go** (lines 64-74):
```go
// processBatch processes multiple files with the given processor function.
// Contexts should be retrieved from cmd.Context() by the processor.
func processBatch(files []string, processor func(file string) error) error {
    var errs []error
    for _, file := range files {
        if err := processor(file); err != nil {
            errs = append(errs, fmt.Errorf("%s: %w", file, err))
        }
    }
    return errors.Join(errs...)
}
```

Note: processBatch doesn't need context parameter since processors access cmd.Context() directly. This avoids breaking all existing command files.

**Acceptance Criteria:**
- main.go creates signal-aware context that cancels on SIGINT/SIGTERM
- cli.ExecuteContext passes context to Cobra's ExecuteContext
- text command uses cmd.Context() for both OCR and PDF extraction
- Ctrl+C during long operations triggers graceful shutdown
- No changes needed to other command files (they don't have long operations yet)

## Verification

**Command sequence:**
```bash
# Verify compilation
go build ./cmd/pdf

# Run unit tests with race detection
go test -race ./internal/ocr/... -v
go test -race ./internal/pdf/... -v
go test -race ./internal/commands/... -v

# Run full test suite
go test -race ./... -short

# Manual verification: test cancellation
echo "Create a large test PDF and run:"
echo "  pdf text large.pdf --ocr"
echo "Press Ctrl+C during processing - should exit gracefully"
```

**Success criteria:**
- All tests pass with zero data races
- Binary builds successfully
- Signal handling works: Ctrl+C during OCR or text extraction exits cleanly
- No timeout regressions (operations complete normally when not cancelled)
- CLI behavior unchanged when operations complete without interruption
