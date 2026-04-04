---
domain: rpi-explain command
feature: rpi-explain
last_updated: 2026-04-04T22:30:00+02:00
updated_by: .rpi/designs/2026-03-21-rpi-explain-command-for-post-implementation-walkthroughs.md
---

# rpi-explain command

## Purpose

A slash command that generates a diff-scoped walkthrough of an implemented solution, explaining what changed and why — with focus on non-obvious decisions. Optionally saves the explanation as an artifact.

## Scenarios

### Explain with artifact chain context
Given a plan path with linked design and research artifacts
When the user runs `/rpi-explain .rpi/plans/foo.md`
Then all linked artifacts are read and used as context, and the walkthrough references rationale from those artifacts

### Explain with no arguments
Given git has changed files compared to the base branch
When the user runs `/rpi-explain` with no arguments
Then changed files are auto-detected and a walkthrough is generated without artifact context

### Explain gracefully handles missing artifacts
Given a path that doesn't exist or has no linked artifacts
When the user runs `/rpi-explain` with that path
Then a diff-only explanation is generated with a note about the missing context

### Non-obvious changes are highlighted
Given a diff containing both trivial and non-trivial changes
When the explanation is generated
Then non-obvious changes get explicit callouts with reasoning while straightforward changes are summarized briefly

### Rationale is attributed to its source
Given a design artifact describes a specific decision
When the walkthrough references that decision
Then it clearly distinguishes whether the rationale comes from an artifact or is inferred from code context

### Artifacts saved only on explicit request
Given the user runs `/rpi-explain` and the explanation completes
When no explicit save request is made
Then no artifact file is created in `.rpi/reviews/`

## Constraints
- Include file:line references in all explanations
- Read all changed files fully before generating explanations
- Do not produce pass/fail judgments or severity ratings (that's `/rpi-verify`)
- Do not hallucinate rationale — flag uncertainty when inferring without artifact backing

## Out of Scope
- New RPI CLI commands or Go code changes
- Automated triggering from other commands
- Branch/commit comparison selection (uses default git changed-files)
