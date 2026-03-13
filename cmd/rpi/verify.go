package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/A-NGJ/ai-agent-research-plan-implement-flow/internal/git"
	"github.com/spf13/cobra"
)

var verifyCmd = &cobra.Command{
	Use:   "verify <action> [file-path]",
	Short: "Deterministic verification checks",
	Long: `Run deterministic verification checks on plans and source files.

Actions:
  completeness <plan-path>  Parse a plan file for checkboxes (- [ ] / - [x])
                            and **File**: path entries. Compare plan files against
                            git changed files. Reports checked/unchecked counts
                            with phase context, plus file coverage (missing from
                            git, unexpected in git).
  markers [file-path]       Scan for TODO/FIXME/HACK markers. Without a file
                            argument, scans git-changed files (excluding .tmpl/.tpl
                            templates).

Output is JSON by default.`,
	Example: `  # Check plan progress and file coverage
  rpi verify completeness .thoughts/plans/2026-03-13-auth.md
  # → {"total_checkboxes": 12, "checked": 8, "unchecked": 4, ...}

  # Scan for TODO/FIXME/HACK in changed files
  rpi verify markers
  # → {"markers": [...], "count": {"TODO": 2}}

  # Scan a specific file
  rpi verify markers cmd/rpi/scan.go`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runVerify,
}

func init() {
	addFormatFlag(verifyCmd)
	rootCmd.AddCommand(verifyCmd)
}

type CheckboxItem struct {
	Text    string `json:"text"`
	Phase   string `json:"phase,omitempty"`
	Section string `json:"section,omitempty"`
}

type CheckboxResult struct {
	Total          int            `json:"total_checkboxes"`
	Checked        int            `json:"checked"`
	Unchecked      int            `json:"unchecked"`
	UncheckedItems []CheckboxItem `json:"unchecked_items"`
}

type CompareResult struct {
	PlanFiles       []string `json:"plan_files"`
	GitChangedFiles []string `json:"git_changed_files"`
	MissingFromGit  []string `json:"missing_from_git"`
	UnexpectedInGit []string `json:"unexpected_in_git"`
}

type CompletenessResult struct {
	CheckboxResult
	CompareResult
}

type Marker struct {
	File string `json:"file"`
	Line int    `json:"line"`
	Type string `json:"type"`
	Text string `json:"text"`
}

type MarkersResult struct {
	Markers []Marker       `json:"markers"`
	Count   map[string]int `json:"count"`
}

var templateExts = map[string]bool{
	".tmpl": true,
	".tpl":  true,
}

func isTemplateFile(path string) bool {
	for ext := range templateExts {
		if strings.HasSuffix(path, ext) {
			return true
		}
	}
	return false
}

func filesToScan(filePath string) ([]string, error) {
	if filePath != "" {
		return []string{filePath}, nil
	}
	files, err := git.ChangedFiles()
	if err != nil {
		return []string{}, nil
	}
	var filtered []string
	for _, f := range files {
		if !isTemplateFile(f) {
			filtered = append(filtered, f)
		}
	}
	if filtered == nil {
		return []string{}, nil
	}
	return filtered, nil
}

func runVerify(cmd *cobra.Command, args []string) error {
	action := args[0]
	filePath := ""
	if len(args) == 2 {
		filePath = args[1]
	}

	format := formatFlag
	if format == "" {
		format = "json"
	}

	switch action {
	case "completeness":
		if filePath == "" {
			return fmt.Errorf("completeness requires a plan path: rpi verify completeness <plan-path>")
		}
		return runVerifyCompleteness(filePath, format)
	case "markers":
		return runVerifyMarkers(filePath, format)
	default:
		return fmt.Errorf("unknown action: %s (expected completeness or markers)", action)
	}
}

func runVerifyCompleteness(planPath, format string) error {
	content, err := os.ReadFile(planPath)
	if err != nil {
		return fmt.Errorf("reading plan: %w", err)
	}

	checkboxes := parseCheckboxes(string(content))
	planFiles := extractPlanFiles(string(content))

	gitFiles, err := git.ChangedFiles()
	if err != nil {
		gitFiles = []string{}
	}

	compare := comparePlanVsGit(planFiles, gitFiles)

	result := CompletenessResult{
		CheckboxResult: checkboxes,
		CompareResult:  compare,
	}

	switch format {
	case "json":
		data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(data))
	default:
		return fmt.Errorf("unknown format: %s (expected json)", format)
	}
	return nil
}

func runVerifyMarkers(filePath, format string) error {
	files, err := filesToScan(filePath)
	if err != nil {
		return err
	}

	result := MarkersResult{
		Markers: []Marker{},
		Count:   map[string]int{"TODO": 0, "FIXME": 0, "HACK": 0},
	}

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		markers := scanMarkers(file, string(content))
		result.Markers = append(result.Markers, markers...)
		for _, m := range markers {
			result.Count[m.Type]++
		}
	}

	switch format {
	case "json":
		data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(data))
	default:
		return fmt.Errorf("unknown format: %s (expected json)", format)
	}
	return nil
}

var (
	uncheckedRe = regexp.MustCompile(`^(\s*)- \[ \] (.+)`)
	checkedRe   = regexp.MustCompile(`^(\s*)- \[x\] (.+)`)
	phaseRe     = regexp.MustCompile(`^## Phase \d+`)
	sectionRe   = regexp.MustCompile(`^### (.+)`)
	planFileRe  = regexp.MustCompile(`\*\*File\*\*:\s*` + "`" + `([^` + "`" + `]+)` + "`")
	markerRe    = regexp.MustCompile(`\b(TODO|FIXME|HACK)\b(.*)`)
)

func parseCheckboxes(content string) CheckboxResult {
	result := CheckboxResult{
		UncheckedItems: []CheckboxItem{},
	}

	var currentPhase, currentSection string

	for _, line := range strings.Split(content, "\n") {
		if phaseRe.MatchString(line) {
			currentPhase = strings.TrimSpace(strings.TrimPrefix(line, "## "))
			currentSection = ""
		} else if m := sectionRe.FindStringSubmatch(line); m != nil {
			currentSection = strings.TrimSpace(m[1])
		}

		if uncheckedRe.MatchString(line) {
			result.Total++
			result.Unchecked++
			m := uncheckedRe.FindStringSubmatch(line)
			result.UncheckedItems = append(result.UncheckedItems, CheckboxItem{
				Text:    m[2],
				Phase:   currentPhase,
				Section: currentSection,
			})
		} else if checkedRe.MatchString(line) {
			result.Total++
			result.Checked++
		}
	}

	return result
}

func extractPlanFiles(content string) []string {
	var files []string
	seen := map[string]bool{}

	for _, line := range strings.Split(content, "\n") {
		if m := planFileRe.FindStringSubmatch(line); m != nil {
			path := m[1]
			if !seen[path] {
				files = append(files, path)
				seen[path] = true
			}
		}
	}

	if files == nil {
		return []string{}
	}
	return files
}

func comparePlanVsGit(planFiles, gitFiles []string) CompareResult {
	gitSet := map[string]bool{}
	for _, f := range gitFiles {
		gitSet[f] = true
	}

	planSet := map[string]bool{}
	for _, f := range planFiles {
		planSet[f] = true
	}

	var missing []string
	for _, f := range planFiles {
		if !gitSet[f] {
			missing = append(missing, f)
		}
	}

	var unexpected []string
	for _, f := range gitFiles {
		if !planSet[f] {
			unexpected = append(unexpected, f)
		}
	}

	if missing == nil {
		missing = []string{}
	}
	if unexpected == nil {
		unexpected = []string{}
	}

	return CompareResult{
		PlanFiles:       planFiles,
		GitChangedFiles: gitFiles,
		MissingFromGit:  missing,
		UnexpectedInGit: unexpected,
	}
}

func scanMarkers(filename, content string) []Marker {
	var markers []Marker
	for i, line := range strings.Split(content, "\n") {
		if m := markerRe.FindStringSubmatch(line); m != nil {
			markers = append(markers, Marker{
				File: filename,
				Line: i + 1,
				Type: m[1],
				Text: strings.TrimSpace(line),
			})
		}
	}
	return markers
}
