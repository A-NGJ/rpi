---
archived_date: "2026-04-02"
date: 2026-03-21T00:18:09+01:00
status: archived
tags:
    - plan
    - workflow
    - testing
    - tdd
topic: add red-green TDD and consolidate testing instructions
---

# Add Red/Green TDD and Consolidate Testing Instructions — Implementation Plan

## Overview

Add a red/green TDD instruction to the implement command and consolidate redundant testing sections in CLAUDE.md.

**Scope**: 2 files modified

## Phase 1: Add TDD and Consolidate

### Overview
Add one bullet to `rpi-implement.md`'s Implementation Philosophy, then merge CLAUDE.md's three overlapping testing sections into one.

### Tasks:

#### 1. Add red/green TDD bullet to implement command
**File**: `internal/workflow/assets/commands/rpi-implement.md`
**Changes**: Add a new bullet after "Follow the plan's intent..." (line 39):
```markdown
- **Red/green TDD for new code**: Write tests first, confirm they fail (red), then implement until they pass (green)
```

#### 2. Consolidate CLAUDE.md testing sections
**File**: `CLAUDE.md`
**Changes**: Merge "Implementing Plans" (lines 54-58), "Development Conventions" (lines 60-62), and "Testing" (lines 64-66) into a single "Development Conventions" section:

Before (3 sections):
```markdown
## Implementing Plans
- When implementing a plan from `.rpi/plans/`, present intended changes...
- After implementing changes, always run the full test suite before commiting...

## Development Conventions
Before implementing any changes, always: 1) Read the current version...

## Testing
After implementing changes, always run the full test suite before committing...
```

After (1 section):
```markdown
## Development Conventions

Before implementing any changes, always: 1) Read the current version of each file you plan to modify, 2) Run the existing test suite to establish a baseline, 3) Implement changes incrementally — one logical unit at a time, 4) Run tests after each unit. If tests fail, fix before proceeding. Do not batch all changes and test at the end.

When implementing a plan from `.rpi/plans/`, present intended changes for each phase before writing code. Pause between phases for manual verification. Update checkboxes in the plan file as items complete, and resume from the first unchecked item if checkboxes already exist.
```

This removes:
- The duplicate "Testing" section entirely
- The duplicate "run the full test suite before committing" instruction (already covered by rpi-implement.md)
- The "Tests commonly break due to" hint (appears twice currently — drop both, the implement command's verification section handles this)

### Success Criteria:

#### Automated Verification:
- [x] `make test` passes (pre-existing failures unrelated to our changes)

#### Manual Verification:
- [x] The red/green TDD bullet reads naturally in the implement command's philosophy section
- [x] CLAUDE.md has one "Development Conventions" section instead of three overlapping sections
- [x] No unique testing guidance was lost

### Commit:
- [x] Stage: `internal/workflow/assets/commands/rpi-implement.md`, `internal/workflow/assets/templates/CLAUDE.md.template`, `CLAUDE.md`
- [x] Message: `docs: add red/green TDD to implement command and consolidate testing sections`

**Note**: Pause for manual confirmation before marking complete.
