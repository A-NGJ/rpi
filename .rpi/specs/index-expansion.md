---
domain: index expansion
id: IX
last_updated: 2026-04-02T01:39:57+02:00
status: complete
updated_by: .rpi/designs/2026-04-02-index-expansion-signature-search-package-overview-import-graph.md
---

# Index Expansion

## Purpose

Expand the codebase index with signature-based symbol search, package-level overview queries, an import graph, and improved MCP tool descriptions — making the index a genuine accelerator for LLM-driven codebase discovery.

## Behavior

### Signature Search
- **IX-1**: `QuerySymbols` accepts an optional `Signature` filter that performs case-insensitive substring match on `Symbol.Signature`
- **IX-2**: The `signature` filter composes with existing `pattern`, `kind`, and `exported` filters (all must match)
- **IX-3**: The `rpi_index_query` MCP tool accepts an optional `signature` parameter
- **IX-4**: The `rpi index query` CLI accepts a `--signature` flag

### Package Filter
- **IX-5**: `QuerySymbols` accepts an optional `Package` filter that performs case-insensitive substring match on `Symbol.Package`
- **IX-6**: The `rpi_index_query` MCP tool accepts an optional `package` parameter
- **IX-7**: The `rpi index query` CLI accepts a `--package` flag

### Package Overview
- **IX-8**: `QueryPackages` returns a list of package summaries, each containing: package name, file paths, file count, exported symbol count, total symbol count, and symbol kind breakdown
- **IX-9**: `QueryPackages` accepts an optional package name filter (substring match)
- **IX-10**: A new `rpi_index_packages` MCP tool exposes `QueryPackages` with optional `package` filter parameter
- **IX-11**: A new `rpi index packages` CLI subcommand exposes `QueryPackages` with `--package` and `--format` flags

### Import Extraction
- **IX-12**: `Build` extracts imports for Go (`import "path"`, `import (...)` blocks)
- **IX-13**: `Build` extracts imports for Python (`import x`, `from x import y`, `from x import (...)`)
- **IX-14**: `Build` extracts imports for JavaScript/TypeScript (`import ... from 'path'`, `import 'path'`, `require('path')`)
- **IX-15**: `Build` extracts imports for Rust (`use x::y`, `use x::{...}`, `mod x`)
- **IX-16**: Multi-line import blocks are handled correctly for all languages (parenthesized Go imports, Python from-imports, JS destructured imports, Rust use blocks)
- **IX-17**: Each import records: source file, import path, alias (empty if none), and line number
- **IX-18**: The `Index` struct includes an `Imports []Import` field
- **IX-19**: The index version is bumped; loading an old-version index returns an error directing the user to rebuild

### Import Queries
- **IX-20**: `QueryImports` returns all imports for a given file (substring match on file path)
- **IX-21**: `QueryImporters` returns all file paths that import a given path (substring match on import path)
- **IX-22**: A new `rpi_index_imports` MCP tool exposes `QueryImports` with a required `file` parameter
- **IX-23**: A new `rpi_index_importers` MCP tool exposes `QueryImporters` with a required `import_path` parameter
- **IX-24**: New `rpi index imports` and `rpi index importers` CLI subcommands expose these queries

### MCP Tool Descriptions
- **IX-25**: All index MCP tool descriptions use `mcpDescriptionWithPrefix` with action-oriented prefixes that communicate when to prefer the tool over Grep/Glob
- **IX-26**: New tools (`rpi_index_packages`, `rpi_index_imports`, `rpi_index_importers`) have descriptions that explain their unique value

## Constraints

### Must
- All import extraction must work with the existing line-by-line scanner (no AST parsing)
- All 4 languages (Go, Python, JS/TS, Rust) supported for every feature from day one
- Malformed or unparseable imports are silently skipped (no crash)
- Index version bump ensures old indexes fail cleanly with a rebuild message
- New MCP tools must be registered in `serve_test.go` tool list assertions

### Must Not
- Must not break existing `QuerySymbols` behavior when new filters are omitted (empty string = no filter)
- Must not add external dependencies
- Must not attempt to resolve dynamic imports (`import()`, `importlib`)

### Out of Scope
- Call graph / reference tracking
- Method-to-receiver mapping queries
- Dynamic imports, conditional imports, re-exports
- Import path resolution or normalization beyond raw string capture

## Test Cases

### IX-1: Signature filter matches
- **Given** an index with a symbol `Build` having signature `func Build(rootPath string, opts BuildOptions) (*Index, error)` **When** `QuerySymbols` is called with `Signature: "BuildOptions"` **Then** `Build` is returned

### IX-2: Signature + name compose
- **Given** the same index **When** `QuerySymbols` is called with `Pattern: "build", Signature: "BuildOptions"` **Then** `Build` is returned
- **Given** the same index **When** `QuerySymbols` is called with `Pattern: "query", Signature: "BuildOptions"` **Then** no results

### IX-5: Package filter
- **Given** an index with symbols in packages `index` and `main` **When** `QuerySymbols` is called with `Package: "index"` **Then** only symbols from package `index` are returned

### IX-8: Package overview aggregation
- **Given** an index with 3 files in package `index` containing 5 exported functions and 2 structs **When** `QueryPackages` is called **Then** the result includes a summary for `index` with `file_count: 3`, `exported_symbols: 7`, `kinds: {"function": 5, "struct": 2, ...}`

### IX-12: Go single-line import
- **Given** a Go file containing `import "fmt"` **When** imports are extracted **Then** an import with `import_path: "fmt"` and empty alias is returned

### IX-12: Go block import
- **Given** a Go file containing `import (\n\t"fmt"\n\t"os"\n)` **When** imports are extracted **Then** two imports are returned for `fmt` and `os`

### IX-12: Go aliased import
- **Given** a Go file containing `import f "fmt"` **When** imports are extracted **Then** an import with `import_path: "fmt"` and `alias: "f"` is returned

### IX-13: Python import
- **Given** a Python file containing `import os` **When** imports are extracted **Then** an import with `import_path: "os"` is returned

### IX-13: Python from-import
- **Given** a Python file containing `from os.path import join` **When** imports are extracted **Then** an import with `import_path: "os.path"` is returned

### IX-14: JS/TS import from
- **Given** a TS file containing `import { useState } from 'react'` **When** imports are extracted **Then** an import with `import_path: "react"` is returned

### IX-14: JS require
- **Given** a JS file containing `const fs = require('fs')` **When** imports are extracted **Then** an import with `import_path: "fs"` is returned

### IX-15: Rust use
- **Given** a Rust file containing `use std::collections::HashMap;` **When** imports are extracted **Then** an import with `import_path: "std::collections::HashMap"` is returned

### IX-16: Multi-line Go import block
- **Given** a Go file with `import (\n\t"fmt"\n\t"os"\n\tlog "github.com/sirupsen/logrus"\n)` **When** imports are extracted **Then** 3 imports are returned, the third having `alias: "log"`

### IX-20: QueryImports
- **Given** an index where `cmd/rpi/serve.go` has 5 imports **When** `QueryImports` is called with `file: "serve.go"` **Then** all 5 imports are returned (substring match)

### IX-21: QueryImporters
- **Given** an index where 3 files import `internal/index` **When** `QueryImporters` is called with `import_path: "internal/index"` **Then** those 3 file paths are returned

### IX-25: Tool descriptions
- **Given** the MCP server is running **When** tool descriptions are inspected **Then** all 7 index tools use `mcpDescriptionWithPrefix` with action-oriented prefixes
