# Shipyard Lessons Learned

## [2026-02-05] Milestone: Remaining Tech Debt (Phases 1-5)

### What Went Well
- Research-first approach saved implementation effort (R16 documented instead of rewritten)
- Parallel builder dispatch worked reliably across all phases
- Quality gates caught real issues (wrong task counts, stale README examples, verification gaps)
- Documentation-only phases are the smoothest — zero quality gate issues
- 80.6% coverage maintained throughout (threshold: 75%)

### Surprises / Discoveries
- Parallel builders can share commits via pre-commit hooks (Phases 3, 4)
- Pre-commit hooks revert STATE.md changes — requires re-application after each wave
- Architect agents don't always follow file naming conventions
- Verifier counts can be inaccurate — always verify in research

### Pitfalls to Avoid
- TDD with strict pre-commit hooks: can't commit failing tests separately
- golangci.yaml global disable of linter rules — prefer targeted exclusions
- Plan file paths may not match actual codebase — always verify

### Process Improvements
- Add file naming validation to architect agent output
- Consider pre-flight checks for parallel builders sharing git state
- Enhance verify commands to cover all documented changes (not just 2 of 5)

---
