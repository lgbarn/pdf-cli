# Phase 5: Code Quality Improvements

## Overview

Replace magic numbers with named constants, make parallelism thresholds configurable and adaptive, and make CI coverage checks portable.

## Requirements

- **R14**: Magic numbers replaced with named constants
- **R15**: Logging consolidated to slog (**SKIPPED** - already complete)
- **R16**: Coverage tooling portable (no bc/awk)
- **R17**: Parallelism thresholds configurable or adaptive to runtime.NumCPU()

## Plans

### Plan 01: Code Quality Improvements (Wave 1)

Single plan with 3 tasks covering all active requirements.

**Files Modified:**
- `internal/ocr/ocr.go` - Add constants for thresholds, timeout, permissions; accept PerformanceConfig
- `internal/pdf/text.go` - Add constants for threshold, progress interval; accept PerformanceConfig
- `internal/fileio/files.go` - Add constants for permissions, validation threshold
- `internal/config/config.go` - Add PerformanceConfig with adaptive defaults
- `.github/workflows/ci.yaml` - Use Go script instead of awk
- `scripts/coverage-check.go` - New portable coverage checker

**Task Breakdown:**
1. **Replace magic numbers** - Extract all hardcoded values to named constants
2. **Configurable parallelism** - Add PerformanceConfig with adaptive defaults based on runtime.NumCPU()
3. **Portable coverage check** - Replace awk/bc with pure Go script

## Research Findings

### R14: Magic Numbers

**Parallelism:**
- Threshold `5` in `internal/ocr/ocr.go:299` and `internal/pdf/text.go:59`
- Max workers `8` in `internal/ocr/ocr.go:344`

**Timeouts:**
- Download timeout `5*time.Minute` in `internal/ocr/ocr.go:175`

**File Permissions:**
- `0750` in `internal/fileio/files.go:28`, `internal/ocr/ocr.go:124`
- `0600` in `internal/config/config.go:139`

**Validation:**
- Threshold `3` in `internal/fileio/files.go:134`

**Progress:**
- Update interval `5` in `internal/pdf/text.go:70,119`

### R15: Logging (Already Complete)

Research shows logging already uses `slog` via `internal/logging` wrapper. Most remaining `fmt.Fprintf(os.Stderr, ...)` calls are intentional user-facing status messages, not logs. No action needed.

### R16: Coverage Tooling

Current CI (`.github/workflows/ci.yaml:51-61`):
```yaml
COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | tr -d '%')
if awk "BEGIN {exit !($COVERAGE < 75)}"; then
```

**Issues:**
- Depends on `awk` for parsing and float comparison
- Uses shell pipeline with `grep`, `tr`
- Not portable to Windows CI runners

**Solution:**
- Create `scripts/coverage-check.go` that parses `go tool cover -func` output
- Replace shell pipeline with: `go run scripts/coverage-check.go coverage.out 75`

### R17: Adaptive Parallelism

**Current State:**
- OCR parallel threshold: hardcoded `5` (line 299)
- Text extraction threshold: hardcoded `5` (line 59)
- Max workers: `min(runtime.NumCPU(), 8)` (line 344)

**Solution:**
Add `PerformanceConfig` to config:
```go
type PerformanceConfig struct {
    OCRParallelThreshold  int `yaml:"ocr_parallel_threshold"`
    TextParallelThreshold int `yaml:"text_parallel_threshold"`
    MaxWorkers            int `yaml:"max_workers"`
}

func DefaultPerformanceConfig() PerformanceConfig {
    numCPU := runtime.NumCPU()
    return PerformanceConfig{
        OCRParallelThreshold:  max(5, numCPU/2),
        TextParallelThreshold: max(5, numCPU/2),
        MaxWorkers:            min(numCPU, 8),
    }
}
```

**Benefits:**
- Adapts to available CPU count at runtime
- Configurable via YAML or environment variables
- Maintains backward compatibility (same behavior on 10+ CPU systems)
- Better performance on systems with fewer CPUs

## Complexity

**Small** - Single plan with focused changes:
- Constants extraction is straightforward refactoring
- Config addition follows existing patterns
- Coverage script is ~50 lines of simple Go

## Dependencies

None - can start immediately.

## Success Criteria

- [ ] All magic numbers replaced with named constants
- [ ] PerformanceConfig added with adaptive defaults
- [ ] Performance settings configurable via YAML and env vars
- [ ] CI coverage check uses pure Go (no awk/bc)
- [ ] All tests pass
- [ ] No behavioral changes for default configuration
- [ ] Documentation updated with new config options

## Notes

- R15 (logging consolidation) skipped - already complete per research
- Backward compatibility maintained - defaults only change on systems with <10 CPUs
- Environment variables follow existing pattern: `PDF_CLI_PERF_*`
