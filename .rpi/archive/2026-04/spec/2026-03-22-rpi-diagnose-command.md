---
archived_date: "2026-04-02"
date: 2026-03-22T12:00:00+01:00
design: .rpi/designs/2026-03-22-rpi-diagnose-command.md
status: archived
tags:
    - spec
topic: rpi-diagnose command for iterative bug analysis
---

# Spec: /rpi-diagnose Command

## Behaviors

### DG-1: Intake clarification

The command accepts a bug description, error message, or path to a failing test. Before investigating, it asks 1-2 clarifying questions to establish expected vs actual behavior. It checks for existing diagnoses on the same topic before proceeding.

### DG-2: Reproduction before investigation

The command reproduces the bug before investigating root cause. This means running the failing test, triggering the error, or confirming the symptom is observable. If the bug cannot be reproduced, it checkpoints with the user.

### DG-3: Root cause tracing with evidence

Investigation traces the code path from symptom to root cause. All findings include file:line references. The command does not propose a fix until root cause is identified.

### DG-4: Autonomous fix-iterate loop (max 3 attempts)

The command attempts up to 3 fix iterations autonomously. Each iteration: apply fix, run relevant tests, evaluate results. If a fix succeeds (tests pass, symptom resolved), proceed to commit. If a fix fails, revert the change, analyze why it failed, and try a different hypothesis.

### DG-5: User checkpoint after 3 failed attempts

After 3 unsuccessful fix attempts, the command stops and presents: what was tried, why each attempt failed, and current understanding of the root cause. It asks the user how to proceed (more attempts, different approach, or escalate).

### DG-6: Commit on successful fix

When a fix succeeds, the command commits using the same flow as `/rpi-implement`: scan for sensitive content, present files + commit message, ask for approval. The fix is not committed without user approval.

### DG-7: Diagnosis artifact creation

The command produces a diagnosis artifact in `.rpi/diagnoses/` containing: bug report (expected/actual/reproduction), root cause analysis with file:line references, investigation log (each attempt with hypothesis/change/result), and resolution status.

### DG-8: Escalation path

If the fix is too complex for inline patching, the command writes the diagnosis artifact with resolution status "escalated" and suggests `/rpi-plan` (root cause known, plan the fix). If the fix requires architectural changes, it suggests `/rpi-propose` instead.

### DG-9: Revert on failure

Each failed fix attempt is fully reverted before the next attempt. The codebase is never left in a broken state between iterations.

## Constraints

### Must

- Establish expected vs actual behavior before investigating
- Reproduce the bug before attempting fixes
- Include file:line references in root cause analysis
- Revert failed fix attempts before trying the next one
- Limit autonomous fix attempts to 3 before checkpointing
- Get user approval before committing fixes
- Scan for sensitive content before committing
- Produce a diagnosis artifact regardless of outcome (fix or escalation)
- Check for existing diagnoses on the same topic before starting

### Must Not

- Propose a fix before identifying root cause
- Leave the codebase in a broken state between fix attempts
- Commit without user approval
- Continue iterating past 3 failures without user checkpoint
- Auto-push or auto-merge changes
- Force a fix when escalation is more appropriate

### Out of Scope

- Exploratory "something feels wrong" investigations with no observable symptom
- Planned feature work (use `/rpi-implement`)
- Design review of architectural changes (use `/rpi-propose`)
- Replacing `/rpi-research` for general codebase exploration

## Test Cases

### TC-1: Simple bug with known reproduction

**Given** a user reports a bug with a failing test path
**When** `/rpi-diagnose path/to/failing_test.py` is invoked
**Then** the command reproduces the failure, identifies root cause, fixes it within 3 attempts, and commits with approval

### TC-2: Bug requiring clarification

**Given** a user provides a vague bug description ("login is broken")
**When** `/rpi-diagnose` is invoked with that description
**Then** the command asks 1-2 clarifying questions to establish expected vs actual behavior before investigating

### TC-3: Fix fails 3 times

**Given** a bug where the first 3 fix hypotheses are incorrect
**When** the command exhausts 3 autonomous attempts
**Then** it presents a summary of what was tried and asks the user how to proceed

### TC-4: Bug requiring architectural fix

**Given** a bug whose root cause is a fundamental design flaw
**When** the command identifies the root cause
**Then** it writes a diagnosis artifact and suggests `/rpi-propose` for design review instead of attempting an inline fix

### TC-5: Existing diagnosis found

**Given** a previous diagnosis artifact exists for the same bug area
**When** `/rpi-diagnose` is invoked for a related issue
**Then** the command surfaces the existing diagnosis and asks whether to continue from it or start fresh

### TC-6: Revert on failed attempt

**Given** a fix attempt that breaks additional tests
**When** the fix is evaluated and found to fail
**Then** the changes are fully reverted before the next attempt, leaving the codebase in its original state
