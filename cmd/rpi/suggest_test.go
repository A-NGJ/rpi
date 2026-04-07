package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/A-NGJ/rpi/internal/scanner"
)

func writeTempArtifact(t *testing.T, dir, subdir, filename, content string) string {
	t.Helper()
	d := filepath.Join(dir, subdir)
	os.MkdirAll(d, 0755)
	p := filepath.Join(d, filename)
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", p, err)
	}
	return p
}

func TestSuggestNextActivePlanUnchecked(t *testing.T) {
	dir := t.TempDir()
	writeTempArtifact(t, dir, "plans", "2026-01-01-test.md", `---
topic: "test plan"
status: active
date: 2026-01-01T00:00:00Z
---

## Phase 1: Work

- [x] Done item
- [ ] Pending item
- [ ] Another pending
`)

	s, err := suggestNext(dir, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(s.Action, "/rpi-implement") {
		t.Errorf("action = %q, want to contain /rpi-implement", s.Action)
	}
	if !strings.Contains(s.Reasoning, "2 unchecked") {
		t.Errorf("reasoning = %q, want to mention 2 unchecked items", s.Reasoning)
	}
	if s.Artifact == "" {
		t.Error("expected non-empty artifact path")
	}
}

func TestSuggestNextActivePlanAllChecked(t *testing.T) {
	dir := t.TempDir()

	specPath := writeTempArtifact(t, dir, "specs", "test-feature.md", `---
feature: test-feature
---

# Test Feature

## Scenarios

### First scenario
Given something
When action
Then result
`)

	planPath := writeTempArtifact(t, dir, "plans", "2026-01-01-test.md", `---
topic: "completed plan"
status: active
date: 2026-01-01T00:00:00Z
spec: `+specPath+`
---

## Phase 1: Done

- [x] All done
- [x] Also done
`)

	s, err := suggestNext(dir, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(s.Action, "/rpi-verify") {
		t.Errorf("action = %q, want to contain /rpi-verify", s.Action)
	}
	if !strings.Contains(s.Action, specPath) {
		t.Errorf("action = %q, want to contain spec path %s", s.Action, specPath)
	}
	if s.Artifact != planPath {
		t.Errorf("artifact = %q, want %q", s.Artifact, planPath)
	}
}

func TestSuggestNextActiveDesignNoPlan(t *testing.T) {
	dir := t.TempDir()
	designPath := writeTempArtifact(t, dir, "designs", "2026-01-01-test.md", `---
topic: "new design"
status: active
date: 2026-01-01T00:00:00Z
---

# New Design
`)

	s, err := suggestNext(dir, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(s.Action, "/rpi-plan") {
		t.Errorf("action = %q, want to contain /rpi-plan", s.Action)
	}
	if s.Artifact != designPath {
		t.Errorf("artifact = %q, want %q", s.Artifact, designPath)
	}
}

func TestSuggestNextDraftDesign(t *testing.T) {
	dir := t.TempDir()
	designPath := writeTempArtifact(t, dir, "designs", "2026-01-01-test.md", `---
topic: "draft design"
status: draft
date: 2026-01-01T00:00:00Z
---

# Draft Design
`)

	s, err := suggestNext(dir, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(s.Action, "Review and approve") {
		t.Errorf("action = %q, want to contain 'Review and approve'", s.Action)
	}
	if s.Artifact != designPath {
		t.Errorf("artifact = %q, want %q", s.Artifact, designPath)
	}
}

func TestSuggestNextCompleteResearchNoDesign(t *testing.T) {
	dir := t.TempDir()
	researchPath := writeTempArtifact(t, dir, "research", "2026-01-01-test.md", `---
topic: "finished research"
status: complete
date: 2026-01-01T00:00:00Z
---

# Finished Research
`)

	s, err := suggestNext(dir, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(s.Action, "/rpi-propose") {
		t.Errorf("action = %q, want to contain /rpi-propose", s.Action)
	}
	if s.Artifact != researchPath {
		t.Errorf("artifact = %q, want %q", s.Artifact, researchPath)
	}
}

func TestSuggestNextDraftResearch(t *testing.T) {
	dir := t.TempDir()
	researchPath := writeTempArtifact(t, dir, "research", "2026-01-01-test.md", `---
topic: "draft research"
status: draft
date: 2026-01-01T00:00:00Z
---

# Draft Research
`)

	s, err := suggestNext(dir, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(s.Action, "Review and finalize") {
		t.Errorf("action = %q, want to contain 'Review and finalize'", s.Action)
	}
	if s.Artifact != researchPath {
		t.Errorf("artifact = %q, want %q", s.Artifact, researchPath)
	}
}

func TestSuggestNextEmpty(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "plans"), 0755)

	s, err := suggestNext(dir, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(s.Action, "/rpi-propose") {
		t.Errorf("action = %q, want to contain /rpi-propose", s.Action)
	}
	if s.Artifact != "" {
		t.Errorf("artifact = %q, want empty", s.Artifact)
	}
}

func TestSuggestNextPriorityOrder(t *testing.T) {
	dir := t.TempDir()

	// Active plan (priority 1) AND complete research (priority 5) — plan should win
	writeTempArtifact(t, dir, "plans", "2026-01-01-plan.md", `---
topic: "active plan"
status: active
date: 2026-01-01T00:00:00Z
---

## Phase 1: Work

- [ ] Pending
`)
	writeTempArtifact(t, dir, "research", "2026-01-01-research.md", `---
topic: "complete research"
status: complete
date: 2026-01-01T00:00:00Z
---

# Research
`)

	s, err := suggestNext(dir, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(s.Action, "/rpi-implement") {
		t.Errorf("action = %q, want /rpi-implement (plan should take priority over research)", s.Action)
	}
}

func TestSuggestNextMostRecentWins(t *testing.T) {
	dir := t.TempDir()

	writeTempArtifact(t, dir, "plans", "2026-01-01-old.md", `---
topic: "old plan"
status: active
date: 2026-01-01T00:00:00Z
---

## Phase 1: Work

- [ ] Old pending
`)
	recentPath := writeTempArtifact(t, dir, "plans", "2026-06-01-recent.md", `---
topic: "recent plan"
status: active
date: 2026-06-01T00:00:00Z
---

## Phase 1: Work

- [ ] Recent pending
`)

	s, err := suggestNext(dir, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Artifact != recentPath {
		t.Errorf("artifact = %q, want %q (most recent should win)", s.Artifact, recentPath)
	}
	if !strings.Contains(s.Reasoning, "recent plan") {
		t.Errorf("reasoning = %q, want to mention 'recent plan'", s.Reasoning)
	}
}

func TestBuildDownstreamMaps(t *testing.T) {
	dir := t.TempDir()

	researchPath := writeTempArtifact(t, dir, "research", "2026-01-01-test.md", `---
topic: "test research"
status: complete
---

# Research
`)

	designPath := writeTempArtifact(t, dir, "designs", "2026-01-01-test.md", `---
topic: "test design"
status: active
related_research: `+researchPath+`
---

# Design
`)

	writeTempArtifact(t, dir, "plans", "2026-01-01-test.md", `---
topic: "test plan"
status: active
design: `+designPath+`
---

# Plan
`)

	allArtifacts, err := scanner.Scan(dir, scanner.Filters{})
	if err != nil {
		t.Fatalf("scan: %v", err)
	}

	planDesigns, designResearch := buildDownstreamMaps(allArtifacts)

	if !planDesigns[designPath] {
		t.Errorf("planDesigns missing %s", designPath)
	}
	if !designResearch[researchPath] {
		t.Errorf("designResearch missing %s", researchPath)
	}
}
