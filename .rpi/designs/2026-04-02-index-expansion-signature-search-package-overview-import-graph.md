---
date: 2026-04-02T01:39:57+02:00
status: complete
tags:
    - design
topic: index expansion — signature search, package overview, import graph
---

# Design: index expansion — signature search, package overview, import graph

## Summary

Expand the codebase index with three capabilities: (1) signature-based symbol search, (2) package-level overview queries, and (3) an import graph. Together these make the index a genuine accelerator for codebase discovery — answering structural questions that Grep/Glob/Read handle poorly.

## Context

The index (`internal/index/`) is fully implemented and tested but underutilized. It builds during `rpi init`/`rpi update`, exposes 4 MCP tools, and stores symbols with name, kind, file, line, package, scope, signature, and export status. However:

- **No skill references it**, so the LLM doesn't know when to prefer it over Grep/Glob.
- **Queries are limited** to name substring + kind/exported filters. The `Signature` field is stored but not queryable. Package info exists but can't be aggregated.
- **No import data** is captured, so "what does this file depend on?" and "who uses this package?" require manual grepping with language-specific patterns.

Additionally, the MCP tool descriptions are mechanical ("case-insensitive substring match") rather than action-oriented. They don't communicate when the index beats Grep.

## Constraints

- **Regex-only extraction** — no AST parsing, no external tools. Must work with the existing line-by-line scanner in `extract.go`.
- **All 4 languages** (Go, Python, JS/TS, Rust) for every feature, including imports.
- **Multi-line imports** — Go `import (...)` blocks, JS/TS multi-line imports, Python multi-line `from x import (...)` — require a small state machine in the extractor.
- **Index version** — adding `Imports` to the `Index` struct requires a version bump. Existing indexes must fail cleanly with a "rebuild" message (already handled by `store.go:41`).
- **MCP tool count** — currently 20 tools registered. Adding 3 new tools brings it to 23. Keep the count reasonable.

## Components

### 1. Signature Search (extend existing)

**What:** Add a `signature` filter to `QueryOptions` and `indexQueryInput`. Case-insensitive substring match on `Symbol.Signature`, composable with existing `pattern`/`kind`/`exported` filters.

**Why not merge into `pattern`?** The common case is name search. A separate parameter keeps name queries fast and lets the LLM combine both (e.g., "functions named Handle that take context.Context").

**Changes:**
- `internal/index/query.go` — add `Signature string` to `QueryOptions`, add filter in `QuerySymbols`
- `cmd/rpi/serve.go` — add `signature` field to `indexQueryInput`
- `cmd/rpi/index.go` — add `--signature` flag to `indexQueryCmd`

### 2. Package Overview (new tool + filter)

Two sub-features:

**a) Package filter on existing query** — add `package` field to `QueryOptions`/`indexQueryInput`. Lets the LLM scope any symbol query to a specific package (e.g., "exported functions in package `index`").

**b) New `rpi_index_packages` tool** — aggregates the existing data into a structural overview:

```json
[
  {
    "name": "index",
    "files": ["internal/index/query.go", "internal/index/index.go", ...],
    "file_count": 6,
    "exported_symbols": 12,
    "total_symbols": 28,
    "kinds": {"function": 8, "struct": 3, "method": 15, "interface": 2}
  }
]
```

Optional `package` filter parameter to zoom into one package.

**Why a separate tool?** The return shape (packages, not symbols or files) is fundamentally different. A distinct tool name helps the LLM know when to reach for it ("what does this package expose?" vs. "find symbol X").

**Changes:**
- `internal/index/query.go` — add `Package string` to `QueryOptions`, add `QueryPackages` function returning `[]PackageSummary`
- `cmd/rpi/serve.go` — add `package` field to `indexQueryInput`, register new `rpi_index_packages` tool
- `cmd/rpi/index.go` — add `--package` flag to `indexQueryCmd`, add `packages` subcommand

### 3. Import Graph (new extraction + new tools)

**New data structure:**

```go
type Import struct {
    File       string `json:"file"`        // file containing the import
    ImportPath string `json:"import_path"` // what's imported
    Alias      string `json:"alias"`       // import alias ("" if none)
    Line       int    `json:"line"`        // line number
}
```

Added to `Index` as `Imports []Import`. Version bump required.

**Extraction — language-specific regex patterns:**

| Language | Patterns | Multi-line handling |
|---|---|---|
| Go | `import "path"`, `import ( ... )` blocks | State machine: track `import (` open, extract until `)` |
| Python | `import x`, `from x import y`, `from x import (...)` | State machine for parenthesized imports |
| JS/TS | `import ... from 'path'`, `import 'path'`, `require('path')` | State machine for multi-line destructured imports |
| Rust | `use x::y`, `use x::{...}`, `mod x` | State machine for multi-line `use` blocks |

The extractor (`extract.go`) currently uses a line-by-line scanner. Import extraction adds a parallel state machine that tracks whether we're inside a multi-line import block. This is scoped to `ExtractSymbols` — we return both symbols and imports from the same file pass.

**New query functions:**
- `QueryImports(idx *Index, file string) []Import` — "what does file X import?" Filter by file path (substring match to handle relative paths).
- `QueryImporters(idx *Index, importPath string) []string` — "who imports package X?" Returns list of file paths. Substring match on import path.

**New MCP tools:**
- `rpi_index_imports` — input: `{file: string}` — returns imports for a file
- `rpi_index_importers` — input: `{import_path: string}` — returns files that import a given path

### 4. Improved MCP Tool Descriptions (all index tools)

Replace mechanical descriptions with action-oriented ones that tell the LLM when to prefer the index over Grep/Glob. Use `mcpDescriptionWithPrefix` for all index tools (currently only `rpi_index_status` uses `mcpDescription`).

| Tool | Prefix |
|---|---|
| `rpi_index_query` | "Find where functions, classes, structs, and interfaces are defined — not just mentioned. Unlike grep, returns only definitions with file, line, kind, and export status. Prefer this when locating a definition or surveying what exists." |
| `rpi_index_files` | "Get a compact structural map of the codebase: files grouped by language with symbol counts. Faster than directory listings for understanding codebase shape." |
| `rpi_index_status` | "Quick orientation: how big is this codebase, what languages, how many symbols. Use early in exploration before diving into files." |
| `rpi_index_build` | "Rebuild the symbol index when stale or missing." |
| `rpi_index_packages` | "Package-level overview: what does each package export? Shows file count, symbol counts by kind. Use to understand package responsibilities before reading files." |
| `rpi_index_imports` | "What does a file depend on? Returns all import/require/use statements with paths and aliases." |
| `rpi_index_importers` | "Who depends on a package or module? Find all files that import a given path. Use to assess blast radius of changes." |

## File Structure

**Modified files:**
- `internal/index/index.go` — add `Import` struct, add `Imports` field to `Index`, update `Build` to collect imports
- `internal/index/extract.go` — add `ExtractImports` or extend `ExtractSymbols` to return imports, add multi-line state machine
- `internal/index/languages.go` — add import patterns per language
- `internal/index/query.go` — add `Signature`/`Package` to `QueryOptions`, add `QueryPackages`, `QueryImports`, `QueryImporters`
- `internal/index/store.go` — bump `CurrentVersion`
- `cmd/rpi/index.go` — add `--signature`/`--package` flags, add `packages` subcommand
- `cmd/rpi/serve.go` — add new input types, register 3 new tools, update descriptions for all index tools

**New test files or additions:**
- `internal/index/extract_test.go` — import extraction tests for all 4 languages, including multi-line blocks
- `internal/index/query_test.go` — signature search, package filter, package overview, import/importer queries
- `cmd/rpi/index_test.go` — CLI tests for new flags/subcommands
- `cmd/rpi/serve_test.go` — MCP tool registration for new tools

## Risks

1. **Multi-line import parsing complexity** — The line-by-line scanner needs a state machine for import blocks. Risk: edge cases in unusual formatting. Mitigation: comprehensive test cases per language, graceful fallback (skip malformed imports rather than crash).

2. **Index size growth** — Adding imports increases `index.json` size. For a typical project this is modest (imports are small strings), but large monorepos could see meaningful growth. Mitigation: monitor size in `rpi_index_status` output.

3. **Version migration** — Old indexes won't have `Imports`. The existing version check in `store.go` already handles this cleanly ("run `rpi index build` to rebuild"). No migration needed.

4. **Import path normalization** — Go uses full module paths (`github.com/user/repo/pkg`), Python uses dotted names (`pkg.mod`), JS uses relative/bare specifiers. `QueryImporters` uses substring match, which handles this reasonably but may return false positives for short patterns. Mitigation: document that patterns should be specific enough to avoid ambiguity.

## Out of Scope

- **Call graph / reference tracking** — requires AST parsing to distinguish calls from type references. Not tractable with regex.
- **Method-to-receiver mapping queries** — the data partially exists (Go receiver is in signature, Python/TS methods have scope), but a dedicated query tool is deferred.
- **Dynamic imports** — `import()` in JS/TS, `importlib` in Python. These are runtime constructs and not reliably extractable with regex.
- **Conditional imports** — `#[cfg()]` in Rust, `if TYPE_CHECKING:` in Python. Treated as regular imports.
- **Re-exports** — `export { x } from 'y'` in JS/TS. Could be added later as a variant of Import.

## References

- Current index implementation: `internal/index/` (6 source files)
- MCP tool registration: `cmd/rpi/serve.go:169-183`
- CLI commands: `cmd/rpi/index.go`
- Architecture doc: `docs/architecture.md:15`
