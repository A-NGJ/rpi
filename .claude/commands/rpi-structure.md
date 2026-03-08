---
description: Define file structure, module boundaries, and interfaces — use for greenfield projects or major reorganizations where file layout needs dedicated attention
model: opus
---

# Structure Definition

Map a design to concrete codebase structure — files, modules, interfaces, and dependencies.

**Prerequisite**: The `rpi` binary must be available in PATH. If not found, run `go build -o bin/rpi ./cmd/rpi` or use `rpi-init` to set it up.

**When to use this command**: For greenfield projects creating many new files, major module reorganizations, or when the file layout itself is complex enough to warrant a dedicated discussion. For most features, the design doc's optional "File Structure" section (via `/rpi-design`) and the plan's file listings (via `/rpi-plan`) provide sufficient coverage without a separate structure stage.

**This command OWNS**: file layout, module boundaries, public interfaces/contracts, dependency graph, and naming conventions.

**This command does NOT own**: codebase research (→ `/rpi-research`), architectural decisions (→ `/rpi-design`), or implementation code and phasing (→ `/rpi-plan`). Define *what* each file exports — the plan stage defines *how* to implement the internals.

## Initial Response

If a file path or design document was provided, read it fully and begin analysis. Otherwise respond:
```
I'll help you define the structure for your implementation.

Please provide:
1. The design document from `.thoughts/designs/`
2. Any additional context or preferences about code organization

Tip: `/rpi-structure .thoughts/designs/2025-01-08-authentication-flow.md`
```

## Process Steps

### Step 1: Context Gathering

1. **Read the design document fully** (no limit/offset) before spawning sub-tasks
2. **Resolve the artifact chain**: Run `rpi chain <design-path>`
   Read all linked research and other documents it identifies.
3. **Spawn parallel research sub-tasks** using the Task tool. Each sub-task should load the appropriate skill first, then perform its work:
   - Sub-task: "Load the `locate-codebase` skill, then find existing files/modules that will be affected by [feature]"
   - Sub-task (@codebase-analyzer): Understand current file organization, module patterns, and naming conventions
   - Sub-task: "Load the `find-patterns` skill, then find how similar features are structured in the codebase (directory layout, file naming, export patterns) for [feature]"
   - Sub-task: "Load the `locate-thoughts` skill, then find any existing structure docs or conventions documentation about [topic]"
4. **Wait for ALL sub-tasks to complete**
5. **Read all files identified by research tasks** - especially module entry points, index files, and existing interfaces
6. **Present structural analysis:**
   ```
   Based on the design and codebase analysis:

   Existing structure relevant to this work:
   - [directory/module] - [what it contains, how it's organized]
   - [file pattern] - [convention used, e.g., "one component per file with co-located test"]

   Conventions I've identified:
   - Naming: [file naming pattern, e.g., kebab-case, PascalCase]
   - Modules: [how modules are organized, e.g., barrel exports, flat structure]
   - Tests: [where tests live, naming convention]
   - Types: [where types/interfaces are defined]

   Questions before I propose the structure:
   - [Question about organization preference or convention]
   ```

### Step 2: Structure Proposal

After clarifications:

1. **Create a todo list** using TodoWrite to track structural decisions
2. **Map each design component to concrete files:**
   - Which existing files need modification?
   - What new files need to be created?
   - What's the dependency direction between modules?
3. **Present the proposed structure:**
   ```
   ## Proposed File Structure

   ### Modified Files:
   - `path/to/existing-file.ext` - [what changes and why]
   - `path/to/another-file.ext` - [what changes and why]

   ### New Files:
   - `path/to/new-file.ext` - [responsibility, what it exports]
   - `path/to/new-test-file.ext` - [what it tests]

   ### Module Boundaries:
   - [Module A] depends on [Module B] via [interface/contract]
   - [Module C] is independent, no new dependencies

   ### Dependency Direction:
   [Module A] -> [Module B] -> [Module C]
   (No circular dependencies)

   Does this organization make sense?
   ```

### Step 3: Interface Definition

After structural buy-in:

1. **Spawn parallel sub-tasks** if needed:
   - Sub-task: "Load the `find-patterns` skill, then find existing interface/type patterns to match for [component]"
   - Sub-task (@codebase-analyzer): Understand existing function signatures at integration points
2. **Define the key interfaces and contracts:**
   ```
   ## Key Interfaces

   ### [Interface/Contract 1]
   Location: `path/to/file.ext`
   ```[language]
   // The public API surface this module exposes
   [type definition or function signature]
   ```

   ### [Interface/Contract 2]
   Location: `path/to/file.ext`
   ```[language]
   [type definition or function signature]
   ```

   ### Integration Points:
   - [Existing file:line] will call [new interface] for [purpose]
   - [New file] will import from [existing module] for [purpose]

   Do these interfaces look right?
   ```

### Step 4: Write the Structure Document

**Create the structure doc**: Run `rpi scaffold structure --topic "..." --design <path> --write`

This creates `.thoughts/structures/YYYY-MM-DD-description.md` with frontmatter pre-populated (`date`, `author`, `git_commit`, `branch`, `repository`, `topic`, `tags`, `status: draft`, `related_design`, `related_research`).

**Fill in the document sections:**
- Overview (brief description of structural changes)
- Context (links to design and research docs)
- Conventions Applied (file naming, module pattern, test placement, type definitions)
- File Changes — Modified Files (current role, changes needed, reason) and New Files (responsibility, exports, depends on, depended on by)
- Module Boundaries (directory, public API, internal, dependencies)
- Key Interfaces (location, used by, interface definition code)
- Dependency Graph (visual representation + dependency rules)
- Migration Notes (if modifying existing structure)
- References

### Step 5: Review & Iterate

1. Present the draft structure document location
2. Iterate based on feedback
3. Ensure all interfaces are concrete enough for the planning stage
4. Mark status as `complete` when finalized

## Guidelines

1. **Follow Existing Conventions**: Match the codebase's naming, organization, and patterns exactly
2. **Be Interactive**: Get buy-in on file layout before defining interfaces
3. **Be Concrete**: Every file should have a clear responsibility and defined exports
4. **Be Minimal**: Don't create files or abstractions that aren't justified by the design
5. **No Circular Dependencies**: Ensure the dependency graph is a DAG
6. **No Open Questions**: Resolve all structural questions before finalizing

## Common Structure Patterns

**Feature Module:** `feature/index.ts` (barrel) + `feature/types.ts` + `feature/[component].ts` + `feature/__tests__/`

**API Endpoint:** Route definition + handler + validation + tests (mirror existing endpoint structure)

**Refactoring:** New structure alongside old -> Migrate consumers -> Remove old structure
