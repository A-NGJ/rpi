---
archived_date: "2026-04-02"
branch: main
date: 2026-03-21T23:59:00+01:00
git_commit: debf4cf
repository: ai-agent-research-plan-implement-flow
researcher: Claude
status: archived
tags:
    - research
topic: rpi-diagnose command for iterative bug analysis
---

# Research: rpi-diagnose command for iterative bug analysis

## Research Question

How should we add a command for deep, iterative bug diagnosis that fills the gap between shallow research and solution-assuming planning?

## Problem Statement

The current pipeline has no command for complex bug analysis. `/rpi-research` interviews but stays shallow. `/rpi-plan` assumes you already know the fix. `/rpi-propose` is for feature design. There's no command that: deeply investigates root cause → attempts a fix → verifies it worked → iterates if not → produces a lightweight artifact summarizing what was found and fixed.

## Summary

A new `/rpi-diagnose` command is needed. It combines research's interview flow with implement's execution loop, adding an autonomous iterate-until-fixed cycle that no existing command has. It fits SDD by framing bugs as "spec deviation analysis" — actual behavior deviates from expected behavior, and the command closes that gap.

## Detailed Findings

### Pipeline Gap

The pipeline currently recommends "Bug fix or small change → Plan → Implement" (`.rpi/PIPELINE.md:177-179`). This works for known bugs but fails for complex ones requiring root cause analysis before a fix can be attempted.

### Command Structure Pattern

All existing commands follow a consistent structure (`.claude/commands/rpi-*.md`):
- **Frontmatter**: description, model: inherit, disable-model-invocation: true
- **Goal**: what it does + pipeline position
- **Invariants**: hard rules (must-do behaviors)
- **Principles**: soft guidance

### Unique Characteristics of Diagnose

1. **Hybrid command**: combines research (understand the bug) with implement (fix + verify). No existing command does both.
2. **Autonomous iteration loop**: try → check → retry. Research stops at findings. Implement follows a plan. This is the only command that adapts mid-execution.
3. **Lightweight artifact**: produces a root cause + fix summary, not a spec or design document.
4. **Escalation path**: if the fix turns out to be too complex, escalates to `/rpi-plan` rather than forcing a solution.

### SDD Alignment

- Bugs are spec deviations — actual behavior differs from expected behavior
- The command requires stating expected vs actual behavior upfront, implicitly referencing a spec
- The artifact documents why the deviation occurred and how it was resolved
- Risk: scope creep if used for vague "something feels wrong" bugs with no clear expected behavior

### No Existing Commands Modified

This is purely additive — a new command file at `.claude/commands/rpi-diagnose.md` and updates to `.rpi/PIPELINE.md` and `CLAUDE.md` to reference it.

## Assessment

The command fills a genuine gap and fits naturally into SDD when framed as spec deviation analysis. The main design challenge is the autonomous iteration loop — deciding when to stop iterating and when to involve the user. The escalation path to `/rpi-plan` keeps it from becoming an unbounded debugging session.

## Suggested Next Steps

Proceed to `/rpi-propose .rpi/research/2026-03-21-rpi-diagnose-command-for-iterative-bug-analysis.md` to design the command's structure, define its flow, and create a spec.

## Decisions

- **Command name**: `/rpi-diagnose` — emphasizes root cause analysis, can escalate to `/rpi-plan` if fix is complex
- **Artifact type**: lightweight post-mortem (root cause + fix summary), not a spec or design
- **SDD framing**: spec deviation analysis — requires expected vs actual behavior upfront
