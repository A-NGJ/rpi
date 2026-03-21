# The `rpi init` Command

The `rpi init` command bootstraps the RPI workflow into any project. All workflow files (agents, commands, skills, scaffold templates) are embedded in the binary -- no external dotfiles or source repo needed.

Supports two targets: **Claude Code** (default) and **OpenCode**.

```bash
# Initialize current directory (Claude Code)
rpi init

# Initialize for OpenCode
rpi init --target opencode

# Initialize a specific project directory
rpi init ~/projects/my-app
rpi init ~/projects/my-app --target opencode

# Options
rpi init --force            # Overwrite existing files and directories
rpi init --no-claude-md     # Skip rules file generation (CLAUDE.md or AGENTS.md)
rpi init --no-mcp           # Skip MCP server registration
rpi init --track-rpi        # Don't gitignore .rpi/ (track in git)
```

## What it creates

### Claude Code target (default)

- `.claude/agents/` -- Agent definitions (e.g., codebase-analyzer)
- `.claude/commands/` -- Slash command definitions (rpi-plan, rpi-research, rpi-propose, etc.)
- `.claude/skills/` -- Skill definitions (find-patterns, locate-codebase, etc.)
- `.claude/templates/` -- Scaffold templates for plans, designs, research docs, etc.
- `.claude/hooks/` -- Empty directory for custom hooks
- `CLAUDE.md` -- Project-level instructions for Claude Code
- MCP server registered via `claude mcp add rpi -- rpi serve`

### OpenCode target

- `.opencode/agents/` -- Agent definitions
- `.opencode/commands/` -- Slash command definitions
- `.opencode/skills/` -- Skill definitions
- `.opencode/templates/` -- Scaffold templates
- `.opencode/hooks/` -- Empty directory for custom hooks
- `AGENTS.md` -- Project-level instructions for OpenCode

### Shared (both targets)

- `.rpi/` -- Directory structure for pipeline artifacts (gitignored by default)
- `.rpi/index.json` -- Codebase symbol index

### MCP Server Configuration

When the target is `claude`, `rpi init` auto-registers an MCP server so the AI calls typed tools (`rpi_scaffold`, `rpi_scan`, etc.) instead of shelling out to the CLI.

- Requires both `rpi` and `claude` to be in PATH
- Skipped with `--no-mcp` or when the target is `opencode`
- Warns and continues (doesn't fail) if `rpi` or `claude` are not found, or if the server is already registered
- Uses `claude mcp add rpi -- rpi serve` to register
- `rpi init --update` does **not** re-register the MCP server (only rebuilds the index and CLI reference)

## Installation

Build and install the `rpi` binary:

```bash
make install
```

This builds the Go binary and copies it to `~/.local/bin/rpi`. Make sure `~/.local/bin` is in your PATH.

## Shell Completion

The `rpi` binary supports autocompletion for bash, zsh, fish, and powershell. Add one of the following lines to your shell configuration file to enable completions for every new session:

```bash
# Bash (~/.bashrc)
source <(rpi completion bash)

# Zsh (~/.zshrc)
source <(rpi completion zsh)

# Fish (~/.config/fish/config.fish)
rpi completion fish | source
```

Run `rpi completion <shell> --help` for more options.
