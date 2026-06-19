---
domain: fused research-to-plan fast-path
feature: rpi-blueprint
last_updated: 2026-06-17T10:00:00+02:00
updated_by: .rpi/designs/2026-06-17-rpi-blueprint-fused-fast-path.md
---

# rpi-blueprint fused research-to-plan fast-path

## Purpose

A one-shot path for solo / low-stakes work that goes from a research note or a
short problem statement straight to a phased implementation plan in a single
pass — without producing a separate reviewable design to approve. It still
guarantees a behavioral contract by emitting a minimal spec, and it records the
design reasoning it would otherwise have written into a design alongside the
plan. It declines and points the user at the full design path when the work
carries genuine tradeoffs or wide blast radius. It is the fused shortcut, as
distinct from fast-forward mode, which runs the full pipeline quickly but still
produces a design.

## Scenarios

### Low-risk research note yields a plan and a minimal spec in one pass
Given the user has a research note and a small, low-risk change
When they invoke the fused fast-path on it
Then a phased plan and a minimal behavioral spec are produced together in a single pass, with no separate design artifact and no separate design-approval gate

### A short problem statement with no research still works
Given the user provides only a short problem statement small enough to reason about in one pass
When they invoke the fused fast-path
Then condensed design reasoning is performed inline and a phased plan plus a minimal spec are produced, without first requiring a standalone research artifact

### Design reasoning is preserved alongside the plan
Given the fused fast-path produces a plan
When the user reads the plan
Then the chosen approach, the alternatives that were considered and dropped, and the blast-radius judgment that justified skipping the full design are recorded with the plan, so the reasoning remains auditable even though no separate design exists

### Genuine tradeoffs trigger a refusal and redirect
Given the work surfaces more than one approach a reasonable engineer would defend, or a wide-reaching multi-component change
When the user invokes the fused fast-path
Then it declines to produce a plan, explains in one or two sentences why the work needs a reviewable design, and suggests the full design path instead

### A minimal spec is always produced, never skipped
Given any successful fused fast-path run
When the plan is written
Then a behavioral spec covering the user-observable behavior is always written and linked, so no implementation is ever planned without a contract

### Fast-forward composes with the fused path
Given the user invokes the fused fast-path with fast-forward enabled
When the plan and minimal spec are written
Then the plan-approval pause is skipped and the run auto-continues into implementation and ends with a verification report, without ever producing or approving a separate design

### Refusal still stops a fast-forward run
Given a fused fast-path run with fast-forward enabled hits the tradeoff-or-blast-radius refusal condition
When the refusal fires
Then the run stops and informs the user rather than silently escalating into the full automated design pipeline, leaving the user to choose the design path deliberately

### Grilling interrogates the fused reasoning before the plan is written
Given the user invokes the fused fast-path with grill mode and the grilling capability is available
When the condensed design reasoning and phase outline are ready
Then the user is interrogated on that reasoning and the phasing in a single pass and revisions are applied before the plan is written; if grilling is unavailable the user is told and offered the standard approval gate

## Constraints

- A successful run always produces both a plan and a linked behavioral spec; a plan with no spec is a contract violation.
- The fused path never produces a separate reviewable design artifact — that is its defining difference from fast-forward mode, which does.
- The tradeoff-and-blast-radius refusal is an integrity gate, not a review pause: it fires even under fast-forward and stops the run.
- Fast-forward and grill modes are mutually exclusive in a single invocation, consistent with the rest of the pipeline.
- The minimal spec describes user-observable behavior only and carries fewer scenarios than a full design's spec, reflecting the small scope that qualifies for this path.
- On success the path suggests the next step (proceed to implementation), consistent with every other stage.

## Out of Scope

- Producing or approving a separate design artifact — the fused path structurally omits it.
- Dropping the behavioral spec — SDD is preserved; only the design deliverable is skipped.
- Auto-escalating into the full automated design pipeline when the refusal fires — the user is left to choose that deliberately.
- Using the fused path for work with genuine tradeoffs, multiple defensible approaches, or wide blast radius — that work is redirected to the full design path.
- Changing fast-forward behavior — the fused path composes with it but does not redefine it.
- A fused mode for any stage other than the research/problem-statement → plan path.
