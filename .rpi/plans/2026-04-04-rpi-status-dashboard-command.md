---
date: 2026-04-04T00:30:18+02:00
design: .rpi/designs/2026-04-04-rpi-status-dashboard-command.md
spec: .rpi/specs/rpi-status-dashboard-command.md
status: active
tags:
    - plan
topic: rpi status dashboard command
---

# rpi status dashboard command — Implementation Plan

## Overview

Add an `rpi status` CLI command that aggregates RPI artifact metadata into a single-screen dashboard: artifact counts by type/status, stale warnings, active plan progress with checkbox completion, and archive readiness.

**Scope**: 2 new files (`cmd/rpi/status.go`, `cmd/rpi/status_test.go`)

## Source Documents
- **Design**: .rpi/designs/2026-04-04-rpi-status-dashboard-command.md
- **Spec**: .rpi/specs/rpi-status-dashboard-command.md

## Phase 1: Command skeleton + artifact summary

### Overview
Create the Cobra command with flags, scan artifacts, group by type/status, and render the Artifacts summary section. Establishes the output structure that later phases extend.

### Tasks:

#### 1. Cobra command + artifact aggregation
**File**: `cmd/rpi/status.go`
**Changes**:
- Define `statusCmd` with `Use: "status"`, `Short` description, `RunE: runStatus`
- Register flags in `init()`: `--rpi-dir` (via `addRpiDirFlag`), `--stale-days` (int, default 14), `--format` (string, default `"text"`)
- `runStatus`: call `scanner.Scan(rpiDir, scanner.Filters{})`, group results into `map[type]map[status]count`, render Artifacts section
- Output sorted by type (design, plan, research, review, spec), omit types with zero artifacts (ST-2)
- Use `text/tabwriter` for aligned columns, no color/unicode (ST-15)
- Archive directory already excluded by `scanner.Scan` (ST-3)

#### 2. Tests
**File**: `cmd/rpi/status_test.go`
**Changes**:
- Helper: `setupStatusTestDir(t)` — creates temp `.rpi/` with fixture artifacts across types/statuses
- Helper: `writeStatusFile(t, dir, relPath, content)` — writes a file with frontmatter
- `TestStatusArtifactSummary` (ST-1): 2 active specs + 1 draft plan + 1 complete design → verify output lines
- `TestStatusOmitsEmptyTypes` (ST-2): only specs and plans present → no designs/research/reviews rows
- `TestStatusExcludesArchive` (ST-3): artifact in `archive/` subdir → not counted

### Success Criteria:

#### Automated Verification:
- [x] `go test ./cmd/rpi/ -run TestStatus` — all pass
- [x] `go build ./cmd/rpi/` — compiles clean
- [x] `go vet ./cmd/rpi/` — no warnings

### Commit:
- [ ] Stage: `cmd/rpi/status.go`, `cmd/rpi/status_test.go`
- [ ] Message: `feat(status): add command skeleton with artifact summary`

---

## Phase 2: Staleness detection

### Overview
Parse frontmatter date fields, compute age, flag non-terminal artifacts exceeding the staleness threshold. Adds the Stale section to output.

### Tasks:

#### 1. Staleness logic
**File**: `cmd/rpi/status.go`
**Changes**:
- Add `computeStaleness` function: for each non-terminal artifact (status not `complete`/`superseded`/`archived`), parse `date` field (plans/designs/research/reviews) or `last_updated` (specs) from frontmatter via `frontmatter.Parse` (ST-5)
- Date parsing: try `time.RFC3339` first, fall back to `"2006-01-02"` layout. Skip on parse failure or missing field (ST-7)
- Compare against injectable `time.Time` (for testability) with `--stale-days` threshold (ST-6)
- Render Stale section: path, status, age in days (ST-8). Omit section header if no stale artifacts

#### 2. Tests
**File**: `cmd/rpi/status_test.go`
**Changes**:
- `TestStatusStaleDetection` (ST-4): plan dated 34 days ago, status draft → appears in Stale section with "34d ago"
- `TestStatusStaleCustomThreshold` (ST-6): plan 10 days old, `--stale-days 7` → stale; `--stale-days 14` → not stale
- `TestStatusStaleMissingDate` (ST-7): active spec with no `last_updated` → appears in summary, not in Stale, no error

### Success Criteria:

#### Automated Verification:
- [x] `go test ./cmd/rpi/ -run TestStatusStale` — all pass
- [x] `go vet ./cmd/rpi/` — no warnings

### Commit:
- [ ] Stage: `cmd/rpi/status.go`, `cmd/rpi/status_test.go`
- [ ] Message: `feat(status): add staleness detection with configurable threshold`

---

## Phase 3: Active plan chains + checkbox progress

### Overview
Resolve one-level chain links for active/draft plans, parse checkboxes for completion percentage. Adds the Active Plans section to output.

### Tasks:

#### 1. Chain resolution + progress
**File**: `cmd/rpi/status.go`
**Changes**:
- For each plan with status `active` or `draft`: call `chain.Resolve(planPath, chain.ResolveOptions{})` (ST-9)
- Filter to one level: only artifacts in the root artifact's `LinksTo` (not transitive links)
- For each linked artifact: display basename, type, status (ST-10)
- Read plan file content, call `parseCheckboxes(content)` (already accessible, same package) for checked/total count and percentage (ST-11)
- Plans with zero checkboxes: omit progress indicator (ST-12)
- Sort plans by path for deterministic output

#### 2. Tests
**File**: `cmd/rpi/status_test.go`
**Changes**:
- `TestStatusActivePlanChain` (ST-9, ST-10): active plan linking to design + spec, design links to research → output shows design + spec only (one level)
- `TestStatusCheckboxProgress` (ST-11): plan with 3 checked, 7 unchecked → "3/10 (30%)"
- `TestStatusNoCheckboxes` (ST-12): draft plan with no checkboxes → no progress indicator in output

### Success Criteria:

#### Automated Verification:
- [x] `go test ./cmd/rpi/ -run TestStatusActive` — all pass
- [x] `go test ./cmd/rpi/ -run TestStatusCheckbox` — all pass
- [x] `go vet ./cmd/rpi/` — no warnings

### Commit:
- [ ] Stage: `cmd/rpi/status.go`, `cmd/rpi/status_test.go`
- [ ] Message: `feat(status): add active plan chains with checkbox progress`

---

## Phase 4: Archive readiness + JSON output

### Overview
Identify archivable artifacts with zero references, add Ready to Archive section. Implement `--format json` output with all four top-level keys.

### Tasks:

#### 1. Archive readiness
**File**: `cmd/rpi/status.go`
**Changes**:
- Call `scanner.Scan(rpiDir, scanner.Filters{Archivable: true})`, then `scanner.CountReferences(rpiDir, path)` for each result (ST-13)
- Filter to zero-reference artifacts, group by type for summary line (ST-14)
- Omit section if no archivable artifacts

#### 2. JSON output
**File**: `cmd/rpi/status.go`
**Changes**:
- Define JSON output structs with keys: `summary`, `stale`, `active_plans`, `archivable` (ST-16)
- When `--format json`: marshal and print instead of text rendering
- Exit code 0 on success, 1 on errors (ST-17)

#### 3. Tests
**File**: `cmd/rpi/status_test.go`
**Changes**:
- `TestStatusArchiveReadiness` (ST-13, ST-14): complete design with 0 refs + complete design with 2 refs → only unreferenced one in output
- `TestStatusJSONOutput` (ST-16): run with `--format json` → valid JSON, has all four keys
- `TestStatusExitCodeOnError` (ST-17): invalid `--rpi-dir` → returns error

### Success Criteria:

#### Automated Verification:
- [x] `go test ./cmd/rpi/ -run TestStatus` — all pass
- [x] `go build ./cmd/rpi/` — compiles and runs `rpi status` against local `.rpi/`
- [x] `go vet ./cmd/rpi/` — no warnings

#### Manual Verification:
- [ ] Run `rpi status` against this repo's `.rpi/` directory — output is readable and fits one screen
- [ ] Run `rpi status --format json` — valid JSON with correct structure

### Commit:
- [ ] Stage: `cmd/rpi/status.go`, `cmd/rpi/status_test.go`
- [ ] Message: `feat(status): add archive readiness and JSON output`

---

## References
- Design: .rpi/designs/2026-04-04-rpi-status-dashboard-command.md
- Spec: .rpi/specs/rpi-status-dashboard-command.md
