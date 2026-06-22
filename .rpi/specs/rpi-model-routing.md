---
domain: Per-skill / per-stage model + reasoning-effort routing
feature: rpi-model-routing
last_updated: 2026-06-17
updated_by: .rpi/designs/2026-06-17-per-skill-model-effort-routing.md
---

# Per-skill / per-stage model + reasoning-effort routing

## Purpose

Give every RPI stage a recommended model tier and reasoning-effort level so users
spend premium reasoning budget on hard stages (propose, verify, plan, diagnose)
and a cheaper tier on mechanical ones (commit, archive, spec-sync, handoff). The
guidance is advisory for plain skills and enforceable only where a stage runs as
a subagent.

## Scenarios

### Mechanical stage is recommended a cheaper tier than a reasoning stage
Given the routing guidance is published
When a user looks up the recommendation for a mechanical stage like committing changes
Then the recommended model tier and effort for that stage are lower than the recommendation for proposing a design

### Hard reasoning stages get the highest recommendation
Given the routing guidance is published
When a user looks up the recommendation for proposing, verifying, planning, or diagnosing
Then each is recommended the premium model tier and high reasoning effort

### Every stage has a recommendation with no gaps
Given the routing guidance is published
When a user looks up any one of the workflow stages
Then a model tier and an effort level are stated for that stage, and no stage is left unspecified

### Guidance is honest about not being enforced for skills
Given a user reads the routing guidance
When they reach the part describing how it takes effect
Then it states plainly that the platform does not auto-switch a skill's model and that they apply the recommendation themselves when choosing a session model

### Guidance is reachable from the user-facing docs
Given a user is reading the project's stage overview or workflow guide
When they want to know which model to use for a stage
Then the routing recommendation is discoverable from those documents

### A configured override for a stage wins over the default
Given a routing configuration that sets a cheaper tier for the commit stage and a premium default
When the recommendation for the commit stage is resolved
Then the commit stage resolves to the cheaper tier rather than the premium default

### A stage-level override outranks a skill-level one
Given a routing configuration that sets one tier at the stage level and a different tier at the skill level for the same stage
When the recommendation for that stage is resolved
Then the stage-level setting is used and the skill-level setting is overridden

### Recommendations are available with no configuration present
Given no routing configuration file exists
When a user asks for the recommendation for any stage
Then a recommendation is still returned from the built-in defaults

## Constraints

- The recommendation for any mechanical stage (commit, archive, spec-sync,
  handoff, explain) is never a higher model tier or higher effort than the
  recommendation for any core reasoning stage (propose, verify, plan, diagnose).
- The published guidance must state that the recommendation is user-applied for
  skills and is not automatically enforced by the platform.
- A recommendation exists for every workflow stage; there are no unspecified
  stages.
- When a configuration is present, a more specific setting overrides a less
  specific one, resolved field by field, with stage-level taking precedence over
  skill-level and skill-level over defaults.
- With no configuration present, the built-in defaults still yield a complete
  recommendation for every stage.

## Out of Scope

- A separate adjudicator/advisor model for reviews.
- Named multi-stage workflow presets that bundle routing.
- Automatically switching the active model for a running skill.
- Detecting or installing the concrete models named by each tier.
