---
archived_date: "2026-04-04"
date: 2026-04-04T19:41:28+02:00
spec: .rpi/specs/rpi-status.md
status: archived
tags:
    - plan
topic: add requirement count to active specs display in rpi status
---

# add requirement count to active specs display in rpi status — Implementation Plan

## Overview

Show the number of behavioral requirements next to each spec name in the "Active Specs" section of `rpi status`. Requirements are identified by the `**XX-N**:` pattern in spec file bodies.

**Scope**: 3 files modified (status.go, status_test.go, spec)

## Source Documents
- **Spec**: .rpi/specs/rpi-status.md

## Phase 1: Add requirement counting to active specs display

### Overview
Introduce an `activeSpec` struct to carry both name and requirement count, count requirements by regex when collecting active specs, and render the count in both text and JSON output.

### Tasks:

#### 1. Add `activeSpec` struct and counting logic
**File**: `cmd/rpi/status.go`
**Changes**:
- Add `activeSpec` struct with `Name string` and `Requirements int` fields
- Add `countSpecRequirements(path string) int` function that reads the file and counts lines matching `\*\*\w+-\d+\*\*:` regex
- Change `activeByType` from `map[string][]string` to `map[string][]activeSpec`
- When collecting active specs (line ~89), also call `countSpecRequirements` and store the result
- Update `renderStatusText` signature and Active Specs rendering to display count (tab-aligned, e.g. `Auth Permissions  19 requirements`)
- Add `ActiveSpecs` field to `statusOutput` struct for JSON output; populate it in `renderStatusJSON`

#### 2. Update spec requirement ST-18
**File**: `.rpi/specs/rpi-status.md`
**Changes**:
- Update ST-18 to specify that each active spec displays its requirement count

#### 3. Tests
**File**: `cmd/rpi/status_test.go`
**Changes**:
- Add `TestStatusActiveSpecsRequirementCount`: create spec files with known requirement patterns, verify count appears in output
- Add test for spec with zero requirements (should show `0 requirements`)
- Update `TestStatusActiveSpecsSection` if output format changes affect existing assertions

### Success Criteria:

#### Automated Verification:
- [x] `go test ./cmd/rpi/ -run TestStatusActiveSpec` passes
- [x] `go test ./cmd/rpi/` passes (all existing tests still green)
- [x] `go vet ./cmd/rpi/` clean

### Commit:
- [ ] Stage: `cmd/rpi/status.go`, `cmd/rpi/status_test.go`, `.rpi/specs/rpi-status.md`
- [ ] Message: `feat(status): show requirement count next to active specs`

**Note**: If all success criteria are covered by automated checks and they pass, proceed to the next phase. Only pause for manual confirmation when the phase includes manual verification items.

---

## References
- Spec: .rpi/specs/rpi-status.md
