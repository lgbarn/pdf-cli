# Phase 1 Research: Dependency Updates and Go Version Alignment

**Research Date**: 2026-01-31
**Researcher**: Domain Researcher Agent
**Phase**: 1 - Dependency Updates and Go Version Alignment

## Executive Summary

This research covers Phase 1 of the pdf-cli technical debt remediation plan, focusing on dependency updates and Go version alignment. The analysis identified **21 outdated dependencies** with available updates and **inconsistent Go version specifications** across project files.

**Key Findings**:
- Current Go version in go.mod: `1.24.1`
- System Go version installed: `go1.25.6`
- README.md states: "Go 1.24 or later"
- 21 dependencies have newer versions available
- 2 dependencies use commit hashes instead of semantic versions
- CI workflow uses `go-version-file: 'go.mod'` (correctly aligned)
- No Dockerfile or other files reference Go version

**Recommended Actions**:
1. Update all 21 outdated dependencies to latest compatible versions
2. Update go.mod to Go 1.24.0 or 1.25.0 (align with stable releases)
3. Update README.md to match go.mod version requirement
4. Regenerate go.sum after dependency updates

---

## 1. Technology Options

### 1.1 Dependency Update Strategies

| Strategy | Pros | Cons | Maturity |
|----------|------|------|----------|
| **Update All at Once** | - Single PR/commit<br>- Faster completion<br>- One round of testing | - Higher risk if multiple breaking changes<br>- Harder to identify source of issues<br>- Larger code review | Standard practice for minor/patch updates |
| **Incremental by Category** | - Easier to isolate issues<br>- Smaller, focused PRs<br>- Better for code review | - More time-consuming<br>- Multiple test cycles<br>- Potential merge conflicts | Good for major version updates |
| **Update Only Critical** | - Minimal risk<br>- Focused on security | - Leaves technical debt<br>- Still outdated after update | Appropriate only for emergency fixes |
| **Automated (Dependabot)** | - Continuous updates<br>- No manual intervention<br>- Catches new releases quickly | - Already configured but not active<br>- Generates many PRs | Industry standard, production-ready |

### 1.2 Go Version Selection

| Version | Pros | Cons | Recommendation |
|---------|------|------|----------------|
| **Go 1.24.0** | - Matches current major version<br>- Stable release<br>- Latest features | - Released 2025, relatively new<br>- May exclude some users | Good choice |
| **Go 1.24.1** (current) | - Includes security/bug fixes<br>- Current specification | - Patch version in go.mod is unusual<br>- Not a baseline requirement | Not recommended for go.mod |
| **Go 1.25.0** | - Latest stable major<br>- Better compatibility with newer deps<br>- System already running 1.25.6 | - Requires code changes if using new features<br>- More restrictive | Best for future-proofing |
| **Go 1.21.x** | - Wider compatibility<br>- Matches some docs | - Older, missing features<br>- Not tested in current CI | Not recommended |

---

## 2. Recommended Approach

### 2.1 Dependency Update Strategy: **Update All at Once**

**Rationale**:
1. All updates are **minor or patch versions** - low risk of breaking changes
2. No major version bumps identified (e.g., v1.x → v2.x)
3. Project has **75% test coverage** with comprehensive test suite
4. **CI pipeline** includes linting, testing, race detection, and security scanning
5. Single update cycle is more efficient for technical debt remediation

**Exception**: Monitor these dependencies closely due to larger version jumps:
- `github.com/jerbob92/wazero-emscripten-embind`: v1.3.0 → v1.5.2 (2 minor versions)
- `github.com/tetratelabs/wazero`: v1.5.0 → v1.11.0 (6 minor versions, requires Go 1.24+)
- `github.com/onsi/ginkgo/v2`: v2.13.0 → v2.28.1 (15 minor versions)

### 2.2 Go Version Strategy: **Update to Go 1.25.0**

**Rationale**:
1. Latest stable Go version (go1.25.6 released 2026-01-15)
2. System already running Go 1.25.6 - better development experience
3. Future-proofs the project for upcoming dependencies
4. `wazero v1.11.0` requires Go 1.24+ anyway - aligns with requirement
5. Go 1.26 expected February 2026 - staying on 1.25 gives stable baseline

**Alignment Points**:
- Update `go.mod`: `go 1.25.0` (not 1.25.6 - use major.minor only)
- Update `README.md` line 70: "Go 1.25 or later"
- CI already uses `go-version-file: 'go.mod'` - no changes needed
- Makefile and GoReleaser don't hardcode Go version - no changes needed

---

## 3. Detailed Dependency Analysis

### 3.1 All Outdated Dependencies

| Package | Current | Latest | Type | Notes |
|---------|---------|--------|------|-------|
| `github.com/chengxilo/virtualterm` | v1.0.4 | v1.0.5 | indirect | Patch update |
| `github.com/clipperhouse/uax29/v2` | v2.2.0 | v2.4.0 | indirect | Minor update, Unicode text segmentation |
| `github.com/cpuguy83/go-md2man/v2` | v2.0.6 | v2.0.7 | indirect | Patch update |
| `github.com/danlock/pkg` | v0.0.17-a9828f2 | v0.0.46-2e8eb6d | indirect | **Commit hash → commit hash** |
| `github.com/go-logr/logr` | v1.2.4 | v1.4.3 | indirect | Minor update |
| `github.com/google/go-cmp` | v0.6.0 | v0.7.0 | indirect | Minor update |
| `github.com/google/pprof` | v0.0.0-20210407192527 | v0.0.0-20260115054156 | indirect | Major timestamp jump (~5 years) |
| `github.com/jerbob92/wazero-emscripten-embind` | v1.3.0 | v1.5.2 | indirect | Minor updates (2 versions) |
| `github.com/onsi/ginkgo/v2` | v2.13.0 | v2.28.1 | indirect | **15 minor versions** - test dependency |
| `github.com/onsi/gomega` | v1.29.0 | v1.39.1 | indirect | **10 minor versions** - test dependency |
| `github.com/stretchr/testify` | v1.9.0 | v1.11.1 | indirect | Minor update - test dependency |
| `github.com/tetratelabs/wazero` | v1.5.0 | v1.11.0 | indirect | **6 minor versions, requires Go 1.24+** |
| `golang.org/x/crypto` | v0.43.0 | v0.47.0 | indirect | Minor update |
| `golang.org/x/exp` | v0.0.0-20231006140011 | v0.0.0-20260112195511 | indirect | ~3 year timestamp jump |
| `golang.org/x/image` | v0.32.0 | v0.35.0 | indirect | Minor update |
| `golang.org/x/mod` | v0.28.0 | v0.32.0 | indirect | Minor update |
| `golang.org/x/net` | v0.45.0 | v0.49.0 | indirect | Minor update |
| `golang.org/x/sync` | v0.17.0 | v0.19.0 | indirect | Minor update |
| `golang.org/x/text` | v0.30.0 | v0.33.0 | indirect | Minor update |
| `golang.org/x/tools` | v0.37.0 | v0.41.0 | indirect | Minor update |
| `gopkg.in/check.v1` | v0.0.0-20161208181325 | v1.0.0-20201130134442 | indirect | **Version tag available** |

**Total**: 21 outdated dependencies
**Direct dependencies**: 0 (all updates are indirect/transitive)
**Test-only dependencies**: 3 (ginkgo, gomega, testify)

### 3.2 Dependencies with Commit Hashes (Direct Dependencies)

These packages use commit hashes instead of semantic versions in go.mod:

```go
github.com/danlock/gogosseract v0.0.11-0ad3421  // No update available
github.com/danlock/pkg v0.0.17-a9828f2          // Update to v0.0.46-2e8eb6d
```

**Impact**:
- Harder to track what changed between versions
- Dependabot may not detect updates
- Manual monitoring required

**Recommendation**: Update to latest commit hash but document in CHANGELOG

### 3.3 Dependency Update Commands

```bash
# Update all dependencies to latest
go get -u ./...

# Or update specific packages:
go get github.com/clipperhouse/uax29/v2@v2.4.0
go get github.com/jerbob92/wazero-emscripten-embind@v1.5.2
go get github.com/tetratelabs/wazero@v1.11.0
go get github.com/onsi/ginkgo/v2@v2.28.1
go get github.com/onsi/gomega@v1.39.1
go get github.com/stretchr/testify@v1.11.1
go get golang.org/x/crypto@v0.47.0
go get golang.org/x/exp@latest
go get golang.org/x/image@v0.35.0
go get golang.org/x/mod@v0.32.0
go get golang.org/x/net@v0.49.0
go get golang.org/x/sync@v0.19.0
go get golang.org/x/text@v0.33.0
go get golang.org/x/tools@v0.41.0

# Update Go version in go.mod
go mod edit -go=1.25

# Regenerate go.sum and clean up
go mod tidy
```

---

## 4. Potential Risks and Mitigations

### 4.1 Breaking Changes

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| **wazero v1.11.0 requires Go 1.24+** | High | Medium | Already planning to update to Go 1.25.0 |
| **wazero v1.11.0 adds golang.org/x/sys dependency** | High | Low | Transparent to consumers; no API changes |
| **Multiple golang.org/x packages updated** | Medium | Low | Well-maintained, backwards compatible |
| **Test framework updates (ginkgo/gomega)** | Medium | Low | Only affects test code, not production |
| **Commit hash dependencies changed** | Low | Medium | No semantic versioning to track changes |

### 4.2 Known Issues

#### wazero v1.11.0
- **Requires Go 1.24+** ([source](https://github.com/tetratelabs/wazero/releases))
- Adds `golang.org/x/sys` dependency (breaking "zero dependency" promise)
- Fixed race condition in refCount initialization (#2447)
- No API breaking changes identified

**Mitigation**: Update to Go 1.25.0 first, then update dependencies

#### Large Version Jumps in Test Dependencies
- `ginkgo/v2`: v2.13.0 → v2.28.1 (15 minor versions)
- `gomega`: v1.29.0 → v1.39.1 (10 minor versions)

**Impact**: Test-only dependencies, unlikely to affect production code
**Mitigation**: Run full test suite with race detector and coverage checks

### 4.3 Testing Strategy

**Pre-update Checklist**:
1. ✅ Comprehensive test suite exists (75%+ coverage)
2. ✅ CI pipeline includes race detector
3. ✅ Security scanning with gosec
4. ✅ Linting with golangci-lint v2.8.0
5. ✅ Integration tests available

**Post-update Verification**:
```bash
# 1. Update dependencies and Go version
go get -u ./...
go mod edit -go=1.25
go mod tidy

# 2. Verify build
make build

# 3. Run full test suite
make test-race

# 4. Check coverage threshold
make coverage-check

# 5. Lint
make lint

# 6. Security scan
gosec -exclude-dir=testdata ./...

# 7. Build all platforms
make build-all
```

### 4.4 Rollback Plan

If issues arise:
```bash
# Restore go.mod and go.sum from git
git checkout go.mod go.sum

# Or revert to specific versions
go get github.com/tetratelabs/wazero@v1.5.0
go mod tidy
```

---

## 5. Go Version Alignment

### 5.1 Current State

| Location | Current Value | Line Number |
|----------|---------------|-------------|
| `go.mod` | `go 1.24.1` | Line 3 |
| `README.md` | "Go 1.24 or later" | Line 70 |
| `README.md` | "Go 1.24 or later" | Line 579 |
| `.github/workflows/ci.yaml` | `go-version-file: 'go.mod'` | Line 23, 42, 87, 117 |
| `.goreleaser.yaml` | *(no explicit Go version)* | N/A |
| `Makefile` | *(no explicit Go version)* | N/A |

**System Go Version**: `go version go1.25.6 darwin/arm64`

### 5.2 Inconsistencies Found

1. **go.mod uses patch version**: `go 1.24.1` - unusual pattern (typically `go 1.24`)
2. **README.md specifies "1.24 or later"** but go.mod is 1.24.1
3. **CONCERNS.md mentions Go 1.21+** compatibility (outdated)
4. **System is running Go 1.25.6** but go.mod specifies 1.24.1

### 5.3 Recommended Alignment

**Target**: Go 1.25.0

| File | Change Required | New Value |
|------|-----------------|-----------|
| `go.mod` | Update line 3 | `go 1.25` |
| `README.md` | Update line 70 | "Go 1.25 or later" |
| `README.md` | Update line 579 | "Go 1.25 or later" |
| `.github/workflows/ci.yaml` | *(no change)* | Uses `go-version-file: 'go.mod'` |
| `.shipyard/codebase/STACK.md` | Update Go version reference | "Go 1.25" |

**No changes needed**:
- CI workflow (uses `go-version-file: 'go.mod'` - automatically aligned)
- Makefile (no hardcoded version)
- GoReleaser (no hardcoded version)

### 5.4 Go Version Format in go.mod

**Best Practice**: Use `go 1.X` format (major.minor only)

```go
// Current (non-standard)
go 1.24.1

// Recommended
go 1.25
```

**Rationale**: The `go` directive specifies the minimum language version, not the toolchain version. Patch versions are handled by the toolchain, not the language specification.

---

## 6. Relevant Documentation Links

### Go Language
- [Go 1.24 Release Notes](https://go.dev/doc/go1.24) - Official release notes
- [Go 1.25 Release Information](https://versionlog.com/golang/1.25/) - Version history
- [Go Release History](https://go.dev/doc/devel/release) - All releases
- [Go endoflife.date](https://endoflife.date/go) - Support timeline

### Key Dependencies
- [wazero Releases](https://github.com/tetratelabs/wazero/releases) - Full changelog
- [wazero v1.11.0 Package Docs](https://pkg.go.dev/github.com/tetratelabs/wazero) - API documentation
- [pdfcpu Documentation](https://github.com/pdfcpu/pdfcpu) - PDF library docs
- [Cobra Documentation](https://github.com/spf13/cobra) - CLI framework

### Project-Specific
- Project CONCERNS.md: `/Users/lgbarn/Personal/pdf-cli/.shipyard/codebase/CONCERNS.md`
- Project STACK.md: `/Users/lgbarn/Personal/pdf-cli/.shipyard/codebase/STACK.md`
- Dependabot config: `/Users/lgbarn/Personal/pdf-cli/.github/dependabot.yaml`

---

## 7. Implementation Considerations

### 7.1 Integration Points

**Files Requiring Changes**:
1. `go.mod` - Update Go version and dependencies
2. `go.sum` - Regenerate checksums
3. `README.md` - Update Go version requirement (2 locations)
4. `.shipyard/codebase/STACK.md` - Update documentation

**Files Requiring Verification** (no changes):
5. All test files - Run full suite
6. CI workflow - Verify pipeline passes
7. Build artifacts - Test cross-compilation

### 7.2 Migration Steps

```bash
# Step 1: Create feature branch
git checkout -b phase-1-dependency-updates

# Step 2: Update Go version in go.mod
go mod edit -go=1.25

# Step 3: Update dependencies
go get -u ./...

# Step 4: Clean up
go mod tidy

# Step 5: Verify build
make build

# Step 6: Run tests
make test-race
make coverage-check

# Step 7: Lint
make lint

# Step 8: Update documentation
# (Manual edit of README.md and STACK.md)

# Step 9: Commit changes
git add go.mod go.sum README.md .shipyard/codebase/STACK.md
git commit -m "chore: update Go to 1.25 and all dependencies to latest versions"
```

### 7.3 Testing Strategy

**Regression Testing**:
- All existing tests must pass (75%+ coverage threshold)
- Race detector must pass (`make test-race`)
- Linter must pass (`make lint`)
- Security scan must pass (`gosec`)
- Cross-compilation must succeed for all platforms

**Specific Test Areas**:
1. **OCR functionality** (wazero updated) - Test WASM backend
2. **PDF operations** - Full command suite
3. **CLI interactions** - All commands with various flags
4. **Configuration loading** - YAML parsing
5. **Progress bars** - Terminal output

### 7.4 Performance Implications

**Expected Changes**:
- wazero v1.11.0 may have performance improvements (check release notes)
- Go 1.25 includes compiler and runtime improvements
- Updated golang.org/x packages may have optimizations

**Benchmarking** (if needed):
```bash
# Run benchmarks before update
go test -bench=. -benchmem ./... > bench_before.txt

# Update dependencies

# Run benchmarks after update
go test -bench=. -benchmem ./... > bench_after.txt

# Compare
benchstat bench_before.txt bench_after.txt
```

### 7.5 Rollback Considerations

**Low Risk Rollback**:
- All changes are in `go.mod`, `go.sum`, and documentation
- No code changes required for dependency updates
- Easy to revert via git if issues arise

**Rollback Command**:
```bash
git checkout main -- go.mod go.sum
go mod download
```

---

## 8. Success Criteria

**Phase 1 Complete When**:
- ✅ All 21 outdated dependencies updated to latest versions
- ✅ Go version updated to 1.25 in go.mod
- ✅ README.md updated to reflect Go 1.25 requirement
- ✅ go.sum regenerated and verified
- ✅ All tests pass with 75%+ coverage
- ✅ CI pipeline passes (lint, test, build, security)
- ✅ Cross-platform builds succeed
- ✅ No new security warnings from gosec
- ✅ Documentation updated (STACK.md)

---

## 9. Timeline Estimate

| Task | Estimated Time | Notes |
|------|----------------|-------|
| Update dependencies | 30 minutes | Mostly automated |
| Update Go version | 15 minutes | Simple change |
| Regenerate go.sum | 5 minutes | Automatic |
| Run full test suite | 10 minutes | CI takes ~5-10 min |
| Fix any test failures | 1-2 hours | Contingency (likely not needed) |
| Update documentation | 30 minutes | README.md, STACK.md |
| Code review | 1 hour | PR review time |
| **Total (best case)** | **2-3 hours** | If no test failures |
| **Total (with fixes)** | **4-5 hours** | If minor issues found |

---

## 10. Open Questions

1. **Should we update to Go 1.25 or stay on 1.24?**
   - **Recommendation**: Go 1.25 for future-proofing and alignment with system version

2. **Should commit-hash dependencies be replaced with tagged releases?**
   - `github.com/danlock/gogosseract` and `github.com/danlock/pkg` use commit hashes
   - **Recommendation**: Update to latest commit hashes for now, investigate tagged releases in Phase 5

3. **Should Dependabot be enabled for automated updates?**
   - Configuration exists but may not be active
   - **Recommendation**: Enable after Phase 1 to catch future updates

4. **Is there a maintenance window for this update?**
   - No code changes, low risk
   - **Recommendation**: Can be done anytime, no downtime required

---

## 11. Conclusion

Phase 1 is a **low-risk, high-value** update that addresses technical debt and security concerns. The update strategy (all-at-once) is appropriate given:

1. All updates are minor/patch versions
2. Comprehensive test coverage and CI pipeline
3. No identified breaking changes (except wazero requiring Go 1.24+, which we're addressing)
4. Single update cycle is more efficient

**Primary Risk**: wazero v1.11.0 requiring Go 1.24+ is already mitigated by updating to Go 1.25.

**Next Steps**:
1. Approve Go 1.25 as target version
2. Execute update commands in feature branch
3. Run full test suite
4. Update documentation
5. Submit PR for review

---

**End of Research Document**

---

## Sources

- [Go 1.24 Release Notes](https://go.dev/doc/go1.24)
- [Go Release History](https://go.dev/doc/devel/release)
- [Go endoflife.date](https://endoflife.date/go)
- [State of GO 2026](https://devnewsletter.com/p/state-of-go-2026)
- [wazero Releases](https://github.com/tetratelabs/wazero/releases)
- [wazero Package Documentation](https://pkg.go.dev/github.com/tetratelabs/wazero)
