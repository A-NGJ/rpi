# CLAUDE.md

This file provides guidance to Claude Code when working with this repository.

## Project Overview

This project follows Spec-Driven Development (SDD). Behavioral specs live in `.rpi/specs/` and serve as the source of truth for expected behavior. Always consult relevant specs before implementing or modifying features.

<!-- TODO: Add brief project description -->

## Git Workflow 

When committing changes, always ask the user which files/directories to include before proposing commits. Never assume all unstaged/staged changes should be committed.

## RPI Artifacts Directory

This project uses a `.rpi/` directory for persistent context:

```
.rpi/
├── research/      # Codebase research notes (optional, from /rpi-research)
├── designs/       # Solution designs (created by /rpi-propose)
├── plans/         # Implementation plans (created by /rpi-plan)
├── specs/         # Living behavioral specs
├── reviews/       # Verification reports
├── archive/       # Archived completed artifacts
```


### Development Pipeline

See `.rpi/PIPELINE.md` for the full workflow guide covering: Research → Propose → Plan → Implement.


## Development Conventions

Before implementing any changes, always: 1) Read the current version of each file you plan to modify, 2) Run the existing test suite to establish a baseline, 3) Implement changes incrementally — one logical unit at a time, 4) Run tests after each unit. If tests fail, fix before proceeding. Do not batch all changes and test at the end.
<!-- TODO: Add project-specific conventions -->
