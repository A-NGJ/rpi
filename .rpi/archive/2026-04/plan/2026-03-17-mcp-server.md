---
archived_date: "2026-04-02"
date: 2026-03-17T23:52:31+01:00
proposal: .rpi/proposals/2026-03-17-mcp-server.md
spec: .rpi/specs/mcp-server.md
status: archived
tags:
    - plan
topic: mcp-server
---

# MCP Server — Implementation Plan

## Overview

Add `rpi serve` subcommand exposing all CLI operations as typed MCP tools over stdio. Uses the official Go MCP SDK (`modelcontextprotocol/go-sdk`).

**Scope**: 1 file modified (`main.go`), 2 new files (`serve.go`, `serve_test.go`)

## Source Documents
- **Proposal**: .rpi/proposals/2026-03-17-mcp-server.md
- **Spec**: .rpi/specs/mcp-server.md

## Phase 1: Server Skeleton + No-Param Tools

### Overview
Add MCP SDK dependency, create `serve.go` with the `rpi serve` command skeleton, register 5 tools that take no input parameters. Write tests validating server setup and tool output.
**Spec behaviors**: MS-1, MS-2, MS-4, MS-5, MS-6, MS-8, MS-23, MS-24, MS-25, MS-26, MS-27

### Tasks:

#### 1. Add MCP SDK dependency
**Changes**: `go get github.com/modelcontextprotocol/go-sdk`

#### 2. Create serve command with server setup
**File**: `cmd/rpi/serve.go`
**Changes**:
- Define `serveCmd` Cobra command (`rpi serve`)
- In `runServe`: create `mcp.NewServer` with `Implementation{Name: "rpi"}`, run on `mcp.StdioTransport{}`
- Register in `init()` to `rootCmd`
- Helper function `jsonResult(v any) (*mcp.CallToolResult, any, error)` that marshals to JSON and wraps in `TextContent`

#### 3. Register 5 no-param tools
**File**: `cmd/rpi/serve.go`
**Changes**:
- `rpi_git_context` — calls `git.GatherContext()` (MS-24)
- `rpi_git_changed_files` — calls `git.ChangedFiles()` (MS-25)
- `rpi_git_sensitive_check` — calls `git.SensitiveCheck()` (MS-26)
- `rpi_index_status` — calls `index.Load()` + `index.Status()` (MS-23)
- `rpi_archive_scan` — calls `scanner.Scan()` with `Archivable: true`, counts refs (MS-27)

Each tool: empty input struct (`struct{}`), handler calling internal package, JSON result via helper.

#### 4. Tests for Phase 1
**File**: `cmd/rpi/serve_test.go`
**Changes**:
- `TestServeCmd_Exists` — verify `serveCmd` is registered on `rootCmd`
- `TestNewRPIServer_ToolCount` — extract server creation into a `newRPIServer()` function, call `tools/list` via MCP client, verify 5 tools present with `rpi_` prefix
- `TestToolHandler_GitContext` — call `rpi_git_context` tool, verify JSON response contains expected fields
- `TestToolHandler_ArchiveScan` — set up temp `.rpi/` dir with test artifacts, call `rpi_archive_scan`, verify JSON response

### Success Criteria:

#### Automated Verification:
- [x] `go test ./cmd/rpi/ -run TestServe` passes
- [x] `go build ./cmd/rpi` succeeds
- [x] `go vet ./...` clean

#### Manual Verification:
- [x] `rpi serve` starts and blocks (ctrl-c to exit)
- [x] Existing `rpi` commands still work (`rpi scan`, `rpi --help`)

### Commit:
- [x] Stage: `go.mod`, `go.sum`, `cmd/rpi/serve.go`, `cmd/rpi/serve_test.go`
- [x] Message: `feat(mcp): add rpi serve command with initial no-param tools`

**Note**: Pause for manual confirmation before proceeding to next phase.

---

## Phase 2: Parameterized Tools

### Overview
Register all remaining 16 tools with typed input structs. Write tests for representative tools covering required params, optional params, and error paths.
**Spec behaviors**: MS-3, MS-7, MS-9, MS-10 through MS-22, MS-28, MS-29, MS-30

### Tasks:

#### 1. Artifact tools (scan, scaffold, frontmatter, chain, extract)
**File**: `cmd/rpi/serve.go`
**Changes**:
- `rpi_scan` — input: `{Type, Status, Proposal, References string; Archivable bool}` → `scanner.Scan()` (MS-10)
- `rpi_scaffold` — input: `{Type, Topic, Ticket, Research, Proposal, Spec, Tags string; Write, Force bool}` → `template.RenderTemplate()` / `writeOutput()` (MS-11)
- `rpi_frontmatter_get` — input: `{File, Field string}` → `frontmatter.Parse()` (MS-12)
- `rpi_frontmatter_set` — input: `{File, Field, Value string}` → `frontmatter.Parse()` + `Write()` (MS-13)
- `rpi_frontmatter_transition` — input: `{File, Status string}` → `frontmatter.Transition()` + `Write()` (MS-14)
- `rpi_chain` — input: `{Path, Sections string}` → `chain.Resolve()` (MS-15)
- `rpi_extract` — input: `{Path, Section string}` → `frontmatter.ExtractSection()` (MS-16)
- `rpi_extract_list_sections` — input: `{Path string}` → `frontmatter.ListSections()` (MS-17)

#### 2. Verification + index tools
**File**: `cmd/rpi/serve.go`
**Changes**:
- `rpi_verify_completeness` — input: `{PlanPath string}` → reuse checkbox/file parsing logic (MS-18)
- `rpi_verify_markers` — input: `{FilePath string}` → reuse marker scanning logic (MS-19)
- `rpi_index_build` — input: `{Lang, Path string; Force bool}` → `index.Build()` + `index.Save()` (MS-20)
- `rpi_index_query` — input: `{Pattern, Kind string; Exported bool}` → `index.QuerySymbols()` (MS-21)
- `rpi_index_files` — input: `{Lang string}` → `index.QueryFiles()` (MS-22)

#### 3. Archive + spec tools
**File**: `cmd/rpi/serve.go`
**Changes**:
- `rpi_archive_check_refs` — input: `{Path string}` → `scanner.FindReferences()` (MS-28)
- `rpi_archive_move` — input: `{Path string; Force bool}` → `doArchiveMove()` (MS-29)
- `rpi_spec_coverage` — input: `{SpecFile string}` → `spec.ParseBehaviors()` + `spec.ComputeCoverage()` (MS-30)

#### 4. Tests for Phase 2
**File**: `cmd/rpi/serve_test.go`
**Changes**:
- `TestToolHandler_Scan` — set up temp `.rpi/` with artifacts, call `rpi_scan` with type filter, verify filtered results (MS-10)
- `TestToolHandler_FrontmatterGet` — create temp artifact, call `rpi_frontmatter_get`, verify JSON matches (MS-12)
- `TestToolHandler_FrontmatterTransition_Invalid` — create draft artifact, call transition to `complete`, verify MCP error (MS-9, MS-14)
- `TestToolHandler_Scaffold` — call `rpi_scaffold` with `write: true`, verify file created and path returned (MS-11)
- `TestToolHandler_Extract` — create temp artifact with sections, call `rpi_extract`, verify content (MS-16)
- `TestToolHandler_VerifyCompleteness` — create temp plan with checkboxes, verify counts (MS-18)

### Success Criteria:

#### Automated Verification:
- [x] `go test ./cmd/rpi/ -run TestTool` passes
- [x] `go build ./cmd/rpi` succeeds
- [x] `go vet ./...` clean

#### Manual Verification:
- [x] All 21 tools visible (20 from spec + confirm count)
- [x] Existing CLI still works

### Commit:
- [x] Stage: `cmd/rpi/serve.go`, `cmd/rpi/serve_test.go`
- [x] Message: `feat(mcp): register all 20 tools with typed input schemas`

**Note**: Pause for manual confirmation before proceeding to next phase.

---

## Phase 3: Integration Test + main.go Wiring

### Overview
Wire `serveCmd` into `main.go`. Add integration test that starts the server process, sends `tools/list`, and verifies all 20 tools are present with correct names.
**Spec behaviors**: MS-3 (e2e)

### Tasks:

#### 1. Wire serve command into main
**File**: `cmd/rpi/main.go`
**Changes**: Already wired via `init()` in `serve.go` — verify `rpi serve --help` works from the built binary.

#### 2. Integration test
**File**: `cmd/rpi/serve_test.go`
**Changes**:
- `TestIntegration_AllToolsRegistered` — create MCP server via `newRPIServer()`, connect with MCP client, call `tools/list`, assert exactly 20 tools, assert all names match expected list, assert all have non-empty descriptions (MS-3, MS-4, MS-5)

### Success Criteria:

#### Automated Verification:
- [x] `go test ./...` all pass
- [x] `go build ./cmd/rpi` succeeds
- [x] `go vet ./...` clean

#### Manual Verification:
- [x] `rpi serve --help` shows description
- [x] `rpi --help` lists `serve` subcommand

### Commit:
- [x] Stage: `cmd/rpi/serve_test.go`
- [x] Message: `test(mcp): add integration test verifying all 21 tools registered`

---

## References
- Proposal: .rpi/proposals/2026-03-17-mcp-server.md
- Spec: .rpi/specs/mcp-server.md
- Official MCP Go SDK: https://github.com/modelcontextprotocol/go-sdk
