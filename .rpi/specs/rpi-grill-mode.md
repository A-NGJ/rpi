---
domain: rpi-propose and rpi-plan approval gates
feature: rpi-grill-mode
last_updated: 2026-05-01T00:00:00+02:00
updated_by: .rpi/plans/2026-05-01-bundle-grill-me-skill-from-mattpocock-skills.md
---

# --grill mode for rpi-propose and rpi-plan

## Purpose

An opt-in mode that hands off `/rpi-propose` and `/rpi-plan` approval gates to
the bundled `grill-me` skill (sourced from
[mattpocock/skills](https://github.com/mattpocock/skills) under MIT) for
adversarial, one-question-at-a-time interrogation before the user accepts the
artifact. `grill-me` ships by default with `rpi init` / `rpi update`; the
fall-back path remains for users who have removed it from their local
installation.

## Scenarios

### Grill-me is available by default
Given a fresh `rpi init` (or `rpi update`) with no manual skill removal
When the user invokes `/rpi-propose --grill` or `/rpi-plan --grill`
Then the bundled `grill-me` skill is available and grilling proceeds without the soft fall-back warning

### Grill triggered via flag
Given the user invokes `/rpi-propose --grill <input>` or `/rpi-plan --grill <input>`
And the `grill-me` skill is available
When the draft reaches the approval gate
Then `grill-me` is invoked to interrogate the draft before approval

### Grill triggered via natural language
Given the user includes phrasing like "grill me on this" or "stress-test this design" alongside their request
And the `grill-me` skill is available
When the draft reaches the approval gate
Then `grill-me` is invoked, identical to the flag form

### Propose grills design and spec together
Given `--grill` mode is active in `/rpi-propose`
When grilling fires
Then it runs once, after both the design and the spec are drafted, treating them as a single unit

### Plan grills the phase outline
Given `--grill` mode is active in `/rpi-plan`
When grilling fires
Then it runs on the phase outline before the full plan is written, mirroring the existing buy-in gate

### Grill-me unavailable, soft fall-back
Given the user has removed the bundled `grill-me` skill or runs in an environment where it is not available
And the user requests grilling
When the approval gate is reached
Then the user is told `grill-me` is not currently available, offered the standard non-grill approval, and the flow proceeds only after explicit confirmation

### Findings shape the artifact inline
Given `grill-me` has just finished interrogating a draft
When the rpi skill applies the resulting revisions
Then the changes appear directly in the design, spec, or phase outline — no separate "interrogation notes" appendix or audit trail file is added

### Standard flow when grill is not requested
Given the user does not pass `--grill` and uses no grill-style phrasing
When the approval gate is reached
Then the existing collaborative approval gate runs unchanged and `grill-me` is never invoked

## Constraints

- `grill-me` is invoked at most once per `--grill` invocation (single-pass, not looped)
- The fall-back path requires explicit user confirmation before proceeding without grilling
- The existing approval gate still runs after grilling — `grill-me` does not bypass approval

## Out of Scope

- Grilling on artifacts produced by skills other than `/rpi-propose` and `/rpi-plan` (no grill mode in research, implement, verify, etc.)
- Recording `grill-me`'s questions and answers as a separate artifact
- Looping `grill-me` until "no more questions"
