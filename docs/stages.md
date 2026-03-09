# How Each Stage Works

## Research (`/rpi-research`)

**Purpose:** Investigate the codebase -- fact-finding with optional assessment.

Works for both focused questions ("how does the auth pipeline work?") and open-ended exploration ("what could we improve about error handling?"). The command spawns parallel sub-agents that use specialized skills:
- **locate-codebase** -- Finds where files and components live
- **codebase-analyzer** -- Understands how specific code works (traces data flow, documents patterns)
- **find-patterns** -- Finds examples of existing patterns to model after
- **locate-thoughts** -- Discovers relevant historical documents in `.thoughts/`
- **analyze-thoughts** -- Extracts key insights from existing documents

For broad queries, Claude shows you what it found in an initial scan and asks if you want to redirect focus before deep-diving.

Findings are always factual with concrete file:line references. When findings reveal clear pain points, opportunities, or trade-offs, an Assessment section provides an opinionated take -- clearly separated from the facts. Output is optionally saved as a structured markdown document. If the research comprehensively documents a module's behavior, it can optionally create or update a spec in `.thoughts/specs/`.

## Design (`/rpi-design`)

**Purpose:** Make architectural decisions and document trade-offs.

Three modes, auto-detected:
- **Lightweight** -- Focused decision between 2-3 options (single component, one pattern choice)
- **Comprehensive** -- Multi-decision feature design with component diagrams and risk tables
- **Incremental** -- Update an existing design with new information

The design stage is interactive. Claude presents options with concrete trade-offs (grounded in your actual codebase, not generic advice), makes a recommendation, and waits for your direction. After you choose, it validates that your combined choices work together before documenting. If the design changes existing behavior documented in `.thoughts/specs/`, it can flag those specs with `pending_changes` for update after implementation.

## Structure (`/rpi-structure`) -- Optional

**Purpose:** Map a design to concrete file layout when the structure itself is complex.

Only needed for greenfield projects or major reorganizations. Defines file changes, module boundaries, public APIs with concrete signatures, and a dependency graph.

## Tickets (`/rpi-tickets`) -- Optional

**Purpose:** Break a large design into independently plannable work units.

Each ticket is self-contained -- it extracts the relevant design decisions, interfaces, and constraints so that `/rpi-plan` can produce a focused plan without reading the full design document. Tickets include dependency graphs and recommended implementation order.

## Plan (`/rpi-plan`)

**Purpose:** Create a phase-by-phase implementation plan with specific code changes and verification steps.

Works in two modes:
- **Standalone** -- For simple tasks. Does its own lightweight research and produces a plan directly.
- **Pipeline** -- For complex tasks with existing design/ticket documents. Reads the full document chain, spot-checks the codebase against the docs, and breaks work into verified phases.

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
2. For each phase, shows a **pre-review** of all intended changes before writing code
3. Implements the phase after approval
4. Runs automated verification (tests, linting, type checks)
5. Updates checkboxes in the plan file
6. Proposes a commit (no Claude attribution)
7. Pauses for manual verification before the next phase

If the plan doesn't match reality (codebase drifted since the plan was written), it stops and clearly explains the mismatch rather than silently improvising.

Resumable: if you invoke `/rpi-implement` on a partially-completed plan, it picks up from the first unchecked item.

## Commit (`/rpi-commit`)

**Purpose:** Create clean, focused git commits without thinking about `git add` and message formatting.

`/rpi-commit` inspects your working tree (staged, unstaged, and untracked files), groups related changes into logical commits, drafts messages matching your repo's existing commit style, and presents the plan for your approval before executing. It never adds Claude attribution or co-author lines -- commits look like you wrote them.

You can use `/rpi-commit` standalone anytime, or let `/rpi-implement` handle commits at the end of each phase.

## Verify (`/rpi-verify`)

**Purpose:** Validate that an implementation matches its design artifacts.

Checks three dimensions:
- **Completeness** -- Are all planned changes implemented?
- **Correctness** -- Does the code match the design decisions?
- **Coherence** -- Do the pieces work together as intended?

Can auto-detect what to verify from recent git changes and active plans, or you can point it at a specific design doc, plan, or ticket. Produces a severity-classified report. Purely advisory -- it doesn't block anything, and can be re-run after fixes.

## Archive (`/rpi-archive`)

**Purpose:** Move completed artifacts to `.thoughts/archive/` to keep the active directory clean.

Two modes:
- **Specific paths** -- `/rpi-archive .thoughts/research/2026-01-15-auth-flow.md`
- **Scan mode** -- `/rpi-archive` with no arguments scans for completed artifacts

Warns before archiving anything still in `draft` or `active` status. Preserves the full directory structure inside `archive/` (e.g., `archive/research/`, `archive/designs/`).
