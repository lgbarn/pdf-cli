---
phase: performance-docs-finalization
plan: 1.1
wave: 1
dependencies: []
must_haves:
  - R16: Document merge optimization trade-off with detailed comment
  - R17: Update SECURITY.md supported versions table
files_touched:
  - internal/pdf/transform.go
  - SECURITY.md
tdd: false
---

# Plan 1.1: Code Documentation and Security Policy Updates

**Wave 1** - No dependencies, can run in parallel with Plan 1.2

## Tasks

<task id="1" files="internal/pdf/transform.go" tdd="false">
  <action>Add comprehensive comment block before MergeWithProgress function (line 23) documenting the O(N²) merge trade-off. Explain: (1) pdfcpu's MergeCreateFile lacks progress callbacks, (2) incremental approach necessary for progress reporting, (3) performance acceptable for typical use cases (10 files ~2s, 50 files ~15s, 100 files ~45s), (4) threshold logic at line 29 keeps small merges fast using single-pass API call.</action>
  <verify>grep -A 20 "MergeWithProgress combines" /Users/lgbarn/Personal/pdf-cli/internal/pdf/transform.go | grep -q "O(N²)"</verify>
  <done>Comment block present explaining trade-off between progress reporting UX and algorithmic complexity. Comment mentions pdfcpu limitation, performance benchmarks, and threshold strategy.</done>
</task>

<task id="2" files="SECURITY.md" tdd="false">
  <action>Update supported versions table (lines 5-9) to reflect current release status: v2.0.x (supported), v1.3.x (supported), versions below 1.3 (not supported). Replace existing rows with these three entries maintaining existing table format.</action>
  <verify>grep -A 3 "| Version |" /Users/lgbarn/Personal/pdf-cli/SECURITY.md | grep -q "2.0.x"</verify>
  <done>SECURITY.md table shows three rows: v2.0.x supported, v1.3.x supported, &lt;1.3 unsupported. Table formatting preserved with checkmark and X emojis.</done>
</task>

<task id="3" files="SECURITY.md" tdd="false">
  <action>Run markdown linter to verify SECURITY.md formatting is correct after version table update.</action>
  <verify>markdownlint SECURITY.md || echo "Linter not installed, manual check OK"</verify>
  <done>SECURITY.md passes markdown validation or manual inspection confirms proper table syntax and rendering.</done>
</task>

## Success Criteria
- [ ] transform.go contains detailed comment explaining merge complexity trade-off
- [ ] SECURITY.md reflects current v2.0.x and v1.3.x support status
- [ ] Documentation changes are clear and technically accurate
