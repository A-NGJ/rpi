---
date: 2026-04-04T00:25:51+02:00
status: complete
tags:
    - design
topic: rpi status dashboard command
---

# Design: rpi status dashboard command

## Summary

Add an `rpi status` CLI command that gives a single-screen overview of all RPI artifacts: counts by type/status, stale artifact warnings, active plan chains with completion percentages, and artifacts ready to archive. Composes existing internal packages (`scanner`, `chain`, `frontmatter`, `verify` checkbox parsing) with no new data stores.

## Context

When returning to a project after days or weeks, there's no way to quickly assess the state of RPI artifacts. Users must manually open each spec, plan, and design to check frontmatter statuses. The individual building blocks exist (`rpi scan`, `rpi chain`, `rpi verify completeness`, `rpi archive scan`) but nothing aggregates them into a human-readable overview.

This is CLI-only — not exposed as an MCP tool. Claude can already compose the individual tools; the dashboard is a human convenience.

## Constraints

- Must use only existing internal packages — no new data stores or caches
- Output must fit a single terminal screen for typical projects (< 30 artifacts)
- Staleness uses frontmatter date fields only (`date`, `last_updated`), no file mtime
- Chain display is one level deep (plan → direct design/spec links), not recursive

## Components

### 1. Artifact aggregation (`cmd/rpi/status.go`)

New Cobra command that calls `scanner.Scan(rpiDir, scanner.Filters{})` to get all artifacts, then groups them by type and status. This is a straightforward map-reduce over the existing `ArtifactInfo` slice.

### 2. Staleness detection

For each non-terminal artifact (status not `complete`/`superseded`/`archived`), parse the frontmatter date field (`date` for plans/designs/research, `last_updated` for specs) and compare against `time.Now()`. Artifacts exceeding the threshold (default 14 days, configurable via `--stale-days`) are flagged.

Date parsing: frontmatter dates are stored as either ISO timestamps (`2026-04-04T00:25:51+02:00`) or plain dates (`2026-04-04`). Both formats are handled via `time.Parse` with fallback patterns. Artifacts with unparseable or missing dates are silently skipped (no false positives).

### 3. Active plan chains

For each plan with status `active` or `draft`:
1. Call `chain.Resolve(planPath, chain.ResolveOptions{})` to get direct links
2. Filter to one level: only artifacts directly in the plan's `LinksTo`
3. Reuse `parseCheckboxes()` from `verify.go` to get completion percentage

This means `parseCheckboxes` (currently unexported, used only in `verify.go`) needs to stay accessible within the `main` package. Since both `status.go` and `verify.go` are in `cmd/rpi/` (same package), no refactoring is needed.

### 4. Archive readiness

Call `scanner.Scan(rpiDir, scanner.Filters{Archivable: true})` and then `scanner.CountReferences()` for each result — identical to the existing `rpi archive scan` logic. Display as a summary line with counts by type.

### 5. Text formatter

Terminal output with aligned columns using `fmt.Fprintf` and `text/tabwriter`. No color codes or Unicode box-drawing — keeps it simple and pipe-friendly. Structure:

```
Artifacts
  specs:    4 active  1 draft  1 complete
  plans:    2 active  2 draft
  designs:  2 complete
  reviews:  4 draft

Stale (no update in 14+ days)
  .rpi/plans/2026-03-19-update-readme-and-docs.md  draft  16d ago

Active Plans
  rpi-explain-command                                active  6/10 (60%)
    design: ...index-expansion...                    complete
    spec:   rpi-explain-command                      active
  benchmark-project...                               active  2/8 (25%)
    design: ...benchmark-project...                  complete
    spec:   benchmark-project...                     active

Ready to Archive
  2 artifacts (1 design, 1 spec) with 0 active references
```

JSON output is also supported via `--format json` for scripting.

## File Structure

| File | Change |
|------|--------|
| `cmd/rpi/status.go` | New — Cobra command, aggregation logic, text formatter |
| `cmd/rpi/status_test.go` | New — unit tests |
| `cmd/rpi/main.go` | No change needed (command registers itself via `init()`) |

## Risks

- **Performance on large `.rpi/` directories**: `scanner.Scan` walks the filesystem, and chain resolution + reference counting add per-artifact I/O. For projects with 100+ artifacts this could be slow. Mitigation: unlikely in practice — RPI artifacts are few. If needed, parallelism can be added later.
- **Date parsing edge cases**: Frontmatter dates are free-form YAML strings. Unrecognized formats silently skip staleness detection rather than erroring.

## Out of Scope

- MCP tool exposure (Claude composes individual tools)
- Color/emoji terminal output
- Archive directory contents (only active artifacts shown)
- Interactive mode or TUI
- Filtering flags on the status command (use `rpi scan` for filtered views)
