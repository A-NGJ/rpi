---
domain: skills
feature: rpi-claim-grounding
last_updated: 2026-06-17T10:00:00+02:00
updated_by: .rpi/designs/2026-06-17-grounding-subagent-for-rpi-verify.md
---

# rpi claim grounding

## Purpose

Define the behavioral contract for the grounding adjudication that runs after the verify stage drafts its findings. An independent read-only pass re-grounds each finding against actual repository state and tags it Verified, Weakened, or Falsified with a one-line evidence pointer. The goal is fewer false-positive blockers in the verification review: a finding may only keep blocking severity if it is confirmed against repo evidence. Covers when grounding runs, what each verdict means for the finding's place in the blocking set, how evidence is cited, and how the stage degrades when grounding is unavailable.

## Scenarios

### A finding contradicted by the code is falsified and excluded from blockers
Given a verification review whose draft contains a blocker claiming a function was removed, but that function still exists in the implementation
When grounding runs over the draft findings
Then the finding is marked Falsified with a one-line pointer to the place the function still lives, and it is excluded from the set of blockers presented to the user

### A confirmed finding keeps its severity with an evidence pointer
Given a draft blocker claiming a required behavior is unimplemented, and the behavior genuinely is absent from the code and tests
When grounding re-grounds the finding against repository state
Then the finding is marked Verified, keeps its blocker severity, and carries a one-line evidence pointer the reader can check

### A partially-true finding is weakened and demoted out of the blocking set
Given a draft blocker claiming a planned test is missing, but a test for that behavior exists under a different path and is thin rather than absent
When grounding adjudicates the finding
Then the finding is marked Weakened, gains a caveat describing what softened it, and is demoted so it no longer counts as a blocker

### Only confirmed findings may block
Given a draft finding asserts blocker severity but grounding cannot anchor the claim to any file, command output, or spec scenario
When grounding completes
Then the finding does not retain blocker severity — an unconfirmed finding is presented at most as a non-blocking item, never in the blocking set on the model's assertion alone

### Grounding is skipped for trivial reviews
Given a verification draft with no blockers and only one or two minor notes
When the verify stage evaluates whether grounding is warranted
Then grounding does not run, and the review is presented without grounding verdicts because the false-positive cost of the minor notes does not justify a second adjudication pass

### The review shows each finding's verdict and a grounding summary
Given grounding has run over a draft with several findings of mixed verdicts
When the verification review is presented to the user
Then each surviving finding shows its verdict and evidence pointer, and the summary reports the before-and-after counts (for example, how many drafted blockers were confirmed versus dropped)

### Grounding degrades gracefully when unavailable
Given the verify stage runs in an environment where the independent grounding adjudicator is not installed
When the stage produces its review
Then the review is presented with the original drafted findings, with no grounding verdicts, and an explicit note that grounding was skipped — the stage never emits a broken or half-annotated review

### The grounding pass never modifies code or artifacts
Given grounding is adjudicating a set of findings
When it gathers evidence to confirm, weaken, or falsify each claim
Then it only reads repository state and runs read-only checks, and it makes no edits to source files, tests, or the review document — adjudication is observation, not remediation

## Constraints

- Grounding is an independent adjudication pass distinct from the pass that drafted the findings; it does not re-derive findings from scratch, it re-grounds the existing ones.
- Each finding receives exactly one verdict: Verified, Weakened, or Falsified.
- A finding may retain blocking severity only when it is Verified against a citable anchor (a file:line, a command output, or a spec scenario). Weakened and Falsified findings are never in the blocking set.
- Falsified findings are excluded from the blocking set and are either dropped or recorded transparently (e.g., struck through) so a bad demotion is visible to the reader.
- Weakened findings survive with a caveat and are demoted out of the blocking set.
- Every verdict carries a one-line evidence pointer the reader can independently verify.
- Grounding runs only when the draft warrants it — at minimum when a blocker is present or the finding count exceeds a small threshold — so trivial reviews are not charged a second pass.
- Grounding is read-only: it never writes or edits source files, tests, or the review document.
- When grounding cannot run (unavailable in the host environment), the verify stage degrades to presenting the drafted findings without verdicts and states that grounding was skipped.
- The blocking set presented to the user is always the post-grounding set when grounding has run.

## Out of Scope

- Grounding any stage other than verify (design constraints, plan phases, research findings).
- A user-invocable grounding command or standalone pipeline stage — grounding is an internal pass of verify.
- Numeric confidence scores or per-finding probabilities; the contract is the three-state verdict.
- A user-tunable threshold for when grounding fires; the warranting condition is fixed.
- Auto-applying fixes for Weakened or Falsified findings; remediation stays with the human and the implement stage.
- Retroactively re-grounding previously written verification reports.
- Adding new deterministic checks beyond those the verify stage already has available.
