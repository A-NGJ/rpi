package git

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type Context struct {
	Branch         string       `json:"branch"`
	Commit         string       `json:"commit"`
	Status         StatusInfo   `json:"status"`
	RecentCommits  []CommitInfo `json:"recent_commits"`
	DiffSummary    DiffSummary  `json:"diff_summary"`
	SensitiveFiles []string     `json:"sensitive_files"`
}

type StatusInfo struct {
	Staged    []string `json:"staged"`
	Modified  []string `json:"modified"`
	Untracked []string `json:"untracked"`
}

type CommitInfo struct {
	Hash    string `json:"hash"`
	Message string `json:"message"`
}

type DiffSummary struct {
	FilesChanged int `json:"files_changed"`
	Insertions   int `json:"insertions"`
	Deletions    int `json:"deletions"`
}

type SensitiveMatch struct {
	File   string `json:"file"`
	Reason string `json:"reason"`
}

type GitignoreMatch struct {
	File    string `json:"file"`
	Pattern string `json:"pattern"`
	Source  string `json:"source"`
}

var sensitiveFilePatterns = []struct {
	pattern *regexp.Regexp
	reason  string
}{
	{regexp.MustCompile(`(?i)\.env$`), "matches .env pattern"},
	{regexp.MustCompile(`(?i)\.env\.`), "matches .env pattern"},
	{regexp.MustCompile(`(?i)credentials`), "matches credentials pattern"},
	{regexp.MustCompile(`(?i)secret`), "matches secret pattern"},
	{regexp.MustCompile(`(?i)api_key`), "matches api_key pattern"},
	{regexp.MustCompile(`(?i)private_key`), "matches private_key pattern"},
	{regexp.MustCompile(`(?i)\.pem$`), "matches .pem pattern"},
	{regexp.MustCompile(`(?i)\.key$`), "matches .key pattern"},
}

var sensitiveContentPatterns = []struct {
	pattern *regexp.Regexp
	reason  string
}{
	{regexp.MustCompile(`password\s*=`), "contains password assignment"},
	{regexp.MustCompile(`API_KEY\s*=`), "contains API_KEY assignment"},
	{regexp.MustCompile(`BEGIN RSA PRIVATE KEY`), "contains RSA private key"},
}

func runGit(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	return strings.TrimRight(string(out), "\n"), nil
}

func GatherContext() (*Context, error) {
	ctx := &Context{}

	branch, err := runGit("branch", "--show-current")
	if err != nil {
		ctx.Branch = "HEAD"
	} else if branch == "" {
		ctx.Branch = "HEAD"
	} else {
		ctx.Branch = branch
	}

	commit, err := runGit("rev-parse", "--short", "HEAD")
	if err != nil {
		ctx.Commit = ""
	} else {
		ctx.Commit = commit
	}

	statusOut, err := runGit("status", "--porcelain")
	if err != nil {
		ctx.Status = StatusInfo{Staged: []string{}, Modified: []string{}, Untracked: []string{}}
	} else {
		ctx.Status = ParseStatus(statusOut)
	}

	logOut, err := runGit("log", "--oneline", "-10")
	if err != nil {
		ctx.RecentCommits = []CommitInfo{}
	} else {
		ctx.RecentCommits = ParseLog(logOut)
	}

	diffOut, err := runGit("diff", "--stat", "HEAD")
	if err != nil {
		ctx.DiffSummary = DiffSummary{}
	} else {
		ctx.DiffSummary = ParseDiffStat(diffOut)
	}

	sensitive, err := SensitiveCheck()
	if err != nil {
		ctx.SensitiveFiles = []string{}
	} else {
		files := make([]string, len(sensitive))
		for i, m := range sensitive {
			files[i] = m.File
		}
		ctx.SensitiveFiles = files
	}

	return ctx, nil
}

func ChangedFiles() ([]string, error) {
	out, err := runGit("diff", "--name-only", "main...HEAD")
	if err != nil || out == "" {
		out, err = runGit("diff", "--name-only", "HEAD~10")
		if err != nil {
			return []string{}, nil
		}
	}
	if out == "" {
		return []string{}, nil
	}
	return strings.Split(out, "\n"), nil
}

// GitignoreCheck scans files appearing in `git status` (staged, modified,
// untracked) against gitignore rules and returns any matches. Uses
// `git check-ignore --no-index` so already-tracked-but-now-ignored files are
// reported (the realistic case: a file was tracked, then later added to
// .gitignore, and is now showing up as a candidate to commit).
func GitignoreCheck() ([]GitignoreMatch, error) {
	statusOut, err := runGit("status", "--porcelain")
	if err != nil {
		return []GitignoreMatch{}, nil
	}
	info := ParseStatus(statusOut)

	seen := map[string]struct{}{}
	var files []string
	for _, group := range [][]string{info.Staged, info.Modified, info.Untracked} {
		for _, f := range group {
			if _, ok := seen[f]; ok {
				continue
			}
			seen[f] = struct{}{}
			files = append(files, f)
		}
	}
	if len(files) == 0 {
		return []GitignoreMatch{}, nil
	}

	matches := []GitignoreMatch{}
	for _, file := range files {
		m, ok, err := checkIgnoreOne(file)
		if err != nil {
			return nil, err
		}
		if ok {
			matches = append(matches, m)
		}
	}
	return matches, nil
}

// checkIgnoreOne runs `git check-ignore --no-index --verbose -- <file>` and
// returns the parsed match. `git check-ignore` exits 0 when a file is ignored,
// 1 when it is not, and other codes on real errors.
func checkIgnoreOne(file string) (GitignoreMatch, bool, error) {
	cmd := exec.Command("git", "check-ignore", "--no-index", "--verbose", "--", file)
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 1 {
				return GitignoreMatch{}, false, nil
			}
			return GitignoreMatch{}, false, fmt.Errorf("git check-ignore %s: exit %d: %s",
				file, exitErr.ExitCode(), strings.TrimSpace(string(exitErr.Stderr)))
		}
		return GitignoreMatch{}, false, fmt.Errorf("git check-ignore %s: %w", file, err)
	}

	line := strings.TrimRight(string(out), "\n")
	return parseCheckIgnoreLine(line)
}

// parseCheckIgnoreLine parses a `git check-ignore --verbose` output line in the
// form `<source>:<lineno>:<pattern>\t<file>`. The source may itself be empty
// (e.g. when no gitignore source is recorded) and the file path may contain
// spaces, so we split on the tab boundary first.
func parseCheckIgnoreLine(line string) (GitignoreMatch, bool, error) {
	tab := strings.IndexByte(line, '\t')
	if tab < 0 {
		return GitignoreMatch{}, false, fmt.Errorf("unexpected check-ignore output: %q", line)
	}
	meta := line[:tab]
	file := line[tab+1:]

	// meta is "<source>:<lineno>:<pattern>". Split from the right twice so a
	// colon in the source path (e.g. Windows) doesn't corrupt the pattern.
	lastColon := strings.LastIndexByte(meta, ':')
	if lastColon < 0 {
		return GitignoreMatch{}, false, fmt.Errorf("unexpected check-ignore meta: %q", meta)
	}
	pattern := meta[lastColon+1:]
	rest := meta[:lastColon]
	secondLastColon := strings.LastIndexByte(rest, ':')
	if secondLastColon < 0 {
		return GitignoreMatch{}, false, fmt.Errorf("unexpected check-ignore meta: %q", meta)
	}
	source := rest[:secondLastColon]

	return GitignoreMatch{File: file, Pattern: pattern, Source: source}, true, nil
}

func SensitiveCheck() ([]SensitiveMatch, error) {
	out, err := runGit("diff", "--cached", "--name-only")
	if err != nil {
		return []SensitiveMatch{}, nil
	}
	if out == "" {
		return []SensitiveMatch{}, nil
	}

	files := strings.Split(out, "\n")
	var matches []SensitiveMatch

	matches = append(matches, SensitiveFilenames(files)...)

	for _, file := range files {
		content, err := runGit("show", ":0:"+file)
		if err != nil {
			continue
		}
		matches = append(matches, SensitiveContent(file, content)...)
	}

	return matches, nil
}

func ParseStatus(output string) StatusInfo {
	info := StatusInfo{
		Staged:    []string{},
		Modified:  []string{},
		Untracked: []string{},
	}
	if output == "" {
		return info
	}

	for _, line := range strings.Split(output, "\n") {
		if len(line) < 4 {
			continue
		}
		x := line[0]
		y := line[1]
		file := line[3:]

		if x == '?' {
			info.Untracked = append(info.Untracked, file)
			continue
		}
		if x != ' ' && x != '?' {
			info.Staged = append(info.Staged, file)
		}
		if y != ' ' && y != '?' {
			info.Modified = append(info.Modified, file)
		}
	}

	return info
}

func ParseDiffStat(output string) DiffSummary {
	ds := DiffSummary{}
	if output == "" {
		return ds
	}

	lines := strings.Split(output, "\n")
	summary := lines[len(lines)-1]

	filesRe := regexp.MustCompile(`(\d+) files? changed`)
	insRe := regexp.MustCompile(`(\d+) insertions?\(\+\)`)
	delRe := regexp.MustCompile(`(\d+) deletions?\(-\)`)

	if m := filesRe.FindStringSubmatch(summary); m != nil {
		ds.FilesChanged, _ = strconv.Atoi(m[1])
	}
	if m := insRe.FindStringSubmatch(summary); m != nil {
		ds.Insertions, _ = strconv.Atoi(m[1])
	}
	if m := delRe.FindStringSubmatch(summary); m != nil {
		ds.Deletions, _ = strconv.Atoi(m[1])
	}

	return ds
}

func ParseLog(output string) []CommitInfo {
	if output == "" {
		return []CommitInfo{}
	}

	var commits []CommitInfo
	for _, line := range strings.Split(output, "\n") {
		parts := strings.SplitN(line, " ", 2)
		if len(parts) != 2 {
			continue
		}
		commits = append(commits, CommitInfo{Hash: parts[0], Message: parts[1]})
	}
	return commits
}

func SensitiveFilenames(files []string) []SensitiveMatch {
	var matches []SensitiveMatch
	for _, file := range files {
		for _, p := range sensitiveFilePatterns {
			if p.pattern.MatchString(file) {
				matches = append(matches, SensitiveMatch{File: file, Reason: p.reason})
				break
			}
		}
	}
	return matches
}

func SensitiveContent(file, content string) []SensitiveMatch {
	var matches []SensitiveMatch
	for _, p := range sensitiveContentPatterns {
		if p.pattern.MatchString(content) {
			matches = append(matches, SensitiveMatch{File: file, Reason: p.reason})
		}
	}
	return matches
}
