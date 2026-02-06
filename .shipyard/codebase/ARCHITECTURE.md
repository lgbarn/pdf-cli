# Architecture

## Overview

pdf-cli follows a layered, modular monolith architecture with strict unidirectional dependencies. The system is built around the Command pattern (via Cobra), with commands orchestrating operations through domain-specific packages (pdf, ocr) that wrap external libraries. Core utilities (fileio, pages, output, pdferrors) provide reusable abstractions without creating circular dependencies.

## Findings

### Architectural Pattern

- **Pattern**: Layered modular monolith with Command pattern for CLI operations
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/cmd/pdf/main.go` (lines 1-36) -- single entry point with command registration
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/commands/info.go` (line 18) -- `cli.AddCommand(infoCmd)` registration pattern used by all 14 commands
  - Commands instantiated via `init()` functions and registered with global root command
  - No microservices, no distributed components -- single binary deployment

- **No circular dependencies**: Dependency graph enforces unidirectional flow
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/docs/architecture.md` (lines 29-54) -- documented dependency hierarchy
  - Leaf packages (fileio, pages, output, pdferrors, progress, retry) have zero internal dependencies
  - Domain packages (pdf, ocr) depend only on leaf packages and external libraries
  - Commands layer depends on domain and leaf packages
  - CLI layer depends on commands via side-effect registration, not direct import

### Layer Boundaries and Data Flow

**Layer 1: Entry Point**
- **Package**: `cmd/pdf`
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/cmd/pdf/main.go` (lines 21-35)
  - Responsibilities: Signal handling, context creation, cleanup orchestration, version injection
  - Creates context with `signal.NotifyContext` for SIGINT/SIGTERM handling
  - Registers deferred cleanup with `defer cleanup.Run()`
  - Sets version info via build flags, delegates execution to CLI layer

**Layer 2: CLI Framework**
- **Package**: `internal/cli`
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/cli/cli.go` (lines 50-88)
  - Responsibilities: Cobra root command setup, global flags, password handling, completion
  - Exposes `AddCommand()` for command registration (called during init)
  - Provides flag helpers: `Verbose()`, `Force()`, `Progress()`, `IsDryRun()`
  - Manages 4-tier password priority: password-file > env var > CLI flag (deprecated) > interactive prompt

**Layer 3: Commands (Orchestration)**
- **Package**: `internal/commands`
  - Evidence: 28 Go files in `/Users/lgbarn/Personal/pdf-cli/internal/commands/` (14 commands + helpers + tests)
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/commands/compress.go` (lines 45-80)
  - Responsibilities: Input validation, batch processing, stdin/stdout handling, progress orchestration, dry-run mode
  - Each command file defines a `cobra.Command` with `RunE` handler
  - Commands use `patterns.StdioHandler` for uniform stdin/stdout pipeline support
  - Batch processing via `processBatch()` helper collects errors with `errors.Join`

**Layer 4: Domain Logic**
- **Package**: `internal/pdf` (wraps pdfcpu + ledongthuc/pdf)
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/pdf/metadata.go`, `transform.go`, `text.go`, `encryption.go`, `watermark.go`, `validation.go`
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/pdf/text.go` (lines 34-42) -- dual library strategy
  - Responsibilities: PDF manipulation, text extraction with fallback, metadata operations
  - Uses ledongthuc/pdf as primary for text extraction, falls back to pdfcpu content parsing
  - All operations accept password parameter for encrypted PDFs
  - Progress reporting integrated for long operations (merge, split, text extraction)

- **Package**: `internal/ocr` (wraps gogosseract + native Tesseract)
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr.go` (lines 120-137)
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/backend.go` (lines 8-13) -- Backend interface
  - Responsibilities: OCR text extraction with pluggable backends, tessdata management, parallel processing
  - Implements Backend interface with Native and WASM implementations
  - Auto-selects native Tesseract if available, else WASM fallback
  - Tessdata download with retry logic, checksum verification, progress bars
  - Parallel processing triggered when image count exceeds configurable threshold

**Layer 5: Utilities (Leaf Packages)**
- **Package**: `internal/fileio`
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/fileio/stdio.go` (lines 12-84)
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/fileio/files.go`
  - Responsibilities: Stdin/stdout handling, file validation, path sanitization, atomic writes
  - `ResolveInputPath()` handles stdin indicator ("-") by creating temp file from stdin
  - `SanitizePath()` prevents directory traversal attacks
  - `AtomicWrite()` uses temp file + rename for crash safety

- **Package**: `internal/pages`
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/pages/parser.go` (lines 10-116)
  - Responsibilities: Parse page ranges ("1-5,7,10-12"), expand to page numbers, validate bounds
  - Deduplicates pages using seen map
  - Supports "end" keyword for last page (in reorder operations)

- **Package**: `internal/output`
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/output/formatter.go` (lines 35-86)
  - Responsibilities: Format output as JSON, CSV, TSV, or human-readable
  - `IsStructured()` determines if machine-readable format is used
  - JSON uses indented encoding for readability

- **Package**: `internal/pdferrors`
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/pdferrors/errors.go` (lines 9-97)
  - Responsibilities: Context-aware error wrapping with operation, file, cause, and hints
  - `WrapError()` detects common error patterns (encryption, file not found, corruption) and adds helpful hints
  - Implements `Unwrap()` for error chain inspection

- **Package**: `internal/progress`
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/progress/progress.go`
  - Responsibilities: Progress bar creation and management using schollz/progressbar
  - Consistent styling across all operations

- **Package**: `internal/cleanup`
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/cleanup/cleanup.go` (lines 10-68)
  - Responsibilities: Thread-safe temp file registry for signal-based cleanup
  - `Register()` returns unregister function for normal cleanup path
  - `Run()` is idempotent, removes files in LIFO order on signal or defer
  - Prevents temp file leaks on SIGINT/SIGTERM

- **Package**: `internal/config`
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/config/config.go` (lines 14-220)
  - Responsibilities: YAML config loading, environment variable overrides, singleton pattern
  - Config path: `$XDG_CONFIG_HOME/pdf-cli/config.yaml` or `~/.config/pdf-cli/config.yaml`
  - `PerformanceConfig` adapts defaults based on `runtime.NumCPU()`
  - Thread-safe singleton using `sync.RWMutex` double-checked locking

- **Package**: `internal/retry`
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/retry/` (referenced in ocr.go)
  - Responsibilities: Exponential backoff retry with permanent error detection
  - Used for network resilience in tessdata downloads

- **Package**: `internal/logging`
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/logging/logger.go`
  - Responsibilities: Structured logging with slog (Go 1.21+), text/JSON formats, multiple levels
  - Thread-safe singleton initialization

### Key Abstractions and Interfaces

**Backend Interface (OCR pluggability)**
- **Definition**: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/backend.go` (lines 8-13)
  ```go
  type Backend interface {
      Name() string
      Available() bool
      ProcessImage(ctx context.Context, imagePath, lang string) (string, error)
      Close() error
  }
  ```
- **Implementations**: NativeBackend (system Tesseract), WASMBackend (gogosseract)
- **Selection logic**: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr.go` (lines 120-137)
  - BackendAuto tries native first, falls back to WASM
  - BackendNative returns error if unavailable
  - BackendWASM always succeeds (embedded binary)

**StdioHandler Pattern (stdin/stdout pipelines)**
- **Definition**: `/Users/lgbarn/Personal/pdf-cli/internal/commands/patterns/stdio.go` (lines 11-92)
- **Purpose**: Uniform handling of stdin input ("-") and stdout output ("--stdout") across commands
- **API**:
  - `Setup()` -- resolves input (may read stdin to temp), resolves output (may create temp for stdout)
  - `Finalize()` -- writes temp file to stdout if needed
  - `Cleanup()` -- removes temp files (safe to call multiple times)
- **Usage**: `/Users/lgbarn/Personal/pdf-cli/internal/commands/compress.go` (lines 103-137)
  - Used by compress, extract, rotate, reorder, encrypt, decrypt, pdfa commands

**Config Singleton with Adaptive Defaults**
- **Pattern**: Thread-safe lazy initialization with double-checked locking
- **Evidence**: `/Users/lgbarn/Personal/pdf-cli/internal/config/config.go` (lines 192-213)
- **Adaptive performance config**: `/Users/lgbarn/Personal/pdf-cli/internal/config/config.go` (lines 53-69)
  - `MaxWorkers` = min(NumCPU, 8) to avoid resource exhaustion
  - `OCRParallelThreshold` = max(NumCPU/2, 5) for adaptive parallelism
  - Overridable via `PDF_CLI_PERF_*` environment variables

### State Management

- **Global state**: Minimal, limited to config singleton and cleanup registry
  - Config: `/Users/lgbarn/Personal/pdf-cli/internal/config/config.go` (lines 188-213) -- lazy-loaded, immutable after load
  - Cleanup: `/Users/lgbarn/Personal/pdf-cli/internal/cleanup/cleanup.go` (lines 10-14) -- mutex-protected path registry
  - Logging: `/Users/lgbarn/Personal/pdf-cli/internal/logging/logger.go` -- singleton logger

- **Request-scoped state**: Context propagation for cancellation
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/cmd/pdf/main.go` (line 26) -- `signal.NotifyContext` creates cancellable context
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/pdf/text.go` (line 28) -- all text extraction accepts `context.Context`
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr.go` (line 326) -- OCR accepts context for cancellation
  - Enables graceful shutdown on SIGINT/SIGTERM

- **No shared mutable state**: Commands are stateless, all state passed via parameters or context

### Dependency Injection and Service Location

- **No DI framework**: Dependencies are wired manually via function parameters
  - Evidence: PDF operations accept password as parameter, not via config
  - Evidence: OCR engine accepts options struct, not global config

- **Service Location**: Minimal use
  - Config singleton via `config.Get()` -- used sparingly, primarily for performance tuning
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/pdf/text.go` (lines 69-73) -- reads parallel threshold from config

- **Explicit dependencies**: Most dependencies passed as parameters
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/commands/compress.go` (line 124) -- `pdf.Compress(input, output, password)`
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr.go` (lines 70-76) -- `NewEngine` returns configured instance

### Data Flow (Representative Example: Compress Command)

1. **User invokes**: `pdf compress input.pdf -o output.pdf --password secret`
2. **Entry point**: `main()` creates context with signal handling
3. **CLI layer**: Cobra parses flags, invokes `runCompress()`
4. **Command layer**: `/Users/lgbarn/Personal/pdf-cli/internal/commands/compress.go`
   - Line 46: Sanitize input args (prevent directory traversal)
   - Line 51: Retrieve password via 4-tier priority
   - Line 64: Check dry-run mode
   - Line 69: Check stdin/stdout mode
   - Line 124: Call `pdf.Compress(input, output, password)`
5. **Domain layer**: `/Users/lgbarn/Personal/pdf-cli/internal/pdf/transform.go`
   - Line 152: Call `api.OptimizeFile(input, output, NewConfig(password))`
6. **External library**: pdfcpu performs optimization
7. **Return**: Success bubbles up, command layer prints success message
8. **Cleanup**: Deferred cleanup removes any temp files created

### Parallelism and Concurrency

**Text Extraction Parallelism**
- **Evidence**: `/Users/lgbarn/Personal/pdf-cli/internal/pdf/text.go` (lines 125-173)
- **Strategy**: Parallel extraction when page count exceeds threshold (default 5)
- **Mechanism**: Goroutine per page, results collected in map, reassembled in page order
- **Limitation**: ledongthuc/pdf Reader is not documented as thread-safe, but works in practice

**OCR Parallelism**
- **Evidence**: `/Users/lgbarn/Personal/pdf-cli/internal/ocr/ocr.go` (lines 460-518)
- **Strategy**: Parallel processing for native backend, sequential for WASM (not thread-safe)
- **Mechanism**: Semaphore-limited goroutines, buffered results channel, ordered reassembly
- **Worker limit**: Configurable via `MaxWorkers` (default: min(NumCPU, 8))
- **Code**: Lines 470-476 create semaphore, lines 478-492 launch workers

**Merge Parallelism**
- **Evidence**: `/Users/lgbarn/Personal/pdf-cli/internal/pdf/transform.go` (lines 22-74)
- **Strategy**: Sequential merge with progress bar (pdfcpu does not support parallel merge)
- **Reason**: PDF structure requires sequential assembly, no parallelism opportunity

**File Validation Parallelism**
- **[Inferred]**: Batch info command likely validates files sequentially
- **Evidence**: `/Users/lgbarn/Personal/pdf-cli/internal/commands/info.go` (lines 151-232) -- loop over files without goroutines

### Module Boundaries

**External Library Isolation**
- **pdfcpu**: Isolated in `internal/pdf` package
  - Evidence: All pdfcpu imports confined to `internal/pdf/*.go`
  - Wrapped API: `NewConfig()`, `GetInfo()`, `Merge()`, `Split()`, `Compress()`, etc.
  - Commands never import pdfcpu directly

- **ledongthuc/pdf**: Isolated in `internal/pdf/text.go`
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/pdf/text.go` (line 11)
  - Used only as fallback for text extraction

- **gogosseract**: Isolated in `internal/ocr/wasm.go`
  - Evidence: Referenced in OCR package, not imported by commands
  - WASM backend implementation detail

- **Tesseract (native)**: Isolated in `internal/ocr/native.go`
  - Evidence: Accessed via exec.Command, not direct Go binding
  - Detection logic in `internal/ocr/detect.go`

**Public Interfaces**
- **PDF package**: Exported functions in `/Users/lgbarn/Personal/pdf-cli/internal/pdf/*.go`
  - `GetInfo()`, `Merge()`, `Split()`, `Compress()`, `ExtractText()`, `Encrypt()`, `Decrypt()`, etc.
  - Types: `Info`, `Metadata`

- **OCR package**: Exported types in `/Users/lgbarn/Personal/pdf-cli/internal/ocr/backend.go`, `ocr.go`
  - `Engine`, `Backend` interface, `EngineOptions`
  - `NewEngine()`, `ExtractTextFromPDF()`

- **Commands package**: No exported API -- commands register themselves via `init()` side effects
  - Evidence: `/Users/lgbarn/Personal/pdf-cli/internal/commands/info.go` (lines 17-22)

## Summary Table

| Aspect | Detail | Confidence |
|--------|--------|------------|
| Architecture | Layered modular monolith, Command pattern | Observed |
| Entry Point | `/Users/lgbarn/Personal/pdf-cli/cmd/pdf/main.go` | Observed |
| Commands | 14 commands, 28 files in `internal/commands/` | Observed |
| Domain Packages | `internal/pdf` (7 files), `internal/ocr` (10+ files) | Observed |
| Utility Packages | 8 leaf packages (fileio, pages, output, pdferrors, progress, cleanup, config, logging, retry) | Observed |
| Dependency Direction | Unidirectional: commands → domain → utilities | Observed |
| External Libraries | pdfcpu (primary), ledongthuc/pdf (text fallback), gogosseract (WASM OCR) | Observed |
| Backend Pattern | OCR Backend interface with Native/WASM implementations | Observed |
| Stdin/Stdout | StdioHandler pattern in `commands/patterns/stdio.go` | Observed |
| Parallelism | Adaptive based on runtime.NumCPU(), configurable thresholds | Observed |
| State Management | Minimal global state (config singleton, cleanup registry), context for cancellation | Observed |
| Error Handling | PDFError wrapper with operation/file context and user-friendly hints | Observed |
| Config Management | YAML + env vars, thread-safe singleton, adaptive performance defaults | Observed |
| Signal Handling | Context cancellation + cleanup registry for graceful shutdown | Observed |

## Open Questions

- **Thread safety of ledongthuc/pdf.Reader**: Parallel text extraction uses goroutines per page with shared Reader instance. Library documentation does not explicitly state thread safety, but code works in practice. Consider adding synchronization or confirming with upstream maintainers.
- **Error propagation in parallel operations**: Currently uses `errors.Join` to collect all errors. For large batches (e.g., 1000 pages), this could result in verbose error output. Consider limiting to first N errors or providing summary.
- **Performance tuning defaults**: Current thresholds (5 pages for parallelism, 8 max workers) are hardcoded or CPU-based. No benchmarking data provided to validate these choices. Consider profiling with representative workloads.
- **Config reload**: Config singleton is loaded once and cached. No mechanism to reload config without restarting process. May be acceptable for CLI tool, but could be limiting for long-running server use case (not currently a requirement).
