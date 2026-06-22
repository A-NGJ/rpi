---
name: rpi-blueprint
description: Go from a research note or a short problem statement straight to a phased plan in one pass, without a separate design to review. Use when user says 'just get me to a plan', 'skip the design, plan it', 'one-shot this from the research note', or 'I don't need a design, just plan this'. Do NOT invoke for genuine tradeoffs, multiple defensible approaches, or high-blast-radius changes — use rpi-propose instead; for a narrow scoped change to existing behavior use rpi-plan.
---

# Fused Research-to-Plan Blueprint

## Goal

Take a research note or a short problem statement to a phased implementation plan
in a single pass: perform condensed design reasoning inline and emit the plan
directly, without producing a separate reviewable design artifact. Always emit a
minimal behavioral spec (the SDD floor) and record the design reasoning that would
have lived in a design file as a `## Design Notes` block inside the plan.

This is the fused shortcut — structurally omitting the design *deliverable* — as
distinct from fast-forward mode, which runs the full pipeline fast but still
produces a design.

Auto-detect the input:
- **Research artifact** (path to a research doc) → resolve its chain, reason inline, emit plan + minimal spec
- **Short problem statement** (plain text small enough to reason about in one pass) → condensed inline reasoning, then emit plan + minimal spec
- **Nothing provided** → ask for a research path or a short problem statement, with brief examples of each

On success, suggest → `/rpi:rpi-implement <plan-path>` (or, under `--ff`, the chain has already continued there).

## Invariants

- See the project's RPI Skill Contract for `--ff` / `--grill` semantics; both flags apply here and are mutually exclusive
- **Refuse-and-redirect hard gate (runs first, even under `--ff`)**: before writing anything, judge whether the work carries genuine tradeoffs, more than one approach a reasonable engineer would defend, or high blast radius. Apply the split-score *signals* — component count, directory spread, multi-spec — to the problem statement / research as one input; scaffold a throwaway design to score only when the call is genuinely ambiguous. If any hold, decline to produce a plan, explain in one or two sentences why the work needs a reviewable design, and suggest `/rpi:rpi-propose <input>`. This gate is an integrity check, not a review pause: it fires even under `--ff` and stops the chain — never silently chain into `/rpi:rpi-propose --ff`. If neither a research artifact nor a problem small enough to reason about in one pass is supplied, ask for a research pass or redirect to `/rpi:rpi-propose`
- If a research artifact is provided, read it fully and resolve its full artifact chain — warn if it is still draft or already complete
- Before drafting, search for prior plans, designs, and specs on this topic; if a prior plan covers the same scope, ask whether to extend it
- Check `.rpi/specs/` for specs covering the affected area — the plan and the minimal spec must satisfy (and must not contradict) these behavioral contracts
- **Condensed design reasoning inline**: commit to the obvious approach and weigh it against the obvious alternative(s) in one pass — no parallel option-exploration fan-out. The fused path does *less* exploration than `rpi-propose`; if exploration surfaces a second defensible approach, that is the refuse signal, not a reason to keep exploring
- **`## Design Notes` block is mandatory** near the top of the produced plan: the chosen approach, the alternative(s) considered and dropped, and the blast-radius judgment that justified blueprint over propose. The fused path omits the design *deliverable*, never the design *reasoning*
- **Minimal spec floor**: always write and link a behavioral spec describing user-observable behavior — 3-5 Given/When/Then scenarios (fewer than a full propose spec, reflecting the small scope that qualifies for this path), plus a Constraints section and an Out of Scope section. Scenarios must not reference internal structure (structs, file paths, function names). Name the spec file after its `feature` field. A plan with no linked spec is a contract violation
- Break the change into ordered phases — each leaves the codebase working and testable; each phase has tasks with file paths, success criteria (automated + manual), and a commit step; include tests in the same phase as the code they test. Map phases to spec scenarios where applicable
- When drafting each phase's Stage list, exclude paths matching `.gitignore` rules — gitignored artifacts (commonly the plan and spec files themselves under the default tracked-specs policy) must not appear in commit instructions
- Get buy-in on the condensed design reasoning and the proposed phases before writing the full plan
- Under `--ff`, skip the plan-outline approval pause — write the plan + minimal spec immediately, then invoke `/rpi:rpi-implement --ff <plan-path>` via the Skill tool and finally `/rpi:rpi-verify <plan-path>` once at the end, propagating `--ff`. The refuse gate above still stops the chain
- Under `--grill` (or matching natural-language phrasing) and when `grill-me` is available, invoke `grill-me` on the condensed design reasoning + phase outline in a single pass before writing the plan; apply revisions inline. If `grill-me` is unavailable, tell the user and ask whether to proceed with the standard approval gate
- Scaffold and save the plan artifact, linking the minimal spec in frontmatter (and the upstream research, when one was the input)
- Transition the research artifact to complete when it was the input and is fully addressed
- On success, suggest the next step (`/rpi:rpi-implement <plan-path>`), consistent with every other stage

## Principles

- Commit to the obvious approach and record why — blueprint does condensed reasoning, not option fan-out; genuine ambiguity is a refuse signal, not a planning input
- Minimal is a floor, not a license — the spec is smaller than a propose spec, never absent
- Omit the deliverable, keep the reasoning — `## Design Notes` makes the dropped design auditable so `/rpi:rpi-verify` has something to check against
- Blueprint is fused deliverables; `--ff` is suppressed pauses — they compose cleanly, but they are not the same axis
- Grilling is opt-in and single-pass — re-invoke if a second round is needed

**Recommended model:** premium tier, high effort — fused research → design → plan reasoning in one pass. Advisory; see `docs/model-routing.md`.
