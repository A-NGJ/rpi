---
domain: documentation
feature: documentation-structure
last_updated: 2026-05-22T11:35:00+02:00
updated_by: .rpi/designs/2026-05-22-shrink-readme-dedupe-install-and-semantic-search.md
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

### Best-fit line states the workflow value
Given a developer reads the opener
When they reach the end of the opener block
Then they see a single "Best fit:" line that names the workflow RPI enables — control over the dev flow with reviewability, a durable trace of decisions, and multi-session continuity — covering both solo developers and teams

### Pain decomposition maps each difficulty to a specific command
Given a developer wants to know which RPI command addresses which problem
When they reach the "What RPI helps with" section between Quick Start and the walkthrough
Then they see a table pairing each named difficulty with the specific RPI artifact or command that addresses it, with phrasing that frames difficulty as the work (not as agent failure)

### Install instructions appear in one canonical section
Given a developer scans the README for how to install RPI
When they look across all sections of the file
Then the Claude Code plugin install commands appear in at most two places — a short Quick Start near the top and one canonical Installation section — never repeated in a third intermediate block

### Quick Start points readers to the canonical Installation section
Given a developer reads the Quick Start at the top of the README
When they finish the install commands and want more options (OpenCode, standalone binary, from source, global setup)
Then they see a single pointer to the canonical Installation section rather than a parallel block of alternative install instructions

### Canonical Installation section covers every supported path
Given a developer needs an install path other than the Claude Code plugin
When they reach the canonical Installation section
Then they find subsections for the Claude Code plugin, OpenCode and standalone CLI (including `rpi init` with `--target opencode`), one-time global setup (`rpi init --global`), from-source builds, and upgrading — each present exactly once

### Migration note for previous standalone users is visible at install
Given a developer previously installed RPI via `rpi init --global` and wants to switch to the Claude Code plugin
When they read the plugin install instructions
Then they see a migration note that tells them to run `rpi uninstall --global` before `/rpi:rpi-setup`

### Semantic search is pitched in the README and configured in docs/
Given a developer skims the README for optional capabilities
When they reach the semantic-search section
Then they see a short paragraph describing what `rpi_search` does and how skills use it, plus a link to `docs/semantic-search.md` — without inline qmd install commands, warmup detail, or status-contract specifics

### Documentation index lists every docs/ file
Given a developer wants deeper documentation on a specific topic
When they look at the Documentation section of the README
Then they see a link entry for every supporting docs/ file that the README references — including `docs/semantic-search.md` when semantic-search content lives there

## Constraints
- README is a single file — no splitting across multiple READMEs
- No inline images or GIFs (visual assets are out of scope)
- Must reference both Claude Code and OpenCode as supported tools
- Badges use shields.io and GitHub-native badge URLs only
- Hero example uses stable command names that exist in the current slash command set
- Existing docs/ files are not modified by README structure work; new docs/ files may be created when README content is migrated out (e.g., setup or troubleshooting that exceeds the README's pitch-level depth)
- The Claude Code plugin install commands (`/plugin marketplace add`, `/plugin install`, `/rpi:rpi-setup`) appear at most twice in the README — once in Quick Start, once in the canonical Installation section

## Out of Scope
- Visual assets (GIF, screencast, terminal recordings)
- Modifications to existing docs/ files (creating new docs/ files is allowed when migrating content out of the README)
- Documentation site generation (MkDocs, Docusaurus, etc.)
- Localization or multi-language support

## Update Log

- **2026-05-22** (`.rpi/designs/2026-05-22-shrink-readme-dedupe-install-and-semantic-search.md`): Added scenarios for install-instruction dedup (single canonical Installation section, Quick Start as pointer, migration note placement) and semantic-search migration to `docs/semantic-search.md`. Softened the "docs/ files are not modified" constraint to allow creating new docs/ files when migrating content out of the README, while still forbidding edits to existing docs/ files. Added an explicit cap of two occurrences for the plugin install commands.
- **2026-05-21** (`.rpi/designs/2026-05-21-tighten-readme-opener-best-fit-line-and-pain-decomposition.md`): Added scenarios for the empowering opener tone, the "Best fit:" line covering solo devs and teams, and the "What RPI helps with" pain-to-command table.
