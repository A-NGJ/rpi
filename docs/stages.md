# How Each Stage Works

## Research (`/rpi-research`)

**Purpose:** Investigate the codebase -- conversational fact-finding with optional research artifact.

Works for both focused questions ("how does the auth pipeline work?") and open-ended exploration ("what could we improve about error handling?"). The command starts with a brief interview to understand motivation, prior attempts, and constraints, then investigates the codebase at a depth proportional to the question.

Research is conversational by default -- no artifact is created unless you ask or it's clearly valuable for cross-session handoff. For broad queries, Claude shows initial findings and asks if you want to redirect focus before deep-diving.

Findings include concrete file:line references. Facts are presented first; an opinionated assessment is offered only when warranted, clearly separated from the facts. If you want to save findings, ask Claude to write them to `.rpi/research/`.

## Propose (`/rpi-propose`)

**Purpose:** Investigate, analyze, and propose solutions with trade-offs.

Three modes, auto-detected:
- **Quick** -- Focused decision between 2-3 options (single component, one pattern choice)
- **Full** -- Multi-decision feature design with component diagrams, risk tables, and file structure
- **Incremental** -- Update an existing proposal with new information

The propose stage is interactive. Claude investigates the codebase, presents options with concrete trade-offs (grounded in your actual codebase, not generic advice), makes a recommendation, and waits for your direction. After you choose, it validates that your combined choices work together before documenting the design in `.rpi/designs/`. If the proposal changes existing behavior documented in `.rpi/specs/`, it can flag those specs with `pending_changes` for update after implementation.

## Plan (`/rpi-plan`)

**Purpose:** Create a phase-by-phase implementation plan with specific code changes and verification steps.

Works in two modes:
- **Standalone** -- For simple tasks. Does its own lightweight research and produces a plan directly.
- **Pipeline** -- For complex tasks with existing proposal documents. Reads the full document chain, spot-checks the codebase against the docs, and breaks work into verified phases. For large proposals, decomposes the work into independently plannable units within the plan.

Every plan phase includes:
- Specific file changes with code snippets
- Tests (in the same phase as the code they test, not a separate "testing phase")
- Automated verification commands (pulled from your project's `CLAUDE.md`)
- Manual verification steps
- A commit with specific files and message

All open questions must be resolved before the plan is finalized.

## Implement (`/rpi-implement`)

**Purpose:** Execute a plan phase-by-phase with verification at each step.

The implementation stage:
1. Reads the plan completely
2. Checks for sensitive files (credentials, secrets) before proceeding
3. For each phase, shows a **pre-review** of all intended changes before writing code
4. Implements the phase after approval
5. Runs automated verification (tests, linting, type checks)
6. Updates checkboxes in the plan file
7. If the plan specifies spec updates, updates the relevant `.rpi/specs/` files
8. Proposes a commit
9. Pauses for manual verification before the next phase

If the plan doesn't match reality (codebase drifted since the plan was written), it stops and clearly explains the mismatch rather than silently improvising.

Resumable: if you invoke `/rpi-implement` on a partially-completed plan, it picks up from the first unchecked item.

## Commit (`/rpi-commit`)

**Purpose:** Create clean, focused git commits without thinking about `git add` and message formatting.

`/rpi-commit` inspects your working tree (staged, unstaged, and untracked files), groups related changes into logical commits, drafts messages matching your repo's existing commit style, and presents the plan for your approval before executing. It never adds Claude attribution or co-author lines -- commits look like you wrote them.

You can use `/rpi-commit` standalone anytime, or let `/rpi-implement` handle commits at the end of each phase.

## Verify (`/rpi-verify`)

**Purpose:** Validate that an implementation matches its proposal artifacts.

Checks three dimensions:
- **Completeness** -- Are all planned changes implemented?
- **Correctness** -- Does the code match the proposal decisions?
- **Coherence** -- Do the pieces work together as intended?

Can auto-detect what to verify from recent git changes and active plans, or you can point it at a specific proposal, plan, or research doc. Produces a severity-classified report. Purely advisory -- it doesn't block anything, and can be re-run after fixes.

## Diagnose (`/rpi-diagnose`)

**Purpose:** Iteratively diagnose and fix complex bugs through root-cause analysis.

Takes a bug description, error message, or path to a failing test. The workflow:
1. Interviews you to establish expected vs actual behavior
2. Checks for existing diagnoses on the same topic
3. Reproduces the bug before investigating
4. Traces the code path with file:line references
5. Attempts a fix (up to 3 attempts, reverting failed ones)
6. Auto-commits on success, escalates to `/rpi-plan` or `/rpi-propose` if the bug is too complex

Always produces a diagnosis artifact in `.rpi/diagnoses/` regardless of outcome -- useful as a post-mortem even if the fix requires a larger effort.

## Explain (`/rpi-explain`)

**Purpose:** Walk through an implemented solution with a diff-scoped explanation.

Two modes, auto-detected:
- **With artifact path** -- Resolves the full artifact chain and walks through the diff with design context. Explains *why* each change was made, citing the design or plan.
- **No arguments** -- Auto-detects changed files from git and explains without artifact context.

File-by-file walkthrough with file:line references. Straightforward changes are summarized briefly; non-obvious changes get detailed reasoning. No artifact is saved by default -- ask if you want one.

## Spec Sync (`/rpi-spec-sync`)

**Purpose:** Sync behavioral specs in `.rpi/specs/` to match the current codebase.

Works in two phases:
1. **Scan** -- Reads all specs and compares against the actual implementation. Detects drift, naming mismatches, orphaned specs, and stale scenarios.
2. **Act** -- Presents a summary table of proposed actions (keep, rewrite, rename, merge, archive) and waits for approval before executing.

When code and spec disagree, code is truth -- the spec gets updated, not the code. On rename or merge, all cross-references across `.rpi/` are updated automatically.

Use after a batch of changes to keep specs current, or periodically as maintenance.

## Archive (`/rpi-archive`)

**Purpose:** Move completed artifacts to `.rpi/archive/` to keep the active directory clean.

Two modes:
- **Specific paths** -- `/rpi-archive .rpi/research/2026-01-15-auth-flow.md`
- **Scan mode** -- `/rpi-archive` with no arguments scans for completed artifacts

Warns before archiving anything still in `draft` or `active` status. Preserves the full directory structure inside `archive/` (e.g., `archive/2026-04/research/`, `archive/2026-04/designs/`).

## Status (`rpi status`)

**Purpose:** Single-screen dashboard of all RPI artifacts.

Shows:
- Artifact counts grouped by type and status
- Active plan progress with checkbox completion percentages
- Stale artifacts (non-terminal status, past a configurable threshold -- default 14 days)
- Specs with scenario counts
- Archive-ready artifacts with reference counts

```bash
rpi status                  # text dashboard
rpi status --format json    # machine-readable output
rpi status --stale-days 7   # flag artifacts stale after 7 days
```
