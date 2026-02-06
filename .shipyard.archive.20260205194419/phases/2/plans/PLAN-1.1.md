# Plan 1.1: Thread-Safe Singletons

## Context
This plan makes the global configuration and logging singletons thread-safe by replacing bare nil-checks with sync.RWMutex-protected initialization. Currently, both `config.Get()/Reset()` and `logging.Get()/Init()/Reset()` use unsynchronized global variables, which creates data races when accessed concurrently.

This is foundational work that must complete before context propagation (Plan 2.1), since later race testing depends on stable global state.

## Dependencies
None. This is the first plan in Phase 2, Wave 1.

## Tasks

### Task 1: Add thread-safe initialization to config package
**Files:** `/Users/lgbarn/Personal/pdf-cli/internal/config/config.go`

**Action:** Modify

**Description:**
Add sync.RWMutex to protect the global config singleton:

1. Add import: `"sync"`
2. Add package-level variables after line 142:
   ```go
   var globalMu sync.RWMutex
   ```
3. Replace `Get()` function (lines 144-155) with double-checked locking pattern:
   ```go
   func Get() *Config {
       globalMu.RLock()
       if global != nil {
           defer globalMu.RUnlock()
           return global
       }
       globalMu.RUnlock()

       globalMu.Lock()
       defer globalMu.Unlock()

       // Double-check after acquiring write lock
       if global != nil {
           return global
       }

       var err error
       global, err = Load()
       if err != nil {
           global = DefaultConfig()
       }
       return global
   }
   ```
4. Replace `Reset()` function (lines 157-160) with mutex protection:
   ```go
   func Reset() {
       globalMu.Lock()
       defer globalMu.Unlock()
       global = nil
   }
   ```

**Acceptance Criteria:**
- `Get()` uses RLock for fast path, upgrades to Lock for initialization
- `Reset()` acquires Lock before modifying global
- Double-checked locking prevents race condition between check and set
- No functional changes to Load() or DefaultConfig()

### Task 2: Add thread-safe initialization to logging package
**Files:** `/Users/lgbarn/Personal/pdf-cli/internal/logging/logger.go`

**Action:** Modify

**Description:**
Add sync.RWMutex to protect the global logger singleton:

1. Add import: `"sync"` (combine with existing imports)
2. Add package-level variable after line 83:
   ```go
   var globalMu sync.RWMutex
   ```
3. Replace `Init()` function (lines 85-88) with mutex protection:
   ```go
   func Init(level Level, format Format) {
       globalMu.Lock()
       defer globalMu.Unlock()
       global = New(level, format, os.Stderr)
   }
   ```
4. Replace `Get()` function (lines 119-125) with double-checked locking:
   ```go
   func Get() *Logger {
       globalMu.RLock()
       if global != nil {
           defer globalMu.RUnlock()
           return global
       }
       globalMu.RUnlock()

       globalMu.Lock()
       defer globalMu.Unlock()

       // Double-check after acquiring write lock
       if global != nil {
           return global
       }

       Init(LevelSilent, FormatText)
       return global
   }
   ```
5. Replace `Reset()` function (lines 127-130) with mutex protection:
   ```go
   func Reset() {
       globalMu.Lock()
       defer globalMu.Unlock()
       global = nil
   }
   ```

**Acceptance Criteria:**
- `Get()` uses RLock for fast path, upgrades to Lock for initialization
- `Init()` and `Reset()` acquire Lock before modifying global
- Double-checked locking prevents race condition
- Package-level helper functions (Debug, Info, Warn, Error, With) remain unchanged

### Task 3: Verify thread safety with race detector
**Files:** N/A (test execution only)

**Action:** Test

**Description:**
Run the full test suite with the race detector to verify no data races exist in config or logging packages:

1. Run: `go test -race ./internal/config/... -v`
   - Verify all tests pass with zero data races
   - Config tests call Reset() at lines 30, 54, 96, 132, 248 - these must be race-free
2. Run: `go test -race ./internal/logging/... -v`
   - Verify all tests pass with zero data races
   - Logging tests call Reset() at lines 188, 207, 224, 241, 270 - these must be race-free
3. Run: `go test -race ./... -short`
   - Verify entire codebase has zero data races
   - Focus on packages that call config.Get() or logging.Get()

**Acceptance Criteria:**
- Zero data races reported by `go test -race`
- All existing tests pass without modification
- No performance regression (mutex overhead is minimal for singleton pattern)

## Verification

**Command sequence:**
```bash
# Verify thread safety
go test -race ./internal/config/... -v
go test -race ./internal/logging/... -v
go test -race ./... -short

# Verify no functional regressions
go test ./internal/config/... -v
go test ./internal/logging/... -v

# Verify build succeeds
go build ./cmd/pdf
```

**Success criteria:**
- All tests pass
- Race detector reports zero data races
- Binary builds successfully
- No changes to public API or CLI behavior
