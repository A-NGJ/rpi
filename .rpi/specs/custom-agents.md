---
domain: custom-agents
feature: custom-agents
last_updated: 2026-04-07T15:57:43+02:00
updated_by: .rpi/designs/2026-04-07-custom-agents-for-verification-and-worktree-implementation.md
---

# Custom Agents

## Purpose

Provide a Claude Code custom agent — a read-only verification agent — so that spec conformance checks can run in parallel with implementation.

## Scenarios

### Init installs agent definitions for Claude target
Given a project being initialized with the Claude target
When `rpi init` completes
Then `.claude/agents/` contains `rpi-verify.md`

### Init skips agent definitions for non-Claude targets
Given a project being initialized with the agents-only or opencode target
When `rpi init` completes
Then no agent definition files are installed

### Update syncs agent definitions
Given a project already initialized for the Claude target
When `rpi update` runs
Then new or changed agent definitions are installed to `.claude/agents/` without overwriting user modifications unless forced

### Verification agent restricted to read-only operations
Given the installed verification agent definition
When inspecting its tool restrictions
Then it permits only read, search, and RPI MCP tools — no file creation or modification tools

### Verification agent returns structured results
Given an active plan with linked specs
When the verification agent is spawned with a plan path
Then it checks each spec scenario against actual code, reports pass/fail per scenario with file references, and returns an overall verdict

### Implement skill uses worktree isolation
Given the implement skill running in Claude Code
When the user invokes the implement skill with a plan
Then implementation runs in an isolated worktree via the Agent tool's `isolation: "worktree"` parameter

### Implement skill falls back to in-place mode on failure
Given an implementation where worktree mode fails (e.g., merge conflicts)
When the skill detects the failure
Then implementation proceeds in-place on the current branch

### Auto-merge on verification pass
Given a worktree implementation has completed all phases and verification passes with no manual verification items
When the main conversation receives the verification result
Then the worktree branch is automatically merged to the base branch and the plan status is updated to complete

### Manual verification pauses merge
Given a worktree implementation has completed all phases but the plan includes manual verification items that automated checks cannot cover
When the main conversation receives the verification result
Then the diff is presented to the user and merge waits for explicit approval

### Base branch stays clean during implementation
Given an implementation running in worktree mode
When the implementation agent is working
Then the base branch has no uncommitted or partial changes from the implementation

## Constraints

- Agent definitions are Claude Code-specific; they are never installed for non-Claude targets
- Verification agent must not modify any files — enforcement via allowed-tools restriction
- Implementation context is passed via the Agent tool's prompt parameter; the agent does not re-read the entire artifact chain
- Merge uses regular merge (not squash) to preserve per-phase commit history
- After successful merge, the worktree and its branch are cleaned up
- No new MCP tools — agents orchestrate existing tools only

## Out of Scope

- Plugin packaging (bundling agents + skills + hooks into one installable unit)
- Per-skill hooks for boundary enforcement
- Artifact navigator agent
- OpenCode or agents-only agent equivalents
- Parallel phase execution within the worktree
- Verification report file generation (the agent returns inline results; the `/rpi-verify` skill handles report files)
