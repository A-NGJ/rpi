---
domain: index expansion
feature: index-expansion
last_updated: 2026-04-04T22:30:00+02:00
updated_by: .rpi/designs/2026-04-02-index-expansion-signature-search-package-overview-import-graph.md
---

# Index Expansion

## Purpose

Expand the codebase index with signature-based symbol search, package-level overview queries, and an import graph — making the index a genuine accelerator for LLM-driven codebase discovery.

## Scenarios

### Symbol query supports signature filtering
Given an index with symbols that have signature metadata
When querying with a signature filter
Then only symbols whose signature contains the filter substring (case-insensitive) are returned, and the filter composes with existing name/kind/exported filters

### Symbol query supports package filtering
Given an index with symbols across multiple packages
When querying with a package filter
Then only symbols from matching packages are returned

### Package overview returns aggregated summaries
Given an index built from a multi-package project
When querying for package summaries
Then each result includes the package name, file count, exported/total symbol counts, and kind breakdown

### Imports extracted for Go, Python, JS/TS, and Rust
Given source files in each supported language
When the index is built
Then imports are extracted from all four languages including single-line and block import syntax

### Multi-line import blocks handled correctly
Given source files with parenthesized or multi-line import blocks
When the index is built
Then all imports within the block are extracted with correct paths and aliases

### Import queries find files and importers
Given a built index with import data
When querying imports by file or by import path
Then the correct set of imports or importing files is returned via substring match

### Index version bump forces rebuild on old indexes
Given an index built with a previous version
When loading the index
Then an error directs the user to rebuild the index

## Constraints
- All import extraction uses the existing line-by-line scanner (no AST parsing)
- All 4 languages supported for every feature from day one
- Malformed imports are silently skipped
- Do not break existing query behavior when new filters are omitted
- Do not add external dependencies or resolve dynamic imports

## Out of Scope
- Call graph or reference tracking
- Method-to-receiver mapping queries
- Dynamic imports, conditional imports, re-exports
- Import path resolution or normalization
