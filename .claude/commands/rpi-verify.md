---
description: Verify implementation against design artifacts for completeness, correctness, and coherence
model: opus
---

# Verify Implementation

Validate that an implementation matches its design artifacts across three dimensions: completeness, correctness, and coherence. Produces a severity-classified verification report.

This command is purely advisory — it does not block anything. It can be re-run after fixes to confirm resolution.

## Step 1: Receive the input

If the user provided a path to a design doc, plan, or ticket as command arguments, proceed to Step 3.

If no arguments were provided, proceed to Step 2 (auto-detection).

## Step 2: Auto-detect what to verify

When no path is provided, detect from recent git changes:

1. Run `git diff --name-only main...HEAD` to find changed files on the current branch
   - If on `main` or the diff is empty, fall back to `git diff --name-only HEAD~5` (last 5 commits)
2. Scan `.thoughts/plans/` for plans with `status: active` or recently completed phases (checked-off items)
3. Scan `.thoughts/designs/` for designs with `status: active`
4. If artifacts found, announce:
   ```
   Found active plan at [path] — verifying against it.
   ```
   Then proceed to Step 3 with the discovered path(s).
5. If nothing found, ask:
   ```
   I couldn't detect an active plan or design from recent git changes.

   What would you like me to verify? Provide a path to:
   - A plan: `.thoughts/plans/YYYY-MM-DD-description.md`
   - A design: `.thoughts/designs/YYYY-MM-DD-description.md`
   - A ticket: `.thoughts/tickets/ticket-id.md`
   ```

## Step 3: Read referenced artifacts

Read the provided or detected artifact(s) fully. Then follow the artifact chain:

- **If given a plan**: read the plan, then read any linked design doc, ticket, and structure doc from its "Source Documents" section
- **If given a design doc**: read the design doc, then scan `.thoughts/plans/` for plans that reference it
- **If given a ticket**: read the ticket, then read its `design:` field target, and scan `.thoughts/plans/` for plans referencing the ticket

Also check:
- `.thoughts/specs/` — if this directory exists, read any specs relevant to the changed domain areas
- `git diff --name-only main...HEAD` (or last 5 commits) — to identify which files were actually changed in the implementation

Present a summary before proceeding:

```
Verifying against:
- Plan: [path] (if found)
- Design: [path] (if found)
- Ticket: [path] (if found)
- Specs: [paths] (if found)

Changed files: [N files from git diff]

Spawning verification sub-agents...
```

## Step 4: Verify across three dimensions

Spawn 3 sub-agents in parallel, one per dimension. Each sub-agent receives the full context from Step 3.

### Sub-agent 1: Completeness

Check whether everything that was planned has been done.

- Sub-task: "Verify completeness of the implementation. You are checking whether all planned work was actually completed.

  Context:
  - [Include the plan content, ticket content, and list of changed files from git diff]

  Check each of these:
  1. **Plan phases**: Are all phases checked off? List any unchecked items.
  2. **Acceptance criteria**: If a ticket exists, are all acceptance criteria met? Read the actual implementation files to verify — don't trust checkboxes alone.
  3. **TODO/FIXME/HACK markers**: Search for these markers in the changed files using Grep. Report any found in new or modified code.
  4. **Test coverage**: Do tests exist for new functionality? Check that new public functions, components, or endpoints have corresponding test files or test cases.
  5. **File coverage**: Were all files mentioned in the plan actually created or modified? Cross-reference the plan's file lists against `git diff --name-only`.

  For each item, report:
  - PASS: [what was satisfied]
  - MISSING (BLOCKER): [critical gap]
  - MISSING (WARNING): [non-critical gap]

  Return your findings as a structured list."

### Sub-agent 2: Correctness

Check whether the implementation matches the design intent.

- Sub-task: "Verify correctness of the implementation against the design. You are checking whether what was built matches what was specified.

  Context:
  - [Include the design doc content, plan content, specs content if available, and list of changed files]

  Check each of these:
  1. **Approach alignment**: Does the implementation follow the chosen approach from the design doc? Read the key implementation files and compare against design decisions.
  2. **API contracts**: If the design specifies interfaces, APIs, or data contracts, verify the implementation matches. Check function signatures, data shapes, and endpoint definitions.
  3. **Edge cases**: Were edge cases identified in the design actually handled in code? Read the relevant code sections to verify.
  4. **Silent deviations**: Are there files changed that weren't in the plan? Are there approaches that differ from what was designed? Flag any undocumented divergence.
  5. **Spec consistency**: If `.thoughts/specs/` files were provided, verify the implementation is consistent with the behavioral specs described there.

  For each item, report:
  - PASS: [what matches]
  - DEVIATION (BLOCKER): [critical mismatch]
  - DEVIATION (WARNING): [minor mismatch, may be intentional]
  - DEVIATION (NOTE): [trivial difference, likely fine]

  Return your findings as a structured list."

### Sub-agent 3: Coherence

Check whether the implementation fits the existing codebase.

- Sub-task: "Verify coherence of the implementation with the existing codebase. You are checking whether the new code fits in naturally.

  Context:
  - [Include the list of changed files and their contents]

  Load the `find-patterns` skill, then check each of these:
  1. **Naming conventions**: Do new files, functions, variables, and classes follow the naming patterns used in the rest of the codebase? Find similar existing code to compare.
  2. **Error handling**: Does error handling follow established patterns? Compare with similar code in the project.
  3. **Code reuse**: Does the new code use existing utilities, helpers, or shared code rather than reinventing? Check for duplication with existing code.
  4. **Dependencies**: Were any unnecessary dependencies introduced? Check for new imports or packages that duplicate existing capabilities.
  5. **Test patterns**: Do the new tests follow the existing test style? Compare with nearby or similar test files for assertion style, setup patterns, and organization.
  6. **File organization**: Do new files follow the project's conventions for directory structure, imports, and module boundaries?

  For each item, report:
  - PASS: [what's consistent]
  - INCONSISTENCY (WARNING): [pattern mismatch]
  - INCONSISTENCY (NOTE): [minor style difference]

  Return your findings as a structured list."

## Step 5: Synthesize verification report

After all three sub-agents complete, synthesize their findings into a single report.

1. **Determine overall status**:
   - **Pass**: No blockers, no warnings (or only notes)
   - **Pass with warnings**: No blockers, but some warnings exist
   - **Issues found**: At least one blocker exists

2. **Classify all findings by severity**:
   - **Blockers (must fix)**: Missing acceptance criteria, critical design deviations, broken contracts
   - **Warnings (should fix)**: Incomplete test coverage, minor deviations, inconsistent patterns
   - **Notes (consider fixing)**: Style nits, minor naming mismatches, suggestions

3. **Write the report** to `.thoughts/reviews/YYYY-MM-DD-verify-[topic].md` using the format below.

4. **Present the summary** to the user:
   ```
   Verification complete: [Pass / Pass with warnings / Issues found]

   - Completeness: [pass / warnings / issues]
   - Correctness: [pass / warnings / issues]
   - Coherence: [pass / warnings / issues]

   [N blockers, N warnings, N notes]

   Full report: .thoughts/reviews/YYYY-MM-DD-verify-[topic].md
   ```

   If blockers exist, list them directly in the summary so the user sees them immediately.

## Report Format

```markdown
---
date: [ISO 8601 datetime with timezone]
topic: "Verification: [Feature/Plan Name]"
tags: [verification, relevant-areas]
type: verification
status: complete
verified_against:
  - [path to design/plan/ticket]
---

# Verification: [Feature/Plan Name]

## Summary
[1-2 sentence overall assessment: pass / pass with warnings / issues found]

## Completeness
### Status: [pass / warnings / issues]
- [x] [Completed item]
- [ ] [Missing item — BLOCKER/WARNING]

## Correctness
### Status: [pass / warnings / issues]
- [Finding with file:line reference and severity]

## Coherence
### Status: [pass / warnings / issues]
- [Finding with file:line reference and severity]

## Issues by Severity

### Blockers (must fix)
- [Issue description with file:line]

### Warnings (should fix)
- [Issue description with file:line]

### Notes (consider fixing)
- [Issue description with file:line]
```

## Guidelines

- **Purely advisory** — this command never blocks anything. Report findings and let the user decide.
- **Re-runnable** — can be run again after fixes. Each run produces a new report file.
- **Works standalone** — does not require the full RPI pipeline to have been used. Can verify any code against any artifact.
- **Read actual code** — don't trust checkboxes or summaries. Sub-agents must read the implementation files to verify claims.
- **Be specific** — every finding should include a `file:line` reference where possible. Vague findings are not actionable.
- **Severity matters** — distinguish genuine blockers from style nits. Don't inflate severity.
- **Check specs** — if `.thoughts/specs/` exists and contains relevant specs, verify behavioral consistency.
- **Scale effort** — if the implementation is small (1-2 files), the sub-agents can be lighter. If it spans many files, be thorough.
