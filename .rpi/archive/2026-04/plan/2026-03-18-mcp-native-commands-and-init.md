---
archived_date: "2026-04-02"
date: 2026-03-18T21:50:00+01:00
proposal: .rpi/proposals/2026-03-18-mcp-native-commands-and-init.md
spec: .rpi/specs/mcp-init.md
status: archived
tags:
    - plan
topic: mcp-native-commands-and-init
---

# mcp-native-commands-and-init — Implementation Plan

## Overview

Update `rpi init` to auto-configure the MCP server in `.claude/settings.local.json`, and rewrite all embedded workflow files (commands, skills) to reference MCP tool names instead of `rpi` CLI commands.

**Scope**: 2 files modified in Go, 9 markdown files updated (7 commands + 2 skills)

## Source Documents
- **Proposal**: .rpi/proposals/2026-03-18-mcp-native-commands-and-init.md
- **Spec**: .rpi/specs/mcp-init.md

## Deviation from Proposal

The proposal lists 4 skills needing updates, but only 2 actually contain `rpi` CLI references:
- `locate-codebase/SKILL.md` — has `rpi index status`, `rpi index query`, `rpi index files`
- `find-patterns/SKILL.md` — has `rpi index status`, `rpi index query`
- `locate-thoughts/SKILL.md` — no `rpi` CLI references (uses grep/glob)
- `analyze-thoughts/SKILL.md` — no `rpi` CLI references (uses direct reading)

Also `PIPELINE.md.template` only references `/rpi-*` slash commands, not `rpi` CLI — no changes needed.

The spec (MC-10, MC-11) is correct: "where they previously referenced CLI commands" means only skills with actual CLI refs need updating.

---

## Phase 1: Init MCP Configuration

**Spec behaviors**: MC-1, MC-2, MC-3, MC-4, MC-5, MC-6, MC-12

### Overview

Add `--no-mcp` flag and `configureMCP()` function to `init_cmd.go`. The function checks PATH, reads/merges `.claude/settings.local.json`, and writes MCP config. Called after workflow file installation, before index building.

### Tasks:

#### 1. Add `--no-mcp` flag and variable
**File**: `cmd/rpi/init_cmd.go`
**Changes**:
- Add `initNoMCP bool` to the var block (line 16-22)
- Register `--no-mcp` flag in `init()` (after line 88)

```go
var (
    // ... existing vars ...
    initNoMCP      bool
)

// In init():
initCmd.Flags().BoolVar(&initNoMCP, "no-mcp", false, "Skip MCP server configuration")
```

#### 2. Add `configureMCP()` function
**File**: `cmd/rpi/init_cmd.go`
**Changes**: Add new function after `ensureGitignoreEntry()`. Import `encoding/json` and `os/exec`.

```go
func configureMCP(w io.Writer, targetDir string) {
    // 1. Check rpi is in PATH
    if _, err := exec.LookPath("rpi"); err != nil {
        logWarning(w, "rpi not found in PATH — skipping MCP server configuration")
        return
    }

    settingsDir := filepath.Join(targetDir, ".claude")
    settingsPath := filepath.Join(settingsDir, "settings.local.json")

    // 2. Read existing settings if present
    var settings map[string]interface{}
    if data, err := os.ReadFile(settingsPath); err == nil {
        if err := json.Unmarshal(data, &settings); err != nil {
            logWarning(w, fmt.Sprintf("Failed to parse %s: %v — skipping MCP configuration", settingsPath, err))
            return
        }
    } else {
        settings = make(map[string]interface{})
    }

    // 3. Check if mcpServers.rpi already exists
    if mcpServers, ok := settings["mcpServers"].(map[string]interface{}); ok {
        if _, exists := mcpServers["rpi"]; exists {
            logWarning(w, "MCP server 'rpi' already configured in settings.local.json")
            return
        }
    }

    // 4. Merge MCP config
    mcpServers, ok := settings["mcpServers"].(map[string]interface{})
    if !ok {
        mcpServers = make(map[string]interface{})
    }
    mcpServers["rpi"] = map[string]interface{}{
        "command": "rpi",
        "args":    []interface{}{"serve"},
    }
    settings["mcpServers"] = mcpServers

    // 5. Write back
    data, err := json.MarshalIndent(settings, "", "  ")
    if err != nil {
        logWarning(w, fmt.Sprintf("Failed to marshal settings: %v", err))
        return
    }
    // Ensure .claude/ dir exists (it should from earlier step)
    os.MkdirAll(settingsDir, 0755)
    if err := os.WriteFile(settingsPath, append(data, '\n'), 0644); err != nil {
        logWarning(w, fmt.Sprintf("Failed to write %s: %v", settingsPath, err))
        return
    }
    logSuccess(w, "Configured MCP server in .claude/settings.local.json")
}
```

#### 3. Call configureMCP from runInit
**File**: `cmd/rpi/init_cmd.go`
**Changes**: Insert call after workflow file installation (line 209) and before index building (line 219). Guard with `!initNoMCP` and claude target check.

```go
// After "Install embedded workflow files" block:

// Configure MCP server (Claude only)
if !initNoMCP && cfg.target == workflow.TargetClaude {
    configureMCP(w, targetDir)
}
```

#### 4. Update resetInitFlags helper
**File**: `cmd/rpi/init_cmd_test.go`
**Changes**: Add `initNoMCP = false` to `resetInitFlags()` (line 11-17).

#### 5. Tests for MCP configuration
**File**: `cmd/rpi/init_cmd_test.go`
**Changes**: Add test functions covering spec behaviors MC-1 through MC-6 and MC-12.

```go
// MC-12: --no-mcp flag exists
func TestInitNoMCPFlag(t *testing.T) {
    // Verify the flag is registered on initCmd
}

// MC-1: Init writes MCP config (rpi is in PATH in test env)
func TestInitWritesMCPConfig(t *testing.T) {
    // Run init, read .claude/settings.local.json, verify mcpServers.rpi
}

// MC-3: Init merges with existing settings
func TestInitMergesMCPConfig(t *testing.T) {
    // Pre-create settings.local.json with {"permissions": {"allow": ["bash"]}}
    // Run init with --force, verify both keys present
}

// MC-4: Init warns on existing rpi MCP entry
func TestInitWarnsExistingMCPEntry(t *testing.T) {
    // Pre-create settings.local.json with mcpServers.rpi
    // Run init with --force, verify entry unchanged and warning printed
}

// MC-5: Init skips MCP with --no-mcp
func TestInitSkipsMCPWithFlag(t *testing.T) {
    // Run init with --no-mcp, verify no settings.local.json created
}

// MC-6: Correct config shape
func TestInitMCPConfigShape(t *testing.T) {
    // Run init, parse settings.local.json, verify exact shape
}
```

Note: MC-2 (rpi not in PATH) is harder to test without mocking `exec.LookPath`. Consider extracting a `lookPath` variable that tests can override, or accept that this path is covered by manual verification.

### Success Criteria:

#### Automated Verification:
- [x] `go test ./cmd/rpi/... -run TestInit` — all existing tests still pass
- [x] `go test ./cmd/rpi/... -run TestInitNoMCP` — new flag test passes
- [x] `go test ./cmd/rpi/... -run TestInitWritesMCP` — MCP config test passes
- [x] `go test ./cmd/rpi/... -run TestInitMerges` — merge test passes
- [x] `go test ./cmd/rpi/... -run TestInitWarnsExisting` — warning test passes
- [x] `go test ./cmd/rpi/... -run TestInitSkipsMCP` — skip test passes
- [x] `go test ./...` — full suite passes

#### Manual Verification:
- [x] `go build -o bin/rpi ./cmd/rpi && bin/rpi init --help` shows `--no-mcp` flag
- [x] Run `bin/rpi init` in a temp dir → `.claude/settings.local.json` contains correct MCP config
- [x] Run `bin/rpi init --no-mcp` in a temp dir → no `settings.local.json` created

### Commit:
- [ ] Stage: `cmd/rpi/init_cmd.go`, `cmd/rpi/init_cmd_test.go`
- [ ] Message: `feat(init): add MCP server auto-configuration`

**Note**: Pause for manual confirmation before proceeding to next phase.

---

## Phase 2: MCP-Native Embedded Assets

**Spec behaviors**: MC-7, MC-8, MC-9, MC-10, MC-11

### Overview

Rewrite all 7 command files and 2 skill files to reference MCP tool names instead of `rpi` CLI commands. Remove the PATH prerequisite block from all commands.

### Tasks:

#### 1. Update command: rpi-research.md
**File**: `internal/workflow/assets/commands/rpi-research.md`
**Changes**:
- Remove lines 12 (PATH prerequisite block)
- Line 49: `Use rpi to check for existing research` → `Use the rpi_scan tool to check for existing research`
- Line 61: `Use rpi to query the codebase index` → `Use the rpi_index_query tool`
- Line 64: `Use rpi to scan for existing documents` → `Use the rpi_scan tool`
- Line 89: `Use rpi to scaffold and save a research artifact` → `Use the rpi_scaffold tool to scaffold and save a research artifact`
- Line 91: `Use rpi to transition the research artifact` → `Use the rpi_frontmatter_transition tool to transition`

#### 2. Update command: rpi-propose.md
**File**: `internal/workflow/assets/commands/rpi-propose.md`
**Changes**:
- Remove line 12 (PATH prerequisite)
- Line 32: `Use rpi to query the codebase index` → `Use the rpi_index_query tool`
- Lines 43-47: Replace `rpi` refs with `rpi_scaffold`, `rpi_frontmatter_transition`, `rpi_frontmatter_get`, `rpi_scan`, `rpi_chain`
- Lines 66-68: `Use rpi to check its status` → `Use the rpi_frontmatter_get tool`, `Use rpi to resolve` → `Use the rpi_chain tool`, `Use rpi to scan` → `Use the rpi_scan tool`, `Use rpi to query` → `Use the rpi_index_query tool`
- Lines 91-96, 106: Replace all remaining `Use rpi` with specific MCP tool names
- Line 121: `use rpi to update frontmatter` → `use the rpi_frontmatter_set tool`

#### 3. Update command: rpi-plan.md
**File**: `internal/workflow/assets/commands/rpi-plan.md`
**Changes**:
- Remove line 10 (PATH prerequisite)
- Line 30: `use rpi to query the codebase index` → `use the rpi_index_query tool`
- Line 31: Same pattern for second occurrence
- Line 38: `Use rpi to scaffold and save` → `Use the rpi_scaffold tool`
- Lines 54-55: `Use rpi to check the proposal's status` → `Use the rpi_frontmatter_get tool`, `Use rpi to resolve the full artifact chain` → `Use the rpi_chain tool`
- Line 77: `Use rpi to scaffold and save a plan artifact` → `Use the rpi_scaffold tool`
- Lines 87-88: `Use rpi to transition` → `Use the rpi_frontmatter_transition tool`

#### 4. Update command: rpi-implement.md
**File**: `internal/workflow/assets/commands/rpi-implement.md`
**Changes**:
- Remove line 10 (PATH prerequisite)
- Line 20: `Use rpi to check the plan's current status` → `Use the rpi_frontmatter_get tool to check the plan's current status`
- Line 24: `use rpi to resolve the plan's artifact chain` → `use the rpi_chain tool to resolve the plan's artifact chain`
- Line 25: `use rpi to transition` → `use the rpi_frontmatter_transition tool`
- Lines 29-30: `Use rpi to check the plan's completeness` → `Use the rpi_verify_completeness tool`
- Line 45: `use rpi to check staged files for sensitive content` → `use the rpi_git_sensitive_check tool`
- Line 144: `use rpi to check the plan's completeness` → `use the rpi_verify_completeness tool`
- Line 158: `use rpi to transition the spec` → `use the rpi_frontmatter_transition tool`
- Line 161: `use rpi to mark the plan as complete` → `use the rpi_frontmatter_transition tool`

#### 5. Update command: rpi-verify.md
**File**: `internal/workflow/assets/commands/rpi-verify.md`
**Changes**:
- Remove line 12 (PATH prerequisite)
- Line 24: `Use rpi to get the list of changed files` → `Use the rpi_git_changed_files tool to get the list of changed files and the rpi_scan tool to find active plans/proposals`
- Line 30: `Use rpi to resolve the artifact chain` → `Use the rpi_chain tool to resolve the artifact chain`
- Line 32: `Use rpi to get the list of changed files` → `Use the rpi_git_changed_files tool`
- Line 43: `Use rpi for mechanical checks` → `Use the rpi_verify_completeness and rpi_verify_markers tools`
- Line 70: `rpi spec coverage` → `the rpi_spec_coverage tool`

#### 6. Update command: rpi-archive.md
**File**: `internal/workflow/assets/commands/rpi-archive.md`
**Changes**:
- Remove line 10 (PATH prerequisite)
- Line 23: `Use rpi to read each artifact's current status` → `Use the rpi_frontmatter_get tool to read each artifact's current status`
- Line 36: `Use rpi to discover archivable artifacts` → `Use the rpi_archive_scan tool`
- Line 79: `Use rpi to check for active references` → `Use the rpi_archive_check_refs tool`
- Line 109: `Use rpi to archive the artifact` → `Use the rpi_archive_move tool`
- Line 128: spec check with `rpi` → `the rpi_spec_coverage tool`

#### 7. Update command: rpi-commit.md
**File**: `internal/workflow/assets/commands/rpi-commit.md`
**Changes**:
- Remove line 10 (PATH prerequisite)
- Line 16: `Use rpi to gather consolidated git context` → `Use the rpi_git_context tool to gather consolidated git context`
- Line 26: `Use rpi to check staged files for sensitive content` → `Use the rpi_git_sensitive_check tool to check staged files for sensitive content`

#### 8. Update skill: locate-codebase/SKILL.md
**File**: `internal/workflow/assets/skills/locate-codebase/SKILL.md`
**Changes**:
- Line 14: `Run rpi index status` → `Use the rpi_index_status tool`
- Line 14-15: `use rpi index query "[topic]" and rpi index files` → `use the rpi_index_query tool with your topic and the rpi_index_files tool`

#### 9. Update skill: find-patterns/SKILL.md
**File**: `internal/workflow/assets/skills/find-patterns/SKILL.md`
**Changes**:
- Line 16: `Run rpi index status` → `Use the rpi_index_status tool`
- Line 16-17: `use rpi index query "[pattern-topic]"` → `use the rpi_index_query tool with your pattern topic`

### Success Criteria:

#### Automated Verification:
- [x] `go test ./...` — full suite passes (existing tests for command content still work)
- [x] `grep -r "must be available in PATH" internal/workflow/assets/commands/` — zero matches (MC-7)
- [x] `grep -rE 'Use `rpi` to|use `rpi` to|Run `rpi ' internal/workflow/assets/commands/` — zero matches (MC-8)
- [x] `grep -rE 'Use `rpi |use `rpi |Run `rpi ' internal/workflow/assets/skills/` — zero matches (MC-10)
- [x] `grep -r "rpi_scaffold\|rpi_scan\|rpi_chain\|rpi_frontmatter" internal/workflow/assets/commands/rpi-propose.md` — confirms MCP tool refs present (MC-9)

#### Manual Verification:
- [x] Read through each modified command file — MCP tool names are correct and contextually appropriate
- [x] Read through each modified skill file — MCP tool names are correct
- [x] Instructions remain clear and unambiguous

### Commit:
- [ ] Stage: `internal/workflow/assets/commands/*.md`, `internal/workflow/assets/skills/locate-codebase/SKILL.md`, `internal/workflow/assets/skills/find-patterns/SKILL.md`
- [ ] Message: `feat(commands): replace rpi CLI references with MCP tool names`

**Note**: Pause for manual confirmation.

---

## Post-Completion

After both phases are done:
1. Transition proposal to `complete` using `rpi_frontmatter_transition`
2. Transition spec from `draft` to `active` (it should move to `approved` after review, but active is the immediate next state)

## References
- Proposal: .rpi/proposals/2026-03-18-mcp-native-commands-and-init.md
- Spec: .rpi/specs/mcp-init.md
- MCP server: cmd/rpi/serve.go (tool names at lines 69-178)
- Init implementation: cmd/rpi/init_cmd.go
