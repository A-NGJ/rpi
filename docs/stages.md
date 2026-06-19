# How Each Stage Works

## Research (`/rpi-research`)

**Purpose:** Investigate the question -- codebase exploration or external systems/libraries/frameworks, with conversational fact-finding and an optional research artifact.

Works for both focused questions ("how does the auth pipeline work?", "what's the state of vector databases in 2026?") and open-ended exploration ("what could we improve about error handling?", "what agentic frameworks exist for X?"). The command starts with a brief interview to understand motivation, prior attempts, and constraints, then investigates at a depth proportional to the question -- across the codebase, external sources, or both.

Research is conversational by default -- no artifact is created unless you ask or it's clearly valuable for cross-session handoff. For broad queries, Claude shows initial findings and asks if you want to redirect focus before deep-diving.

Findings carry source anchors: `file:line` references for codebase claims, URL or quoted documentation for external claims. Authoritative external sources (project README, official docs, release notes) are preferred; blog posts and forum threads are flagged as such. Facts are presented first; an opinionated assessment is offered only when warranted, clearly separated from the facts. If you want to save findings, ask Claude to write them to `.rpi/research/`.

## Propose (`/rpi-propose`)

**Purpose:** Investigate, analyze, and propose solutions with trade-offs.

Three modes, auto-detected:
- **Quick** -- Focused decision between 2-3 options (single component, one pattern choice)
- **Full** -- Multi-decision feature design with component diagrams, risk tables, and file structure
- **Incremental** -- Update an existing proposal with new information

The propose stage is interactive. Claude investigates the codebase, presents options with concrete trade-offs (grounded in your actual codebase, not generic advice), makes a recommendation, and waits for your direction. After you choose, it validates that your combined choices work together before documenting the design in `.rpi/designs/`. If the proposal changes existing behavior documented in `.rpi/specs/`, it can flag those specs with `pending_changes` for update after implementation.

**Pre-lock audit:** Before the design/spec approval gate, a read-only `rpi-slice-audit` pass checks the drafted Components hold together -- every `## File Structure` entry is introduced by some Component and vice-versa (coverage), no Component references a file or symbol no Component defines (mismatch), and no Component contradicts a decision recorded earlier in the design (decision-drift). Findings surface at the gate and block approval until resolved or waived. A single-Component design runs only the lightweight coverage check.

**Grill mode (opt-in):** Pass `--grill` (or use phrasing like "grill me on this", "stress-test this") to invoke the bundled `grill-me` skill (sourced from [mattpocock/skills](https://github.com/mattpocock/skills) under MIT) at the approval gate. Once the design and spec are drafted, `grill-me` interrogates them one question at a time, walking every branch of the decision tree before you accept the artifact. Findings are applied inline to the design and spec; no separate audit log is written. Grilling is single-pass -- re-invoke if you want a second round.

If a user has removed the bundled `grill-me` skill, you'll be told and asked whether to proceed with the standard approval gate.

**Fast-forward mode (opt-in):** Pass `--ff` to suppress all approval gates (trade-off buy-in, mid-flight checkpoints, spec approval) and auto-chain into `/rpi-plan --ff <design-path>` immediately. The chain continues through `/rpi-implement` and ends at `/rpi-verify`, producing a verification report in `.rpi/reviews/` as the terminal artifact. Mutually exclusive with `--grill`; the explicit flag is required (no natural-language trigger). Use when you trust the defaults and want full autopilot. Safety gates (codebase drift detection, sensitive-content scan) and a hard pre-lock coverage gap still stop the chain.

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

**Pre-lock audit:** Before the phase-outline buy-in gate, a read-only `rpi-slice-audit` pass checks the drafted phases hold together -- every success criterion and planned file maps to real work (coverage), no phase edits a file only a later phase creates (forward-reference), and no phase contradicts a decision recorded upstream (decision-drift). Findings surface at the gate and block approval until resolved or waived. A single-phase standalone plan skips the cross-slice pass and runs only the lightweight coverage check; in a split, the audit runs per sibling plan.

**Grill mode (opt-in):** Pass `--grill` (or use phrasing like "grill me on this") to invoke the bundled `grill-me` skill on the *phase outline* before the full plan is written. Same fall-back behavior as Propose -- if a user has removed `grill-me` locally, you'll be asked whether to proceed without it.

**Fast-forward mode (opt-in):** Pass `--ff` to skip the phase outline buy-in and auto-chain into `/rpi-implement --ff <plan-path>`, terminating at `/rpi-verify`. Pre-flight checks (design status, artifact chain, drift spot-check) and the design-coverage check still run and can stop the chain. Pre-lock audit findings are recorded rather than blocking under `--ff` -- **except a hard coverage gap** (a success criterion or planned file mapping to no work), which stops the chain even under `--ff`, since it would waste the entire downstream run. Mutually exclusive with `--grill`.

## Blueprint (`/rpi-blueprint`)

**Purpose:** Fuse Propose and Plan into a single pass for low-stakes solo work -- go from a research note or a short problem statement straight to a phased plan, without producing a separate reviewable design.

Blueprint does *condensed* design reasoning inline (it commits to the obvious approach and weighs it against the obvious alternatives, without the parallel option-exploration `/rpi-propose` runs) and emits the plan directly. Two things it never drops:

- **A minimal behavioral spec.** SDD is non-negotiable -- every plan that lands behavior is backed by a spec. The blueprint spec is *minimal* (3-5 user-observable scenarios) rather than the 5-8 a full propose spec carries, reflecting the small scope that qualifies for this path. A blueprint plan with no linked spec is a contract violation.
- **The design reasoning.** Because there is no design file, the chosen approach, the alternatives considered and dropped, and the blast-radius judgment are recorded in a `## Design Notes` block near the top of the plan -- keeping the rationale auditable and giving `/rpi-verify` something to check against.

**Fused vs fast-forward -- the deliverable axis.** This is the distinction to keep straight. `--ff` runs the *full* pipeline fast: it suppresses review *pauses* but still produces a design artifact in `.rpi/designs/`. `/rpi-blueprint` is orthogonal -- it structurally *omits* the design *deliverable*, fusing the design reasoning into the plan. The two compose: `rpi-blueprint --ff` means "fuse the design into the plan *and* don't pause for plan approval, then auto-run implement → verify." And `/rpi-blueprint` differs from `/rpi-plan` in the other direction: `/rpi-plan` standalone is scoped to changes to *existing* behavior with no tradeoffs and carries no design reasoning, while blueprint does condensed design reasoning and is research-grounded.

**Refuse-and-redirect (hard gate).** Blueprint is for low-stakes work. It declines and points you at `/rpi-propose` when the work surfaces genuine tradeoffs, more than one approach a reasonable engineer would defend, or high blast radius (using the same complexity *signals* `rpi split-score` encodes -- component count, directory spread, multi-spec -- as one input). This is an integrity gate, not a review pause: it fires **even under `--ff`** and stops the chain rather than silently escalating into `rpi-propose --ff`, leaving you to choose the full design path deliberately.

**Fast-forward / grill (opt-in):** `--ff` skips the plan-outline approval pause, writes the plan + minimal spec, then auto-chains `/rpi-implement --ff <plan-path>` and terminates at `/rpi-verify`. `--grill` runs a single-pass interrogation of the condensed design reasoning + phase outline before the plan is written. Mutually exclusive, per the shared flag contract; the refuse gate still stops a fast-forward run.

## Implement (`/rpi-implement`)

**Purpose:** Execute a plan phase-by-phase with verification at each step.

The implementation stage works phase-by-phase:

1. Reads the plan completely
2. Checks for sensitive files (credentials, secrets) before proceeding
3. For each phase, presents intended changes before writing code
4. Implements the phase
5. Runs automated verification (tests, linting, type checks)
6. Auto-commits when checks pass -- no manual confirmation needed
7. Updates checkboxes in the plan file
8. Advances to the next phase automatically if all success criteria are covered by automated checks; pauses only when the plan includes manual verification items
9. On completion, verifies spec conformance for all linked specs -- extracts Given/When/Then scenarios and checks each against the implementation

If the plan doesn't match reality (codebase drifted since the plan was written), it stops and clearly explains the mismatch rather than silently improvising.

Resumable: if you invoke `/rpi-implement` on a partially-completed plan, it picks up from the first unchecked item.

**Fast-forward mode (opt-in):** Pass `--ff` to skip the per-phase pre-review and any manual verification pauses. After the plan completes, `/rpi-verify <plan-path>` is invoked automatically, producing a verification report in `.rpi/reviews/` as the chain's terminal artifact. The "On mismatch" gate, the sensitive-content scan, end-of-plan spec verification, and phase failure handling all run unchanged -- `--ff` is a review override, not a safety override. Mutually exclusive with `--grill`.

## Revise (`/rpi-revise`)

**Purpose:** Amend an *existing* plan in `.rpi/plans/` when a new constraint or a review finding lands after the plan is drafted or partially implemented -- without redoing finished work or skipping the audit a fresh plan would get.

Use it when the situation changed under a plan you already have: an API shifted, a requirement was added, or a `/rpi-verify` review surfaced a gap. This is distinct from `/rpi-plan`, which creates a *fresh* scoped change from scratch. A change that invalidates the underlying design routes back through `/rpi-propose` instead -- revise operates on the plan layer only.

The amendment loop:

1. Captures a *before* snapshot of every checked/unchecked item (with its phase and `**File**:` context) as the preservation baseline
2. Identifies the changed-phase set the request touches and presents it for buy-in before writing
3. Edits only the affected phases, leaving unaffected phases byte-stable; affected items are re-keyed by their text (not line position) so reordering preserves `[x]`, new items arrive unchecked, removed items disappear
4. Re-runs the slice / pre-lock audit on the changed phases only -- this is `/rpi-plan`'s audit applied to what changed, not a reimplementation; a renumbering seam counts as changed so the ordered-phase invariant is re-checked across it
5. Captures an *after* snapshot and asserts no completed item was silently reset -- if a change would reopen finished work, it stops and shows exactly which items and asks for explicit confirmation
6. Hands back to `/rpi-implement`, which resumes from the first unchecked item

A `draft` plan stays `draft` and an `active` plan stays `active`, amended in place. A `complete` plan is never silently reopened: revise stops and offers an explicit choice -- a guarded, confirmed reopen, or supersede it via `/rpi-plan` carrying the unchanged phases into a successor plan.

This closes the **verify → revise → implement** loop: a verification finding becomes the change request, the affected phase is amended, and implementation resumes -- no fresh design pass required.

**Fast-forward / grill (opt-in):** Per the shared flag contract, `--ff` skips the approval gate and auto-chains back into `/rpi-implement --ff <plan-path>` -- but the protection against silently undoing completed work is **not** an approval gate and still fires (a revision that would reopen finished work stops and reports). `--grill` stress-tests the proposed changed-phase amendment before writing. Mutually exclusive.

## Commit (`/rpi-commit`)

**Purpose:** Create clean, focused git commits without thinking about `git add` and message formatting.

`/rpi-commit` inspects your working tree (staged, unstaged, and untracked files), groups related changes into logical commits, drafts messages matching your repo's existing commit style, and presents the plan for your approval before executing. It never adds Claude attribution or co-author lines -- commits look like you wrote them.

You can use `/rpi-commit` standalone anytime, or let `/rpi-implement` handle commits at the end of each phase.

## Verify (`/rpi-verify`)

**Purpose:** Validate that an implementation matches its proposal artifacts. **This is the closing checkpoint of the pipeline -- not an optional add-on.**

A passing test suite proves the code is internally consistent; it does not prove the implementation matches the design. Verify is the step that catches the gap between intent and code, and it's the one that's easiest to skip and most worth running. Treat it as part of the normal Plan → Implement → Verify rhythm.

Checks three dimensions:
- **Completeness** -- Are all planned changes implemented?
- **Correctness** -- Does the code match the proposal decisions? Extracts Given/When/Then scenarios from linked specs and verifies each against actual code and tests, reporting pass/fail per scenario with file:line references.
- **Coherence** -- Do the pieces work together as intended?

Can auto-detect what to verify from recent git changes and active plans, or you can point it at a specific proposal, plan, or research doc. Produces a severity-classified report in `.rpi/reviews/`. It's advisory rather than a blocking gate -- you keep ownership of which findings to act on -- and it's idempotent, so re-run it freely after fixes. When you invoke `/rpi-implement --ff`, verify runs automatically at the end of the chain; otherwise, run it yourself.

**Grounding pass.** When a review draft contains at least one blocker or more than three findings, verify hands its findings to a read-only `rpi-ground` subagent that re-anchors each one against actual repository state and tags it `Verified | Weakened | Falsified` with a one-line evidence pointer. Only `Verified` findings keep blocker severity; `Weakened` findings are demoted out of the blocking set with a caveat, and `Falsified` findings (contradicted by the code) are excluded -- so you see fewer false-positive blockers. Trivial reviews skip grounding, and on non-Claude targets where the subagent isn't installed, verify degrades to a single-pass review with an explicit "grounding skipped" note.

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

## Resume (`rpi resume`)

**Purpose:** Session-level overview to answer "where did I leave my work?" after returning to a project.

Shows:
- Active and draft artifacts across all types (path, type, status, topic)
- Current phase, checkbox progress, and the next unchecked items of the most recent active plan
- A suggested next pipeline action (e.g., "design has no implementation plan -- run /rpi-plan")

```bash
rpi resume                  # human-readable text summary (default)
rpi resume --format json    # structured JSON (same shape as the MCP tool)
```

Also exposed as the `rpi_session_resume` MCP tool. Claude Code's `rpi init` wires a `SessionStart` hook that prompts the assistant to call it automatically, so the AI orients itself on the in-flight work before you type. You can invoke it manually any time by running `rpi resume` or by asking the AI something like "where did I leave my work?".
