---
domain: skill-descriptions
feature: skill-descriptions
last_updated: 2026-05-06T10:44:18+02:00
updated_by: .rpi/designs/2026-05-06-skill-descriptions.md
---

# skill-descriptions

## Purpose

Define the quality contract for the `description` field of every RPI skill so the skill is reliably triggered from natural conversation, not just slash commands. The contract covers structure (what each description must contain), voice (consistent across skills), disambiguation (between overlapping skills), and a measured acceptance threshold.

## Scenarios

### Description states a user goal in its first sentence
Given any RPI skill description
When the first sentence is inspected
Then it describes what the user gets — phrased as the user's goal — and does not describe the skill's mechanism, output format, or internal artifacts

### Description specifies when to invoke
Given any RPI skill description
When the description is inspected
Then it contains a sentence beginning with `Use when` that lists user-facing situations the skill should fire on

### Description includes verbatim user phrasings
Given any RPI skill description for which real user vocabulary has been sampled
When the trigger sentence is inspected
Then it quotes at least three user phrasings inside single or double quotes

### Overlapping skills include explicit negative gates
Given two RPI skills that share natural-language surface area — specifically `rpi-plan`↔`rpi-propose`, `rpi-research`↔`rpi-propose`, and `rpi-research`↔`rpi-diagnose`
When each skill's description is inspected
Then it contains a sentence of the form `Do NOT invoke for X — use rpi-Y instead`, where rpi-Y is the named sibling

### Voice is consistent across all RPI skills
Given the descriptions of all RPI skills
When their trigger sentences are inspected
Then every one uses the imperative form `Use when …` — none use third-person variants such as `This skill should be used when …`

### Description meets the measured trigger threshold
Given a rewritten description and an eval set of 20 realistic queries with a 60% train / 40% test split
When the `skill-creator` description optimizer runs over the eval set with the test split held out
Then the test-split trigger rate is at least 80%

### Eval query set is persisted alongside the description
Given a skill whose description was tuned via the optimizer
When the rewrite is merged
Then the 20 queries used for the eval are stored as a CSV file under the design's eval directory so the same queries can be replayed when the description is later re-tuned

## Constraints
- Source of truth for skill files is `internal/workflow/assets/skills/<name>/SKILL.md`. Deployed copies under `.claude/skills/` are not edited directly.
- Each rewrite changes only the `description:` line in frontmatter. The body of every SKILL.md remains byte-identical to its pre-rewrite content.
- Compatibility with the existing Agent Skills format spec (`agent-skills.md`) remains intact: every skill still has `name` and `description`, every name still matches its parent directory and the naming regex.
- Voice convention is imperative (`Use when …`), applied uniformly across all RPI skills.
- Acceptance threshold: ≥80% trigger rate on the optimizer's test split.
- The measurement instrument is Anthropic's `skill-creator` description optimizer using its default 60/40 train/test split, 5-iteration cap, and 3 trials per query.

## Out of Scope
- Triggering behavior of non-RPI skills installed alongside the RPI suite.
- Quality or content of SKILL.md bodies (this spec covers descriptions only).
- Adding, removing, or renaming skills.
- CLAUDE.md guidance that influences activation outside the description signal.
- Hooks, harness signals, or plugin-marketplace metadata.
- Automated CI integration of the optimizer (manual runs before merge are sufficient for v1).
