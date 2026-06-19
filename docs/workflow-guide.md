# Choosing Your Path

Not every task needs every stage. Pick the path that matches your task's complexity.

## Small Tasks (bug fixes, config changes, simple refactors)

**Path: Plan -> Implement**

Skip exploration and proposals entirely. `/rpi-plan` has a standalone mode that does lightweight research on the fly and produces a focused plan.

**Example -- Fix a broken date formatter:**
```
You:  /rpi-plan Fix the date formatter in utils/dates.ts that returns "NaN" for ISO strings without timezone
```

Claude researches the file, writes a 1-phase plan with the fix and test, saves it to `.rpi/plans/`.

```
You:  /rpi-implement .rpi/plans/2026-03-04-fix-date-formatter.md
```

Claude implements the phase, runs tests, and auto-commits when they pass.

**Example -- Add a missing environment variable:**
```
You:  /rpi-plan Add REDIS_URL to the config module with a default of localhost:6379
```

One phase, two files modified, done in minutes.

## Medium Tasks (focused features, single-concern changes)

**Path: Propose -> Plan -> Implement**

Use when the feature touches multiple files, involves a choice between approaches, or you're working in unfamiliar code. Optionally run `/rpi-research` first if you need to understand the codebase -- or survey external systems/frameworks -- before proposing.

**Example -- Add rate limiting to the API:**

**Step 1 (optional): Research**
```
You:  /rpi-research How does the API middleware chain work? Where are requests authenticated and validated?
```
Claude explores your codebase conversationally. You discuss findings interactively -- no artifact is created by default. If the exploration is thorough enough, you can ask it to save findings to `.rpi/research/`.

For external-system questions (e.g., *"what agentic frameworks exist for a data analytics tool?"*), `/rpi-research` surveys docs, READMEs, and release notes instead of source files -- citing URLs or quoted documentation rather than `file:line` references.

**Step 2: Propose**
```
You:  /rpi-propose I want to add per-endpoint rate limiting. Should handle both authenticated and anonymous users.
```
Claude investigates the codebase, presents 2-3 options (e.g., in-memory vs Redis, middleware vs decorator pattern), with pros/cons tied to your actual codebase. You pick an approach. It writes the design with the decision rationale and risk assessment to `.rpi/designs/`.

**Step 3: Plan**
```
You:  /rpi-plan .rpi/designs/2026-03-04-api-rate-limiting.md
```
Claude reads the proposal, spot-checks the codebase against the docs, and breaks the work into phases:
- Phase 1: Rate limiter core module + unit tests
- Phase 2: Middleware integration + integration tests
- Phase 3: Per-endpoint configuration + documentation

Each phase has specific file changes, code snippets, automated verification commands (pulled from your `CLAUDE.md`), and manual verification steps.

**Step 4: Implement**
```
You:  /rpi-implement .rpi/plans/2026-03-04-api-rate-limiting.md
```
For each phase, Claude runs the test suite, auto-commits when checks pass, updates checkboxes in the plan, and advances to the next phase. It only pauses when the plan includes manual verification items not covered by automated tests.

If something doesn't match the plan -- say a file was refactored since the proposal was written -- Claude stops and tells you exactly what diverged, rather than silently improvising.

## Large Tasks (multi-concern features, major refactors, greenfield projects)

**Path: Propose -> Plan (decomposes) -> Implement (per unit)**

Use when the work spans multiple systems, would produce more than ~4 implementation phases, or you want to ship incrementally over multiple sessions.

**Example -- Build a notification system (email, push, in-app):**

**Step 1 (optional): Research**
```
You:  /rpi-research How do we currently send emails? Is there any notification infrastructure? How do user preferences work?
```
Conversational research to build understanding before proposing.

**Step 2: Propose**
```
You:  /rpi-propose Build a notification system supporting email, push, and in-app channels. Users should be able to set per-channel preferences.
```
Claude goes into Full mode -- investigates the codebase deeply, identifies key design dimensions (channel abstraction, delivery strategy, preference storage, template system), presents options with trade-offs for each, validates that the chosen options compose well together, and writes the full design to `.rpi/designs/`.

**Step 3: Plan (with decomposition)**
```
You:  /rpi-plan .rpi/designs/2026-03-04-notification-system.md
```
Claude reads the proposal and decomposes it into independently plannable units:
```
1. Core notification model + channel interface
2. Email channel implementation
3. Push notification channel
4. In-app notification channel + UI
5. User preference management
6. Notification dispatch service (orchestrator)
```
Each unit has its own phases, file changes, and verification steps within the plan.

**Step 4: Implement (per unit)**
```
You:  /rpi-implement .rpi/plans/2026-03-04-notification-system.md
```

Claude implements unit by unit. You can stop between units, come back the next day, and pick up where you left off -- the `.rpi/` directory preserves all context and checkboxes track progress.

## Fused Shortcut (low-stakes solo work)

**Path: Blueprint -> Implement**

When you have a research note or a small, well-understood change and you *don't* want a separate design to review, `/rpi-blueprint` fuses Propose and Plan into one pass: it does condensed design reasoning inline and emits a phased plan directly, plus a minimal behavioral spec. The design reasoning that would have lived in a design file is recorded in a `## Design Notes` block at the top of the plan.

**Example -- from a research note straight to a plan:**
```
You:  /rpi-blueprint .rpi/research/2026-03-04-cache-warmup.md
```
Claude reads the research, commits to the obvious approach (recording the alternatives it dropped and why in `## Design Notes`), writes a phased plan and a minimal spec to `.rpi/plans/` and `.rpi/specs/`, then suggests `/rpi-implement`. No design artifact is produced.

**Example -- from a short problem statement:**
```
You:  /rpi-blueprint Add a --quiet flag to the CLI that suppresses progress output
```
Small enough to reason about in one pass -- Claude emits the plan + minimal spec directly.

**Power-user composition:**
```
You:  /rpi-blueprint --ff .rpi/research/2026-03-04-cache-warmup.md
```
Skips the plan-outline approval pause and auto-chains into `/rpi-implement --ff`, terminating at `/rpi-verify`.

### Where blueprint draws the line (refuse-and-redirect)

Blueprint is for low-stakes work, so it **declines and redirects to `/rpi-propose`** when the work deserves a reviewed design -- when condensed reasoning surfaces more than one approach a reasonable engineer would defend, or the change is wide-reaching / multi-component (high blast radius). Instead of landing a thin plan, it explains in a sentence or two why the work needs a design and points you at `/rpi-propose <input>`.

This refusal is a **hard gate, not a review pause**: it fires even under `--ff` and *stops* the run rather than silently escalating into `rpi-propose --ff`. Escalating autopilot across a deliverable boundary you didn't ask for would be surprising -- so blueprint stops and leaves the choice of the design path to you.

**Fused vs fast-forward.** Don't conflate the two "fast" concepts: `--ff` runs the *full* pipeline fast (suppressing review pauses) but still produces a design; `/rpi-blueprint` *omits the design deliverable* entirely, fusing its reasoning into the plan. They compose (`rpi-blueprint --ff`) but they are different axes.

## Not Sure Where to Start?

Use `/rpi-research` even when you have a vague idea. It handles focused codebase questions ("how does auth work?"), open-ended exploration ("what could we improve about error handling?"), and external surveys ("what agentic frameworks exist for X?", "what's the state of vector databases in 2026?"). It's conversational -- you discuss findings interactively and decide whether to save research or move straight to `/rpi-propose`.

```
You:  /rpi-research What could we improve about error handling?
```

## Before You Approve

### The pre-lock audit -- catching incoherent plans before you commit to them

When `/rpi-propose` drafts a design's components or `/rpi-plan` drafts a plan's phases, a read-only `rpi-slice-audit` pass runs **before** you're asked to approve. It catches the defects a single authoring pass routinely misses:

- **Coverage** -- a success criterion or planned file that maps to no real work, or a file slated for commit that no task produces.
- **Forward-references** -- a phase that edits or depends on something only a later phase creates.
- **Decision-drift** -- a slice whose behavior contradicts a decision recorded upstream (or earlier in the same design).

A clean audit is invisible -- you see a one-line "audit: clean" note and proceed to the normal gate. When it finds something, each finding names the slice and what it collides with. **Resolve** it (revise the slice) or **waive** it (acknowledge and proceed) -- you can't approve an interactive plan or design until every finding is resolved or waived.

**Under `--ff`**, findings are recorded with the artifact and do **not** stop the run -- with one exception. A **hard coverage gap** (a criterion or planned file mapping to no work) stops the chain even under `--ff`, because it would waste the entire downstream autopilot run on a plan that can't deliver what it promises. This is the only pre-lock finding that blocks `--ff`; forward-references and decision-drift are recorded but never block fast-forward.

A trivial single-phase standalone plan skips the cross-slice audit (one slice can't forward-reference a sibling) and runs only the lightweight coverage check.

## After Implementation

### Verify -- not optional, the closing checkpoint

`/rpi-verify` is the step most people are tempted to skip, and the one that pays off the most. A passing test suite tells you the code is internally consistent; it does **not** tell you the implementation matches what you actually designed. Verify is what catches the gap.

```
/rpi-verify .rpi/plans/2026-03-04-my-feature.md
```

It checks three dimensions -- **completeness** (are all planned changes present?), **correctness** (do the Given/When/Then scenarios in the linked specs match the actual code and tests, with file:line citations?), and **coherence** (do the pieces fit together?). The output is a severity-classified report in `.rpi/reviews/`, not a green/red gate, so you keep ownership of which findings to act on. Re-run after fixes -- it's cheap and idempotent.

When a review has a blocker or more than a few findings, verify runs a second read-only **grounding** pass (the `rpi-ground` subagent) that re-anchors each finding against the actual repo and demotes anything it can't confirm. Each finding gains a verdict and an evidence pointer, and the summary reports the before/after blocker count, e.g.:

```
Blockers: 3 drafted → 2 Verified, 1 Falsified (dropped)
- [Verified]  Missing migration for the new column (db/schema.sql:1 — column absent)
- [Verified]  No test covers the retry path (internal/queue/retry_test.go — no case)
- [Falsified] `parseConfig` was removed (still defined at internal/config/parse.go:42)
```

Only `Verified` findings stay in the blocking set, so the blockers you read are the ones the repo actually backs. On non-Claude targets the subagent isn't installed and verify falls back to a single-pass review with a "grounding skipped" note.

Treat verify as part of the normal Plan → Implement → Verify rhythm, not an optional add-on. If you used `--ff`, it already runs automatically at the end of the chain; if you didn't, run it yourself.

### Revise -- when the plan has to change after it's drafted

Plans don't survive contact with reality unchanged. A constraint lands mid-implementation, or a verify pass finds a gap, and the plan you already have needs to absorb it. Editing the plan file by hand is risky: it's easy to clobber the `[x]` of work you already finished, or to amend a phase without the audit a fresh plan would get. `/rpi-revise` does the amendment safely -- it edits only the affected phases, preserves the checkbox state of everything else, re-audits just what changed, and refuses to silently reopen completed work.

It's distinct from `/rpi-plan`: revise *amends an existing plan*, while plan *creates a fresh one*. (A change that invalidates the underlying design goes back through `/rpi-propose` instead.)

**Case 1 -- a constraint arrives mid-implementation.** You're partway through the rate-limiting plan when the team standardizes on Redis for shared state:
```
You:  /rpi-revise .rpi/plans/2026-03-04-api-rate-limiting.md the limiter must use Redis, not in-memory, so it works across instances
```
Claude snapshots the current checkbox state, identifies that only the limiter-core phase is affected, shows you that changed-phase set, rewrites just that phase (the already-checked middleware-integration work stays checked), re-runs the audit on the changed phase, confirms no completed item was reset, and suggests `/rpi-implement` -- which resumes at the first unchecked item.

**Case 2 -- a review finding flows back into the plan (verify → revise → implement).** A verify run flags a missing concern:
```
You:  /rpi-verify .rpi/plans/2026-03-04-api-rate-limiting.md
      → Blocker: no phase covers limiter eviction / TTL cleanup
You:  /rpi-revise .rpi/plans/2026-03-04-api-rate-limiting.md add eviction + TTL cleanup with tests to the limiter-core phase
You:  /rpi-implement .rpi/plans/2026-03-04-api-rate-limiting.md
```
The affected phase is amended to close the gap, unaffected phases keep their state, and implementation picks up the new work -- closing the loop without a fresh design pass.

A `complete` plan is never silently reopened: revise stops and offers an explicit choice -- a guarded, confirmed reopen, or supersede it via `/rpi-plan` carrying the unchanged phases forward. Under `--ff` the approval pause is skipped, but the protection against silently undoing completed work still fires.

### Other commands that close the loop

- **`/rpi-explain`** -- Walks through the diff with a file-by-file explanation. Useful for self-review or explaining changes to a teammate.
- **`/rpi-spec-sync`** -- Syncs specs in `.rpi/specs/` to match the current codebase. Run it after a batch of changes to detect drift, rewrite stale scenarios, rename or merge specs, and archive obsolete ones.
- **`/rpi-archive`** -- Moves completed artifacts to `.rpi/archive/` to keep the active directory clean. Run it when a feature is fully shipped and you no longer need the research/designs/plan documents in the active directories.
- **`rpi status`** -- Dashboard showing artifact counts, active plan progress, stale artifacts, and archive readiness. Run it anytime to see the overall state of `.rpi/`.

## Tips

- **Use `/rpi-diagnose` for complex bugs.** When a bug spans multiple files or the root cause isn't obvious, `/rpi-diagnose` will iteratively trace, fix, and verify -- producing a post-mortem artifact even if the fix requires escalation to `/rpi-plan`.
- **Start small.** Try `/rpi-plan` on a bug fix to see how the plan -> implement cycle feels before using the full pipeline.
- **Edit the artifacts.** The `.rpi/` documents are yours. If a proposal decision is wrong, edit it before planning. If a plan phase is unnecessary, delete it.
- **Use CLAUDE.md.** Add your project's test commands, linting setup, and conventions to `CLAUDE.md`. The pipeline stages pull verification commands from there.
- **Redirect during research.** When `/rpi-research` shows initial findings, tell it to focus on specific areas rather than researching everything.
- **Skip stages when they don't add value.** The full pipeline exists for complex work. Most daily tasks only need Plan -> Implement.
- **Review the pre-review.** `/rpi-implement` shows you exactly what it plans to change before writing code. This is your last checkpoint -- use it.
