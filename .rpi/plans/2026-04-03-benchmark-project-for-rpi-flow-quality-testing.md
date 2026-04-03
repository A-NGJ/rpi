---
date: 2026-04-03T00:08:40+02:00
design: .rpi/designs/2026-04-03-benchmark-project-for-rpi-flow-quality-testing.md
spec: .rpi/specs/benchmark-project-for-rpi-flow-quality-testing.md
status: active
tags:
    - plan
topic: benchmark project for rpi flow quality testing
---

# Benchmark Project for RPI Flow Quality Testing — Implementation Plan

## Overview

Create a self-contained multi-language data pipeline project at `testdata/benchmark-project/` that serves as an A/B testing harness for RPI. The project includes Go, Python, and TypeScript components with shared JSON Schema files, a defined feature request (conditional transform), and pre-written tests (13 baseline passing + 10 new failing).

**Scope**: 21 new files, 0 modified files

## Source Documents
- **Design**: .rpi/designs/2026-04-03-benchmark-project-for-rpi-flow-quality-testing.md
- **Spec**: .rpi/specs/benchmark-project-for-rpi-flow-quality-testing.md

---

## Phase 1: Go Core (pipeline engine + registry + schema)

### Overview
Create the Go pipeline engine with plugin registry, schema types, validation, and tests. This establishes the canonical types that the rest of the project mirrors. Covers spec BP-17 (exported/unexported symbols).

### Tasks:

#### 1. Schema types
**File**: `testdata/benchmark-project/core/schema.go`
**Changes**: Define `Record` (`map[string]interface{}`), `Field` (Name, Type, Required), `FieldType` enum (String, Int, Float, Bool, Date), `ValidationError`, and `TransformFunc` type. Mix of exported (8+) and unexported (4+) symbols to exercise index filtering.

#### 2. Plugin registry
**File**: `testdata/benchmark-project/core/registry.go`
**Changes**: Package-level registry map. `Register(name string, fn TransformFunc)` and `Lookup(name string) (TransformFunc, error)`. This is the pattern the implementer must discover and follow.

#### 3. Pipeline runner
**File**: `testdata/benchmark-project/core/pipeline.go`
**Changes**: `Pipeline` struct with `Steps []Step`, `Step` struct (TransformName, Params), `Run(records []Record) ([]Record, error)` that looks up transforms by name and applies them sequentially.

#### 4. Schema validator
**File**: `testdata/benchmark-project/core/validator.go`
**Changes**: `Validate(record Record, fields []Field) []ValidationError` — checks required fields, type matching.

#### 5. Go module
**File**: `testdata/benchmark-project/core/go.mod`
**Changes**: `module benchmark/core` with `go 1.21`.

#### 6. Pipeline tests
**File**: `testdata/benchmark-project/core/pipeline_test.go`
**Changes**: 4 baseline tests: TestPipelineRunsSteps, TestPipelineStopsOnError, TestRegistryLookup, TestRegistryUnknownTransform. 3 new tests (will fail): TestConditionalApplies, TestConditionalSkips, TestConditionalRegistered.

#### 7. Validator tests
**File**: `testdata/benchmark-project/core/validator_test.go`
**Changes**: Baseline tests for required field validation and type checking. These count toward the 4 baseline Go tests.

### Success Criteria:

#### Automated Verification:
- [x] `cd testdata/benchmark-project/core && go build ./...` compiles cleanly
- [x] `cd testdata/benchmark-project/core && go test ./... 2>&1 | grep -c PASS` shows 4 passing tests
- [x] `cd testdata/benchmark-project/core && go test ./... 2>&1 | grep -c FAIL` shows 3 failing tests (conditional)

### Commit:
- [ ] Stage: `testdata/benchmark-project/core/`
- [ ] Message: `feat(benchmark): add Go core — pipeline engine, registry, schema, validator`

---

## Phase 2: Shared JSON Schemas

### Overview
Create the JSON Schema files that serve as the cross-language contract. Both Python and TypeScript validate against these. Covers spec BP-19 (cross-language references).

### Tasks:

#### 1. Pipeline config schema
**File**: `testdata/benchmark-project/schemas/pipeline.json`
**Changes**: JSON Schema defining `PipelineConfig` with `name`, `description`, `fields` array (field definitions), and `steps` array (transform steps with name + params).

#### 2. Transforms schema
**File**: `testdata/benchmark-project/schemas/transforms.json`
**Changes**: JSON Schema defining transform declarations. Includes schemas for existing transforms (clean, convert, validate) but NOT conditional — the implementer must add that.

### Success Criteria:

#### Automated Verification:
- [x] Both files are valid JSON (parseable with `python3 -c "import json; json.load(open('...'))"`)
- [x] Schema structure matches what Phase 3 (Python) and Phase 4 (TypeScript) expect

### Commit:
- [ ] Stage: `testdata/benchmark-project/schemas/`
- [ ] Message: `feat(benchmark): add shared JSON Schema for pipeline config and transforms`

---

## Phase 3: Python Transforms

### Overview
Create the Python transform framework with decorator-based registration and 3 built-in transforms. Covers spec BP-18 (5+ classes, 10+ functions/methods).

### Tasks:

#### 1. Transform base + decorator
**File**: `testdata/benchmark-project/transforms/__init__.py`
**Changes**: `Transform` base class with `apply(record, params) -> record` method. `@transform("name")` decorator that registers classes in a module-level dict. `get_transform(name)` lookup function.

#### 2. String cleaner
**File**: `testdata/benchmark-project/transforms/clean.py`
**Changes**: `@transform("clean")` class `StringCleaner` with operations: trim, lowercase, strip_html. Params: `field` (target field), `operations` (list of ops to apply).

#### 3. Type converter
**File**: `testdata/benchmark-project/transforms/convert.py`
**Changes**: `@transform("convert")` class `TypeConverter` with conversions: string->int, date parsing (ISO format), bool coercion. Params: `field`, `target_type`.

#### 4. Field validator
**File**: `testdata/benchmark-project/transforms/validate.py`
**Changes**: `@transform("validate")` class `FieldValidator` with checks: required, min_length/max_length, regex pattern. Params: `field`, `rules` (dict of rule->value).

#### 5. Tests
**File**: `testdata/benchmark-project/transforms/test_transforms.py`
**Changes**: 6 baseline tests: test_clean_trim, test_clean_lowercase, test_convert_to_int, test_convert_to_bool, test_validate_required, test_validate_pattern. 4 new tests (will fail): test_conditional_equals, test_conditional_not_equals, test_conditional_contains, test_conditional_unknown_operator_raises.

### Success Criteria:

#### Automated Verification:
- [x] `cd testdata/benchmark-project && python3 -m pytest transforms/ -v 2>&1 | grep -c "PASSED"` shows 6 passing
- [x] `cd testdata/benchmark-project && python3 -m pytest transforms/ -v 2>&1 | grep -c "FAILED"` shows 4 failing

### Commit:
- [ ] Stage: `testdata/benchmark-project/transforms/`
- [ ] Message: `feat(benchmark): add Python transforms — clean, convert, validate with decorator registry`

---

## Phase 4: TypeScript CLI (config parser + builder + types)

### Overview
Create the TypeScript config layer with interfaces mirroring Go types, a config parser, and a pipeline builder. Covers spec BP-14 (TypeScript type definitions).

### Tasks:

#### 1. Type definitions
**File**: `testdata/benchmark-project/cli/src/types.ts`
**Changes**: Interfaces: `PipelineConfig`, `TransformStep`, `FieldDefinition`, `TransformParams`, `CleanParams`, `ConvertParams`, `ValidateParams`. These mirror Go's `schema.go` — the cross-language link that import analysis surfaces.

#### 2. Config parser
**File**: `testdata/benchmark-project/cli/src/config.ts`
**Changes**: `parseConfig(yamlString: string): PipelineConfig` — parses YAML config, validates transform names against known list, resolves params to typed objects. Uses `js-yaml` for parsing.

#### 3. Pipeline builder
**File**: `testdata/benchmark-project/cli/src/builder.ts`
**Changes**: `PipelineBuilder` class — takes `PipelineConfig`, validates all referenced transforms exist, builds execution plan. `build()` returns validated pipeline or throws with missing transform details.

#### 4. Tests
**File**: `testdata/benchmark-project/cli/src/config.test.ts`
**Changes**: 3 baseline tests: test config parses valid YAML, test unknown transform throws, test builder validates field references. 3 new tests (will fail): test conditional config parses, test builder resolves inner transform ref, test ConditionalParams type exists.

#### 5. Package config
**File**: `testdata/benchmark-project/cli/package.json`
**Changes**: Dependencies: `js-yaml`, `ajv` (JSON Schema validation). DevDependencies: `typescript`, `vitest`, `@types/js-yaml`. Scripts: `test`, `build`.

#### 6. TypeScript config
**File**: `testdata/benchmark-project/cli/tsconfig.json`
**Changes**: Standard strict TypeScript config targeting ES2020.

### Success Criteria:

#### Automated Verification:
- [x] `cd testdata/benchmark-project/cli && npm install && npm test 2>&1 | grep -c "pass"` shows 3 passing
- [x] `cd testdata/benchmark-project/cli && npm test 2>&1 | grep -c "fail"` shows 3 failing

### Commit:
- [ ] Stage: `testdata/benchmark-project/cli/`
- [ ] Message: `feat(benchmark): add TypeScript CLI — config parser, builder, types with vitest`

---

## Phase 5: Infrastructure (setup, verify, task)

### Overview
Create the setup script, verification script, and task document. Covers spec BP-4, BP-5, BP-9, BP-15, BP-16.

### Tasks:

#### 1. Setup script
**File**: `testdata/benchmark-project/setup.sh`
**Changes**: Takes target dir as argument. Copies project files (excluding setup.sh itself and .git), runs `git init && git add . && git commit -m "initial"`, runs `rpi init`, runs `rpi index build`. Checks for required tools (go, python3, node, rpi) upfront.

#### 2. Verify script
**File**: `testdata/benchmark-project/verify.sh`
**Changes**: Runs all 3 test suites, parses output for pass/fail counts per language. Reports: Go (N/7), Python (N/10), TypeScript (N/6), Total (N/23). Exit code 0 if all 23 pass, 1 otherwise.

#### 3. Task document
**File**: `testdata/benchmark-project/task.md`
**Changes**: The feature request for the conditional transform. Includes: description, example config YAML, requirements list (5 items mapping to the 5 areas that need changes), and acceptance criteria (all tests pass).

### Success Criteria:

#### Automated Verification:
- [x] `bash testdata/benchmark-project/setup.sh /tmp/bench-test` completes without errors
- [x] `/tmp/bench-test/.rpi/` directory exists with index.json
- [x] `cd /tmp/bench-test && bash verify.sh` reports 13/23 passing (baseline)

#### Manual Verification:
- [x] `task.md` is clear and unambiguous — each of the 10 new tests maps to a specific requirement

### Commit:
- [ ] Stage: `testdata/benchmark-project/setup.sh`, `testdata/benchmark-project/verify.sh`, `testdata/benchmark-project/task.md`
- [ ] Message: `feat(benchmark): add setup, verify scripts and task document`

---

## Spec Coverage

| Spec ID | Description | Phase |
|---------|-------------|-------|
| BP-1 | Project structure (Go + Python + TS + schemas) | 1-4 |
| BP-2 | ~21 files total | 1-5 |
| BP-3 | task.md defines feature request | 5 |
| BP-4 | setup.sh creates clean environment | 5 |
| BP-5 | rpi index build produces multi-language index | 5 |
| BP-6 | No network for setup (vendored deps) | 4 (package.json) |
| BP-7 | 13 baseline tests pass | 1, 3, 4 |
| BP-8 | Baseline coverage areas | 1, 3, 4 |
| BP-9 | Conditional transform feature request | 5 |
| BP-10 | Implementation requires all 3 languages + schema | 1, 2, 3, 4 |
| BP-11 | 10 new failing tests | 1, 3, 4 |
| BP-12 | Go conditional tests | 1 |
| BP-13 | Python conditional tests | 3 |
| BP-14 | TypeScript conditional tests | 4 |
| BP-15 | verify.sh reports pass/fail | 5 |
| BP-16 | Partial scoring works | 5 |
| BP-17 | 8+ exported, 4+ unexported Go symbols | 1 |
| BP-18 | 5+ Python classes, 10+ functions | 3 |
| BP-19 | Cross-language schema references | 2, 3, 4 |
| BP-20 | 3+ distinct packages in index | 1, 3, 4 |

## References
- Design: .rpi/designs/2026-04-03-benchmark-project-for-rpi-flow-quality-testing.md
- Spec: .rpi/specs/benchmark-project-for-rpi-flow-quality-testing.md
