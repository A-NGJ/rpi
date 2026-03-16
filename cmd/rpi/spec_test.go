package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestSpecCoverageBasic(t *testing.T) {
	binary := buildBinary(t)
	dir := t.TempDir()

	// Create a spec file
	specDir := filepath.Join(dir, ".rpi", "specs")
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatal(err)
	}
	specPath := filepath.Join(specDir, "test.md")
	if err := os.WriteFile(specPath, []byte(`---
domain: test-module
id: TM
status: approved
---

# test-module

## Behavior
- **TM-1**: First behavior
- **TM-2**: Second behavior
`), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a test file with a matching spec comment
	pkgDir := filepath.Join(dir, "pkg")
	if err := os.MkdirAll(pkgDir, 0755); err != nil {
		t.Fatal(err)
	}
	// Need go.mod so findProjectRoot works
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test\n\ngo 1.21\n"), 0644); err != nil {
		t.Fatal(err)
	}
	testFile := filepath.Join(pkgDir, "thing_test.go")
	if err := os.WriteFile(testFile, []byte("package pkg\n\n// spec:TM-1\nfunc TestThing(t *testing.T) {}\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Run from the temp dir so findProjectRoot finds our go.mod
	stdout, _, exitCode := runRPIInDir(t, binary, dir, "spec", "coverage", specPath)

	// Exit code 1 because TM-2 is missing
	if exitCode != 1 {
		t.Errorf("expected exit 1 (missing coverage), got %d", exitCode)
	}
	if !strings.Contains(stdout, "TM-1") {
		t.Error("output should contain TM-1")
	}
	if !strings.Contains(stdout, "1 covered") {
		t.Errorf("output should show 1 covered, got:\n%s", stdout)
	}
	if !strings.Contains(stdout, "1 missing") {
		t.Errorf("output should show 1 missing, got:\n%s", stdout)
	}
}

func TestSpecCoverageMissing(t *testing.T) {
	binary := buildBinary(t)
	dir := t.TempDir()

	specDir := filepath.Join(dir, ".rpi", "specs")
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatal(err)
	}
	specPath := filepath.Join(specDir, "test.md")
	if err := os.WriteFile(specPath, []byte(`---
domain: test-module
id: TM
status: approved
---

# test-module

## Behavior
- **TM-1**: First behavior
`), 0644); err != nil {
		t.Fatal(err)
	}

	// go.mod but no test files with spec comments
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test\n\ngo 1.21\n"), 0644); err != nil {
		t.Fatal(err)
	}

	_, _, exitCode := runRPIInDir(t, binary, dir, "spec", "coverage", specPath)

	if exitCode != 1 {
		t.Errorf("expected exit 1 (all missing), got %d", exitCode)
	}
}

func TestSpecCoverageJSON(t *testing.T) {
	binary := buildBinary(t)
	dir := t.TempDir()

	specDir := filepath.Join(dir, ".rpi", "specs")
	if err := os.MkdirAll(specDir, 0755); err != nil {
		t.Fatal(err)
	}
	specPath := filepath.Join(specDir, "test.md")
	if err := os.WriteFile(specPath, []byte(`---
domain: test-module
id: TM
status: approved
---

# test-module

## Behavior
- **TM-1**: First behavior
`), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test\n\ngo 1.21\n"), 0644); err != nil {
		t.Fatal(err)
	}

	stdout, _, exitCode := runRPIInDir(t, binary, dir, "spec", "coverage", "--format", "json", specPath)

	// Exit code 1 because behavior is uncovered
	if exitCode != 1 {
		t.Errorf("expected exit 1, got %d", exitCode)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(stdout), &result); err != nil {
		t.Fatalf("invalid JSON output: %v\nGot:\n%s", err, stdout)
	}

	if result["SpecDomain"] != "test-module" {
		t.Errorf("SpecDomain = %v, want test-module", result["SpecDomain"])
	}
}

func TestSpecCoverageMissingFile(t *testing.T) {
	binary := buildBinary(t)

	_, _, exitCode := runRPI(t, binary, "spec", "coverage", "/nonexistent/spec.md")

	if exitCode != 2 {
		t.Errorf("expected exit 2, got %d", exitCode)
	}
}

// runRPIInDir runs the rpi binary with a specific working directory.
func runRPIInDir(t *testing.T, binary, dir string, args ...string) (stdout, stderr string, exitCode int) {
	t.Helper()

	cmd := exec.Command(binary, args...)
	cmd.Dir = dir

	var outBuf, errBuf strings.Builder
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()
	exitCode = 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("unexpected error running rpi: %v", err)
		}
	}
	return outBuf.String(), errBuf.String(), exitCode
}
