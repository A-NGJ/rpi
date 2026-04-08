---
domain: documentation
feature: documentation-structure
last_updated: 2026-04-08T12:14:26+02:00
updated_by: .rpi/designs/2026-04-08-documentation-restructure-for-external-users.md
---

# documentation-structure

## Purpose

Define the expected structure and content of the README so that external developers browsing GitHub can understand RPI's value within the first scroll and have a clear path to trying it.

## Scenarios

### Scanning the README above the fold
Given a developer opens the README on GitHub
When they see the first viewport (~25 lines)
Then they see: project name, badges (license, version, CI), and an opener that names the problem (steering AI agents) and the solution (staged framework with reviewable artifacts)

### Understanding the workflow from the hero example
Given a developer reads the hero example section
When they reach the end of the example
Then they understand the full Research → Propose → Plan → Implement cycle from a single concrete scenario, with each step showing the command, what happens, and what artifact is produced

### Differentiating RPI from alternatives
Given a developer wants to know how RPI differs from other approaches
When they look for comparison information
Then they find a positioning section before the Quick Start that leads with RPI's differentiators (reviewable artifacts + compiled CLI) followed by brief comparisons to specific alternatives

### Following Quick Start to first result
Given a developer decides to try RPI
When they follow the Quick Start section
Then they can install, initialize a project, and run a concrete first command that produces a visible result — not just "use the slash commands"

### Discovering available commands
Given a developer wants to know what commands are available
When they look at the slash commands table
Then they see every command with its purpose and output location, grouped in a single reference table

### Finding detailed documentation
Given a developer wants deeper information on a specific topic
When they look for documentation links
Then they find a documentation section that links to workflow guide, stage descriptions, directory structure, CLI reference, and architecture — each with a one-line description

### Opener tone conveys empowerment not criticism
Given a developer reads the opening paragraph
When they interpret the framing of AI coding agents
Then the text conveys that agents are capable and the challenge is directing them, not that agents are unreliable or produce bad code

## Constraints
- README is a single file — no splitting across multiple READMEs
- No inline images or GIFs (visual assets are out of scope)
- Must reference both Claude Code and OpenCode as supported tools
- Badges use shields.io and GitHub-native badge URLs only
- Hero example uses stable command names that exist in the current slash command set
- docs/ files are not modified — README links to them as supporting references

## Out of Scope
- Visual assets (GIF, screencast, terminal recordings)
- Changes to any file in docs/
- Documentation site generation (MkDocs, Docusaurus, etc.)
- Localization or multi-language support
