# Plan 1.1: Dependency Updates and Go Version Alignment

## Context
This plan updates all 21 outdated dependencies to their latest compatible versions and aligns the Go version string across all project files. The Go version will be updated from 1.24.1 to 1.25 (matching the development environment running Go 1.25.6), and all dependencies will be updated using `go get -u`. All updates are minor or patch versions, representing low-risk changes with no expected breaking changes.

Key dependency updates include:
- wazero: v1.5.0 → v1.11.0 (largest jump, requires Go 1.24+)
- ginkgo: v2.13.0 → v2.28.1
- 19 other minor/patch updates across direct and transitive dependencies

This is a pure dependency maintenance task with no functional code changes.

## Dependencies
None - this is the first plan in Phase 1.

## Tasks

### Task 1: Update Go Version and Dependencies
**Files:**
- `/Users/lgbarn/Personal/pdf-cli/go.mod`
- `/Users/lgbarn/Personal/pdf-cli/go.sum`

**Action:** modify

**Description:**
Update the Go version directive and all dependencies in go.mod:

1. Update Go version from 1.24.1 to 1.25:
   ```bash
   cd /Users/lgbarn/Personal/pdf-cli
   go mod edit -go=1.25
   ```

2. Update all dependencies to latest compatible versions:
   ```bash
   go get -u ./...
   ```

3. Clean up and verify go.mod/go.sum:
   ```bash
   go mod tidy
   ```

4. Verify the build still works:
   ```bash
   go build ./...
   ```

**Acceptance Criteria:**
- go.mod contains `go 1.25` (not 1.24.1)
- All 21 previously outdated dependencies are updated to latest versions
- `go mod tidy` produces no additional changes
- `go build ./...` completes successfully with no errors

### Task 2: Update README Go Version References
**Files:**
- `/Users/lgbarn/Personal/pdf-cli/README.md`

**Action:** modify

**Description:**
Update the two Go version references in README.md to match the new version:

1. Line 71: Change "Go 1.24 or later" to "Go 1.25 or later"
2. Line 579: Change "Go 1.24 or later" to "Go 1.25 or later"

Use exact string replacement to maintain consistency with surrounding documentation.

**Acceptance Criteria:**
- Line 71 contains "Go 1.25 or later"
- Line 579 contains "Go 1.25 or later"
- No other changes to README.md
- Formatting and markdown structure preserved

### Task 3: Verify Build and Tests
**Files:** N/A (verification task)

**Action:** test

**Description:**
Run the full test suite with race detection to verify all dependency updates are compatible and no regressions were introduced:

1. Run tests with race detection:
   ```bash
   cd /Users/lgbarn/Personal/pdf-cli
   go test -race ./...
   ```

2. Run tests with coverage (optional verification):
   ```bash
   go test -cover ./...
   ```

3. Verify no new lint issues were introduced:
   ```bash
   go vet ./...
   ```

**Acceptance Criteria:**
- `go test -race ./...` passes with all tests successful
- No new race conditions detected
- `go vet ./...` reports no issues
- No compilation errors or warnings

## Verification

After completing all tasks, verify the plan was executed correctly:

1. **Version Consistency Check:**
   ```bash
   cd /Users/lgbarn/Personal/pdf-cli
   grep -n "^go " go.mod
   grep -n "Go 1\." README.md
   ```
   Expected: go.mod shows `go 1.25`, README shows "Go 1.25 or later" at lines 71 and 579

2. **Dependency Freshness:**
   ```bash
   go list -u -m all
   ```
   Expected: No available updates (all dependencies at latest versions)

3. **Clean State:**
   ```bash
   go mod tidy
   git diff go.mod go.sum
   ```
   Expected: No changes after running `go mod tidy`

4. **Full Build and Test:**
   ```bash
   go build ./... && go test -race ./...
   ```
   Expected: All builds and tests pass successfully

5. **CI Compatibility:**
   The CI workflow uses `go-version-file: 'go.mod'`, so it will automatically pick up the Go 1.25 version - no CI configuration changes needed.

## Risk Assessment

**Low Risk** - All dependency updates are minor/patch versions with no known breaking changes. The Go version update from 1.24 to 1.25 is a minor version update within the same major release.

## Estimated Effort

**Small (S)** - Approximately 15-30 minutes of work:
- Task 1: 10-15 minutes (dependency updates + verification)
- Task 2: 2-3 minutes (simple text replacement)
- Task 3: 5-10 minutes (test execution)

## Rollback Plan

If issues are discovered after merging:
1. Revert the commit using `git revert`
2. Dependencies will revert to previous versions
3. Go version will revert to 1.24.1
4. No code changes means no functional impact to roll back
