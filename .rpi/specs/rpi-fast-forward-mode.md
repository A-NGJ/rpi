---
domain: rpi-propose, rpi-plan, rpi-implement, and rpi-verify chaining
feature: rpi-fast-forward-mode
last_updated: 2026-04-29T23:44:59+02:00
updated_by: .rpi/designs/2026-04-29-fast-forward-mode-for-rpi-implement.md
---

# --ff (fast-forward) mode for the rpi pipeline

## Purpose

An opt-in autopilot mode that suppresses the user-approval gates in
`/rpi-propose`, `/rpi-plan`, and `/rpi-implement` and auto-chains into the
next pipeline stage, ending with a verification report from `/rpi-verify`.
The user trades review for speed when the task is small, the defaults are
trusted, or agentic operation is acceptable. Automated integrity and safety
checks still run and can stop the chain.

## Scenarios

### Fast-forward through implement runs all phases without per-phase pauses
Given the user invokes `/rpi-implement --ff <plan>`
When implementation begins
Then the per-phase intended-changes preview and manual verification pauses are skipped, and phases run back-to-back until the plan is complete or a safety gate stops the run

### Fast-forward through propose auto-accepts and chains to plan
Given the user invokes `/rpi-propose --ff <input>`
When the design and spec are drafted
Then the trade-off buy-in, mid-flight checkpoints, and spec approval gates are skipped, the artifacts are saved with the standard status transitions applied, and `/rpi-plan --ff <design-path>` is invoked automatically

### Fast-forward through plan auto-accepts the outline and chains to implement
Given the user invokes `/rpi-plan --ff <design>`
When the phase outline is ready
Then the buy-in gate is skipped, the full plan is written, and `/rpi-implement --ff <plan-path>` is invoked automatically

### Fast-forward chains into verify after implement completes
Given a fast-forward run reaches the end of `/rpi-implement` successfully
When the plan transitions to complete
Then `/rpi-verify <plan-path>` is invoked automatically and produces a verification report in `.rpi/reviews/` as the terminal artifact of the run

### Safety and integrity gates still stop the chain
Given a fast-forward run is in progress
When an automated check detects codebase drift, a dropped design decision, sensitive content in a staged file, or any "On mismatch" condition
Then the chain stops at that point, the user is informed, and they decide whether to resume manually

### Fast-forward and grill modes are mutually exclusive
Given the user passes both `--ff` and `--grill` in a single invocation
When the skill begins
Then it errors immediately and asks the user to pick one mode, without taking any other action

### Fast-forward requires the explicit flag
Given the user uses phrasing like "autopilot", "no pauses", or "just run it" without passing `--ff`
When the skill runs
Then the standard collaborative gates run unchanged — natural-language phrasing alone never triggers fast-forward

### Fast-forward is not available in research
Given the user invokes `/rpi-research --ff <topic>`
When the skill runs
Then `--ff` is treated as ordinary input (or ignored), research proceeds normally with no auto-chain to propose, and no autopilot semantics are applied

## Constraints

- `--ff` is invocation-scoped: it is not persisted in artifact frontmatter, so re-invocations on the same artifact require the flag to be passed again
- `--ff` and `--grill` cannot be combined in one invocation
- `--ff` is propagated by every skill in the chain — when a skill chains into the next, it includes `--ff` in the chained invocation's arguments
- `--ff` does not modify status transitions; existing per-skill rules apply unchanged whether the chain completes or breaks
- The codebase-drift "On mismatch" gate, plan-time drift spot-check, design-coverage check, and sensitive-content scan are preserved under `--ff` — they are integrity/safety checks, not review gates
- Phase-level automated check failures (tests, build, lint) still stop the run normally; `--ff` does not affect failure handling
- If `--ff` is passed without input, the skill follows the existing "nothing provided" branch (asks for input); the flag is preserved for the next invocation

## Out of Scope

- A partial fast-forward (e.g., skip pre-review but keep manual verification) — `--ff` is monolithic
- Auto-chaining `/rpi-research` into `/rpi-propose` under `--ff`
- Auto-running `/rpi-commit`, `/rpi-spec-sync`, or any other post-verify skill under `--ff`
- Persisting `--ff` state across re-invocations or in plan/design metadata
- Natural-language triggers for `--ff` (must be the explicit flag)
- Adding `--ff` to skills outside the propose → plan → implement → verify chain
