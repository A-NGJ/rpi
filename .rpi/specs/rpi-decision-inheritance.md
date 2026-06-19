---
domain: skills
feature: rpi-decision-inheritance
last_updated: 2026-06-17T11:00:00+02:00
updated_by: .rpi/designs/2026-06-17-decision-inheritance-across-artifacts.md
---

# rpi decision inheritance

## Purpose

Define the behavioral contract for carrying decisions forward across the pipeline. When an upstream artifact records decisions, the downstream design or plan proposed from it must contain those decisions verbatim, each pointing back to the artifact where it was decided. The goal is traceability: any commitment a reader finds in a plan or design can be followed up the chain to its origin, instead of being silently re-derived or lost. Covers when decisions are inherited, how provenance is attached, how inherited decisions stay distinct from newly-made ones, and how the absence of upstream decisions is handled.

## Scenarios

### Recorded research decisions appear in a design proposed from it
Given a research note that records two decisions
When a design is proposed from that research note
Then both decisions appear in the design, grouped under an inherited-decisions heading, each shown with the research note as its source

### A design's own decisions are inherited by a plan built from it
Given a design that records its own decisions and was itself proposed from a research note
When a plan is created from that design
Then the plan carries forward both the design's decisions and the research note's decisions, each attributed to the artifact that originally recorded it

### Inherited decisions stay distinct from decisions made at the current stage
Given a design that inherits decisions from upstream and also settles a new decision of its own
When the design is read
Then the inherited decisions and the new decision are shown under separate headings, so a reader can tell which commitments were carried forward and which were made here

### Provenance points to the true origin across multiple hops
Given a chain where a plan descends from a design that descends from a research note, and a particular decision was first recorded in the research note
When the plan presents that inherited decision
Then it attributes the decision to the research note where it originated, not merely to the design it passed through

### No upstream decisions means no inherited block is fabricated
Given an upstream artifact that records no decisions
When a design or plan is created from it
Then no inherited-decisions block is added, and the downstream artifact is produced normally without an empty or invented section

### An inherited decision can be read without opening the source artifact
Given a plan that inherited a decision from a research note which has since been archived
When a reviewer reads the plan in isolation
Then the decision's text is present in the plan itself, so the reviewer can understand what was decided without retrieving the original artifact, while still seeing the path it came from

### Inherited decisions are available for a later drift check
Given a plan that carries inherited decisions and a set of implementation phases
When the plan is audited for consistency
Then each inherited decision is present and attributed in a form the audit can compare phases against, so a phase that contradicts an inherited commitment can be surfaced

### Each inherited decision names exactly one source
Given a downstream artifact that inherits several decisions from different upstream artifacts
When the inherited decisions are presented
Then each decision is shown with the single artifact path that recorded it, grouped by source, with no decision left unattributed

## Constraints

- Inheritance happens only when an upstream artifact is provided and that artifact (or one further up its chain) records decisions; otherwise no inherited block is produced.
- Inherited decisions are carried forward verbatim, not paraphrased, so the downstream meaning matches the upstream commitment.
- Every inherited decision names the artifact path where it was originally recorded; an unattributed inherited decision is invalid.
- On a multi-hop chain, attribution points to the artifact that first recorded the decision, not to an intermediate artifact it passed through.
- Inherited decisions (carried forward, with source) are kept under a heading distinct from the artifact's own decisions (made at this stage).
- The downstream artifact is self-contained: the inherited decision's text is present in it, so it is readable without opening the source artifact.
- Recording decisions in an artifact is optional; an artifact with nothing to record omits its decisions section, and downstream consumers treat that as the empty case.
- Inheritance is best-effort and never blocks drafting: if no upstream decisions exist, the design or plan is still produced.

## Out of Scope

- Detecting or resolving a downstream phase that contradicts an inherited decision — inheritance makes the decisions available; the drift check that compares phases against them is a separate contract.
- A user-invocable command or flag dedicated to decision inheritance; it is an internal behavior of the propose and plan stages.
- Inheriting sections other than decisions (such as constraints or risks) across the chain.
- Merging, deduplicating, or reconciling conflicting decisions across the chain; decisions are carried forward faithfully, not arbitrated.
- Promoting interview or stress-test output into a decision artifact, or adding a discovery stage that produces decisions.
- Retroactively adding inherited decisions to artifacts that were already written.
