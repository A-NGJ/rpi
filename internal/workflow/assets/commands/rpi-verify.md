---
description: Verify implementation against proposal artifacts for completeness, correctness, and coherence
model: opus
disable-model-invocation: true
---

# Verify Implementation

Validate that an implementation matches its proposal artifacts across three dimensions: completeness, correctness, and coherence. Produces a severity-classified verification report.

This command is purely advisory — it does not block anything. It can be re-run after fixes to confirm resolution.

## Step 1: Receive the input

If the user provided a path to a proposal, plan, or other artifact as command arguments, proceed to Step 3.

If no arguments were provided, proceed to Step 2 (auto-detection).

## Step 2: Auto-detect what to verify

When no path is provided, detect from recent git changes:

1. Use the rpi_git_changed_files tool to get the list of changed files and the rpi_scan tool to find active plans/proposals
2. If artifacts found, announce what you're verifying and proceed to Step 3
3. If nothing found, ask for a path to a plan or proposal

## Step 3: Read referenced artifacts

Read the provided or detected artifact(s) fully. Use the rpi_chain tool to resolve the artifact chain — this returns the full chain (plan → proposal → research) with metadata. Read all linked files.

Also check `.rpi/specs/` for relevant specs, and use the rpi_git_changed_files tool to get the list of changed files.

Present a brief summary of what you're verifying before proceeding.

## Step 4: Verify across three dimensions

Verify all three dimensions — parallelize when possible. Each dimension requires reading the actual implementation files, not trusting summaries or checkboxes.

### Completeness

Check whether everything planned has been done:
- Are all plan phases and tasks complete? Use the rpi_verify_completeness and rpi_verify_markers tools for mechanical checks (checkbox counts, file coverage, marker scans for TODO/FIXME/HACK).
- If a ticket exists, are all acceptance criteria met?
- Do tests exist for new functionality?
- Were all planned files created or modified?

### Correctness

Check whether the implementation matches the design intent:
- **Spec conformance**: If specs exist in `.rpi/specs/` for the affected modules, read each spec behavior (XX-N) and verify the implementation satisfies it by reading the actual code and tests — not by looking for comments or markers. Report behaviors that appear unimplemented or untested.
- Does the implementation follow the approach chosen in the proposal?
- Do API contracts, function signatures, and data shapes match what was specified?
- Were edge cases identified in the proposal handled in code?
- Are there silent deviations — files changed or approaches used that weren't in the plan?

### Coherence

Check whether the implementation fits the existing codebase:
- Do naming conventions, error handling, and code organization follow existing patterns?
- Does the new code reuse existing utilities rather than reinventing?
- Were unnecessary dependencies introduced?
- Do tests follow the project's existing test style?

For each finding, classify as: blocker (must fix), warning (should fix), or note (consider fixing).

## Step 5: Synthesize verification report

After all three dimensions are verified:

1. **Determine overall status**: Pass / Pass with warnings / Issues found
2. **Write the report**: Use the rpi_scaffold tool to scaffold a verification report in `.rpi/reviews/`. Fill in findings grouped by dimension and severity.
3. **Present the summary** — overall status, counts by severity, report path. If blockers exist, list them directly so the user sees them immediately.

## Guidelines

- **Purely advisory** — this command never blocks anything
- **Re-runnable** — each run produces a new report file
- **Read actual code** — don't trust checkboxes or summaries
- **Be specific** — every finding should include a file:line reference
- **Severity matters** — distinguish genuine blockers from style nits
- **Check specs** — if `.rpi/specs/` contains relevant specs, verify behavioral consistency
- **Scale effort** — small implementations get lighter verification; large ones get thorough checks
