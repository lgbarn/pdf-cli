# Phase 3: Security Hardening - Implementation Plans

## Overview
This phase addresses all P0 critical security issues identified in the technical debt remediation roadmap. Breaking changes are acceptable as these are high-priority security fixes.

## Requirements
- **R1**: Passwords must not be visible in process listings — use stdin, env vars, or file-based input instead of CLI flags
- **R2**: Downloaded tessdata files must be verified with SHA256 checksums
- **R3**: All file path inputs must be sanitized against path traversal

## Plan Structure

### Wave 1 (Parallel Execution)
Both plans in Wave 1 can be executed in parallel as they have no dependencies on each other.

#### Plan 1.1: Password Security (R1)
- **File**: `plans/01-password-security.md`
- **Complexity**: Medium (3 tasks)
- **Dependencies**: None
- **Impact**: HIGH - Eliminates password exposure in process listings
- **Breaking**: Minor - Error messages change for missing passwords
- **Summary**: Create secure password reading with 4-tier priority: --password-file > PDF_CLI_PASSWORD > --password (deprecated) > interactive prompt. Update all 14 password-accepting commands.

#### Plan 1.2: Path Sanitization (R3)
- **File**: `plans/02-path-sanitization.md`
- **Complexity**: Medium (3 tasks)
- **Dependencies**: None
- **Impact**: HIGH - Prevents directory traversal attacks
- **Breaking**: None - Only rejects malicious paths
- **TDD**: Yes - Comprehensive test suite for SanitizePath
- **Summary**: Centralize path validation in internal/fileio/SanitizePath. Reject paths containing ".." after cleaning. Apply at all entry points and internal operations.

### Wave 2 (Sequential Execution)
Wave 2 depends on Wave 1 completion.

#### Plan 2.1: Tessdata Checksum Verification (R2)
- **File**: `plans/03-tessdata-checksums.md`
- **Complexity**: Medium (3 tasks)
- **Dependencies**: Plan 1.2 (requires path sanitization)
- **Impact**: MEDIUM - Protects against supply chain attacks
- **Breaking**: None - Minor performance overhead (~1-2%)
- **TDD**: Yes - Comprehensive checksum tests
- **Summary**: Embed SHA256 checksums for 10+ common languages. Verify downloads before renaming. Warn for unknown languages but allow.

## Execution Strategy

1. **Wave 1**: Execute Plans 1.1 and 1.2 in parallel
   - Plan 1.1: Password security (3 tasks)
   - Plan 1.2: Path sanitization (3 tasks)

2. **Wave 2**: Execute Plan 2.1 after Wave 1 completion
   - Plan 2.1: Tessdata checksums (3 tasks, depends on 1.2)

Total: **9 tasks across 3 plans**

## Files Touched

### New Files
- `internal/cli/password.go` - Secure password reading
- `internal/cli/password_test.go` - Password reading tests
- `internal/ocr/checksums.go` - Embedded SHA256 checksums
- `internal/ocr/checksums_test.go` - Checksum validation tests

### Modified Files
- `internal/cli/flags.go` - Add GetPasswordSecure, AddPasswordFileFlag
- `internal/fileio/files.go` - Add SanitizePath, update CopyFile
- `internal/fileio/files_test.go` - Add path sanitization tests
- `internal/ocr/ocr.go` - Add checksum verification, path sanitization
- `internal/ocr/ocr_test.go` - Add checksum verification tests
- `internal/commands/*.go` - 14+ files updated with password/path security

## Verification Criteria

### Security Testing
1. **Password exposure**: Verify passwords not visible in `ps aux`
2. **Directory traversal**: Test `pdf info "../../etc/passwd"` is rejected
3. **Checksum verification**: Test corrupted tessdata is rejected
4. **gosec scan**: No new security warnings

### Functional Testing
1. All password input methods work (file, env, deprecated flag, prompt)
2. Legitimate file paths (absolute, relative, stdin) still work
3. Tessdata downloads complete successfully with verification
4. All integration tests pass

### Performance Testing
1. Checksum computation adds <2% overhead to downloads
2. Path sanitization adds <1ms per file operation

## Risk Assessment

### High Risk Items
- **Password flag deprecation**: May break user scripts using --password
  - Mitigation: Keep flag with warning, document migration path

- **Checksum verification**: May reject valid downloads if checksums incorrect
  - Mitigation: Test against live tessdata_fast repository before release

### Medium Risk Items
- **Path sanitization**: May reject edge-case legitimate paths
  - Mitigation: Comprehensive test suite, allow stdin marker "-"

### Low Risk Items
- All other changes are defensive and well-tested

## Success Metrics

- [ ] Zero passwords visible in process listings
- [ ] Zero directory traversal vulnerabilities (gosec clean)
- [ ] 10+ languages with embedded checksums
- [ ] >90% test coverage for new security functions
- [ ] All existing integration tests pass
- [ ] No performance regression >5%

## Documentation Impact

### README Updates
- Add security section for password handling
- Document OCR checksum verification
- Update examples to show --password-file usage

### CHANGELOG Updates
```
### Security
- **BREAKING**: Passwords via --password flag now show deprecation warning
- Added secure password input: --password-file, PDF_CLI_PASSWORD env var, interactive prompt
- Added path sanitization to prevent directory traversal attacks
- Added SHA256 checksum verification for tessdata downloads

### Changed
- --password flag deprecated in favor of --password-file (more secure)
- File path validation now rejects directory traversal attempts
```

## Post-Implementation

### Follow-up Items
1. Monitor for user reports of path validation false positives
2. Consider adding checksums for more languages based on usage
3. Add security documentation to SECURITY.md
4. Consider signed checksum manifest in future

### Metrics to Track
- Percentage of users switching from --password to secure methods
- Number of directory traversal attempts blocked (telemetry)
- Checksum verification success rate
