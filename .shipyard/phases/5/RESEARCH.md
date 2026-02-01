# Phase 5: Code Quality Improvements - Research Document

**Date:** 2026-01-31
**Phase:** 5 - Code Quality Improvements
**Status:** Research Complete

## Executive Summary

This research document covers four code quality improvements for pdf-cli:
1. **R14:** Replace magic numbers with named constants
2. **R15:** Consolidate logging to slog
3. **R16:** Make coverage tooling portable (remove bc/awk dependencies)
4. **R17:** Make parallelism thresholds configurable/adaptive

All improvements are achievable with low to moderate effort. The codebase already uses slog in a centralized logging package, simplifying R15. Several magic numbers exist across files that should be extracted to constants. The CI coverage check uses awk, which can be replaced with pure Go tooling. Parallelism thresholds are currently hardcoded but can be made adaptive to `runtime.NumCPU()`.

---

## 1. Technology Options

### R14: Magic Numbers → Named Constants

**Option A: Define constants in each package**
- Pros: Co-located with usage, clear context
- Cons: May duplicate similar values across packages
- Maturity: Standard Go practice

**Option B: Centralized constants package**
- Pros: Single source of truth, easy to audit
- Cons: Can become a dumping ground, less contextual
- Maturity: Common in larger codebases

**Option C: Configuration-driven (via config package)**
- Pros: User-configurable without rebuilding
- Cons: Overkill for internal thresholds
- Maturity: Appropriate for user-facing settings

### R15: Logging Consolidation to slog

**Option A: Continue using existing internal/logging wrapper**
- Pros: Already implemented, consistent API
- Cons: None identified
- Maturity: The codebase already uses this approach

**Option B: Direct slog usage throughout codebase**
- Pros: One less abstraction layer
- Cons: Less control over global behavior
- Maturity: Standard since Go 1.21

### R16: Portable Coverage Tooling

**Option A: Pure shell script with go tool cover**
- Pros: No external dependencies beyond Go
- Cons: Shell arithmetic still needed
- Maturity: Built into Go toolchain
- Example:
  ```bash
  COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | tr -d '%')
  if awk "BEGIN {exit !($COVERAGE < 75)}"; then
    echo "Coverage too low"
  fi
  ```

**Option B: vladopajic/go-test-coverage tool**
- Pros: Pure Go, no shell dependencies, declarative config, GitHub Action support
- Cons: External dependency, requires config file
- Maturity: Well-maintained, 800+ stars on GitHub
- Repo: https://github.com/vladopajic/go-test-coverage
- Example usage:
  ```yaml
  - name: check test coverage
    uses: vladopajic/go-test-coverage@v2
    with:
      profile: cover.out
      threshold-total: 75
  ```

**Option C: Custom Go script**
- Pros: Full control, zero external deps
- Cons: Maintenance burden, reinventing the wheel
- Maturity: N/A (would be new code)

### R17: Adaptive Parallelism Thresholds

**Option A: Hardcode based on runtime.NumCPU()**
- Pros: Automatic, no configuration needed
- Cons: May not be optimal for all workloads
- Example: `workers := runtime.NumCPU()`

**Option B: Formula-based (e.g., min(NumCPU, 8))**
- Pros: Caps maximum workers to avoid resource exhaustion
- Cons: Still not user-configurable
- Example: `workers := min(runtime.NumCPU(), 8)`

**Option C: Configuration-driven with adaptive defaults**
- Pros: User can override, sensible defaults
- Cons: More complex implementation
- Config fields: `MaxWorkers`, `ParallelThreshold`

---

## 2. Recommended Approach

### R14: Magic Numbers
**Recommendation: Option A (Package-level constants)**

Define constants at the top of each file where they're used. This keeps the context clear and follows standard Go conventions.

**Why not Option B?** A centralized constants package becomes harder to navigate and may not provide enough context for each constant's purpose.

**Why not Option C?** Most magic numbers are internal implementation details (file permissions, timeouts) that users shouldn't configure.

### R15: Logging Consolidation
**Recommendation: Option A (Continue with internal/logging wrapper)**

The codebase already has a well-structured logging package that wraps slog. The work is to replace the remaining `fmt.Fprintf(os.Stderr, ...)` calls with `logging.Info()` or `logging.Error()`.

**Why not Option B?** The wrapper provides useful abstractions like global logger management and consistent error handling.

### R16: Coverage Tooling
**Recommendation: Option B (vladopajic/go-test-coverage)**

This is the cleanest solution for portable coverage checking without shell dependencies. It's pure Go, has GitHub Action support, and is actively maintained.

**Why not Option A?** Still requires awk for floating-point comparison, defeating the portability goal.

**Why not Option C?** Reinventing functionality that already exists in a well-tested tool.

### R17: Adaptive Parallelism
**Recommendation: Option C (Configuration-driven with adaptive defaults)**

Add config fields to `internal/config/config.go`:
- `Performance.MaxWorkers` (default: `runtime.NumCPU()`)
- `Performance.ParallelThreshold` (default: 5 for OCR, 5 for text extraction)

This provides the best balance of sensible defaults and user control.

**Why not Option A?** Container environments may have CPU limits that don't match `runtime.NumCPU()`.

**Why not Option B?** Hardcoded caps (like 8) are arbitrary and may not suit all environments.

---

## 3. Potential Risks and Mitigations

### R14: Magic Numbers

**Risk:** Over-extraction creates constant clutter
**Mitigation:** Only extract values that are:
- Used multiple times, OR
- Have business/domain meaning, OR
- May need to change in the future

**Risk:** Breaking existing behavior by misidentifying values
**Mitigation:** Add unit tests for edge cases before refactoring

### R15: Logging Consolidation

**Risk:** Changing log output breaks user scripts that parse stderr
**Mitigation:**
- Use structured logging only for internal/debugging logs
- Keep user-facing messages (e.g., "Compressed to X") as simple `fmt.Fprintf`
- Document logging levels and formats

**Risk:** Performance regression from structured logging
**Mitigation:** slog is performant (~650ns/op), but avoid logging in hot loops

### R16: Coverage Tooling

**Risk:** New tool has different threshold calculation
**Mitigation:** Test locally before merging; verify threshold values match

**Risk:** GitHub Action version incompatibility
**Mitigation:** Pin to specific version (e.g., `@v2`) and test in CI

### R17: Adaptive Parallelism

**Risk:** `runtime.NumCPU()` in containers returns node CPU count, not container limit
**Mitigation:**
- Document that users in containerized environments should set `PDF_CLI_MAX_WORKERS` env var
- Consider reading from cgroup limits (Go 1.19+ runtime already does this for GOMAXPROCS)

**Risk:** Too many workers cause memory exhaustion
**Mitigation:** Cap default at reasonable maximum (e.g., 16 workers) even if NumCPU is higher

**Risk:** Changed defaults alter performance characteristics
**Mitigation:** Maintain backward-compatible defaults (5 for thresholds, NumCPU capped at 8 for workers)

---

## 4. Relevant Documentation Links

### R14: Magic Numbers
- [Effective Go - Constants](https://go.dev/doc/effective_go#constants)
- [Go Code Review Comments - Named Result Parameters](https://github.com/golang/go/wiki/CodeReviewComments)

### R15: Logging with slog
- [Official slog Package Documentation](https://pkg.go.dev/log/slog)
- [Go Blog: Structured Logging with slog](https://go.dev/blog/slog)
- [Logging in Go with Slog: The Ultimate Guide](https://betterstack.com/community/guides/logging/logging-in-go/)
- [Deep Dive and Migration Guide to slog](https://leapcell.io/blog/deep-dive-and-migration-guide-to-go-1-21-s-structured-logging-with-slog)
- [Logging in Go with Slog: A Practitioner's Guide](https://www.dash0.com/guides/logging-in-go-with-slog)
- [Complete Guide to Logging in Golang with slog](https://signoz.io/guides/golang-slog/)

### R16: Coverage Tooling
- [go tool cover Documentation](https://pkg.go.dev/cmd/cover)
- [vladopajic/go-test-coverage GitHub](https://github.com/vladopajic/go-test-coverage)
- [go-test-coverage GitHub Action](https://github.com/marketplace/actions/go-test-coverage)
- [go-test-coverage Configuration Example](https://github.com/vladopajic/go-test-coverage/blob/main/.testcoverage.example.yml)
- [Coverage profiling support for integration tests](https://go.dev/doc/build-cover)

### R17: Adaptive Parallelism
- [runtime.NumCPU() Documentation](https://pkg.go.dev/runtime#NumCPU)
- [3 rules for efficient parallel computation](https://yourbasic.org/golang/efficient-parallel-computation/)
- [Understanding Goroutines, Concurrency, and Parallelism in Go](https://dev.to/lovestaco/understanding-goroutines-concurrency-and-parallelism-in-go-355d)
- [CPU throttling for containerized Go applications](https://kanishk.io/posts/cpu-throttling-in-containerized-go-apps/)
- [runtime: CPU limit-aware GOMAXPROCS default](https://github.com/golang/go/issues/73193)

---

## 5. Implementation Considerations

### R14: Magic Numbers Inventory

Based on codebase analysis, here are the magic numbers found:

| File | Line | Value | Meaning | Proposed Constant |
|------|------|-------|---------|-------------------|
| `internal/ocr/ocr.go` | 299 | `5` | Parallel threshold for OCR image processing | `DefaultOCRParallelThreshold` |
| `internal/ocr/ocr.go` | 344 | `8` | Max concurrent OCR workers | `DefaultMaxOCRWorkers` |
| `internal/ocr/ocr.go` | 175 | `5*time.Minute` | Tessdata download timeout | `TessdataDownloadTimeout` |
| `internal/pdf/text.go` | 59 | `5` | Parallel threshold for text extraction | `DefaultTextParallelThreshold` |
| `internal/fileio/files.go` | 134 | `3` | Validation parallel threshold | `ValidationParallelThreshold` |
| `internal/fileio/files.go` | 28, 124 | `0750` | Directory permissions | `DefaultDirPerms` |
| `internal/fileio/files.go` | 188-190 | `1024`, `KB*1024`, `MB*1024` | File size units | Already defined as `KB`, `MB`, `GB` |
| `internal/config/config.go` | 130, 139 | `0750`, `0600` | Config dir/file permissions | `ConfigDirPerms`, `ConfigFilePerms` |
| `internal/progress/progress.go` | N/A | Threshold passed by caller | Progress bar display threshold | Move to config |

**File permissions (0750, 0600):** These appear in multiple places and should be centralized.

**Progress bar thresholds (1, 5):** Currently passed as arguments. Consider moving to config or keeping as caller responsibility.

### R15: Logging Locations to Update

Found 9 occurrences of `fmt.Fprintf(os.Stderr, ...)` that should be evaluated:

| File | Line | Current Code | Replacement Strategy |
|------|------|--------------|---------------------|
| `internal/ocr/ocr.go` | 173 | `fmt.Fprintf(os.Stderr, "Downloading...")` | Keep as user-facing status message OR use `logging.Info` |
| `internal/cli/cli.go` | 115, 121, 127 | Dry-run, verbose, status prints | Keep as-is (user-facing output, not logs) |
| `internal/cli/cli.go` | 133 | Progress message | Keep as-is (user-facing) |
| `internal/commands/*.go` | Various | Success messages like "Compressed to X" | Keep as-is (user-facing output) |

**Recommendation:** Most of these are **user-facing status messages**, not logs. Only the tessdata download message should potentially use structured logging. The current approach is appropriate.

**Clarification needed:** Should user-facing messages go to stderr (current) or stdout? Convention suggests:
- **stdout:** Primary output (extracted text, info command JSON/CSV)
- **stderr:** Status messages, progress, errors

Current usage follows this convention correctly.

### R16: CI Coverage Implementation

**Current CI workflow (lines 51-61):**
```yaml
- name: Check coverage threshold (75%)
  run: |
    COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | tr -d '%')
    echo "Total coverage: ${COVERAGE}%"
    if awk "BEGIN {exit !($COVERAGE < 75)}"; then
      echo "::error::Coverage ${COVERAGE}% is below 75% threshold"
      exit 1
    else
      echo "Coverage ${COVERAGE}% meets 75% threshold"
    fi
```

**Problems:**
- Uses `awk` for both extraction and comparison
- Not portable to environments without awk (Windows, minimal containers)

**Proposed replacement:**
```yaml
- name: Install go-test-coverage
  run: go install github.com/vladopajic/go-test-coverage/v2@latest

- name: Check coverage threshold
  run: |
    go-test-coverage --config=./.testcoverage.yml
```

**Configuration file (`.testcoverage.yml`):**
```yaml
profile: coverage.out
threshold:
  total: 75
```

**Alternative (inline config):**
```yaml
- name: Check coverage threshold
  uses: vladopajic/go-test-coverage@v2
  with:
    profile: coverage.out
    threshold-total: 75
```

**Recommendation:** Use GitHub Action approach for simplicity.

### R17: Configuration Schema Changes

Add new section to `internal/config/config.go`:

```go
// PerformanceConfig holds performance tuning settings.
type PerformanceConfig struct {
    MaxWorkers         int `yaml:"max_workers"`          // Max concurrent workers (0 = auto)
    OCRParallelThreshold    int `yaml:"ocr_parallel_threshold"`    // Min images for parallel OCR
    TextParallelThreshold   int `yaml:"text_parallel_threshold"`   // Min pages for parallel text extraction
}
```

**Default values:**
```go
Performance: PerformanceConfig{
    MaxWorkers:              0, // 0 = auto-detect, capped at 8
    OCRParallelThreshold:    5,
    TextParallelThreshold:   5,
},
```

**Environment variable overrides:**
- `PDF_CLI_MAX_WORKERS`
- `PDF_CLI_OCR_PARALLEL_THRESHOLD`
- `PDF_CLI_TEXT_PARALLEL_THRESHOLD`

**Runtime calculation:**
```go
func getMaxWorkers(config int) int {
    if config > 0 {
        return config
    }
    // Auto-detect: use NumCPU capped at 8
    return min(runtime.NumCPU(), 8)
}
```

**Why cap at 8?** Based on current code (`internal/ocr/ocr.go:344`), the existing hardcoded limit is 8. This prevents resource exhaustion on high-core-count machines.

**Container awareness:** Modern Go runtime (1.19+) already makes GOMAXPROCS container-aware via cgroup limits. Users can override via env var if needed.

### Integration Points

**Files to modify:**

1. **Magic numbers:**
   - `internal/ocr/ocr.go` - Extract thresholds and timeouts
   - `internal/pdf/text.go` - Extract parallel threshold
   - `internal/fileio/files.go` - Extract permissions and validation threshold
   - `internal/config/config.go` - Extract permissions

2. **Logging:**
   - Review `internal/ocr/ocr.go:173` - Decide on structured vs user-facing
   - No other changes needed (other uses are already appropriate)

3. **Coverage tooling:**
   - `.github/workflows/ci.yaml` - Replace awk-based check
   - `.testcoverage.yml` - Add config file

4. **Parallelism:**
   - `internal/config/config.go` - Add PerformanceConfig
   - `internal/ocr/ocr.go` - Use config for thresholds and workers
   - `internal/pdf/text.go` - Use config for threshold

### Migration Concerns

**Backward compatibility:** Adding config fields with sensible defaults maintains backward compatibility. Existing users see no behavior change unless they opt into configuration.

**Performance implications:**
- Adaptive worker count may increase parallelism on high-core machines (good)
- Configurable thresholds allow users to tune for their workload (good)
- No performance regressions expected

### Testing Strategy

**Unit tests:**
- Test constant values are used correctly
- Test config loading with performance settings
- Test worker calculation with various NumCPU values
- Test threshold logic with edge cases (0, 1, threshold-1, threshold, threshold+1)

**Integration tests:**
- Verify CI coverage check works with new tooling
- Test environment variable overrides
- Test parallel processing at various thresholds

**Manual testing:**
- Run on machines with different core counts
- Run in Docker containers with CPU limits
- Verify log output format is preserved

---

## 6. Inconclusive Areas / Further Investigation

### Logging Strategy Clarification

**Question:** Should the tessdata download message use structured logging or remain as user-facing output?

**Current:** `fmt.Fprintf(os.Stderr, "Downloading tessdata for '%s'...\n", lang)`

**Options:**
1. Keep as-is (user-facing status message)
2. Change to `logging.Info("downloading tessdata", "language", lang)`
3. Use both: structured log + user message

**Recommendation:** Keep as-is. Users need to know what's happening during a potentially long download. Structured logs are better for debugging, not user communication.

### Progress Bar Threshold Configurability

**Question:** Should progress bar display thresholds be configurable?

**Current:** Thresholds passed to `NewProgressBar()` vary by operation (1 for OCR, 5 for text extraction, 0 for merge/split)

**Consideration:** These are UX decisions (when to show progress), not performance tuning. Making them configurable may be overkill.

**Recommendation:** Leave as-is unless users request configurability.

### Container CPU Limit Detection

**Question:** Should we explicitly detect cgroup CPU limits, or rely on runtime.NumCPU()?

**Current state:** Go 1.19+ runtime already adjusts GOMAXPROCS based on cgroup limits. However, `runtime.NumCPU()` still returns the node CPU count.

**Options:**
1. Read `/sys/fs/cgroup/cpu.max` directly (Linux-specific)
2. Trust users to set environment variable in containers
3. Document the limitation

**Recommendation:** Option 3. Document that containerized deployments should set `PDF_CLI_MAX_WORKERS` environment variable. Reading cgroups directly is complex and platform-specific.

---

## 7. Summary Table

| Requirement | Effort | Risk | Dependencies | Recommended Tool/Approach |
|-------------|--------|------|--------------|---------------------------|
| R14: Magic numbers | Low | Low | None | Package-level constants |
| R15: Logging consolidation | Low | Low | Existing `internal/logging` | Keep current approach, minimal changes |
| R16: Portable coverage | Low | Low | `vladopajic/go-test-coverage@v2` | GitHub Action + config file |
| R17: Adaptive parallelism | Medium | Medium | `internal/config` | Config-driven with adaptive defaults |

**Estimated total effort:** 1-2 days for implementation + testing

---

## Appendix: Code Examples

### Example: Extracting Magic Numbers in ocr.go

**Before:**
```go
const parallelThreshold = 5

func (e *Engine) processImages(...) {
    workers := min(runtime.NumCPU(), 8)
    // ...
}
```

**After:**
```go
const (
    // DefaultOCRParallelThreshold is the minimum number of images to trigger parallel processing
    DefaultOCRParallelThreshold = 5

    // DefaultMaxOCRWorkers is the default maximum number of concurrent OCR workers
    DefaultMaxOCRWorkers = 8
)

func (e *Engine) processImages(...) {
    cfg := config.Get()
    threshold := cfg.Performance.OCRParallelThreshold
    maxWorkers := getMaxWorkers(cfg.Performance.MaxWorkers)
    workers := min(runtime.NumCPU(), maxWorkers)
    // ...
}

func getMaxWorkers(configured int) int {
    if configured > 0 {
        return configured
    }
    return DefaultMaxOCRWorkers
}
```

### Example: .testcoverage.yml Configuration

```yaml
# Minimum coverage threshold
profile: coverage.out

threshold:
  total: 75

# Optional: per-file/package thresholds
# threshold:
#   file: 70
#   package: 75
#   total: 75

# Optional: exclude patterns
# exclude:
#   paths:
#     - .*_test\.go
#     - internal/testing
```

---

## Sources

- [cover command - cmd/cover - Go Packages](https://pkg.go.dev/cmd/cover)
- [GitHub - vladopajic/go-test-coverage](https://github.com/vladopajic/go-test-coverage)
- [go-test-coverage GitHub Action](https://github.com/marketplace/actions/go-test-coverage)
- [slog package - log/slog - Go Packages](https://pkg.go.dev/log/slog)
- [Logging in Go with Slog: The Ultimate Guide](https://betterstack.com/community/guides/logging/logging-in-go/)
- [Deep Dive and Migration Guide to Go 1.21+'s slog](https://leapcell.io/blog/deep-dive-and-migration-guide-to-go-1-21-s-structured-logging-with-slog)
- [Structured Logging with slog - The Go Programming Language](https://go.dev/blog/slog)
- [3 rules for efficient parallel computation](https://yourbasic.org/golang/efficient-parallel-computation/)
- [Understanding Goroutines, Concurrency, and Parallelism in Go](https://dev.to/lovestaco/understanding-goroutines-concurrency-and-parallelism-in-go-355d)
- [runtime: CPU limit-aware GOMAXPROCS default](https://github.com/golang/go/issues/73193)
- [CPU throttling for containerized Go applications](https://kanishk.io/posts/cpu-throttling-in-containerized-go-apps/)
- [runtime package - runtime - Go Packages](https://pkg.go.dev/runtime)
