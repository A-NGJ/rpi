---
archived_date: "2026-04-02"
domain: design-artifact
id: DA
last_updated: 2026-03-19T09:51:50+01:00
status: archived
updated_by: .rpi/proposals/2026-03-19-rename-proposal-to-design.md
---

# design-artifact

## Purpose

The "design" artifact type replaces "proposal" as the output of the `/rpi-propose` command. It captures architecture (components, interfaces, data flow) with integrated decisions, scaffolded via `rpi scaffold design` and stored in `.rpi/designs/`.

## Behavior

### Scaffold

- **DA-1**: `rpi scaffold design --topic "foo"` creates a file at `.rpi/designs/YYYY-MM-DD-foo.md` using the `design.tmpl` template
- **DA-2**: The scaffolded design contains sections: Summary, Context, Constraints, Components, File Structure, Risks, Out of Scope, References
- **DA-3**: The scaffolded design frontmatter includes `tags: [design]` and `status: draft`
- **DA-4**: `rpi scaffold design --research <path>` populates `related_research` in frontmatter
- **DA-5**: `rpi scaffold plan --design <path>` populates `design` in plan frontmatter
- **DA-6**: `rpi scaffold verify-report --design <path>` populates `design` in verify-report frontmatter

### Scan and filter

- **DA-7**: `rpi scan --type design` returns only artifacts in `.rpi/designs/`
- **DA-8**: `rpi scan --design <path>` returns artifacts whose frontmatter `design` field matches the path
- **DA-9**: The `propose` and `proposal` type names are not recognized (no backward compat aliases)

### Chain resolution

- **DA-10**: The `design` frontmatter field is a link field — `rpi chain` follows it when resolving artifact chains
- **DA-11**: Artifacts in `.rpi/designs/` are inferred as type `design`

### Init

- **DA-12**: `rpi init` creates `.rpi/designs/` directory (not `proposals/`)
- **DA-13**: `rpi init --update` generates command files that reference `design` artifacts

### Template content

- **DA-14**: The design template does NOT include Interfaces or Data Flow sections — those are added by the command prompt when needed
- **DA-15**: The Components section supports inline decision documentation (no standalone Decisions section)

## Constraints

### Must
- The `propose` artifact type must be fully removed — no aliases, no fallbacks
- All frontmatter fields referencing proposals (`proposal:`) must become `design:`
- The `/rpi-propose` command file must stay named `rpi-propose.md`
- Template type label for `design` must be `"Design"`

### Must Not
- Must not include Interfaces or Data Flow in `design.tmpl` — these are command-prompt-driven
- Must not rename the `/rpi-propose` command itself
- Must not change the spec template structure

### Out of Scope
- Migration tooling for existing `.rpi/proposals/` directories
- Changes to spec template
- Restructuring other command files beyond reference updates

## Test Cases

### DA-1: Scaffold creates design in correct directory
- **Given** no existing designs **When** `rpi scaffold design --topic "auth-flow" --write` **Then** file created at `.rpi/designs/YYYY-MM-DD-auth-flow.md`

### DA-2: Scaffolded design has correct sections
- **Given** a scaffolded design **When** reading its content **Then** it contains h2 headers: Summary, Context, Constraints, Components, File Structure, Risks, Out of Scope, References

### DA-5: Plan scaffold links to design
- **Given** a design at `.rpi/designs/2026-03-19-foo.md` **When** `rpi scaffold plan --design .rpi/designs/2026-03-19-foo.md --write` **Then** plan frontmatter contains `design: .rpi/designs/2026-03-19-foo.md`

### DA-7: Scan filters by design type
- **Given** artifacts in `.rpi/designs/` and `.rpi/research/` **When** `rpi scan --type design` **Then** only design artifacts are returned

### DA-9: Propose type not recognized
- **Given** the updated CLI **When** `rpi scaffold propose --topic "foo"` **Then** error: invalid type

### DA-10: Chain resolves design links
- **Given** a plan with `design: .rpi/designs/2026-03-19-foo.md` **When** `rpi chain .rpi/plans/2026-03-19-bar.md` **Then** the chain includes the design artifact

### DA-12: Init creates designs directory
- **Given** no `.rpi/` directory **When** `rpi init` **Then** `.rpi/designs/` directory is created (not `proposals/`)
