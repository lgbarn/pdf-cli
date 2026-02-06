# Plan 1.1: OCR Checksum Expansion

## Context
This plan implements R1 from Phase 2: Security Hardening. It expands OCR tessdata checksum verification from 1 language (eng) to 21 total languages covering the top 20 most commonly used languages worldwide.

The checksum verification happens during tessdata download in `internal/ocr/ocr.go` (lines 317-333). When a checksum is known, downloads are verified for integrity. When unknown, a warning is printed with the computed SHA256 but the download proceeds.

This plan adds 20 new language checksums to the existing map in `internal/ocr/checksums.go`.

## Dependencies
None. This plan can execute independently in Wave 1.

## Tasks

### Task 1: Download tessdata files and compute checksums
**Files:** None (local verification task)
**Action:** Download
**Description:**
Download each of the 20 tessdata_fast .traineddata files from GitHub and compute their SHA256 checksums locally. This verification task ensures we have correct checksums before hardcoding them.

Languages to download and verify:
1. fra (French)
2. deu (German)
3. spa (Spanish)
4. ita (Italian)
5. por (Portuguese)
6. nld (Dutch)
7. pol (Polish)
8. rus (Russian)
9. jpn (Japanese)
10. chi_sim (Chinese Simplified)
11. chi_tra (Chinese Traditional)
12. kor (Korean)
13. ara (Arabic)
14. hin (Hindi)
15. tur (Turkish)
16. vie (Vietnamese)
17. ukr (Ukrainian)
18. ces (Czech)
19. swe (Swedish)
20. nor (Norwegian)

For each language, run:
```bash
curl -sL "https://github.com/tesseract-ocr/tessdata_fast/raw/main/<LANG>.traineddata" | shasum -a 256
```

Store the computed checksums for Task 2.

**Acceptance Criteria:**
- 20 SHA256 checksums computed (one per language)
- All checksums are 64-character lowercase hexadecimal strings
- Checksums match the format: `[a-f0-9]{64}`

### Task 2: Add checksums to KnownChecksums map
**Files:** `/Users/lgbarn/Personal/pdf-cli/internal/ocr/checksums.go`
**Action:** Modify
**Description:**
Add all 20 language checksums to the `KnownChecksums` map in alphabetical order (keeping "eng" first for backward compatibility, then alphabetical).

Edit the map at lines 9-11 to include entries for all 20 new languages in this format:
```go
var KnownChecksums = map[string]string{
    "eng":     "7d4322bd2a7749724879683fc3912cb542f19906c83bcc1a52132556427170b2",
    "ara":     "<checksum from Task 1>",
    "ces":     "<checksum from Task 1>",
    "chi_sim": "<checksum from Task 1>",
    "chi_tra": "<checksum from Task 1>",
    // ... continue for all 20 languages
}
```

**Acceptance Criteria:**
- KnownChecksums map contains exactly 21 entries (eng + 20 new)
- All entries follow format: `"lang": "64-char-hex-checksum"`
- Entries are in alphabetical order (except "eng" first)
- No syntax errors introduced

### Task 3: Verify implementation with tests
**Files:** `/Users/lgbarn/Personal/pdf-cli/internal/ocr/checksums_test.go`
**Action:** Test
**Description:**
Run existing tests to verify the new checksums are valid. The existing test `TestAllChecksumsValidFormat` (lines 25-38 of checksums_test.go) automatically validates that all checksums are 64-character lowercase hex strings.

Run the full ocr test suite with race detection:
```bash
go test -race ./internal/ocr/...
```

**Acceptance Criteria:**
- `TestAllChecksumsValidFormat` passes (validates all 21 checksums)
- `TestGetChecksum` passes (validates GetChecksum function)
- `TestHasChecksum` passes (validates HasChecksum function)
- All tests in `internal/ocr/...` pass with race detection
- Test coverage for `checksums.go` remains at 100%

## Verification

Run all verification commands:

```bash
# Verify checksum count
grep -c 'eng\|fra\|deu\|spa' /Users/lgbarn/Personal/pdf-cli/internal/ocr/checksums.go
# Expected output: >= 20 (actually will be 21 since eng appears in KnownChecksums)

# Verify all checksums are valid format
go test -race -run TestAllChecksumsValidFormat ./internal/ocr/

# Run full OCR test suite
go test -race ./internal/ocr/...

# Verify no syntax errors
go build ./internal/ocr/
```

## Success Criteria
- KnownChecksums map contains 21 entries
- All tests in `./internal/ocr/...` pass with `-race` flag
- `grep -c 'eng\|fra\|deu\|spa' internal/ocr/checksums.go` returns >= 20
- Test coverage >= 75% for ocr package (already at 82.9%)
- No regression in existing functionality
