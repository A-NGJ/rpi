---
archived_date: "2026-04-02"
date: 2026-04-02T01:08:28+02:00
status: archived
tags:
    - plan
topic: readme-installation-rewrite
---

# readme-installation-rewrite — Implementation Plan

## Overview

Rewrite the README Installation/Quick Start section for clarity. Separate binary installation from project setup, fix orphaned numbering (line 71's "4."), and clean up indentation. Preserve all existing content.

**Scope**: 1 file modified (README.md)

## Phase 1: Rewrite Installation Section

### Overview
Restructure the Quick Start / Installation section into clearly numbered steps with consistent formatting.

### Tasks:

#### 1. Rewrite README.md lines 26–71
**File**: `README.md`
**Changes**:
- Replace the current "Quick Start" section (lines 26–71) with a restructured version
- Keep "Prerequisites" as-is
- Split into two clearly labeled steps:
  1. **Install `rpi`** — curl|bash (recommended), pinned version variant, and `go install` from source
  2. **Initialize your project** — `rpi init` with target flags, what it creates, and `rpi update` for syncing
- Fix the orphaned "4." (line 71) — integrate "Start your AI coding tool" as step 3
- Use consistent indentation (no unnecessary nesting)
- Preserve all existing information verbatim (URLs, commands, directory listings)

### Success Criteria:

#### Automated Verification:
- [x] No orphaned numbering — all numbered items are part of a continuous sequence
- [x] All commands from the original section are present: `curl -sSfL ...`, `VERSION=v0.1.0 curl ...`, `go install ...`, `rpi init`, `rpi update`, `rpi update --force`

#### Manual Verification:
- [x] README renders correctly in GitHub markdown preview (nested code blocks, lists)

### Commit:
- [x] Stage: `README.md`
- [x] Message: `docs: rewrite installation section for clarity`

---

## References
- Current README.md lines 26–71
