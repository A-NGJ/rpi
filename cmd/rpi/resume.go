package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/A-NGJ/rpi/internal/frontmatter"
	"github.com/A-NGJ/rpi/internal/scanner"
	"github.com/spf13/cobra"
)

// ResumeResult is the session-level overview returned by rpi resume.
type ResumeResult struct {
	Artifacts  []ResumeArtifact   `json:"artifacts"`
	ActivePlan *ActivePlanSummary `json:"active_plan,omitempty"`
	Suggestion *Suggestion        `json:"suggestion,omitempty"`
}

// ResumeArtifact is a compact artifact entry in the resume output.
type ResumeArtifact struct {
	Path   string `json:"path"`
	Type   string `json:"type"`
	Status string `json:"status"`
	Topic  string `json:"topic"`
}

// ActivePlanSummary is a compact plan context for the session resume.
type ActivePlanSummary struct {
	Path         string   `json:"path"`
	Topic        string   `json:"topic"`
	CurrentPhase string   `json:"current_phase"`
	Progress     Progress `json:"progress"`
	NextItems    []string `json:"next_items"`
}

var resumeCmd = &cobra.Command{
	Use:   "resume",
	Short: "Show active work and suggested next steps",
	Long: `Return a session-level overview of active RPI work.

Lists all active and draft artifacts, shows the current implementation
context for the most recent active plan, and suggests the next pipeline
action. Designed for session start to quickly restore context.`,
	Example: `  rpi resume`,
	Args:    cobra.NoArgs,
	RunE:    runResume,
}

func init() {
	addRpiDirFlag(resumeCmd)
	rootCmd.AddCommand(resumeCmd)
}

func runResume(cmd *cobra.Command, args []string) error {
	result, err := assembleResume(rpiDirFlag)
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

// assembleResume builds a session-level overview: active artifacts, plan
// context, and a pipeline suggestion.
func assembleResume(rpiDir string) (*ResumeResult, error) {
	allArtifacts, err := scanner.Scan(rpiDir, scanner.Filters{})
	if err != nil {
		return nil, fmt.Errorf("scan artifacts: %w", err)
	}

	result := &ResumeResult{
		Artifacts: []ResumeArtifact{},
	}

	// Filter to active/draft artifacts for the list
	for _, a := range allArtifacts {
		if a.Status == nil {
			continue
		}
		status := *a.Status
		if status != "active" && status != "draft" {
			continue
		}
		topic := ""
		if a.Title != nil {
			topic = *a.Title
		}
		result.Artifacts = append(result.Artifacts, ResumeArtifact{
			Path:   a.Path,
			Type:   a.Type,
			Status: status,
			Topic:  topic,
		})
	}

	// Build active plan context for the most recent active plan
	activePlans := filterArtifacts(allArtifacts, "plan", "active")
	if len(activePlans) > 0 {
		sortByDateDesc(activePlans)
		plan := activePlans[0]
		result.ActivePlan = buildActivePlanSummary(plan)
	}

	// Pipeline suggestion
	suggestion, err := suggestNext(rpiDir, "")
	if err == nil {
		result.Suggestion = suggestion
	}

	return result, nil
}

// buildActivePlanSummary constructs a compact plan context from an artifact.
func buildActivePlanSummary(plan scanner.ArtifactInfo) *ActivePlanSummary {
	content, err := os.ReadFile(plan.Path)
	if err != nil {
		return nil
	}

	doc, err := frontmatter.Parse(plan.Path)
	if err != nil {
		return nil
	}

	topic, _ := doc.Frontmatter["topic"].(string)
	checkboxes := parseCheckboxes(string(content))
	currentPhase, nextItems := detectCurrentPhase(string(content))

	return &ActivePlanSummary{
		Path:         plan.Path,
		Topic:        topic,
		CurrentPhase: currentPhase,
		Progress: Progress{
			Checked: checkboxes.Checked,
			Total:   checkboxes.Total,
		},
		NextItems: nextItems,
	}
}
