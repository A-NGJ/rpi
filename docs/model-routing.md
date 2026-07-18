# Model & reasoning-effort routing

RPI stages differ wildly in how much reasoning they need. Proposing a design or
diagnosing a bug is hard reasoning; committing or archiving is mechanical work the
compiled CLI already does. This page recommends a **model tier** and a
**reasoning-effort** level for every stage so you spend premium budget where it
pays off and a cheap, fast tier on the rest.

> **Advisory, not enforced.** Claude Code does not auto-switch a skill's model;
> apply this with `/model`. Only subagents are enforced.

## Recommended routing table

| Stage / skill   | Model tier | Effort | Rationale |
|-----------------|-----------|--------|-----------|
| `rpi-propose`   | premium   | high   | Tradeoff analysis, spec authoring — hardest reasoning |
| `rpi-verify`    | premium   | high   | Adversarial conformance check; false negatives are costly |
| `rpi-plan`      | premium   | high   | Phase decomposition, dependency reasoning |
| `rpi-diagnose`  | premium   | high   | Iterative root-cause; needs deep reasoning |
| `rpi-blueprint` | premium   | high   | Fused research → design → plan reasoning in a single pass |
| `rpi-revise`    | premium   | high   | Re-plans affected phases; dependency/ordering re-reasoning |
| `rpi-spec`      | premium   | high   | Spec + goal-envelope authoring — ranks with propose |
| `rpi-research`  | premium   | medium | Broad investigation; high model but exploration tolerates less peak effort |
| `rpi-implement` | premium   | medium | Code execution against a fixed plan; correctness matters, search space is bounded |
| `grill-me`      | premium   | medium | Interactive stress-test; benefits from a strong model, no peak effort |
| `rpi-explain`   | cheap     | low    | Narrate an existing diff — mechanical |
| `rpi-commit`    | cheap     | low    | Message + scans are deterministic-CLI-backed |
| `rpi-archive`   | cheap     | low    | File moves + frontmatter — CLI does the work |
| `rpi-spec-sync` | cheap     | low    | Drift detection is CLI-backed; rewrite is light |
| `rpi-handoff`   | cheap     | low    | Context capture to a temp file |
| `rpi-setup`     | cheap     | low    | Binary install/upgrade — mechanical |

Every mechanical stage (explain, commit, archive, spec-sync, handoff, setup) is
recommended a tier and effort no higher than any core reasoning stage (propose,
verify, plan, diagnose), so following the table never spends *more* on routine
work than on hard work.

## Tier → concrete model

Tiers are named abstractly so the table survives model renames. This is the one
place that maps a tier to a concrete model — update it here when models change.

| Tier      | Concrete model (current) | `/model` alias |
|-----------|--------------------------|----------------|
| `premium` | Claude Opus 4.8 (`claude-opus-4-8`) — strongest reasoning | `opus` |
| `cheap`   | Claude Haiku 4.5 (`claude-haiku-4-5`) — fast / low cost | `haiku` |

Effort levels (`high` / `medium` / `low`) map to the harness's reasoning-effort
control for the session or subagent.

## How it takes effect

- **Plain skills (most stages): advisory.** An Agent Skill's frontmatter is
  restricted to `name` + `description` (see `.rpi/specs/agent-skills.md`), and
  Claude Code has no per-skill model override. You apply the recommendation
  yourself by choosing a session model with `/model` before running the stage.
  Each affected `SKILL.md` carries a one-line note repeating its own tier+effort
  so the recommendation is visible at the point of use.
- **Subagents: enforced.** A subagent profile
  (`internal/workflow/assets/agents/<name>.md`) may carry a `model:` field, so a
  stage that delegates to a subagent runs on its assigned tier regardless of the
  session model. Today `rpi-verify` is pinned to the premium tier this way; as
  more grounding subagents are added, each takes its tier from this table.

## Layer 2 (deferred — designed, not yet built)

A future, opt-in layer adds a project-local override file and an advisory CLI so
the table can be tuned per project. It is **not built yet**; the table above is
the built-in baseline and is complete on its own.

- **Override file:** `.rpi/models.json` (project-local, travels with the repo's
  specs and plans), with an optional user-global fallback at
  `~/.config/rpi/models.json`. Shape:

  ```json
  {
    "defaults": { "model": "premium", "effort": "high" },
    "stages":   { "commit": { "model": "cheap", "effort": "low" } },
    "skills":   { "rpi-archive": { "model": "cheap" } }
  }
  ```

- **Cascade (most specific wins, composing per-field):** `stages[stage]` →
  `skills[skill]` → `defaults`, with each field (`model`, `effort`) resolved
  independently — so an entry may override only `model` and inherit `effort`. The
  built-in table above is the implicit baseline when no file is present, so
  recommendations exist with zero configuration.
- **Advisory CLI:** a read-only `rpi models <skill>` (and `rpi models` for the
  whole table) that prints the resolved tier+effort as JSON, carrying the same
  "not auto-enforced for skills" caveat. It changes nothing — routing only takes
  effect via `/model` or a subagent `model:` field.
