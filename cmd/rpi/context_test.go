package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectCurrentPhase(t *testing.T) {
	content := `## Phase 1: Setup

### Tasks:
- [x] Create directory
- [x] Write config

---

## Phase 2: Core Logic

### Tasks:
- [x] Add parser
- [ ] Add validator
- [ ] Add serializer

---

## Phase 3: Tests

### Tasks:
- [ ] Unit tests
- [ ] Integration tests
`
	phase, items := detectCurrentPhase(content)
	if phase != "Phase 2: Core Logic" {
		t.Errorf("phase = %q, want %q", phase, "Phase 2: Core Logic")
	}
	if len(items) != 2 {
		t.Fatalf("items count = %d, want 2", len(items))
	}
	if items[0] != "Add validator" {
		t.Errorf("items[0] = %q, want %q", items[0], "Add validator")
	}
	if items[1] != "Add serializer" {
		t.Errorf("items[1] = %q, want %q", items[1], "Add serializer")
	}
}

func TestDetectCurrentPhaseAllComplete(t *testing.T) {
	content := `## Phase 1: Setup

- [x] Create directory
- [x] Write config

## Phase 2: Done

- [x] All done
`
	phase, items := detectCurrentPhase(content)
	if phase != "" {
		t.Errorf("phase = %q, want empty", phase)
	}
	if len(items) != 0 {
		t.Errorf("items count = %d, want 0", len(items))
	}
}

func TestDetectCurrentPhaseMaxItems(t *testing.T) {
	content := `## Phase 1: Big Phase

- [ ] Item one
- [ ] Item two
- [ ] Item three
- [ ] Item four
- [ ] Item five
`
	_, items := detectCurrentPhase(content)
	if len(items) != 3 {
		t.Fatalf("items count = %d, want 3 (max)", len(items))
	}
	if items[2] != "Item three" {
		t.Errorf("items[2] = %q, want %q", items[2], "Item three")
	}
}

func TestAssembleContextNoPlan(t *testing.T) {
	dir := t.TempDir()
	plansDir := filepath.Join(dir, "plans")
	os.MkdirAll(plansDir, 0755)

	result, err := assembleContext(dir, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Plan != nil {
		t.Error("expected nil Plan when no active plans exist")
	}
}

func TestAssembleContextExplicitPath(t *testing.T) {
	dir := t.TempDir()
	plansDir := filepath.Join(dir, "plans")
	os.MkdirAll(plansDir, 0755)

	planContent := `---
topic: "test plan"
status: active
---

# Test Plan

## Phase 1: Setup

- [x] Done item

## Phase 2: Work

- [ ] Pending item
`
	planPath := filepath.Join(plansDir, "2026-01-01-test.md")
	os.WriteFile(planPath, []byte(planContent), 0644)

	result, err := assembleContext(dir, planPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Plan == nil {
		t.Fatal("expected non-nil Plan")
	}
	if result.Plan.Path != planPath {
		t.Errorf("path = %q, want %q", result.Plan.Path, planPath)
	}
	if result.Plan.Topic != "test plan" {
		t.Errorf("topic = %q, want %q", result.Plan.Topic, "test plan")
	}
	if result.Plan.CurrentPhase != "Phase 2: Work" {
		t.Errorf("current_phase = %q, want %q", result.Plan.CurrentPhase, "Phase 2: Work")
	}
	if result.Plan.Progress.Checked != 1 || result.Plan.Progress.Total != 2 {
		t.Errorf("progress = %d/%d, want 1/2", result.Plan.Progress.Checked, result.Plan.Progress.Total)
	}
}

func TestAssembleContextAutoDetect(t *testing.T) {
	dir := t.TempDir()
	plansDir := filepath.Join(dir, "plans")
	specsDir := filepath.Join(dir, "specs")
	designsDir := filepath.Join(dir, "designs")
	os.MkdirAll(plansDir, 0755)
	os.MkdirAll(specsDir, 0755)
	os.MkdirAll(designsDir, 0755)

	specContent := `---
feature: test-feature
---

# Test Feature

## Scenarios

### First scenario
Given something
When action
Then result

### Second scenario
Given other
When another action
Then another result
`
	specPath := filepath.Join(specsDir, "test-feature.md")
	os.WriteFile(specPath, []byte(specContent), 0644)

	designContent := `---
topic: "test design"
status: complete
spec: ` + specPath + `
---

# Test Design

## Constraints

- Must be fast
- Must be portable
`
	designPath := filepath.Join(designsDir, "2026-01-01-test.md")
	os.WriteFile(designPath, []byte(designContent), 0644)

	planContent := `---
topic: "auto detect plan"
status: active
date: 2026-01-02T00:00:00Z
design: ` + designPath + `
spec: ` + specPath + `
---

# Auto Detect Plan

## Phase 1: Done

- [x] Completed

## Phase 2: Current

- [ ] Next thing
`
	planPath := filepath.Join(plansDir, "2026-01-02-auto.md")
	os.WriteFile(planPath, []byte(planContent), 0644)

	result, err := assembleContext(dir, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Plan == nil {
		t.Fatal("expected non-nil Plan from auto-detect")
	}
	if result.Plan.Path != planPath {
		t.Errorf("path = %q, want %q", result.Plan.Path, planPath)
	}
	if result.Plan.CurrentPhase != "Phase 2: Current" {
		t.Errorf("current_phase = %q, want %q", result.Plan.CurrentPhase, "Phase 2: Current")
	}

	// Spec
	if result.Spec == nil {
		t.Fatal("expected non-nil Spec")
	}
	if result.Spec.Feature != "test-feature" {
		t.Errorf("feature = %q, want %q", result.Spec.Feature, "test-feature")
	}
	if len(result.Spec.ScenarioTitles) != 2 {
		t.Fatalf("scenario_titles count = %d, want 2", len(result.Spec.ScenarioTitles))
	}
	if result.Spec.ScenarioTitles[0] != "First scenario" {
		t.Errorf("scenario_titles[0] = %q, want %q", result.Spec.ScenarioTitles[0], "First scenario")
	}

	// Constraints
	if result.Constraints == "" {
		t.Error("expected non-empty Constraints")
	}
}

func TestAssembleContextConstraintsTruncation(t *testing.T) {
	dir := t.TempDir()
	plansDir := filepath.Join(dir, "plans")
	designsDir := filepath.Join(dir, "designs")
	os.MkdirAll(plansDir, 0755)
	os.MkdirAll(designsDir, 0755)

	// Build a long constraints section (>200 chars)
	longConstraint := "- This is a very long constraint that goes on and on and on to exceed the two hundred character limit that we have set for the constraints field in the context output. It keeps going and going and going and then some more text."
	designContent := `---
topic: "long design"
status: complete
---

# Long Design

## Constraints

` + longConstraint + `
`
	designPath := filepath.Join(designsDir, "2026-01-01-long.md")
	os.WriteFile(designPath, []byte(designContent), 0644)

	planContent := `---
topic: "truncation test"
status: active
date: 2026-01-01T00:00:00Z
design: ` + designPath + `
---

# Truncation Test

## Phase 1: Work

- [ ] Do something
`
	planPath := filepath.Join(plansDir, "2026-01-01-trunc.md")
	os.WriteFile(planPath, []byte(planContent), 0644)

	result, err := assembleContext(dir, planPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Constraints) > 203 { // 200 + "..."
		t.Errorf("constraints length = %d, want <= 203", len(result.Constraints))
	}
	if result.Constraints[len(result.Constraints)-3:] != "..." {
		t.Errorf("constraints should end with '...', got %q", result.Constraints[len(result.Constraints)-3:])
	}
}
