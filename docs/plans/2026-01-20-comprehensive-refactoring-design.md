# Comprehensive Refactoring Design

**Date:** 2026-01-20
**Status:** Approved
**Scope:** Architecture, reliability, developer experience, operational improvements

## Overview

This design addresses accumulated technical debt in pdf-cli through a phased refactoring that improves code organization, test coverage, developer experience, and adds operational features—all while maintaining backward compatibility.

## Goals

1. **Reliability** - Increase test coverage from 69.6% to 75%+
2. **Developer Experience** - Better linting, architecture docs, contribution workflow
3. **Operational Improvements** - Dry-run mode, config file support, structured logging
4. **Architecture** - Eliminate duplication, improve modularity, enhance testability

## Current State Assessment

| Area | Status | Issue |
|------|--------|-------|
| Package dependencies | Clean | No circular dependencies |
| Code duplication | High | 40% boilerplate in commands, stdin/stdout logic copied 7 times |
| pdf.go | Monolithic | 689 LOC mixing 6+ concerns |
| util/ package | Sprawling | 7 files with mixed responsibilities |
| Interfaces | Missing | No command interface, hard to test |
| Test infrastructure | Weak | No mocks, global state dependencies |

### Test Coverage Baseline

| Package | Current | Target |
|---------|---------|--------|
| cli | 98.0% | Maintain |
| util | 90.7% | Maintain |
| ocr | 75.5% | 80%+ |
| commands | 60.7% | 75%+ |
| pdf | 57.9% | 75%+ |
| **Overall** | **69.6%** | **75%+** |

---

## Architecture Changes

### 1. Extract Stdio Handler Pattern

**Problem:** 7 commands duplicate stdin/stdout handling (~30 LOC each = 210 LOC total).

**Solution:** Create reusable patterns in `internal/commands/patterns/`:

```
internal/commands/patterns/
├── stdio.go      # StdioHandler for input/output management
└── batch.go      # BatchProcessor for multi-file operations
```

**StdioHandler responsibilities:**
- Resolve input (file path or stdin to temp file)
- Manage output (file path or stdout from temp file)
- Handle cleanup automatically via defer
- Provide consistent error wrapping

**Affected commands:** compress, decrypt, encrypt, extract, pdfa, reorder, rotate

### 2. Split pdf.go (689 LOC → 6 files)

**New structure:**

```
internal/pdf/
├── pdf.go          # Config, public API surface (~80 LOC)
├── metadata.go     # GetInfo, PageCount, GetMetadata, SetMetadata (~100 LOC)
├── transform.go    # Merge, Split, ExtractPages, Rotate, Compress (~150 LOC)
├── encryption.go   # Encrypt, Decrypt (~80 LOC)
├── text.go         # Text extraction with fallback logic (~180 LOC)
├── watermark.go    # AddWatermark, AddImageWatermark (~60 LOC)
└── validation.go   # Validate, ValidateToBuffer (~40 LOC)
```

**Principles:**
- Public API remains unchanged
- Each file has single responsibility
- Related functions grouped together
- Internal helpers stay with their consumers

### 3. Reorganize util/ Package

**Current:** 7 files with mixed concerns in `internal/util/`

**New structure:**

```
internal/
├── fileio/         # File operations + stdio
│   ├── files.go    # ValidatePDFFile, AtomicWrite, etc.
│   ├── stdio.go    # ResolveInputPath, WriteToStdout
│   └── fileio_test.go
├── pages/          # Page range parsing
│   ├── parser.go   # ParsePageRanges, ExpandPageRanges
│   ├── reorder.go  # ParseReorderSequence
│   └── pages_test.go
├── output/         # Output formatting
│   ├── formatter.go # OutputFormatter, formats
│   ├── table.go    # Table rendering logic
│   └── output_test.go
├── errors/         # Error handling
│   ├── errors.go   # PDFError, WrapError
│   └── errors_test.go
└── progress/       # Progress bar utilities
    └── progress.go
```

**Migration strategy:**
- Create new packages with content from util/
- Update imports throughout codebase
- Remove old util/ package
- Ensure all tests pass

---

## Operational Features

### 4. Dry-Run Mode

**Implementation:** Global `--dry-run` flag that:
- Shows what would happen without executing
- Prints planned operations in structured format
- Works with all modifying commands

**Output example:**
```
$ pdf-cli merge --dry-run file1.pdf file2.pdf -o combined.pdf
[dry-run] Would merge 2 files:
  - file1.pdf (12 pages)
  - file2.pdf (8 pages)
[dry-run] Output: combined.pdf (20 pages total)
```

**Affected commands:** merge, split, rotate, compress, encrypt, decrypt, watermark, meta set, extract, reorder

### 5. Config File Support

**Location:** `~/.config/pdf-cli/config.yaml` (XDG compliant)

**Supported settings:**
```yaml
defaults:
  output_format: json      # json, csv, tsv, human
  verbose: false
  show_progress: true

compress:
  quality: medium          # low, medium, high

encrypt:
  algorithm: aes256        # aes128, aes256

ocr:
  language: eng
  backend: auto            # auto, native, wasm
```

**New package:** `internal/config/`
- Load config at startup
- Config values are defaults, CLI flags override
- Support environment variable overrides

### 6. Structured Logging

**Flags:** `--log-level` and `--log-format`

**Log levels:** `debug`, `info`, `warn`, `error`, `silent` (default: `silent`)

**Log format:** `text` (human), `json` (machine-parseable)

**New package:** `internal/logging/`
- Wraps `log/slog` (Go 1.21+)
- Contextual fields (operation, file, duration)
- Writes to stderr (stdout reserved for command output)

**Example:**
```
$ pdf-cli merge --log-level=debug --log-format=json file1.pdf file2.pdf
{"level":"debug","msg":"validating input","file":"file1.pdf","pages":12}
{"level":"info","msg":"merge complete","output":"merged.pdf","duration_ms":245}
```

---

## Developer Experience

### 7. Enhanced Linting

**Updated `.golangci.yaml`:**
```yaml
linters:
  enable:
    # Current
    - govet
    - ineffassign
    - staticcheck
    - unused
    # New
    - gofmt
    - goimports
    - misspell
    - gocritic
    - revive
    - errcheck
    - gosimple
    - typecheck
```

### 8. Test Infrastructure

**New package:** `internal/testing/`
- `mock_pdf.go` - Mock PDF operations for command testing
- `mock_ocr.go` - Mock OCR backend (no tessdata downloads)
- `fixtures.go` - Shared test helpers, embedded test PDFs

**CI coverage enforcement:**
```yaml
- name: Check coverage threshold
  run: |
    coverage=$(go tool cover -func=cover.out | grep total | awk '{print $3}' | tr -d '%')
    if (( $(echo "$coverage < 75" | bc -l) )); then
      echo "Coverage $coverage% is below 75% threshold"
      exit 1
    fi
```

### 9. Documentation

**New files:**
- `docs/architecture.md` - Package diagram, data flow, design decisions
- Updated `CONTRIBUTING.md` - Dev setup, testing guidelines, PR checklist

**New Makefile targets:**
```make
lint-fix:      # Auto-fix linting issues
test-coverage: # Generate HTML coverage report
docs:          # Generate godoc locally
check-all:     # Full pre-commit check
```

---

## Implementation Phases

### Phase 1: Foundation (Low Risk)
1. Enhanced linting configuration
2. Add mock infrastructure (`internal/testing/`)
3. Create `docs/architecture.md`
4. Update `CONTRIBUTING.md`

### Phase 2: Package Reorganization
1. Split `internal/util/` → `fileio/`, `pages/`, `output/`, `errors/`, `progress/`
2. Update all imports across codebase
3. Ensure tests pass after reorganization

### Phase 3: PDF Package Refactoring
1. Split `pdf.go` into focused modules
2. Keep public API unchanged
3. Add tests for gaps discovered

### Phase 4: Command Layer Improvements
1. Create `internal/commands/patterns/` with StdioHandler and BatchProcessor
2. Refactor 7 commands to use new patterns
3. Add missing command tests

### Phase 5: Operational Features
1. Add config file support (`internal/config/`)
2. Add structured logging (`internal/logging/`)
3. Add `--dry-run` flag to modifying commands

### Phase 6: Test Coverage Push
1. Target remaining gaps in `pdf/` and `commands/`
2. Add CI coverage threshold enforcement
3. Final documentation polish

---

## Estimated Impact

**File changes:**
- ~25 new files created
- ~40 existing files modified
- ~15 files split/reorganized
- Net LOC: Roughly neutral (duplication removed, structure added)

**Risk assessment:**
- Phases 1-3: Low risk (additive changes, no behavior changes)
- Phase 4: Medium risk (refactoring working code)
- Phase 5: Low risk (new features, additive)
- Phase 6: Low risk (tests only)

**Backward compatibility:** All changes maintain existing CLI interface and behavior.

---

## Success Criteria

1. Test coverage ≥ 75% overall
2. All linters pass with expanded configuration
3. Architecture documentation complete
4. Config file support working
5. Dry-run mode working for all modifying commands
6. Structured logging available
7. No regressions in existing functionality
