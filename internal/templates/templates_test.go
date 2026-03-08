package templates

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot determine test file path")
	}
	// internal/templates/templates_test.go -> repo root is two levels up
	return filepath.Join(filepath.Dir(filename), "..", "..")
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

func TestGet_PIPELINE(t *testing.T) {
	content, err := Get("PIPELINE.md")
	if err != nil {
		t.Fatalf("Get(PIPELINE.md) returned error: %v", err)
	}
	if content == "" {
		t.Fatal("Get(PIPELINE.md) returned empty content")
	}
	if !strings.Contains(content, "# Development Pipeline") {
		t.Error("PIPELINE.md template missing expected '# Development Pipeline' header")
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

func TestGet_MatchesSource(t *testing.T) {
	root := repoRoot(t)
	sources := map[string]string{
		"CLAUDE.md":   filepath.Join(root, "bin", "templates", "CLAUDE.md.template"),
		"PIPELINE.md": filepath.Join(root, "bin", "templates", "PIPELINE.md.template"),
	}

	for name, sourcePath := range sources {
		t.Run(name, func(t *testing.T) {
			expected, err := os.ReadFile(sourcePath)
			if err != nil {
				t.Fatalf("read source %s: %v", sourcePath, err)
			}

			got, err := Get(name)
			if err != nil {
				t.Fatalf("Get(%s): %v", name, err)
			}

			if got != string(expected) {
				t.Errorf("embedded %s does not match source file %s", name, sourcePath)
			}
		})
	}
}

func TestNames(t *testing.T) {
	names := Names()
	if len(names) != 2 {
		t.Fatalf("expected 2 names, got %d: %v", len(names), names)
	}

	expected := []string{"CLAUDE.md", "PIPELINE.md"}
	for i, name := range expected {
		if names[i] != name {
			t.Errorf("names[%d] = %q, want %q", i, names[i], name)
		}
	}
}
