---
domain: RPI Benchmark Project
feature: benchmark
last_updated: 2026-04-04T22:30:00+02:00
updated_by: .rpi/designs/2026-04-03-benchmark-project-for-rpi-flow-quality-testing.md
---

# RPI Benchmark Project

## Purpose

A self-contained multi-language project (Go + Python + TypeScript) with a defined feature request and pre-written tests that enables quality comparison between RPI-assisted and prompt-only development approaches. Quality is measured as the number of passing tests after implementation.

## Scenarios

### Benchmark project has correct structure
Given the benchmark project at `testdata/benchmark-project/`
When inspecting the directory
Then it contains Go, Python, and TypeScript source code, shared JSON Schema files, a feature request (`task.md`), and setup/verify scripts

### Setup creates clean git environment with RPI
Given an empty target directory
When `setup.sh <target-dir>` runs
Then a git repository is initialized with all project files committed and `.rpi/` configured, without requiring network access

### Baseline tests pass before implementation
Given a freshly setup benchmark project with no modifications
When `verify.sh` runs
Then all 13 existing tests pass and all 10 new feature tests fail

### Feature request is fully specified in task.md
Given the `task.md` file
When read by an implementer
Then the conditional transform feature is unambiguously specified and requires changes in all 3 languages plus the shared schema

### New tests verify feature across all languages
Given pre-written tests for the conditional transform feature
When a correct implementation is provided
Then 10 new tests pass covering Go (registry, predicate match/skip), Python (decorator, operators), and TypeScript (config parsing, type definitions)

### Verify script reports per-language scores
Given any state of the benchmark implementation
When `verify.sh` runs
Then it reports pass/fail counts per language and a total score out of 23

### Codebase exercises index tools across languages
Given a setup benchmark project
When the codebase index is built
Then it contains symbols from all 3 languages with cross-language references via shared schema definitions

## Constraints
- All tests are deterministic — no network, no randomness
- Setup is idempotent — running twice overwrites cleanly
- Each language runs tests independently
- Do not require Docker or specific runtime versions beyond Go 1.21+, Python 3.10+, Node 18+
- Do not include the solution — tests define behavior but implementation is what's tested

## Out of Scope
- Automated A/B runner
- Statistical analysis of multiple runs
- Token usage or latency measurement
- Rust language support in the benchmark
- CI integration for the benchmark
