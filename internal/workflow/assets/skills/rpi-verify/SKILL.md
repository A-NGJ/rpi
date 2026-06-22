---
name: rpi-verify
description: Verify an implementation against its design and spec across completeness, correctness, and coherence — produce a severity-classified review. Use when user says 'verify the implementation', 'check the implementation matches the design', or 'review what I just shipped'.
---

# Verify Implementation

## Goal

Validate that an implementation matches its design artifacts across three dimensions: completeness, correctness, and coherence. Produce a severity-classified verification report in `.rpi/reviews/`. This command is purely advisory — it never blocks anything.

If no path provided, auto-detect from recent git changes. If artifacts found, announce what you're verifying.

## Invariants

- Resolve the artifact chain from the provided or detected artifact — read all linked files (plan → design → research)
- Search for prior verify-reports covering the same area; use them as historical context, not as a new spec source
- Check `.rpi/specs/` for relevant specs and get the list of changed files
- Read actual implementation files — never trust summaries or checkboxes
- **Completeness**: check all plan phases/tasks done, tests exist, all planned files created/modified, scan for TODO/FIXME/HACK markers
- **Correctness**: extract scenarios from linked specs using the verify spec tool, then verify each scenario against actual code and tests with pass/fail per scenario and file:line references; check API contracts match design, flag silent deviations
- **Coherence**: verify naming conventions, error handling, code organization follow existing patterns; check for unnecessary dependencies
- Classify each finding as: blocker (must fix), warning (should fix), or note (consider fixing)
- **Grounding gate**: after drafting findings, run the grounding pass only when the draft has at least one blocker OR more than 3 total findings; otherwise skip grounding and present the review without verdicts
- **Grounding invocation**: pass the drafted finding list (each with its dimension, severity, claim text, and cited anchor) to the read-only `rpi-ground` subagent — it adjudicates the existing findings against repo state, it does not re-derive them
- **Verdict application**: the blocking set presented to the user is the post-grounding set — a finding keeps blocker severity only if `rpi-ground` marks it Verified with a citable anchor; Weakened findings are demoted out of the blocking set with a caveat; Falsified findings are excluded (dropped, or struck-through for transparency)
- **Grounding annotation**: each surviving finding shows its verdict (Verified | Weakened | Falsified) and a one-line evidence pointer; the summary reports before/after counts (e.g. "3 blockers → 2 Verified, 1 Falsified (dropped)")
- **Graceful degradation**: if the `rpi-ground` subagent is unavailable (e.g. opencode / agents-only targets), present the drafted findings with no grounding annotation and an explicit "grounding skipped (subagent unavailable)" note — never a half-annotated review
- Scaffold a verification report, fill in findings grouped by dimension and severity
- Present summary: overall status, counts by severity, report path; list the post-grounding blockers directly

## Principles

- Be specific — every finding includes a file:line reference
- Severity matters — distinguish genuine blockers from style nits
- Scale effort — small implementations get lighter verification; large ones get thorough checks
- Re-runnable — each run produces a new report file

**Recommended model:** premium tier, high effort — adversarial conformance checking, where false negatives are costly (the verify subagent itself is pinned to premium). Advisory; see `docs/model-routing.md`.
