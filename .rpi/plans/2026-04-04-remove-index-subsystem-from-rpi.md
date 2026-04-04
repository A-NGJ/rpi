---
date: 2026-04-04T21:12:30+02:00
design: .rpi/designs/2026-04-04-remove-index-subsystem-from-rpi.md
spec: .rpi/specs/index-removal.md
status: complete
tags:
    - plan
topic: remove index subsystem from rpi
---

# Remove Index Subsystem from RPI — Implementation Plan

## Overview

Remove the entire index subsystem: MCP tools, CLI commands, Go package, and documentation references.

**Scope**: 12 files deleted, 6 files modified

## Source Documents
- **Design**: .rpi/designs/2026-04-04-remove-index-subsystem-from-rpi.md
- **Spec**: .rpi/specs/index-removal.md

## Phase 1: Strip index from MCP server [IR-3, IR-4, IR-5]

### Overview
Remove all index-related code from the MCP server and its tests. After this phase, the server compiles and passes tests without any index dependency.

### Tasks:

#### 1. Remove index tool registrations and handlers from serve.go
**File**: `cmd/rpi/serve.go`
**Changes**:
- Remove `"github.com/A-NGJ/rpi/internal/index"` import (line 15)
- Remove `"os"` import if no longer used after removing handleIndexStatus
- Remove `rpi_index_status` tool registration (lines 102-106)
- Remove `rpi_index_build` through `rpi_index_importers` tool registrations (lines 169-198)
- Remove `handleIndexStatus` handler (lines 238-249)
- Remove `handleIndexBuild` through `handleIndexImporters` handlers (lines 572-663)
- Remove input structs: `indexBuildInput`, `indexQueryInput`, `indexFilesInput`, `indexPackagesInput`, `indexImportsInput`, `indexImportersInput` (lines 338-366)

#### 2. Update serve_test.go
**File**: `cmd/rpi/serve_test.go`
**Changes**:
- Remove `TestHandleIndexStatus_ReturnsJSON` test (lines 79-89)
- Remove 7 index tools from `expectedTools` in `TestIntegration_AllToolsRegistered` (lines 403, 416-421)

### Success Criteria:

#### Automated Verification:
- [x] `go build ./cmd/rpi/` succeeds
- [x] `go test ./cmd/rpi/ -count=1` passes

### Commit:
- [x] Stage: `cmd/rpi/serve.go`, `cmd/rpi/serve_test.go`
- [x] Message: `refactor(serve): remove index MCP tools and handlers`

---

## Phase 2: Delete index package and CLI command [IR-1, IR-2, IR-6]

### Overview
Delete the `internal/index/` package and the CLI `rpi index` command files.

### Tasks:

#### 1. Delete index package
**Directory**: `internal/index/`
**Changes**: Delete entire directory (8 files: `index.go`, `store.go`, `query.go`, `extract.go`, `languages.go`, `index_test.go`, `extract_test.go`, `query_test.go`)

#### 2. Delete CLI command
**Files**: `cmd/rpi/index.go`, `cmd/rpi/index_test.go`
**Changes**: Delete both files

### Success Criteria:

#### Automated Verification:
- [x] `go build ./cmd/rpi/` succeeds
- [x] `go vet ./...` clean
- [x] `internal/index/` directory does not exist
- [x] `cmd/rpi/index.go` and `cmd/rpi/index_test.go` do not exist

### Commit:
- [x] Stage: deleted files + `cmd/rpi/sync.go` (also had index dependency)
- [x] Message: `refactor: delete index package and CLI command`

---

## Phase 3: Clean up init and documentation [IR-8, IR-9]

### Overview
Remove index references from init command and all documentation.

### Tasks:

#### 1. Remove index.json gitignore from init
**File**: `cmd/rpi/init_cmd.go`
**Changes**:
- Remove lines 181-184 (the `ensureGitignoreEntry` call for `.rpi/index.json`)
- Remove index references from the `Long` description (lines 56-57)

#### 2. Update README.md
**File**: `README.md`
**Changes**:
- Remove line 60: `.rpi/index.json` -- Codebase symbol index (gitignored)
- Update line 66: remove "rebuild index" from `rpi update` description

#### 3. Update architecture.md
**File**: `docs/architecture.md`
**Changes**:
- Remove "Codebase indexing" bullet (line 15)
- Remove `internal/index/` from project structure (line 29)

#### 4. Update rpi-init.md
**File**: `docs/rpi-init.md`
**Changes**:
- Remove line 49: `.rpi/index.json` -- Codebase symbol index

#### 5. Update thoughts-directory.md
**File**: `docs/thoughts-directory.md`
**Changes**:
- Remove `index.json` from directory structure (line 7)
- Remove "codebase index" mention from intro paragraph (line 1)

### Success Criteria:

#### Automated Verification:
- [x] `go build ./cmd/rpi/` succeeds
- [x] `go test ./cmd/rpi/ -count=1 -run TestInit` passes (if init tests exist)
- [x] No matches for `index.json`, `rpi index`, `rpi_index`, or `internal/index` in `docs/`, `README.md`, or `cmd/rpi/init_cmd.go`

### Commit:
- [x] Stage: `cmd/rpi/init_cmd.go`, `cmd/rpi/sync.go`, `README.md`, `docs/architecture.md`, `docs/rpi-init.md`, `docs/thoughts-directory.md`
- [x] Message: `docs: remove index references from init and documentation`

---

## References
- Design: .rpi/designs/2026-04-04-remove-index-subsystem-from-rpi.md
- Spec: .rpi/specs/index-removal.md
