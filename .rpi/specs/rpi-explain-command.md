---
domain: rpi-explain command
id: EX
last_updated: 2026-03-21T23:53:38+01:00
status: active
updated_by: .rpi/designs/2026-03-21-rpi-explain-command-for-post-implementation-walkthroughs.md
---

# rpi-explain command

## Purpose

A slash command that generates a diff-scoped walkthrough of an implemented solution, explaining what changed and why — with focus on non-obvious decisions. Optionally saves the explanation as an artifact.

## Behavior

### Input resolution
- **EX-1**: When given an artifact path, resolve its chain (plan → design → research) and use as context for explanations
- **EX-2**: When given no arguments, auto-detect changed files from git and proceed without artifact context
- **EX-3**: When given a path that doesn't exist or has no linked artifacts, proceed with diff-only explanation and note the missing context

### Diff walkthrough
- **EX-4**: Walk through changes file-by-file, providing a factual summary of what changed in each file
- **EX-5**: For non-obvious changes, provide explicit callouts explaining the reasoning — inferred from artifacts when available, from code context otherwise
- **EX-6**: Clearly distinguish between rationale sourced from artifacts vs inferred from code context
- **EX-7**: Summarize straightforward changes briefly (1-2 sentences) rather than explaining the obvious

### Artifact saving
- **EX-8**: Do not save an artifact by default — only when the user explicitly requests it
- **EX-9**: When saving, use `.rpi/reviews/` directory with a descriptive filename

## Constraints

### Must
- Include file:line references in all explanations
- Read all changed files fully before generating explanations
- Prioritize non-obvious changes over boilerplate/mechanical changes

### Must Not
- Produce pass/fail judgments or severity ratings (that's `/rpi-verify`)
- Auto-save artifacts without user request
- Hallucinate rationale — flag uncertainty when inferring without artifact backing

### Out of Scope
- New RPI CLI commands or Go code changes
- Automated triggering from other commands
- Branch/commit comparison selection (uses default git changed-files)

## Test Cases

### EX-1: Artifact chain resolution
- **Given** a plan path with linked design and research **When** `/rpi-explain .rpi/plans/foo.md` **Then** all linked artifacts are read and referenced in the walkthrough

### EX-2: No arguments
- **Given** git has changed files vs main **When** `/rpi-explain` **Then** changed files are detected and walkthrough is generated without artifact context

### EX-5: Non-obvious change callout
- **Given** a diff with a non-trivial refactor **When** explanation is generated **Then** the non-obvious parts get explicit callouts with reasoning

### EX-6: Source attribution
- **Given** a plan describes a design decision **When** the walkthrough references that decision **Then** it attributes the rationale to the plan artifact

### EX-8: No auto-save
- **Given** user runs `/rpi-explain` **When** explanation completes **Then** no artifact file is created unless user explicitly requests it
