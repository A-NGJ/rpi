package workflow

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// --- Agent Skills format compliance tests ---

func TestAllSkillsPresent(t *testing.T) {
	// AS-12: All 9 pipeline skills must be present
	expected := []string{
		"rpi-research", "rpi-propose", "rpi-plan", "rpi-implement",
		"rpi-verify", "rpi-diagnose", "rpi-explain", "rpi-commit", "rpi-archive",
	}

	entries, err := fs.ReadDir(assets, "assets/skills")
	if err != nil {
		t.Fatalf("read assets/skills: %v", err)
	}

	found := make(map[string]bool)
	for _, e := range entries {
		if e.IsDir() {
			found[e.Name()] = true
		}
	}

	for _, name := range expected {
		if !found[name] {
			t.Errorf("missing skill directory: %s", name)
		}
	}
	if len(found) != len(expected) {
		t.Errorf("expected %d skills, found %d", len(expected), len(found))
	}
}

func TestSkillNameMatchesDir(t *testing.T) {
	// AS-7: name field must match parent directory name
	// AS-8: name must match regex and ≤64 chars
	nameRegex := regexp.MustCompile(`^[a-z][a-z0-9]*(-[a-z0-9]+)*$`)

	entries, err := fs.ReadDir(assets, "assets/skills")
	if err != nil {
		t.Fatalf("read assets/skills: %v", err)
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dirName := e.Name()
		data, err := assets.ReadFile("assets/skills/" + dirName + "/SKILL.md")
		if err != nil {
			t.Errorf("read %s/SKILL.md: %v", dirName, err)
			continue
		}

		name := extractFrontmatterField(string(data), "name")
		if name != dirName {
			t.Errorf("skill %s: name field %q doesn't match directory", dirName, name)
		}
		if len(name) > 64 {
			t.Errorf("skill %s: name exceeds 64 chars (%d)", dirName, len(name))
		}
		if !nameRegex.MatchString(name) {
			t.Errorf("skill %s: name %q doesn't match naming regex", dirName, name)
		}
	}
}

func TestCanonicalSkillsHaveNoToolFields(t *testing.T) {
	// AS-10: canonical SKILL.md must not contain model, disable-model-invocation, or tools
	toolFields := []string{"model:", "disable-model-invocation:", "tools:"}

	entries, err := fs.ReadDir(assets, "assets/skills")
	if err != nil {
		t.Fatalf("read assets/skills: %v", err)
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		data, err := assets.ReadFile("assets/skills/" + e.Name() + "/SKILL.md")
		if err != nil {
			t.Errorf("read %s/SKILL.md: %v", e.Name(), err)
			continue
		}
		fm := extractFrontmatter(string(data))
		for _, field := range toolFields {
			if strings.Contains(fm, field) {
				t.Errorf("skill %s: canonical SKILL.md contains tool-specific field %s", e.Name(), field)
			}
		}
	}
}

func TestSkillDescriptionValid(t *testing.T) {
	// AS-9: description must be 1-1024 chars
	entries, err := fs.ReadDir(assets, "assets/skills")
	if err != nil {
		t.Fatalf("read assets/skills: %v", err)
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		data, err := assets.ReadFile("assets/skills/" + e.Name() + "/SKILL.md")
		if err != nil {
			t.Errorf("read %s/SKILL.md: %v", e.Name(), err)
			continue
		}
		desc := extractFrontmatterField(string(data), "description")
		if len(desc) == 0 {
			t.Errorf("skill %s: description is empty", e.Name())
		}
		if len(desc) > 1024 {
			t.Errorf("skill %s: description exceeds 1024 chars (%d)", e.Name(), len(desc))
		}
	}
}

// --- InstallSkills tests ---

func TestInstallSkills_AgentsOnly(t *testing.T) {
	agentsDir := filepath.Join(t.TempDir(), ".agents", "skills")

	count, err := InstallSkills(agentsDir, "", TargetAgentsOnly, false)
	if err != nil {
		t.Fatalf("InstallSkills error: %v", err)
	}
	if count != 9 {
		t.Errorf("expected 9 files installed, got %d", count)
	}

	// Verify all 9 canonical skills exist
	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		t.Fatalf("read agents dir: %v", err)
	}
	if len(entries) != 9 {
		t.Errorf("expected 9 skill dirs, got %d", len(entries))
	}

	// Verify canonical files have no tool-specific fields
	for _, e := range entries {
		data, err := os.ReadFile(filepath.Join(agentsDir, e.Name(), "SKILL.md"))
		if err != nil {
			t.Errorf("read %s: %v", e.Name(), err)
			continue
		}
		fm := extractFrontmatter(string(data))
		if strings.Contains(fm, "model:") {
			t.Errorf("canonical %s contains model: field", e.Name())
		}
	}
}

func TestInstallSkills_Claude(t *testing.T) {
	dir := t.TempDir()
	agentsDir := filepath.Join(dir, ".agents", "skills")
	toolDir := filepath.Join(dir, ".claude")

	count, err := InstallSkills(agentsDir, toolDir, TargetClaude, false)
	if err != nil {
		t.Fatalf("InstallSkills error: %v", err)
	}
	// 9 canonical + 9 tool-specific = 18
	if count != 18 {
		t.Errorf("expected 18 files installed, got %d", count)
	}

	// Verify rpi-archive has model and disable-model-invocation in tool copy
	archiveData, err := os.ReadFile(filepath.Join(toolDir, "skills", "rpi-archive", "SKILL.md"))
	if err != nil {
		t.Fatalf("read tool rpi-archive: %v", err)
	}
	archiveFM := extractFrontmatter(string(archiveData))
	if !strings.Contains(archiveFM, "model: haiku") {
		t.Error("tool rpi-archive should have model: haiku")
	}
	if !strings.Contains(archiveFM, "disable-model-invocation: true") {
		t.Error("tool rpi-archive should have disable-model-invocation: true")
	}

	// Verify rpi-commit has model in tool copy
	commitData, err := os.ReadFile(filepath.Join(toolDir, "skills", "rpi-commit", "SKILL.md"))
	if err != nil {
		t.Fatalf("read tool rpi-commit: %v", err)
	}
	if !strings.Contains(extractFrontmatter(string(commitData)), "model: haiku") {
		t.Error("tool rpi-commit should have model: haiku")
	}

	// Verify rpi-research (no overrides) — tool copy matches canonical
	canonData, err := os.ReadFile(filepath.Join(agentsDir, "rpi-research", "SKILL.md"))
	if err != nil {
		t.Fatalf("read canonical rpi-research: %v", err)
	}
	toolData, err := os.ReadFile(filepath.Join(toolDir, "skills", "rpi-research", "SKILL.md"))
	if err != nil {
		t.Fatalf("read tool rpi-research: %v", err)
	}
	if string(canonData) != string(toolData) {
		t.Error("rpi-research without overrides: tool copy should match canonical")
	}
}

func TestInstallSkills_OpenCode(t *testing.T) {
	dir := t.TempDir()
	agentsDir := filepath.Join(dir, ".agents", "skills")
	toolDir := filepath.Join(dir, ".opencode")

	_, err := InstallSkills(agentsDir, toolDir, TargetOpenCode, false)
	if err != nil {
		t.Fatalf("InstallSkills error: %v", err)
	}

	// Verify OpenCode translates model alias to full ID
	archiveData, err := os.ReadFile(filepath.Join(toolDir, "skills", "rpi-archive", "SKILL.md"))
	if err != nil {
		t.Fatalf("read opencode rpi-archive: %v", err)
	}
	if !strings.Contains(string(archiveData), "model: anthropic/claude-haiku-4-5-20251001") {
		t.Error("OpenCode rpi-archive should have full model ID")
	}
}

func TestInstallSkills_ContentParity(t *testing.T) {
	// AS-11 / TC-5: body content must be identical between canonical and tool copies
	dir := t.TempDir()
	agentsDir := filepath.Join(dir, ".agents", "skills")
	toolDir := filepath.Join(dir, ".claude")

	_, err := InstallSkills(agentsDir, toolDir, TargetClaude, false)
	if err != nil {
		t.Fatalf("InstallSkills error: %v", err)
	}

	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		t.Fatalf("read agents dir: %v", err)
	}

	for _, e := range entries {
		canonData, err := os.ReadFile(filepath.Join(agentsDir, e.Name(), "SKILL.md"))
		if err != nil {
			t.Errorf("read canonical %s: %v", e.Name(), err)
			continue
		}
		toolData, err := os.ReadFile(filepath.Join(toolDir, "skills", e.Name(), "SKILL.md"))
		if err != nil {
			t.Errorf("read tool %s: %v", e.Name(), err)
			continue
		}

		canonBody := extractBody(string(canonData))
		toolBody := extractBody(string(toolData))
		if canonBody != toolBody {
			t.Errorf("skill %s: body content differs between canonical and tool copy", e.Name())
		}
	}
}

func TestInstallSkills_NoOverwriteWithoutForce(t *testing.T) {
	dir := t.TempDir()
	agentsDir := filepath.Join(dir, ".agents", "skills")

	// First install
	_, err := InstallSkills(agentsDir, "", TargetAgentsOnly, false)
	if err != nil {
		t.Fatalf("first install: %v", err)
	}

	// Modify a file
	modPath := filepath.Join(agentsDir, "rpi-research", "SKILL.md")
	if err := os.WriteFile(modPath, []byte("custom content"), 0644); err != nil {
		t.Fatalf("modify file: %v", err)
	}

	// Second install without force — should not overwrite
	_, err = InstallSkills(agentsDir, "", TargetAgentsOnly, false)
	if err != nil {
		t.Fatalf("second install: %v", err)
	}
	data, _ := os.ReadFile(modPath)
	if string(data) != "custom content" {
		t.Error("file should not be overwritten without force")
	}

	// Third install with force — should overwrite
	count, err := InstallSkills(agentsDir, "", TargetAgentsOnly, true)
	if err != nil {
		t.Fatalf("force install: %v", err)
	}
	if count != 9 {
		t.Errorf("force install: expected 9 files, got %d", count)
	}
	data, _ = os.ReadFile(modPath)
	if string(data) == "custom content" {
		t.Error("file should be overwritten with force")
	}
}

// --- injectFrontmatter tests ---

func TestInjectFrontmatter_AddsFields(t *testing.T) {
	input := "---\nname: test\ndescription: a test\n---\n\n# Body\n"
	fields := map[string]string{"model": "haiku"}
	got := string(injectFrontmatter([]byte(input), fields, TargetClaude))

	if !strings.Contains(got, "model: haiku") {
		t.Error("should inject model: haiku")
	}
	if !strings.Contains(got, "name: test") {
		t.Error("should preserve existing fields")
	}
	if !strings.Contains(got, "# Body") {
		t.Error("should preserve body")
	}
}

func TestInjectFrontmatter_OpenCodeTranslatesModel(t *testing.T) {
	input := "---\nname: test\ndescription: a test\n---\n\n# Body\n"
	fields := map[string]string{"model": "haiku"}
	got := string(injectFrontmatter([]byte(input), fields, TargetOpenCode))

	if !strings.Contains(got, "model: anthropic/claude-haiku-4-5-20251001") {
		t.Errorf("OpenCode should translate model alias, got:\n%s", got)
	}
}

func TestInjectFrontmatter_MultipleFields(t *testing.T) {
	input := "---\nname: test\ndescription: a test\n---\n\n# Body\n"
	fields := map[string]string{"model": "haiku", "disable-model-invocation": "true"}
	got := string(injectFrontmatter([]byte(input), fields, TargetClaude))

	if !strings.Contains(got, "model: haiku") {
		t.Error("should inject model")
	}
	if !strings.Contains(got, "disable-model-invocation: true") {
		t.Error("should inject disable-model-invocation")
	}
}

func TestInjectFrontmatter_NoFrontmatter(t *testing.T) {
	input := "# No frontmatter\nJust body.\n"
	got := injectFrontmatter([]byte(input), map[string]string{"model": "haiku"}, TargetClaude)
	if string(got) != input {
		t.Error("content without frontmatter should be returned as-is")
	}
}

// --- Prompt structure tests ---

func TestPromptStructure_HasRequiredSections(t *testing.T) {
	// Pipeline skills should have Goal, Invariants, and Principles sections
	pipelineSkills := []string{
		"rpi-research", "rpi-propose", "rpi-plan", "rpi-implement",
		"rpi-verify", "rpi-diagnose", "rpi-explain", "rpi-commit", "rpi-archive",
	}

	for _, skill := range pipelineSkills {
		data, err := ReadAsset("skills/" + skill + "/SKILL.md")
		if err != nil {
			t.Fatalf("read %s: %v", skill, err)
		}
		content := string(data)
		for _, section := range []string{"## Goal", "## Invariants", "## Principles"} {
			if !strings.Contains(content, section) {
				t.Errorf("%s missing required section %q", skill, section)
			}
		}
	}
}

func TestPromptStructure_LineCount(t *testing.T) {
	// Each pipeline skill prompt ≤50 lines excluding YAML frontmatter
	pipelineSkills := []string{
		"rpi-research", "rpi-propose", "rpi-plan", "rpi-implement",
		"rpi-verify", "rpi-diagnose", "rpi-explain", "rpi-commit", "rpi-archive",
	}

	for _, skill := range pipelineSkills {
		data, err := ReadAsset("skills/" + skill + "/SKILL.md")
		if err != nil {
			t.Fatalf("read %s: %v", skill, err)
		}
		body := extractBody(string(data))
		lines := strings.Count(strings.TrimSpace(body), "\n") + 1
		if lines > 50 {
			t.Errorf("%s has %d lines (excluding frontmatter), max is 50", skill, lines)
		}
	}
}

func TestPromptStructure_NoToolReferences(t *testing.T) {
	// No skill should contain rpi_ (MCP tool names) or backtick-quoted rpi CLI invocations
	entries, err := fs.ReadDir(assets, "assets/skills")
	if err != nil {
		t.Fatalf("read assets/skills: %v", err)
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		data, err := assets.ReadFile("assets/skills/" + e.Name() + "/SKILL.md")
		if err != nil {
			t.Errorf("read %s: %v", e.Name(), err)
			continue
		}
		content := string(data)
		if strings.Contains(content, "rpi_") {
			t.Errorf("%s contains MCP tool name reference (rpi_)", e.Name())
		}
		if strings.Contains(content, "`rpi ") {
			t.Errorf("%s contains backtick-quoted rpi CLI invocation", e.Name())
		}
	}
}

// --- InstallTo tests (templates only) ---

func TestInstallTo_SkipsSkills(t *testing.T) {
	dir := t.TempDir()

	count, err := InstallTo(dir, TargetClaude, false)
	if err != nil {
		t.Fatalf("InstallTo error: %v", err)
	}

	// Should install templates but not skills
	if _, err := os.Stat(filepath.Join(dir, "templates", "CLAUDE.md.template")); err != nil {
		t.Error("templates should be installed")
	}
	if _, err := os.Stat(filepath.Join(dir, "skills")); err == nil {
		t.Error("skills directory should not be created by InstallTo")
	}
	if count == 0 {
		t.Error("expected at least template files to be installed")
	}
}

func TestInstall_BackwardCompatible(t *testing.T) {
	dir := t.TempDir()

	count, err := Install(dir, false)
	if err != nil {
		t.Fatalf("Install error: %v", err)
	}
	if count == 0 {
		t.Fatal("expected files to be installed")
	}

	// Should install templates
	if _, err := os.Stat(filepath.Join(dir, "templates", "CLAUDE.md.template")); err != nil {
		t.Error("Install() should install templates")
	}
}

// --- helpers ---

func extractFrontmatter(content string) string {
	if !strings.HasPrefix(content, "---\n") {
		return ""
	}
	end := strings.Index(content[4:], "\n---")
	if end < 0 {
		return ""
	}
	return content[4 : 4+end]
}

func extractBody(content string) string {
	parts := strings.SplitN(content, "---", 3)
	if len(parts) >= 3 {
		return strings.TrimSpace(parts[2])
	}
	return strings.TrimSpace(content)
}

func extractFrontmatterField(content, field string) string {
	fm := extractFrontmatter(content)
	for _, line := range strings.Split(fm, "\n") {
		if strings.HasPrefix(line, field+": ") {
			return strings.TrimPrefix(line, field+": ")
		}
	}
	return ""
}
