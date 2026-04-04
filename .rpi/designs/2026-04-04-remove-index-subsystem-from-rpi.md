---
date: 2026-04-04T20:55:13+02:00
status: complete
tags:
    - design
topic: remove index subsystem from rpi
---

# Design: Remove Index Subsystem from RPI

## Summary

Remove the entire index subsystem (codebase symbol indexing, import graph, package overviews) from RPI. The index provides capabilities that AI agents already have natively through grep, glob, and LSP — making it unused overhead that adds complexity and pollutes the MCP tool list.

## Context

The index was built to give AI agents a structural map of the codebase: symbols, imports, packages. In practice, agents never call the 7 index MCP tools because they have superior built-in alternatives:

- **Symbol search** → agents use grep/glob/LSP, which are always up-to-date and language-aware
- **Import graph** → agents grep for import statements on demand
- **Package overviews** → agents use `rpi_scan` and direct file exploration

The index is regex-based (inherently less accurate than LSP), requires manual rebuilding (`rpi index build`), and adds ~1,700 lines of code + tests for zero observed utility.

## Constraints

- Must not break any remaining CLI commands or MCP tools
- Must cleanly remove all dead code — no stubs or deprecation shims
- Must update all documentation that references the index
- Must archive the `index-expansion` spec

## Components

### Files to delete (10 files, ~1,100 lines)

| File | Purpose |
|------|---------|
| `internal/index/index.go` | Data structures, `Build()` |
| `internal/index/store.go` | JSON persistence (`Save`/`Load`) |
| `internal/index/query.go` | Query operations |
| `internal/index/extract.go` | Symbol/import extraction |
| `internal/index/languages.go` | Language detection, regex configs |
| `internal/index/index_test.go` | Build/store tests |
| `internal/index/extract_test.go` | Extraction tests |
| `internal/index/query_test.go` | Query tests |
| `cmd/rpi/index.go` | CLI `rpi index` command + subcommands |
| `cmd/rpi/index_test.go` | CLI command tests |

### Files to modify

| File | Change |
|------|--------|
| `cmd/rpi/serve.go` | Remove 7 MCP tool registrations, input structs, handlers (~150 lines) |
| `cmd/rpi/serve_test.go` | Remove 7 tools from `expectedTools` |
| `cmd/rpi/init_cmd.go` | Remove `.rpi/index.json` gitignore entry |
| `README.md` | Remove index references |
| `docs/architecture.md` | Remove index feature description |
| `docs/rpi-init.md` | Remove index mention |
| `docs/thoughts-directory.md` | Remove index from directory structure |

### Artifacts to archive

- `.rpi/specs/index-expansion.md (deleted)` → archive as complete (the feature it specified is being intentionally removed)

## File Structure

```
# Deleted
internal/index/           # entire directory
cmd/rpi/index.go
cmd/rpi/index_test.go

# Modified
cmd/rpi/serve.go          # remove index tool registrations + handlers
cmd/rpi/serve_test.go     # remove index from expected tools
cmd/rpi/init_cmd.go       # remove index.json gitignore entry
README.md                 # remove index references
docs/architecture.md      # remove index section
docs/rpi-init.md          # remove index mention
docs/thoughts-directory.md # remove index from structure
```

## Risks

- **Low**: Someone has a workflow depending on `rpi index` commands → mitigated by the observation that agents don't use it; CLI users can use standard tools instead
- **Low**: Generated `.rpi/index.json` left behind in existing projects → harmless, gets ignored; users can delete manually

## Out of Scope

- Replacing index with an alternative code navigation system
- Modifying any non-index MCP tools or commands

## References

- `.rpi/specs/index-expansion.md (deleted)` — the spec being retired
