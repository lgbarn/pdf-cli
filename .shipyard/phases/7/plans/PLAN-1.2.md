---
phase: docs-and-tests
plan: 1.2
wave: 1
dependencies: []
must_haves:
  - R13: Documentation (README, architecture.md) aligned with current code
files_touched:
  - README.md
  - docs/architecture.md
tdd: false
---

# Plan 1.2: Documentation Alignment

## Context

After 6 phases of changes, README.md and docs/architecture.md are outdated. Key changes that need documenting:
- Phase 3: Password input changed from --password flag to --password-file / env var / interactive prompt
- Phase 4: New internal/cleanup package for signal-based temp file cleanup
- Phase 5: New PerformanceConfig, configurable parallelism, environment overrides
- Phase 6: New internal/retry package for exponential backoff

## Tasks

<task id="1" files="README.md" tdd="false">
  <action>
    Update README.md to reflect all changes from Phases 1-6.

    **Read README.md first.**

    Updates needed:

    1. **Go version**: Verify go.mod version matches README. Update if needed.

    2. **Password handling** (Phase 3 breaking change):
       - Remove any examples showing `--password` as a direct flag
       - Document the new 4-tier password input:
         1. `--password-file <path>` — read from file
         2. `PDF_CLI_PASSWORD` environment variable
         3. `--password` flag (deprecated, shows warning)
         4. Interactive terminal prompt (if stdin is a terminal)
       - Update encrypt/decrypt command examples

    3. **New packages** (if project structure is documented):
       - `internal/cleanup` — signal-based temp file cleanup
       - `internal/retry` — exponential backoff retry logic

    4. **Configuration** (Phase 5):
       - Document performance environment variables:
         - `PDF_CLI_PERF_OCR_THRESHOLD`
         - `PDF_CLI_PERF_TEXT_THRESHOLD`
         - `PDF_CLI_PERF_MAX_WORKERS`

    5. **OCR improvements**:
       - Mention retry logic for tessdata downloads (auto-retries on network failures)
       - Mention SHA256 checksum verification

    Do NOT add excessive documentation. Keep changes minimal and aligned with existing README style.
  </action>
  <verify>
    Review README.md manually for accuracy.
  </verify>
  <done>
    - README reflects current password input methods
    - Go version is correct
    - New features documented appropriately
    - Examples are accurate
  </done>
</task>

<task id="2" files="docs/architecture.md" tdd="false">
  <action>
    Update docs/architecture.md to reflect new packages and changes.

    **Read docs/architecture.md first.**

    Updates needed:

    1. **New packages**:
       - `internal/cleanup` — Thread-safe cleanup registry for temp files. Register/Run API. Integrated with signal handler in main.go for SIGINT/SIGTERM cleanup.
       - `internal/retry` — Generic retry helper with exponential backoff. Used by tessdata downloads. PermanentError type for non-retryable errors.

    2. **Updated packages**:
       - `internal/config` — New PerformanceConfig struct with adaptive defaults based on runtime.NumCPU()
       - `internal/ocr` — Now uses retry for downloads, error collection with errors.Join, configurable parallelism
       - `internal/fileio` — Path sanitization (SanitizePath), CopyFile close error propagation, AtomicWrite with cleanup registration
       - `internal/cli` — Secure password reading (ReadPassword) with file/env/flag/prompt tiers

    3. **Architecture changes**:
       - Signal handling flow: main.go → signal.NotifyContext → cleanup.Run on exit
       - Error propagation: parallel processing now collects all errors via errors.Join

    Keep the existing style and structure. Add new packages where the package list is. Update descriptions of modified packages.
  </action>
  <verify>
    Review docs/architecture.md manually for accuracy.
  </verify>
  <done>
    - architecture.md lists all new packages
    - Updated package descriptions reflect current functionality
    - Signal handling and error propagation documented
  </done>
</task>

## Verification

```bash
cd /Users/lgbarn/Personal/pdf-cli/.worktrees/phase-2-concurrency

# Verify docs exist and are non-empty
wc -l README.md docs/architecture.md

# Verify no code changes (docs only)
go test -race ./... -short -count=1
```

## Success Criteria

- README reflects current Go version, new password input method, and CLI changes
- architecture.md reflects new packages (cleanup, retry) and updated function signatures
- No functional code changes
- All tests still pass
