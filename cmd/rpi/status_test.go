package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func writeStatusFile(t *testing.T, dir, relPath, content string) {
	t.Helper()
	full := filepath.Join(dir, relPath)
	if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func runStatusCmd(t *testing.T, dir string) string {
	t.Helper()
	oldFlag := rpiDirFlag
	rpiDirFlag = dir
	defer func() { rpiDirFlag = oldFlag }()

	buf := new(bytes.Buffer)
	statusCmd.SetOut(buf)
	if err := statusCmd.RunE(statusCmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return buf.String()
}

func TestStatusArtifactSummary(t *testing.T) {
	dir := t.TempDir()
	writeStatusFile(t, dir, "specs/s1.md", "---\nstatus: active\ntopic: S1\n---\n")
	writeStatusFile(t, dir, "specs/s2.md", "---\nstatus: active\ntopic: S2\n---\n")
	writeStatusFile(t, dir, "plans/p1.md", "---\nstatus: draft\ntopic: P1\n---\n")
	writeStatusFile(t, dir, "designs/d1.md", "---\nstatus: complete\ntopic: D1\n---\n")

	output := runStatusCmd(t, dir)

	if strings.Contains(output, "specs:") {
		t.Errorf("specs should not appear in Artifacts section, got:\n%s", output)
	}
	if !strings.Contains(output, "plans:") || !strings.Contains(output, "1 draft") {
		t.Errorf("expected plans with 1 draft, got:\n%s", output)
	}
	if !strings.Contains(output, "designs:") || !strings.Contains(output, "1 complete") {
		t.Errorf("expected designs with 1 complete, got:\n%s", output)
	}
}

func TestStatusOmitsEmptyTypes(t *testing.T) {
	dir := t.TempDir()
	writeStatusFile(t, dir, "specs/s1.md", "---\nstatus: active\ntopic: S1\n---\n")
	writeStatusFile(t, dir, "plans/p1.md", "---\nstatus: draft\ntopic: P1\n---\n")

	output := runStatusCmd(t, dir)

	for _, absent := range []string{"designs:", "research:", "reviews:"} {
		if strings.Contains(output, absent) {
			t.Errorf("should not contain %q, got:\n%s", absent, output)
		}
	}
}

func TestStatusExcludesArchive(t *testing.T) {
	dir := t.TempDir()
	writeStatusFile(t, dir, "specs/s1.md", "---\nstatus: active\ntopic: S1\n---\n")
	writeStatusFile(t, dir, "archive/2026-03/specs/s2.md", "---\nstatus: archived\ntopic: S2\n---\n")

	output := runStatusCmd(t, dir)

	if strings.Contains(output, "archived") {
		t.Errorf("should not contain archived artifacts, got:\n%s", output)
	}
	if !strings.Contains(output, "Specifications") {
		t.Errorf("expected Specifications section for non-archived spec, got:\n%s", output)
	}
}

func TestStatusStaleDetection(t *testing.T) {
	dir := t.TempDir()
	// Plan dated 2026-03-01 — 34 days before our fixed "now" of 2026-04-04
	writeStatusFile(t, dir, "plans/p1.md", "---\nstatus: draft\ntopic: P1\ndate: 2026-03-01\n---\n")

	oldNow := nowFunc
	nowFunc = func() time.Time { return time.Date(2026, 4, 4, 0, 0, 0, 0, time.UTC) }
	defer func() { nowFunc = oldNow }()

	output := runStatusCmd(t, dir)

	if !strings.Contains(output, "Stale") {
		t.Errorf("expected Stale section, got:\n%s", output)
	}
	if !strings.Contains(output, "34d ago") {
		t.Errorf("expected 34d ago, got:\n%s", output)
	}
}

func TestStatusStaleCustomThreshold(t *testing.T) {
	dir := t.TempDir()
	// Plan dated 2026-03-25 — 10 days before now
	writeStatusFile(t, dir, "plans/p1.md", "---\nstatus: draft\ntopic: P1\ndate: 2026-03-25\n---\n")

	oldNow := nowFunc
	nowFunc = func() time.Time { return time.Date(2026, 4, 4, 0, 0, 0, 0, time.UTC) }
	defer func() { nowFunc = oldNow }()

	// With threshold 7: should be stale
	oldStaleDays := staleDays
	staleDays = 7
	output := runStatusCmd(t, dir)
	if !strings.Contains(output, "Stale") {
		t.Errorf("expected Stale section with --stale-days 7, got:\n%s", output)
	}

	// With threshold 14: should not be stale
	staleDays = 14
	output = runStatusCmd(t, dir)
	if strings.Contains(output, "Stale") {
		t.Errorf("should not have Stale section with --stale-days 14, got:\n%s", output)
	}

	staleDays = oldStaleDays
}

func TestStatusStaleMissingDate(t *testing.T) {
	dir := t.TempDir()
	// Active spec with no last_updated field
	writeStatusFile(t, dir, "specs/s1.md", "---\nstatus: active\ntopic: S1\n---\n")

	oldNow := nowFunc
	nowFunc = func() time.Time { return time.Date(2026, 4, 4, 0, 0, 0, 0, time.UTC) }
	defer func() { nowFunc = oldNow }()

	output := runStatusCmd(t, dir)

	// Should appear in Specifications section
	if !strings.Contains(output, "Specifications") {
		t.Errorf("expected Specifications section, got:\n%s", output)
	}
	// Should NOT appear in Stale section
	if strings.Contains(output, "Stale") {
		t.Errorf("should not have Stale section for missing date, got:\n%s", output)
	}
}

func chdirTemp(t *testing.T, dir string) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(orig) })
}

func TestStatusActivePlanChain(t *testing.T) {
	dir := t.TempDir()
	chdirTemp(t, dir)

	// Active plan links to design and spec
	writeStatusFile(t, dir, "plans/p1.md",
		"---\nstatus: active\ntopic: My Plan\ndesign: designs/d1.md\nspec: specs/s1.md\n---\n# Plan\n- [x] done\n- [ ] todo\n")
	// Design links to research (should NOT appear — one level only)
	writeStatusFile(t, dir, "designs/d1.md",
		"---\nstatus: complete\ntopic: My Design\nresearch: research/r1.md\n---\n")
	writeStatusFile(t, dir, "specs/s1.md",
		"---\nstatus: active\ntopic: My Spec\n---\n")
	writeStatusFile(t, dir, "research/r1.md",
		"---\nstatus: complete\ntopic: My Research\n---\n")

	output := runStatusCmd(t, dir)

	if !strings.Contains(output, "Active Plans") {
		t.Fatalf("expected Active Plans section, got:\n%s", output)
	}
	if !strings.Contains(output, "My Plan") {
		t.Errorf("expected plan topic, got:\n%s", output)
	}
	// Should NOT show linked artifact sub-rows under Active Plans (ST-10 revised)
	activePlansIdx := strings.Index(output, "Active Plans")
	specsIdx := strings.Index(output, "Specifications")
	if activePlansIdx >= 0 {
		end := len(output)
		if specsIdx > activePlansIdx {
			end = specsIdx
		}
		activePlansSection := output[activePlansIdx:end]
		if strings.Contains(activePlansSection, "design:") {
			t.Errorf("should not show design link under Active Plans, got:\n%s", activePlansSection)
		}
		if strings.Contains(activePlansSection, "spec:") {
			t.Errorf("should not show spec link under Active Plans, got:\n%s", activePlansSection)
		}
	}
	// Specifications section should list the spec
	if !strings.Contains(output, "Specifications") {
		t.Errorf("expected Specifications section, got:\n%s", output)
	}
	if !strings.Contains(output, "My Spec") {
		t.Errorf("expected active spec name in Active Specs section, got:\n%s", output)
	}
}

func TestStatusCheckboxProgress(t *testing.T) {
	dir := t.TempDir()
	chdirTemp(t, dir)

	// Plan with 3 checked, 7 unchecked = 3/10 (30%)
	checkboxes := "- [x] a\n- [x] b\n- [x] c\n- [ ] d\n- [ ] e\n- [ ] f\n- [ ] g\n- [ ] h\n- [ ] i\n- [ ] j\n"
	writeStatusFile(t, dir, "plans/p1.md",
		"---\nstatus: active\ntopic: Progress Plan\n---\n"+checkboxes)

	output := runStatusCmd(t, dir)

	if !strings.Contains(output, "3/10 (30%)") {
		t.Errorf("expected 3/10 (30%%), got:\n%s", output)
	}
}

func TestStatusNoCheckboxes(t *testing.T) {
	dir := t.TempDir()
	chdirTemp(t, dir)

	writeStatusFile(t, dir, "plans/p1.md",
		"---\nstatus: draft\ntopic: Empty Plan\n---\n# No checkboxes here\n")

	output := runStatusCmd(t, dir)

	if !strings.Contains(output, "Active Plans") {
		t.Fatalf("expected Active Plans section, got:\n%s", output)
	}
	if !strings.Contains(output, "Empty Plan") {
		t.Errorf("expected plan topic, got:\n%s", output)
	}
	// Should not have any progress indicator
	if strings.Contains(output, "/") && strings.Contains(output, "%") {
		t.Errorf("should not show progress for plan with no checkboxes, got:\n%s", output)
	}
}

func TestStatusArchiveReadiness(t *testing.T) {
	dir := t.TempDir()

	// Complete design with 0 references — should appear
	writeStatusFile(t, dir, "designs/d1.md",
		"---\nstatus: complete\ntopic: Unreferenced Design\n---\n")
	// Complete design referenced by a plan — should NOT appear
	writeStatusFile(t, dir, "designs/d2.md",
		"---\nstatus: complete\ntopic: Referenced Design\n---\n")
	writeStatusFile(t, dir, "plans/p1.md",
		"---\nstatus: active\ntopic: Plan\ndesign: designs/d2.md\n---\n")

	output := runStatusCmd(t, dir)

	if !strings.Contains(output, "Ready to Archive") {
		t.Fatalf("expected Ready to Archive section, got:\n%s", output)
	}
	if !strings.Contains(output, "1 designs") {
		t.Errorf("expected 1 designs in archive summary, got:\n%s", output)
	}
}

func runStatusCmdWithFormat(t *testing.T, dir, format string) string {
	t.Helper()
	oldFlag := rpiDirFlag
	oldFormat := formatFlag
	rpiDirFlag = dir
	formatFlag = format
	defer func() {
		rpiDirFlag = oldFlag
		formatFlag = oldFormat
	}()

	buf := new(bytes.Buffer)
	statusCmd.SetOut(buf)
	if err := statusCmd.RunE(statusCmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return buf.String()
}

func TestStatusJSONOutput(t *testing.T) {
	dir := t.TempDir()
	writeStatusFile(t, dir, "specs/s1.md", "---\nstatus: active\ntopic: S1\n---\n")
	writeStatusFile(t, dir, "plans/p1.md", "---\nstatus: draft\ntopic: P1\n---\n")

	output := runStatusCmdWithFormat(t, dir, "json")

	var result statusOutput
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, output)
	}

	// Check all four keys are present
	if result.Summary == nil {
		t.Error("expected summary key in JSON")
	}
	if result.Stale == nil {
		t.Error("expected stale key in JSON")
	}
	if result.ActivePlans == nil {
		t.Error("expected active_plans key in JSON")
	}
	if result.Archivable == nil {
		t.Error("expected archivable key in JSON")
	}

	// Specs should not appear in summary (specs are statusless)
	if _, ok := result.Summary["spec"]; ok {
		t.Errorf("specs should not appear in summary, got %v", result.Summary)
	}
}

func TestStatusSpecificationsSection(t *testing.T) {
	dir := t.TempDir()
	writeStatusFile(t, dir, "specs/s1.md", "---\ntopic: Auth Permissions\n---\n## Scenarios\n\n### First scenario\nGiven x\nWhen y\nThen z\n\n### Second scenario\nGiven a\nWhen b\nThen c\n")
	writeStatusFile(t, dir, "specs/s2.md", "---\ntopic: Rate Limiting\n---\n## Scenarios\n\n### Only scenario\nGiven x\nWhen y\nThen z\n")
	writeStatusFile(t, dir, "specs/s3.md", "---\ntopic: Old Spec\n---\n")

	output := runStatusCmd(t, dir)

	if !strings.Contains(output, "Specifications") {
		t.Fatalf("expected Specifications section, got:\n%s", output)
	}
	if !strings.Contains(output, "Auth Permissions") {
		t.Errorf("expected spec name 'Auth Permissions', got:\n%s", output)
	}
	if !strings.Contains(output, "Rate Limiting") {
		t.Errorf("expected spec name 'Rate Limiting', got:\n%s", output)
	}
	if !strings.Contains(output, "Old Spec") {
		t.Errorf("expected spec name 'Old Spec', got:\n%s", output)
	}
}

func TestStatusSpecsScenarioCount(t *testing.T) {
	dir := t.TempDir()
	// Spec with 3 scenarios
	writeStatusFile(t, dir, "specs/s1.md", "---\ntopic: Auth Permissions\n---\n## Scenarios\n\n"+
		"### First scenario\nGiven x\nWhen y\nThen z\n\n### Second scenario\nGiven a\nWhen b\nThen c\n\n### Third scenario\nGiven d\nWhen e\nThen f\n")
	// Spec with 0 scenarios
	writeStatusFile(t, dir, "specs/s2.md", "---\ntopic: Empty Spec\n---\nNo scenarios here.\n")

	output := runStatusCmd(t, dir)

	if !strings.Contains(output, "Auth Permissions") || !strings.Contains(output, "3 scenarios") {
		t.Errorf("expected 'Auth Permissions' with '3 scenarios', got:\n%s", output)
	}
	if !strings.Contains(output, "Empty Spec") || !strings.Contains(output, "0 scenarios") {
		t.Errorf("expected 'Empty Spec' with '0 scenarios', got:\n%s", output)
	}
}

func TestStatusSpecsScenarioCountJSON(t *testing.T) {
	dir := t.TempDir()
	writeStatusFile(t, dir, "specs/s1.md", "---\ntopic: Auth Permissions\n---\n## Scenarios\n\n"+
		"### First scenario\nGiven x\nWhen y\nThen z\n\n### Second scenario\nGiven a\nWhen b\nThen c\n")

	output := runStatusCmdWithFormat(t, dir, "json")

	var result statusOutput
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, output)
	}
	if len(result.Specs) != 1 {
		t.Fatalf("expected 1 spec, got %d", len(result.Specs))
	}
	if result.Specs[0].Name != "Auth Permissions" {
		t.Errorf("expected name 'Auth Permissions', got %q", result.Specs[0].Name)
	}
	if result.Specs[0].Scenarios != 2 {
		t.Errorf("expected 2 scenarios, got %d", result.Specs[0].Scenarios)
	}
}

func TestStatusNoActiveSectionWhenEmpty(t *testing.T) {
	dir := t.TempDir()
	// Only draft and complete designs — no active ones, no specs
	writeStatusFile(t, dir, "designs/d1.md", "---\nstatus: draft\ntopic: Draft Design\n---\n")
	writeStatusFile(t, dir, "designs/d2.md", "---\nstatus: complete\ntopic: Done Design\n---\n")

	output := runStatusCmd(t, dir)

	if strings.Contains(output, "Active Designs") {
		t.Errorf("should not show Active Designs when none active, got:\n%s", output)
	}
	if strings.Contains(output, "Specifications") {
		t.Errorf("should not show Specifications when no specs exist, got:\n%s", output)
	}
	if strings.Contains(output, "Active Plans") {
		t.Errorf("should not show Active Plans when none active, got:\n%s", output)
	}
}

func TestStatusExitCodeOnError(t *testing.T) {
	oldFlag := rpiDirFlag
	rpiDirFlag = "/nonexistent/path/that/does/not/exist"
	defer func() { rpiDirFlag = oldFlag }()

	buf := new(bytes.Buffer)
	statusCmd.SetOut(buf)
	err := statusCmd.RunE(statusCmd, nil)
	if err == nil {
		t.Error("expected error for nonexistent rpi-dir")
	}
}
