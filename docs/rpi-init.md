# The `rpi init` Command

The `rpi init` command bootstraps the RPI workflow into any project. All workflow files (agents, commands, skills, scaffold templates) are embedded in the binary -- no external dotfiles or source repo needed.

```bash
# Initialize current directory
rpi init

# Initialize a specific project directory
rpi init ~/projects/my-app

# Init with .thoughts/ tracked in git (for team sharing)
rpi init --track-thoughts

# Options
rpi init --force            # Overwrite existing .claude/ and workflow files
rpi init --no-claude-md     # Skip CLAUDE.md creation
rpi init --track-thoughts   # Don't gitignore .thoughts/ (track in git)
```

## What it creates

- `.claude/agents/` -- Agent definitions (e.g., codebase-analyzer)
- `.claude/commands/` -- Slash command definitions (rpi-plan, rpi-research, etc.)
- `.claude/skills/` -- Skill definitions (find-patterns, locate-codebase, etc.)
- `.claude/templates/` -- Scaffold templates for plans, designs, research docs, etc.
- `.claude/hooks/` -- Empty directory for custom hooks
- `.thoughts/` -- Directory structure for pipeline artifacts (gitignored by default)
- `CLAUDE.md` -- Project-level instructions for Claude Code
- `.thoughts/PIPELINE.md` -- Pipeline reference guide

## Update Mode

Use `--update` to sync custom configs from a dotfiles directory into an existing project. This is useful for overlaying personal agents, commands, or skills on top of the embedded defaults.

- **New files** are copied automatically
- **Unchanged files** are skipped
- **Differing files** are skipped with a warning (use `--force` to overwrite)
- **CLAUDE.md** sections missing from the template are appended

```bash
rpi init --update                    # Sync all components from dotfiles
rpi init --update --agents-only      # Sync only agents
rpi init --update --commands-only    # Sync only commands
rpi init --update --skills-only      # Sync only skills
rpi init --update --force            # Overwrite differing files
```

The dotfiles source defaults to `~/.claude/`. Set the `DOTFILES_CLAUDE` environment variable to use a different directory:

```bash
DOTFILES_CLAUDE=~/dotfiles/.claude rpi init --update
```

## Installation

Build and install the `rpi` binary:

```bash
make install
```

This builds the Go binary and copies it to `~/.local/bin/rpi`. Make sure `~/.local/bin` is in your PATH.
