---
date: 2026-04-04T20:27:41+02:00
topic: "verify index-expansion and unified-status-lifecycle specs"
tags: [verify]
spec: .rpi/specs/index-expansion.md (deleted)
---

# Verification Report: Index Expansion + Unified Status Lifecycle Specs

## Summary

Both specs are fully implemented and tested. **39 of 40 requirements verified as PASS**, 1 SKIP (UL-10 — applies to external skill behavior, not this codebase).

| Spec | Requirements | Pass | Skip |
|------|-------------|------|------|
| Index Expansion (IX) | 26 | 26 | 0 |
| Unified Status Lifecycle (UL) | 14 | 13 | 1 |

**Overall status: PASS**

## Completeness

### Index Expansion (IX-1 through IX-26) — 26/26 PASS

**Signature Filter (IX-1 to IX-4)**
- PASS IX-1: `QueryOptions.Signature` field with case-insensitive substring match — `internal/index/query.go:36`
- PASS IX-2: Signature composes with pattern/kind/exported filters — `internal/index/query.go:48-49`
- PASS IX-3: MCP tool `rpi_index_query` accepts `signature` parameter — `cmd/rpi/serve.go:348`
- PASS IX-4: CLI `--signature` flag — `cmd/rpi/index.go:163`

**Package Filter (IX-5 to IX-7)**
- PASS IX-5: `QueryOptions.Package` field with case-insensitive substring match — `internal/index/query.go:15`
- PASS IX-6: MCP tool accepts `package` parameter — `cmd/rpi/serve.go:349`
- PASS IX-7: CLI `--package` flag — `cmd/rpi/index.go:164`

**Package Overview (IX-8 to IX-11)**
- PASS IX-8: `QueryPackages` returns summaries with all required fields — `internal/index/query.go:59-67`
- PASS IX-9: Optional package name filter (substring match) — `internal/index/query.go:71`
- PASS IX-10: MCP tool `rpi_index_packages` — `cmd/rpi/serve.go:186-188`
- PASS IX-11: CLI `rpi index packages` — `cmd/rpi/index.go:112-126`

**Import Extraction (IX-12 to IX-19)**
- PASS IX-12: Go imports (single, block, aliased) — `internal/index/extract.go:277-285`
- PASS IX-13: Python imports (import, from-import) — `internal/index/extract.go:287-299`
- PASS IX-14: JS/TS imports (import, require) — `internal/index/extract.go:301-323`
- PASS IX-15: Rust imports (use, mod) — `internal/index/extract.go:325-375`
- PASS IX-16: Multi-line blocks handled for all languages — `internal/index/extract.go:215-382`
- PASS IX-17: Import records: file, import_path, alias, line — `internal/index/index.go:44-49`
- PASS IX-18: `Index.Imports []Import` field — `internal/index/index.go:56`
- PASS IX-19: Index version bumped to "2" — `internal/index/store.go:11`

**Import Queries (IX-20 to IX-24)**
- PASS IX-20: `QueryImports` with file substring match — `internal/index/query.go:120-130`
- PASS IX-21: `QueryImporters` with deduplicated results — `internal/index/query.go:133-148`
- PASS IX-22: MCP tool `rpi_index_imports` — `cmd/rpi/serve.go:191-193`
- PASS IX-23: MCP tool `rpi_index_importers` — `cmd/rpi/serve.go:195-198`
- PASS IX-24: CLI subcommands `imports` and `importers` — `cmd/rpi/index.go:128-154`

**MCP Descriptions (IX-25 to IX-26)**
- PASS IX-25: All 7 index tools use `mcpDescriptionWithPrefix` — `cmd/rpi/serve.go:170-198`
- PASS IX-26: New tools have action-oriented descriptions — `cmd/rpi/serve.go:187,192,197`

### Unified Status Lifecycle (UL-1 through UL-14) — 13/14 PASS, 1 SKIP

**State Machine (UL-1 to UL-6)**
- PASS UL-1: Valid statuses: draft, active, complete, superseded, archived — `internal/frontmatter/transition.go:6-10`
- PASS UL-2: `approved`/`implemented` rejected — `internal/frontmatter/frontmatter_test.go:183-211`
- PASS UL-3: All 7 allowed transitions match spec — `internal/frontmatter/transition.go:6-10`
- PASS UL-4: `archived` is terminal — `internal/frontmatter/transition.go:21-26`
- PASS UL-5: Missing status treated as draft — `internal/frontmatter/transition.go:16-19`
- PASS UL-6: complete→active (reopen) valid — `internal/frontmatter/transition.go:9`

**Specs Are Statusless (UL-7 to UL-10)**
- PASS UL-7: Spec template has no `status` field — `.rpi/templates/spec.tmpl`
- PASS UL-8: Specs excluded from archivable scanner — `internal/scanner/scan.go:118-121`
- PASS UL-9: `rpi archive move` works on specs — `cmd/rpi/archive.go:155-215`
- SKIP UL-10: Skills don't transition spec status — enforced at external skill level, not in this codebase

**Archivable Filter (UL-11 to UL-12)**
- PASS UL-11: Non-spec artifacts archivable when complete/superseded — `internal/scanner/scan.go:125-128`
- PASS UL-12: `implemented` not valid archivable status — excluded by UL-2 and scan filter

**Display and Tools (UL-13 to UL-14)**
- PASS UL-13: Display order: active, draft, complete, superseded — `cmd/rpi/status.go:64`
- PASS UL-14: `rpi_frontmatter_transition` lists valid targets only — `cmd/rpi/serve.go:137-139`, `cmd/rpi/frontmatter.go:21-27`

## Correctness

No deviations found. Implementation matches both specs precisely.

## Coherence

Both implementations follow existing codebase patterns — case-insensitive matching, MCP tool registration, CLI flag conventions, and test structure are consistent.

## Issues

### Blockers

None.

### Warnings

None.

### Notes

- **UL-10 (SKIP)**: The requirement that skills don't transition spec status is enforced at the Claude skill level (external to this repo), not in the Go codebase. Confirmed by absence of spec-specific transition logic in the MCP tools.

## References

- `.rpi/specs/index-expansion.md (deleted)` — IX spec (26 requirements)
- `.rpi/specs/status-lifecycle.md` — UL spec (14 requirements)
