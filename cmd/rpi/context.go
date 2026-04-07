package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/A-NGJ/rpi/internal/chain"
	"github.com/A-NGJ/rpi/internal/frontmatter"
	"github.com/A-NGJ/rpi/internal/git"
	"github.com/A-NGJ/rpi/internal/scanner"
	"github.com/spf13/cobra"
)

const maxNextItems = 3

var contextCmd = &cobra.Command{
	Use:   "context [plan-path]",
	Short: "Show active implementation context",
	Long: `Return a compact snapshot of the active implementation context.

Includes current plan phase and progress, linked spec scenario titles,
design constraints, and git state. Designed for context recovery after
conversation compaction.

When no plan path is given, auto-detects the most recently dated active
plan in .rpi/plans/.`,
	Example: `  # Auto-detect active plan
  rpi context

  # Explicit plan path
  rpi context .rpi/plans/2026-04-07-my-plan.md`,
	Args: cobra.MaximumNArgs(1),
	RunE: runContext,
}

func init() {
	addRpiDirFlag(contextCmd)
	rootCmd.AddCommand(contextCmd)
}

// Output structs

type ContextResult struct {
	Plan        *PlanContext `json:"plan,omitempty"`
	Spec        *SpecContext `json:"spec,omitempty"`
	Constraints string       `json:"constraints,omitempty"`
	Git         *GitContext  `json:"git,omitempty"`
}

type PlanContext struct {
	Path         string   `json:"path"`
	Topic        string   `json:"topic"`
	CurrentPhase string   `json:"current_phase"`
	Progress     Progress `json:"progress"`
	NextItems    []string `json:"next_items"`
}

type Progress struct {
	Checked int `json:"checked"`
	Total   int `json:"total"`
}

type SpecContext struct {
	Path           string   `json:"path"`
	Feature        string   `json:"feature"`
	ScenarioTitles []string `json:"scenario_titles"`
}

type GitContext struct {
	Branch           string `json:"branch"`
	UncommittedFiles int    `json:"uncommitted_files"`
}

func runContext(cmd *cobra.Command, args []string) error {
	planPath := ""
	if len(args) == 1 {
		planPath = args[0]
	}

	result, err := assembleContext(rpiDirFlag, planPath)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal result: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

// detectCurrentPhase walks plan content and returns the name of the first
// phase that has unchecked items, plus up to maxNextItems unchecked item texts.
func detectCurrentPhase(content string) (string, []string) {
	var currentPhase string
	var foundPhase string
	var nextItems []string

	for _, line := range strings.Split(content, "\n") {
		if phaseRe.MatchString(line) {
			currentPhase = strings.TrimSpace(strings.TrimPrefix(line, "## "))
		}

		if uncheckedRe.MatchString(line) {
			if foundPhase == "" {
				foundPhase = currentPhase
			}
			if currentPhase == foundPhase && len(nextItems) < maxNextItems {
				m := uncheckedRe.FindStringSubmatch(line)
				nextItems = append(nextItems, m[2])
			}
		}
	}

	if nextItems == nil {
		nextItems = []string{}
	}
	return foundPhase, nextItems
}

// assembleContext builds a compact context snapshot for the given plan
// (or auto-detects the most recently dated active plan).
func assembleContext(rpiDir, planPath string) (*ContextResult, error) {
	result := &ContextResult{}

	// Auto-detect active plan if no path given
	if planPath == "" {
		detected, err := findActivePlan(rpiDir)
		if err != nil {
			return nil, err
		}
		if detected == "" {
			return result, nil
		}
		planPath = detected
	}

	// Read plan
	planContent, err := os.ReadFile(planPath)
	if err != nil {
		return nil, fmt.Errorf("read plan: %w", err)
	}

	doc, err := frontmatter.Parse(planPath)
	if err != nil {
		return nil, fmt.Errorf("parse plan: %w", err)
	}

	topic, _ := doc.Frontmatter["topic"].(string)
	checkboxes := parseCheckboxes(string(planContent))
	currentPhase, nextItems := detectCurrentPhase(string(planContent))

	result.Plan = &PlanContext{
		Path:         planPath,
		Topic:        topic,
		CurrentPhase: currentPhase,
		Progress: Progress{
			Checked: checkboxes.Checked,
			Total:   checkboxes.Total,
		},
		NextItems: nextItems,
	}

	// Resolve chain to find spec and design
	chainResult, err := chain.Resolve(planPath, chain.ResolveOptions{
		Sections: []string{"Constraints"},
	})
	if err == nil {
		for _, artifact := range chainResult.Artifacts {
			switch artifact.Type {
			case "spec":
				result.Spec = buildSpecContext(artifact.Path)
			case "design":
				if s, ok := artifact.Sections["## Constraints"]; ok {
					result.Constraints = truncateConstraints(s)
				}
			}
		}
	}

	// Git state
	result.Git = buildGitContext()

	return result, nil
}

func findActivePlan(rpiDir string) (string, error) {
	results, err := scanner.Scan(rpiDir, scanner.Filters{
		Type:   "plan",
		Status: "active",
	})
	if err != nil {
		return "", fmt.Errorf("scan for active plans: %w", err)
	}
	if len(results) == 0 {
		return "", nil
	}

	// Pick the most recently dated plan
	best := results[0]
	bestTime := parseDate(best.Date)
	for _, r := range results[1:] {
		t := parseDate(r.Date)
		if t.After(bestTime) {
			best = r
			bestTime = t
		}
	}
	return best.Path, nil
}

func buildSpecContext(specPath string) *SpecContext {
	doc, err := frontmatter.Parse(specPath)
	if err != nil {
		return nil
	}

	feature, _ := doc.Frontmatter["feature"].(string)
	scenarios := parseScenarios(doc.Body)

	titles := make([]string, len(scenarios))
	for i, s := range scenarios {
		titles[i] = s.Title
	}

	return &SpecContext{
		Path:           specPath,
		Feature:        feature,
		ScenarioTitles: titles,
	}
}

func truncateConstraints(section string) string {
	// Strip the heading line
	text := section
	if idx := strings.Index(text, "\n"); idx != -1 {
		text = text[idx+1:]
	}
	text = strings.TrimSpace(text)

	if len(text) > 200 {
		return text[:200] + "..."
	}
	return text
}

func parseDate(s *string) time.Time {
	if s == nil {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339, *s)
	if err != nil {
		return time.Time{}
	}
	return t
}

func buildGitContext() *GitContext {
	ctx := &GitContext{}

	gitResult, err := git.GatherContext()
	if err == nil {
		ctx.Branch = gitResult.Branch
	}

	files, err := git.ChangedFiles()
	if err == nil {
		ctx.UncommittedFiles = len(files)
	}

	return ctx
}
