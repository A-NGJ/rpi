package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteContractBlock_MissingFileNoOp(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "CLAUDE.md")

	buf := new(bytes.Buffer)
	if err := writeContractBlock(buf, path); err != nil {
		t.Fatalf("writeContractBlock returned error: %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("file should not exist after writer call on missing path, got err=%v", err)
	}
}

func TestWriteContractBlock_AppendWhenAbsent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "CLAUDE.md")
	prior := "# CLAUDE.md\n\nProject overview here.\n"
	if err := os.WriteFile(path, []byte(prior), 0644); err != nil {
		t.Fatal(err)
	}

	buf := new(bytes.Buffer)
	if err := writeContractBlock(buf, path); err != nil {
		t.Fatalf("writeContractBlock: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(got, []byte(prior)) {
		t.Errorf("prior content not preserved at file start.\nfile: %q", got)
	}
	if !bytes.Contains(got, []byte("<!-- rpi:contract:begin")) {
		t.Error("contract begin marker not appended")
	}
	if !bytes.Contains(got, []byte("<!-- rpi:contract:end -->")) {
		t.Error("contract end marker not appended")
	}
	// Exactly one blank line separator between prior content and block.
	beginIdx := bytes.Index(got, []byte("<!-- rpi:contract:begin"))
	if beginIdx < 2 {
		t.Fatalf("begin marker at unexpected index %d", beginIdx)
	}
	if string(got[beginIdx-2:beginIdx]) != "\n\n" {
		t.Errorf("expected exactly one blank line before block, got %q", got[beginIdx-3:beginIdx])
	}
}

func TestWriteContractBlock_ReplaceWhenPresent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "CLAUDE.md")
	stale := "# CLAUDE.md\n\nIntro paragraph.\n\n<!-- rpi:contract:begin v=1 -->\n## Stale Heading\n\nObsolete text.\n<!-- rpi:contract:end -->\n\n## After Section\n\nPost block.\n"
	if err := os.WriteFile(path, []byte(stale), 0644); err != nil {
		t.Fatal(err)
	}

	buf := new(bytes.Buffer)
	if err := writeContractBlock(buf, path); err != nil {
		t.Fatalf("writeContractBlock: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(got, []byte("Stale Heading")) {
		t.Error("stale block contents were not replaced")
	}
	if !bytes.Contains(got, []byte("## RPI Skill Contract")) {
		t.Error("fresh block heading missing after replace")
	}
	if !bytes.Contains(got, []byte("Intro paragraph.")) {
		t.Error("content before block was not preserved")
	}
	if !bytes.Contains(got, []byte("## After Section")) {
		t.Error("content after block was not preserved")
	}
	if !bytes.Contains(got, []byte("Post block.")) {
		t.Error("trailing user content was not preserved")
	}
}

func TestWriteContractBlock_IdempotentOnRerun(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "CLAUDE.md")
	if err := os.WriteFile(path, []byte("# CLAUDE.md\n\nIntro.\n"), 0644); err != nil {
		t.Fatal(err)
	}

	buf := new(bytes.Buffer)
	if err := writeContractBlock(buf, path); err != nil {
		t.Fatalf("first call: %v", err)
	}
	first, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	firstStat, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}

	// Second call must not rewrite the file (mtime preserved AND byte-identical).
	buf.Reset()
	if err := writeContractBlock(buf, path); err != nil {
		t.Fatalf("second call: %v", err)
	}
	second, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(first, second) {
		t.Error("second writeContractBlock call mutated file content")
	}
	secondStat, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if !secondStat.ModTime().Equal(firstStat.ModTime()) {
		t.Error("second writeContractBlock call rewrote file (mtime changed)")
	}
}

func TestWriteContractBlock_MalformedBeginOnly(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "CLAUDE.md")
	malformed := "# CLAUDE.md\n\n<!-- rpi:contract:begin v=1 -->\n## Skill Contract\n\nNo end marker here.\n"
	if err := os.WriteFile(path, []byte(malformed), 0644); err != nil {
		t.Fatal(err)
	}

	buf := new(bytes.Buffer)
	if err := writeContractBlock(buf, path); err != nil {
		t.Fatalf("writeContractBlock: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != malformed {
		t.Errorf("malformed file was modified.\nwant: %q\ngot:  %q", malformed, got)
	}
	if !strings.Contains(buf.String(), "malformed contract block") {
		t.Errorf("expected warning about malformed block, got: %q", buf.String())
	}
}

func TestWriteContractBlock_MalformedEndOnly(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "CLAUDE.md")
	malformed := "# CLAUDE.md\n\n## Skill Contract\n\nNo begin marker here.\n<!-- rpi:contract:end -->\n"
	if err := os.WriteFile(path, []byte(malformed), 0644); err != nil {
		t.Fatal(err)
	}

	buf := new(bytes.Buffer)
	if err := writeContractBlock(buf, path); err != nil {
		t.Fatalf("writeContractBlock: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != malformed {
		t.Errorf("malformed file was modified.\nwant: %q\ngot:  %q", malformed, got)
	}
	if !strings.Contains(buf.String(), "malformed contract block") {
		t.Errorf("expected warning about malformed block, got: %q", buf.String())
	}
}

func TestWriteContractBlock_PreservesOuterContent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "CLAUDE.md")
	before := "# CLAUDE.md\n\n## User Section Before\n\nUser line A.\nUser line B.\n\n"
	staleBlock := "<!-- rpi:contract:begin v=1 -->\n## Outdated\n\nObsolete.\n<!-- rpi:contract:end -->\n"
	after := "\n## User Section After\n\nUser line C.\nUser line D.\n"
	original := before + staleBlock + after
	if err := os.WriteFile(path, []byte(original), 0644); err != nil {
		t.Fatal(err)
	}

	buf := new(bytes.Buffer)
	if err := writeContractBlock(buf, path); err != nil {
		t.Fatalf("writeContractBlock: %v", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	// Outer content must be byte-identical: everything before begin marker
	// matches `before`, and the suffix from the line after end marker matches `after`.
	if !bytes.HasPrefix(got, []byte(before)) {
		t.Errorf("content before block was modified.\nwant prefix: %q\ngot:         %q", before, got)
	}
	if !bytes.HasSuffix(got, []byte(after)) {
		t.Errorf("content after block was modified.\nwant suffix: %q\ngot:         %q", after, got)
	}
}
