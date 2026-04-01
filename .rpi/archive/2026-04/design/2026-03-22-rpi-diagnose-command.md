---
archived_date: "2026-04-02"
date: 2026-03-22T12:00:00+01:00
research: .rpi/research/2026-03-21-rpi-diagnose-command-for-iterative-bug-analysis.md
status: archived
tags:
    - design
topic: rpi-diagnose command for iterative bug analysis
---

# Design: /rpi-diagnose Command

## Problem

The RPI pipeline has no command for complex bug analysis. `/rpi-research` interviews but stays shallow and doesn't attempt fixes. `/rpi-plan` assumes you already know the fix. `/rpi-propose` is for feature design. There's no command that deeply investigates root cause, attempts a fix, verifies it, and iterates if it didn't work.

## Solution

A new `/rpi-diagnose` command that combines research's interview flow with implement's execution loop, adding an autonomous iterate-until-fixed cycle. It frames bugs as spec deviation analysis: actual behavior deviates from expected behavior, and the command closes that gap.

## Design Decisions

### D1: Artifact location — new `.rpi/diagnoses/` directory

Diagnoses are fundamentally different from verification reports. Reviews validate known specs; diagnoses investigate unknowns. A separate directory keeps semantics clean. Artifact format is a lightweight post-mortem: root cause, what was tried, what fixed it.

### D2: Semi-autonomous iteration with escalation — 3 attempts then checkpoint

The fix-iterate loop runs autonomously for up to 3 attempts. Each attempt: apply fix, run relevant tests, evaluate results. After 3 failed attempts, checkpoint with the user to reassess the approach. This prevents runaway debugging sessions while keeping simple fixes fast.

### D3: No spec reference required — but expected-vs-actual is mandatory

Requiring a spec reference would make the command unusable for bugs in unspecced areas. Instead, the first step is always clarifying "what should happen" vs "what actually happens," even if no formal spec exists. This preserves the SDD framing without creating a hard dependency.

### D4: Escalation path — `/rpi-plan` by default, `/rpi-propose` for architectural changes

When a fix is too complex for inline patching, escalate to `/rpi-plan` since the root cause is now known and a plan is sufficient. Suggest `/rpi-propose` only if the fix requires architectural changes that need design review.

### D5: Commit inline on success

When the fix works and tests pass, commit it using the same approval flow as `/rpi-implement` (present files + message, ask approval, scan for sensitive content). The whole point is to close the loop.

## Command Flow

```
1. Intake
   - Accept: bug description, error message, or path to failing test
   - Interview: 1-2 clarifying questions to establish expected vs actual behavior
   - Check for existing diagnoses on the same topic

2. Investigation
   - Reproduce the bug (run failing test or trigger the error)
   - Trace the code path from symptom to root cause
   - Document findings with file:line references

3. Fix-Iterate Loop (max 3 autonomous attempts)
   - Propose a fix hypothesis
   - Apply the fix
   - Run relevant tests / reproduce steps
   - If fixed → proceed to commit
   - If not fixed → revert, analyze why, try next hypothesis
   - After 3 failures → checkpoint with user

4. Resolution
   - On fix: commit with approval, write diagnosis artifact
   - On escalation: write diagnosis artifact (root cause + what was tried),
     suggest /rpi-plan or /rpi-propose
```

## Artifact Format

```markdown
---
date: ...
topic: "..."
tags: [diagnose]
status: draft | active | complete
spec: (optional, if a spec was referenced)
---

# Diagnosis: [title]

## Bug Report
- **Expected**: ...
- **Actual**: ...
- **Reproduction**: ...

## Root Cause
[What caused the bug, with file:line references]

## Investigation Log
### Attempt 1
- **Hypothesis**: ...
- **Change**: ...
- **Result**: ...

## Resolution
- **Fix applied**: yes | no (escalated)
- **Fix summary**: ...
- **Tests added/modified**: ...
- **Escalation**: (if applicable) suggested next command + rationale
```

## Pipeline Integration

- Add to CLAUDE.md pipeline description
- New command file: `.claude/commands/rpi-diagnose.md`
- New artifact directory: `.rpi/diagnoses/`

## What This Does NOT Do

- Replace `/rpi-research` for exploratory investigation
- Replace `/rpi-implement` for planned work
- Handle vague "something feels wrong" without a reproducible symptom
- Auto-merge or push changes
