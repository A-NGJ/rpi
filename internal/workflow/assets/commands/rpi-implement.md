---
description: Implement technical plans from .rpi/plans with verification
model: sonnet
disable-model-invocation: true
---

# Implement Plan

You are tasked with implementing an approved technical plan from `.rpi/plans/`. These plans contain phases with specific changes and success criteria.

Plans come in two forms:
- **Pipeline plans**: Reference proposals (`.rpi/proposals/`) and optionally research docs. Read these when you need deeper context.
- **Standalone plans**: Self-contained with all context inline (no proposals). These are typically for simpler tasks.

## Getting Started

When given a plan path:

- **Validate plan status**: Use the rpi_frontmatter_get tool to check the plan's current status.
  - If `draft` or `active`: proceed (draft = fresh plan, active = resuming)
  - If `complete`: warn the user that proceeding may duplicate work
- Read the plan completely and check for any existing checkmarks (- [x])
- Resolve the artifact chain: use the rpi_chain tool to resolve the plan's artifact chain and read upstream context.
  This returns linked proposals, research docs. Read the files it identifies.
- Read all files mentioned in the plan
- **Read files fully** - never use limit/offset parameters, you need complete context
- Think deeply about how the pieces fit together
- Update the plan status: use the rpi_frontmatter_transition tool to transition the plan to active
- Check current progress: use the rpi_verify_completeness tool to check the plan's completeness — completed vs remaining items
- Start implementing if you understand what needs to be done

If no plan path provided, ask for one.

## Implementation Philosophy

Plans are carefully designed, but reality can be messy. Your job is to:

- **Preview before writing**: Before modifying any files in a phase, present a summary of all intended changes for approval (see Pre-Review below)
- Follow the plan's intent while adapting to what you find
- Implement each phase fully before moving to the next
- Verify your work makes sense in the broader codebase context
- Update checkboxes in the plan as you complete sections
- Commit changes after each phase (after automated and manual testing have passed)
  - **Before staging**: use the rpi_git_sensitive_check tool to check staged files for sensitive content — if it flags any files, warn the user and exclude them from the commit
  - List the files you plan to add for each commit
  - Show the commit message(s) you'll use. Try to keep them concise yet descriptive.
  - Ask: "I plan to create [N] commit(s) with these changes. Shall I proceed?"
  - **After hook failure**: read the error output, fix the issue, re-stage, and create a **new** commit (never use `--amend`)
- Wait for user manual input after each phase before proceeding

When things don't match the plan exactly, think about why and communicate clearly. The plan is your guide, but your judgment matters too.

If you encounter a mismatch:

- STOP and think deeply about why the plan can't be followed
- Present the issue clearly:

  ```
  Issue in Phase [N]:
  Expected: [what the plan says]
  Found: [actual situation]
  Why this matters: [explanation]

  How should I proceed?
  ```

## Pre-Review: Preview Changes Before Writing

Before writing any code for a phase, present a change preview for user approval:

```
Phase [N] Pre-Review: [Phase Name]

Files to modify:
- `path/to/file.ext` — [what will change]
  ```[language]
  // key code to add/modify (the important parts, not boilerplate)
  ```
- `path/to/other.ext` — [what will change]
  ```[language]
  // key code to add/modify
  ```

Files to create:
- `path/to/new.ext` — [responsibility]
  ```[language]
  // key implementation
  ```

Deviations from plan:
- [Any differences from what the plan specified, and why]
  (or "None — matches plan exactly")

Shall I proceed with these changes?

**Rules:**
- Show the meaningful code — skip trivial imports, boilerplate, or obvious glue
- If a change is large, summarize the approach and show the critical sections
- Always flag deviations from the plan — don't silently diverge
- If the user rejects or adjusts, incorporate feedback before writing
- Once approved, implement the full phase including all details not shown in the preview

## Verification Approach

After implementing a phase:

- Run the success criteria checks (usually `make check test` covers everything)
- Fix any issues before proceeding
- Update your progress in both the plan and your todos
- Check off completed items in the plan file itself using Edit
- **Pause for human verification**: After completing all automated verification for a phase, pause and inform the human that the phase is ready for manual testing. Use this format:

  ```
  Phase [N] Complete - Ready for Manual Verification

  Automated verification passed:
  - [List automated checks that passed]

  Please perform the manual verification steps listed in the plan:
  - [List manual verification items from the plan]

  Let me know when manual testing is complete so I can proceed to Phase [N+1].

If instructed to execute multiple phases consecutively, skip the pause until the last phase. Otherwise, assume you are just doing one phase.

Do not check off items in the manual testing steps until confirmed by the user.

## If You Get Stuck

When something isn't working as expected:

- First, make sure you've read and understood all the relevant code
- Consider if the codebase has evolved since the plan was written
- Present the mismatch clearly and ask for guidance

If you're stuck on unfamiliar code, research it before guessing.

## Resuming Work

If the plan has existing checkmarks:

- Trust that completed work is done
- Use the rpi_verify_completeness tool to check the plan's completeness — what's done vs remaining
- Pick up from the first unchecked item
- Verify previous work only if something seems off

Remember: You're implementing a solution, not just checking boxes. Keep the end goal in mind and maintain forward momentum.

## Completion

When all phases are done and verified:

1. **Verify spec conformance** — check `.rpi/specs/` for specs linked to this plan or covering affected modules:
   - For each spec behavior (XX-N), verify the implementation satisfies it by reading the actual code and tests
   - If a mismatch is found: STOP. Present the divergence clearly. The spec must be amended first (get user approval on the spec change), then continue.
   - Once all behaviors are verified, use the rpi_frontmatter_transition tool to transition the spec to `implemented`
   - Skip this step if no relevant specs exist

2. **Update the plan status**: use the rpi_frontmatter_transition tool to mark the plan as complete
3. **Announce**: "All phases complete. Plan status updated to `complete`."
