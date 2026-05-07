package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// TestSyncProjectGlobalSkipsProjectBlocks verifies that syncProject with
// global: true installs skills/agents/settings into the target dir but
// does NOT create .rpi/, templates, the rules file, or .gitignore.
func TestSyncProjectGlobalSkipsProjectBlocks(t *testing.T) {
	tmp := t.TempDir()

	cfg, err := resolveTargetConfig("claude")
	if err != nil {
		t.Fatalf("resolveTargetConfig: %v", err)
	}

	// Create the tool subdirs upfront — runInit normally does this, and
	// syncProject expects them to exist for skill/agent install. Mirror
	// the global init code path which skips the "tool dir already
	// exists" guard.
	for _, d := range cfg.subdirs {
		if err := os.MkdirAll(filepath.Join(tmp, cfg.toolDir, d), 0755); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
	}

	if err := syncProject(syncOptions{
		targetDir: tmp,
		cfg:       cfg,
		global:    true,
		w:         &bytes.Buffer{},
	}); err != nil {
		t.Fatalf("syncProject(global=true) failed: %v", err)
	}

	// Project-side blocks must be skipped.
	for _, missing := range []string{
		".rpi",
		".rpi/templates",
		"CLAUDE.md",
		".gitignore",
	} {
		if _, err := os.Stat(filepath.Join(tmp, missing)); !os.IsNotExist(err) {
			t.Errorf("global mode created %s; expected it to be skipped", missing)
		}
	}

	// Skills install still runs.
	skillPath := filepath.Join(tmp, ".claude", "skills", "rpi-research", "SKILL.md")
	if _, err := os.Stat(skillPath); err != nil {
		t.Errorf("skills not installed under global mode: %v", err)
	}

	// Agents install still runs (Claude target only).
	agentPath := filepath.Join(tmp, ".claude", "agents", "rpi-verify.md")
	if _, err := os.Stat(agentPath); err != nil {
		t.Errorf("agents not installed under global mode: %v", err)
	}

	// settings.json still gets configured.
	settingsPath := filepath.Join(tmp, ".claude", "settings.json")
	if _, err := os.Stat(settingsPath); err != nil {
		t.Errorf("settings.json not written under global mode: %v", err)
	}
}
