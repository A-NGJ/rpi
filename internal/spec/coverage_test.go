package spec

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, dir, relPath, content string) string {
	t.Helper()
	full := filepath.Join(dir, relPath)
	if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return full
}

func TestParseBehaviors(t *testing.T) {
	dir := t.TempDir()
	specPath := writeFile(t, dir, "spec.md", `---
domain: test-module
id: TM
status: approved
---

# test-module

## Behavior
### Parsing
- **TM-1**: Parses input correctly
- **TM-2**: Handles empty input

### Validation
- **TM-3**: Rejects invalid data
`)

	behaviors, prefix, err := ParseBehaviors(specPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if prefix != "TM" {
		t.Errorf("prefix = %q, want %q", prefix, "TM")
	}

	if len(behaviors) != 3 {
		t.Fatalf("got %d behaviors, want 3", len(behaviors))
	}

	if behaviors[0].ID != "TM-1" {
		t.Errorf("behaviors[0].ID = %q, want %q", behaviors[0].ID, "TM-1")
	}
	if behaviors[0].Description != "Parses input correctly" {
		t.Errorf("behaviors[0].Description = %q, want %q", behaviors[0].Description, "Parses input correctly")
	}
	if behaviors[2].ID != "TM-3" {
		t.Errorf("behaviors[2].ID = %q, want %q", behaviors[2].ID, "TM-3")
	}
}

func TestParseBehaviorsNoID(t *testing.T) {
	dir := t.TempDir()
	specPath := writeFile(t, dir, "spec.md", `---
domain: test-module
status: draft
---

# test-module

## Behavior
- **TM-1**: Some behavior
`)

	_, _, err := ParseBehaviors(specPath)
	if err == nil {
		t.Fatal("expected error for spec without id field")
	}
}

func TestScanTestFiles(t *testing.T) {
	dir := t.TempDir()

	writeFile(t, dir, "pkg/thing_test.go", `package pkg

// spec:TM-1
func TestParsing(t *testing.T) {}

// spec:TM-2 spec:TM-3
func TestMultiple(t *testing.T) {}
`)

	writeFile(t, dir, "pkg/other_test.go", `package pkg

// spec:TM-1
func TestParsingAgain(t *testing.T) {}
`)

	// Non-test file should be ignored
	writeFile(t, dir, "pkg/thing.go", `package pkg
// spec:TM-1
func Parse() {}
`)

	refs, err := ScanTestFiles(dir, "TM")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(refs) != 4 {
		t.Fatalf("got %d refs, want 4", len(refs))
	}

	// Verify all refs have the TM prefix
	for _, r := range refs {
		if r.ID[:3] != "TM-" {
			t.Errorf("unexpected ref ID: %s", r.ID)
		}
	}
}

func TestComputeCoverage(t *testing.T) {
	behaviors := []Behavior{
		{ID: "TM-1", Description: "First behavior"},
		{ID: "TM-2", Description: "Second behavior"},
		{ID: "TM-3", Description: "Third behavior"},
	}

	refs := []TestRef{
		{ID: "TM-1", File: "pkg/thing_test.go", Line: 5},
		{ID: "TM-3", File: "pkg/thing_test.go", Line: 15},
	}

	report := ComputeCoverage(behaviors, refs, "test-module", "TM")

	if report.Total != 3 {
		t.Errorf("Total = %d, want 3", report.Total)
	}
	if report.Covered != 2 {
		t.Errorf("Covered = %d, want 2", report.Covered)
	}
	if report.Missing != 1 {
		t.Errorf("Missing = %d, want 1", report.Missing)
	}

	if len(report.MissingBehaviors) != 1 || report.MissingBehaviors[0].ID != "TM-2" {
		t.Errorf("expected TM-2 in missing, got %v", report.MissingBehaviors)
	}
}

func TestComputeCoverageAllCovered(t *testing.T) {
	behaviors := []Behavior{
		{ID: "TM-1", Description: "First"},
		{ID: "TM-2", Description: "Second"},
	}

	refs := []TestRef{
		{ID: "TM-1", File: "a_test.go", Line: 1},
		{ID: "TM-2", File: "a_test.go", Line: 5},
	}

	report := ComputeCoverage(behaviors, refs, "test-module", "TM")

	if report.Missing != 0 {
		t.Errorf("Missing = %d, want 0", report.Missing)
	}
	if len(report.MissingBehaviors) != 0 {
		t.Errorf("expected empty MissingBehaviors, got %v", report.MissingBehaviors)
	}
}

func TestComputeCoverageNoneFound(t *testing.T) {
	behaviors := []Behavior{
		{ID: "TM-1", Description: "First"},
		{ID: "TM-2", Description: "Second"},
	}

	var refs []TestRef

	report := ComputeCoverage(behaviors, refs, "test-module", "TM")

	if report.Covered != 0 {
		t.Errorf("Covered = %d, want 0", report.Covered)
	}
	if report.Missing != 2 {
		t.Errorf("Missing = %d, want 2", report.Missing)
	}
}
