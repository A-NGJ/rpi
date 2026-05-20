#!/bin/sh
# Launcher for the rpi MCP server. Resolves ~/.rpi/bin/rpi at runtime
# (Claude Code spawns the manifest's `command` directly, so tilde and $HOME
# inside the manifest itself are not expanded). Also emits a clear message
# to stderr when the binary is missing, so the user sees something more
# actionable than a generic "MCP server failed to start".

bin="$HOME/.rpi/bin/rpi"

if [ ! -x "$bin" ]; then
  cat >&2 <<EOF
rpi binary not found at $bin.
The plugin is installed but the binary is missing. Run /rpi:rpi-setup
inside Claude Code to fetch the matching release and install it.
EOF
  exit 1
fi

exec "$bin" "$@"
