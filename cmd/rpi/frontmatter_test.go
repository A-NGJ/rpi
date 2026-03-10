package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/A-NGJ/ai-agent-research-plan-implement-flow/internal/frontmatter"
)

func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.md")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestFrontmatterGetAll(t *testing.T) {
	path := writeTempFile(t, "---\nstatus: draft\ntitle: Test Doc\n---\n# Body\n")

	doc, err := frontmatter.Parse(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if doc.Frontmatter["status"] != "draft" {
		t.Errorf("status = %v, want %q", doc.Frontmatter["status"], "draft")
	}
	if doc.Frontmatter["title"] != "Test Doc" {
		t.Errorf("title = %v, want %q", doc.Frontmatter["title"], "Test Doc")
	}
}

func TestFrontmatterGetField(t *testing.T) {
	path := writeTempFile(t, "---\nstatus: active\ntitle: Hello\n---\n# Body\n")

	doc, err := frontmatter.Parse(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val, ok := doc.Frontmatter["status"]
	if !ok {
		t.Fatal("status field not found")
	}
	if val != "active" {
		t.Errorf("status = %v, want %q", val, "active")
	}
}

func TestFrontmatterGetNoFrontmatter(t *testing.T) {
	path := writeTempFile(t, "# Just a heading\n\nNo frontmatter.\n")

	doc, err := frontmatter.Parse(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(doc.Frontmatter) != 0 {
		t.Errorf("expected empty frontmatter, got %v", doc.Frontmatter)
	}
}

func TestFrontmatterSet(t *testing.T) {
	path := writeTempFile(t, "---\nstatus: draft\n---\n# Body\n\nOriginal content.\n")

	doc, err := frontmatter.Parse(path)
	if err != nil {
		t.Fatal(err)
	}

	doc.Frontmatter["new_field"] = "new_value"
	if err := frontmatter.Write(doc); err != nil {
		t.Fatal(err)
	}

	// Re-parse and verify
	doc2, err := frontmatter.Parse(path)
	if err != nil {
		t.Fatal(err)
	}

	if doc2.Frontmatter["new_field"] != "new_value" {
		t.Errorf("new_field = %v, want %q", doc2.Frontmatter["new_field"], "new_value")
	}
	if doc2.Body != doc.Body {
		t.Errorf("body changed after set:\ngot:  %q\nwant: %q", doc2.Body, doc.Body)
	}
}

func TestFrontmatterTransitionValid(t *testing.T) {
	path := writeTempFile(t, "---\nstatus: draft\n---\n# Body\n")

	doc, err := frontmatter.Parse(path)
	if err != nil {
		t.Fatal(err)
	}

	if err := frontmatter.Transition(doc, "active"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := frontmatter.Write(doc); err != nil {
		t.Fatal(err)
	}

	// Re-parse and verify
	doc2, err := frontmatter.Parse(path)
	if err != nil {
		t.Fatal(err)
	}
	if doc2.Frontmatter["status"] != "active" {
		t.Errorf("status = %v, want %q", doc2.Frontmatter["status"], "active")
	}
}

func TestTransitionCascadesToTicket(t *testing.T) {
	dir := t.TempDir()
	thoughtsDir := filepath.Join(dir, ".thoughts")
	os.MkdirAll(filepath.Join(thoughtsDir, "plans"), 0755)
	os.MkdirAll(filepath.Join(thoughtsDir, "tickets"), 0755)

	// Create a ticket
	ticketPath := filepath.Join(thoughtsDir, "tickets", "feat-001-something.md")
	os.WriteFile(ticketPath, []byte("---\nticket_id: feat-001\nstatus: draft\ntitle: Something\n---\n# Ticket\n"), 0644)

	// Create a plan referencing the ticket
	planPath := filepath.Join(thoughtsDir, "plans", "2026-03-10-feat-001-something.md")
	os.WriteFile(planPath, []byte("---\nstatus: draft\nticket: feat-001\ntopic: Something\n---\n# Plan\n"), 0644)

	// Transition plan to active
	oldFlag := thoughtsDirFlag
	thoughtsDirFlag = thoughtsDir
	defer func() { thoughtsDirFlag = oldFlag }()

	doc, err := frontmatter.Parse(planPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := frontmatter.Transition(doc, "active"); err != nil {
		t.Fatal(err)
	}
	if err := frontmatter.Write(doc); err != nil {
		t.Fatal(err)
	}
	cascadeToTicket(doc, "active")

	// Verify ticket is now active
	ticketDoc, err := frontmatter.Parse(ticketPath)
	if err != nil {
		t.Fatal(err)
	}
	if ticketDoc.Frontmatter["status"] != "active" {
		t.Errorf("ticket status = %v, want active", ticketDoc.Frontmatter["status"])
	}
}

func TestTransitionCascadesToTicketComplete(t *testing.T) {
	dir := t.TempDir()
	thoughtsDir := filepath.Join(dir, ".thoughts")
	os.MkdirAll(filepath.Join(thoughtsDir, "plans"), 0755)
	os.MkdirAll(filepath.Join(thoughtsDir, "tickets"), 0755)

	ticketPath := filepath.Join(thoughtsDir, "tickets", "feat-002-other.md")
	os.WriteFile(ticketPath, []byte("---\nticket_id: feat-002\nstatus: active\ntitle: Other\n---\n# Ticket\n"), 0644)

	planPath := filepath.Join(thoughtsDir, "plans", "2026-03-10-feat-002-other.md")
	os.WriteFile(planPath, []byte("---\nstatus: active\nticket: feat-002\ntopic: Other\n---\n# Plan\n"), 0644)

	oldFlag := thoughtsDirFlag
	thoughtsDirFlag = thoughtsDir
	defer func() { thoughtsDirFlag = oldFlag }()

	doc, err := frontmatter.Parse(planPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := frontmatter.Transition(doc, "complete"); err != nil {
		t.Fatal(err)
	}
	if err := frontmatter.Write(doc); err != nil {
		t.Fatal(err)
	}
	cascadeToTicket(doc, "complete")

	ticketDoc, err := frontmatter.Parse(ticketPath)
	if err != nil {
		t.Fatal(err)
	}
	if ticketDoc.Frontmatter["status"] != "complete" {
		t.Errorf("ticket status = %v, want complete", ticketDoc.Frontmatter["status"])
	}
}

func TestTransitionNoCascadeWithoutTicketField(t *testing.T) {
	dir := t.TempDir()
	thoughtsDir := filepath.Join(dir, ".thoughts")
	os.MkdirAll(filepath.Join(thoughtsDir, "plans"), 0755)

	planPath := filepath.Join(thoughtsDir, "plans", "2026-03-10-standalone.md")
	os.WriteFile(planPath, []byte("---\nstatus: draft\ntopic: Standalone\n---\n# Plan\n"), 0644)

	oldFlag := thoughtsDirFlag
	thoughtsDirFlag = thoughtsDir
	defer func() { thoughtsDirFlag = oldFlag }()

	doc, err := frontmatter.Parse(planPath)
	if err != nil {
		t.Fatal(err)
	}
	// Should not panic or error — just silently skip
	cascadeToTicket(doc, "active")
}

func TestTransitionNoCascadeTicketNotFound(t *testing.T) {
	dir := t.TempDir()
	thoughtsDir := filepath.Join(dir, ".thoughts")
	os.MkdirAll(filepath.Join(thoughtsDir, "plans"), 0755)
	os.MkdirAll(filepath.Join(thoughtsDir, "tickets"), 0755)

	planPath := filepath.Join(thoughtsDir, "plans", "2026-03-10-missing.md")
	os.WriteFile(planPath, []byte("---\nstatus: draft\nticket: nonexistent-001\ntopic: Missing\n---\n# Plan\n"), 0644)

	oldFlag := thoughtsDirFlag
	thoughtsDirFlag = thoughtsDir
	defer func() { thoughtsDirFlag = oldFlag }()

	doc, err := frontmatter.Parse(planPath)
	if err != nil {
		t.Fatal(err)
	}
	// Should not panic or error
	cascadeToTicket(doc, "active")
}

func TestTransitionNoCascadeTicketAlreadyAtStatus(t *testing.T) {
	dir := t.TempDir()
	thoughtsDir := filepath.Join(dir, ".thoughts")
	os.MkdirAll(filepath.Join(thoughtsDir, "plans"), 0755)
	os.MkdirAll(filepath.Join(thoughtsDir, "tickets"), 0755)

	// Ticket already active
	ticketPath := filepath.Join(thoughtsDir, "tickets", "feat-003-already.md")
	os.WriteFile(ticketPath, []byte("---\nticket_id: feat-003\nstatus: active\ntitle: Already\n---\n# Ticket\n"), 0644)

	planPath := filepath.Join(thoughtsDir, "plans", "2026-03-10-feat-003-already.md")
	os.WriteFile(planPath, []byte("---\nstatus: draft\nticket: feat-003\ntopic: Already\n---\n# Plan\n"), 0644)

	oldFlag := thoughtsDirFlag
	thoughtsDirFlag = thoughtsDir
	defer func() { thoughtsDirFlag = oldFlag }()

	doc, err := frontmatter.Parse(planPath)
	if err != nil {
		t.Fatal(err)
	}
	// Cascade active to already-active ticket — should silently skip (invalid transition active→active)
	cascadeToTicket(doc, "active")

	// Verify ticket status unchanged
	ticketDoc, err := frontmatter.Parse(ticketPath)
	if err != nil {
		t.Fatal(err)
	}
	if ticketDoc.Frontmatter["status"] != "active" {
		t.Errorf("ticket status = %v, want active (unchanged)", ticketDoc.Frontmatter["status"])
	}
}

func TestFrontmatterTransitionInvalidExitCode(t *testing.T) {
	// Build the binary first
	binPath := filepath.Join(t.TempDir(), "rpi")
	build := exec.Command("go", "build", "-o", binPath, "./")
	build.Dir = filepath.Join(".", ".")
	// We need to find the project root for the build
	// Since we're in cmd/rpi/, we build from here
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build failed: %v\n%s", err, out)
	}

	tmpFile := writeTempFile(t, "---\nstatus: draft\n---\n# Body\n")

	cmd := exec.Command(binPath, "frontmatter", "transition", tmpFile, "archived")
	out, err := cmd.CombinedOutput()
	if err == nil {
		t.Fatal("expected non-zero exit code for invalid transition")
	}

	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("expected ExitError, got %T", err)
	}
	if exitErr.ExitCode() != 2 {
		t.Errorf("exit code = %d, want 2\noutput: %s", exitErr.ExitCode(), out)
	}

	if !strings.Contains(string(out), "invalid status transition") {
		t.Errorf("stderr should contain 'invalid status transition', got: %s", out)
	}
}
