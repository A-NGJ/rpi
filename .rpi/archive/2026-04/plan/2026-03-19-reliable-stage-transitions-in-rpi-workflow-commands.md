---
archived_date: "2026-04-02"
date: 2026-03-19T12:06:35+01:00
status: archived
tags:
    - plan
topic: reliable stage transitions in rpi workflow commands
---

# Reliable Stage Transitions in RPI Workflow Commands — Implementation Plan

## Overview

Embed imperative, visually prominent transition blocks directly into the approval/completion steps of each RPI command file, replacing the current passive suggestions that sit at the end of long files where the agent loses track of them.

**Scope**: 3 files modified (rpi-research.md, rpi-propose.md, rpi-plan.md). No Go code changes.

## Source Documents
- **Research**: .rpi/research/2026-03-18-fragile-stage-transitions-in-rpi-workflow.md

## Phase 1: Embed Transition Blocks in Command Files

### Overview

Replace passive, distant transition instructions with imperative blocks embedded at the exact point where the user approves work. Uses a consistent visual format across all three non-terminal commands.

### Tasks:

#### 1. rpi-research.md — Merge transition into approval steps
**File**: `.claude/commands/rpi-research.md`
**Changes**:
- In Step 7 (Present findings), add a transition block after "Ask if they have follow-up questions" for the case where research is actionable but not saved:
  ```markdown
  > **NEXT STAGE** — When findings point to something actionable and the user is ready to move forward:
  > You MUST suggest: `-> /rpi-propose` (or `/rpi-propose .rpi/research/YYYY-MM-DD-topic.md` if a research artifact was saved)
  ```
- In Step 8 (Optional summary save), add the transition block after saving the artifact
- Remove Step 9 entirely (its content is now embedded in Steps 7 and 8)

#### 2. rpi-propose.md — Embed transition in spec approval steps
**File**: `.claude/commands/rpi-propose.md`
**Changes**:
- In Quick mode Step 3, replace the standalone "Then suggest:" line (line 52) with a prominent transition block embedded right after the spec transition to `approved`:
  ```markdown
  > **NEXT STAGE** — You MUST do this immediately when the user approves the spec:
  > Suggest: `-> /rpi-plan .rpi/designs/YYYY-MM-DD-description.md`
  > Include the actual path of the design artifact you just created.
  ```
- In Full mode Step 5, apply the same replacement to line 107

#### 3. rpi-plan.md — Embed transition in review steps
**File**: `.claude/commands/rpi-plan.md`
**Changes**:
- In Standalone Step 3 (Review & iterate), add the transition block after "Keep iterating until the user confirms":
  ```markdown
  > **NEXT STAGE** — You MUST do this immediately when the user confirms the plan:
  > Suggest: `-> /rpi-implement .rpi/plans/YYYY-MM-DD-description.md`
  > Include the actual path of the plan artifact you just created.
  ```
- In Pipeline Step 6 (Review & iterate), add the same block
- Remove the transition line from the Guidelines section (line 107)

#### 4. No tests needed
This is a prompt-only change. No Go code is affected.

### Success Criteria:

#### Automated Verification:
- [x] `make test` passes (sanity check — no Go changes)

#### Manual Verification:
- [ ] Each modified command file has the transition block in the approval step (not at the end)
- [ ] All transition blocks use imperative wording ("You MUST")
- [ ] All transition blocks include the artifact path pattern
- [ ] No orphaned/duplicate transition instructions remain
- [ ] rpi-research.md no longer has a separate Step 9
- [ ] rpi-plan.md Guidelines section no longer contains a transition instruction

### Commit:
- [ ] Stage: `.claude/commands/rpi-research.md`, `.claude/commands/rpi-propose.md`, `.claude/commands/rpi-plan.md`
- [ ] Message: `fix(commands): embed imperative stage transitions in approval steps`

---

## References
- Research: .rpi/research/2026-03-18-fragile-stage-transitions-in-rpi-workflow.md
