---
domain: RPI Benchmark Project
id: BP
last_updated: 2026-04-03T00:00:00Z
status: active
updated_by: .rpi/designs/2026-04-03-benchmark-project-for-rpi-flow-quality-testing.md
---

# RPI Benchmark Project

## Purpose

A self-contained multi-language project (Go + Python + TypeScript) with a defined feature request and pre-written tests that enables quality comparison between RPI-assisted and prompt-only development approaches. Quality is measured as the number of passing tests after implementation.

## Behavior

### Project Structure
- **BP-1**: The benchmark project lives at `testdata/benchmark-project/` and contains Go (core/), Python (transforms/), and TypeScript (cli/) source code plus shared JSON Schema files (schemas/).
- **BP-2**: The project contains exactly 20 files: 15 source files, 3 test files, and 2 shell scripts (setup.sh, verify.sh).
- **BP-3**: `task.md` defines the feature request — adding a `conditional` transform — and is the only input given to both approaches.

### Setup
- **BP-4**: `setup.sh <target-dir>` copies the project to a new directory, runs `git init`, `git add .`, `git commit`, and `rpi init`.
- **BP-5**: After setup, `rpi index build` produces an index with symbols from all 3 languages and cross-package imports.
- **BP-6**: Setup must not require network access (no `go mod download`, `npm install` for the initial copy — dependencies are vendored or mocked).

### Existing Tests (Baseline)
- **BP-7**: All existing tests (13 total: 4 Go, 6 Python, 3 TypeScript) pass before the feature is implemented.
- **BP-8**: Existing tests cover: pipeline execution, schema validation, transform decoration, string cleaning, type conversion, field validation, config parsing.

### Feature Request
- **BP-9**: The feature request (`task.md`) asks for a `conditional` transform that wraps another transform and applies it only when a field matches a predicate (operator: equals, not_equals, contains, exists).
- **BP-10**: A correct implementation requires changes in all 3 languages plus the shared schema: Go (registry), Python (transform class), TypeScript (types + config), JSON Schema (transform definition).

### New Tests (Verification)
- **BP-11**: 10 new tests (3 Go, 4 Python, 3 TypeScript) are pre-written and fail before implementation.
- **BP-12**: Go tests verify: conditional transform is registered, applies inner transform when predicate matches, skips when predicate doesn't match.
- **BP-13**: Python tests verify: `@transform("conditional")` decorator, equals/not_equals/contains/exists operators, error on unknown operator.
- **BP-14**: TypeScript tests verify: conditional config parses correctly, builder resolves inner transform reference, type definitions include ConditionalParams.

### Scoring
- **BP-15**: `verify.sh` runs all test suites (`go test ./...`, `pytest`, `npm test`) and reports pass/fail counts per language and total.
- **BP-16**: Quality score = passing tests / 23. A score of 13/23 means no new tests pass (baseline only). A score of 23/23 means full correctness.

### Index Tool Exercisability
- **BP-17**: The Go code contains at least 8 exported symbols (functions, structs, interfaces) and 4 unexported symbols.
- **BP-18**: The Python code contains at least 5 classes and 10 functions/methods.
- **BP-19**: Cross-language references exist via shared schema: Python and TypeScript both import/reference paths that appear in the Go schema definitions.
- **BP-20**: `rpi index packages` returns at least 3 distinct packages (one per language component).

## Constraints

### Must
- All tests are deterministic — no network, no filesystem I/O beyond reading the schema files, no randomness
- setup.sh is idempotent — running it twice on the same target overwrites cleanly
- The codebase compiles/runs in each language independently (Go, Python, TypeScript each have their own entry point for tests)
- The feature request is solvable by reading only the existing code — no external documentation needed

### Must Not
- Must not require Docker or any container runtime
- Must not require specific versions of Go/Python/Node beyond reasonable minimums (Go 1.21+, Python 3.10+, Node 18+)
- Must not include the solution — the pre-written tests define behavior but the implementation is what's being tested
- Must not leak benchmark infrastructure into the main RPI codebase (no imports from testdata/)

### Out of Scope
- Automated A/B runner (invoking Claude with different tool configs)
- Statistical analysis of multiple runs
- Token usage or latency measurement
- Rust language support in the benchmark project
- CI integration for the benchmark

## Test Cases

### BP-4: Setup produces clean environment
- **Given** an empty directory `/tmp/bench-test` **When** `setup.sh /tmp/bench-test` runs **Then** the directory contains a git repo with all project files committed and `.rpi/` initialized

### BP-5: Index covers all languages
- **Given** a setup benchmark project **When** `rpi index build` runs **Then** `rpi index status` shows files in go, py, and ts languages with >0 symbols each

### BP-7: Baseline tests pass
- **Given** a setup benchmark project with no modifications **When** `verify.sh` runs **Then** 13/13 existing tests pass and 10/10 new tests fail

### BP-9: Feature fully specified
- **Given** `task.md` **When** read by an implementer **Then** the requirements are unambiguous — each of the 10 new tests maps to a specific sentence in the task description

### BP-15: Verify script reports correctly
- **Given** a complete correct implementation **When** `verify.sh` runs **Then** output shows 23/23 passing with per-language breakdown

### BP-16: Partial implementation scored
- **Given** an implementation that only adds the Go registry entry but no Python/TS code **When** `verify.sh` runs **Then** score is 16/23 (13 baseline + 3 Go new tests)
