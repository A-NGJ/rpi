package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func resetUninstallFlags() {
	uninstallGlobal = false
	uninstallDryRun = false
}

// writeStandaloneSkill drops a minimal SKILL.md under
// <home>/.claude/skills/<name>/. When marker is true, the frontmatter
// includes a matching `name:` line so the authorship check passes;
// otherwise the frontmatter is empty.
func writeStandaloneSkill(t *testing.T, home, name string, marker bool) {
	t.Helper()
	dir := filepath.Join(home, ".claude", "skills", name)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("mkdir skill %s: %v", name, err)
	}
	body := "---\n---\n\n# stub\n"
	if marker {
		body = "---\nname: " + name + "\ndescription: stub for tests\n---\n\n# stub\n"
	}
	if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte(body), 0644); err != nil {
		t.Fatalf("write SKILL.md: %v", err)
	}
}

func writePluginBinary(t *testing.T, home string) {
	t.Helper()
	binDir := filepath.Join(home, ".rpi", "bin")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		t.Fatalf("mkdir bin: %v", err)
	}
	if err := os.WriteFile(filepath.Join(binDir, "rpi"), []byte("#!/bin/sh\nexit 0\n"), 0755); err != nil {
		t.Fatalf("write bin: %v", err)
	}
}

func writeSettingsJSON(t *testing.T, home string, payload map[string]any) {
	t.Helper()
	claudeDir := filepath.Join(home, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("mkdir .claude: %v", err)
	}
	raw, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		t.Fatalf("marshal settings: %v", err)
	}
	if err := os.WriteFile(filepath.Join(claudeDir, "settings.json"), raw, 0644); err != nil {
		t.Fatalf("write settings.json: %v", err)
	}
}

// snapshotTree builds a deterministic map of every path under root to a
// content fingerprint. Used by dry-run tests to assert nothing changed.
func snapshotTree(t *testing.T, root string) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if info.IsDir() {
			out[rel+"/"] = "dir"
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		out[rel] = string(data)
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", root, err)
	}
	return out
}

func TestUninstallDetectInstallState(t *testing.T) {
	tests := []struct {
		name      string
		seed      func(t *testing.T, home string)
		wantClass string
	}{
		{
			name:      "empty home",
			seed:      func(t *testing.T, home string) {},
			wantClass: installStateNothing,
		},
		{
			name: "plugin binary only",
			seed: func(t *testing.T, home string) {
				writePluginBinary(t, home)
			},
			wantClass: installStatePluginMode,
		},
		{
			name: "standalone skills only",
			seed: func(t *testing.T, home string) {
				writeStandaloneSkill(t, home, "rpi-research", true)
				writeStandaloneSkill(t, home, "rpi-plan", true)
			},
			wantClass: installStateStandalone,
		},
		{
			name: "standalone MCP entry only",
			seed: func(t *testing.T, home string) {
				writeSettingsJSON(t, home, map[string]any{
					"mcpServers": map[string]any{
						"rpi": map[string]any{
							"command": "rpi",
							"args":    []string{"serve"},
						},
					},
				})
			},
			wantClass: installStateStandalone,
		},
		{
			name: "standalone skills plus plugin binary - standalone wins",
			seed: func(t *testing.T, home string) {
				writeStandaloneSkill(t, home, "rpi-research", true)
				writePluginBinary(t, home)
			},
			wantClass: installStateStandalone,
		},
		{
			name: "rpi-prefixed dir without canonical name",
			seed: func(t *testing.T, home string) {
				writeStandaloneSkill(t, home, "rpi-something", true)
			},
			wantClass: installStateNothing,
		},
		{
			name: "rpi-research dir but no frontmatter marker",
			seed: func(t *testing.T, home string) {
				writeStandaloneSkill(t, home, "rpi-research", false)
			},
			wantClass: installStateNothing,
		},
		{
			name: "standalone permission entry only",
			seed: func(t *testing.T, home string) {
				writeSettingsJSON(t, home, map[string]any{
					"permissions": map[string]any{
						"allow": []string{"Bash(rpi scan:*)"},
					},
				})
			},
			wantClass: installStateStandalone,
		},
		{
			name: "standalone hook marker only",
			seed: func(t *testing.T, home string) {
				writeSettingsJSON(t, home, map[string]any{
					"hooks": map[string]any{
						"PostCompact": []map[string]any{{
							"matcher": "",
							"hooks": []map[string]string{{
								"type":    "command",
								"command": "cat <<'HOOK_EOF'\nrpi_context_essentials\nHOOK_EOF",
							}},
						}},
					},
				})
			},
			wantClass: installStateStandalone,
		},
		{
			name: "mcpServers.rpi pointing at the plugin binary - not standalone",
			seed: func(t *testing.T, home string) {
				writeSettingsJSON(t, home, map[string]any{
					"mcpServers": map[string]any{
						"rpi": map[string]any{
							"command": filepath.Join(home, ".rpi", "bin", "rpi"),
							"args":    []string{"serve"},
						},
					},
				})
			},
			wantClass: installStateNothing,
		},
		{
			name: "non-rpi MCP server named rpi-like is ignored",
			seed: func(t *testing.T, home string) {
				writeSettingsJSON(t, home, map[string]any{
					"mcpServers": map[string]any{
						"rpi-like": map[string]any{
							"command": "something-else",
							"args":    []string{"serve"},
						},
					},
				})
			},
			wantClass: installStateNothing,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			home := t.TempDir()
			tc.seed(t, home)
			state, err := detectInstallState(home)
			if err != nil {
				t.Fatalf("detectInstallState: %v", err)
			}
			if got := state.classification(); got != tc.wantClass {
				t.Errorf("classification = %q, want %q", got, tc.wantClass)
			}
		})
	}
}

func TestUninstallDryRunLeavesDiskUnchanged(t *testing.T) {
	home := t.TempDir()
	writeStandaloneSkill(t, home, "rpi-research", true)
	writePluginBinary(t, home)
	writeSettingsJSON(t, home, map[string]any{
		"mcpServers": map[string]any{
			"rpi": map[string]any{
				"command": "rpi",
				"args":    []string{"serve"},
			},
		},
		"permissions": map[string]any{
			"allow": []string{"mcp__rpi__*", "Bash(rpi scan:*)", "Bash(npm test:*)"},
		},
		"hooks": map[string]any{
			"PostCompact": []map[string]any{{
				"matcher": "",
				"hooks": []map[string]string{{
					"type":    "command",
					"command": "cat <<'HOOK_EOF'\nrpi_context_essentials\nHOOK_EOF",
				}},
			}},
		},
	})

	before := snapshotTree(t, home)

	t.Setenv("HOME", home)
	resetUninstallFlags()
	uninstallGlobal = true
	uninstallDryRun = true
	buf := new(bytes.Buffer)
	cmd := uninstallCmd
	cmd.SetOut(buf)
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("uninstall --dry-run: %v", err)
	}

	after := snapshotTree(t, home)
	if !reflect.DeepEqual(before, after) {
		t.Errorf("dry-run modified the filesystem.\nbefore: %v\nafter:  %v", before, after)
	}

	output := buf.String()
	if !strings.Contains(output, "standalone install detected") {
		t.Errorf("expected standalone classification in plan output; got: %s", output)
	}
	for _, fragment := range []string{
		"rpi-research",
		"mcpServers.rpi",
		"permission entries",
		"RPI hook entries",
	} {
		if !strings.Contains(output, fragment) {
			t.Errorf("expected dry-run plan to mention %q, got: %s", fragment, output)
		}
	}
}

func TestUninstallNothingToRemove(t *testing.T) {
	home := t.TempDir()

	t.Setenv("HOME", home)
	resetUninstallFlags()
	uninstallGlobal = true
	buf := new(bytes.Buffer)
	cmd := uninstallCmd
	cmd.SetOut(buf)
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("uninstall: %v", err)
	}
	if !strings.Contains(buf.String(), "Nothing to remove") {
		t.Errorf("expected 'Nothing to remove' message, got: %s", buf.String())
	}
}

func TestUninstallRequiresGlobalFlag(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	resetUninstallFlags()
	buf := new(bytes.Buffer)
	cmd := uninstallCmd
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	if err := cmd.RunE(cmd, nil); err == nil {
		t.Fatal("expected error when --global is not set")
	}
}

// runUninstallGlobal exercises `rpi uninstall --global` with HOME
// redirected to the given dir, no dry-run, and returns the captured
// stdout for assertions.
func runUninstallGlobal(t *testing.T, home string) *bytes.Buffer {
	t.Helper()
	t.Setenv("HOME", home)
	resetUninstallFlags()
	uninstallGlobal = true
	buf := new(bytes.Buffer)
	cmd := uninstallCmd
	cmd.SetOut(buf)
	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("uninstall --global: %v", err)
	}
	return buf
}

func TestUninstallStandaloneFullRemoval(t *testing.T) {
	home := t.TempDir()
	writeStandaloneSkill(t, home, "rpi-research", true)
	writeStandaloneSkill(t, home, "rpi-plan", true)
	// A non-RPI skill that must remain untouched.
	nonRPIDir := filepath.Join(home, ".claude", "skills", "other-skill")
	if err := os.MkdirAll(nonRPIDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(nonRPIDir, "SKILL.md"), []byte("---\nname: other-skill\n---\n"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	// A rpi-prefixed dir lacking the authorship marker — must remain.
	writeStandaloneSkill(t, home, "rpi-something", true) // not in canonical set
	writePluginBinary(t, home)
	writeSettingsJSON(t, home, map[string]any{
		"mcpServers": map[string]any{
			"rpi": map[string]any{
				"command": "rpi",
				"args":    []string{"serve"},
			},
			"rpi-like": map[string]any{
				"command": "rpi-like",
				"args":    []string{"serve"},
			},
		},
		"permissions": map[string]any{
			"allow": []string{
				"mcp__rpi__*",
				"Bash(rpi scan:*)",
				"Bash(rpi status:*)",
				"Bash(rm /tmp/claude-handoff-*.md)",
				"Bash(npm test:*)",
			},
		},
		"hooks": map[string]any{
			"PostCompact": []map[string]any{{
				"matcher": "",
				"hooks": []map[string]string{{
					"type":    "command",
					"command": "cat <<'HOOK_EOF'\nrpi_context_essentials\nHOOK_EOF",
				}},
			}},
			"PreToolUse": []map[string]any{{
				"matcher": "",
				"hooks": []map[string]string{{
					"type":    "command",
					"command": "echo user-hook",
				}},
			}},
		},
		"customKey": "value",
	})

	runUninstallGlobal(t, home)

	// RPI skill dirs gone.
	for _, name := range []string{"rpi-research", "rpi-plan"} {
		if _, err := os.Stat(filepath.Join(home, ".claude", "skills", name)); !os.IsNotExist(err) {
			t.Errorf("standalone skill dir %s should be removed", name)
		}
	}
	// Non-RPI skill preserved.
	if _, err := os.Stat(nonRPIDir); err != nil {
		t.Errorf("non-RPI skill dir was deleted: %v", err)
	}
	// rpi-something (no authorship marker) preserved.
	if _, err := os.Stat(filepath.Join(home, ".claude", "skills", "rpi-something")); err != nil {
		t.Errorf("rpi-prefixed non-RPI dir was deleted: %v", err)
	}
	// ~/.rpi removed.
	if _, err := os.Stat(filepath.Join(home, ".rpi")); !os.IsNotExist(err) {
		t.Errorf("~/.rpi should be removed after standalone cleanup")
	}

	// settings.json must remain valid JSON and preserve unrelated keys.
	data, err := os.ReadFile(filepath.Join(home, ".claude", "settings.json"))
	if err != nil {
		t.Fatalf("read settings.json: %v", err)
	}
	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("settings.json is not valid JSON after scrub: %v", err)
	}
	content := string(data)

	// mcpServers.rpi gone, rpi-like preserved.
	if strings.Contains(content, `"rpi":`) {
		t.Errorf("mcpServers.rpi should be removed; got: %s", content)
	}
	if !strings.Contains(content, "rpi-like") {
		t.Errorf("non-RPI mcpServers entry rpi-like was lost")
	}
	// RPI-owned perms gone, user perms preserved.
	for _, removed := range []string{
		"mcp__rpi__*",
		"Bash(rpi scan:*)",
		"Bash(rpi status:*)",
		"Bash(rm /tmp/claude-handoff-*.md)",
	} {
		if strings.Contains(content, removed) {
			t.Errorf("expected %q to be scrubbed from settings.json, still present", removed)
		}
	}
	if !strings.Contains(content, "Bash(npm test:*)") {
		t.Error("user perm Bash(npm test:*) was lost")
	}
	// PostCompact (RPI) gone, PreToolUse (user) preserved.
	if strings.Contains(content, "rpi_context_essentials") {
		t.Error("RPI hook should be scrubbed")
	}
	if !strings.Contains(content, "PreToolUse") || !strings.Contains(content, "user-hook") {
		t.Error("user hook PreToolUse was lost")
	}
	if val, _ := parsed["customKey"].(string); val != "value" {
		t.Errorf("customKey lost; got %v", parsed["customKey"])
	}
}

func TestUninstallPluginModeRemoval(t *testing.T) {
	home := t.TempDir()
	writePluginBinary(t, home)
	// .claude/ exists but contains no RPI artifacts — must be untouched.
	claudeDir := filepath.Join(home, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	keep := filepath.Join(claudeDir, "settings.json")
	keepContent := `{"customKey":"value"}` + "\n"
	if err := os.WriteFile(keep, []byte(keepContent), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	runUninstallGlobal(t, home)

	if _, err := os.Stat(filepath.Join(home, ".rpi", "bin", "rpi")); !os.IsNotExist(err) {
		t.Errorf("binary should be removed")
	}
	if _, err := os.Stat(filepath.Join(home, ".rpi")); !os.IsNotExist(err) {
		t.Errorf("~/.rpi should be removed")
	}
	// .claude/ untouched.
	got, err := os.ReadFile(keep)
	if err != nil {
		t.Fatalf("read settings.json: %v", err)
	}
	if string(got) != keepContent {
		t.Errorf("plugin-mode uninstall must not touch ~/.claude — content changed:\nwant: %qgot:  %q", keepContent, string(got))
	}
}

func TestUninstallDoubleRunIsNoOp(t *testing.T) {
	home := t.TempDir()
	writeStandaloneSkill(t, home, "rpi-research", true)
	writePluginBinary(t, home)
	writeSettingsJSON(t, home, map[string]any{
		"mcpServers": map[string]any{
			"rpi": map[string]any{"command": "rpi", "args": []string{"serve"}},
		},
	})

	runUninstallGlobal(t, home)
	before := snapshotTree(t, home)
	buf := runUninstallGlobal(t, home)
	after := snapshotTree(t, home)

	if !reflect.DeepEqual(before, after) {
		t.Errorf("second uninstall modified filesystem:\nbefore: %v\nafter:  %v", before, after)
	}
	if !strings.Contains(buf.String(), "Nothing to remove") {
		t.Errorf("expected 'Nothing to remove' on second run, got: %s", buf.String())
	}
}

func TestUninstallPartialPermissionState(t *testing.T) {
	home := t.TempDir()
	writeSettingsJSON(t, home, map[string]any{
		"permissions": map[string]any{
			"allow": []string{"Bash(rpi scan:*)"},
		},
	})

	runUninstallGlobal(t, home)

	data, err := os.ReadFile(filepath.Join(home, ".claude", "settings.json"))
	if err != nil {
		t.Fatalf("settings.json deleted: %v", err)
	}
	var parsed map[string]any
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("settings.json no longer valid JSON: %v\ncontent: %s", err, string(data))
	}
	if strings.Contains(string(data), "Bash(rpi scan:*)") {
		t.Errorf("RPI permission still present after uninstall: %s", string(data))
	}
}

func TestUninstallNonRPIMcpServerPreserved(t *testing.T) {
	home := t.TempDir()
	writeSettingsJSON(t, home, map[string]any{
		"mcpServers": map[string]any{
			"rpi-like": map[string]any{
				"command": "rpi-like",
				"args":    []string{"serve"},
			},
		},
	})

	buf := runUninstallGlobal(t, home)

	if !strings.Contains(buf.String(), "Nothing to remove") {
		t.Errorf("expected nothing-to-remove for non-RPI MCP, got: %s", buf.String())
	}
	data, _ := os.ReadFile(filepath.Join(home, ".claude", "settings.json"))
	if !strings.Contains(string(data), "rpi-like") {
		t.Errorf("non-RPI MCP server was touched: %s", string(data))
	}
}

func TestUninstallNonRPISkillDirPreserved(t *testing.T) {
	home := t.TempDir()
	// Both: a rpi-prefixed dir without the marker, and a non-rpi dir.
	writeStandaloneSkill(t, home, "rpi-research", false)
	custom := filepath.Join(home, ".claude", "skills", "custom-skill")
	if err := os.MkdirAll(custom, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(custom, "SKILL.md"), []byte("---\nname: custom-skill\n---\n"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	buf := runUninstallGlobal(t, home)

	if !strings.Contains(buf.String(), "Nothing to remove") {
		t.Errorf("expected nothing-to-remove; got: %s", buf.String())
	}
	if _, err := os.Stat(filepath.Join(home, ".claude", "skills", "rpi-research", "SKILL.md")); err != nil {
		t.Errorf("rpi-research dir without marker was deleted: %v", err)
	}
	if _, err := os.Stat(custom); err != nil {
		t.Errorf("custom skill dir was deleted: %v", err)
	}
}
