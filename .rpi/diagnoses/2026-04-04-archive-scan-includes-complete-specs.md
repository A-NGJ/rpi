---
title: archive scan includes complete specs
status: complete
date: "2026-04-04"
---

## Bug Report

**Expected**: `rpi archive scan` should only include specs with status "superseded" — specs are living documents and "complete" or "implemented" statuses don't indicate they should be archived.

**Actual**: Specs with any archivable status ("complete", "superseded", "implemented") appeared in scan results.

**Reproduction**: Run `rpi archive scan` with a spec that has `status: complete` in frontmatter — it appears as an archive candidate.

## Root Cause

`internal/scanner/scan.go:118-126` — the `matches()` function applied the same archivable status filter to all artifact types without distinguishing specs from designs/plans/research.

## Investigation Log

1. **Hypothesis**: The scan filter doesn't differentiate artifact types for archivability.
   **Finding**: Confirmed. The filter checked `s != "complete" && s != "superseded" && s != "implemented"` uniformly for all types.

## Resolution

**Status**: fixed

**Changes**:
- `internal/scanner/scan.go:118-130` — added `info.Type == "spec"` branch that only allows "superseded" status for specs
- `internal/scanner/scan_test.go` — updated `TestScanArchivable` expected count from 3→2, added `TestScanArchivableSpecSuperseded` for positive case
