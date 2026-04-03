---
domain: rpi-status
id: ST
last_updated: 2026-04-04T00:00:00Z
status: active
updated_by: .rpi/designs/2026-04-04-rpi-status-dashboard-command.md
---

# rpi status dashboard command

## Purpose

Provide a single CLI command (`rpi status`) that aggregates all RPI artifact metadata into a human-readable dashboard showing artifact counts, stale warnings, active plan progress, and archive readiness.

## Behavior

### Artifact Summary
- **ST-1**: Group all non-archived artifacts by type (`spec`, `plan`, `design`, `research`, `review`) and status, displaying counts for each type-status combination.
- **ST-2**: Omit types with zero artifacts from the summary (no empty rows).
- **ST-3**: Exclude the `archive/` subdirectory from all scanning.

### Staleness Detection
- **ST-4**: Flag artifacts with status `draft` or `active` whose frontmatter date exceeds the staleness threshold (default: 14 days).
- **ST-5**: Use `date` field for plans/designs/research/reviews, `last_updated` field for specs.
- **ST-6**: Accept `--stale-days N` flag to override the default threshold.
- **ST-7**: Skip artifacts with missing or unparseable date fields silently (no error, no false positive).
- **ST-8**: Display each stale artifact with its path, status, and age in days.

### Active Plan Chains
- **ST-9**: For each plan with status `active` or `draft`, resolve one level of frontmatter links (the plan's direct `design`, `spec`, and other link fields).
- **ST-10**: Linked artifact sub-rows are not shown under Active Plans in text output. Links remain available in JSON output (`active_plans[].links`).
- **ST-11**: Parse plan checkboxes (`- [ ]` / `- [x]`) and display checked/total count with percentage.
- **ST-12**: Plans with zero checkboxes show no progress indicator (not "0/0 (0%)").

### Active Artifact Sections
- **ST-18**: When active specs exist, display an "Active Specs" section listing each by name (title from frontmatter, falling back to filename). Omit section when no active specs exist.
- **ST-19**: Section order: Artifacts → Active Plans → Active Specs → Stale → Ready to Archive. Each section is omitted when empty.

### Archive Readiness
- **ST-13**: List artifacts with archivable status (`complete`, `superseded`, `implemented` — or `superseded` only for specs) that have zero active references.
- **ST-14**: Display as a summary count grouped by type (e.g., "2 artifacts (1 design, 1 spec)").

### Output Format
- **ST-15**: Default output is human-readable text with aligned columns.
- **ST-16**: `--format json` outputs a structured JSON object with keys: `summary`, `stale`, `active_plans`, `archivable`.
- **ST-17**: Exit code 0 on success, 1 on errors (filesystem, parse failures).

## Constraints

### Must
- Compose only existing internal packages (`scanner`, `chain`, `frontmatter`, verify's `parseCheckboxes`)
- Register as a Cobra subcommand via `init()` following existing patterns
- Support `--rpi-dir` flag consistent with other commands
- Produce stable, deterministic output ordering (sorted by type, then path)

### Must Not
- Write to any files or modify artifact state
- Introduce new dependencies beyond the standard library and existing imports
- Include archived artifacts in any section
- Emit color codes or Unicode special characters in text output

### Out of Scope
- MCP tool registration
- Filtering flags (use `rpi scan` for filtered queries)
- Archive directory listing
- Multi-level chain resolution (use `rpi chain` for full depth)
- File modification time (`os.Stat`) — frontmatter dates only

## Test Cases

### ST-1: Artifact summary grouping
- **Given** `.rpi/` contains 2 active specs, 1 draft plan, and 1 complete design **When** `rpi status` is run **Then** output shows `specs: 2 active`, `plans: 1 draft`, `designs: 1 complete`

### ST-2: Empty types omitted
- **Given** `.rpi/` contains only specs and plans (no designs, research, reviews) **When** `rpi status` is run **Then** output contains no `designs`, `research`, or `reviews` rows

### ST-4: Stale artifact detection
- **Given** a plan with `date: 2026-03-01` and status `draft` **When** `rpi status` is run on 2026-04-04 (34 days later, > 14 day default) **Then** the plan appears in the Stale section with "34d ago"

### ST-6: Custom stale threshold
- **Given** a plan with `date: 2026-03-25` (10 days old) **When** `rpi status --stale-days 7` is run **Then** the plan appears in Stale section
- **Given** the same plan **When** `rpi status --stale-days 14` is run **Then** the plan does not appear in Stale section

### ST-7: Missing date skipped
- **Given** an active spec with no `last_updated` field **When** `rpi status` is run **Then** the spec appears in the summary but not in the Stale section, and no error is printed

### ST-9: One-level chain resolution
- **Given** an active plan linking to `design: .rpi/designs/foo.md` and `spec: .rpi/specs/bar.md`, where the design further links to `research: .rpi/research/baz.md` **When** `rpi status` is run **Then** the plan shows the design and spec, but not the research (one level only)

### ST-11: Checkbox progress
- **Given** an active plan with 3 `- [x]` and 7 `- [ ]` checkboxes **When** `rpi status` is run **Then** the plan shows "3/10 (30%)"

### ST-12: No checkboxes
- **Given** a draft plan with no checkboxes **When** `rpi status` is run **Then** the plan row has no progress indicator

### ST-13: Archive readiness
- **Given** a complete design with 0 references and a complete design with 2 references **When** `rpi status` is run **Then** only the unreferenced design appears in Ready to Archive

### ST-16: JSON output
- **Given** any `.rpi/` state **When** `rpi status --format json` is run **Then** output is valid JSON with keys `summary`, `stale`, `active_plans`, `archivable`
