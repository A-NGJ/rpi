---
date: 2026-04-07T15:00:00+02:00
related_research: .rpi/research/2026-04-07-rpi-improvements-via-claude-code-internals.md
spec: .rpi/specs/session-awareness.md
status: complete
tags:
    - design
    - mcp
    - session
    - pipeline
topic: session awareness tools
---

# Design: Session Awareness Tools

## Summary

Add two MCP tools (`rpi_session_resume`, `rpi_suggest_next`) and corresponding CLI commands (`rpi resume`, `rpi next`) that give AI assistants awareness of active work and pipeline flow. `rpi_session_resume` returns a compact summary of all active artifacts, the current implementation context, and a suggested next action. `rpi_suggest_next` analyzes artifact state and recommends the next pipeline step. Both wire into Claude Code hooks (SessionStart, Stop) for automatic triggering; other tools call them via MCP or CLI.

## Context

The research (.rpi/research/2026-04-07-rpi-improvements-via-claude-code-internals.md) identified session awareness as the broadest-impact improvement. Currently, an AI assistant starting a new session has no way to discover active work without multiple `rpi_scan` calls and manual interpretation. Similarly, when finishing a task, there's no programmatic way to determine the next pipeline step — that logic lives only in CLAUDE.md prose and skill prompts, and is lost after context compaction.

The research proposed 6 new MCP tools. After analysis, 3 were dropped as redundant or low-value:
- `rpi_plan_progress` — reading covered by `rpi_context_essentials`, writing by direct file editing
- `rpi_validate_action` — enforcement only automatic in Claude Code; prompt-level guidance is sufficient
- `rpi_spec_drift_check` — overlaps with `rpi_verify_spec`; better added as a mode to that tool later

## Constraints

- **MCP-first**: All logic lives in the Go binary. Claude Code hooks are thin prompt-injection triggers
- **Portability**: Works identically across all MCP clients. Hook configuration is Claude Code-only and optional
- **Reuse existing internals**: `scanner.Scan`, `assembleContext`/`detectCurrentPhase`, `chain.Resolve`, `parseCheckboxes`, `frontmatter.Parse`
- **No new packages**: Logic fits in `cmd/rpi/` alongside existing command files
- **Pipeline rules are deterministic**: Priority-ordered evaluation, not heuristics

## Components

### 1. `rpi_suggest_next` MCP Tool + `rpi next` CLI Command

The pipeline suggestion engine. Used standalone and embedded in session resume.

**Input**: Optional `artifact_path` string — when provided, suggests the next step for that specific artifact instead of scanning all artifacts.

**Output**:

```json
{
  "action": "/rpi-implement .rpi/plans/2026-04-07-context-preservation.md",
  "reasoning": "Active plan 'context preservation' has 7 unchecked items remaining in Phase 3",
  "artifact": ".rpi/plans/2026-04-07-context-preservation.md"
}
```

When nothing is actionable:

```json
{
  "action": "/rpi-propose or /rpi-research",
  "reasoning": "No active or draft artifacts found",
  "artifact": ""
}
```

**Pipeline rules** (evaluated in priority order — later stages first):

| Priority | Condition | Suggested Action |
|---|---|---|
| 1 | Active plan with unchecked items | `/rpi-implement <plan-path>` |
| 2 | Active plan with all items checked | `/rpi-verify <spec-path>` (resolves chain to find spec) |
| 3 | Active design with no downstream plan | `/rpi-plan <design-path>` |
| 4 | Draft design | Review and approve design at `<path>` |
| 5 | Complete research with no downstream design | `/rpi-propose` referencing the research |
| 6 | Draft research | Review and finalize research at `<path>` |
| 7 | Nothing active | `/rpi-propose` or `/rpi-research` |

**Multiple artifacts at same priority**: Most recently dated artifact wins (by frontmatter `date` field).

**"No downstream" detection**: After scanning all artifacts, read frontmatter for each plan to collect their `design` field values, and for each design to collect their `related_research` field values. This builds two sets: "designs that have plans" and "research that has designs." In-memory comparison — no additional scans needed.

**Algorithm**:

```
suggestNext(rpiDir, artifactPath):
  if artifactPath given:
    read artifact frontmatter → determine type and status → apply matching rule
    return suggestion

  allArtifacts = scanner.Scan(rpiDir, Filters{})

  // Build downstream lookup maps
  planDesigns = set{}    // design paths that have at least one plan
  designResearch = set{} // research paths that have at least one design
  for each artifact where type=plan:
    doc = frontmatter.Parse(artifact.path)
    if doc has "design" field: planDesigns.add(design value)
  for each artifact where type=design:
    doc = frontmatter.Parse(artifact.path)
    if doc has "related_research" field: designResearch.add(research value)

  // Evaluate rules in priority order
  activePlans = filter(allArtifacts, type=plan, status=active) sorted by date desc
  for each plan:
    checkboxes = parseCheckboxes(readFile(plan.path))
    if unchecked > 0: return implement suggestion
    else: resolve chain → find spec → return verify suggestion

  activeDesigns = filter(allArtifacts, type=design, status=active) sorted by date desc
  for each design not in planDesigns:
    return plan suggestion

  draftDesigns = filter(allArtifacts, type=design, status=draft) sorted by date desc
  if any: return review suggestion for most recent

  completeResearch = filter(allArtifacts, type=research, status=complete) sorted by date desc
  for each research not in designResearch:
    return propose suggestion

  draftResearch = filter(allArtifacts, type=research, status=draft) sorted by date desc
  if any: return review suggestion for most recent

  return "start new work" suggestion
```

**Alternatives considered for downstream detection**:
- **`scanner.CountReferences`**: Checks all artifacts for references, not just the target type. A design referenced only by a review would incorrectly appear to have a downstream artifact. Rejected.
- **`scanner.Scan` with `References` filter per candidate**: Correct but O(n) scans per candidate. Current approach does O(n) file reads upfront, which is simpler and bounded.

### 2. `rpi_session_resume` MCP Tool + `rpi resume` CLI Command

Assembles a session-level overview by combining artifact scanning, plan context, and pipeline suggestion.

**Input**: No parameters.

**Output**:

```json
{
  "artifacts": [
    { "path": ".rpi/plans/foo.md", "type": "plan", "status": "active", "topic": "context preservation" },
    { "path": ".rpi/designs/bar.md", "type": "design", "status": "active", "topic": "session awareness" }
  ],
  "active_plan": {
    "path": ".rpi/plans/foo.md",
    "topic": "context preservation",
    "current_phase": "Phase 3: MCP Tool Registration",
    "progress": { "checked": 5, "total": 12 },
    "next_items": ["Add sessionInput struct", "Add handleSessionResume handler", "Register tool in serve.go"]
  },
  "suggestion": {
    "action": "/rpi-implement .rpi/plans/foo.md",
    "reasoning": "Active plan 'context preservation' has 7 unchecked items in Phase 3",
    "artifact": ".rpi/plans/foo.md"
  }
}
```

When nothing is active:

```json
{
  "artifacts": [],
  "active_plan": null,
  "suggestion": {
    "action": "/rpi-propose or /rpi-research",
    "reasoning": "No active or draft artifacts found",
    "artifact": ""
  }
}
```

**Logic**:
1. `scanner.Scan(rpiDir, Filters{})` — get all non-archived artifacts
2. Filter to active/draft status only for the `artifacts` list
3. If active plans exist, build `active_plan` context — reuses `detectCurrentPhase` and `parseCheckboxes` from `context.go` for the most recently dated active plan
4. Call `suggestNext` (from component 1) for the suggestion
5. Compose result

**Differences from `rpi_context_essentials`**:

| Aspect | `rpi_context_essentials` | `rpi_session_resume` |
|---|---|---|
| Scope | Single active plan | All active/draft artifacts |
| Plan context | Phase + progress + next items | Same (reuses logic) |
| Spec scenarios | Yes (titles) | No (session-level, not implementation-level) |
| Design constraints | Yes | No |
| Suggestion | No | Yes |
| Git state | Yes (branch, uncommitted count) | No (available via `rpi_git_context`) |
| Input | Optional `plan_path` | None |

The tools are complementary: `rpi_session_resume` answers "where did I leave off?", while `rpi_context_essentials` answers "what are the details of my current implementation?"

### 3. Hook Configurations

Added to `.claude/settings.json` by `rpi init` (Claude target only), alongside the existing PostCompact hook.

**SessionStart** — remind to call session resume:
```json
{
  "type": "command",
  "command": "cat <<'HOOK_EOF'\nCall the rpi_session_resume MCP tool to see active work and suggested next steps.\nHOOK_EOF"
}
```

**Stop** — remind to check next step:
```json
{
  "type": "command",
  "command": "cat <<'HOOK_EOF'\nCall the rpi_suggest_next MCP tool to determine the appropriate next pipeline step.\nHOOK_EOF"
}
```

Same pattern as the PostCompact hook: thin prompt injection, AI calls the MCP tool itself via normal MCP request/response flow.

### 4. `rpi init` Integration

The `configureHooks` function (`cmd/rpi/init_cmd.go:299`) currently handles only PostCompact. It needs to be generalized to configure all three hooks (PostCompact, SessionStart, Stop) using the same merge pattern:

1. Read existing settings.json
2. Parse existing hooks map
3. For each hook event, check if our entry is already present (by marker string)
4. If not present, append our entry to the hook array
5. Write back

The marker strings for deduplication:
- PostCompact: `"rpi_context_essentials"` (existing)
- SessionStart: `"rpi_session_resume"`
- Stop: `"rpi_suggest_next"`

This replaces the current single-hook function with a loop over a hook definition table, keeping the merge logic identical.

## File Structure

| File | Change |
|---|---|
| `cmd/rpi/suggest.go` | **New** — pipeline suggestion logic, `suggestNext` function, CLI command, output structs |
| `cmd/rpi/suggest_test.go` | **New** — tests for each pipeline rule, priority ordering, downstream detection |
| `cmd/rpi/resume.go` | **New** — session resume assembly, CLI command, output structs |
| `cmd/rpi/resume_test.go` | **New** — tests for artifact grouping, plan context inclusion, empty state |
| `cmd/rpi/serve.go` | **Modified** — register `rpi_session_resume` and `rpi_suggest_next` MCP tools, add input structs |
| `cmd/rpi/init_cmd.go` | **Modified** — generalize `configureHooks` to handle SessionStart and Stop alongside PostCompact |
| `cmd/rpi/init_cmd_test.go` | **Modified** — add tests for new hooks |

## Risks

- **Stop hook noise**: The Stop hook fires after every task completion, not just pipeline-relevant ones. The AI may call `rpi_suggest_next` after trivial actions like answering a question. Mitigation: the tool is cheap (one scan + a few file reads), and the AI can ignore irrelevant suggestions. If noise proves problematic, the hook can be removed without affecting the MCP tool.
- **Pipeline rules don't cover all edge cases**: Superseded artifacts, plans with no checkboxes, specs with no upstream plan. Mitigation: unrecognized states fall through to the "start new work" suggestion. Rules can be expanded incrementally.
- **Stale frontmatter dates**: If artifacts have incorrect or missing dates, the "most recent" tiebreaker may pick the wrong artifact. Mitigation: `rpi scaffold` sets dates automatically; manual files are the user's responsibility.

## Out of Scope

- Plan boundary enforcement (`rpi_validate_action`)
- Spec drift detection (future `rpi_verify_spec` mode)
- Plan progress writing (`rpi_plan_progress`)
- Plugin packaging
- Custom agent definitions
- Hook configuration for non-Claude Code tools (prompt guidance only)

## References

- Research: .rpi/research/2026-04-07-rpi-improvements-via-claude-code-internals.md
- Related design: .rpi/designs/2026-04-07-context-preservation-via-rpi-context-essentials.md
