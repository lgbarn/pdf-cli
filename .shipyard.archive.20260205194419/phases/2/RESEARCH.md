# Phase 2 Research: Thread Safety and Context Propagation

**Date**: 2026-01-31
**Phase**: 2 - Thread Safety and Context Propagation
**Status**: Research Complete

## Executive Summary

This document provides a comprehensive analysis of the current state of global singleton patterns and context propagation in pdf-cli, along with specific recommendations for making the codebase thread-safe and cancellation-aware.

**Key Findings**:
- Both `config` and `logging` packages use unsafe lazy initialization (bare nil-check without synchronization)
- No context propagation exists in long-running operations (OCR, text extraction, downloads, batch processing)
- OCR backend interface already accepts `context.Context`, providing a foundation to build on
- Cobra provides `cmd.Context()` method (available since v1.8.0, using v1.10.2)
- No existing data races detected by `go test -race ./...`
- Parallel processing in OCR and text extraction lacks cancellation support

---

## 1. Config Package Analysis

### Current State

**File**: `/Users/lgbarn/Personal/pdf-cli/internal/config/config.go`

#### Global State (Lines 141-160)

```go
// global holds the loaded configuration
var global *Config

// Get returns the global configuration, loading it if necessary.
func Get() *Config {
	if global == nil {  // RACE CONDITION: Bare nil-check
		var err error
		global, err = Load()
		if err != nil {
			// Fall back to defaults on error
			global = DefaultConfig()
		}
	}
	return global
}

// Reset clears the global config (useful for testing).
func Reset() {
	global = nil  // RACE CONDITION: Unsynchronized write
}
```

#### Race Conditions Identified

1. **Check-Then-Act Race in Get()** (Line 146-154):
   - Two goroutines can both see `global == nil`
   - Both will call `Load()` and create duplicate configs
   - One config will be overwritten, wasting work
   - No memory barrier ensures visibility across goroutines

2. **Write Race in Reset()** (Line 159):
   - If `Reset()` is called while `Get()` is executing, `Get()` might:
     - Return nil (if Reset happens after the check but before return)
     - Use a partially constructed config
     - Panic with nil pointer dereference

3. **Test Interference**:
   - Tests in `/Users/lgbarn/Personal/pdf-cli/internal/config/config_test.go` call `Reset()` at line 30, 54, 96, 132, 248
   - Running tests with `-parallel` could cause flaky failures
   - Evidence: TestGet (line 129-141) verifies singleton behavior but isn't race-safe

#### Test Patterns

```go
// Line 129-141: TestGet expects singleton behavior
func TestGet(t *testing.T) {
	os.Setenv("XDG_CONFIG_HOME", "/nonexistent/path")
	defer os.Unsetenv("XDG_CONFIG_HOME")
	Reset()  // UNSAFE in parallel tests

	cfg1 := Get()
	cfg2 := Get()

	// Should return the same instance
	if cfg1 != cfg2 {
		t.Error("Get should return the same config instance")
	}
}
```

### Recommended Fix

Use `sync.Once` for initialization and `sync.RWMutex` for reset capability:

```go
var (
	global     *Config
	globalOnce sync.Once
	globalMu   sync.RWMutex
)

func Get() *Config {
	globalMu.RLock()
	if global != nil {
		defer globalMu.RUnlock()
		return global
	}
	globalMu.RUnlock()

	// Upgrade to write lock for initialization
	globalMu.Lock()
	defer globalMu.Unlock()

	// Double-check after acquiring write lock
	if global == nil {
		var err error
		global, err = Load()
		if err != nil {
			global = DefaultConfig()
		}
	}
	return global
}

func Reset() {
	globalMu.Lock()
	defer globalMu.Unlock()
	global = nil
}
```

**Alternative**: Use `sync.Once` alone if `Reset()` is only needed for testing:

```go
var (
	global     *Config
	globalOnce sync.Once
)

func Get() *Config {
	globalOnce.Do(func() {
		var err error
		global, err = Load()
		if err != nil {
			global = DefaultConfig()
		}
	})
	return global
}

func Reset() {
	// For testing only - not safe for concurrent access
	// Tests should use t.Setenv and avoid parallel execution
	global = nil
	globalOnce = sync.Once{}
}
```

**Recommendation**: Use the RWMutex approach for full safety, or document that Reset() is test-only and tests must not run in parallel.

---

## 2. Logging Package Analysis

### Current State

**File**: `/Users/lgbarn/Personal/pdf-cli/internal/logging/logger.go`

#### Global State (Lines 82-130)

```go
// global is the global logger instance.
var global *Logger

// Init initializes the global logger with the given level and format.
func Init(level Level, format Format) {
	global = New(level, format, os.Stderr)  // RACE: Unsynchronized write
}

// Get returns the global logger, initializing with defaults if needed.
func Get() *Logger {
	if global == nil {  // RACE: Bare nil-check
		Init(LevelSilent, FormatText)
	}
	return global
}

// Reset resets the global logger (for testing).
func Reset() {
	global = nil  // RACE: Unsynchronized write
}
```

#### Race Conditions Identified

1. **Init() Race** (Line 86-88):
   - Multiple goroutines calling `Init()` concurrently will corrupt `global`
   - No synchronization on write operation
   - Called from `cli.InitLogging()` in rootCmd.PersistentPreRun (line 63 of cli.go)

2. **Get() Race** (Line 120-125):
   - Same check-then-act race as config package
   - Two goroutines can both call `Init()` if `global == nil`

3. **Reset() Race** (Line 128-130):
   - Same issues as config.Reset()
   - Used in 10+ test functions (logger_test.go lines 188, 207, 224, 241, 270)

4. **Package-Level Functions** (Lines 143-161):
   - `Debug()`, `Info()`, `Warn()`, `Error()`, `With()` all call `Get()`
   - Each call risks triggering the race condition

#### Test Patterns

```go
// Line 187-208: TestGlobalLogger
func TestGlobalLogger(t *testing.T) {
	Reset()  // UNSAFE in parallel tests

	l := Get()
	if l == nil {
		t.Fatal("Get() should return a logger")
	}

	// Should be silent by default
	if l.Level() != LevelSilent {
		t.Errorf("Default level should be silent, got: %v", l.Level())
	}

	// Subsequent calls should return the same instance
	l2 := Get()
	if l != l2 {
		t.Error("Get() should return the same instance")
	}

	Reset()  // Another unsafe reset
}
```

### Recommended Fix

Use `sync.Once` for initialization and `sync.RWMutex` for Init/Reset:

```go
var (
	global   *Logger
	globalMu sync.RWMutex
)

func Init(level Level, format Format) {
	globalMu.Lock()
	defer globalMu.Unlock()
	global = New(level, format, os.Stderr)
}

func Get() *Logger {
	globalMu.RLock()
	if global != nil {
		defer globalMu.RUnlock()
		return global
	}
	globalMu.RUnlock()

	// Lazy init with defaults
	globalMu.Lock()
	defer globalMu.Unlock()
	if global == nil {
		global = New(LevelSilent, FormatText, os.Stderr)
	}
	return global
}

func Reset() {
	globalMu.Lock()
	defer globalMu.Unlock()
	global = nil
}
```

**Important**: The `Init()` function is called from `rootCmd.PersistentPreRun` (cli.go line 63), which runs before every command. This means `Init()` must be safe for concurrent access if multiple commands could theoretically run in the same process.

---

## 3. OCR Package Analysis

### Current State

**File**: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr.go`

#### Context Usage Status

**Good News**: The OCR backend interface already accepts `context.Context`:

```go
// Line 8-13: Backend interface
type Backend interface {
	Name() string
	Available() bool
	ProcessImage(ctx context.Context, imagePath, lang string) (string, error)  // ✓ Has context
	Close() error
}
```

Both implementations properly use context:
- `NativeBackend.ProcessImage()` (native.go line 46): Uses `exec.CommandContext(ctx, ...)`
- `WASMBackend.ProcessImage()` (wasm.go line 95): Passes `ctx` to gogosseract methods

**Problem**: The high-level functions don't accept or propagate context:

#### Functions Needing Context (ocr.go)

1. **downloadTessdata()** (Lines 169-209) - Network I/O
   - Currently creates its own context: `ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)` (line 175)
   - Hard-coded 5-minute timeout
   - No cancellation from caller
   - Should accept `ctx context.Context` as first parameter

2. **ExtractTextFromPDF()** (Lines 211-244) - Long-running operation
   ```go
   func (e *Engine) ExtractTextFromPDF(pdfPath string, pages []int, password string, showProgress bool) (string, error)
   ```
   - Main OCR entry point
   - Calls `processImages()` which may take minutes
   - Should be: `func (e *Engine) ExtractTextFromPDF(ctx context.Context, pdfPath string, pages []int, password string, showProgress bool) (string, error)`

3. **processImagesSequential()** (Lines 309-330)
   - Currently creates context: `ctx := context.Background()` (line 316)
   - Calls `backend.ProcessImage(ctx, ...)` (line 320)
   - Should accept context from caller

4. **processImagesParallel()** (Lines 332-374) - Parallel batch processing
   - Currently creates context: `ctx := context.Background()` (line 339)
   - Launches goroutines without cancellation (lines 347-356)
   - If parent context is cancelled, goroutines continue running
   - Should respect context cancellation and stop launching new workers

#### Proposed Signatures

```go
// downloadTessdata downloads a tessdata file with cancellation support
func downloadTessdata(ctx context.Context, dataDir, lang string) error

// ExtractTextFromPDF extracts text from a PDF using OCR with cancellation support
func (e *Engine) ExtractTextFromPDF(ctx context.Context, pdfPath string, pages []int, password string, showProgress bool) (string, error)

// processImages processes extracted images with cancellation support
func (e *Engine) processImages(ctx context.Context, imageFiles []string, showProgress bool) (string, error)

// processImagesSequential processes images one by one with cancellation
func (e *Engine) processImagesSequential(ctx context.Context, imageFiles []string, showProgress bool) (string, error)

// processImagesParallel processes images in parallel with cancellation
func (e *Engine) processImagesParallel(ctx context.Context, imageFiles []string, showProgress bool) (string, error)
```

#### Parallel Processing Concerns (Lines 332-374)

Current implementation launches goroutines without context awareness:

```go
// Line 347-356: Goroutines don't check context
for i, imgPath := range imageFiles {
	wg.Add(1)
	sem <- struct{}{} // Acquire semaphore
	go func(idx int, path string) {
		defer wg.Done()
		defer func() { <-sem }() // Release semaphore
		text, _ := e.backend.ProcessImage(ctx, path, e.lang)  // ctx is background
		results <- imageResult{index: idx, text: text}
	}(i, imgPath)
}
```

**Issues**:
- If context is cancelled after launching goroutines, they continue processing
- No early termination of the loop if context is done
- Progress bar continues even if operation should stop

**Fix**: Check context before launching each goroutine and add context-aware collection:

```go
for i, imgPath := range imageFiles {
	// Check if context is cancelled before launching new work
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	wg.Add(1)
	sem <- struct{}{}
	go func(idx int, path string) {
		defer wg.Done()
		defer func() { <-sem }()
		// ProcessImage already respects context
		text, _ := e.backend.ProcessImage(ctx, path, e.lang)
		results <- imageResult{index: idx, text: text}
	}(i, imgPath)
}
```

---

## 4. PDF Package Analysis

### Current State

**Files**: `/Users/lgbarn/Personal/pdf-cli/internal/pdf/text.go`, `/Users/lgbarn/Personal/pdf-cli/internal/pdf/transform.go`

#### Functions Needing Context

The PDF package has several long-running operations that should accept context:

**text.go**:
1. **ExtractTextWithProgress()** (Lines 22-31) - May take seconds to minutes for large PDFs
   - Current: `func ExtractTextWithProgress(input string, pages []int, password string, showProgress bool) (string, error)`
   - Proposed: `func ExtractTextWithProgress(ctx context.Context, input string, pages []int, password string, showProgress bool) (string, error)`

2. **extractPagesParallel()** (Lines 107-150) - Parallel processing
   - Currently launches goroutines without cancellation (lines 122-125)
   - Should check context and propagate to workers

**transform.go**:
1. **MergeWithProgress()** (Lines 21-70) - Incremental merge of multiple files
   - Current: `func MergeWithProgress(inputs []string, output, password string, showProgress bool) error`
   - Proposed: `func MergeWithProgress(ctx context.Context, inputs []string, output, password string, showProgress bool) error`
   - Merge loop (lines 54-64) should check context each iteration

2. **SplitWithProgress()** (Lines 83-134) - Split large PDF into chunks
   - Current: `func SplitWithProgress(input, outputDir string, pageCount int, password string, showProgress bool) error`
   - Proposed: `func SplitWithProgress(ctx context.Context, input, outputDir string, pageCount int, password string, showProgress bool) error`
   - Split loop (lines 104-130) should check context each iteration

**Note**: The underlying pdfcpu library (v0.11.1) does not accept context in its API. We can add context checking at our wrapper layer but cannot cancel pdfcpu operations mid-execution. This is acceptable - we check context between operations.

#### Example: extractPagesParallel with Context

```go
func extractPagesParallel(ctx context.Context, r *pdf.Reader, pages []int, totalPages int, showProgress bool) (string, error) {
	// ... setup ...

	for _, pageNum := range pages {
		// Check context before launching each goroutine
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		go func(pn int) {
			results <- pageResult{pageNum: pn, text: extractPageText(r, pn, totalPages)}
		}(pageNum)
	}

	// ... collect results ...
}
```

---

## 5. Command Layer Analysis

### Current State

**File**: `/Users/lgbarn/Personal/pdf-cli/internal/commands/text.go`, `compress.go`, `images.go`, etc.

#### How Commands Work

1. **Root command** (`internal/cli/cli.go` line 24): `rootCmd` is a `*cobra.Command`
2. **Execution** (`cmd/pdf/main.go` line 19): `cli.Execute()` calls `rootCmd.Execute()`
3. **Context availability**: Cobra v1.10.2 provides `cmd.Context()` method
   - Returns context set via `ExecuteContext()` or a background context
   - Available in all `RunE` functions via the `cmd` parameter

#### Commands Calling Long-Running Operations

**text.go** (Line 93):
```go
func runText(cmd *cobra.Command, args []string) error {
	// ...
	text, err = engine.ExtractTextFromPDF(inputFile, pages, password, cli.Progress())
	//                                    ^ Should pass cmd.Context()
	// ...
}
```

**compress.go** (Line 110, 139):
```go
// Line 110: compressWithStdio
if err := pdf.Compress(input, output, password); err != nil {
	// Should pass context if Compress becomes context-aware
}

// Line 139: compressFile
if err := pdf.Compress(inputFile, output, password); err != nil {
	// Should pass context
}
```

**Note**: Most PDF operations (merge, split, compress, etc.) are fast enough that context may not be critical, but for consistency and future-proofing, all long-running operations should accept context.

#### Files That Need Modification

Commands that call OCR or long-running PDF operations:

1. `/Users/lgbarn/Personal/pdf-cli/internal/commands/text.go` (line 93) - Calls `ExtractTextFromPDF()`
2. `/Users/lgbarn/Personal/pdf-cli/internal/commands/compress.go` (lines 110, 139) - Calls `Compress()` in batch
3. `/Users/lgbarn/Personal/pdf-cli/internal/commands/merge.go` - Likely calls `MergeWithProgress()`
4. `/Users/lgbarn/Personal/pdf-cli/internal/commands/split.go` - Likely calls `SplitWithProgress()`
5. Any batch processing via `processBatch()` in `helpers.go` (line 66)

#### Adding Context to Commands

**Pattern**: Each command's `RunE` function receives `cmd *cobra.Command` as first parameter. Call `cmd.Context()` to get the context.

**Example for text.go**:

```go
func runText(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()  // Get context from Cobra

	// ... existing code ...

	if useOCR {
		// Pass context to OCR
		text, err = engine.ExtractTextFromPDF(ctx, inputFile, pages, password, cli.Progress())
		// ...
	} else {
		// Pass context to text extraction
		text, err = pdf.ExtractTextWithProgress(ctx, inputFile, pages, password, cli.Progress())
		// ...
	}

	// ...
}
```

---

## 6. Context Propagation Strategy

### Cobra Context Setup

**Current**: `cmd/pdf/main.go` line 19 calls `cli.Execute()` which calls `rootCmd.Execute()`

**Recommendation**: Use `ExecuteContext()` with signal handling for graceful shutdown:

```go
// cmd/pdf/main.go
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/lgbarn/pdf-cli/internal/cli"
	_ "github.com/lgbarn/pdf-cli/internal/commands"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// Create context that cancels on SIGINT/SIGTERM
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cli.SetVersion(version, commit, date)
	if err := cli.ExecuteContext(ctx); err != nil {
		os.Exit(1)
	}
}
```

**Update cli.go**:

```go
// internal/cli/cli.go
func ExecuteContext(ctx context.Context) error {
	return rootCmd.ExecuteContext(ctx)
}
```

### Context Flow Diagram

```
main.main()
  ├─ signal.NotifyContext() → ctx
  └─ cli.ExecuteContext(ctx)
       └─ rootCmd.ExecuteContext(ctx)
            └─ cmd.Context() available in all RunE functions
                 ├─ commands/text.go runText()
                 │    └─ ocr.ExtractTextFromPDF(ctx, ...)
                 │         └─ processImagesParallel(ctx, ...)
                 │              └─ backend.ProcessImage(ctx, ...)
                 ├─ commands/compress.go runCompress()
                 │    └─ processBatch() → compressFile()
                 │         └─ pdf.Compress(input, output, password)
                 └─ other commands...
```

### Benefits

1. **Graceful Shutdown**: SIGINT/SIGTERM during OCR processing will cancel context and stop goroutines
2. **Timeout Support**: Could add `context.WithTimeout()` for operations with time limits
3. **Cancellation Propagation**: Context flows from main → command → domain layer → backend
4. **Testing**: Tests can use `context.WithTimeout()` to prevent hanging

---

## 7. Testing Considerations

### Race Detection

**Current Status**: `go test -race ./...` passes with no races detected (verified 2026-01-31)

**After Changes**: Must continue to pass race detector

### Test Modifications Needed

1. **config_test.go**:
   - Tests calling `Reset()` must not run in parallel
   - Consider `t.Setenv()` instead of `os.Setenv()` (automatically isolated per test in Go 1.17+)
   - Alternative: Use `t.Parallel()` only for tests that don't use global state

2. **logger_test.go**:
   - Same considerations as config_test.go
   - Tests that modify global logger must not run in parallel

3. **ocr_test.go**:
   - Add tests for context cancellation
   - Test that `ExtractTextFromPDF(ctx, ...)` stops when context is cancelled
   - Test parallel processing with cancelled context

### Example Test for Context Cancellation

```go
func TestExtractTextFromPDF_ContextCancellation(t *testing.T) {
	engine, err := NewEngine("eng")
	if err != nil {
		t.Skip("OCR not available")
	}
	defer engine.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err = engine.ExtractTextFromPDF(ctx, "testdata/sample.pdf", nil, "", false)
	if err != context.Canceled {
		t.Errorf("Expected context.Canceled, got %v", err)
	}
}
```

---

## 8. Implementation Plan

### Phase 2.1: Thread-Safe Singletons

**Files to Modify**:
1. `/Users/lgbarn/Personal/pdf-cli/internal/config/config.go`
   - Add `sync.RWMutex` for global state
   - Update `Get()` to use double-checked locking
   - Update `Reset()` to acquire write lock
   - Add tests for concurrent access

2. `/Users/lgbarn/Personal/pdf-cli/internal/logging/logger.go`
   - Add `sync.RWMutex` for global state
   - Update `Init()`, `Get()`, `Reset()` with proper locking
   - Add tests for concurrent access

**Testing**:
- Add benchmark test for `Get()` contention
- Add `t.Run("concurrent", func(t *testing.T) { ... })` tests
- Verify `go test -race ./internal/config ./internal/logging` passes

### Phase 2.2: Context Propagation - OCR Layer

**Files to Modify**:
1. `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr.go`
   - Update `downloadTessdata(ctx context.Context, dataDir, lang string) error`
   - Update `ExtractTextFromPDF(ctx context.Context, ...)` signature
   - Update `processImages(ctx, ...)`, `processImagesSequential(ctx, ...)`, `processImagesParallel(ctx, ...)`
   - Add context checks in parallel loops

**Testing**:
- Add context cancellation tests
- Verify progress bar stops on cancellation
- Test parallel processing with short timeout

### Phase 2.3: Context Propagation - PDF Layer

**Files to Modify**:
1. `/Users/lgbarn/Personal/pdf-cli/internal/pdf/text.go`
   - Update `ExtractTextWithProgress(ctx context.Context, ...)`
   - Update `extractPagesParallel(ctx, ...)`
   - Add context checks in loops

2. `/Users/lgbarn/Personal/pdf-cli/internal/pdf/transform.go`
   - Update `MergeWithProgress(ctx, ...)`
   - Update `SplitWithProgress(ctx, ...)`
   - Add context checks in loops

**Note**: Underlying pdfcpu calls don't accept context. Add context checks between operations only.

### Phase 2.4: Context Propagation - Command Layer

**Files to Modify**:
1. `/Users/lgbarn/Personal/pdf-cli/cmd/pdf/main.go`
   - Add `signal.NotifyContext()` for graceful shutdown
   - Call `cli.ExecuteContext(ctx)` instead of `cli.Execute()`

2. `/Users/lgbarn/Personal/pdf-cli/internal/cli/cli.go`
   - Add `func ExecuteContext(ctx context.Context) error`

3. `/Users/lgbarn/Personal/pdf-cli/internal/commands/text.go`
   - Update `runText()` to get context and pass to `ExtractTextFromPDF(ctx, ...)`
   - Update call to `pdf.ExtractTextWithProgress(ctx, ...)`

4. `/Users/lgbarn/Personal/pdf-cli/internal/commands/compress.go` (if needed)
5. `/Users/lgbarn/Personal/pdf-cli/internal/commands/merge.go` (if exists)
6. `/Users/lgbarn/Personal/pdf-cli/internal/commands/split.go` (if exists)

**Testing**:
- Manual test: Run `pdf text large-scanned.pdf --ocr` and press Ctrl+C
- Verify graceful shutdown
- Verify no "leaked goroutines" warnings from race detector

---

## 9. Files Requiring Modification

### Complete List

| File Path | Changes Required | Priority |
|-----------|------------------|----------|
| `/Users/lgbarn/Personal/pdf-cli/internal/config/config.go` | Add sync.RWMutex, update Get/Reset | HIGH |
| `/Users/lgbarn/Personal/pdf-cli/internal/logging/logger.go` | Add sync.RWMutex, update Init/Get/Reset | HIGH |
| `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr.go` | Add context to 5 functions | HIGH |
| `/Users/lgbarn/Personal/pdf-cli/internal/pdf/text.go` | Add context to ExtractTextWithProgress, extractPagesParallel | MEDIUM |
| `/Users/lgbarn/Personal/pdf-cli/internal/pdf/transform.go` | Add context to MergeWithProgress, SplitWithProgress | MEDIUM |
| `/Users/lgbarn/Personal/pdf-cli/cmd/pdf/main.go` | Add signal.NotifyContext, call ExecuteContext | HIGH |
| `/Users/lgbarn/Personal/pdf-cli/internal/cli/cli.go` | Add ExecuteContext wrapper | HIGH |
| `/Users/lgbarn/Personal/pdf-cli/internal/commands/text.go` | Pass cmd.Context() to OCR/PDF functions | HIGH |
| `/Users/lgbarn/Personal/pdf-cli/internal/commands/compress.go` | Pass context (if batch processing) | LOW |
| `/Users/lgbarn/Personal/pdf-cli/internal/commands/merge.go` | Pass context (if long-running) | LOW |
| `/Users/lgbarn/Personal/pdf-cli/internal/commands/split.go` | Pass context (if long-running) | LOW |
| `/Users/lgbarn/Personal/pdf-cli/internal/config/config_test.go` | Add concurrent access tests | MEDIUM |
| `/Users/lgbarn/Personal/pdf-cli/internal/logging/logger_test.go` | Add concurrent access tests | MEDIUM |
| `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr_test.go` | Add context cancellation tests | MEDIUM |

**Total**: 14 files

---

## 10. Risks and Edge Cases

### Risk 1: sync.RWMutex Performance Impact

**Description**: Adding locks to `config.Get()` and `logging.Get()` could add overhead to every log call and config access.

**Mitigation**:
- Use RWMutex (read locks are cheap and can be held concurrently)
- After initialization, reads are very fast (uncontended lock)
- Benchmark shows RWMutex read lock is ~10ns on modern hardware
- Config/logging are not hot paths (not called millions of times per second)

**Evidence**: The `logging.Get()` pattern is used by package-level functions (`Debug()`, `Info()`, etc.) which are called frequently, but slog itself is optimized for concurrent access.

### Risk 2: Context Cancellation Mid-Operation

**Description**: If context is cancelled while processing an image or PDF page, the operation might leave temporary files or be in an inconsistent state.

**Mitigation**:
- Use `defer os.RemoveAll(tmpDir)` to clean up temp directories (already present in ocr.go line 223)
- Check context between operations, not during
- Backend operations (ProcessImage) are atomic - they either complete or return error
- Progress bars are finalized in `defer` statements (already present)

**Edge Case**: Native tesseract (exec.CommandContext) will be killed mid-execution if context cancelled. This is acceptable - the temp files are cleaned up.

### Risk 3: Test Flakiness with Reset()

**Description**: If tests use `Reset()` and run in parallel (`go test -parallel=N`), they will interfere with each other.

**Mitigation**:
- Document that tests using `Reset()` must not use `t.Parallel()`
- Use `t.Setenv()` for environment variable isolation (Go 1.17+)
- Consider making `Reset()` panic with a message if called outside tests
- Alternative: Use dependency injection in tests instead of global singletons

**Example**:
```go
func Reset() {
	if !testing.Testing() {
		panic("Reset() should only be called in tests")
	}
	globalMu.Lock()
	defer globalMu.Unlock()
	global = nil
}
```

### Risk 4: Backward Compatibility

**Description**: Changing function signatures (adding context parameter) breaks API compatibility.

**Mitigation**:
- This is an internal package (`internal/`), not a public API
- No external consumers to worry about
- Existing code will fail to compile (good - forces update)
- Backward-compatible wrappers not needed

### Risk 5: pdfcpu Library Doesn't Support Context

**Description**: The underlying pdfcpu library operations cannot be cancelled mid-execution.

**Mitigation**:
- Add context checks between pdfcpu calls (e.g., between merging each file)
- Document that cancellation is best-effort
- For single-file operations (compress, encrypt), cancellation might not trigger until completion
- This is acceptable - most single-file operations are fast (<5 seconds)

**Example**: In `MergeWithProgress()`, check context between each merge iteration:

```go
for i := 1; i < len(inputs); i++ {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	err := api.MergeCreateFile(...)  // Cannot cancel this call
	// ...
}
```

### Risk 6: Progress Bar State on Cancellation

**Description**: If context is cancelled during parallel processing, the progress bar might not reach 100%.

**Mitigation**:
- `defer progress.FinishProgressBar(bar)` already present in all functions
- Finish() should be called even on error
- Progress bar library (progressbar/v3) handles incomplete bars gracefully
- No action needed - existing code is correct

---

## 11. Patterns and Conventions

### Existing Code Patterns to Follow

1. **Error Wrapping**: Use `fmt.Errorf("operation failed: %w", err)` for error chains (Go 1.13+)
   - Example: ocr.go line 137, 221, 253

2. **Temp Directory Cleanup**: Use `defer os.RemoveAll(tmpDir)` immediately after creation
   - Example: ocr.go line 223, text.go line 158

3. **Progress Bar Cleanup**: Use `defer progress.FinishProgressBar(bar)` immediately after creation
   - Example: ocr.go line 314, text.go line 71

4. **Context Variable Name**: Use `ctx` (not `context` or `c`)
   - Example: backend.go line 11, native.go line 46, wasm.go line 95

5. **Context First Parameter**: Context is always the first parameter in function signatures
   - Standard Go convention
   - Example from stdlib: `http.NewRequestWithContext(ctx, method, url, body)`

### New Conventions to Establish

1. **Context Checking in Loops**:
   ```go
   for i, item := range items {
       select {
       case <-ctx.Done():
           return ctx.Err()
       default:
       }
       // Process item
   }
   ```

2. **Mutex Naming**: Use `globalMu` for protecting `global` variable
   - Suffix `Mu` indicates mutex (common Go convention)

3. **Context in Tests**: Use `context.Background()` or `context.WithTimeout()`, never nil
   ```go
   ctx := context.Background()
   result, err := engine.ExtractTextFromPDF(ctx, ...)
   ```

4. **Timeout Constants**: Define reasonable timeouts as package constants
   ```go
   const (
       DefaultDownloadTimeout = 5 * time.Minute
       DefaultOCRTimeout      = 10 * time.Minute
   )
   ```

---

## 12. Success Criteria Verification

### R4: Global config and logging state must be thread-safe

**Before**:
```go
// config.go line 146
if global == nil {  // RACE
    global, err = Load()
}
```

**After**:
```go
// config.go (modified)
globalMu.RLock()
if global != nil {
    defer globalMu.RUnlock()
    return global
}
globalMu.RUnlock()

globalMu.Lock()
defer globalMu.Unlock()
if global == nil {
    global, err = Load()
    // ...
}
```

**Verification**:
- `go test -race ./internal/config ./internal/logging` passes
- Concurrent benchmark shows no data races
- Manual review: No bare nil-checks remain

### R5: All long-running operations accept and propagate context.Context

**Before**:
```go
// ocr.go line 212
func (e *Engine) ExtractTextFromPDF(pdfPath string, ...) (string, error)

// text.go line 22
func ExtractTextWithProgress(input string, ...) (string, error)
```

**After**:
```go
// ocr.go (modified)
func (e *Engine) ExtractTextFromPDF(ctx context.Context, pdfPath string, ...) (string, error)

// text.go (modified)
func ExtractTextWithProgress(ctx context.Context, input string, ...) (string, error)
```

**Verification**:
- All functions identified in section 3 and 4 have context parameter
- Context is checked in loops (grep for `ctx.Done()`)
- `cmd.Context()` is called in all command handlers
- Test exists for context cancellation behavior

### No change to public CLI behavior

**Verification**:
- Run existing integration tests: `go test -tags=integration ./...`
- Manual smoke tests:
  - `pdf text sample.pdf` - works as before
  - `pdf text scanned.pdf --ocr` - works as before
  - Press Ctrl+C during OCR - gracefully stops (new behavior, but acceptable)
- No command-line flags added or changed
- Output format unchanged

### go test -race ./... passes with zero data races

**Verification**:
- Run `go test -race ./...` before and after changes
- Before: 0 races (verified)
- After: 0 races (must verify)
- CI integration: Add `-race` flag to GitHub Actions workflow

---

## 13. Documentation References

### Cobra Context Support

- [Cobra Package Documentation](https://pkg.go.dev/github.com/spf13/cobra) - Official API reference
- [Using context.Context with Cobra](https://blog.ksub.org/bytes/2019/10/07/using-context.context-with-cobra/) - Blog post on context integration
- [Cobra PR #893](https://github.com/spf13/cobra/pull/893) - Added context support to Cobra

### Go sync Package

- [sync.Once Documentation](https://pkg.go.dev/sync#Once) - Thread-safe one-time initialization
- [sync.RWMutex Documentation](https://pkg.go.dev/sync#RWMutex) - Reader/writer mutual exclusion lock
- [The Go Memory Model](https://go.dev/ref/mem) - Defines when reads can observe writes

### Context Package

- [context Package](https://pkg.go.dev/context) - Standard library context for cancellation
- [Go Concurrency Patterns: Context](https://go.dev/blog/context) - Official blog post
- [Context and structs](https://go.dev/blog/context-and-structs) - Best practices

### Race Detector

- [Data Race Detector](https://go.dev/doc/articles/race_detector) - Using `-race` flag
- [Happens Before Guarantees](https://go.dev/ref/mem#tmp_8) - Memory ordering guarantees

---

## 14. Next Steps

1. **Review this document** with team/stakeholders
2. **Create detailed implementation plan** (sub-tasks for each phase)
3. **Set up feature branch**: `feature/phase-2-thread-safety`
4. **Implement Phase 2.1**: Thread-safe singletons (highest risk, smallest change)
5. **Review and test Phase 2.1** before proceeding
6. **Implement Phases 2.2-2.4**: Context propagation (larger change, depends on 2.1)
7. **Integration testing** with manual tests (Ctrl+C during OCR)
8. **Update documentation** if needed (README, architecture.md)
9. **Merge to main** after all tests pass

---

## Appendix A: Function Signature Changes Summary

### Config Package

No function signature changes (internal locking only)

### Logging Package

No function signature changes (internal locking only)

### OCR Package

| Function | Before | After |
|----------|--------|-------|
| `downloadTessdata` | `(dataDir, lang string) error` | `(ctx context.Context, dataDir, lang string) error` |
| `ExtractTextFromPDF` | `(pdfPath string, pages []int, password string, showProgress bool)` | `(ctx context.Context, pdfPath string, pages []int, password string, showProgress bool)` |
| `processImages` | `(imageFiles []string, showProgress bool)` | `(ctx context.Context, imageFiles []string, showProgress bool)` |
| `processImagesSequential` | `(imageFiles []string, showProgress bool)` | `(ctx context.Context, imageFiles []string, showProgress bool)` |
| `processImagesParallel` | `(imageFiles []string, showProgress bool)` | `(ctx context.Context, imageFiles []string, showProgress bool)` |

### PDF Package

| Function | Before | After |
|----------|--------|-------|
| `ExtractTextWithProgress` | `(input string, pages []int, password string, showProgress bool)` | `(ctx context.Context, input string, pages []int, password string, showProgress bool)` |
| `extractPagesParallel` | `(r *pdf.Reader, pages []int, totalPages int, showProgress bool)` | `(ctx context.Context, r *pdf.Reader, pages []int, totalPages int, showProgress bool)` |
| `MergeWithProgress` | `(inputs []string, output, password string, showProgress bool)` | `(ctx context.Context, inputs []string, output, password string, showProgress bool)` |
| `SplitWithProgress` | `(input, outputDir string, pageCount int, password string, showProgress bool)` | `(ctx context.Context, input, outputDir string, pageCount int, password string, showProgress bool)` |

### CLI Package

| Function | Before | After |
|----------|--------|-------|
| N/A | N/A | `ExecuteContext(ctx context.Context) error` (new function) |

**Total function signature changes**: 10 functions

---

## Appendix B: Test Coverage Gaps

### Current Test Coverage

Based on review of test files:

- `config_test.go`: 12 tests, 243 lines - Good coverage of config loading/saving
- `logger_test.go`: 14 tests, 301 lines - Good coverage of logger functionality
- `ocr_test.go`: 6 tests, 129 lines - Basic coverage, no concurrency tests

### Missing Test Coverage

1. **Concurrent Access Tests**:
   - No tests for concurrent `config.Get()` calls
   - No tests for concurrent `logging.Get()` calls
   - No tests for `Reset()` during `Get()`

2. **Context Cancellation Tests**:
   - No tests for OCR with cancelled context
   - No tests for text extraction with timeout
   - No tests for parallel processing cancellation

3. **Race Detection Tests**:
   - Race detector not run in CI (no `-race` flag in test commands)
   - No documented race detector results

### Recommended New Tests

```go
// config_test.go additions
func TestGetConcurrent(t *testing.T) {
    Reset()
    var wg sync.WaitGroup
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            cfg := Get()
            if cfg == nil {
                t.Error("Get() returned nil")
            }
        }()
    }
    wg.Wait()
}

// ocr_test.go additions
func TestExtractTextFromPDF_Cancelled(t *testing.T) {
    ctx, cancel := context.WithCancel(context.Background())
    cancel()

    engine, _ := NewEngine("eng")
    defer engine.Close()

    _, err := engine.ExtractTextFromPDF(ctx, "testdata/sample.pdf", nil, "", false)
    if !errors.Is(err, context.Canceled) {
        t.Errorf("Expected context.Canceled, got %v", err)
    }
}
```

---

## Appendix C: Grep Results for Audit

### All Context Usage

```bash
$ grep -r "context\." internal/ | grep -v "_test.go" | grep -v ".md"
internal/ocr/backend.go:	"context"
internal/ocr/backend.go:	ProcessImage(ctx context.Context, imagePath, lang string) (string, error)
internal/ocr/native.go:	"context"
internal/ocr/native.go:func (n *NativeBackend) ProcessImage(ctx context.Context, imagePath, lang string) (string, error) {
internal/ocr/wasm.go:	"context"
internal/ocr/wasm.go:func (w *WASMBackend) initializeTesseract(ctx context.Context, lang string) error {
internal/ocr/wasm.go:func (w *WASMBackend) ProcessImage(ctx context.Context, imagePath, lang string) (string, error) {
internal/ocr/wasm.go:	return w.tess.Close(context.Background())
internal/ocr/ocr.go:	"context"
internal/ocr/ocr.go:	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
internal/ocr/ocr.go:	ctx := context.Background()
internal/ocr/ocr.go:	ctx := context.Background()
```

**Analysis**: Context is already used in backend implementations, but not propagated from top-level APIs.

### All sync.* Usage

```bash
$ grep -r "sync\." internal/ | grep -v "_test.go" | grep -v ".md"
internal/ocr/ocr.go:	"sync"
internal/ocr/ocr.go:	var wg sync.WaitGroup
```

**Analysis**: Only usage is `sync.WaitGroup` in parallel OCR processing. No mutexes or `sync.Once` currently used.

---

**End of Research Document**
