---
title: rpi status shows specs with legacy statuses
status: complete
date: "2026-04-04"
---

## Bug Report

**Expected**: `rpi status` should not show specs in the Artifacts section (since specs are statusless), and the dedicated specs section should be labeled "Specifications" listing all specs.

**Actual**: Specs appeared in the Artifacts section grouped by status (active, draft, complete, unknown), and the dedicated section was labeled "Active Specs" filtering only `status: active` specs.

**Reproduction**: Run `rpi status` with specs that have legacy `status:` frontmatter — they appear in the Artifacts summary with status counts.

## Root Cause

`cmd/rpi/status.go:80-88` — the main artifact loop counted specs in the `summary` map like any other type, and line 90 filtered specs by `status == "active"` for the dedicated section. After the statusless-specs refactoring (commit 8ba9f89), specs should no longer participate in status-based grouping.

## Investigation Log

1. **Hypothesis**: The status command was not updated when specs became statusless.
   **Finding**: Confirmed. Five locations needed updating: the summary loop, the Specifications section heading, the JSON struct field, the staleness filter (still checked spec status), and a stale test in `internal/template/render_test.go`.

## Resolution

**Status**: fixed

**Changes**:
- `cmd/rpi/status.go` — removed "spec" from `statusTypeOrder` and `statusTypePlurals`; restructured artifact loop to collect all specs separately (no status filtering); renamed "Active Specs" → "Specifications"; renamed JSON field `active_specs` → `specs`; excluded specs from staleness detection
- `cmd/rpi/status_test.go` — updated 7 tests to reflect specs not appearing in Artifacts section, "Specifications" heading, and `Specs` JSON field
- `internal/template/render_test.go` — removed stale `status: draft` expectation from spec template test
