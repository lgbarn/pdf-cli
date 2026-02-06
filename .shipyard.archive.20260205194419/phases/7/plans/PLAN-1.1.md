---
phase: docs-and-tests
plan: 1.1
wave: 1
dependencies: []
must_haves:
  - R10: Test files over 500 lines should be split into focused files
files_touched:
  - internal/pdf/pdf_test.go
  - internal/pdf/text_test.go
  - internal/pdf/transform_test.go
  - internal/pdf/metadata_test.go
  - internal/pdf/validate_test.go
  - internal/pdf/images_test.go
  - internal/pdf/content_parsing_test.go
  - internal/commands/commands_integration_test.go
  - internal/commands/integration_transform_test.go
  - internal/commands/integration_content_test.go
  - internal/commands/integration_batch_test.go
  - internal/commands/additional_coverage_test.go
  - internal/commands/coverage_images_test.go
  - internal/commands/coverage_batch_test.go
tdd: false
---

# Plan 1.1: Split Large Test Files

## Context

Three test files exceed the 500-line threshold. They need to be split into focused files organized by topic. This is a purely mechanical refactoring — no test logic changes, just moving functions between files.

Large files:
1. `internal/pdf/pdf_test.go` — 2,344 lines (96 test functions)
2. `internal/commands/commands_integration_test.go` — 882 lines (34 test functions)
3. `internal/commands/additional_coverage_test.go` — 620 lines (30 test functions)

## Tasks

<task id="1" files="internal/pdf/pdf_test.go,internal/pdf/text_test.go,internal/pdf/transform_test.go,internal/pdf/metadata_test.go,internal/pdf/validate_test.go,internal/pdf/images_test.go,internal/pdf/content_parsing_test.go" tdd="false">
  <action>
    Split `internal/pdf/pdf_test.go` (2,344 lines) into focused test files.

    **IMPORTANT:** Read the ENTIRE file first. Identify all test functions and shared helpers (testdataDir, helper functions, etc.).

    Shared helpers (like testdataDir, setup functions) should remain in `pdf_test.go` or a `helpers_test.go` file so all split files can use them.

    Split by topic:

    **Keep in pdf_test.go** (~250 lines):
    - Shared test helpers (testdataDir, any setup functions)
    - TestPagesToStrings, TestNewConfig (utility tests)
    - TestGetInfo*, TestPageCount* (core PDF info tests)
    - TestValidateValid, TestValidatePDFA*, TestConvertToPDFA*

    **New: text_test.go** (~400 lines):
    - TestExtractText*, TestExtractTextWithPages, TestExtractTextWithProgress*
    - TestExtractTextFallbackPath, TestExtractTextPrimaryWithSpecificPages
    - TestExtractPageText*, TestExtractPagesSequential*, TestExtractPagesParallel*

    **New: transform_test.go** (~400 lines):
    - TestMerge*, TestMergeWithProgress*, TestSplit*, TestSplitByPageCount*
    - TestSplitWithProgress*, TestRotate*, TestCompress*, TestEncrypt*, TestDecrypt*
    - TestExtractPages*, TestReorder* (if any)

    **New: metadata_test.go** (~200 lines):
    - TestGetMetadata*, TestSetMetadata*, TestInfoStruct, TestMetadataStruct
    - TestPDFAValidationResultStruct

    **New: images_test.go** (~150 lines):
    - TestCreatePDFFromImages*, TestExtractImages*

    **New: content_parsing_test.go** (~200 lines):
    - TestParseTextFromPDFContent*, TestExtractParenString*
    - TestAddWatermark*, TestAddImageWatermark*

    All files must be `package pdf` (same package). Run `go test ./internal/pdf/...` after each split to verify nothing breaks.
  </action>
  <verify>
    ```bash
    cd /Users/lgbarn/Personal/pdf-cli/.worktrees/phase-2-concurrency
    go test -race ./internal/pdf/... -count=1
    wc -l internal/pdf/*_test.go | sort -rn
    # Verify no file exceeds 500 lines
    ```
  </verify>
  <done>
    - pdf_test.go split into 6+ focused test files
    - No test file in internal/pdf/ exceeds 500 lines
    - All tests pass unchanged
    - Each file has a clear topic focus
  </done>
</task>

<task id="2" files="internal/commands/commands_integration_test.go,internal/commands/integration_transform_test.go,internal/commands/integration_content_test.go,internal/commands/integration_batch_test.go" tdd="false">
  <action>
    Split `internal/commands/commands_integration_test.go` (882 lines) into focused files.

    **Read the file first.** Identify shared helpers.

    Split by command category:

    **Keep in commands_integration_test.go** (~300 lines):
    - Shared test helpers (testdataDir, executeCommand, etc.)
    - TestCompressCommand*, TestRotateCommand* (transform commands)

    **New: integration_content_test.go** (~300 lines):
    - TestTextCommand*, TestInfoCommand, TestExtractCommand*
    - TestMetaCommand (basic), TestPdfaValidateCommand*, TestPdfaConvertCommand*

    **New: integration_batch_test.go** (~300 lines):
    - TestMergeCommand*, TestSplitCommand*
    - TestWatermarkCommand*, TestReorderCommand*
    - TestEncryptCommand*, TestDecryptCommand*
    - TestInfoCommand_MultipleFiles, TestMetaCommand_MultipleFiles
    - TestCommands_NonExistentFile

    All files must be `package commands` (same package).
  </action>
  <verify>
    ```bash
    cd /Users/lgbarn/Personal/pdf-cli/.worktrees/phase-2-concurrency
    go test -race ./internal/commands/... -count=1
    wc -l internal/commands/*_test.go | sort -rn
    ```
  </verify>
  <done>
    - commands_integration_test.go split into 3 focused files
    - No integration test file exceeds 500 lines
    - All tests pass unchanged
  </done>
</task>

<task id="3" files="internal/commands/additional_coverage_test.go,internal/commands/coverage_images_test.go,internal/commands/coverage_batch_test.go" tdd="false">
  <action>
    Split `internal/commands/additional_coverage_test.go` (620 lines) into focused files.

    **Read the file first.**

    Split by topic:

    **Keep in additional_coverage_test.go** (~250 lines):
    - TestTextCommand_WithOCR*, TestMerge_NonexistentFiles, TestSplit_InvalidPagesPerFile
    - TestPdfaValidate_JSONFormat, TestCompressCommand_FileSizeIncrease
    - TestReorderCommand_EndKeyword, TestExtract_NoPages

    **New: coverage_images_test.go** (~200 lines):
    - TestImagesCommand*, TestCombineImagesCommand*

    **New: coverage_batch_test.go** (~200 lines):
    - TestInfoCommand_Batch*, TestMetaCommand_Batch*, TestMetaCommand_SetMultipleOnBatch
    - TestCompressCommand_BatchWithOutputFlag, TestEncryptCommand_BatchWithOutputFlag
    - TestRotateCommand_BatchWithOutputFlag, TestWatermarkCommand_BatchWithOutputFlag

    All files must be `package commands` (same package).
  </action>
  <verify>
    ```bash
    cd /Users/lgbarn/Personal/pdf-cli/.worktrees/phase-2-concurrency
    go test -race ./internal/commands/... -count=1
    wc -l internal/commands/*_test.go | sort -rn
    ```
  </verify>
  <done>
    - additional_coverage_test.go split into 3 focused files
    - No coverage test file exceeds 500 lines
    - All tests pass unchanged
  </done>
</task>

## Verification

```bash
cd /Users/lgbarn/Personal/pdf-cli/.worktrees/phase-2-concurrency

# All tests pass
go test -race ./... -short -count=1

# No test file exceeds 500 lines
find . -name "*_test.go" -exec wc -l {} + | sort -rn | head -5
# First line should be < 500

# Coverage unchanged
go test -coverprofile=cover.out ./... -short
go tool cover -func=cover.out | tail -1
```

## Success Criteria

- No test file exceeds 500 lines
- Each split test file has a clear focus indicated by its filename
- go test ./... passes (test behavior unchanged)
- Coverage remains >= 81%
