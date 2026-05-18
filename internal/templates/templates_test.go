package templates

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGet_AGENTS(t *testing.T) {
	content, err := Get("AGENTS.md")
	if err != nil {
		t.Fatalf("Get(AGENTS.md) returned error: %v", err)
	}
	if content == "" {
		t.Fatal("Get(AGENTS.md) returned empty content")
	}
	if !strings.Contains(content, "# AGENTS.md") {
		t.Error("AGENTS.md template missing expected '# AGENTS.md' header")
	}
	if !strings.Contains(content, ".rpi/") {
		t.Error("AGENTS.md template missing .rpi/ directory reference")
	}
}

func TestGet_CLAUDE(t *testing.T) {
	content, err := Get("CLAUDE.md")
	if err != nil {
		t.Fatalf("Get(CLAUDE.md) returned error: %v", err)
	}
	if content == "" {
		t.Fatal("Get(CLAUDE.md) returned empty content")
	}
	if !strings.Contains(content, "# CLAUDE.md") {
		t.Error("CLAUDE.md template missing expected '# CLAUDE.md' header")
	}
}

func TestGet_Unknown(t *testing.T) {
	_, err := Get("nonexistent")
	if err == nil {
		t.Fatal("Get(nonexistent) should return error")
	}
	if !strings.Contains(err.Error(), "unknown template") {
		t.Errorf("error should mention 'unknown template', got: %v", err)
	}
}

func TestGet_MatchesWorkflowAsset(t *testing.T) {
	for name := range map[string]string{
		"AGENTS.md": "templates/AGENTS.md.template",
		"CLAUDE.md": "templates/CLAUDE.md.template",
	} {
		t.Run(name, func(t *testing.T) {
			setupTestHome(t) // ensure no user overrides

			got, err := Get(name)
			if err != nil {
				t.Fatalf("Get(%s): %v", name, err)
			}
			if got == "" {
				t.Fatalf("Get(%s) returned empty content", name)
			}
		})
	}
}

func setupTestHome(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)
	return tmpDir
}

func TestEnsureUserTemplates(t *testing.T) {
	home := setupTestHome(t)

	err := EnsureUserTemplates()
	if err != nil {
		t.Fatalf("EnsureUserTemplates() returned error: %v", err)
	}

	templatesDir := filepath.Join(home, ".rpi", "templates")
	for _, name := range Names() {
		path := filepath.Join(templatesDir, name+".template")
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("expected file %s to exist: %v", path, err)
		}
		if len(data) == 0 {
			t.Errorf("file %s is empty", path)
		}
	}
}

func TestEnsureUserTemplates_NoOverwrite(t *testing.T) {
	home := setupTestHome(t)

	templatesDir := filepath.Join(home, ".rpi", "templates")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		t.Fatal(err)
	}

	customContent := "# My Custom CLAUDE.md"
	customPath := filepath.Join(templatesDir, "CLAUDE.md.template")
	if err := os.WriteFile(customPath, []byte(customContent), 0644); err != nil {
		t.Fatal(err)
	}

	err := EnsureUserTemplates()
	if err != nil {
		t.Fatalf("EnsureUserTemplates() returned error: %v", err)
	}

	data, err := os.ReadFile(customPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != customContent {
		t.Errorf("EnsureUserTemplates overwrote custom content: got %q, want %q", string(data), customContent)
	}
}

func TestGet_PrefersUserTemplate(t *testing.T) {
	home := setupTestHome(t)

	templatesDir := filepath.Join(home, ".rpi", "templates")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		t.Fatal(err)
	}

	customContent := "# Customized CLAUDE.md"
	if err := os.WriteFile(filepath.Join(templatesDir, "CLAUDE.md.template"), []byte(customContent), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := Get("CLAUDE.md")
	if err != nil {
		t.Fatalf("Get(CLAUDE.md) returned error: %v", err)
	}
	if got != customContent {
		t.Errorf("Get(CLAUDE.md) = %q, want user-customized %q", got, customContent)
	}
}

func TestGet_FallsBackToEmbedded(t *testing.T) {
	setupTestHome(t) // empty home, no ~/.rpi/templates/

	content, err := Get("CLAUDE.md")
	if err != nil {
		t.Fatalf("Get(CLAUDE.md) returned error: %v", err)
	}
	if !strings.Contains(content, "# CLAUDE.md") {
		t.Error("expected embedded content with '# CLAUDE.md' header")
	}
}

func TestGetCLAUDEMD_ContainsContractBlock(t *testing.T) {
	setupTestHome(t)

	content, err := Get("CLAUDE.md")
	if err != nil {
		t.Fatalf("Get(CLAUDE.md): %v", err)
	}
	if !strings.Contains(content, "<!-- rpi:contract:begin") {
		t.Error("CLAUDE.md missing contract begin marker")
	}
	if !strings.Contains(content, "<!-- rpi:contract:end -->") {
		t.Error("CLAUDE.md missing contract end marker")
	}
	if !strings.Contains(content, "## RPI Skill Contract") {
		t.Error("CLAUDE.md missing '## RPI Skill Contract' heading")
	}
}

func TestGetAGENTSMD_ContainsContractBlock(t *testing.T) {
	setupTestHome(t)

	content, err := Get("AGENTS.md")
	if err != nil {
		t.Fatalf("Get(AGENTS.md): %v", err)
	}
	if !strings.Contains(content, "<!-- rpi:contract:begin") {
		t.Error("AGENTS.md missing contract begin marker")
	}
	if !strings.Contains(content, "<!-- rpi:contract:end -->") {
		t.Error("AGENTS.md missing contract end marker")
	}
	if !strings.Contains(content, "## RPI Skill Contract") {
		t.Error("AGENTS.md missing '## RPI Skill Contract' heading")
	}
}

func TestGetReplacesSlotOnce(t *testing.T) {
	setupTestHome(t)

	for _, name := range []string{"CLAUDE.md", "AGENTS.md"} {
		t.Run(name, func(t *testing.T) {
			content, err := Get(name)
			if err != nil {
				t.Fatalf("Get(%s): %v", name, err)
			}
			if strings.Contains(content, "<!-- rpi:contract:slot -->") {
				t.Error("placeholder slot still present after splice")
			}
			if got := strings.Count(content, "<!-- rpi:contract:begin"); got != 1 {
				t.Errorf("begin marker appears %d times, want 1", got)
			}
			if got := strings.Count(content, "<!-- rpi:contract:end -->"); got != 1 {
				t.Errorf("end marker appears %d times, want 1", got)
			}
		})
	}
}

func TestGetWithUserTemplateMissingSlot_NoSplice(t *testing.T) {
	home := setupTestHome(t)

	templatesDir := filepath.Join(home, ".rpi", "templates")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Legacy user template — no slot placeholder.
	legacy := "# Custom CLAUDE.md\n\nNo slot here.\n"
	if err := os.WriteFile(filepath.Join(templatesDir, "CLAUDE.md.template"), []byte(legacy), 0644); err != nil {
		t.Fatal(err)
	}

	got, err := Get("CLAUDE.md")
	if err != nil {
		t.Fatalf("Get(CLAUDE.md): %v", err)
	}
	if got != legacy {
		t.Errorf("Get(CLAUDE.md) = %q, want verbatim %q", got, legacy)
	}
	if strings.Contains(got, "<!-- rpi:contract:begin") {
		t.Error("contract block was spliced into legacy template without slot placeholder")
	}
}

func TestNames(t *testing.T) {
	names := Names()
	if len(names) != 2 {
		t.Fatalf("expected 2 names, got %d: %v", len(names), names)
	}

	expected := []string{"AGENTS.md", "CLAUDE.md"}
	for i, name := range expected {
		if names[i] != name {
			t.Errorf("names[%d] = %q, want %q", i, names[i], name)
		}
	}
}
