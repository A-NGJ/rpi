package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func resetUpdateFlags() {
	updateForce = false
	updateNoClaudeMD = false
}

func initThenUpdate(t *testing.T, dir string) (*bytes.Buffer, error) {
	t.Helper()
	// Init first
	resetInitFlags()
	buf := new(bytes.Buffer)
	cmd := initCmd
	cmd.SetOut(buf)
	if err := cmd.RunE(cmd, []string{dir}); err != nil {
		t.Fatalf("init failed: %v", err)
	}
	// Run update
	resetUpdateFlags()
	buf = new(bytes.Buffer)
	cmd = updateCmd
	cmd.SetOut(buf)
	err := cmd.RunE(cmd, []string{dir})
	return buf, err
}

func TestUpdateRequiresExistingProject(t *testing.T) {
	dir := t.TempDir()

	resetUpdateFlags()
	buf := new(bytes.Buffer)
	cmd := updateCmd
	cmd.SetOut(buf)
	err := cmd.RunE(cmd, []string{dir})
	if err == nil {
		t.Fatal("update should error without prior init")
	}
	if !strings.Contains(err.Error(), "not initialized") {
		t.Errorf("expected 'not initialized' error, got: %v", err)
	}
}

func TestUpdateCreatesMissingDirs(t *testing.T) {
	dir := t.TempDir()

	// Init
	resetInitFlags()
	buf := new(bytes.Buffer)
	cmd := initCmd
	cmd.SetOut(buf)
	if err := cmd.RunE(cmd, []string{dir}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// Remove some dirs
	os.Remove(filepath.Join(dir, ".rpi", "reviews"))
	os.Remove(filepath.Join(dir, ".claude", "hooks"))

	// Update
	resetUpdateFlags()
	buf = new(bytes.Buffer)
	cmd = updateCmd
	cmd.SetOut(buf)
	if err := cmd.RunE(cmd, []string{dir}); err != nil {
		t.Fatalf("update error: %v", err)
	}

	// Verify dirs recreated
	if _, err := os.Stat(filepath.Join(dir, ".rpi", "reviews")); err != nil {
		t.Error(".rpi/reviews not recreated")
	}
	if _, err := os.Stat(filepath.Join(dir, ".claude", "hooks")); err != nil {
		t.Error(".claude/hooks not recreated")
	}

	output := buf.String()
	if !strings.Contains(output, "Created .rpi/reviews/") {
		t.Error("output missing .rpi/reviews/ creation message")
	}
	if !strings.Contains(output, "Created .claude/hooks/") {
		t.Error("output missing .claude/hooks/ creation message")
	}
}

func TestUpdateDoesNotOverwriteWithoutForce(t *testing.T) {
	dir := t.TempDir()

	// Init
	resetInitFlags()
	buf := new(bytes.Buffer)
	cmd := initCmd
	cmd.SetOut(buf)
	if err := cmd.RunE(cmd, []string{dir}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// Modify a skill file
	skillFile := filepath.Join(dir, ".claude", "skills", "rpi-plan", "SKILL.md")
	os.WriteFile(skillFile, []byte("custom content"), 0644)

	// Update without --force
	resetUpdateFlags()
	buf = new(bytes.Buffer)
	cmd = updateCmd
	cmd.SetOut(buf)
	if err := cmd.RunE(cmd, []string{dir}); err != nil {
		t.Fatalf("update error: %v", err)
	}

	// File should NOT be overwritten
	data, _ := os.ReadFile(skillFile)
	if string(data) != "custom content" {
		t.Error("update without --force should not overwrite existing files")
	}
}

func TestUpdateForceOverwritesFiles(t *testing.T) {
	dir := t.TempDir()

	// Init
	resetInitFlags()
	buf := new(bytes.Buffer)
	cmd := initCmd
	cmd.SetOut(buf)
	if err := cmd.RunE(cmd, []string{dir}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// Modify a skill file
	skillFile := filepath.Join(dir, ".claude", "skills", "rpi-plan", "SKILL.md")
	os.WriteFile(skillFile, []byte("custom content"), 0644)

	// Update with --force
	resetUpdateFlags()
	updateForce = true
	buf = new(bytes.Buffer)
	cmd = updateCmd
	cmd.SetOut(buf)
	if err := cmd.RunE(cmd, []string{dir}); err != nil {
		t.Fatalf("update --force error: %v", err)
	}

	// File should be overwritten
	data, _ := os.ReadFile(skillFile)
	if string(data) == "custom content" {
		t.Error("update --force should overwrite existing files")
	}
}

func TestUpdateRegeneratesIndex(t *testing.T) {
	dir := t.TempDir()

	buf, err := initThenUpdate(t, dir)
	if err != nil {
		t.Fatalf("update error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Built codebase index") {
		t.Error("output missing index build message")
	}

	indexPath := filepath.Join(dir, ".rpi", "index.json")
	if _, err := os.Stat(indexPath); err != nil {
		t.Error("index.json not present after update")
	}
}

// spec:IU-3
func TestUpdatePreservesRulesFileWithoutForce(t *testing.T) {
	dir := t.TempDir()

	// Init
	resetInitFlags()
	buf := new(bytes.Buffer)
	cmd := initCmd
	cmd.SetOut(buf)
	if err := cmd.RunE(cmd, []string{dir}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// Modify CLAUDE.md
	claudeMD := filepath.Join(dir, "CLAUDE.md")
	os.WriteFile(claudeMD, []byte("custom content"), 0644)

	// Update without --force
	resetUpdateFlags()
	buf = new(bytes.Buffer)
	cmd = updateCmd
	cmd.SetOut(buf)
	if err := cmd.RunE(cmd, []string{dir}); err != nil {
		t.Fatalf("update error: %v", err)
	}

	// CLAUDE.md should be untouched
	data, _ := os.ReadFile(claudeMD)
	if string(data) != "custom content" {
		t.Error("update without --force should not overwrite CLAUDE.md")
	}
}

// spec:IU-4
func TestUpdateForceOverwritesRulesFile(t *testing.T) {
	dir := t.TempDir()

	// Init
	resetInitFlags()
	buf := new(bytes.Buffer)
	cmd := initCmd
	cmd.SetOut(buf)
	if err := cmd.RunE(cmd, []string{dir}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// Modify CLAUDE.md
	claudeMD := filepath.Join(dir, "CLAUDE.md")
	os.WriteFile(claudeMD, []byte("custom content"), 0644)

	// Update with --force
	resetUpdateFlags()
	updateForce = true
	buf = new(bytes.Buffer)
	cmd = updateCmd
	cmd.SetOut(buf)
	if err := cmd.RunE(cmd, []string{dir}); err != nil {
		t.Fatalf("update --force error: %v", err)
	}

	// CLAUDE.md should be overwritten with template
	data, _ := os.ReadFile(claudeMD)
	if string(data) == "custom content" {
		t.Error("update --force should overwrite CLAUDE.md with template")
	}
	if !strings.Contains(string(data), "# CLAUDE.md") {
		t.Error("CLAUDE.md missing template header after update --force")
	}
}

func TestUpdateNoClaudeMDSkipsRulesFile(t *testing.T) {
	dir := t.TempDir()

	// Init
	resetInitFlags()
	buf := new(bytes.Buffer)
	cmd := initCmd
	cmd.SetOut(buf)
	if err := cmd.RunE(cmd, []string{dir}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// Modify CLAUDE.md
	claudeMD := filepath.Join(dir, "CLAUDE.md")
	os.WriteFile(claudeMD, []byte("custom content"), 0644)

	// Update with --no-claude-md
	resetUpdateFlags()
	updateNoClaudeMD = true
	buf = new(bytes.Buffer)
	cmd = updateCmd
	cmd.SetOut(buf)
	if err := cmd.RunE(cmd, []string{dir}); err != nil {
		t.Fatalf("update error: %v", err)
	}

	// CLAUDE.md should be untouched
	data, _ := os.ReadFile(claudeMD)
	if string(data) != "custom content" {
		t.Error("update --no-claude-md should not modify CLAUDE.md")
	}
}

func TestUpdateAddsSettingsPermission(t *testing.T) {
	dir := t.TempDir()

	// Init
	resetInitFlags()
	buf := new(bytes.Buffer)
	cmd := initCmd
	cmd.SetOut(buf)
	if err := cmd.RunE(cmd, []string{dir}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// Remove the permission from settings.json
	settingsPath := filepath.Join(dir, ".claude", "settings.json")
	os.WriteFile(settingsPath, []byte(`{"permissions":{"allow":[]}}`), 0644)

	// Update should re-add it
	resetUpdateFlags()
	buf = new(bytes.Buffer)
	cmd = updateCmd
	cmd.SetOut(buf)
	if err := cmd.RunE(cmd, []string{dir}); err != nil {
		t.Fatalf("update error: %v", err)
	}

	data, _ := os.ReadFile(settingsPath)
	if !strings.Contains(string(data), "mcp__rpi__*") {
		t.Error("update should add mcp__rpi__* permission to settings.json")
	}
}

func TestUpdateDetectsOpenCodeTarget(t *testing.T) {
	dir := t.TempDir()

	// Init with opencode
	resetInitFlags()
	initTarget = "opencode"
	buf := new(bytes.Buffer)
	cmd := initCmd
	cmd.SetOut(buf)
	if err := cmd.RunE(cmd, []string{dir}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// Remove a subdir
	os.Remove(filepath.Join(dir, ".opencode", "hooks"))

	// Update (should auto-detect opencode)
	resetUpdateFlags()
	buf = new(bytes.Buffer)
	cmd = updateCmd
	cmd.SetOut(buf)
	if err := cmd.RunE(cmd, []string{dir}); err != nil {
		t.Fatalf("update error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, ".opencode", "hooks")); err != nil {
		t.Error(".opencode/hooks not recreated")
	}

	output := buf.String()
	if !strings.Contains(output, "Created .opencode/hooks/") {
		t.Error("output missing .opencode/hooks/ creation message")
	}
}

// TC-6: Update with existing commands dir
func TestUpdatePreservesExistingCommandsDir(t *testing.T) {
	dir := t.TempDir()

	// Manually set up an old-style project with .claude/commands/
	os.MkdirAll(filepath.Join(dir, ".claude", "skills"), 0755)
	os.MkdirAll(filepath.Join(dir, ".claude", "hooks"), 0755)
	os.MkdirAll(filepath.Join(dir, ".claude", "commands"), 0755)
	os.WriteFile(filepath.Join(dir, ".claude", "commands", "rpi-propose.md"), []byte("old command"), 0644)
	os.MkdirAll(filepath.Join(dir, ".rpi", "plans"), 0755)

	// Run update
	resetUpdateFlags()
	updateForce = true
	buf := new(bytes.Buffer)
	cmd := updateCmd
	cmd.SetOut(buf)
	if err := cmd.RunE(cmd, []string{dir}); err != nil {
		t.Fatalf("update error: %v", err)
	}

	// AS-13: .claude/commands/ must not be deleted
	data, err := os.ReadFile(filepath.Join(dir, ".claude", "commands", "rpi-propose.md"))
	if err != nil {
		t.Error(".claude/commands/rpi-propose.md was deleted (AS-13)")
	} else if string(data) != "old command" {
		t.Error(".claude/commands/rpi-propose.md was modified (AS-13)")
	}

	// .claude/skills/ should have new skills
	entries, err := os.ReadDir(filepath.Join(dir, ".claude", "skills"))
	if err != nil {
		t.Fatalf("read .claude/skills/: %v", err)
	}
	if len(entries) != 9 {
		t.Errorf("expected 9 skill dirs in .claude/skills/, got %d", len(entries))
	}
}

func TestUpdateDetectsAgentsOnlyTarget(t *testing.T) {
	dir := t.TempDir()

	// Init with agents-only
	resetInitFlags()
	initTarget = "agents-only"
	buf := new(bytes.Buffer)
	cmd := initCmd
	cmd.SetOut(buf)
	if err := cmd.RunE(cmd, []string{dir}); err != nil {
		t.Fatalf("init failed: %v", err)
	}

	// Update (should auto-detect agents-only)
	resetUpdateFlags()
	buf = new(bytes.Buffer)
	cmd = updateCmd
	cmd.SetOut(buf)
	if err := cmd.RunE(cmd, []string{dir}); err != nil {
		t.Fatalf("update error: %v", err)
	}

	// Should still have 9 skills
	entries, err := os.ReadDir(filepath.Join(dir, ".agents", "skills"))
	if err != nil {
		t.Fatalf("read .agents/skills/: %v", err)
	}
	if len(entries) != 9 {
		t.Errorf("expected 9 skill dirs, got %d", len(entries))
	}

	// No tool-specific dirs should be created
	if _, err := os.Stat(filepath.Join(dir, ".claude")); err == nil {
		t.Error(".claude/ should not be created for agents-only update")
	}
}
