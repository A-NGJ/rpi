# Choosing Your Path

Not every task needs every stage. Pick the path that matches your task's complexity.

## Small Tasks (bug fixes, config changes, simple refactors)

**Path: Plan -> Implement**

Skip research and design entirely. `/rpi-plan` has a standalone mode that does lightweight research on the fly and produces a focused plan.

**Example -- Fix a broken date formatter:**
```
You:  /rpi-plan Fix the date formatter in utils/dates.ts that returns "NaN" for ISO strings without timezone
```

Claude researches the file, writes a 1-phase plan with the fix and test, saves it to `.thoughts/plans/`.

```
You:  /rpi-implement .thoughts/plans/2026-03-04-fix-date-formatter.md
```

Claude shows you the intended changes, you approve, it implements, runs tests, and proposes a commit.

**Example -- Add a missing environment variable:**
```
You:  /rpi-plan Add REDIS_URL to the config module with a default of localhost:6379
```

One phase, two files modified, done in minutes.

## Medium Tasks (focused features, single-concern changes)

**Path: Research -> Design -> Plan -> Implement**

Use when the feature touches multiple files, involves a choice between approaches, or you're working in unfamiliar code.

**Example -- Add rate limiting to the API:**

**Step 1: Research**
```
You:  /rpi-research How does the API middleware chain work? Where are requests authenticated and validated?
```
Claude spawns sub-agents to explore your codebase in parallel. Returns a document with every relevant file path, the middleware execution order, existing patterns, and how requests flow through the system.

You review `.thoughts/research/2026-03-04-api-middleware.md` -- maybe you redirect: "Focus more on the error handling middleware, I want rate limit errors to follow that pattern."

**Step 2: Design**
```
You:  /rpi-design .thoughts/research/2026-03-04-api-middleware.md
      I want to add per-endpoint rate limiting. Should handle both authenticated and anonymous users.
```
Claude presents 2-3 options (e.g., in-memory vs Redis, middleware vs decorator pattern), with pros/cons tied to your actual codebase. You pick an approach. It writes the design document with the decision rationale, component diagram, and risk assessment.

**Step 3: Plan**
```
You:  /rpi-plan .thoughts/designs/2026-03-04-api-rate-limiting.md
```
Claude reads the design, spot-checks the codebase against the docs, and breaks the work into phases:
- Phase 1: Rate limiter core module + unit tests
- Phase 2: Middleware integration + integration tests
- Phase 3: Per-endpoint configuration + documentation

Each phase has specific file changes, code snippets, automated verification commands (pulled from your `CLAUDE.md`), and manual verification steps.

**Step 4: Implement**
```
You:  /rpi-implement .thoughts/plans/2026-03-04-api-rate-limiting.md
```
Claude implements Phase 1, shows you a preview of all changes before writing code, runs the test suite, updates checkboxes in the plan, and proposes a commit. Then it pauses for your manual verification before starting Phase 2.

If something doesn't match the plan -- say a file was refactored since the design was written -- Claude stops and tells you exactly what diverged, rather than silently improvising.

## Large Tasks (multi-concern features, major refactors, greenfield projects)

**Path: Research -> Design -> Tickets -> Plan (per ticket) -> Implement (per ticket)**

Use when the work spans multiple systems, would produce more than ~4 implementation phases, or you want to ship incrementally over multiple sessions.

**Example -- Build a notification system (email, push, in-app):**

**Step 1: Research**
```
You:  /rpi-research How do we currently send emails? Is there any notification infrastructure? How do user preferences work?
```
Produces a thorough map of everything notification-adjacent in your codebase.

**Step 2: Design**
```
You:  /rpi-design .thoughts/research/2026-03-04-notification-infrastructure.md
      Build a notification system supporting email, push, and in-app channels. Users should be able to set per-channel preferences.
```
Claude goes into comprehensive mode -- spawns multiple research sub-agents, identifies key design dimensions (channel abstraction, delivery strategy, preference storage, template system), presents options with trade-offs for each, validates that the chosen options compose well together, and writes the full design document.

**Step 3: Tickets**
```
You:  /rpi-tickets .thoughts/designs/2026-03-04-notification-system.md
```
Claude decomposes the design into discrete work units:
```
1. notif-001: Core notification model + channel interface
2. notif-002: Email channel implementation
3. notif-003: Push notification channel
4. notif-004: In-app notification channel + UI
5. notif-005: User preference management
6. notif-006: Notification dispatch service (orchestrator)
```
Each ticket is self-contained with its own scope, acceptance criteria, file list, and extracted design context. An index file shows the dependency graph and recommended implementation order.

**Step 4-5: Plan and Implement (per ticket)**
```
You:  /rpi-plan .thoughts/tickets/notif-001-core-model.md
You:  /rpi-implement .thoughts/plans/2026-03-04-notif-001-core-model.md

You:  /rpi-plan .thoughts/tickets/notif-002-email-channel.md
You:  /rpi-implement .thoughts/plans/2026-03-04-notif-002-email-channel.md

...and so on for each ticket
```

Each ticket becomes its own plan -> implement cycle. You can stop between tickets, come back the next day, and pick up where you left off -- the `.thoughts/` directory preserves all context.

## Greenfield or Major Reorganization

**Path: Research -> Design -> Structure -> Tickets -> Plan -> Implement**

Add the `/rpi-structure` stage when you need to make deliberate decisions about file layout, module boundaries, and interfaces before planning implementation -- typically for new projects with many new files or large-scale reorganizations.

```
You:  /rpi-structure .thoughts/designs/2026-03-04-new-service.md
```

This produces a structure document with every new and modified file, their responsibilities, module boundaries, public interfaces with concrete signatures, and a dependency graph. The structure document then feeds into `/rpi-tickets` or `/rpi-plan`.

## Not Sure Where to Start?

Use `/rpi-research` even when you have a vague idea. It handles both focused questions ("how does auth work?") and open-ended exploration ("what could we improve about error handling?"). For exploratory queries, it surfaces opportunities and trade-offs in an Assessment section and suggests which pipeline path to take next.

```
You:  /rpi-research What could we improve about error handling?
```

## After Implementation

Two optional commands help close the loop:

- **`/rpi-verify`** -- Validates that your implementation matches the design artifacts. Checks completeness, correctness, and coherence. Run it after `/rpi-implement` or anytime you want a second opinion on whether the code matches the plan.
- **`/rpi-archive`** -- Moves completed artifacts to `.thoughts/archive/` to keep the active directory clean. Run it when a feature is fully shipped and you no longer need the research/design/plan documents in the active directories.
