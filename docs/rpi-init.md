# The `rpi init` Command

The `rpi init` command bootstraps the workflow into any project:

```bash
# Basic init (creates .claude/ directory structure + CLAUDE.md + .thoughts/)
rpi init

# Init with all agents, commands, and skills from your dotfiles
rpi init --all

# Init a specific project directory
rpi init --all ~/projects/my-app

# Init with .thoughts/ tracked in git (for team sharing)
rpi init --all --track-thoughts

# Update existing configs from dotfiles (preserves local changes)
rpi init --update

# Options
rpi init --force            # Overwrite existing .claude/
rpi init --no-claude-md     # Skip CLAUDE.md creation
rpi init --agents-only      # Only copy agents (use with --all or --update)
rpi init --commands-only    # Only copy commands (use with --all or --update)
rpi init --skills-only      # Only copy skills (use with --all or --update)
rpi init --track-thoughts   # Don't gitignore .thoughts/ (track in git)
```

The command copies agents, commands, skills, and hooks from your global `~/.claude/` directory. Set the `DOTFILES_CLAUDE` environment variable to use a different source directory (e.g., `DOTFILES_CLAUDE=~/dotfiles/.claude rpi init --all`).

## Update Mode

Use `--update` to sync configs from dotfiles into an existing project:

- **New files** are copied automatically
- **Unchanged files** are skipped
- **Differing files** are skipped with a warning (use `--force` to overwrite)
- **CLAUDE.md** sections missing from the template are appended

```bash
rpi init --update                    # Sync all components
rpi init --update --agents-only      # Sync only agents
rpi init --update --force            # Overwrite differing files
```

## Installation

Build and install the `rpi` binary:

```bash
make install
```

This builds the Go binary and copies it to `~/.local/bin/rpi`. Make sure `~/.local/bin` is in your PATH.
