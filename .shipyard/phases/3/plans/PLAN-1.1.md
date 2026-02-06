---
phase: phase-3-concurrency-error-handling
plan: 1.1
wave: 1
dependencies: []
must_haves:
  - R7: Cleanup registry uses map[string]struct{} (no idx variable in Register function)
  - Existing tests pass with map-based implementation
  - No changes needed to callers of Register/Unregister
files_touched:
  - internal/cleanup/cleanup.go
  - internal/cleanup/cleanup_test.go
tdd: false
---

# Plan 1.1: Cleanup Registry Map Conversion (R7)

## Overview
Convert cleanup registry from slice-based to map-based tracking to eliminate index invalidation issues and improve semantic clarity. The current slice approach has a fragile contract where unregister calls after Run() are silently ignored due to index invalidation.

## Tasks

<task id="1" files="internal/cleanup/cleanup.go" tdd="false">
  <action>
Replace slice-based path tracking with map-based tracking in cleanup registry:

1. Line 12: Change `paths []string` to `paths map[string]struct{}`
2. Initialize in Register function (around line 21): Add `if paths == nil { paths = make(map[string]struct{}) }`
3. Register function (lines 20-34): Replace index-based logic with map operations:
   - Remove `idx := len(paths)` (line 24)
   - Replace `paths = append(paths, path)` with `paths[path] = struct{}{}`
   - Update unregister closure to use `delete(paths, path)` instead of `if idx < len(paths) { paths[idx] = "" }`
4. Run function (lines 38-59): Update iteration logic:
   - Remove reverse iteration (no LIFO requirement with map)
   - Replace `for i := len(paths) - 1; i >= 0; i--` with `for p := range paths`
   - Replace `p := paths[i]` and `if p == "" { continue }` with direct map iteration
   - Keep `paths = nil` to clear registry after cleanup
5. Reset function (line 66): Change `paths = nil` to clear map (no change needed, nil works for both)

Expected diff pattern:
```go
var (
    mu     sync.Mutex
    paths  map[string]struct{}  // was: []string
    hasRun bool
)

func Register(path string) func() {
    mu.Lock()
    defer mu.Unlock()

    if paths == nil {
        paths = make(map[string]struct{})
    }
    paths[path] = struct{}{}

    return func() {
        mu.Lock()
        defer mu.Unlock()
        delete(paths, path)
    }
}

func Run() error {
    mu.Lock()
    defer mu.Unlock()

    if hasRun {
        return nil
    }
    hasRun = true

    var firstErr error
    for p := range paths {
        if err := os.RemoveAll(p); err != nil && firstErr == nil {
            firstErr = err
        }
    }
    paths = nil
    return firstErr
}
```
  </action>
  <verify>go test -v -race /Users/lgbarn/Personal/pdf-cli/internal/cleanup/... && go test -race /Users/lgbarn/Personal/pdf-cli/internal/pdf/... /Users/lgbarn/Personal/pdf-cli/internal/ocr/... | grep -E "(PASS|FAIL)"</verify>
  <done>All existing cleanup tests pass. No idx variable exists in Register function. Map-based tracking successfully replaces slice-based approach. Callers in internal/pdf/text.go:185 and internal/ocr/ocr.go:236,350 continue to work without modification.</done>
</task>

<task id="2" files="internal/cleanup/cleanup_test.go" tdd="true">
  <action>
Add test case for "unregister after Run()" scenario to verify map-based implementation handles this edge case correctly:

Add new test function after TestRunIdempotent (after line 123):
```go
func TestUnregisterAfterRun(t *testing.T) {
    setup(t)

    dir := t.TempDir()
    f1 := filepath.Join(dir, "test.tmp")

    if err := os.WriteFile(f1, []byte("test"), 0600); err != nil {
        t.Fatal(err)
    }

    unregister := Register(f1)

    // Run cleanup first - file should be removed
    if err := Run(); err != nil {
        t.Fatalf("Run returned error: %v", err)
    }

    if _, err := os.Stat(f1); !os.IsNotExist(err) {
        t.Errorf("expected f1 to be removed by Run, got err=%v", err)
    }

    // Calling unregister after Run should be safe (no panic, no error)
    // This is a no-op with map-based implementation
    unregister()

    // Verify we can still call Reset and reuse the registry
    Reset()
    f2 := filepath.Join(dir, "test2.tmp")
    if err := os.WriteFile(f2, []byte("test2"), 0600); err != nil {
        t.Fatal(err)
    }
    Register(f2)

    if err := Run(); err != nil {
        t.Fatalf("second Run returned error: %v", err)
    }

    if _, err := os.Stat(f2); !os.IsNotExist(err) {
        t.Errorf("expected f2 to be removed, got err=%v", err)
    }
}
```

This test verifies:
1. Unregister after Run() doesn't panic (was a risk with slice approach)
2. The operation is safe even though the path is already gone
3. Registry can be Reset and reused after this sequence
  </action>
  <verify>go test -v -run TestUnregisterAfterRun /Users/lgbarn/Personal/pdf-cli/internal/cleanup/...</verify>
  <done>TestUnregisterAfterRun passes. Test confirms that calling unregister after Run() is safe with map-based implementation, and registry can be Reset and reused afterward.</done>
</task>

<task id="3" files="internal/cleanup/cleanup.go" tdd="false">
  <action>
Verify cleanup registry map conversion by running race detector across all packages that use the cleanup registry:

Run comprehensive race detection tests:
1. Test cleanup package itself
2. Test pdf package (uses cleanup.Register at text.go:185)
3. Test ocr package (uses cleanup.Register at ocr.go:236,350)

This ensures:
- No race conditions in map access (all protected by mutex)
- Callers work correctly with new map-based implementation
- Concurrent registration still safe (tested by TestConcurrentRegister)
  </action>
  <verify>go test -race /Users/lgbarn/Personal/pdf-cli/internal/cleanup/... /Users/lgbarn/Personal/pdf-cli/internal/pdf/... /Users/lgbarn/Personal/pdf-cli/internal/ocr/... 2>&1 | tee /tmp/race-test.log && ! grep -i "DATA RACE" /tmp/race-test.log</verify>
  <done>All race detector tests pass with zero DATA RACE warnings. Map-based cleanup registry is thread-safe and all callers work correctly with the new implementation.</done>
</task>

## Success Criteria
- ✓ Cleanup registry uses `map[string]struct{}` instead of `[]string`
- ✓ No `idx` variable exists in Register function
- ✓ All existing tests pass (TestRegisterAndRun, TestUnregister, TestConcurrentRegister, TestRunIdempotent)
- ✓ New test TestUnregisterAfterRun passes
- ✓ Race detector shows zero warnings across cleanup, pdf, and ocr packages
- ✓ No changes needed to callers in internal/pdf/text.go or internal/ocr/ocr.go

## Notes
- The map-based approach eliminates the fragile index-based contract
- Unregister operations are now idempotent (deleting non-existent key is safe)
- LIFO ordering from Run() is lost, but this is acceptable (cleanup order rarely matters for temp files)
- Setting `paths = nil` after cleanup works for both slice and map (both become nil)
