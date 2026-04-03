---
date: 2026-04-02T01:43:59+02:00
design: .rpi/designs/2026-04-02-index-expansion-signature-search-package-overview-import-graph.md
spec: .rpi/specs/index-expansion.md
status: complete
tags:
    - plan
topic: index expansion — signature search, package overview, import graph
---

# index expansion — signature search, package overview, import graph — Implementation Plan

## Overview

Expand the codebase index with three capabilities: signature-based symbol search, package-level overview queries, and an import graph. Also improve all index MCP tool descriptions to be action-oriented.

**Scope**: 7 files modified, 0 new files

## Source Documents
- **Design**: .rpi/designs/2026-04-02-index-expansion-signature-search-package-overview-import-graph.md
- **Spec**: .rpi/specs/index-expansion.md

## Phase 1: Signature + Package Filters on Existing Query

### Overview
Add `Signature` and `Package` filters to `QueryOptions` and wire them through the CLI and MCP tool. This extends the existing query infrastructure — no new tools or data structures.

**Spec coverage**: IX-1, IX-2, IX-3, IX-4, IX-5, IX-6, IX-7

### Tasks:

#### 1. Query logic
**File**: `internal/index/query.go`
**Changes**:
- Add `Signature string` and `Package string` fields to `QueryOptions`
- Add two filter clauses in `QuerySymbols`: case-insensitive substring match on `Symbol.Signature` (if `Signature != ""`) and on `Symbol.Package` (if `Package != ""`)

#### 2. CLI flags
**File**: `cmd/rpi/index.go`
**Changes**:
- Add `indexSignatureFlag` and `indexPackageFlag` vars
- Register `--signature` and `--package` flags on `indexQueryCmd`
- Pass them through to `QueryOptions` in `runIndexQuery`

#### 3. MCP tool input
**File**: `cmd/rpi/serve.go`
**Changes**:
- Add `Signature string` and `Package string` fields to `indexQueryInput` with jsonschema descriptions
- Pass them through to `QueryOptions` in `handleIndexQuery`

#### 4. Tests
**File**: `internal/index/query_test.go`
**Changes**:
- Update `sampleIndex()` to include `Signature` and `Package` fields on symbols
- Add `TestQuerySymbolsSignatureFilter` — matches substring in signature
- Add `TestQuerySymbolsSignatureAndPatternCompose` — both must match
- Add `TestQuerySymbolsPackageFilter` — filters by package name

### Success Criteria:

#### Automated Verification:
- [x] `go test ./internal/index/... -run TestQuerySymbolsSignature` passes
- [x] `go test ./internal/index/... -run TestQuerySymbolsPackage` passes
- [x] `go vet ./...` clean

### Commit:
- [x] Stage: `internal/index/query.go`, `internal/index/query_test.go`, `cmd/rpi/index.go`, `cmd/rpi/serve.go`
- [x] Message: `feat(index): add signature and package filters to symbol query`

---

## Phase 2: Package Overview

### Overview
Add a `QueryPackages` function that aggregates existing symbol data into package-level summaries, expose it as a new MCP tool (`rpi_index_packages`) and CLI subcommand (`rpi index packages`).

**Spec coverage**: IX-8, IX-9, IX-10, IX-11

### Tasks:

#### 1. Query function
**File**: `internal/index/query.go`
**Changes**:
- Add `PackageSummary` struct with fields: `Name`, `Files` ([]string), `FileCount`, `ExportedSymbols`, `TotalSymbols`, `Kinds` (map[string]int)
- Add `QueryPackages(idx *Index, pkg string) []PackageSummary` — iterates symbols, groups by package, optionally filters by package name (case-insensitive substring)

#### 2. CLI subcommand
**File**: `cmd/rpi/index.go`
**Changes**:
- Add `indexPackagesCmd` cobra command (`rpi index packages`)
- Add `--package` flag for optional filtering
- Add `--format` flag (json/md)
- Register under `indexCmd`
- Implement `runIndexPackages`

#### 3. MCP tool
**File**: `cmd/rpi/serve.go`
**Changes**:
- Add `indexPackagesInput` struct with optional `Package` field
- Add `handleIndexPackages` handler calling `index.QueryPackages`
- Register `rpi_index_packages` tool in `registerTools`

#### 4. Tests
**Files**: `internal/index/query_test.go`, `cmd/rpi/serve_test.go`
**Changes**:
- Add `TestQueryPackages` — verifies aggregation (file count, exported count, kinds)
- Add `TestQueryPackagesFilter` — verifies package name filter
- Update `expectedTools` in `serve_test.go` to include `rpi_index_packages`

### Success Criteria:

#### Automated Verification:
- [x] `go test ./internal/index/... -run TestQueryPackages` passes
- [x] `go test ./cmd/rpi/... -run TestIntegration_AllToolsRegistered` passes (21 tools)
- [x] `go vet ./...` clean

### Commit:
- [x] Stage: `internal/index/query.go`, `internal/index/query_test.go`, `cmd/rpi/index.go`, `cmd/rpi/serve.go`, `cmd/rpi/serve_test.go`
- [x] Message: `feat(index): add package overview query and MCP tool`

---

## Phase 3: Import Extraction

### Overview
Add the `Import` struct, bump the index version, and implement multi-line import extraction for all 4 languages. This is the heaviest phase — the state machine for multi-line import blocks is the main complexity.

**Spec coverage**: IX-12, IX-13, IX-14, IX-15, IX-16, IX-17, IX-18, IX-19

### Tasks:

#### 1. Data structures + version bump
**File**: `internal/index/index.go`
**Changes**:
- Add `Import` struct with fields: `File string`, `ImportPath string`, `Alias string`, `Line int`
- Add `Imports []Import` field to `Index` struct

**File**: `internal/index/store.go`
**Changes**:
- Bump `CurrentVersion` from `"1"` to `"2"`

#### 2. Import extraction
**File**: `internal/index/extract.go`
**Changes**:
- Add `ExtractImports(filePath string, lang string) ([]Import, error)` function
- Implement per-language import regex patterns and multi-line state machine:
  - **Go**: `import "path"` single-line, `import ( ... )` blocks with optional alias
  - **Python**: `import x`, `from x import y`, `from x import (...)` multi-line
  - **JS/TS**: `import ... from 'path'`, `import 'path'`, `require('path')`, multi-line destructured imports
  - **Rust**: `use x::y;`, `use x::{...}` multi-line blocks, `mod x;`
- State machine tracks whether scanner is inside a multi-line import block; extracts each import path + alias + line number
- Malformed imports are silently skipped

#### 3. Wire into Build
**File**: `internal/index/index.go`
**Changes**:
- In `Build`, call `ExtractImports` after `ExtractSymbols` for each file
- Rewrite `Import.File` to relative path
- Collect into `idx.Imports`

#### 4. Tests
**File**: `internal/index/extract_test.go`
**Changes**:
- Add `TestExtractImportsGo` — single-line, block, aliased imports
- Add `TestExtractImportsPython` — `import x`, `from x import y`, multi-line `from x import (...)`
- Add `TestExtractImportsJavaScript` — `import from`, `import 'path'`, `require()`, multi-line destructured
- Add `TestExtractImportsTypeScript` — same patterns as JS (shared lang handling)
- Add `TestExtractImportsRust` — `use`, multi-line `use {}`, `mod`
- Add `TestExtractImportsMultilineGoBlock` — specifically tests the state machine with aliases and blank lines
- Each test writes a temp file, calls `ExtractImports`, asserts import count, paths, aliases, and line numbers

### Success Criteria:

#### Automated Verification:
- [x] `go test ./internal/index/... -run TestExtractImports` passes (all language tests)
- [x] `go test ./internal/index/... -run TestBuild` passes (if exists) — Build still works end-to-end
- [x] `go vet ./...` clean
- [x] Existing `TestExtract*` symbol tests still pass (no regressions)

### Commit:
- [x] Stage: `internal/index/index.go`, `internal/index/store.go`, `internal/index/extract.go`, `internal/index/extract_test.go`
- [x] Message: `feat(index): extract imports for Go, Python, JS/TS, and Rust with multi-line support`

---

## Phase 4: Import Queries + MCP Tools

### Overview
Add query functions for imports and importers, expose them as MCP tools and CLI subcommands.

**Spec coverage**: IX-20, IX-21, IX-22, IX-23, IX-24

### Tasks:

#### 1. Query functions
**File**: `internal/index/query.go`
**Changes**:
- Add `QueryImports(idx *Index, file string) []Import` — returns imports where `Import.File` contains `file` (case-insensitive substring)
- Add `QueryImporters(idx *Index, importPath string) []string` — returns deduplicated file paths where `Import.ImportPath` contains `importPath` (case-insensitive substring)

#### 2. CLI subcommands
**File**: `cmd/rpi/index.go`
**Changes**:
- Add `indexImportsCmd` (`rpi index imports <file>`) — takes a file path, prints imports
- Add `indexImportersCmd` (`rpi index importers <import_path>`) — takes an import path, prints files
- Both support `--format json|md`
- Register under `indexCmd`

#### 3. MCP tools
**File**: `cmd/rpi/serve.go`
**Changes**:
- Add `indexImportsInput` struct with required `File` field
- Add `indexImportersInput` struct with required `ImportPath` field
- Add `handleIndexImports` and `handleIndexImporters` handlers
- Register `rpi_index_imports` and `rpi_index_importers` tools in `registerTools`

#### 4. Tests
**Files**: `internal/index/query_test.go`, `cmd/rpi/serve_test.go`
**Changes**:
- Update `sampleIndex()` to include `Imports` data
- Add `TestQueryImports` — substring match on file path
- Add `TestQueryImporters` — substring match on import path, returns file list
- Update `expectedTools` in `serve_test.go` to include `rpi_index_imports` and `rpi_index_importers` (23 tools total)

### Success Criteria:

#### Automated Verification:
- [x] `go test ./internal/index/... -run TestQueryImport` passes
- [x] `go test ./cmd/rpi/... -run TestToolRegistration` passes (23 tools)
- [x] `go vet ./...` clean

### Commit:
- [x] Stage: `internal/index/query.go`, `internal/index/query_test.go`, `cmd/rpi/index.go`, `cmd/rpi/serve.go`, `cmd/rpi/serve_test.go`
- [x] Message: `feat(index): add import and importer queries with MCP tools`

---

## Phase 5: MCP Tool Descriptions

### Overview
Update all index MCP tool descriptions to use `mcpDescriptionWithPrefix` with action-oriented prefixes that tell the LLM when to prefer the index over Grep/Glob.

**Spec coverage**: IX-25, IX-26

### Tasks:

#### 1. Update tool registrations
**File**: `cmd/rpi/serve.go`
**Changes**:
- Change `rpi_index_query` to use `mcpDescriptionWithPrefix` with: "Find where functions, classes, structs, and interfaces are defined — not just mentioned. Unlike grep, returns only definitions with file, line, kind, and export status. Prefer this when locating a definition or surveying what exists."
- Change `rpi_index_files` to use `mcpDescriptionWithPrefix` with: "Get a compact structural map of the codebase: files grouped by language with symbol counts. Faster than directory listings for understanding codebase shape."
- Change `rpi_index_status` to use `mcpDescriptionWithPrefix` with: "Quick orientation: how big is this codebase, what languages, how many symbols. Use early in exploration before diving into files."
- Change `rpi_index_build` to use `mcpDescriptionWithPrefix` with: "Rebuild the symbol index when stale or missing."
- Verify `rpi_index_packages`, `rpi_index_imports`, `rpi_index_importers` already have prefixes from Phases 2/4 (update if needed)

### Success Criteria:

#### Automated Verification:
- [x] `go test ./cmd/rpi/... -run TestToolRegistration` passes
- [x] `go vet ./...` clean

#### Manual Verification:
- [x] Inspect tool descriptions in `serve.go` — all 7 index tools use `mcpDescriptionWithPrefix`

### Commit:
- [x] Stage: `cmd/rpi/serve.go`
- [x] Message: `feat(index): improve MCP tool descriptions with action-oriented prefixes`

---

## References
- Design: .rpi/designs/2026-04-02-index-expansion-signature-search-package-overview-import-graph.md
- Spec: .rpi/specs/index-expansion.md
