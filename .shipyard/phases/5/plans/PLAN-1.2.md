---
phase: performance-docs-finalization
plan: 1.2
wave: 1
dependencies: []
must_haves:
  - R15: Document WASM thread-safety limitations in troubleshooting
  - R18: Update README password flag documentation and security warnings
files_touched:
  - README.md
tdd: false
---

# Plan 1.2: README User Documentation Updates

**Wave 1** - No dependencies, can run in parallel with Plan 1.1

## Tasks

<task id="1" files="README.md" tdd="false">
  <action>Add new troubleshooting section after line 782 ("The first time you use WASM OCR..."). Title: "OCR performance with WASM backend". Content: Explain WASM backend forces sequential processing due to lack of thread-safety, recommend native Tesseract (brew install tesseract) for large batch OCR operations, note WASM is convenient for single-file or small-batch use.</action>
  <verify>grep -A 5 "OCR performance with WASM backend" /Users/lgbarn/Personal/pdf-cli/README.md</verify>
  <done>New troubleshooting section present explaining WASM sequential processing limitation and recommending native Tesseract for performance-critical workloads.</done>
</task>

<task id="2" files="README.md" tdd="false">
  <action>Update password flag documentation (lines 463-530): (1) Line 463: Update --password description to explicitly state it's deprecated and shows error if used without --allow-insecure-password flag. (2) Add new table row after line 463 for --allow-insecure-password flag. (3) Line 465: Change default log level from "silent" to "error". (4) Line 482: Remove or fix insecure password example in dry-run section. (5) Lines 526-530: Update flag description from "shows warning" to "shows error (requires --allow-insecure-password to use)".</action>
  <verify>grep -q "allow-insecure-password" /Users/lgbarn/Personal/pdf-cli/README.md && grep "default: error" /Users/lgbarn/Personal/pdf-cli/README.md</verify>
  <done>Password flag documentation reflects current error behavior (not warning), --allow-insecure-password flag documented, default log level shows "error", dry-run example does not show insecure password usage.</done>
</task>

<task id="3" files="README.md" tdd="false">
  <action>Add security warning box after line 530 in the "Working with Encrypted PDFs" section. Use markdown alert syntax ("> [!WARNING]"). Content: "Avoid using --password flag in production scripts or shared environments. Process listings (ps, htop) expose command-line arguments including passwords. Prefer --password-file, environment variables, or interactive prompts for sensitive operations."</action>
  <verify>grep -A 3 "\[!WARNING\]" /Users/lgbarn/Personal/pdf-cli/README.md | grep -q "process listings"</verify>
  <done>Warning box present explaining process listing security risk, recommending secure alternatives to --password flag.</done>
</task>

## Success Criteria
- [ ] WASM performance limitation documented in troubleshooting section
- [ ] Password flag documentation accurate to current v2.0 behavior (error not warning)
- [ ] --allow-insecure-password flag documented
- [ ] Default log level corrected to "error"
- [ ] Security warning about process listings present
- [ ] No examples showing insecure password usage
