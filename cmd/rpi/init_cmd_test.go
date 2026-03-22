package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func resetInitFlags() {
	initNoClaudeMD = false
	initTrackRpi = false
	initTarget = "claude"
	initNoMCP = false
}

func runInitInDir(t *testing.T, dir string) (*bytes.Buffer, error) {
	t.Helper()
	resetInitFlags()
	buf := new(bytes.Buffer)
	cmd := initCmd
	cmd.SetOut(buf)
	err := cmd.RunE(cmd, []string{dir})
	return buf, err
}

func TestInitCreatesAllDirs(t *testing.T) {
	dir := t.TempDir()
	buf, err := runInitInDir(t, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify .claude/ subdirs
	claudeSubdirs := []string{"agents", "commands", "skills", "hooks"}
	for _, d := range claudeSubdirs {
		path := filepath.Join(dir, ".claude", d)
		info, err := os.Stat(path)
		if err != nil {
			t.Errorf(".claude/%s not created: %v", d, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf(".claude/%s is not a directory", d)
		}
	}

	// Verify .rpi/ subdirs
	rpiSubdirs := []string{
		"research", "designs",
		"plans", "specs", "reviews", "archive",
	}
	for _, d := range rpiSubdirs {
		path := filepath.Join(dir, ".rpi", d)
		info, err := os.Stat(path)
		if err != nil {
			t.Errorf(".rpi/%s not created: %v", d, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf(".rpi/%s is not a directory", d)
		}
	}

	// Verify CLAUDE.md created
	if _, err := os.Stat(filepath.Join(dir, "CLAUDE.md")); err != nil {
		t.Error("CLAUDE.md not created")
	}

	// Verify .gitignore entries
	gitignore, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if err != nil {
		t.Fatalf("failed to read .gitignore: %v", err)
	}
	if !strings.Contains(string(gitignore), ".claude/") {
		t.Error(".gitignore missing .claude/ entry")
	}
	if !strings.Contains(string(gitignore), ".rpi/") {
		t.Error(".gitignore missing .rpi/ entry")
	}

	output := buf.String()
	if !strings.Contains(output, "Created .claude/agents/") {
		t.Error("output missing .claude/agents/ creation message")
	}
	if !strings.Contains(output, "Created .rpi/research/") {
		t.Error("output missing .rpi/research/ creation message")
	}
}

func TestInitIdempotent(t *testing.T) {
	dir := t.TempDir()

	// First run
	_, err := runInitInDir(t, dir)
	if err != nil {
		t.Fatalf("first run error: %v", err)
	}

	// Second run should error (`.claude/` exists)
	_, err = runInitInDir(t, dir)
	if err == nil {
		t.Fatal("second run should return error")
	}
	if !strings.Contains(err.Error(), ".claude/ already exists") {
		t.Errorf("expected '.claude/ already exists' error, got: %v", err)
	}
}

func TestInitPartial(t *testing.T) {
	dir := t.TempDir()

	// Pre-create some .rpi/ dirs but not .claude/
	os.MkdirAll(filepath.Join(dir, ".rpi", "research"), 0755)
	os.MkdirAll(filepath.Join(dir, ".rpi", "plans"), 0755)

	_, err := runInitInDir(t, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// All dirs should exist
	for _, d := range []string{"designs", "specs", "reviews", "archive"} {
		path := filepath.Join(dir, ".rpi", d)
		if _, err := os.Stat(path); err != nil {
			t.Errorf(".rpi/%s not created: %v", d, err)
		}
	}
}

func TestInitNoClaudeMD(t *testing.T) {
	dir := t.TempDir()

	resetInitFlags()
	initNoClaudeMD = true
	buf := new(bytes.Buffer)
	cmd := initCmd
	cmd.SetOut(buf)
	err := cmd.RunE(cmd, []string{dir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// CLAUDE.md should not exist
	if _, err := os.Stat(filepath.Join(dir, "CLAUDE.md")); err == nil {
		t.Error("CLAUDE.md should not be created with --no-claude-md")
	}

}

func TestInitTrackRpi(t *testing.T) {
	dir := t.TempDir()

	resetInitFlags()
	initTrackRpi = true
	buf := new(bytes.Buffer)
	cmd := initCmd
	cmd.SetOut(buf)
	err := cmd.RunE(cmd, []string{dir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	gitignore, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if err != nil {
		t.Fatalf("failed to read .gitignore: %v", err)
	}

	if strings.Contains(string(gitignore), ".rpi/") {
		t.Error(".rpi/ should not be in .gitignore with --track-rpi")
	}
	if !strings.Contains(string(gitignore), ".claude/") {
		t.Error(".claude/ should still be in .gitignore")
	}
}

func TestInitTargetDir(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "myproject")
	os.MkdirAll(target, 0755)

	_, err := runInitInDir(t, target)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(target, ".claude", "commands")); err != nil {
		t.Error(".claude/commands not created in target dir")
	}
	if _, err := os.Stat(filepath.Join(target, ".rpi", "plans")); err != nil {
		t.Error(".rpi/plans not created in target dir")
	}
	if _, err := os.Stat(filepath.Join(target, "CLAUDE.md")); err != nil {
		t.Error("CLAUDE.md not created in target dir")
	}
}

func TestInitExistingClaudeDir(t *testing.T) {
	dir := t.TempDir()

	// Pre-create .claude/
	os.MkdirAll(filepath.Join(dir, ".claude"), 0755)

	_, err := runInitInDir(t, dir)
	if err == nil {
		t.Fatal("should error when .claude/ already exists")
	}
	if !strings.Contains(err.Error(), ".claude/ already exists") {
		t.Errorf("expected '.claude/ already exists' error, got: %v", err)
	}
}

func TestInitGitignore(t *testing.T) {
	dir := t.TempDir()

	_, err := runInitInDir(t, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if err != nil {
		t.Fatalf("failed to read .gitignore: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, ".claude/") {
		t.Error("missing .claude/ entry")
	}
	if !strings.Contains(content, ".rpi/") {
		t.Error("missing .rpi/ entry")
	}
}

func TestInitTemplateContent(t *testing.T) {
	dir := t.TempDir()

	_, err := runInitInDir(t, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify CLAUDE.md has template content
	claudeData, err := os.ReadFile(filepath.Join(dir, "CLAUDE.md"))
	if err != nil {
		t.Fatalf("read CLAUDE.md: %v", err)
	}
	if !strings.Contains(string(claudeData), "# CLAUDE.md") {
		t.Error("CLAUDE.md missing expected template header")
	}

}

func TestInitInstallsWorkflowFiles(t *testing.T) {
	dir := t.TempDir()
	buf, err := runInitInDir(t, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify embedded agents are installed
	if _, err := os.Stat(filepath.Join(dir, ".claude", "agents", "codebase-analyzer.md")); err != nil {
		t.Error("agents/codebase-analyzer.md not installed")
	}

	// Verify embedded commands are installed
	for _, cmd := range []string{"rpi-plan.md", "rpi-research.md", "rpi-propose.md", "rpi-implement.md"} {
		if _, err := os.Stat(filepath.Join(dir, ".claude", "commands", cmd)); err != nil {
			t.Errorf("commands/%s not installed", cmd)
		}
	}

	// Verify embedded skills are installed
	for _, skill := range []string{"find-patterns", "analyze-thoughts", "locate-codebase", "locate-thoughts"} {
		if _, err := os.Stat(filepath.Join(dir, ".claude", "skills", skill, "SKILL.md")); err != nil {
			t.Errorf("skills/%s/SKILL.md not installed", skill)
		}
	}

	output := buf.String()
	if !strings.Contains(output, "Installed") {
		t.Error("output missing install confirmation")
	}
}

func TestInitAddsRpiToGitignore(t *testing.T) {
	dir := t.TempDir()

	_, err := runInitInDir(t, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if err != nil {
		t.Fatalf("failed to read .gitignore: %v", err)
	}

	if !strings.Contains(string(data), ".rpi/") {
		t.Error(".gitignore missing .rpi/ entry")
	}
}

func TestInitBuildsIndex(t *testing.T) {
	dir := t.TempDir()

	// Create a Go source file so the index has something to find.
	os.MkdirAll(filepath.Join(dir, "pkg"), 0755)
	os.WriteFile(filepath.Join(dir, "pkg", "main.go"), []byte("package main\n\nfunc main() {}\n"), 0644)

	buf, err := runInitInDir(t, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify index file was created.
	indexPath := filepath.Join(dir, ".rpi", "index.json")
	if _, err := os.Stat(indexPath); err != nil {
		t.Errorf("index file not created: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Built codebase index") {
		t.Errorf("output missing index build message, got: %s", output)
	}
}

func TestInitSucceedsWithEmptyDir(t *testing.T) {
	dir := t.TempDir()

	buf, err := runInitInDir(t, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Init should succeed even with no source files (0 files, 0 symbols).
	output := buf.String()
	if !strings.Contains(output, "Built codebase index (0 files, 0 symbols)") {
		t.Errorf("expected empty index message, got: %s", output)
	}
}

func TestCopyDirectory(t *testing.T) {
	src := t.TempDir()
	dest := filepath.Join(t.TempDir(), "dest")

	// Create test files
	os.WriteFile(filepath.Join(src, "a.txt"), []byte("aaa"), 0644)
	os.WriteFile(filepath.Join(src, "b.txt"), []byte("bbb"), 0644)
	os.MkdirAll(filepath.Join(src, "sub"), 0755)
	os.WriteFile(filepath.Join(src, "sub", "c.txt"), []byte("ccc"), 0644)

	count, err := copyDirectory(src, dest)
	if err != nil {
		t.Fatalf("copyDirectory error: %v", err)
	}
	if count != 3 { // a.txt, b.txt, sub/
		t.Errorf("expected 3 items copied, got %d", count)
	}

	// Verify content
	data, _ := os.ReadFile(filepath.Join(dest, "a.txt"))
	if string(data) != "aaa" {
		t.Errorf("a.txt content: got %q, want %q", data, "aaa")
	}
	data, _ = os.ReadFile(filepath.Join(dest, "sub", "c.txt"))
	if string(data) != "ccc" {
		t.Errorf("sub/c.txt content: got %q, want %q", data, "ccc")
	}
}

func runInitOpenCode(t *testing.T, dir string) (*bytes.Buffer, error) {
	t.Helper()
	resetInitFlags()
	initTarget = "opencode"
	buf := new(bytes.Buffer)
	cmd := initCmd
	cmd.SetOut(buf)
	err := cmd.RunE(cmd, []string{dir})
	return buf, err
}

func TestInitOpenCode(t *testing.T) {
	dir := t.TempDir()
	buf, err := runInitOpenCode(t, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify .opencode/ subdirs
	for _, d := range []string{"agents", "commands", "skills", "hooks"} {
		path := filepath.Join(dir, ".opencode", d)
		info, err := os.Stat(path)
		if err != nil {
			t.Errorf(".opencode/%s not created: %v", d, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf(".opencode/%s is not a directory", d)
		}
	}

	// Verify AGENTS.md generated, CLAUDE.md absent
	if _, err := os.Stat(filepath.Join(dir, "AGENTS.md")); err != nil {
		t.Error("AGENTS.md not created")
	}
	if _, err := os.Stat(filepath.Join(dir, "CLAUDE.md")); err == nil {
		t.Error("CLAUDE.md should not be created for opencode target")
	}

	// Verify .opencode/ in .gitignore
	gitignore, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if err != nil {
		t.Fatalf("failed to read .gitignore: %v", err)
	}
	if !strings.Contains(string(gitignore), ".opencode/") {
		t.Error(".gitignore missing .opencode/ entry")
	}
	// .claude/ should NOT be in .gitignore
	if strings.Contains(string(gitignore), ".claude/") {
		t.Error(".gitignore should not contain .claude/ for opencode target")
	}

	// Verify command frontmatter transform (model: opus → full ID)
	cmdData, err := os.ReadFile(filepath.Join(dir, ".opencode", "commands", "rpi-research.md"))
	if err != nil {
		t.Fatalf("read rpi-research.md: %v", err)
	}
	if !strings.Contains(string(cmdData), "model: inherit") {
		t.Error("command model: inherit should pass through unchanged for opencode")
	}

	// Verify command body is tool-agnostic (no Sub-task syntax)
	if strings.Contains(string(cmdData), "Sub-task") {
		t.Error("command body should not contain tool-specific Sub-task syntax")
	}

	// Verify agent frontmatter transform
	agentData, err := os.ReadFile(filepath.Join(dir, ".opencode", "agents", "codebase-analyzer.md"))
	if err != nil {
		t.Fatalf("read codebase-analyzer.md: %v", err)
	}
	if !strings.Contains(string(agentData), "mode: subagent") {
		t.Error("agent should have mode: subagent")
	}

	// Verify TodoWrite removed from rpi-plan.md
	planData, err := os.ReadFile(filepath.Join(dir, ".opencode", "commands", "rpi-plan.md"))
	if err != nil {
		t.Fatalf("read rpi-plan.md: %v", err)
	}
	if strings.Contains(string(planData), "TodoWrite") {
		t.Error("TodoWrite should be removed from command body")
	}

	output := buf.String()
	if !strings.Contains(output, "Created .opencode/agents/") {
		t.Error("output missing .opencode/agents/ creation message")
	}
	if !strings.Contains(output, "Installed") {
		t.Error("output missing install confirmation")
	}
}

func TestInitOpenCodeAlreadyExists(t *testing.T) {
	dir := t.TempDir()

	// First run
	_, err := runInitOpenCode(t, dir)
	if err != nil {
		t.Fatalf("first run error: %v", err)
	}

	// Second run should error
	_, err = runInitOpenCode(t, dir)
	if err == nil {
		t.Fatal("second run should return error")
	}
	if !strings.Contains(err.Error(), ".opencode/ already exists") {
		t.Errorf("expected '.opencode/ already exists' error, got: %v", err)
	}
}

func TestInitInvalidTarget(t *testing.T) {
	dir := t.TempDir()
	resetInitFlags()
	initTarget = "invalid"
	buf := new(bytes.Buffer)
	cmd := initCmd
	cmd.SetOut(buf)
	err := cmd.RunE(cmd, []string{dir})
	if err == nil {
		t.Fatal("should error for invalid target")
	}
	if !strings.Contains(err.Error(), "unknown target") {
		t.Errorf("expected 'unknown target' error, got: %v", err)
	}
}

func TestInitGeneratesCLIReference(t *testing.T) {
	dir := t.TempDir()
	_, err := runInitInDir(t, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cliRefPath := filepath.Join(dir, ".rpi", "cli-reference.md")
	data, err := os.ReadFile(cliRefPath)
	if err != nil {
		t.Fatalf("CLI reference not created: %v", err)
	}
	if !strings.Contains(string(data), "# RPI CLI Reference") {
		t.Error("CLI reference missing expected header")
	}
	if !strings.Contains(string(data), "rpi scan") {
		t.Error("CLI reference missing rpi scan command")
	}
}

// spec:MC-12
func TestInitNoMCPFlag(t *testing.T) {
	flag := initCmd.Flags().Lookup("no-mcp")
	if flag == nil {
		t.Fatal("--no-mcp flag not registered")
	}
	if flag.DefValue != "false" {
		t.Errorf("--no-mcp default = %q, want %q", flag.DefValue, "false")
	}
}

// spec:MC-1 spec:MC-6
func TestInitWritesMCPConfig(t *testing.T) {
	dir := t.TempDir()

	var calls [][]string
	orig := mcpCommandRunner
	mcpCommandRunner = func(name string, args ...string) ([]byte, error) {
		calls = append(calls, append([]string{name}, args...))
		// First call: "claude mcp get rpi" → not found
		if len(calls) == 1 {
			return []byte("No MCP server found"), fmt.Errorf("exit 1")
		}
		// Second call: "claude mcp add rpi -- rpi serve" → success
		return []byte("Added"), nil
	}
	t.Cleanup(func() { mcpCommandRunner = orig })

	buf, err := runInitInDir(t, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(calls) != 2 {
		t.Fatalf("expected 2 claude mcp calls, got %d: %v", len(calls), calls)
	}
	// Verify the add command shape
	addCall := calls[1]
	expected := []string{"claude", "mcp", "add", "rpi", "--", "rpi", "serve"}
	if len(addCall) != len(expected) {
		t.Fatalf("add call = %v, want %v", addCall, expected)
	}
	for i, v := range expected {
		if addCall[i] != v {
			t.Errorf("add call[%d] = %q, want %q", i, addCall[i], v)
		}
	}

	output := buf.String()
	if !strings.Contains(output, "Configured MCP server via claude mcp add") {
		t.Errorf("expected success message, got: %s", output)
	}
}

// spec:MC-3
func TestInitCallsClaudeMCPAdd(t *testing.T) {
	dir := t.TempDir()

	var addCalled bool
	orig := mcpCommandRunner
	mcpCommandRunner = func(name string, args ...string) ([]byte, error) {
		if len(args) >= 2 && args[0] == "mcp" && args[1] == "get" {
			return nil, fmt.Errorf("not found")
		}
		if len(args) >= 2 && args[0] == "mcp" && args[1] == "add" {
			addCalled = true
			return []byte("Added"), nil
		}
		return nil, fmt.Errorf("unexpected call")
	}
	t.Cleanup(func() { mcpCommandRunner = orig })

	_, err := runInitInDir(t, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !addCalled {
		t.Error("claude mcp add was not called")
	}
}

// spec:MC-4
func TestInitWarnsExistingMCPEntry(t *testing.T) {
	dir := t.TempDir()

	var addCalled bool
	orig := mcpCommandRunner
	mcpCommandRunner = func(name string, args ...string) ([]byte, error) {
		if len(args) >= 2 && args[0] == "mcp" && args[1] == "get" {
			// Server already exists
			return []byte("rpi: connected"), nil
		}
		if len(args) >= 2 && args[0] == "mcp" && args[1] == "add" {
			addCalled = true
			return nil, nil
		}
		return nil, fmt.Errorf("unexpected call")
	}
	t.Cleanup(func() { mcpCommandRunner = orig })

	buf, err := runInitInDir(t, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "already configured") {
		t.Error("expected warning about existing MCP entry")
	}
	if addCalled {
		t.Error("claude mcp add should not be called when server already exists")
	}
}

// spec:MC-5
func TestInitSkipsMCPWithFlag(t *testing.T) {
	dir := t.TempDir()

	var called bool
	orig := mcpCommandRunner
	mcpCommandRunner = func(name string, args ...string) ([]byte, error) {
		called = true
		return nil, nil
	}
	t.Cleanup(func() { mcpCommandRunner = orig })

	resetInitFlags()
	initNoMCP = true
	buf := new(bytes.Buffer)
	cmd := initCmd
	cmd.SetOut(buf)
	err := cmd.RunE(cmd, []string{dir})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if called {
		t.Error("claude mcp should not be called with --no-mcp")
	}
}

// spec:MC-6
func TestInitMCPAddCommandShape(t *testing.T) {
	dir := t.TempDir()

	var addArgs []string
	orig := mcpCommandRunner
	mcpCommandRunner = func(name string, args ...string) ([]byte, error) {
		if len(args) >= 2 && args[0] == "mcp" && args[1] == "get" {
			return nil, fmt.Errorf("not found")
		}
		if len(args) >= 2 && args[0] == "mcp" && args[1] == "add" {
			addArgs = append([]string{name}, args...)
			return []byte("Added"), nil
		}
		return nil, fmt.Errorf("unexpected call")
	}
	t.Cleanup(func() { mcpCommandRunner = orig })

	_, err := runInitInDir(t, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify exact command: claude mcp add rpi -- rpi serve
	expected := []string{"claude", "mcp", "add", "rpi", "--", "rpi", "serve"}
	if len(addArgs) != len(expected) {
		t.Fatalf("add args = %v, want %v", addArgs, expected)
	}
	for i, v := range expected {
		if addArgs[i] != v {
			t.Errorf("add args[%d] = %q, want %q", i, addArgs[i], v)
		}
	}
}
