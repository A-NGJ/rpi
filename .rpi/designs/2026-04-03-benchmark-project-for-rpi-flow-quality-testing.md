---
date: 2026-04-03T00:03:25+02:00
status: complete
tags:
    - design
    - testing
    - benchmark
topic: benchmark project for rpi flow quality testing
---

# Design: Benchmark Project for RPI Flow Quality Testing

## Summary

A self-contained, multi-language data pipeline project (`testdata/benchmark-project/`) that serves as an A/B testing harness for RPI. It provides a realistic codebase with a defined feature request and pre-written failing tests, enabling quality comparison between RPI-assisted development (MCP tools + structured workflow) and prompt-only approaches.

## Context

We have 23 MCP tools and a structured workflow (research -> propose -> plan -> implement -> verify) but no empirical way to measure whether they improve implementation quality. We need a controlled environment where:

1. The codebase is complex enough that index tools (symbol query, package overview, import graph) provide real navigation value
2. A feature request is well-defined enough that correctness is mechanically verifiable
3. Both approaches (RPI vs. prompt-only) start from identical baselines

The project must exercise the index tools specifically — multi-language, cross-package imports, mixed symbol kinds — because that's where the value hypothesis is strongest.

## Constraints

- Must support Go, Python, and TypeScript (the 3 most common RPI-indexed languages)
- Must be buildable/testable without external services (no databases, APIs, etc.)
- Must be small enough to set up quickly (~15-20 files) but large enough that grep-and-pray fails
- Pre-written tests must be deterministic — no timing, no network, no randomness
- The feature request must touch all 3 languages to exercise cross-language understanding
- Setup script must produce a clean git repo with `rpi init` already run

## Components

### 1. Pipeline Core (Go) — 6 files

The engine that loads, validates, and runs transformation pipelines.

```
core/
├── pipeline.go          # Pipeline struct, Run() method, step execution loop
├── registry.go          # Plugin registry — Register(), Lookup() by name
├── schema.go            # Shared types: Record, Field, FieldType, ValidationError
├── validator.go         # Schema validation — validates records against field definitions
├── pipeline_test.go     # Tests for pipeline execution with mock transforms
└── validator_test.go    # Tests for schema validation
```

Key design choices:
- **Plugin registry pattern**: Transforms register by name string, looked up at runtime. This means the implementer must find the registry and follow the registration pattern — exactly the kind of task where `rpi index query register --kind function` helps.
- **Record type**: `map[string]interface{}` with typed Field definitions. Simple but requires understanding the schema to write correct transforms.

### 2. Transforms (Python) — 5 files

Built-in data transformations that plug into the Go pipeline via a shared JSON Schema contract.

```
transforms/
├── __init__.py          # Transform base class, @transform decorator
├── clean.py             # StringCleaner: trim, lowercase, strip_html
├── convert.py           # TypeConverter: string->int, date parsing, bool coercion
├── validate.py          # FieldValidator: required, min/max, regex pattern
├── test_transforms.py   # Tests for all transforms including the one to be added
```

Key design choices:
- **Decorator-based registration**: `@transform("name")` registers a class. Mirrors the Go registry pattern but in Python idiom. Implementer must find and follow this pattern.
- **Shared JSON Schema**: Transforms declare their config schema in `schemas/transforms.json`. The implementer must update this file too.

### 3. Config & CLI (TypeScript) — 5 files

Configuration parser and pipeline builder that reads YAML configs and validates them against the JSON Schema.

```
cli/
├── src/
│   ├── config.ts        # Config parser — reads YAML, resolves transform references
│   ├── builder.ts       # PipelineBuilder — assembles pipeline from config
│   ├── types.ts         # TypeScript interfaces matching Go types + JSON Schema
│   └── config.test.ts   # Tests for config parsing including new transform config
└── package.json
```

Key design choices:
- **TypeScript interfaces mirror Go structs**: `types.ts` defines `PipelineConfig`, `TransformStep`, `FieldDefinition` matching Go's `schema.go`. This cross-language dependency is exactly what `rpi index imports` and `rpi index importers` surface.

### 4. Shared Schema — 2 files

```
schemas/
├── pipeline.json        # JSON Schema for pipeline configuration files
└── transforms.json      # JSON Schema for transform declarations (name, params, config)
```

These are the glue. Both Python and TypeScript validate against these schemas. The Go core defines the canonical types that the schemas describe.

### 5. Setup & Verification — 2 files

```
setup.sh                 # Creates clean git repo, runs rpi init, builds index
task.md                  # The feature request given to both approaches
verify.sh                # Runs all test suites, reports pass/fail per language
```

## The Feature Request (task.md)

> **Add a `conditional` transform**
>
> A new transform that wraps another transform and only applies it when a field matches a predicate.
>
> Example config:
> ```yaml
> steps:
>   - transform: conditional
>     params:
>       field: status
>       operator: equals
>       value: "active"
>       then: clean.lowercase
> ```
>
> Requirements:
> 1. Register the `conditional` transform in the Go plugin registry
> 2. Implement the Python transform class following the existing decorator pattern
> 3. Add the transform's config schema to `schemas/transforms.json`
> 4. Update TypeScript types and config parser to handle the `conditional` transform
> 5. All existing tests must continue to pass
> 6. New pre-written tests for `conditional` must pass

This feature was chosen because it:
- Requires understanding the plugin registration pattern (Go)
- Requires following the decorator pattern (Python)
- Requires updating shared schema (JSON)
- Requires updating types and parsing logic (TypeScript)
- Has a "wraps another transform" aspect that tests deeper architectural understanding

## Scoring

Each language has pre-written tests with the following breakdown:

| Language | Existing tests (must stay green) | New tests (for conditional) | Total |
|----------|----------------------------------|-----------------------------|-------|
| Go       | 4                                | 3                           | 7     |
| Python   | 6                                | 4                           | 10    |
| TypeScript | 3                              | 3                           | 6     |
| **Total** | **13**                          | **10**                      | **23** |

**Quality score** = number of passing tests after implementation / 23

Additional qualitative criteria (human-evaluated):
- Does the implementation follow existing patterns? (e.g., uses `@transform` decorator, not a manual class)
- Does it update all necessary files? (schema, types, registry, implementation, config)
- Are there regressions in existing tests?

## File Structure

```
testdata/benchmark-project/
├── setup.sh
├── verify.sh
├── task.md
├── schemas/
│   ├── pipeline.json
│   └── transforms.json
├── core/
│   ├── go.mod
│   ├── pipeline.go
│   ├── registry.go
│   ├── schema.go
│   ├── validator.go
│   ├── pipeline_test.go
│   └── validator_test.go
├── transforms/
│   ├── __init__.py
│   ├── clean.py
│   ├── convert.py
│   ├── validate.py
│   └── test_transforms.py
└── cli/
    ├── package.json
    ├── tsconfig.json
    └── src/
        ├── config.ts
        ├── builder.ts
        ├── types.ts
        └── config.test.ts
```

Total: 20 files (15 source + 3 test + 2 scripts)

## Risks

1. **Go module in a subdirectory** — The benchmark project's Go code needs its own `go.mod`. The setup script handles this, but it means the parent repo's Go toolchain doesn't interfere.
2. **Python/TS tooling requirements** — Running the benchmark requires `python3`/`pytest` and `node`/`npm`. The verify script should check for these and fail early with clear messages.
3. **Test determinism** — All tests are pure computation (no I/O, no network). Randomness risk is zero.
4. **Feature request too easy** — If the task is trivially solvable without understanding the codebase, the A/B comparison is meaningless. The "wraps another transform" aspect adds enough architectural depth. Can iterate if needed.

## Out of Scope

- Automated A/B runner (running Claude twice with different tool configs) — that's a future concern
- Statistical significance / multiple runs — this is a qualitative first pass
- Performance benchmarking (latency, token usage) — focus is on correctness only
- Support for Rust in the benchmark project — 3 languages is sufficient

## References

- Prior research conversation on this topic (2026-04-02)
- RPI MCP tools: 23 tools registered in `cmd/rpi/serve.go`
- Index spec: `.rpi/specs/index-expansion.md`
