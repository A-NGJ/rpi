---
date: 2026-04-08T12:17:22+02:00
design: .rpi/designs/2026-04-08-documentation-restructure-for-external-users.md
spec: .rpi/specs/documentation-structure.md
status: active
tags:
    - plan
topic: documentation restructure for external users
---

# documentation restructure for external users — Implementation Plan

## Overview

Restructure README.md as an external-facing landing page. Single phase — rewrite opener, add badges, add hero example, reposition comparison, fix Quick Start ending, reorder sections.

**Scope**: 1 file modified (README.md)

## Source Documents
- **Design**: .rpi/designs/2026-04-08-documentation-restructure-for-external-users.md
- **Spec**: .rpi/specs/documentation-structure.md

## Phase 1: Restructure README.md

### Overview

Rewrite README.md with new section order, empowerment-focused opener, badges, hero example, repositioned comparison, and concrete Quick Start ending. The "Why This Exists" section is removed — its best points are absorbed into the opener and comparison section.

### Tasks:

#### 1. New opener and badges
**File**: README.md
**Changes**:
- Keep `# AI Agent: Research-Propose-Plan-Implement Flow` as H1
- Add 3 badges immediately after H1: License (MIT), Release (latest), CI (release workflow)
  - License: `[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)`
  - Release: `[![Release](https://img.shields.io/github/v/release/A-NGJ/rpi)](https://github.com/A-NGJ/rpi/releases/latest)`
  - CI: `[![Release](https://github.com/A-NGJ/rpi/actions/workflows/release.yml/badge.svg)](https://github.com/A-NGJ/rpi/actions/workflows/release.yml)`
- Replace current opener paragraph (lines 3-4) with empowerment-focused text: AI agents are capable, the challenge is steering them, RPI gives you a framework with staged decisions / reviewable artifacts / specs
- Mention Claude Code and OpenCode support in the opener or immediately after
- **Spec scenarios**: "Scanning the README above the fold", "Opener tone conveys empowerment"

#### 2. Hero example
**File**: README.md
**Changes**:
- Remove ASCII artifact-flow diagram (lines 7-14)
- Add a "See It in Action" section after the opener
- Condense the rate limiting example from `docs/workflow-guide.md:36-68` into 4 steps:
  1. Research (optional): `/rpi-research` — one-line command + what happens
  2. Propose: `/rpi-propose` — command + what happens + artifact produced
  3. Plan: `/rpi-plan` — command + what happens + artifact produced
  4. Implement: `/rpi-implement` — command + what happens + result
- Keep it under 20 lines total
- **Spec scenario**: "Understanding the workflow from the hero example"

#### 3. Repositioned comparison section
**File**: README.md
**Changes**:
- Move "How It Compares" (currently near bottom) to after hero example
- Rename to "How RPI Is Different"
- Lead with the two differentiators: reviewable artifacts that keep a human in the loop, compiled CLI that keeps bookkeeping out of the LLM's context
- Keep vs. OpenSpec and vs. unstructured prompting comparisons
- Absorb the strongest "Why This Exists" points (separating thinking from doing, creating review checkpoints, keeping context window small) into this section's framing
- Remove "Why This Exists" as a standalone section
- **Spec scenario**: "Differentiating RPI from alternatives"

#### 4. Concrete Quick Start ending
**File**: README.md
**Changes**:
- Keep install and init steps as-is
- Replace step 3 ("Start coding" → "use the slash commands") with a concrete example:
  - Show `/rpi-plan` with a simple bug fix description
  - Show `/rpi-implement` with the resulting plan path
  - One line describing the result
- **Spec scenario**: "Following Quick Start to first result"

#### 5. Section reorder and cleanup
**File**: README.md
**Changes**:
- Final section order:
  1. Title + Badges
  2. Opener paragraph
  3. Hero example ("See It in Action")
  4. "How RPI Is Different" (repositioned comparison)
  5. Quick Start (install → init → try it)
  6. Slash Commands table
  7. Choosing Your Path
  8. Documentation links
  9. MCP Server
  10. Acknowledgments
  11. License
- Ensure all internal links (`docs/workflow-guide.md`, etc.) are still valid
- **Spec scenarios**: "Discovering available commands", "Finding detailed documentation"

### Success Criteria:

#### Manual Verification:
- [ ] First viewport (~25 lines) shows: title, badges, empowerment-focused opener — **spec: "Scanning above the fold"**
- [ ] Hero example shows full cycle with commands, results, artifacts in under 20 lines — **spec: "Understanding the workflow from the hero example"**
- [ ] Comparison section appears before Quick Start, leads with differentiators — **spec: "Differentiating from alternatives"**
- [ ] Quick Start ends with concrete command + visible result — **spec: "Following Quick Start to first result"**
- [ ] Slash commands table present with purpose and output — **spec: "Discovering available commands"**
- [ ] Documentation section links to all 5 docs/ files with descriptions — **spec: "Finding detailed docs"**
- [ ] Opener frames agents as capable, challenge is directing them — **spec: "Opener tone conveys empowerment"**
- [ ] All internal links (docs/*.md, LICENSE) resolve correctly
- [ ] Markdown renders correctly on GitHub (badges, tables, code blocks)

### Commit:
- [ ] Stage: README.md
- [ ] Message: `docs: restructure README as external-facing landing page`

---

## References
- Design: .rpi/designs/2026-04-08-documentation-restructure-for-external-users.md
- Spec: .rpi/specs/documentation-structure.md
