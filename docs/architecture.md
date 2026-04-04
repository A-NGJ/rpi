# Architecture

## Why a Go Binary

The `rpi` CLI exists to keep mechanical work out of the LLM's context window. Every token an LLM spends on parsing YAML frontmatter, resolving file links, or generating boilerplate is a token not spent on design thinking or code generation.

The binary handles operations that are **deterministic and error-prone for LLMs**:

- **Template scaffolding** -- `rpi scaffold` generates documents with correct frontmatter, dates, and file paths. An LLM asked to do this will occasionally hallucinate fields or misformat dates.
- **Artifact chain resolution** -- `rpi chain` follows frontmatter links recursively (plan → proposal → research) and returns a flat list of files to load. This is a mechanical graph traversal, not a creative task.
- **Frontmatter manipulation** -- `rpi frontmatter` reads, writes, and validates status transitions. YAML parsing in natural language is fragile; a CLI does it reliably every time.
- **Directory scanning and filtering** -- `rpi scan` walks `.rpi/`, parses metadata, and filters by status/type. Fast and deterministic vs. asking the LLM to shell out and parse results.
- **Verification checks** -- `rpi verify` counts checkboxes, checks file coverage against git changes, and scans for TODO markers. Mechanical validation that should never consume context tokens.
- **Section extraction** -- `rpi extract` pulls a single heading's content from a markdown file, so the LLM can load exactly the section it needs instead of reading an entire document.
- **MCP server** -- `rpi serve` exposes all of the above as typed [MCP](https://modelcontextprotocol.io/) tools over stdio. The LLM calls `rpi_scaffold`, `rpi_scan`, `rpi_chain`, etc. with validated JSON schemas instead of constructing shell commands. `rpi init` auto-registers the server with Claude Code.

Everything is embedded in a single binary via Go's `embed` package -- no external config repos, no dotfile dependencies. `rpi init` bootstraps any project from the binary alone.

## Project Structure

```
.
├── cmd/rpi/                              # CLI binary + MCP server (Go)
├── internal/
│   ├── chain/                            # Artifact chain resolution
│   ├── frontmatter/                      # YAML frontmatter parsing, writing, transitions
│   ├── git/                              # Git context gathering
│   ├── plan/                             # Plan progress parsing
│   ├── scanner/                          # .rpi/ directory scanning
│   ├── template/                         # Template resolution with user-override support
│   ├── templates/                        # Go template rendering
│   └── workflow/
│       └── assets/                       # All embedded assets (installed by rpi init)
│           ├── agents/                   # Agent definitions
│           ├── commands/                 # Slash command definitions (rpi-plan, rpi-research, etc.)
│           ├── skills/                   # Skill definitions (find-patterns, locate-codebase, etc.)
│           └── templates/                # Scaffold + document templates (.tmpl, .template)
├── docs/
│   ├── architecture.md                   # This file
│   ├── workflow-guide.md                 # Choosing Your Path (detailed examples)
│   ├── stages.md                         # How Each Stage Works (detailed)
│   ├── thoughts-directory.md             # .rpi/ directory documentation
│   └── rpi-init.md                       # rpi init command documentation
```
