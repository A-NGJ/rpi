---
archived_date: "2026-04-04"
date: 2026-04-04T01:18:37+02:00
plan: .rpi/plans/2026-04-04-rpi-status-dashboard-command.md
spec: .rpi/specs/rpi-status.md
status: archived
tags:
    - verify
topic: rpi status dashboard command
---

# Verification Report: rpi status dashboard command

## Summary

**Status: PASS** — All 19 spec behaviors (ST-1 through ST-19) verified against implementation. Tests pass, build is clean, no vet warnings. Two minor findings (1 warning, 1 note).

| Severity | Count |
|----------|-------|
| Blocker  | 0     |
| Warning  | 1     |
| Note     | 1     |

## Completeness

All 4 phases of the main plan have automated verification checked off. The follow-up plan (show artifact names / remove plan links) Phase 1 is also checked off. Tests cover all spec behaviors.

| Spec | Description | Status |
|------|-------------|--------|
| ST-1 | Artifact summary grouping | Pass — `TestStatusArtifactSummary` |
| ST-2 | Empty types omitted | Pass — `TestStatusOmitsEmptyTypes` |
| ST-3 | Exclude archive | Pass — `TestStatusExcludesArchive` |
| ST-4 | Stale detection | Pass — `TestStatusStaleDetection` |
| ST-5 | date vs last_updated | Pass — code at `status.go:254-257` |
| ST-6 | --stale-days flag | Pass — `TestStatusStaleCustomThreshold` |
| ST-7 | Missing date skipped | Pass — `TestStatusStaleMissingDate` |
| ST-8 | Stale display format | Pass — `status.go:163` renders path, status, age |
| ST-9 | One-level chain resolution | Pass — `TestStatusActivePlanChain` |
| ST-10 | No link sub-rows in text | Pass — `TestStatusActivePlanChain` asserts no `design:` sub-rows |
| ST-11 | Checkbox progress | Pass — `TestStatusCheckboxProgress` |
| ST-12 | No checkboxes = no progress | Pass — `TestStatusNoCheckboxes` |
| ST-13 | Archive readiness (0 refs) | Pass — `TestStatusArchiveReadiness` |
| ST-14 | Archive summary by type | Pass — `TestStatusArchiveReadiness` checks "1 designs" |
| ST-15 | Aligned text columns | Pass — uses `text/tabwriter` |
| ST-16 | JSON output with 4 keys | Pass — `TestStatusJSONOutput` |
| ST-17 | Exit code on error | Pass — `TestStatusExitCodeOnError` |
| ST-18 | Active Specs section | Pass — `TestStatusActiveSpecsSection` |
| ST-19 | Section order | Pass — code renders: Artifacts -> Active Plans -> Active Specs -> Stale -> Ready to Archive |

- No TODO/FIXME/HACK markers in `status.go` or `status_test.go`
- `go build`, `go vet`, `go test` all clean

## Correctness

All spec behaviors verified by reading code and cross-referencing tests. No silent deviations found. API contracts (JSON struct keys, flag names, exit codes) match the spec.

Constraints verified:
- Uses only existing internal packages (scanner, chain, frontmatter) — no new dependencies
- Registered as Cobra subcommand via `init()` following existing patterns
- Supports `--rpi-dir` flag via `addRpiDirFlag`
- Output is deterministic (sorted by type, then path)
- Read-only — no file writes or state modifications
- No color codes or Unicode special characters

## Coherence

- Naming follows existing codebase conventions (`runStatus`, `statusCmd`, `addRpiDirFlag`)
- Error handling consistent with other commands (return `error` from `RunE`)
- `nowFunc` override pattern for time-sensitive testing is clean and testable
- `statusTypeOrder` and `statusTypePlurals` include `diagnosis` type added in a follow-up fix commit

## Issues

### Blockers

None.

### Warnings

1. **Dead code: `activeByType["plan"]` populated but unused** — `status.go:89` collects active plan names into `activeByType["plan"]`, but `renderStatusText` only reads `activeByType["spec"]` (line 153). The Active Plans section uses `findActivePlans()` instead. The plan data is collected for nothing.
   - File: `cmd/rpi/status.go:89`
   - Fix: Either remove `a.Type == "plan"` from the condition on line 89, or use `activeByType["plan"]` for Active Plans display names instead of re-deriving them in `findActivePlans`.

### Notes

1. **Plan called for "Active Designs" section that was excluded** — The follow-up plan (`2026-04-04-rpi-status-show-artifact-names-and-remove-plan-links.md`) Task 2 specifies rendering an "Active Designs" section and creating a `TestStatusActiveDesignsSection` test. Neither was implemented, and the spec was updated to omit Active Designs. This appears to be a deliberate scope reduction. The plan's status is still `active` — consider updating it to reflect the final scope or marking it complete.

## References

- Spec: .rpi/specs/rpi-status.md
- Design: .rpi/designs/2026-04-04-rpi-status-dashboard-command.md
- Main plan: .rpi/plans/2026-04-04-rpi-status-dashboard-command.md
- Follow-up plan: .rpi/plans/2026-04-04-rpi-status-show-artifact-names-and-remove-plan-links.md
- Implementation: cmd/rpi/status.go, cmd/rpi/status_test.go
