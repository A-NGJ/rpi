---
description: Investigate codebase opportunities and ideas without committing to a direction
model: opus
---

# Explore Codebase

Investigate a topic, question, or vague idea to surface opportunities, trade-offs, and recommended next steps — without committing to a direction.

## Why opinionated (not neutral like research)

This command fills the gap before research and design. Research (`/rpi-research`) is cartography — mapping the terrain accurately with zero opinions. Exploration is scouting — you're looking for what's worth doing, what isn't, and where to focus next. Your job here is to form a view and share it clearly.

| Aspect | `/rpi-explore` | `/rpi-research` |
|--------|----------------|-----------------|
| Intent | What should we do? | How does it work? |
| Opinions | Yes — assessments, priorities | No — pure facts |
| Depth | Broad, shallow | Narrow, deep |
| Output | Conversational, optional save | Structured document, always saved |

## Step 1: Receive the topic

If the user provided a topic, question, or idea as command arguments, proceed directly to Step 2.

If no arguments were provided, ask:
```
What would you like me to explore? Give me a topic, question, or vague idea:

- "what could we improve about the auth flow?"
- "is it worth migrating to X?"
- "the API feels slow, where should we look?"
- "what's the tech debt situation in the data layer?"
```

## Step 2: Check for existing context

Before investigating the codebase, check `.thoughts/` for existing work on the topic:

- Sub-task: "Load the `locate-thoughts` skill, then find documents related to [topic]"

If relevant documents are found:
```
Found existing context on this topic:
- [path] — [brief description]

I'll build on this as I explore.
```

If nothing found, continue without comment.

## Step 3: Investigate broadly

Spawn sub-agents to survey the relevant areas. Scale effort to the topic's complexity:

**For focused topics** (e.g., "is our error handling consistent?"):
- 1-2 targeted sub-agents are enough

**For broad topics** (e.g., "what's the tech debt situation?"):
- Sub-task: "Load the `locate-codebase` skill, then find files related to [topic]"
- Sub-task: "Load the `find-patterns` skill, then find how [topic area] is currently handled across the codebase"
- Sub-task (@codebase-analyzer): Understand the specific code areas relevant to the topic

Don't go as deep as research would — you're scanning for patterns, pain points, and opportunities, not documenting every detail.

## Step 4: Form opinionated assessment

This is where exploration diverges from research. Synthesize your findings into an opinionated take:

- **Pain points**: What's causing friction, complexity, or risk?
- **Opportunities**: What could be improved, simplified, or added?
- **Trade-offs**: What are the costs and benefits of different directions?
- **Quick wins vs larger efforts**: What's easy to fix now vs what needs a bigger investment?
- **What's NOT worth doing**: Equally important — call out areas that look problematic but aren't worth the effort

Present findings conversationally — organize by theme or insight, not by file. No rigid template. Think of this as a briefing to a colleague: "Here's what I found and here's what I think we should do about it."

## Step 5: Offer actionable transitions

End with concrete next steps, mapping each to the appropriate pipeline command:

```
Suggested next steps:

- Want me to research [specific area] in depth? → `/rpi-research`
- Ready to design [specific approach]? → `/rpi-design`
- This looks like a small fix — want me to plan it directly? → `/rpi-plan`
```

Tailor the suggestions to what you actually found — don't offer all three if only one makes sense. If the exploration revealed nothing worth pursuing, say so honestly.

## Step 6: Optional save

After presenting findings, ask if the user wants to save:

```
Want me to save these findings? I'll write them to `.thoughts/research/YYYY-MM-DD-explore-[topic].md`.
```

If the user declines or doesn't respond, that's fine — exploration can be purely conversational.

If saving, use this format:

```markdown
---
date: [ISO 8601 datetime with timezone]
topic: "[Exploration Topic]"
tags: [exploration, relevant-areas]
type: exploration
status: complete
---

# Exploration: [Topic]

## Question
[What we were investigating]

## Findings
[What we discovered — organized by theme, not by file]

## Assessment
[Opinionated take — what's worth doing, what isn't, priorities]

## Suggested Next Steps
- [ ] [Action item with suggested command]
- [ ] [Action item]
```

## Guidelines

- Be opinionated — that's the whole point. Don't hedge everything.
- Stay broad — resist the urge to go deep on one area. Flag areas that need depth for `/rpi-research`.
- Keep it conversational — no rigid structure required in the verbal output.
- Save is always optional — never write a file without asking first.
- Use `type: exploration` in frontmatter to distinguish saved explorations from research docs.
- Scale investigation effort to the topic — a simple question doesn't need three sub-agents.
