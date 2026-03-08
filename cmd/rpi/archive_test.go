package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/A-NGJ/ai-agent-research-plan-implement-flow/internal/scanner"
)

func setupArchiveTestDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	writeArchiveFile(t, dir, "plans/p1.md",
		"---\ntopic: \"Plan One\"\nstatus: complete\ndesign: designs/d1.md\n---\n# P1\n")
	writeArchiveFile(t, dir, "designs/d1.md",
		"---\ntopic: \"Design One\"\nstatus: draft\n---\n# D1\n")
	writeArchiveFile(t, dir, "research/r1.md",
		"---\ntopic: \"Research One\"\nstatus: superseded\n---\n# R1\nReferences plans/p1.md in body.\n")
	writeArchiveFile(t, dir, "tickets/t1.md",
		"---\ntopic: \"Ticket One\"\nstatus: active\n---\n# T1\n")

	return dir
}

func writeArchiveFile(t *testing.T, dir, relPath, content string) {
	t.Helper()
	full := filepath.Join(dir, relPath)
	if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestArchiveScanReturnsRefCount(t *testing.T) {
	dir := setupArchiveTestDir(t)
	oldFlag := thoughtsDirFlag
	thoughtsDirFlag = dir
	defer func() { thoughtsDirFlag = oldFlag }()

	buf := new(bytes.Buffer)
	cmd := archiveScanCmd
	cmd.SetOut(buf)

	if err := cmd.RunE(cmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var results []archiveScanResult
	if err := json.Unmarshal(buf.Bytes(), &results); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, buf.String())
	}

	// p1 (complete) and r1 (superseded) are archivable
	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}

	for _, r := range results {
		if r.Type == "plan" {
			// p1 is referenced by r1 in body
			if r.RefCount < 1 {
				t.Errorf("plan ref_count should be >= 1, got %d", r.RefCount)
			}
		}
	}
}

func TestArchiveCheckRefsReturnsDetails(t *testing.T) {
	dir := setupArchiveTestDir(t)
	oldFlag := thoughtsDirFlag
	thoughtsDirFlag = dir
	defer func() { thoughtsDirFlag = oldFlag }()

	buf := new(bytes.Buffer)
	cmd := archiveCheckRefsCmd
	cmd.SetOut(buf)

	if err := cmd.RunE(cmd, []string{"designs/d1.md"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var refs []scanner.ReferenceDetail
	if err := json.Unmarshal(buf.Bytes(), &refs); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, buf.String())
	}

	// p1 references designs/d1.md via frontmatter design field
	if len(refs) < 1 {
		t.Errorf("expected at least 1 reference, got %d", len(refs))
	}
}

func TestArchiveCheckRefsEmpty(t *testing.T) {
	dir := setupArchiveTestDir(t)
	oldFlag := thoughtsDirFlag
	thoughtsDirFlag = dir
	defer func() { thoughtsDirFlag = oldFlag }()

	buf := new(bytes.Buffer)
	cmd := archiveCheckRefsCmd
	cmd.SetOut(buf)

	if err := cmd.RunE(cmd, []string{"nonexistent/file.md"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var refs []scanner.ReferenceDetail
	if err := json.Unmarshal(buf.Bytes(), &refs); err != nil {
		t.Fatalf("invalid JSON: %v\noutput: %s", err, buf.String())
	}

	if len(refs) != 0 {
		t.Errorf("expected 0 refs, got %d", len(refs))
	}
}
