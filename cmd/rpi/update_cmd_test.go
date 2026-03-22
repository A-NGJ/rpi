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

	// Modify a workflow file
	cmdFile := filepath.Join(dir, ".claude", "commands", "rpi-plan.md")
	os.WriteFile(cmdFile, []byte("custom content"), 0644)

	// Update without --force
	resetUpdateFlags()
	buf = new(bytes.Buffer)
	cmd = updateCmd
	cmd.SetOut(buf)
	if err := cmd.RunE(cmd, []string{dir}); err != nil {
		t.Fatalf("update error: %v", err)
	}

	// File should NOT be overwritten
	data, _ := os.ReadFile(cmdFile)
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

	// Modify a workflow file
	cmdFile := filepath.Join(dir, ".claude", "commands", "rpi-plan.md")
	os.WriteFile(cmdFile, []byte("custom content"), 0644)

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
	data, _ := os.ReadFile(cmdFile)
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
	if !strings.Contains(output, "Rebuilt codebase index") {
		t.Error("output missing index rebuild message")
	}

	indexPath := filepath.Join(dir, ".rpi", "index.json")
	if _, err := os.Stat(indexPath); err != nil {
		t.Error("index.json not present after update")
	}
}

func TestUpdateUpdatesRulesFile(t *testing.T) {
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

	// Update
	resetUpdateFlags()
	buf = new(bytes.Buffer)
	cmd = updateCmd
	cmd.SetOut(buf)
	if err := cmd.RunE(cmd, []string{dir}); err != nil {
		t.Fatalf("update error: %v", err)
	}

	// CLAUDE.md should be overwritten with template
	data, _ := os.ReadFile(claudeMD)
	if string(data) == "custom content" {
		t.Error("update should overwrite CLAUDE.md with template")
	}
	if !strings.Contains(string(data), "# CLAUDE.md") {
		t.Error("CLAUDE.md missing template header after update")
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
