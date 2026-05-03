package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestParseStatus(t *testing.T) {
	output := `M  staged.go
 M modified.go
MM both.go
A  added.go
?? untracked.go`

	info := ParseStatus(output)

	if len(info.Staged) != 3 {
		t.Errorf("expected 3 staged files, got %d: %v", len(info.Staged), info.Staged)
	}
	if len(info.Modified) != 2 {
		t.Errorf("expected 2 modified files, got %d: %v", len(info.Modified), info.Modified)
	}
	if len(info.Untracked) != 1 {
		t.Errorf("expected 1 untracked file, got %d: %v", len(info.Untracked), info.Untracked)
	}
	if info.Untracked[0] != "untracked.go" {
		t.Errorf("expected untracked.go, got %s", info.Untracked[0])
	}
}

func TestParseStatusEmpty(t *testing.T) {
	info := ParseStatus("")
	if len(info.Staged) != 0 || len(info.Modified) != 0 || len(info.Untracked) != 0 {
		t.Errorf("expected empty arrays, got staged=%v modified=%v untracked=%v",
			info.Staged, info.Modified, info.Untracked)
	}
}

func TestParseDiffStat(t *testing.T) {
	output := ` file1.go | 10 ++++------
 file2.go |  5 +++++
 3 files changed, 12 insertions(+), 6 deletions(-)`

	ds := ParseDiffStat(output)
	if ds.FilesChanged != 3 {
		t.Errorf("expected 3 files changed, got %d", ds.FilesChanged)
	}
	if ds.Insertions != 12 {
		t.Errorf("expected 12 insertions, got %d", ds.Insertions)
	}
	if ds.Deletions != 6 {
		t.Errorf("expected 6 deletions, got %d", ds.Deletions)
	}
}

func TestParseDiffStatNoChanges(t *testing.T) {
	ds := ParseDiffStat("")
	if ds.FilesChanged != 0 || ds.Insertions != 0 || ds.Deletions != 0 {
		t.Errorf("expected zeroes, got %+v", ds)
	}
}

func TestParseLog(t *testing.T) {
	output := `abc1234 feat: add something
def5678 fix: broken test
ghi9012 docs: update readme`

	commits := ParseLog(output)
	if len(commits) != 3 {
		t.Fatalf("expected 3 commits, got %d", len(commits))
	}
	if commits[0].Hash != "abc1234" {
		t.Errorf("expected hash abc1234, got %s", commits[0].Hash)
	}
	if commits[0].Message != "feat: add something" {
		t.Errorf("expected 'feat: add something', got %s", commits[0].Message)
	}
}

func TestSensitiveFilenames(t *testing.T) {
	files := []string{".env", "config/credentials.json", "src/main.go", "certs/server.pem", "tls/key.key"}
	matches := SensitiveFilenames(files)

	if len(matches) != 4 {
		t.Fatalf("expected 4 matches, got %d: %v", len(matches), matches)
	}

	found := map[string]bool{}
	for _, m := range matches {
		found[m.File] = true
	}
	for _, expected := range []string{".env", "config/credentials.json", "certs/server.pem", "tls/key.key"} {
		if !found[expected] {
			t.Errorf("expected %s to be flagged", expected)
		}
	}
}

func TestSensitiveContent(t *testing.T) {
	content := `DB_HOST=localhost
password=supersecret
API_KEY=abc123
-----BEGIN RSA PRIVATE KEY-----`

	matches := SensitiveContent("config.txt", content)
	if len(matches) != 3 {
		t.Fatalf("expected 3 matches, got %d: %v", len(matches), matches)
	}
}

func TestSensitiveNoMatches(t *testing.T) {
	matches := SensitiveFilenames([]string{"main.go", "utils.go", "README.md"})
	if len(matches) != 0 {
		t.Errorf("expected 0 matches, got %d: %v", len(matches), matches)
	}

	contentMatches := SensitiveContent("main.go", "package main\n\nfunc main() {}\n")
	if len(contentMatches) != 0 {
		t.Errorf("expected 0 content matches, got %d: %v", len(contentMatches), contentMatches)
	}
}

// setupTempRepo creates a temp git repo, chdirs into it for the test, and
// returns its path. Caller can use the returned path for absolute file ops;
// relative ops (git add, etc.) use the chdir.
func setupTempRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Chdir(dir)

	for _, args := range [][]string{
		{"init", "-q", "-b", "main"},
		{"config", "user.email", "test@example.com"},
		{"config", "user.name", "Test"},
		{"config", "commit.gpgsign", "false"},
	} {
		if err := exec.Command("git", args...).Run(); err != nil {
			t.Fatalf("git %v: %v", args, err)
		}
	}
	return dir
}

func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

func gitCmd(t *testing.T, args ...string) {
	t.Helper()
	if out, err := exec.Command("git", args...).CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func TestGitignoreCheck_NoMatches(t *testing.T) {
	dir := setupTempRepo(t)
	writeFile(t, dir, "main.go", "package main\n")
	gitCmd(t, "add", "main.go")

	matches, err := GitignoreCheck()
	if err != nil {
		t.Fatalf("GitignoreCheck: %v", err)
	}
	if len(matches) != 0 {
		t.Errorf("expected 0 matches, got %d: %v", len(matches), matches)
	}
}

func TestGitignoreCheck_EmptyRepo(t *testing.T) {
	setupTempRepo(t)

	matches, err := GitignoreCheck()
	if err != nil {
		t.Fatalf("GitignoreCheck: %v", err)
	}
	if len(matches) != 0 {
		t.Errorf("expected 0 matches in clean repo, got %d: %v", len(matches), matches)
	}
}

func TestGitignoreCheck_TrackedButNowIgnored(t *testing.T) {
	dir := setupTempRepo(t)
	// Commit foo.log first (no .gitignore yet)
	writeFile(t, dir, "foo.log", "line1\n")
	gitCmd(t, "add", "foo.log")
	gitCmd(t, "commit", "-q", "-m", "add foo.log")

	// Add .gitignore matching *.log, then modify foo.log so it shows in status
	writeFile(t, dir, ".gitignore", "*.log\n")
	writeFile(t, dir, "foo.log", "line1\nline2\n")

	matches, err := GitignoreCheck()
	if err != nil {
		t.Fatalf("GitignoreCheck: %v", err)
	}
	found := false
	for _, m := range matches {
		if m.File == "foo.log" {
			found = true
			if m.Pattern != "*.log" {
				t.Errorf("expected pattern '*.log', got %q", m.Pattern)
			}
			if m.Source == "" {
				t.Error("expected non-empty Source (gitignore file path)")
			}
		}
	}
	if !found {
		t.Errorf("expected foo.log to be flagged as ignored, got matches: %v", matches)
	}
}

func TestGitignoreCheck_UntrackedAndStaged(t *testing.T) {
	dir := setupTempRepo(t)
	writeFile(t, dir, ".gitignore", "*.tmp\n")
	gitCmd(t, "add", ".gitignore")
	gitCmd(t, "commit", "-q", "-m", "add gitignore")

	// Untracked file matching the ignore pattern, force-added so it appears in status
	writeFile(t, dir, "bar.tmp", "scratch\n")
	gitCmd(t, "add", "-f", "bar.tmp")

	matches, err := GitignoreCheck()
	if err != nil {
		t.Fatalf("GitignoreCheck: %v", err)
	}
	found := false
	for _, m := range matches {
		if m.File == "bar.tmp" && m.Pattern == "*.tmp" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected bar.tmp to be flagged with pattern *.tmp, got matches: %v", matches)
	}
}
