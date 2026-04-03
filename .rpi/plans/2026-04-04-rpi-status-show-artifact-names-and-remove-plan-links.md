---
date: 2026-04-04T01:00:26+02:00
spec: .rpi/specs/rpi-status-dashboard-command.md
status: active
tags:
    - plan
topic: rpi status show artifact names and remove plan links
---

# rpi status show artifact names and remove plan links — Implementation Plan

## Overview

Modify `rpi status` text output to: (1) add dedicated Active Designs, Active Plans, and Active Specs sections that list each active artifact by name, and (2) remove nested design/spec link sub-rows from Active Plans. The existing Artifacts summary section remains unchanged. JSON output retains existing structure for backward compatibility.

**Scope**: 3 files modified (status.go, status_test.go, spec), 0 new files

## Source Documents
- **Spec**: .rpi/specs/rpi-status-dashboard-command.md

## Phase 1: Add active artifact sections and remove plan link sub-rows

### Overview

Keep the Artifacts count summary as-is. Add three new sections — Active Designs, Active Specs, Active Plans — each listing active artifacts by name. The existing Active Plans section gains the names but loses the nested design/spec sub-rows. Sections only appear when there are active artifacts of that type.

Target text output:
```
Artifacts
  designs:  2 active  1 draft
  plans:    1 active
  specs:    3 active

Active Designs
  Auth flow redesign
  Cache layer design

Active Plans
  Implement caching                       active    6/10 (60%)

Active Specs
  Auth permissions
  Cache invalidation
  Rate limiting

Stale (no update in 14+ days)
  ...

Ready to Archive
  ...
```

### Tasks:

#### 1. Collect active artifacts by type
**File**: `cmd/rpi/status.go`
**Changes**:
- Add a new collection (e.g., `activeByType map[string][]string`) populated during the existing `for _, a := range artifacts` loop (lines 77-86). For types `design`, `plan`, and `spec` with status `active`, record the artifact's title (fall back to filename basename without `.md` if `Title` is nil). Sort names alphabetically within each type.
- Pass this map through to `renderStatusText`.

#### 2. Render Active Designs and Active Specs sections
**File**: `cmd/rpi/status.go` — `renderStatusText`
**Changes**:
- After the Artifacts section (line 124) and before the existing Active Plans section, render an "Active Designs" section listing each active design by name — only if there are active designs.
- After the Active Plans section, render an "Active Specs" section listing each active spec by name — only if there are active specs.
- Section order: Artifacts → Active Designs → Active Plans → Active Specs → Stale → Ready to Archive.
- Each entry is indented with two spaces, showing just the name.

#### 3. Remove linked artifact sub-rows from Active Plans
**File**: `cmd/rpi/status.go` — `renderStatusText` (lines 142-144)
**Changes**:
- Remove the inner loop that prints `l.Type`, `l.Name`, `l.Status` under each active plan.
- Keep `findActivePlans` resolving chains so that `Links` is still populated for JSON output (`renderStatusJSON`).

#### 4. Update tests
**File**: `cmd/rpi/status_test.go`
**Changes**:
- **`TestStatusActivePlanChain`**: Remove assertions that `design:` and `spec:` sub-rows appear under Active Plans. Add assertion that `design:` does NOT appear in the Active Plans section.
- **New test `TestStatusActiveDesignsSection`**: Create active and non-active designs. Verify "Active Designs" section appears listing only active ones by name.
- **New test `TestStatusActiveSpecsSection`**: Create active and non-active specs. Verify "Active Specs" section appears listing only active ones by name.
- **New test `TestStatusNoActiveSectionWhenEmpty`**: Create only draft/complete artifacts for a type. Verify the corresponding "Active X" section does not appear.

#### 5. Update spec
**File**: `.rpi/specs/rpi-status-dashboard-command.md`
**Changes**:
- Revise **ST-10** to state that linked artifact sub-rows are no longer shown under Active Plans in text output (links remain in JSON).
- Add **ST-18**: When active designs exist, display an "Active Designs" section listing each by name.
- Add **ST-19**: When active specs exist, display an "Active Specs" section listing each by name.
- Add **ST-20**: Active Designs/Specs/Plans sections are omitted when no active artifacts of that type exist.

### Success Criteria:

#### Automated Verification:
- [x] `go test ./cmd/rpi/ -run TestStatus -count=1` — all status tests pass
- [x] `go vet ./cmd/rpi/` — no warnings
- [x] "Active Designs" section lists active design names (not shown when none active)
- [x] "Active Specs" section lists active spec names (not shown when none active)
- [x] Active Plans section shows plan names with progress, no nested design/spec sub-rows
- [x] JSON output still contains `links` array in `active_plans` entries

### Commit:
- [ ] Stage: `cmd/rpi/status.go`, `cmd/rpi/status_test.go`, `.rpi/specs/rpi-status-dashboard-command.md`
- [ ] Message: `feat(status): add active artifact sections, remove plan link sub-rows`

**Note**: If all success criteria are covered by automated checks and they pass, proceed to the next phase. Only pause for manual confirmation when the phase includes manual verification items.

---

## References
- Spec: .rpi/specs/rpi-status-dashboard-command.md
