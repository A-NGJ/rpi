---
domain: index removal
feature: index-removal
last_updated: 2026-04-04T22:30:00+02:00
updated_by: .rpi/designs/2026-04-04-remove-index-subsystem-from-rpi.md
---

# Index Removal

## Purpose

Ensure the index subsystem is fully removed from RPI without breaking remaining functionality.

## Scenarios

### Index package and CLI files are removed
Given the project after the removal is complete
When checking for index-related files
Then `internal/index/` directory, `cmd/rpi/index.go`, and `cmd/rpi/index_test.go` do not exist

### No Go files import the index package
Given the project after the removal is complete
When grepping all `.go` files for `internal/index`
Then zero matches are found

### MCP server has no index tools registered
Given the MCP server is running
When listing registered tools
Then no tool name starts with `rpi_index_` and all non-index tools remain present

### CLI rejects index command
Given the built binary
When the user runs `rpi index`
Then it exits with an unknown-command error

### Init does not reference index in gitignore
Given a fresh directory
When the user runs `rpi init`
Then `.gitignore` does not contain `index.json`

## Constraints
- Remove all index code, tests, MCP tools, and CLI commands
- Preserve all non-index functionality unchanged
- Update all documentation that references the index
- Do not leave stubs, deprecation shims, or dead imports

## Out of Scope
- Cleaning up `.rpi/index.json` from existing projects
- Providing a replacement for index functionality
