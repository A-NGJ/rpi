---
description: Research the codebase to understand problems before proposing solutions
model: sonnet
---

# Research Codebase

Investigate the codebase conversationally to understand how things work, find patterns, and surface insights — without being forced to produce artifacts.

This is the entry point for unclear or open-ended questions: "How does X work?", "What's going on with Y?", "Can we improve Z?". When insights crystallize into something actionable, suggest `/rpi-propose`.

## Step 1: Receive the topic

If the user provided a question or topic as command arguments, proceed directly to Step 2.

If no arguments were provided, ask what they'd like to research — this works for focused questions ("how does auth work?"), open-ended exploration ("what could we improve about error handling?"), and vague ideas ("something feels off about the API").

## Step 2: Discovery interview

Always conduct a brief interview to understand the user's needs before investigating. Even for seemingly specific questions, the interview surfaces context that improves research quality — what prompted the question, what they already know, what outcome they're looking for.

### Interview approach

Conversational, not a rigid checklist. Ask 1-2 questions at a time, adapt follow-ups based on answers. Cover these dimensions as relevant:

1. **Problem/need**: What are you trying to accomplish? What's not working or missing?
2. **Motivation**: What prompted this? Why now?
3. **Prior attempts**: What have you already tried or considered?
4. **Constraints**: What must be preserved? What can't change?
5. **Success criteria**: What does "done" look like? How will you know it works?
6. **Scope**: What's in vs. out of scope?

Not every dimension applies to every question — skip what's irrelevant. Stop interviewing when you have enough to investigate productively. You can always ask more questions later if findings raise new questions.

## Step 3: Synthesize understanding

Reflect back a concise problem statement that captures what was learned from the interview. Get confirmation before proceeding to codebase investigation. This becomes the refined research question that guides the rest of the work.

If the user corrects or refines the problem statement, update it and confirm again.

## Step 4: Read mentioned files and check existing context

- If the user mentions specific files, read them fully before doing anything else
- Use the rpi_scan tool to check for existing research artifacts on this topic
  - If relevant research exists, mention it and build on it

## Step 5: Assess scope and investigate

Scale investigation effort to the question:

- **Focused question** (e.g., "how does the webhook handler work?"): Read the relevant files directly — minimal research needed.
- **Broad or exploratory question** (e.g., "what's the tech debt situation?"): Investigate multiple areas in parallel.

To investigate (adapt to scope — parallelize when possible):

1. Use the rpi_index_query tool to find files related to the topic, then read them
2. Understand how the key code works — read the implementation and trace the logic
3. Find existing patterns related to the topic — look for similar code with file:line refs
4. Use the rpi_scan tool to check for existing documents about this topic in `.rpi/`

For focused questions, skip tasks that aren't relevant.

## Step 6: Checkpoint (broad/exploratory questions only)

After initial findings return, present what was found and let the user redirect. Skip for focused questions — just proceed to presenting findings.

## Step 7: Present findings

Synthesize all results:
- Prioritize live codebase findings as the primary source of truth
- Include specific file paths and line numbers
- Connect findings across different components
- Answer the user's specific questions with concrete evidence
- Tie findings back to the problem statement from the interview
- When findings reveal clear pain points or opportunities, include your assessment — but keep facts and opinions clearly separated

Ask if they have follow-up questions.

## Step 8: Optional summary save

If the user asks to save findings ("save this", "summarize what we found"), or if the investigation is substantial enough that cross-session handoff is clearly valuable:

Offer to save to `.rpi/research/`. If they agree:
1. Use the rpi_scaffold tool to scaffold and save a research artifact for this topic
2. Fill in the problem statement from the interview, key findings, file references, and assessment
3. Use the rpi_frontmatter_transition tool to transition the research artifact to active status

**Do not force artifact creation.** The default is conversational — only save when requested or clearly useful.

## Step 9: Transition out

When insights crystallize into something actionable, suggest continuing with `/rpi-propose`. If in the same session, the full context carries over. If in a new session, they can pass the saved summary: `/rpi-propose .rpi/research/YYYY-MM-DD-topic.md`.

## Guidelines

- **Do NOT use `EnterPlanMode`** — this command has its own structured flow; plan mode restricts tools and causes steps to be skipped
- **Always interview** — even focused questions benefit from understanding motivation and context before diving in
- **Don't over-interview** — stop when you have enough to investigate; you can always ask more later
- **Scale to the question** — a simple question doesn't need broad parallel research
- **No forced artifacts** — this command is conversational by default
- **Be concrete** — always include file:line references, not vague descriptions
- **Be interactive** — checkpoint after initial findings for broad questions
- **Cross-session handoff** — when saving, include enough context that a future `/rpi-propose` can skip re-investigating documented areas
- **Facts first, opinions when warranted** — present what exists before suggesting what should change
- **Follow-ups welcome** — append to the same research doc rather than creating a new one
