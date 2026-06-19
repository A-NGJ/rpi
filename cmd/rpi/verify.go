package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/A-NGJ/rpi/internal/coverage"
	"github.com/A-NGJ/rpi/internal/frontmatter"
	"github.com/A-NGJ/rpi/internal/git"
	"github.com/spf13/cobra"
)

var verifyCmd = &cobra.Command{
	Use:   "verify <action> [file-path]",
	Short: "Deterministic verification checks",
	Long: `Run deterministic verification checks on plans, specs, and source files.

Actions:
  completeness <plan-path>  Parse a plan file for checkboxes (- [ ] / - [x])
                            and **File**: path entries. Compare plan files against
                            git changed files. Reports checked/unchecked counts
                            with phase context, plus file coverage (missing from
                            git, unexpected in git).
  coverage --pre-lock <plan-path>
                            Pre-lock audit of a drafted plan (no git diff): reports
                            coverage gaps (orphaned criteria, uncovered/unjustified
                            files), phase-ordering forward-references and cycles,
                            and edit-target existence, with a hardFailure flag.
  markers [file-path]       Scan for TODO/FIXME/HACK markers. Without a file
                            argument, scans git-changed files (excluding .tmpl/.tpl
                            templates).
  spec <spec-path>          Parse a spec file's ## Scenarios section and extract
                            Given/When/Then scenario blocks as structured JSON.

Output is JSON by default.`,
	Example: `  # Check plan progress and file coverage
  rpi verify completeness .rpi/plans/2026-03-13-auth.md
  # → {"total_checkboxes": 12, "checked": 8, "unchecked": 4, ...}

  # Scan for TODO/FIXME/HACK in changed files
  rpi verify markers
  # → {"markers": [...], "count": {"TODO": 2}}

  # Scan a specific file
  rpi verify markers cmd/rpi/scan.go

  # Pre-lock coverage audit of a drafted plan
  rpi verify coverage --pre-lock .rpi/plans/2026-03-13-auth.md
  # → {"coverage": {...}, "ordering": {...}, "existence": {...}, "hardFailure": false}

  # Parse spec scenarios
  rpi verify spec .rpi/specs/my-feature.md
  # → {"spec": "...", "feature": "...", "scenarios": [...], "total": 6}`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runVerify,
}

var preLockFlag bool

func init() {
	addFormatFlag(verifyCmd)
	verifyCmd.Flags().BoolVar(&preLockFlag, "pre-lock", false, "Run coverage in pre-lock mode (drafted plan, no git diff)")
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
	case "coverage":
		if filePath == "" {
			return fmt.Errorf("coverage requires a plan path: rpi verify coverage --pre-lock <plan-path>")
		}
		return runVerifyCoverage(filePath, format)
	case "markers":
		return runVerifyMarkers(filePath, format)
	case "spec":
		if filePath == "" {
			return fmt.Errorf("spec requires a spec path: rpi verify spec <spec-path>")
		}
		return runVerifySpec(filePath, format)
	default:
		return fmt.Errorf("unknown action: %s (expected completeness, coverage, markers, or spec)", action)
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

// runVerifyCoverage runs the deterministic pre-lock coverage/ordering/existence
// audit over a drafted plan and prints the JSON verdict. Unlike completeness,
// it consults no git diff — it reasons purely about whether the drafted plan
// coheres with itself.
func runVerifyCoverage(planPath, format string) error {
	if !preLockFlag {
		return fmt.Errorf("rpi verify coverage currently supports only --pre-lock mode; pass --pre-lock")
	}

	content, err := os.ReadFile(planPath)
	if err != nil {
		return fmt.Errorf("reading plan: %w", err)
	}

	root, err := os.Getwd()
	if err != nil {
		root = ""
	}

	result := coverage.Analyze(string(content), root)

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
	markerRe = regexp.MustCompile(`\b(TODO|FIXME|HACK)\b(.*)`)
)

// parseCheckboxes delegates to the shared coverage extractor (the single plan
// parser) and assembles the completeness-shaped CheckboxResult.
func parseCheckboxes(content string) CheckboxResult {
	result := CheckboxResult{
		UncheckedItems: []CheckboxItem{},
	}

	for _, cb := range coverage.ParseCheckboxes(content) {
		result.Total++
		if cb.Checked {
			result.Checked++
			continue
		}
		result.Unchecked++
		result.UncheckedItems = append(result.UncheckedItems, CheckboxItem{
			Text:    cb.Text,
			Phase:   cb.Phase,
			Section: cb.Section,
		})
	}

	return result
}

func extractPlanFiles(content string) []string {
	return coverage.ExtractPlanFiles(content)
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

// Spec verification types and functions

type Scenario struct {
	Title string `json:"title"`
	Given string `json:"given"`
	When  string `json:"when"`
	Then  string `json:"then"`
}

type SpecResult struct {
	Spec      string     `json:"spec"`
	Feature   string     `json:"feature"`
	Scenarios []Scenario `json:"scenarios"`
	Total     int        `json:"total"`
}

func runVerifySpec(specPath, format string) error {
	doc, err := frontmatter.Parse(specPath)
	if err != nil {
		return fmt.Errorf("reading spec: %w", err)
	}

	feature, _ := doc.Frontmatter["feature"].(string)

	scenarios := parseScenarios(doc.Body)

	result := SpecResult{
		Spec:      specPath,
		Feature:   feature,
		Scenarios: scenarios,
		Total:     len(scenarios),
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

// parseScenarios extracts scenario blocks from a ## Scenarios section.
// Each scenario starts with a ### heading and contains Given/When/Then steps.
// Continuation lines (without a keyword) are appended to the current step.
func parseScenarios(body string) []Scenario {
	section, ok := frontmatter.ExtractSection(body, "Scenarios")
	if !ok {
		return []Scenario{}
	}

	var scenarios []Scenario
	var current *Scenario
	var step *string // points to the field currently being appended to

	for _, line := range strings.Split(section, "\n") {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "### ") {
			if current != nil {
				scenarios = append(scenarios, *current)
			}
			current = &Scenario{Title: strings.TrimPrefix(trimmed, "### ")}
			step = nil
			continue
		}

		if current == nil {
			continue
		}

		lower := strings.ToLower(trimmed)
		if strings.HasPrefix(lower, "given ") {
			current.Given = trimmed[len("Given "):]
			step = &current.Given
		} else if strings.HasPrefix(lower, "when ") {
			current.When = trimmed[len("When "):]
			step = &current.When
		} else if strings.HasPrefix(lower, "then ") {
			current.Then = trimmed[len("Then "):]
			step = &current.Then
		} else if trimmed != "" && step != nil {
			*step += " " + trimmed
		}
	}

	if current != nil {
		scenarios = append(scenarios, *current)
	}

	return scenarios
}
