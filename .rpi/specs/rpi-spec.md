---
domain: workflow
feature: rpi-spec
last_updated: 2026-07-17
updated_by: rpi-propose
---

# rpi-spec — goal-ready spec fast path

## Purpose

A primary fast path from a task description or research note to autonomous agent execution: in one pass it produces a full living behavioral spec plus an ephemeral goal envelope — a lightweight work order with a concrete requirements checklist, scope boundaries, machine-checkable verification commands, and a ready-to-paste goal condition for a condition-based agent loop. The envelope replaces the phased plan on this path; verification remains the final gate. Design reasoning is condensed inline by default, escalates to a full design when tradeoffs warrant it, and refuses only for extreme blast radius.

## Scenarios

### One pass yields a living spec and a goal envelope

Given a small, low-risk task description or research note
When the user invokes the goal-spec fast path
Then a full behavioral spec and a goal envelope (requirements checklist, scope boundaries, verification commands, ready-to-paste goal condition) are produced together in a single pass, with no phased plan and no separate design-approval gate

### The emitted goal condition is ready for the agent loop

Given an approved goal envelope
When the user starts the condition-based agent loop with the emitted goal condition
Then the condition fits the loop runner's size limit, names one measurable end state, points at the envelope's verification commands and requirements checklist, states the constraints that must hold en route, and includes a turn-bounding clause

### Upstream decisions are inherited with attribution

Given a research note whose artifact chain carries recorded decisions
When the fast path is invoked on it
Then those decisions are carried into the produced artifacts verbatim, each attributed to the artifact that recorded it, and the research note is marked complete when fully addressed

### Genuine tradeoffs escalate to a full design inline

Given work that surfaces multiple defensible approaches or spans several components
When the user invokes the fast path
Then a full design artifact is produced inline — after a brief tradeoff checkpoint with the user — before the spec and envelope are drafted, rather than refusing the work

### Extreme blast radius still refuses

Given work whose blast radius is extreme (wide-reaching restructuring across many areas)
When the user invokes the fast path, with or without fast-forward
Then it declines to produce artifacts, explains why in one or two sentences, and redirects to the full design path

### Progress survives interruption

Given a goal run that checked off some envelope requirements before the session ended
When a new session resumes the work
Then the envelope's checklist reflects the prior progress and the loop continues from the first unchecked requirement instead of restarting

### Verification remains the final gate

Given the goal condition has been met and the loop has stopped
When the user follows the suggested next step
Then the verification stage is suggested — not run automatically — and it validates the implementation against the living spec

### The living spec stays clean after completion

Given a completed and verified goal run
When the goal envelope is archived
Then the behavioral spec remains in the living specs collection unchanged, carrying no task-scoped goal state — no verification commands, turn budgets, or checklists

## Constraints

- The emitted spec is full grade — 5-8 user-observable Given/When/Then scenarios plus Constraints and Out of Scope — and obeys the existing rule that scenarios never reference internal structure; all concreteness lives in the envelope.
- The goal condition always fits the loop runner's limit (4,000 characters) and always includes a bounding clause.
- The envelope is ephemeral: it transitions to complete after verification passes and is then archivable; the spec is living and has no completion lifecycle.
- The fast path composes with fast-forward (skip checkpoints and the approval gate, auto-accept, print the condition) and with grilling (single-pass interrogation before the gate); the two modes are mutually exclusive, and the extreme-blast-radius refusal fires even under fast-forward.
- The fast path never starts the agent loop itself — it hands off a ready condition for the user to run.
- The condition-based loop is the documented consumer; interval-based recurrence is at most a secondary unattended pattern.

## Out of Scope

- Running or supervising the agent loop itself — loop mechanics belong to the host tool.
- Replacing the full design path or the phased-plan path for high-stakes work — both remain available and recommended there.
- Changing the verification stage's behavior — it stays the final gate as-is.
- Migrating existing phased plans into goal envelopes.
- Automated test generation from scenarios.
- Guaranteeing loop completion — the envelope makes completion checkable, not inevitable.
