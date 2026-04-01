---
archived_date: "2026-04-02"
branch: main
date: 2026-03-18T23:51:17+01:00
git_commit: 5de3542
repository: ai-agent-research-plan-implement-flow
researcher: Claude
status: archived
tags:
    - research
    - workflow
    - transitions
    - agent-behavior
topic: fragile stage transitions in rpi workflow
---

# Research: Fragile Stage Transitions in RPI Workflow

## Research Question

Why does the agent forget to suggest or invoke the next stage command when the user approves work, and what can be done to make transitions reliable?

## Problem Statement

When the user signals approval ("looks good", "yes", short affirmations) at the end of an RPI workflow stage, the agent often fails to suggest or invoke the next command in the pipeline. The handoff between stages (research → propose → plan → implement) relies on the agent remembering transition instructions that are buried at the end of long command files, far from where the agent's attention is when the approval happens.

## Summary

Transition instructions in all four pipeline commands are passive suggestions placed at the end of the command file. After a complex workflow involving 50-100+ tool calls, these instructions are thousands of tokens back in context and easily lost. The problem is structural — not a wording issue — because agent attention degrades with distance in context.

## Detailed Findings

### Current transition instructions

| Command | Instruction | Location | Wording |
|---------|------------|----------|---------|
| rpi-research | `rpi-research.md:95` | Step 9 (last step) | "suggest continuing with `/rpi-propose`" |
| rpi-propose | `rpi-propose.md:51,106` | End of Quick/Full mode | "Then suggest: `→ /rpi-plan ...`" |
| rpi-plan | `rpi-plan.md:106` | After guidelines | "proceed to `/rpi-implement` with the plan path" |
| rpi-implement | End of command | Completion section | No transition — just announces completion |

All four are **passive suggestions** ("suggest", "then suggest"), not imperative instructions. They sit at the end of commands that can be 100-130 lines long.

### Why transitions fail

1. **Distance from action point**: By the time the user says "looks good" after reviewing a proposal + spec, the agent has processed the entire command file (130+ lines), made 50-100 tool calls, and the transition instruction is thousands of tokens back. Agent attention to instructions degrades with context distance.

2. **Soft wording**: "Then suggest" is easy to deprioritize compared to the active work of writing artifacts, calling tools, and managing frontmatter transitions. The agent treats it as optional.

3. **No structural marker**: The transition instruction isn't visually or structurally distinguished from surrounding prose. It blends into the step descriptions.

4. **CLAUDE.md partial patch**: `CLAUDE.md` contains "When the user says 'looks good' or similar short affirmations during planning, proceed immediately with implementation" — but this only covers plan → implement, not the other three transitions.

### What a reliable transition would need

- **Proximity to the approval point**: The transition instruction should be near where the user says "looks good," not at the bottom of a long file
- **Imperative wording**: "You MUST suggest" not "suggest"
- **Structural prominence**: A clearly marked block that stands out from surrounding text
- **The artifact path**: The agent needs to know what path to pass to the next command — this requires knowing which file was just created

## Assessment

The transition problem is a **prompt design issue**, not a tooling gap. The fix is in the command files themselves — making transitions impossible to miss rather than easy to forget. Potential approaches:

1. **Standardized "Next Stage" block** at the end of every approval step — visually prominent, imperative wording, includes the artifact path
2. **Repeat the transition instruction** at multiple points — after artifact creation, after spec approval, in guidelines
3. **Move the transition into the approval gate itself** — "Present the spec for approval. When approved, transition spec to `approved` AND suggest the next command"
4. **Shorter commands** — if commands are shorter, the transition is closer to the action

Approach 3 seems most promising: embedding the transition into the same step where approval happens, rather than having it as a separate step at the end.

## Suggested Next Steps

- Propose a fix as part of a broader workflow command restructuring (likely combined with the proposal → design rename)
- The fix is purely in the command markdown files — no Go code changes needed

## Decisions

None yet — this is a research document for a future proposal.
