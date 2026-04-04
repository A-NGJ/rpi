---
name: rpi-verify
description: Verify implementation against design artifacts for completeness, correctness, and coherence
---

# Verify Implementation

## Goal

Validate that an implementation matches its design artifacts across three dimensions: completeness, correctness, and coherence. Produce a severity-classified verification report in `.rpi/reviews/`. This command is purely advisory — it never blocks anything.

If no path provided, auto-detect from recent git changes. If artifacts found, announce what you're verifying.

## Invariants

- Resolve the artifact chain from the provided or detected artifact — read all linked files (plan → design → research)
- Check `.rpi/specs/` for relevant specs and get the list of changed files
- Read actual implementation files — never trust summaries or checkboxes
- **Completeness**: check all plan phases/tasks done, tests exist, all planned files created/modified, scan for TODO/FIXME/HACK markers
- **Correctness**: use the `rpi_verify_spec` MCP tool to extract scenarios from linked specs, then verify each scenario against actual code and tests with pass/fail per scenario and file:line references; check API contracts match design, flag silent deviations
- **Coherence**: verify naming conventions, error handling, code organization follow existing patterns; check for unnecessary dependencies
- Classify each finding as: blocker (must fix), warning (should fix), or note (consider fixing)
- Scaffold a verification report, fill in findings grouped by dimension and severity
- Present summary: overall status, counts by severity, report path; list blockers directly

## Principles

- Be specific — every finding includes a file:line reference
- Severity matters — distinguish genuine blockers from style nits
- Scale effort — small implementations get lighter verification; large ones get thorough checks
- Re-runnable — each run produces a new report file
