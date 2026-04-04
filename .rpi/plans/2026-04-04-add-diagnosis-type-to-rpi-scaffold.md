---
date: 2026-04-04T01:28:16+02:00
status: complete
tags:
    - plan
topic: add diagnosis type to rpi-scaffold
---

# add diagnosis type to rpi-scaffold — Implementation Plan

## Overview

Add `diagnosis` as a sixth artifact type to `rpi scaffold`. Diagnosis artifacts are already recognized by `scan`, `status`, `sync`, and `init`, but scaffold doesn't support creating them yet. This follows the exact same pattern as the 5 existing types.

**Scope**: 4 files modified, 2 new files

## Phase 1: Register type, add template, update tests

### Overview

Single phase — register the diagnosis type in the scaffold type registry, add filename generation, create the template, update tests, and update the MCP schema tag.

### Tasks:

#### 1. Type registry
**File**: `cmd/rpi/scaffold.go`
**Changes**:
- Add `"diagnosis": "diagnoses"` to `typeDirs` map
- Add `"diagnosis"` to `validTypes` slice

#### 2. Filename generation
**File**: `internal/template/render.go`
**Changes**:
- Add `case "diagnosis"` to `GenerateFilename` switch — same pattern as `research`/`design`: `YYYY-MM-DD-<slug>.md`

#### 3. Template files
**Files**: `.rpi/templates/diagnosis.tmpl`, `internal/workflow/assets/templates/diagnosis.tmpl`
**Changes**:
- Create `diagnosis.tmpl` in both locations (filesystem + embedded fallback)
- Frontmatter: `date`, `topic`, `tags: [diagnosis, ...]`, `status: draft`
- Body sections: `# Diagnosis: {{.Topic}}`, `## Bug Report` (Expected/Actual/Reproduction), `## Root Cause`, `## Investigation Log`, `## Resolution`

#### 4. MCP schema
**File**: `cmd/rpi/serve.go`
**Changes**:
- Update `scaffoldInput.Type` jsonschema tag to include `diagnosis`

#### 5. Tests
**File**: `cmd/rpi/scaffold_test.go`
**Changes**:
- Add `{"diagnosis", []string{"--topic", "Test diagnosis"}}` to `TestScaffoldAllTypes` table

### Success Criteria:

#### Automated Verification:
- [x] `go test ./...` passes
- [x] `go vet ./...` clean

#### Manual Verification:
- [x] `go run ./cmd/rpi scaffold diagnosis --topic "test bug"` produces valid output with frontmatter and diagnosis sections

### Commit:
- [ ] Stage: `cmd/rpi/scaffold.go`, `cmd/rpi/scaffold_test.go`, `cmd/rpi/serve.go`, `internal/template/render.go`, `.rpi/templates/diagnosis.tmpl`, `internal/workflow/assets/templates/diagnosis.tmpl`
- [ ] Message: `feat(scaffold): add diagnosis artifact type`
