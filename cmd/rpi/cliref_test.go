package main

import (
	"strings"
	"testing"
)

func TestGenerateCLIReferenceContainsAllCommands(t *testing.T) {
	out := generateCLIReference(rootCmd)

	expectedCommands := []string{
		"rpi scan",
		"rpi scaffold",
		"rpi chain",
		"rpi verify",
		"rpi frontmatter",
		"rpi extract",
		"rpi git-context",
		"rpi index build",
		"rpi index query",
		"rpi index files",
		"rpi index status",
		"rpi archive scan",
		"rpi archive check-refs",
		"rpi archive move",
		"rpi spec coverage",
		"rpi init",
	}

	for _, cmd := range expectedCommands {
		if !strings.Contains(out, cmd) {
			t.Errorf("output missing command %q", cmd)
		}
	}
}

func TestGenerateCLIReferenceExcludesInternalCommands(t *testing.T) {
	out := generateCLIReference(rootCmd)

	for _, cmd := range []string{"rpi completion", "rpi help"} {
		if strings.Contains(out, "| `"+cmd+"`") {
			t.Errorf("output should not contain internal command %q", cmd)
		}
	}
}

func TestGenerateCLIReferenceIncludesFlags(t *testing.T) {
	out := generateCLIReference(rootCmd)

	// The scan command should have --type, --status, --format flags
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		if strings.Contains(line, "| `rpi scan`") {
			if !strings.Contains(line, "`--type`") {
				t.Error("scan row missing --type flag")
			}
			if !strings.Contains(line, "`--status`") {
				t.Error("scan row missing --status flag")
			}
			if !strings.Contains(line, "`--format`") {
				t.Error("scan row missing --format flag")
			}
			return
		}
	}
	t.Error("could not find rpi scan row in output")
}

func TestGenerateCLIReferenceIncludesStatuses(t *testing.T) {
	out := generateCLIReference(rootCmd)

	expectedStatuses := []string{
		"draft",
		"active",
		"approved",
		"implemented",
		"complete",
		"archived",
		"superseded",
	}

	for _, s := range expectedStatuses {
		if !strings.Contains(out, s) {
			t.Errorf("output missing status %q", s)
		}
	}
}
