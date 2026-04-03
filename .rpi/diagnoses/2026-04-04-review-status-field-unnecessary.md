---
date: 2026-04-04T01:25:00+02:00
topic: "review artifacts have unnecessary status field"
tags: [diagnosis]
status: resolved
---

# Diagnosis: review artifacts have unnecessary status field

## Bug Report

**Expected**: Review artifacts should not carry a `status` field — they are write-once artifacts that are complete the moment they're created.

**Actual**: All 5 reviews have `status: draft` because the verify-report template hardcodes it. No code ever transitions review status, so they remain `draft` forever.

**Reproduction**: `grep '^status:' .rpi/reviews/*.md` — all show `draft`.

## Root Cause

Both verify-report templates hardcode `status: draft`:
- `internal/workflow/assets/templates/verify-report.tmpl:5`
- `.rpi/templates/verify-report.tmpl:5`

No code transitions review status. The scanner reads it but nothing acts on it — reviews aren't filtered by status in any command, and the archivable filter skips nil-status artifacts (correct behavior for reviews).

## Investigation Log

### Attempt 1 — Remove status from templates

**Hypothesis**: Removing `status: draft` from both templates is sufficient. No code depends on review status.

**Verification**:
- `rpi status` handles nil status as "unknown" in counts — acceptable, reviews just show count without status breakdown
- Scanner archivable filter: `info.Status == nil → return false` — correct, reviews shouldn't be auto-archivable by status
- No other code paths filter reviews by status

**Change**: Removed `status: draft` line from both `.tmpl` files.

**Result**: All tests pass (`go test ./internal/workflow/... ./internal/template/... ./cmd/rpi/`). New reviews will be created without a status field.

## Resolution

**Status**: Resolved — removed `status: draft` from both verify-report templates. Existing reviews retain their `status: draft` (harmless, no need to bulk-edit).
