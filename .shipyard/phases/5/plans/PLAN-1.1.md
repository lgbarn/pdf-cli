---
phase: code-quality-improvements
plan: 01
wave: 1
dependencies: []
must_haves:
  - R14: Magic numbers replaced with named constants
  - R16: Coverage tooling portable (no bc/awk)
  - R17: Parallelism thresholds configurable or adaptive to runtime.NumCPU()
files_touched:
  - internal/ocr/ocr.go
  - internal/pdf/text.go
  - internal/fileio/files.go
  - internal/config/config.go
  - .github/workflows/ci.yaml
  - scripts/coverage-check.go
tdd: false
---

# Plan 01: Code Quality Improvements

Replace magic numbers with named constants, make parallelism thresholds configurable, and make CI coverage check portable.

## Context

Current state:
- Magic numbers scattered across codebase (file permissions, thresholds, timeouts)
- Parallelism thresholds hardcoded to `5` in OCR and text extraction
- Max workers hardcoded to `min(runtime.NumCPU(), 8)`
- CI coverage check uses `awk` for float comparison (not portable)
- Logging already uses slog via internal/logging wrapper (R15 complete)

## Tasks

<task id="1" files="internal/ocr/ocr.go,internal/pdf/text.go,internal/fileio/files.go" tdd="false">
  <action>
Replace all magic numbers with named constants:

**internal/ocr/ocr.go:**
- `const DefaultParallelThreshold = 5` (line 299)
- `const DefaultMaxWorkers = 8` (line 344)
- `const DefaultDownloadTimeout = 5 * time.Minute` (line 175)
- `const DefaultDataDirPerm = 0750` (line 124)

**internal/pdf/text.go:**
- `const ParallelThreshold = 5` (line 59)
- `const ProgressUpdateInterval = 5` (line 70, 119)

**internal/fileio/files.go:**
- `const DefaultDirPerm = 0750` (line 28)
- `const DefaultFilePerm = 0600` (internal/config/config.go line 139)
- `const ParallelValidationThreshold = 3` (line 134)

Update all usage sites to reference the new constants.
  </action>
  <verify>
go test ./internal/ocr ./internal/pdf ./internal/fileio -v
grep -r "0750\|0600" internal/ | grep -v "const\|//"
grep -rE "([^a-zA-Z]5[^0-9]|=\s*5\s*\*)" internal/ocr internal/pdf internal/fileio | grep -v "const\|//"
  </verify>
  <done>
All tests pass, no standalone magic numbers remain in affected files (only const declarations).
  </done>
</task>

<task id="2" files="internal/config/config.go,internal/ocr/ocr.go,internal/pdf/text.go" tdd="false">
  <action>
Make parallelism thresholds configurable via PerformanceConfig:

1. Add to `internal/config/config.go`:
   ```go
   // PerformanceConfig holds performance tuning settings.
   type PerformanceConfig struct {
       OCRParallelThreshold  int `yaml:"ocr_parallel_threshold"`
       TextParallelThreshold int `yaml:"text_parallel_threshold"`
       MaxWorkers            int `yaml:"max_workers"`
   }
   ```

2. Add to Config struct:
   ```go
   Performance PerformanceConfig `yaml:"performance"`
   ```

3. Add DefaultPerformanceConfig() helper:
   ```go
   func DefaultPerformanceConfig() PerformanceConfig {
       numCPU := runtime.NumCPU()
       return PerformanceConfig{
           OCRParallelThreshold:  max(5, numCPU/2),
           TextParallelThreshold: max(5, numCPU/2),
           MaxWorkers:            min(numCPU, 8),
       }
   }
   ```

4. Update DefaultConfig() to call DefaultPerformanceConfig()

5. Add env overrides in applyEnvOverrides():
   - PDF_CLI_PERF_OCR_THRESHOLD
   - PDF_CLI_PERF_TEXT_THRESHOLD
   - PDF_CLI_PERF_MAX_WORKERS

6. Update OCR engine to accept PerformanceConfig:
   - Add to EngineOptions: `PerfConfig *PerformanceConfig`
   - Use in processImages() instead of const parallelThreshold
   - Use in processImagesParallel() instead of const DefaultMaxWorkers

7. Update pdf.ExtractTextWithProgress() to accept PerformanceConfig and use it

8. Update all callers (cmd/pdf commands) to pass config.Get().Performance
  </action>
  <verify>
go test ./internal/config ./internal/ocr ./internal/pdf -v
go run ./cmd/pdf text testdata/sample.pdf --help
PDF_CLI_PERF_MAX_WORKERS=4 go run ./cmd/pdf text testdata/sample.pdf
  </verify>
  <done>
All tests pass, performance config can be set via YAML or env vars, OCR and text extraction respect the config values. Default values adapt to runtime.NumCPU().
  </done>
</task>

<task id="3" files=".github/workflows/ci.yaml,scripts/coverage-check.go" tdd="false">
  <action>
Replace awk-based coverage check with portable Go script:

1. Create `scripts/coverage-check.go`:
   ```go
   //go:build ignore

   package main

   import (
       "bufio"
       "fmt"
       "os"
       "strconv"
       "strings"
   )

   func main() {
       if len(os.Args) < 3 {
           fmt.Fprintf(os.Stderr, "Usage: %s <coverage-file> <threshold>\n", os.Args[0])
           os.Exit(1)
       }

       coverageFile := os.Args[1]
       threshold, err := strconv.ParseFloat(os.Args[2], 64)
       if err != nil {
           fmt.Fprintf(os.Stderr, "Invalid threshold: %v\n", err)
           os.Exit(1)
       }

       coverage, err := parseCoverage(coverageFile)
       if err != nil {
           fmt.Fprintf(os.Stderr, "Error parsing coverage: %v\n", err)
           os.Exit(1)
       }

       fmt.Printf("Total coverage: %.1f%%\n", coverage)

       if coverage < threshold {
           fmt.Fprintf(os.Stderr, "::error::Coverage %.1f%% is below %.0f%% threshold\n", coverage, threshold)
           os.Exit(1)
       }

       fmt.Printf("Coverage %.1f%% meets %.0f%% threshold\n", coverage, threshold)
   }

   func parseCoverage(filename string) (float64, error) {
       f, err := os.Open(filename)
       if err != nil {
           return 0, err
       }
       defer f.Close()

       scanner := bufio.NewScanner(f)
       for scanner.Scan() {
           line := scanner.Text()
           if strings.HasPrefix(line, "total:") {
               fields := strings.Fields(line)
               if len(fields) >= 3 {
                   pctStr := strings.TrimSuffix(fields[2], "%")
                   return strconv.ParseFloat(pctStr, 64)
               }
           }
       }
       return 0, fmt.Errorf("no coverage total found")
   }
   ```

2. Update `.github/workflows/ci.yaml` line 51-61:
   ```yaml
   - name: Check coverage threshold (75%)
     run: go run scripts/coverage-check.go coverage.out 75
   ```

3. Remove awk/tr/grep pipeline
  </action>
  <verify>
go run scripts/coverage-check.go coverage.out 75
go run scripts/coverage-check.go coverage.out 100  # Should fail
  </verify>
  <done>
Coverage check runs successfully in CI using only Go (no awk/bc/tr), properly detects coverage above/below threshold with appropriate exit codes.
  </done>
</task>

## Verification

```bash
# Run all tests
go test ./... -v

# Verify no magic numbers remain (except in consts)
grep -r "0750\|0600" internal/ cmd/ | grep -v "const\|//\|_test.go"

# Verify config defaults adapt to CPU count
go run -race ./cmd/pdf --version

# Verify coverage check works
go test -coverprofile=coverage.out ./...
go run scripts/coverage-check.go coverage.out 75

# Test environment overrides
PDF_CLI_PERF_MAX_WORKERS=2 go run ./cmd/pdf text testdata/sample.pdf
```

## Success Criteria

- All magic numbers replaced with descriptive named constants
- Performance config added with adaptive defaults based on runtime.NumCPU()
- Performance settings configurable via YAML config or env vars
- Coverage check portable (pure Go, no shell dependencies)
- All existing tests pass
- No behavioral changes (same defaults if runtime.NumCPU() >= 10)
