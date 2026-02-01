---
phase: error-handling-reliability
plan: 2.1
wave: 2
dependencies: [1.1]
must_haves:
  - R11: Temp file cleanup on crash/interrupt via signal handlers
files_touched:
  - internal/cleanup/cleanup.go
  - internal/cleanup/cleanup_test.go
  - cmd/pdf/main.go
  - internal/ocr/ocr.go
  - internal/ocr/native.go
  - internal/pdf/text.go
  - internal/pdf/transform.go
  - internal/commands/patterns/stdio.go
  - internal/fileio/stdio.go
  - internal/fileio/files.go
tdd: true
---

# Plan 2.1: Signal-Based Temp File Cleanup

## Overview

Create a centralized cleanup registry for temporary files and integrate it with existing signal handling (signal.NotifyContext) to ensure temp files are removed on crash/interrupt. This plan addresses R11 (temp file cleanup on signals).

## Context

Current state:
- 8 temp file creation points in production code lack crash cleanup
- main.go already uses `signal.NotifyContext` from Phase 2 (line 25)
- Most temp files use defer os.Remove() which doesn't run on signals
- No centralized cleanup mechanism exists

Temp file creation locations:
1. internal/ocr/ocr.go:193 — downloadTessdata (tessdata download)
2. internal/ocr/ocr.go:219 — ExtractTextFromPDF (temp directory)
3. internal/ocr/native.go:49 — ProcessImage (output file)
4. internal/pdf/text.go:~90 — extractTextFallback (temp directory)
5. internal/pdf/transform.go:35 — MergeWithProgress (temp merge file)
6. internal/commands/patterns/stdio.go:37 — StdioHandler.Handle (temp operation file)
7. internal/fileio/stdio.go:32 — ReadFromStdin (temp stdin file)
8. internal/fileio/files.go:48 — AtomicWrite (temp write file)

## Tasks

<task id="1" files="internal/cleanup/cleanup.go,internal/cleanup/cleanup_test.go" tdd="true">
  <action>
    Create internal/cleanup package with thread-safe cleanup registry:

    **internal/cleanup/cleanup.go:**
    ```go
    package cleanup

    import (
        "fmt"
        "os"
        "sync"
    )

    var (
        mu       sync.Mutex
        paths    []string
        disabled bool
    )

    // Register adds a file or directory path for cleanup on exit/signal.
    // Returns a function that can be called to unregister the path.
    // Safe to call from multiple goroutines.
    func Register(path string) func() {
        mu.Lock()
        defer mu.Unlock()

        if disabled {
            return func() {}
        }

        paths = append(paths, path)
        index := len(paths) - 1

        return func() {
            mu.Lock()
            defer mu.Unlock()
            if index < len(paths) && paths[index] == path {
                paths[index] = "" // Mark as removed
            }
        }
    }

    // Run executes cleanup of all registered paths.
    // Called automatically by signal handler or manually in defer.
    // Removes files/directories in reverse registration order.
    // Safe to call multiple times (idempotent).
    func Run() error {
        mu.Lock()
        defer mu.Unlock()

        if len(paths) == 0 {
            return nil
        }

        var errs []error
        // Clean up in reverse order (LIFO)
        for i := len(paths) - 1; i >= 0; i-- {
            path := paths[i]
            if path == "" {
                continue // Already unregistered
            }

            if err := os.RemoveAll(path); err != nil && !os.IsNotExist(err) {
                errs = append(errs, fmt.Errorf("cleanup %s: %w", path, err))
            }
        }

        paths = nil // Clear registry

        if len(errs) > 0 {
            return fmt.Errorf("cleanup errors: %v", errs)
        }
        return nil
    }

    // Disable prevents new registrations (for testing).
    func Disable() {
        mu.Lock()
        defer mu.Unlock()
        disabled = true
    }

    // Enable re-enables registrations (for testing).
    func Enable() {
        mu.Lock()
        defer mu.Unlock()
        disabled = false
    }

    // Reset clears all registered paths (for testing).
    func Reset() {
        mu.Lock()
        defer mu.Unlock()
        paths = nil
        disabled = false
    }
    ```

    **internal/cleanup/cleanup_test.go:**
    ```go
    package cleanup

    import (
        "os"
        "path/filepath"
        "sync"
        "testing"
    )

    func TestRegisterAndRun(t *testing.T) {
        Reset()

        tmpDir := t.TempDir()
        file1 := filepath.Join(tmpDir, "file1.txt")
        file2 := filepath.Join(tmpDir, "file2.txt")

        os.WriteFile(file1, []byte("test1"), 0600)
        os.WriteFile(file2, []byte("test2"), 0600)

        Register(file1)
        Register(file2)

        if err := Run(); err != nil {
            t.Fatalf("Run() failed: %v", err)
        }

        // Verify files removed
        if _, err := os.Stat(file1); !os.IsNotExist(err) {
            t.Errorf("file1 still exists")
        }
        if _, err := os.Stat(file2); !os.IsNotExist(err) {
            t.Errorf("file2 still exists")
        }
    }

    func TestUnregister(t *testing.T) {
        Reset()

        tmpDir := t.TempDir()
        file := filepath.Join(tmpDir, "file.txt")
        os.WriteFile(file, []byte("test"), 0600)

        unregister := Register(file)
        unregister() // Unregister immediately

        if err := Run(); err != nil {
            t.Fatalf("Run() failed: %v", err)
        }

        // File should still exist (unregistered)
        if _, err := os.Stat(file); err != nil {
            t.Errorf("file should not be removed: %v", err)
        }
    }

    func TestConcurrentRegister(t *testing.T) {
        Reset()

        tmpDir := t.TempDir()
        var wg sync.WaitGroup

        for i := 0; i < 100; i++ {
            wg.Add(1)
            go func(n int) {
                defer wg.Done()
                file := filepath.Join(tmpDir, fmt.Sprintf("file%d.txt", n))
                os.WriteFile(file, []byte("test"), 0600)
                Register(file)
            }(i)
        }

        wg.Wait()

        if err := Run(); err != nil {
            t.Fatalf("Run() failed: %v", err)
        }

        // Verify all files removed
        entries, _ := os.ReadDir(tmpDir)
        if len(entries) > 0 {
            t.Errorf("expected all files removed, got %d remaining", len(entries))
        }
    }

    func TestRunIdempotent(t *testing.T) {
        Reset()

        tmpDir := t.TempDir()
        file := filepath.Join(tmpDir, "file.txt")
        os.WriteFile(file, []byte("test"), 0600)
        Register(file)

        if err := Run(); err != nil {
            t.Fatalf("first Run() failed: %v", err)
        }

        // Second Run should be no-op
        if err := Run(); err != nil {
            t.Fatalf("second Run() failed: %v", err)
        }
    }
    ```
  </action>
  <verify>
    Run unit tests:

    ```bash
    cd /Users/lgbarn/Personal/pdf-cli/.worktrees/phase-2-concurrency
    go test -v ./internal/cleanup/...
    go test -race -v ./internal/cleanup/...  # Race detector
    ```
  </verify>
  <done>
    - internal/cleanup package created with Register/Run API
    - Thread-safe with mutex protection
    - Unregister function returned from Register
    - LIFO cleanup order (reverse registration)
    - Idempotent Run (safe to call multiple times)
    - Unit tests pass including concurrent access
    - Race detector clean
  </done>
</task>

<task id="2" files="cmd/pdf/main.go,internal/ocr/ocr.go,internal/ocr/native.go,internal/pdf/text.go,internal/pdf/transform.go,internal/commands/patterns/stdio.go,internal/fileio/stdio.go,internal/fileio/files.go" tdd="false">
  <action>
    Integrate cleanup.Run with signal handler and all temp file creation sites:

    **1. cmd/pdf/main.go — Add cleanup to signal handler (lines 24-27):**
    ```go
    import (
        "context"
        "os"
        "os/signal"
        "syscall"

        "github.com/lgbarn/pdf-cli/internal/cleanup"
        "github.com/lgbarn/pdf-cli/internal/cli"
        _ "github.com/lgbarn/pdf-cli/internal/commands"
    )

    func run() int {
        ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
        defer stop()
        defer cleanup.Run()  // Add this - runs on normal exit or signal

        cli.SetVersion(version, commit, date)
        if err := cli.ExecuteContext(ctx); err != nil {
            return 1
        }
        return 0
    }
    ```

    **2. internal/ocr/ocr.go — downloadTessdata (after line 193):**
    ```go
    tmpFile, err := os.CreateTemp(dataDir, "tessdata-*.tmp")
    if err != nil {
        return err
    }
    tmpPath := tmpFile.Name()
    unregister := cleanup.Register(tmpPath)
    defer unregister()
    defer os.Remove(tmpPath)
    ```

    **3. internal/ocr/ocr.go — ExtractTextFromPDF (after line 219):**
    ```go
    tmpDir, err := os.MkdirTemp("", "pdf-ocr-*")
    if err != nil {
        return "", fmt.Errorf("failed to create temp directory: %w", err)
    }
    unregister := cleanup.Register(tmpDir)
    defer unregister()
    defer os.RemoveAll(tmpDir)
    ```

    **4. internal/ocr/native.go — ProcessImage (after line 49):**
    ```go
    tmpFile, err := os.CreateTemp("", "ocr-output-*.txt")
    if err != nil {
        return "", fmt.Errorf("failed to create temp file: %w", err)
    }
    tmpPath := tmpFile.Name()
    unregister := cleanup.Register(tmpPath)
    defer unregister()
    defer os.Remove(tmpPath)
    ```

    **5. internal/pdf/text.go — extractTextFallback (find os.MkdirTemp call):**
    ```go
    tmpDir, err := os.MkdirTemp("", "pdf-text-*")
    if err != nil {
        return "", err
    }
    unregister := cleanup.Register(tmpDir)
    defer unregister()
    defer os.RemoveAll(tmpDir)
    ```

    **6. internal/pdf/transform.go — MergeWithProgress (after line 35):**
    ```go
    tmpFile, err := os.CreateTemp("", "pdf-merge-*.pdf")
    if err != nil {
        return err
    }
    tmpPath := tmpFile.Name()
    unregister := cleanup.Register(tmpPath)
    defer func() {
        unregister()
        _ = tmpFile.Close()
        _ = os.Remove(tmpPath)
    }()
    ```

    **7. internal/commands/patterns/stdio.go — StdioHandler.Handle (after line 37):**
    ```go
    tmpFile, err := os.CreateTemp("", "pdf-cli-"+h.Operation+"-*.pdf")
    if err != nil {
        return fmt.Errorf("failed to create temp output: %w", err)
    }
    tmpPath := tmpFile.Name()
    unregister := cleanup.Register(tmpPath)
    defer func() {
        unregister()
        _ = tmpFile.Close()
        _ = os.Remove(tmpPath)
    }()
    ```

    **8. internal/fileio/stdio.go — ReadFromStdin (after line 32):**
    ```go
    tmpFile, err := os.CreateTemp("", "pdf-cli-stdin-*.pdf")
    if err != nil {
        return "", fmt.Errorf("failed to create temp file: %w", err)
    }
    tmpPath := tmpFile.Name()
    unregister := cleanup.Register(tmpPath)
    defer func() {
        unregister()
        _ = tmpFile.Close()
        _ = os.Remove(tmpPath)
    }()
    ```

    **9. internal/fileio/files.go — AtomicWrite (after line 48):**
    ```go
    tmpFile, err := os.CreateTemp(dir, ".pdf-cli-tmp-*")
    if err != nil {
        return fmt.Errorf("failed to create temp file: %w", err)
    }
    tmpPath := tmpFile.Name()
    unregister := cleanup.Register(tmpPath)

    defer func() {
        if tmpFile != nil {
            unregister()
            _ = tmpFile.Close()
            _ = os.Remove(tmpPath)
        }
    }()
    ```

    Add import "github.com/lgbarn/pdf-cli/internal/cleanup" to all modified files.
  </action>
  <verify>
    Manual integration test with signal handling:

    ```bash
    cd /Users/lgbarn/Personal/pdf-cli/.worktrees/phase-2-concurrency
    go build -o pdf-cli ./cmd/pdf

    # Test 1: OCR with interrupt
    ./pdf-cli ocr large-file.pdf &
    PID=$!
    sleep 2
    kill -INT $PID
    # Verify no files left in /tmp matching pdf-ocr-* or ocr-output-*

    # Test 2: Merge with interrupt
    ./pdf-cli merge file1.pdf file2.pdf output.pdf &
    PID=$!
    sleep 1
    kill -INT $PID
    # Verify no files left in /tmp matching pdf-merge-*

    # Test 3: Stdin with interrupt
    cat input.pdf | ./pdf-cli split --pages 1-5 &
    PID=$!
    sleep 1
    kill -TERM $PID
    # Verify no files left in /tmp matching pdf-cli-stdin-*
    ```
  </verify>
  <done>
    - cleanup.Run called in main.go defer chain
    - All 8 temp file creation sites register with cleanup
    - Unregister called in each defer to prevent double cleanup
    - cleanup import added to all modified files
    - Manual signal tests verify no temp file leaks
    - Normal exit and signal exit both clean up properly
  </done>
</task>

<task id="3" files="internal/cleanup/cleanup_integration_test.go" tdd="true">
  <action>
    Create integration test for signal-based cleanup:

    **internal/cleanup/cleanup_integration_test.go:**
    ```go
    //go:build integration

    package cleanup_test

    import (
        "os"
        "os/exec"
        "path/filepath"
        "strings"
        "syscall"
        "testing"
        "time"
    )

    func TestSignalCleanup(t *testing.T) {
        // Build test binary
        binary := filepath.Join(t.TempDir(), "test-cleanup")
        cmd := exec.Command("go", "build", "-o", binary, "./testdata/signal_test_main.go")
        if err := cmd.Run(); err != nil {
            t.Fatalf("failed to build test binary: %v", err)
        }

        // Run binary that creates temp file and registers cleanup
        cmd = exec.Command(binary)
        cmd.Stdout = os.Stdout
        cmd.Stderr = os.Stderr

        if err := cmd.Start(); err != nil {
            t.Fatalf("failed to start binary: %v", err)
        }

        // Wait for temp file creation (read from stdout)
        time.Sleep(500 * time.Millisecond)

        // Send SIGINT
        if err := cmd.Process.Signal(syscall.SIGINT); err != nil {
            t.Fatalf("failed to send signal: %v", err)
        }

        // Wait for process to exit
        cmd.Wait()

        // Verify no temp files left behind
        entries, err := os.ReadDir(os.TempDir())
        if err != nil {
            t.Fatalf("failed to read temp dir: %v", err)
        }

        for _, e := range entries {
            if strings.HasPrefix(e.Name(), "cleanup-test-") {
                t.Errorf("temp file not cleaned up: %s", e.Name())
            }
        }
    }

    func TestNormalExitCleanup(t *testing.T) {
        // Similar test but let process exit normally
        // Verify cleanup still happens
    }
    ```

    **internal/cleanup/testdata/signal_test_main.go:**
    ```go
    package main

    import (
        "context"
        "fmt"
        "os"
        "os/signal"
        "syscall"
        "time"

        "github.com/lgbarn/pdf-cli/internal/cleanup"
    )

    func main() {
        ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
        defer stop()
        defer cleanup.Run()

        tmpFile, err := os.CreateTemp("", "cleanup-test-*.tmp")
        if err != nil {
            panic(err)
        }
        tmpPath := tmpFile.Name()
        unregister := cleanup.Register(tmpPath)
        defer unregister()

        fmt.Fprintf(os.Stderr, "Created temp file: %s\n", tmpPath)
        tmpFile.Close()

        // Wait for signal
        <-ctx.Done()
        fmt.Fprintf(os.Stderr, "Signal received, cleaning up...\n")
    }
    ```
  </action>
  <verify>
    Run integration tests:

    ```bash
    cd /Users/lgbarn/Personal/pdf-cli/.worktrees/phase-2-concurrency
    go test -v -tags=integration ./internal/cleanup/...
    ```
  </verify>
  <done>
    - Integration test verifies signal-based cleanup
    - Test binary simulates real signal handling flow
    - SIGINT and SIGTERM both tested
    - Normal exit cleanup also tested
    - No temp files leaked in any scenario
    - Integration tests pass
  </done>
</task>

## Verification

Full verification suite:

```bash
cd /Users/lgbarn/Personal/pdf-cli/.worktrees/phase-2-concurrency

# Unit tests
go test -v ./internal/cleanup/...
go test -race -v ./internal/cleanup/...

# Integration tests
go test -v -tags=integration ./internal/cleanup/...

# Build and manual test
go build -o pdf-cli ./cmd/pdf

# Test signal cleanup
./pdf-cli ocr test.pdf --lang eng &
PID=$!
sleep 2
kill -INT $PID
sleep 1
ls /tmp/pdf-ocr-* 2>&1 | grep -q "No such file" && echo "PASS: OCR cleanup" || echo "FAIL: OCR cleanup"

./pdf-cli merge f1.pdf f2.pdf out.pdf &
PID=$!
sleep 1
kill -TERM $PID
sleep 1
ls /tmp/pdf-merge-* 2>&1 | grep -q "No such file" && echo "PASS: Merge cleanup" || echo "FAIL: Merge cleanup"
```

## Success Criteria

- internal/cleanup package created with Register/Run API
- Thread-safe cleanup registry with mutex protection
- cleanup.Run integrated in main.go defer chain
- All 8 temp file creation sites register with cleanup
- Unregister function prevents double cleanup on normal path
- Unit tests verify concurrent access and idempotency
- Integration tests verify signal-based cleanup works
- Manual tests confirm no temp file leaks on SIGINT/SIGTERM
- Race detector reports no issues
- go test ./... passes
- No temp files remain after signal interruption

## Dependencies

Depends on Plan 1.1 completion because:
- Error handling patterns from 1.1 should be applied to cleanup errors
- cleanup.Run should use proper error propagation (though typically logged, not returned)
- Integration approach follows error handling testing patterns
