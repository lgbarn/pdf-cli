---
phase: error-handling-reliability
plan: 1.1
wave: 1
dependencies: []
must_haves:
  - R6: Parallel processing must surface all errors (not silently drop)
  - R8: File close errors must be checked and propagated for write operations
files_touched:
  - internal/ocr/ocr.go
  - internal/fileio/files.go
tdd: true
---

# Plan 1.1: Error Propagation and Close Handling

## Overview

Fix silent error swallowing in OCR parallel processing and propagate file close errors on write paths. This plan addresses R6 (parallel error surfacing) and R8 (close error checking).

## Context

Current issues:
1. **processImagesParallel** (ocr.go:359): Errors from `e.backend.ProcessImage()` are discarded with `_`, causing silent failures
2. **processImagesSequential** (ocr.go:322-325): Errors checked but not all propagated to caller
3. **imageResult struct** (ocr.go:293-296): Only carries text, not errors
4. **Write-path close errors**: Two locations ignore close errors on temp files during write operations
   - internal/fileio/files.go:104 — CopyFile's dstFile.Close()
   - internal/ocr/ocr.go:202,205 — downloadTessdata's tmpFile.Close()

## Tasks

<task id="1" files="internal/ocr/ocr.go" tdd="true">
  <action>
    Add error field to imageResult struct and collect all processing errors:

    1. Modify imageResult struct (line 293):
       ```go
       type imageResult struct {
           index int
           text  string
           err   error  // Add this field
       }
       ```

    2. Update processImagesParallel (lines 356-361) to capture errors:
       ```go
       go func(idx int, path string) {
           defer wg.Done()
           defer func() { <-sem }()
           text, err := e.backend.ProcessImage(ctx, path, e.lang)
           results <- imageResult{index: idx, text: text, err: err}
       }(i, imgPath)
       ```

    3. Collect and join errors after results loop (after line 377):
       ```go
       // Collect results in order using a slice
       texts := make([]string, len(imageFiles))
       var errs []error
       for res := range results {
           texts[res.index] = res.text
           if res.err != nil {
               errs = append(errs, fmt.Errorf("image %d: %w", res.index, res.err))
           }
           if bar != nil {
               _ = bar.Add(1)
           }
       }

       if len(errs) > 0 {
           return "", errors.Join(errs...)
       }

       return joinNonEmpty(texts, "\n"), nil
       ```

    4. Update processImagesSequential (lines 322-325) to collect all errors:
       ```go
       texts := make([]string, 0, len(imageFiles))
       var errs []error

       for i, imgPath := range imageFiles {
           if ctx.Err() != nil {
               return "", ctx.Err()
           }
           text, err := e.backend.ProcessImage(ctx, imgPath, e.lang)
           if err != nil {
               errs = append(errs, fmt.Errorf("image %d: %w", i, err))
           } else {
               texts = append(texts, text)
           }
           if bar != nil {
               _ = bar.Add(1)
           }
       }

       if len(errs) > 0 {
           return "", errors.Join(errs...)
       }

       return joinNonEmpty(texts, "\n"), nil
       ```

    5. Add "errors" import to package imports (line 3)
  </action>
  <verify>
    Run unit tests and add test case for error propagation:

    ```bash
    cd /Users/lgbarn/Personal/pdf-cli/.worktrees/phase-2-concurrency
    go test -v -run TestProcessImages ./internal/ocr/...
    ```

    Add test in internal/ocr/ocr_test.go:
    ```go
    func TestProcessImagesParallelErrorPropagation(t *testing.T) {
        // Create mock backend that returns errors for specific images
        // Verify errors.Join is used and all errors are returned
    }
    ```
  </verify>
  <done>
    - imageResult struct has error field
    - Both processImagesParallel and processImagesSequential collect all errors
    - errors.Join combines multiple errors
    - Unit tests pass with error propagation test
    - No errors silently swallowed
  </done>
</task>

<task id="2" files="internal/fileio/files.go,internal/ocr/ocr.go" tdd="true">
  <action>
    Fix file close error handling on write paths using named return + defer closure pattern:

    **Location 1: internal/fileio/files.go — CopyFile (lines 84-111)**

    Change function signature (line 85):
    ```go
    func CopyFile(src, dst string) (err error) {
    ```

    Replace dstFile defer (line 104):
    ```go
    defer func() {
        if cerr := dstFile.Close(); cerr != nil && err == nil {
            err = fmt.Errorf("failed to close destination file: %w", cerr)
        }
    }()
    ```

    **Location 2: internal/ocr/ocr.go — downloadTessdata (lines 169-209)**

    Change function signature (line 169):
    ```go
    func downloadTessdata(ctx context.Context, dataDir, lang string) (err error) {
    ```

    Replace first tmpFile.Close (line 202):
    ```go
    if _, err := io.Copy(io.MultiWriter(tmpFile, bar), resp.Body); err != nil {
        _ = tmpFile.Close()  // Keep this - error path, main error takes precedence
        return err
    }
    ```

    Replace second tmpFile.Close and add defer closure (lines 205-206):
    ```go
    defer func() {
        if cerr := tmpFile.Close(); cerr != nil && err == nil {
            err = fmt.Errorf("failed to close temp file: %w", cerr)
        }
        tmpFile = nil
    }()
    progress.FinishProgressBar(bar)

    if err := os.Rename(tmpPath, dataFile); err != nil {
        return fmt.Errorf("failed to rename temp file: %w", err)
    }

    return nil
    ```

    Note: The read-only defer closes (8 total) are intentionally left as-is since close errors on read-only files are not actionable.
  </action>
  <verify>
    Run targeted tests and simulate close errors:

    ```bash
    cd /Users/lgbarn/Personal/pdf-cli/.worktrees/phase-2-concurrency
    go test -v -run TestCopyFile ./internal/fileio/...
    go test -v -run TestDownloadTessdata ./internal/ocr/...
    ```

    Add test case in internal/fileio/files_test.go:
    ```go
    func TestCopyFileCloseError(t *testing.T) {
        // Test that close errors are propagated
        // Use file on read-only filesystem or similar
    }
    ```
  </verify>
  <done>
    - CopyFile uses named return + defer closure for dstFile.Close()
    - downloadTessdata uses named return + defer closure for tmpFile.Close()
    - Close errors on write paths are checked and propagated
    - Unit tests verify close error handling
    - Read-only close errors remain ignored (as intended)
  </done>
</task>

<task id="3" files="internal/ocr/ocr.go,internal/fileio/files.go" tdd="false">
  <action>
    Add integration test for end-to-end error propagation:

    Create internal/ocr/ocr_integration_test.go:
    ```go
    //go:build integration

    package ocr

    import (
        "context"
        "testing"
    )

    func TestParallelProcessingWithErrors(t *testing.T) {
        // Test that parallel OCR with some failing images returns all errors
        // Use 10+ images to trigger parallel path
        // Mock some to fail, verify errors.Join output
    }

    func TestSequentialProcessingWithErrors(t *testing.T) {
        // Test that sequential OCR with some failing images returns all errors
        // Use <5 images to trigger sequential path
    }
    ```

    Add file operation integration tests in internal/fileio/files_integration_test.go:
    ```go
    //go:build integration

    package fileio

    import (
        "testing"
    )

    func TestCopyFileToReadOnlyDest(t *testing.T) {
        // Verify close errors are caught when destination becomes read-only
        // during write operation
    }
    ```
  </action>
  <verify>
    Run integration tests:

    ```bash
    cd /Users/lgbarn/Personal/pdf-cli/.worktrees/phase-2-concurrency
    go test -v -tags=integration ./internal/ocr/...
    go test -v -tags=integration ./internal/fileio/...
    ```
  </verify>
  <done>
    - Integration tests verify error propagation end-to-end
    - Parallel and sequential paths both tested
    - File close error handling tested with realistic scenarios
    - All integration tests pass
  </done>
</task>

## Verification

Run full test suite:

```bash
cd /Users/lgbarn/Personal/pdf-cli/.worktrees/phase-2-concurrency
go test -v ./internal/ocr/...
go test -v ./internal/fileio/...
go test -v -tags=integration ./...
```

Verify error handling with manual test:

```bash
# Test parallel error collection (requires 5+ page PDF)
./pdf-cli ocr test-multi.pdf --lang eng

# Test close error on write (requires permission manipulation)
chmod 444 /tmp/test-dest
./pdf-cli merge input1.pdf input2.pdf /tmp/test-dest/output.pdf
```

## Success Criteria

- imageResult struct contains error field
- processImagesParallel collects and joins all errors with errors.Join
- processImagesSequential collects and joins all errors with errors.Join
- CopyFile propagates dstFile.Close() errors
- downloadTessdata propagates tmpFile.Close() errors
- Unit tests cover error propagation paths
- Integration tests verify end-to-end error handling
- No errors are silently swallowed
- go test ./... passes
