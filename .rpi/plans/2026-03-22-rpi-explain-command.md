---
date: 2026-03-22T21:57:43+01:00
design: .rpi/designs/2026-03-21-rpi-explain-command-for-post-implementation-walkthroughs.md
spec: .rpi/specs/rpi-explain-command.md
status: active
tags:
    - plan
topic: rpi-explain command
---

# rpi-explain command — Implementation Plan

## Overview

Add a `/rpi-explain` slash command that generates diff-scoped walkthroughs of implemented solutions. Prompt-only — no Go code changes.

**Scope**: 1 new file, 1 modified file

## Source Documents
- **Design**: .rpi/designs/2026-03-21-rpi-explain-command-for-post-implementation-walkthroughs.md
- **Spec**: .rpi/specs/rpi-explain-command.md

## Phase 1: Create `/rpi-explain` command and update pipeline docs

### Overview
Create the slash command file following existing command conventions (frontmatter + Goal + Invariants + Principles). Update CLAUDE.md to mention explain as an optional post-implement step.

### Tasks:

#### 1. Slash command
**File**: `.claude/commands/rpi-explain.md`
**Changes**: Create new command file with:
- Frontmatter: description, model: inherit, disable-model-invocation: true
- Goal section: diff-scoped walkthrough of implemented changes
- Invariants covering: input resolution (EX-1, EX-2, EX-3), diff walkthrough (EX-4, EX-5, EX-6, EX-7), artifact saving (EX-8, EX-9)
- Principles section: explain don't judge, attribute sources, scale to diff size

#### 2. Pipeline documentation
**File**: `CLAUDE.md`
**Changes**: Add `/rpi-explain` as optional post-implement step in the pipeline description

### Success Criteria:

#### Manual Verification:
- [x] Command file follows the same structure as `rpi-verify.md` and `rpi-research.md` (frontmatter + Goal + Invariants + Principles)
- [x] All spec behaviors EX-1 through EX-9 are covered by invariants
- [x] CLAUDE.md pipeline description mentions explain
- [ ] Run `/rpi-explain` on current working tree — produces a meaningful diff walkthrough

### Commit:
- [x] Stage: `CLAUDE.md`, `internal/workflow/assets/commands/rpi-explain.md`, `internal/workflow/assets/templates/CLAUDE.md.template`
- [x] Message: `feat: add /rpi-explain command for post-implementation walkthroughs`

---

## References
- Design: .rpi/designs/2026-03-21-rpi-explain-command-for-post-implementation-walkthroughs.md
- Spec: .rpi/specs/rpi-explain-command.md
