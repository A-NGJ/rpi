package workflow

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTransformCommandFrontmatter_Opus(t *testing.T) {
	input := []byte("---\ndescription: test\nmodel: opus\n---\n\n# Body\n")
	got := string(transformCommandFrontmatter(input))
	if !strings.Contains(got, "model: anthropic/claude-opus-4-6") {
		t.Errorf("expected model: anthropic/claude-opus-4-6, got:\n%s", got)
	}
	if strings.Contains(got, "model: opus") {
		t.Error("original model: opus should be replaced")
	}
}

func TestTransformCommandFrontmatter_Sonnet(t *testing.T) {
	input := []byte("---\ndescription: test\nmodel: sonnet\n---\n\n# Body\n")
	got := string(transformCommandFrontmatter(input))
	if !strings.Contains(got, "model: anthropic/claude-sonnet-4-6") {
		t.Errorf("expected model: anthropic/claude-sonnet-4-6, got:\n%s", got)
	}
}

func TestTransformCommandFrontmatter_UnknownModel(t *testing.T) {
	input := []byte("---\ndescription: test\nmodel: custom-model\n---\n\n# Body\n")
	got := string(transformCommandFrontmatter(input))
	if !strings.Contains(got, "model: custom-model") {
		t.Errorf("unknown model should pass through unchanged, got:\n%s", got)
	}
}

func TestTransformCommandFrontmatter_PreservesBody(t *testing.T) {
	input := []byte("---\nmodel: opus\n---\n\n# Body\nSome content here\n")
	got := string(transformCommandFrontmatter(input))
	if !strings.Contains(got, "# Body") {
		t.Error("body content should be preserved")
	}
	if !strings.Contains(got, "Some content here") {
		t.Error("body content should be preserved")
	}
}

func TestTransformAgentFrontmatter_Basic(t *testing.T) {
	input := "---\nname: codebase-analyzer\ndescription: Analyzes stuff\ntools: Read, Grep, Glob, LS\nmodel: inherit\n---\n\nBody content here.\n"
	got, err := transformAgentFrontmatter([]byte(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	s := string(got)

	if !strings.Contains(s, "mode: subagent") {
		t.Error("expected mode: subagent")
	}
	if !strings.Contains(s, "bash: false") {
		t.Error("expected bash: false in tools deny-map")
	}
	if !strings.Contains(s, "write: false") {
		t.Error("expected write: false in tools deny-map")
	}
	if !strings.Contains(s, "edit: false") {
		t.Error("expected edit: false in tools deny-map")
	}
	if strings.Contains(s, "model:") {
		t.Error("model: field should be removed")
	}
	if strings.Contains(s, "Read, Grep") {
		t.Error("original tools string should be replaced")
	}
	if !strings.Contains(s, "Body content here.") {
		t.Error("body content should be preserved")
	}
	if !strings.Contains(s, "name: codebase-analyzer") {
		t.Error("name should be preserved")
	}
}

func TestTransformAgentFrontmatter_NoFrontmatter(t *testing.T) {
	input := []byte("# No frontmatter\nJust body.\n")
	got, err := transformAgentFrontmatter(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(got) != string(input) {
		t.Error("content without frontmatter should be returned as-is")
	}
}

func TestInstallTo_OpenCode(t *testing.T) {
	dir := t.TempDir()

	count, err := InstallTo(dir, TargetOpenCode, false)
	if err != nil {
		t.Fatalf("InstallTo error: %v", err)
	}
	if count == 0 {
		t.Fatal("expected files to be installed")
	}

	// Verify command frontmatter transform
	cmdData, err := os.ReadFile(filepath.Join(dir, "commands", "rpi-research.md"))
	if err != nil {
		t.Fatalf("read rpi-research.md: %v", err)
	}
	if !strings.Contains(string(cmdData), "model: inherit") {
		t.Error("command model: inherit should pass through unchanged")
	}

	// Verify command body is tool-agnostic (no Sub-task syntax)
	if strings.Contains(string(cmdData), "Sub-task") {
		t.Error("command body should not contain tool-specific Sub-task syntax")
	}

	// Verify agent frontmatter transform
	agentData, err := os.ReadFile(filepath.Join(dir, "agents", "codebase-analyzer.md"))
	if err != nil {
		t.Fatalf("read codebase-analyzer.md: %v", err)
	}
	s := string(agentData)
	if !strings.Contains(s, "mode: subagent") {
		t.Error("agent should have mode: subagent")
	}
	if !strings.Contains(s, "bash: false") {
		t.Error("agent should have bash: false")
	}
	if strings.Contains(s, "tools: Read, Grep") {
		t.Error("original tools string should be replaced")
	}

	// Verify skills are copied as-is (no transform needed)
	skillData, err := os.ReadFile(filepath.Join(dir, "skills", "locate-codebase", "SKILL.md"))
	if err != nil {
		t.Fatalf("read skill: %v", err)
	}
	if !strings.Contains(string(skillData), "name: locate-codebase") {
		t.Error("skill should be copied unchanged")
	}
}

func TestInstallTo_Claude_Unchanged(t *testing.T) {
	dir := t.TempDir()

	_, err := InstallTo(dir, TargetClaude, false)
	if err != nil {
		t.Fatalf("InstallTo error: %v", err)
	}

	// Verify command keeps short model alias
	cmdData, err := os.ReadFile(filepath.Join(dir, "commands", "rpi-research.md"))
	if err != nil {
		t.Fatalf("read rpi-research.md: %v", err)
	}
	if !strings.Contains(string(cmdData), "model: inherit") {
		t.Error("Claude target should preserve original model: inherit")
	}

	// Verify agent keeps original format
	agentData, err := os.ReadFile(filepath.Join(dir, "agents", "codebase-analyzer.md"))
	if err != nil {
		t.Fatalf("read codebase-analyzer.md: %v", err)
	}
	if !strings.Contains(string(agentData), "tools: Read, Grep, Glob, LS") {
		t.Error("Claude target should preserve original tools format")
	}
}

// --- Prompt structure tests (GG-5, GG-6, GG-7) ---

var commandFiles = []string{
	"commands/rpi-research.md",
	"commands/rpi-propose.md",
	"commands/rpi-plan.md",
	"commands/rpi-implement.md",
	"commands/rpi-verify.md",
	"commands/rpi-archive.md",
	"commands/rpi-commit.md",
}

func TestPromptStructure_HasRequiredSections(t *testing.T) {
	// GG-5: Each prompt contains Goal, Invariants, and Principles sections
	for _, file := range commandFiles {
		data, err := ReadAsset(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		content := string(data)
		for _, section := range []string{"## Goal", "## Invariants", "## Principles"} {
			if !strings.Contains(content, section) {
				t.Errorf("%s missing required section %q", file, section)
			}
		}
	}
}

func TestPromptStructure_LineCount(t *testing.T) {
	// GG-6: Each prompt is ≤50 lines excluding YAML frontmatter
	for _, file := range commandFiles {
		data, err := ReadAsset(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		content := string(data)

		// Strip frontmatter (between first and second ---)
		parts := strings.SplitN(content, "---", 3)
		body := ""
		if len(parts) >= 3 {
			body = strings.TrimSpace(parts[2])
		} else {
			body = strings.TrimSpace(content)
		}

		lines := strings.Count(body, "\n") + 1
		if lines > 50 {
			t.Errorf("%s has %d lines (excluding frontmatter), max is 50", file, lines)
		}
	}
}

func TestPromptStructure_NoToolReferences(t *testing.T) {
	// GG-7: No prompt contains rpi_ (MCP tool names) or backtick-quoted rpi CLI invocations
	for _, file := range commandFiles {
		data, err := ReadAsset(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		content := string(data)
		if strings.Contains(content, "rpi_") {
			t.Errorf("%s contains MCP tool name reference (rpi_)", file)
		}
		if strings.Contains(content, "`rpi ") {
			t.Errorf("%s contains backtick-quoted rpi CLI invocation", file)
		}
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

	// Should behave identically to InstallTo with TargetClaude
	cmdData, err := os.ReadFile(filepath.Join(dir, "commands", "rpi-research.md"))
	if err != nil {
		t.Fatalf("read rpi-research.md: %v", err)
	}
	if !strings.Contains(string(cmdData), "model: inherit") {
		t.Error("Install() should preserve Claude Code format")
	}
}
