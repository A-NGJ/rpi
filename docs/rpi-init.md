# The `rpi-init` Script

The `bin/rpi-init` script bootstraps the workflow into any project:

```bash
# Basic init (creates .claude/ directory structure + CLAUDE.md + .thoughts/)
rpi-init

# Init with all agents, commands, and skills from your dotfiles
rpi-init --all

# Init a specific project directory
rpi-init --all ~/projects/my-app

# Init with .thoughts/ tracked in git (for team sharing)
rpi-init --all --track-thoughts

# Update existing configs from dotfiles (preserves local changes)
rpi-init --update

# Options
rpi-init --force            # Overwrite existing .claude/
rpi-init --no-claude-md     # Skip CLAUDE.md creation
rpi-init --agents-only      # Only copy agents
rpi-init --commands-only    # Only copy commands
rpi-init --skills-only     # Only copy skills
rpi-init --track-thoughts   # Don't gitignore .thoughts/ (track in git)
```

The script copies agents, commands, skills, and hooks from your global `~/.claude/` directory. Set the `DOTFILES_CLAUDE` environment variable to use a different source directory (e.g., `DOTFILES_CLAUDE=~/dotfiles/.claude rpi-init --all`).

## RPI Binary

The `rpi` CLI binary is automatically built and installed during initialization:

- If `bin/rpi` doesn't exist in the source repo, it is built with `go build` (requires Go)
- The binary is installed to `~/.local/bin/rpi`
- If `~/.local/bin` is not in your PATH, the script adds it to your shell profile (`~/.zshrc`, `~/.bash_profile`, `~/.bashrc`, or `~/.profile`)
- Running `rpi-init --update` rebuilds and reinstalls the binary
