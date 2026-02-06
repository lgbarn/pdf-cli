---
phase: security-hardening
plan: 03
wave: 2
dependencies: [02]
must_haves:
  - Embedded SHA256 checksums for common tessdata languages
  - Verify downloaded files before renaming into place
  - Clear error on checksum mismatch
  - Warning for unknown languages without checksums
files_touched:
  - internal/ocr/checksums.go (new)
  - internal/ocr/ocr.go
  - internal/ocr/ocr_test.go
tdd: true
---

# Plan 2.1: Tessdata Checksum Verification (R2)

## Goal
Verify integrity of downloaded tessdata files using SHA256 checksums to prevent supply chain attacks and corrupted downloads. Since tesseract-ocr/tessdata_fast doesn't publish official checksums, we'll compute and embed them for the most common language files.

## Context
Currently, downloadTessdata in internal/ocr/ocr.go (lines 169-208) downloads .traineddata files from GitHub without any integrity verification. This creates two risks:
1. Supply chain attack: compromised GitHub could serve malicious files
2. Corrupted downloads: network issues could produce incomplete/corrupted files

We cannot rely on external checksums (none published), so we must compute SHA256 hashes for common languages and embed them in the codebase. For uncommon languages, we'll warn but allow download.

This plan depends on Plan 1.2 (path sanitization) because we need SanitizePath in the download flow.

## Tasks

<task id="1" files="internal/ocr/checksums.go" tdd="true">
  <action>
Create internal/ocr/checksums.go with embedded SHA256 checksums for common tessdata_fast languages.

File structure:
```go
package ocr

// KnownChecksums maps language codes to SHA256 checksums for tessdata_fast files.
// These checksums are for the tessdata_fast repository from tesseract-ocr.
// Generated from: https://github.com/tesseract-ocr/tessdata_fast/tree/main
//
// To add a new language checksum:
// 1. Download: curl -L https://github.com/tesseract-ocr/tessdata_fast/raw/main/LANG.traineddata -o /tmp/LANG.traineddata
// 2. Compute: sha256sum /tmp/LANG.traineddata
// 3. Add entry: "LANG": "SHA256_HEX_STRING"
var KnownChecksums = map[string]string{
  // Common Western languages
  "eng": "7d4322bd2a7749724879683fc3912cb542f19906c83bcc1a52132556427170b2",
  "fra": "907915a1a0e6b58f99620b3f03186c98815b6c96e4b975b7b45d0d0c0d8f5c3d",
  "spa": "f74d2b5e7e6c6c3b8c99c6c5b4c0e7c0f7c8c6c8c6c8c6c8c6c8c6c8c6c8c6c8",
  "deu": "2a67a3c9e3d5c2c7e8c9c6c5b4c0e7c0f7c8c6c8c6c8c6c8c6c8c6c8c6c8c6c8",
  "ita": "3b78b4d0f4e6d3d8f9d0d7d6c5d1f8d1g8d7d7d9d7d9d7d9d7d9d7d9d7d9d7d9",
  "por": "4c89c5e1g5f7e4e9g1e1e8e7d6e2g9e2h9e8e8e0e8e0e8e0e8e0e8e0e8e0e8e0",

  // Common Asian languages
  "jpn": "8e12f23a4b56c78d90ef12a34b56c78d90ef12a34b56c78d90ef12a34b56c78d",
  "chi_sim": "9f23g34b5c67d89e01fg23a45b67c89d01ef23a45b67c89d01ef23a45b67c89d",
  "chi_tra": "0g34h45c6d78e90f12gh34b56c78d90e12fg34b56c78d90e12fg34b56c78d90e",
  "kor": "1h45i56d7e89f01g23hi45c67d89e01f23gh45c67d89e01f23gh45c67d89e01f",

  // Other common languages
  "rus": "2i56j67e8f90g12h34ij56d78e90f12g34hi56d78e90f12g34hi56d78e90f12g",
  "ara": "3j67k78f9g01h23i45jk67e89f01g23h45ij67e89f01g23h45ij67e89f01g23h",
  "hin": "4k78l89g0h12i34j56kl78f90g12h34i56jk78f90g12h34i56jk78f90g12h34i",
  "osd": "5l89m90h1i23j45k67lm89g01h23i45j67kl89g01h23i45j67kl89g01h23i45j",
}

// GetChecksum returns the known SHA256 checksum for a language, if available.
// Returns empty string if no checksum is known for the language.
func GetChecksum(lang string) string {
  return KnownChecksums[lang]
}

// HasChecksum returns true if a checksum is known for the language.
func HasChecksum(lang string) bool {
  _, ok := KnownChecksums[lang]
  return ok
}
```

IMPORTANT: The checksums above are PLACEHOLDERS. Before implementing this task, the builder MUST:
1. Download actual .traineddata files from https://github.com/tesseract-ocr/tessdata_fast/raw/main/
2. Compute real SHA256 checksums using: `sha256sum LANG.traineddata`
3. Replace placeholder values with actual checksums

Create test file internal/ocr/checksums_test.go:
```go
package ocr

import "testing"

func TestGetChecksum(t *testing.T) {
  // Test known language
  if checksum := GetChecksum("eng"); checksum == "" {
    t.Error("Expected checksum for 'eng', got empty string")
  }

  // Test unknown language
  if checksum := GetChecksum("xyz"); checksum != "" {
    t.Errorf("Expected empty checksum for unknown language, got: %s", checksum)
  }
}

func TestHasChecksum(t *testing.T) {
  if !HasChecksum("eng") {
    t.Error("Expected HasChecksum('eng') to be true")
  }

  if HasChecksum("xyz") {
    t.Error("Expected HasChecksum('xyz') to be false")
  }
}

func TestAllChecksumsValid(t *testing.T) {
  for lang, checksum := range KnownChecksums {
    // SHA256 checksums are 64 hex characters
    if len(checksum) != 64 {
      t.Errorf("Invalid checksum length for %s: got %d, want 64", lang, len(checksum))
    }

    // Verify all characters are valid hex
    for _, c := range checksum {
      if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
        t.Errorf("Invalid hex character in checksum for %s: %c", lang, c)
        break
      }
    }
  }
}
```
  </action>
  <verify>
Tests should pass:
```bash
cd /Users/lgbarn/Personal/pdf-cli/.worktrees/phase-2-concurrency
go test -v ./internal/ocr -run TestChecksum
go test -v ./internal/ocr -run TestAllChecksumsValid
```

Verify at least 10 languages have checksums:
```bash
cd /Users/lgbarn/Personal/pdf-cli/.worktrees/phase-2-concurrency
grep -c '": "' internal/ocr/checksums.go
```
  </verify>
  <done>
- internal/ocr/checksums.go created with real SHA256 checksums
- At least 10 common languages covered (eng, fra, spa, deu, jpn, chi_sim, etc.)
- GetChecksum and HasChecksum helper functions implemented
- All tests pass
- All checksums are valid 64-character hex strings
  </done>
</task>

<task id="2" files="internal/ocr/ocr.go" tdd="false">
  <action>
Update downloadTessdata function to verify checksums after download.

Modify internal/ocr/ocr.go downloadTessdata function (lines 169-208):

1. Add crypto/sha256 and encoding/hex to imports (after line 13):
```go
import (
  "context"
  "crypto/sha256"
  "encoding/hex"
  "fmt"
  ...
)
```

2. Replace the function starting at line 169 with:
```go
func downloadTessdata(ctx context.Context, dataDir, lang string) error {
  url := fmt.Sprintf("%s/%s.traineddata", TessdataURL, lang)
  dataFile := filepath.Join(dataDir, lang+".traineddata")

  // Validate path (uses Plan 1.2 path sanitization)
  dataFile, err := fileio.SanitizePath(dataFile)
  if err != nil {
    return fmt.Errorf("invalid tessdata path for language %s: %w", lang, err)
  }

  fmt.Fprintf(os.Stderr, "Downloading tessdata for '%s'...\n", lang)

  ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
  defer cancel()

  req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
  if err != nil {
    return err
  }

  resp, err := http.DefaultClient.Do(req)
  if err != nil {
    return err
  }
  defer resp.Body.Close()

  if resp.StatusCode != http.StatusOK {
    return fmt.Errorf("failed to download: HTTP %d", resp.StatusCode)
  }

  tmpFile, err := os.CreateTemp(dataDir, "tessdata-*.tmp")
  if err != nil {
    return err
  }
  tmpPath := tmpFile.Name()
  defer os.Remove(tmpPath)

  // Download with progress bar AND compute SHA256 simultaneously
  hasher := sha256.New()
  bar := progress.NewBytesProgressBar(fmt.Sprintf("Downloading %s.traineddata", lang), resp.ContentLength)
  multiWriter := io.MultiWriter(tmpFile, bar, hasher)

  if _, err := io.Copy(multiWriter, resp.Body); err != nil {
    _ = tmpFile.Close()
    return err
  }
  _ = tmpFile.Close()
  progress.FinishProgressBar(bar)

  // Verify checksum if known
  computedHash := hex.EncodeToString(hasher.Sum(nil))
  if expectedHash := GetChecksum(lang); expectedHash != "" {
    if computedHash != expectedHash {
      return fmt.Errorf(
        "checksum verification failed for %s.traineddata\nExpected: %s\nGot:      %s\n"+
        "This may indicate a corrupted download or supply chain attack. "+
        "Please report this issue if it persists.",
        lang, expectedHash, computedHash,
      )
    }
    fmt.Fprintf(os.Stderr, "Checksum verified for %s.traineddata\n", lang)
  } else {
    // Warn for unknown checksums but allow download
    fmt.Fprintf(os.Stderr,
      "WARNING: No checksum available for language '%s'. "+
      "File integrity cannot be verified. Computed SHA256: %s\n",
      lang, computedHash,
    )
  }

  return os.Rename(tmpPath, dataFile)
}
```

Key changes:
- Add path sanitization (depends on Plan 1.2)
- Create SHA256 hasher and add to MultiWriter
- Verify checksum before renaming file into place
- Fail with clear error on mismatch
- Warn but allow for unknown languages
- Display computed hash for unknown languages (helps users/maintainers add new checksums)
  </action>
  <verify>
Build succeeds:
```bash
cd /Users/lgbarn/Personal/pdf-cli/.worktrees/phase-2-concurrency
go build ./internal/ocr
```

Test download with checksum verification:
```bash
cd /Users/lgbarn/Personal/pdf-cli/.worktrees/phase-2-concurrency
# Remove existing tessdata to force fresh download
rm -rf ~/.config/pdf-cli/tessdata/eng.traineddata

# Test download - should verify checksum
go test -v ./internal/ocr -run TestEnsureTessdata

# Verify checksum message in output
```

Test checksum mismatch by temporarily corrupting a checksum in checksums.go, then restore.
  </verify>
  <done>
- downloadTessdata function updated with checksum verification
- SHA256 computed during download (no extra read pass)
- Checksum verification happens before file is renamed into place
- Clear error message on mismatch with both expected and computed hashes
- Warning for unknown languages includes computed hash
- Tests pass including checksum verification
  </done>
</task>

<task id="3" files="internal/ocr/ocr_test.go" tdd="true">
  <action>
Add comprehensive tests for checksum verification in OCR download.

Create or update internal/ocr/ocr_test.go with:

```go
func TestDownloadTessdataChecksumVerification(t *testing.T) {
  // This test requires network access and is slow, so skip in short mode
  if testing.Short() {
    t.Skip("Skipping network test in short mode")
  }

  // Create temp directory for test
  tmpDir := t.TempDir()

  // Test successful download and verification for a known language
  t.Run("valid_checksum", func(t *testing.T) {
    err := downloadTessdata(context.Background(), tmpDir, "eng")
    if err != nil {
      t.Fatalf("Expected successful download and verification, got error: %v", err)
    }

    // Verify file exists
    dataFile := filepath.Join(tmpDir, "eng.traineddata")
    if _, err := os.Stat(dataFile); os.IsNotExist(err) {
      t.Error("Expected file to exist after download")
    }
  })

  // Test unknown language (should warn but succeed)
  t.Run("unknown_language", func(t *testing.T) {
    // Use a language without checksum (if all common ones have checksums,
    // use a rare one like 'afr' or 'isl')
    // This test verifies warning is shown but download succeeds
    err := downloadTessdata(context.Background(), tmpDir, "afr")
    if err != nil {
      // If language doesn't exist on GitHub, that's also acceptable
      // The point is we don't fail on checksum
      t.Logf("Download failed (may be expected): %v", err)
    }
  })
}

func TestChecksumMismatch(t *testing.T) {
  // Unit test for checksum verification logic
  // This tests the verification code without actual download

  // Save original checksum
  origChecksum := KnownChecksums["eng"]
  defer func() {
    KnownChecksums["eng"] = origChecksum
  }()

  // Set wrong checksum
  KnownChecksums["eng"] = "0000000000000000000000000000000000000000000000000000000000000000"

  // Create temp dir and download
  tmpDir := t.TempDir()

  // This should fail with checksum mismatch
  err := downloadTessdata(context.Background(), tmpDir, "eng")
  if err == nil {
    t.Fatal("Expected checksum mismatch error, got nil")
  }

  if !strings.Contains(err.Error(), "checksum verification failed") {
    t.Errorf("Expected 'checksum verification failed' in error, got: %v", err)
  }
}

func TestPathSanitizationInDownload(t *testing.T) {
  // Test that malicious language codes are rejected
  tmpDir := t.TempDir()

  maliciousLangs := []string{
    "../../etc/passwd",
    "../../../tmp/evil",
    "../../escape",
  }

  for _, lang := range maliciousLangs {
    err := downloadTessdata(context.Background(), tmpDir, lang)
    if err == nil {
      t.Errorf("Expected error for malicious language '%s', got nil", lang)
    }
    if !strings.Contains(err.Error(), "directory traversal") &&
       !strings.Contains(err.Error(), "invalid") {
      t.Errorf("Expected traversal/invalid error for '%s', got: %v", lang, err)
    }
  }
}
```

Also add benchmark to measure overhead of checksum computation:
```go
func BenchmarkDownloadWithChecksum(b *testing.B) {
  tmpDir := b.TempDir()
  ctx := context.Background()

  b.ResetTimer()
  for i := 0; i < b.N; i++ {
    // Download small language file
    _ = downloadTessdata(ctx, tmpDir, "osd")
  }
}
```
  </action>
  <verify>
Run tests:
```bash
cd /Users/lgbarn/Personal/pdf-cli/.worktrees/phase-2-concurrency
# Run all checksum tests
go test -v ./internal/ocr -run Checksum

# Run download tests (may be slow)
go test -v ./internal/ocr -run Download

# Run in short mode (skip network tests)
go test -short -v ./internal/ocr

# Run benchmark
go test -bench=. ./internal/ocr
```

Verify coverage:
```bash
go test -cover ./internal/ocr
```
  </verify>
  <done>
- Comprehensive test suite for checksum verification
- Tests cover: valid checksum, unknown language, checksum mismatch, path sanitization
- Tests skip in short mode to avoid slow network operations
- Benchmark added to measure checksum overhead
- All tests pass
- Test coverage >85% for ocr package
  </done>
</task>

## Verification Strategy

Manual testing checklist:
1. Download tessdata for known language (eng) - should verify checksum
2. Download tessdata for unknown language - should warn but succeed
3. Corrupt a checksum in checksums.go - should fail download
4. Test with slow/interrupted network - temp file should be cleaned up
5. Verify computed hash is displayed for unknown languages

Security testing:
```bash
# Test supply chain protection
# 1. Download eng.traineddata
pdf text --ocr scanned.pdf

# 2. Manually corrupt downloaded file
echo "corrupted" >> ~/.config/pdf-cli/tessdata/eng.traineddata

# 3. Try to use it - OCR should still work (file already validated)
# 4. Delete file and re-download - should verify checksum again
rm ~/.config/pdf-cli/tessdata/eng.traineddata
pdf text --ocr scanned.pdf  # Should re-download and verify
```

Success criteria:
- All tessdata downloads for known languages verify checksums
- Checksum mismatches are rejected with clear error
- Unknown languages warn but allow download
- Computed hash displayed for all downloads
- No performance regression (SHA256 computed during download, not separate read)
- All tests pass including integration tests

## Breaking Changes

**None** - This is a security enhancement. Behavior is unchanged for users, except:
- Downloads now take slightly longer (SHA256 computation, ~1-2% overhead)
- Failed checksum verification will block OCR (intentional security feature)
- Warning messages for languages without embedded checksums

## Security Impact

**MEDIUM IMPACT**: Protects against:
1. Supply chain attacks (compromised GitHub serving malicious files)
2. Network corruption (incomplete/corrupted downloads)
3. MITM attacks (though HTTPS already protects this)

Does NOT protect against:
- Checksum database being compromised (checksums are in source code)
- Languages without embedded checksums (warns but allows)

Future enhancement: Consider downloading checksums from a signed manifest.

## Documentation Updates

Add to README.md in OCR section:
```
### Security

Downloaded tessdata files are verified using SHA256 checksums for integrity.
Supported languages have embedded checksums that are verified automatically.
For languages without checksums, a warning is displayed but download proceeds.

If checksum verification fails, it may indicate:
- Corrupted download (try again)
- Supply chain compromise (report issue immediately)
```
