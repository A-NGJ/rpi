---
name: rpi-revise
description: Amend an existing implementation plan in .rpi/plans when a new constraint or a review finding lands after the plan is drafted or partly built. Use when user says 'the constraint changed, update the plan', 'revise the plan for X', 'a review found a gap, amend the plan', or 'the API changed, the plan needs to account for it'. Do NOT invoke to create a fresh self-contained scoped change from scratch — use rpi-plan instead.
---

# Revise Plan

## Goal

Amend an existing plan in `.rpi/plans/` when a new constraint or a review finding lands after the plan is drafted or partially implemented. Edit only the affected phases, preserve the checkbox state of everything else, re-audit just what changed, and hand back to implementation.

Fires to amend an *existing* plan. A fresh, self-contained scoped change is the planning skill's job — steer those to it (the negative gate names `rpi-plan`). A change that invalidates the upstream design routes back through the design stage, not here.

After a successful revision, suggest → `/rpi-implement <plan-path>`, which resumes from the first unchecked item.

## Invariants

- See the project's RPI Skill Contract for `--ff` / `--grill` semantics; both flags apply here and are mutually exclusive
- Validate the target plan exists and read it fully, plus its upstream chain (design, spec, research)
- Capture a *before* snapshot of the plan's checked / unchecked items, each with its Phase and `**File**:` context, using the completeness oracle — this is the preservation baseline
- Identify the changed-phase set the request touches; present it for buy-in before writing (creation-style approval, but scoped). Under `--ff`, skip this approval gate
- Edit only affected phases; leave unaffected phases byte-stable so their state is untouched by construction
- Re-key affected items by their normalized text (not line position) so reordering preserves `[x]`; new items are added unchecked; removed items disappear with their state
- Re-run the slice / pre-lock audit — the planning skill's audit, applied to the changed phases only, not a reimplementation. Unchanged phases keep their prior audit standing. A renumbering seam counts as changed so the ordered-phase invariant is re-checked across it
- Capture an *after* snapshot and assert no item that was `[x]` is now `[ ]`. If a change would reopen completed work, stop and show exactly which items, and require explicit confirmation — in every mode including fast-forward (this is a safety invariant, not an approval gate)
- Status handling via the validated status transition: a `draft` plan stays `draft` and an `active` plan stays `active`, amended in place. A `complete` plan is never silently reopened — offer two explicit choices: a guarded, user-confirmed reopen, or supersede it and hand off to the planning skill to carry the unchanged phases into a successor plan. Under `--ff`, a complete plan stops and reports rather than guessing
- Under `--grill` (when `grill-me` is available), stress-test the proposed changed-phase amendment before writing, apply revisions inline, then proceed through the standard gate; if unavailable, say so and offer the standard gate
- Under `--ff`, after a successful revision invoke `/rpi-implement --ff <plan-path>` via the Skill tool, resuming from the first unchecked item — but never skip the completed-work assertion
- For the review-driven loop, accept a verification finding as the change request, amend the affected phase, and hand back to implement — closing the verify → revise → implement loop without a fresh design pass
- If the scoped changed-phase audit capability is not yet available, fall back to re-auditing the changed phases with the full plan audit and say so, rather than failing

## Principles

- Preserve completed work above all — the before/after snapshots are the evidence, not recollection
- Scope tightly — re-examine only what the change touches; the renumbering seam is the only sanctioned widening
- Amend, don't recreate — a plan that has grown past the split threshold defers to the planning skill's split flow rather than splitting mid-revision
- Stay coupled by reference — the audit lives in the planning skill; inherit it, don't fork it

**Recommended model:** premium tier, high effort — re-plans affected phases with dependency/ordering re-checks. Advisory; see `docs/model-routing.md`.
