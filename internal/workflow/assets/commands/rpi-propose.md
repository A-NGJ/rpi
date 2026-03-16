---
description: Investigate, analyze, and propose solutions — from quick decisions to complex features
model: opus
---

# Solution Proposal

Investigate the codebase, analyze trade-offs, and produce a proposal with a behavioral spec that captures what we learned, what we decided, and what the system should do.

This is part of the pipeline: **research → propose → plan → implement**. Propose is where the hard choices happen — understanding the terrain, weighing options, and committing to an approach. The spec is the culminating deliverable — a behavioral contract with test cases that defines what success looks like before any code is written. The proposal captures *why* (reasoning, trade-offs); the spec captures *what* (behavioral contract).

**Prerequisite**: The `rpi` binary must be available in PATH. If not found, run `go build -o bin/rpi ./cmd/rpi` or `make install`. See `.rpi/cli-reference.md` for available commands.

## Initial Response

Auto-detect the mode:

- **Path to an existing proposal doc** → **Incremental mode** (updating a previous proposal)
- **Focused decision** (e.g., "should we use X or Y?", "design the caching approach") → **Quick mode**
- **Complex feature description or path to research doc** → **Full mode**
- **Nothing provided** → Ask for input with brief examples of each mode

---

## Quick Mode

For focused decisions: choosing between approaches, designing a single component's interface, deciding on a data model shape.

### Step 1: Understand the decision

1. Read any mentioned files fully
2. Use `rpi` to query the codebase index for files related to the topic, then read them. For decisions touching multiple areas, also look for similar patterns in the codebase.
3. Present the decision frame: what needs to be decided, the relevant codebase context (with file:line refs), and the constraints. If the user already specified the options, skip straight to analysis.

### Step 2: Trade-off analysis

Present options with concrete trade-offs — how each works, pros, cons, and whether it fits existing patterns (with file:line evidence). Give a clear recommendation with reasoning tied to the specific constraints.

### Step 3: Write proposal and spec

After the user confirms direction:

Use `rpi` to scaffold and save a proposal artifact for this topic. Fill in the relevant sections — for Quick mode, focus on: Summary, Constraints & Requirements, Design Decisions, References. Skip sections that don't apply.

Use `rpi` to transition the proposal to active status.

If this proposal was created from a research doc, use `rpi` to check whether the research findings are fully addressed, then transition it to complete. If gaps remain, note them and ask.

**Create the spec contract**: Use `rpi` to scaffold a spec for the affected module(s) using the SDD template. Fill in: Purpose, Behavior (with prefixed IDs), Constraints (Must / Must Not / Out of Scope), and Test Cases (Given/When/Then). If a spec already exists for the affected module, update it with new/changed behaviors instead of creating a new one.

Present the spec to the user for approval or editing. Once accepted, use `rpi` to transition the spec from `draft` to `approved`.

Then suggest: `→ /rpi-plan .thoughts/proposals/YYYY-MM-DD-description.md`

---

## Full Mode

For features that involve multiple interacting decisions, new components, or significant architectural changes.

### Step 1: Investigate codebase

Build a thorough understanding of the terrain before proposing solutions.

1. **Read all mentioned files fully** before investigating further
2. **Check upstream context** — if a research doc was provided, use `rpi` to check its status and resolve the full artifact chain. Read all files it identifies. Warn if the research is still in draft or already complete. Also check for existing proposals on the same topic.
3. **Investigate the relevant areas** (parallelize when possible):
   - Use `rpi` to scan for existing documents about this topic in `.thoughts/`
   - Use `rpi` to query the codebase index for files related to the feature, then read them
   - Understand the current architecture and patterns in the affected areas
   - Find how similar problems are solved in the codebase — concrete examples with file:line refs
4. **Surface relevant non-functional concerns** — consider which genuinely matter (performance, reliability, security, observability). Don't force-fit all of them.
5. **Present findings and open questions** before exploring design options. If the user wants to proceed without checkpoints ("just design it", "I trust your judgment"), compress steps and present the synthesized proposal directly.

### Step 2: Explore design options

Map the independent design decisions. Investigate further if needed — look for similar patterns in the codebase, do web research for library docs or benchmarks when valuable. Present options with concrete trade-offs and a recommendation for each. Use diagrams when they clarify relationships or data flow. Get buy-in before proceeding.

### Step 3: Synthesize

After the user selects directions, validate the combined choices work together:

- Check for contradictions or unexpected complexity when combined
- Check integration with existing code — where does the proposal touch existing systems? Do interfaces need to change?
- Identify risks that only appear when seeing the whole design
- Define file structure when the feature involves new files or module reorganization

Present the cohesive proposal for review before writing.

### Step 4: Write proposal and spec

Use `rpi` to scaffold and save a proposal artifact (linking to the research doc if one exists). Fill in all sections: Summary, Investigation Findings, Constraints & Requirements, Design Decisions, Architecture, File Structure, Risks & Mitigations, What This Proposal Does NOT Cover, References.

Use `rpi` to transition the proposal to active status.

If created from a research doc, verify the research findings are addressed, then use `rpi` to transition it to complete. If gaps remain, note them and ask.

**Create the spec contract**: Use `rpi` to scaffold a spec for the affected module(s) using the SDD template. Fill in: Purpose, Behavior (with prefixed IDs), Constraints (Must / Must Not / Out of Scope), and Test Cases (Given/When/Then). If a spec already exists for the affected module, update it with new/changed behaviors instead of creating a new one.

Present the spec to the user for approval or editing.

### Step 5: Review spec contract & iterate

Focus on the spec as the contract that drives all downstream work. Iterate based on feedback — refine behaviors, constraints, and test cases until the user is satisfied. Resolve all open questions before proceeding.

Once accepted, use `rpi` to transition the spec from `draft` to `approved`.

Then suggest: `→ /rpi-plan .thoughts/proposals/YYYY-MM-DD-description.md`

---

## Incremental Mode

When the user provides a path to an existing proposal that needs updating:

1. **Read the existing proposal fully**
2. **Understand what's changing** — ask what prompted the update (new requirements, implementation findings, changed constraints)
3. **Assess impact** — which decisions are affected? Which still hold?
4. **Research if needed** — investigate only the areas that changed
5. **Propose changes** — present what you'd update and why, get buy-in before modifying
6. **Update the document** — modify in place, use `rpi` to update frontmatter with update metadata (set status to "updated", add last_updated date and update_reason)
7. Add an `## Update Log` section at the bottom if one doesn't exist, with a dated entry explaining what changed and why
8. **Update affected specs** if the changes alter documented behavior

---

## Guidelines

1. **Be opinionated** — present recommendations with clear reasoning grounded in evidence
2. **Be interactive** — get buy-in at each checkpoint; a proposal that surprises the user during review means the process failed
3. **Be evidence-based** — ground decisions in codebase patterns, constraints, and concrete trade-offs with file:line refs
4. **Be focused** — design at the right level of abstraction: architecture and key interfaces, not implementation details
5. **Resolve open questions** — don't finalize with unresolved questions
6. **Respect existing patterns** — prefer solutions that align with how the codebase already works
7. **Specs are the contract** — every proposal culminates in a spec presented for approval. The spec captures *what*, the proposal captures *why*
