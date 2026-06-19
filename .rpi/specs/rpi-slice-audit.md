---
domain: rpi-slice-audit
feature: rpi-slice-audit
last_updated: 2026-06-17T14:00:00+02:00
updated_by: .rpi/designs/2026-06-17-slice-pre-lock-audit.md
---

# rpi-slice-audit

## Purpose

Before a freshly-drafted plan or design is finalized and shown to the user for approval, an independent adversarial pass audits each slice of the work — each plan phase, or each design component — for internal incoherence. It catches a phase that depends on something only a later phase produces, a slice that references work no other slice ever does, a slice that contradicts a decision already recorded upstream, and acceptance criteria or planned files that map to no actual work. Findings are surfaced at the approval gate so the user never approves a plan or design that does not hold together. In interactive runs the user must resolve or waive each finding before approving; in fast-forward runs findings are recorded and do not stop the run, except a coverage gap severe enough to invalidate the plan, which stops it regardless.

## Scenarios

### A coherent multi-phase plan passes the audit invisibly
Given a drafted plan whose phases each build only on earlier phases and whose every success criterion and planned file maps to real work
When the planning step runs the pre-lock audit before asking for approval
Then the user is shown a clean audit result and proceeds directly to the normal approval gate, with the plan unchanged in shape from one produced before this audit existed.

### A phase that relies on later work is flagged before approval
Given a drafted plan in which one phase depends on a file or capability that only a later phase produces
When the pre-lock audit runs
Then the user is shown a forward-reference finding identifying both phases, before the approval gate, and the plan is not finalized until the user resolves or waives the finding.

### A planned file that no phase produces is flagged
Given a drafted plan in which a phase references a file that no phase in the plan creates
When the pre-lock audit runs
Then the user is shown a coverage finding naming the unmet reference before the approval gate, and cannot approve the plan until it is resolved or waived.

### An acceptance criterion that maps to no work is flagged
Given a drafted plan that lists a success criterion which no phase's tasks actually deliver
When the pre-lock audit runs
Then the user is shown an orphaned-criterion finding before the approval gate, and the plan is treated as having a coverage gap.

### A design component that contradicts an upstream decision is flagged
Given a drafted design in which one component describes behavior that conflicts with a decision recorded earlier in the same design
When the pre-lock audit runs before the design is presented for approval
Then the user is shown a decision-drift finding naming the conflicting component and decision, and the design is not finalized until the user resolves or waives it.

### The audit is independent of the author's perspective
Given any drafted plan or design under audit
When the pre-lock audit runs
Then the audit is performed by an independent read-only pass that produces only findings and never edits, rewrites, or saves any plan, design, spec, or source file.

### Fast-forward records findings without stopping, but a hard coverage gap still stops it
Given a drafted plan is produced with the fast-forward option
When the pre-lock audit finds non-coverage issues such as a forward-reference or a contradicted decision
Then those findings are recorded with the plan and the run proceeds without pausing; but if the audit finds a hard coverage gap — a criterion or planned file mapping to no work — the run stops and surfaces that gap even under fast-forward.

### A trivial single-slice plan skips the cross-slice audit
Given a drafted standalone plan with a single phase and no upstream design
When the planning step reaches the pre-lock audit
Then no cross-slice forward-reference or decision-drift check is performed (a single slice cannot reference a sibling or an upstream decision), the lightweight coverage check still runs, and the user proceeds to approval without an unnecessary audit pass.

## Constraints

- The audit always runs before the artifact is finalized and before the user is asked to approve — never after approval and never after implementation.
- The audit is read-only: it emits findings and never modifies any plan, design, spec, or source file.
- In interactive runs, every finding must be resolved or explicitly waived before the user can approve the plan or design.
- Under the fast-forward option, findings are recorded and non-blocking, with the single exception of a hard coverage gap (a criterion or planned file mapping to no work), which blocks even under fast-forward.
- The fast-forward option and the grilling option remain mutually exclusive; when grilling is active, the audit runs before the grilling pass so grilling operates on already-audited slices.
- A plan or design that passes the audit cleanly is unchanged in shape and behavior from one produced before the audit existed.
- When a design is split into sibling plans, the audit runs over each sibling plan's own phases, and references that cross between sibling plans are checked for coverage.
- A standalone plan with no upstream design is not subject to the decision-drift check; the other checks still apply.
- The coverage and ordering checks are deterministic — the same drafted plan against the same working tree yields the same coverage and forward-reference verdict every time.

## Out of Scope

- Verifying an implementation after the plan has been executed — that is the separate verification stage.
- Automatically fixing any finding — the audit reports; the author or user resolves.
- Reconciling slices across unrelated plans or designs — the audit is scoped to one plan's phases or one design's components (and the sibling plans of a single split).
- Detecting semantically near-duplicate slices — coverage is structural, not similarity-based.
- Restructuring or renaming phases or components — the audit reads existing shapes without reshaping them.
- Deciding whether a design should be split into sibling plans — that remains a separate, independent step.
- Auditing artifacts other than plans and designs (research notes, specs, diagnoses).
