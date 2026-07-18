---
name: rpi-spec
description: Go from a task description or research note straight to a living behavioral spec plus a goal envelope — a ready-to-run work order for a condition-based agent loop — in one pass, with no separate design to review and no phased plan. Use when user says 'make a goal-ready spec', 'spec this for an agent loop', 'turn this into a goal for /goal', or 'one-shot a spec and goal from this'. Do NOT invoke for genuine tradeoffs the user wants to review (use rpi-propose), a narrow scoped change wanting a phased plan (use rpi-plan), or a phased plan without a design (use rpi-blueprint).
---

# Goal-Ready Spec Fast Path

## Goal

Take a task description or research note to two artifacts in a single pass: a
full-grade living behavioral spec and an ephemeral goal envelope — a work order
carrying a requirements checklist, scope with file paths, verification commands,
and a ready-to-paste `/goal` condition for a condition-based agent loop. The
envelope replaces the phased plan on this path; the living spec is the
centerpiece deliverable. The skill never starts the loop — it hands off a
condition for the user to run.

Auto-detect the input:
- **Task description** (plain text small enough to reason about in one pass)
- **Research artifact** (path to a research doc) → resolve its chain, reason inline
- **Nothing provided** → ask for either, with brief examples of each

On success, suggest → `/rpi-verify <spec-path>` as the final gate once the goal clears.

## Invariants

- See the project's RPI Skill Contract for `--ff` / `--grill` semantics; both apply here and are mutually exclusive
- **Gate judgment (runs before drafting, even under `--ff`)**: judge blast radius. Extreme (wide restructuring across many areas) → decline in 1-2 sentences and redirect to the full design path (`/rpi-propose <input>`); this refusal fires even under `--ff`. Genuine tradeoffs or multi-component signals → produce a full design artifact inline after a brief tradeoff checkpoint, then continue. Otherwise → condensed `## Design Notes` in the envelope
- If a research artifact is the input, read it fully and resolve its chain — warn if it is still draft or already complete
- Before drafting, search for prior specs, designs, and research on this topic (semantic search with keyword-scan fallback, per the cross-skill contract); collect upstream `## Decisions` for verbatim, per-source inheritance into the envelope (and the design, when one is produced)
- Ground in the codebase before drafting — file:line evidence feeds the envelope's Scope and Verification
- **Living spec (full grade)**: 5-8 user-observable Given/When/Then scenarios plus Constraints and Out of Scope; scenarios never reference internal structure (structs, file paths, function names); name the spec file after its `feature` field
- **Goal envelope**: a requirements `- [ ]` checklist the executing agent updates as it works; a Scope section with file paths and explicit must-not-change boundaries; Verification commands with expected results; a goal condition ≤4,000 chars naming one measurable end state, pointing at the checklist and verification commands, stating the constraints that must hold en route, and carrying a turn-bounding clause
- Get buy-in on the spec + envelope before finalizing; iterate until accepted. Under `--ff`, skip the gate, auto-accept, and print the condition. Under `--grill` (when `grill-me` is available), interrogate the drafts single-pass before the gate; if unavailable, say so and ask whether to proceed
- Scaffold and save both artifacts, linking the spec in the envelope frontmatter (and the upstream research, when one was the input)
- Transition the research artifact to complete when it was the input and fully addressed; the envelope starts active; the spec is living (no status lifecycle)
- **Handoff**: print the ready-to-paste `/goal` condition and state that verification is the suggested (not auto-run) next step after the goal clears — never start the loop yourself

## Principles

- The spec is the centerpiece — full grade, a living contract; the envelope is its ephemeral work order
- Concreteness lives only in the envelope — the spec stays abstract so it survives as a clean behavioral contract
- Escalate when it makes sense, refuse only at the extreme — condensed notes by default, a real design inline for tradeoffs, the slow path only for wide blast radius
- Hand off, never drive — the loop is the user's to start; the skill readies the condition and stops

**Recommended model:** premium tier, high effort — spec + envelope authoring ranks with propose. Advisory; see `docs/model-routing.md`.
