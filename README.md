# AI Agent: Research-Plan-Implement Flow

A structured development workflow for AI coding agents that turns vague feature requests into shipped code through a pipeline of discrete, reviewable stages. Built for [Claude Code](https://docs.anthropic.com/en/docs/claude-code), but the underlying methodology -- Research -> Design -> Plan -> Implement -- works with any AI coding tool.

Instead of asking an AI to "just implement it" and hoping for the best, this workflow forces deliberate progression through **Research -> Design -> Plan -> Implement** -- with optional stages for complex work. Each stage produces a document you can review, edit, and approve before moving on.

```
Research -> Design -> Plan -> Implement
   |          |        |        |
   v          v        v        v
.thoughts/  .thoughts/ .thoughts/ code +
research/   designs/   plans/     tests +
                                  commits
```

## Table of Contents

- [Why This Exists](#why-this-exists)
- [Quick Start](#quick-start)
- [Choosing Your Path](#choosing-your-path)
- [How Each Stage Works](#how-each-stage-works)
- [The `.thoughts/` Directory](#the-thoughts-directory)
- [The `rpi-init` Script](#the-rpi-init-script)
- [Tips](#tips)
- [Using with Other AI Coding Tools](#using-with-other-ai-coding-tools)
- [Project Structure](#project-structure)
- [License](#license)

## Why This Exists

AI coding assistants are powerful but unpredictable when given large tasks. They skip steps, make questionable architectural choices, and produce code that doesn't fit the codebase. This workflow solves that by:

- **Separating thinking from doing** -- Research documents facts without opinions. Design makes decisions with trade-offs. Plans specify exact changes. Implementation follows the plan.
- **Creating review checkpoints** -- You approve each stage before the next one starts. Bad decisions get caught early, not after 500 lines of wrong code.
- **Building persistent context** -- All artifacts live in `.thoughts/`, so you and your team (or the AI) can pick up where you left off across sessions.
- **Scaling to complexity** -- Simple bug fix? Skip straight to Plan -> Implement. Complex feature spanning multiple systems? Use the full pipeline with Tickets.
- **Keeping the context window small** -- LLMs produce better output when focused. By breaking work into stages, each conversation stays scoped to one job (research *or* design *or* implementation) rather than cramming everything into a single bloated context. The `.thoughts/` documents carry knowledge between stages, so the AI starts each stage with exactly the context it needs -- no more, no less.

## Quick Start

### Prerequisites

- [Claude Code CLI](https://docs.anthropic.com/en/docs/claude-code) installed and configured

### Installation

1. Clone this repository:
   ```bash
   git clone <repo-url>
   cd ai-agent-research-plan-implement-flow
   ```

2. Run the initialization script in your target project:
   ```bash
   # From your project directory
   /path/to/ai-agent-research-plan-implement-flow/bin/rpi-init --all /path/to/your/project
   ```

   This creates:
   - `.claude/` -- Agents, commands, skills, and hooks for Claude Code
   - `.thoughts/` -- Directory for all pipeline artifacts (gitignored by default)
   - `CLAUDE.md` -- Project-level instructions for Claude Code

   Add `--track-thoughts` to commit `.thoughts/` to git so your team can share research, designs, and plans.

3. Alternatively, copy the `.claude/` directory manually into your project:
   ```bash
   cp -r .claude/ /path/to/your/project/.claude/
   mkdir -p /path/to/your/project/.thoughts/{research,designs,plans,structures,tickets,specs,prs,reviews}
   ```

4. Start Claude Code in your project and use the slash commands.

### The Slash Commands

| Command | What It Does | Output |
|---------|-------------|--------|
| `/rpi-research` | Investigates the codebase -- fact-finding with optional assessment | `.thoughts/research/YYYY-MM-DD-topic.md` |
| `/rpi-design` | Makes architectural decisions with trade-off analysis | `.thoughts/designs/YYYY-MM-DD-topic.md` |
| `/rpi-structure` | Defines file layout, module boundaries, interfaces | `.thoughts/structures/YYYY-MM-DD-topic.md` |
| `/rpi-tickets` | Breaks a design into independently plannable work units | `.thoughts/tickets/prefix-NNN-name.md` |
| `/rpi-plan` | Creates phased implementation plan with success criteria | `.thoughts/plans/YYYY-MM-DD-topic.md` |
| `/rpi-implement` | Executes a plan phase-by-phase with verification | Code, tests, and commits |
| `/rpi-commit` | Creates focused git commits with smart grouping | Git commits |
| `/rpi-verify` | Validates implementation matches design artifacts | Verification report |
| `/rpi-archive` | Archives completed artifacts to keep `.thoughts/` clean | Moves files to `.thoughts/archive/` |

## Choosing Your Path

Not every task needs every stage. Match the path to your task's complexity:

- **Small tasks** (bug fixes, config changes) -- skip straight to **Plan -> Implement**. `/rpi-plan` does lightweight research on the fly.
- **Medium tasks** (focused features, single-concern changes) -- use the full **Research -> Design -> Plan -> Implement** pipeline.
- **Large tasks** (multi-concern features, major refactors) -- add **Tickets** to break the design into independently plannable units: Research -> Design -> Tickets -> Plan -> Implement (per ticket).
- **Greenfield or major reorganizations** -- add **Structure** before tickets to define file layout and interfaces upfront.

Not sure where to start? Use `/rpi-research` with any question -- it handles both focused investigation and open-ended exploration.

See the [full workflow guide](docs/workflow-guide.md) for detailed examples of each path.

## How Each Stage Works

Each slash command maps to a pipeline stage with a specific purpose. Research gathers facts, Design makes decisions, Plan specifies changes, and Implement executes them. Optional stages (Structure, Tickets) add precision for complex work.

See [detailed stage descriptions](docs/stages.md) for how each command works, its modes, and what it produces.

## The `.thoughts/` Directory

All pipeline artifacts live in `.thoughts/`, organized by type (research, designs, plans, tickets, specs, etc.). Files follow a `YYYY-MM-DD-descriptive-name.md` naming convention and track progress through a `draft -> active -> complete` status lifecycle.

By default `.thoughts/` is gitignored, but you can share it with your team using `--track-thoughts` during init.

See [full `.thoughts/` documentation](docs/thoughts-directory.md) for directory structure, naming conventions, specs, status lifecycle, and team sharing options.

## The `rpi-init` Script

The `bin/rpi-init` script bootstraps the workflow into any project. It copies agents, commands, skills, and hooks from your global `~/.claude/` directory.

```bash
rpi-init --all ~/projects/my-app              # Full init
rpi-init --all --track-thoughts                # Share .thoughts/ via git
rpi-init --update                              # Update existing configs
```

See [full `rpi-init` documentation](docs/rpi-init.md) for all options and flags.

## Tips

- **Start small.** Try `/rpi-plan` on a bug fix to see how the plan -> implement cycle feels before using the full pipeline.
- **Edit the artifacts.** The `.thoughts/` documents are yours. If a design decision is wrong, edit it before planning. If a plan phase is unnecessary, delete it.
- **Use CLAUDE.md.** Add your project's test commands, linting setup, and conventions to `CLAUDE.md`. The pipeline stages pull verification commands from there.
- **Redirect during research.** When `/rpi-research` shows initial findings, tell it to focus on specific areas rather than exploring everything.
- **Skip stages when they don't add value.** The full pipeline exists for complex work. Most daily tasks only need Plan -> Implement.
- **Review the pre-review.** `/rpi-implement` shows you exactly what it plans to change before writing code. This is your last checkpoint -- use it.

## Using with Other AI Coding Tools

This workflow is built for Claude Code, but the methodology applies to any AI coding agent. The `.thoughts/` directory, document templates, and staged pipeline work regardless of tooling. For other tools, follow their documentation on how to register custom commands and load prompt files, then adapt the files in `.claude/` accordingly.

## Project Structure

```
.
├── bin/
|   ├── rpi-init                    # Project initialization script
|   └── templates/
|       ├── CLAUDE.md.template         # Template for project CLAUDE.md
|       └── PIPELINE.md.template       # Pipeline reference document
├── docs/
|   ├── workflow-guide.md              # Choosing Your Path (detailed examples)
|   ├── stages.md                      # How Each Stage Works (detailed)
|   ├── thoughts-directory.md          # .thoughts/ directory documentation
|   └── rpi-init.md                 # rpi-init script documentation
├── .claude/
|   ├── agents/
|   |   └── codebase-analyzer.md       # Agent: traces code flow and documents implementations
|   ├── commands/
|   |   ├── rpi-research.md            # Command: /rpi-research
|   |   ├── rpi-design.md              # Command: /rpi-design
|   |   ├── rpi-structure.md           # Command: /rpi-structure
|   |   ├── rpi-tickets.md             # Command: /rpi-tickets
|   |   ├── rpi-plan.md                # Command: /rpi-plan
|   |   ├── rpi-implement.md           # Command: /rpi-implement
|   |   ├── rpi-verify.md              # Command: /rpi-verify
|   |   ├── rpi-archive.md             # Command: /rpi-archive
|   |   └── rpi-commit.md              # Command: /rpi-commit
|   └── skills/
|       ├── find-patterns/SKILL.md     # Skill: finds existing code patterns to model after
|       ├── analyze-thoughts/SKILL.md  # Skill: extracts insights from .thoughts/ documents
|       ├── locate-thoughts/SKILL.md   # Skill: discovers documents in .thoughts/
|       └── locate-codebase/SKILL.md   # Skill: finds where code lives in the codebase
```

## Acknowledgments

Inspired by [HumanLayer](https://github.com/humanlayer/humanlayer) -- their work on human-in-the-loop patterns for AI agents informed the design of this workflow.

## License

MIT
