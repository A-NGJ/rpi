---
description: Diagnose bugs through iterative root-cause analysis and fix attempts
model: inherit
---

# Diagnose Bug

## Goal

Iteratively diagnose and fix complex bugs through root-cause analysis. This combines research's interview flow with implement's execution loop, adding an autonomous iterate-until-fixed cycle.

Accept a bug description, error message, or path to a failing test. Clarify expected vs actual behavior, reproduce, trace root cause, then attempt fixes autonomously (up to 3 attempts). Produce a diagnosis artifact in `.rpi/diagnoses/` regardless of outcome.

When the bug is fixed and committed, announce completion. When escalation is needed, suggest → `/rpi-plan` (or `/rpi-propose` for architectural issues).

## Invariants

- Always interview before investigating — ask 1-2 clarifying questions to establish expected vs actual behavior
- Check for existing diagnoses on the same topic before starting — surface them and ask whether to continue or start fresh
- Reproduce the bug before investigating root cause — run the failing test, trigger the error, or confirm the symptom is observable; if unreproducible, checkpoint with user
- Trace the code path from symptom to root cause — all findings must include file:line references
- Do not propose a fix until root cause is identified
- **Fix-iterate loop**: attempt up to 3 fix iterations autonomously; each iteration: apply fix, run relevant tests, evaluate results
- **Revert on failure**: fully revert each failed fix attempt before trying the next — the codebase must never be left in a broken state between attempts
- **Checkpoint after 3 failures**: stop and present what was tried, why each attempt failed, and current understanding of root cause; ask the user how to proceed (more attempts, different approach, or escalate)
- **Commit on success**: when a fix works and tests pass, use the same commit flow as `/rpi-implement` — scan for sensitive content, present files + commit message, ask for approval
- **Diagnosis artifact**: always produce an artifact in `.rpi/diagnoses/` containing bug report (expected/actual/reproduction), root cause with file:line references, investigation log (each attempt with hypothesis/change/result), and resolution status
- **Escalation**: if the fix is too complex for inline patching, write the diagnosis artifact with resolution status "escalated" and suggest `/rpi-plan`; if the fix requires architectural changes, suggest `/rpi-propose` instead
- Do not auto-push or auto-merge changes

## Principles

- Bugs are spec deviations — frame diagnosis as closing the gap between expected and actual behavior
- Escalate, don't force — a clear diagnosis artifact handed off to `/rpi-plan` is more valuable than a fragile hack
- Right-size iteration — simple bugs may be fixed on the first attempt; don't over-investigate when the fix is obvious
