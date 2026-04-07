package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/A-NGJ/rpi/internal/chain"
	"github.com/A-NGJ/rpi/internal/frontmatter"
	"github.com/A-NGJ/rpi/internal/scanner"
	"github.com/spf13/cobra"
)

// Suggestion represents a recommended next pipeline action.
type Suggestion struct {
	Action    string `json:"action"`
	Reasoning string `json:"reasoning"`
	Artifact  string `json:"artifact"`
}

var nextCmd = &cobra.Command{
	Use:   "next [artifact-path]",
	Short: "Suggest the next pipeline action",
	Long: `Analyze artifact state and recommend the next pipeline step.

Evaluates pipeline rules in priority order: active plans (implement/verify),
designs without plans, draft designs, complete research without designs,
draft research. When multiple artifacts compete at the same priority,
the most recently dated artifact wins.

Optionally pass an artifact path to get a suggestion specific to that artifact.`,
	Example: `  # Auto-detect from all artifacts
  rpi next

  # Suggest for specific artifact
  rpi next .rpi/designs/2026-04-07-my-design.md`,
	Args: cobra.MaximumNArgs(1),
	RunE: runNext,
}

func init() {
	addRpiDirFlag(nextCmd)
	rootCmd.AddCommand(nextCmd)
}

func runNext(cmd *cobra.Command, args []string) error {
	artifactPath := ""
	if len(args) == 1 {
		artifactPath = args[0]
	}

	suggestion, err := suggestNext(rpiDirFlag, artifactPath)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(suggestion, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal result: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

// suggestNext analyzes artifact state and recommends the next pipeline action.
// If artifactPath is provided, suggests the next step for that specific artifact.
// Otherwise, scans all artifacts and applies priority-ordered pipeline rules.
func suggestNext(rpiDir, artifactPath string) (*Suggestion, error) {
	if artifactPath != "" {
		return suggestForArtifact(artifactPath)
	}

	allArtifacts, err := scanner.Scan(rpiDir, scanner.Filters{})
	if err != nil {
		return nil, fmt.Errorf("scan artifacts: %w", err)
	}

	planDesigns, designResearch := buildDownstreamMaps(allArtifacts)

	// Priority 1+2: Active plans
	activePlans := filterArtifacts(allArtifacts, "plan", "active")
	sortByDateDesc(activePlans)
	for _, plan := range activePlans {
		content, err := os.ReadFile(plan.Path)
		if err != nil {
			continue
		}
		checkboxes := parseCheckboxes(string(content))
		topic := titleOrPath(plan)
		if checkboxes.Unchecked > 0 {
			return &Suggestion{
				Action:    fmt.Sprintf("/rpi-implement %s", plan.Path),
				Reasoning: fmt.Sprintf("Active plan '%s' has %d unchecked items remaining", topic, checkboxes.Unchecked),
				Artifact:  plan.Path,
			}, nil
		}
		// All checked — suggest verification
		specPath := resolveSpecFromPlan(plan.Path)
		if specPath != "" {
			return &Suggestion{
				Action:    fmt.Sprintf("/rpi-verify %s", specPath),
				Reasoning: fmt.Sprintf("Active plan '%s' has all items checked — verify against spec", topic),
				Artifact:  plan.Path,
			}, nil
		}
		return &Suggestion{
			Action:    fmt.Sprintf("/rpi-verify (plan: %s)", plan.Path),
			Reasoning: fmt.Sprintf("Active plan '%s' has all items checked — verify completion", topic),
			Artifact:  plan.Path,
		}, nil
	}

	// Priority 3: Active designs without downstream plans
	activeDesigns := filterArtifacts(allArtifacts, "design", "active")
	sortByDateDesc(activeDesigns)
	for _, design := range activeDesigns {
		if !planDesigns[design.Path] {
			return &Suggestion{
				Action:    fmt.Sprintf("/rpi-plan %s", design.Path),
				Reasoning: fmt.Sprintf("Active design '%s' has no implementation plan", titleOrPath(design)),
				Artifact:  design.Path,
			}, nil
		}
	}

	// Priority 4: Draft designs
	draftDesigns := filterArtifacts(allArtifacts, "design", "draft")
	sortByDateDesc(draftDesigns)
	if len(draftDesigns) > 0 {
		d := draftDesigns[0]
		return &Suggestion{
			Action:    fmt.Sprintf("Review and approve design at %s", d.Path),
			Reasoning: fmt.Sprintf("Draft design '%s' needs review", titleOrPath(d)),
			Artifact:  d.Path,
		}, nil
	}

	// Priority 5: Complete research without downstream designs
	completeResearch := filterArtifacts(allArtifacts, "research", "complete")
	sortByDateDesc(completeResearch)
	for _, research := range completeResearch {
		if !designResearch[research.Path] {
			return &Suggestion{
				Action:    fmt.Sprintf("/rpi-propose %s", research.Path),
				Reasoning: fmt.Sprintf("Complete research '%s' has no downstream design", titleOrPath(research)),
				Artifact:  research.Path,
			}, nil
		}
	}

	// Priority 6: Draft research
	draftResearch := filterArtifacts(allArtifacts, "research", "draft")
	sortByDateDesc(draftResearch)
	if len(draftResearch) > 0 {
		r := draftResearch[0]
		return &Suggestion{
			Action:    fmt.Sprintf("Review and finalize research at %s", r.Path),
			Reasoning: fmt.Sprintf("Draft research '%s' needs review", titleOrPath(r)),
			Artifact:  r.Path,
		}, nil
	}

	// Priority 7: Nothing active
	return &Suggestion{
		Action:    "/rpi-propose or /rpi-research",
		Reasoning: "No active or draft artifacts found",
		Artifact:  "",
	}, nil
}

// suggestForArtifact suggests the next step for a specific artifact.
func suggestForArtifact(artifactPath string) (*Suggestion, error) {
	doc, err := frontmatter.Parse(artifactPath)
	if err != nil {
		return nil, fmt.Errorf("parse artifact: %w", err)
	}

	artType := scanner.InferType(artifactPath)
	status, _ := doc.Frontmatter["status"].(string)
	topic, _ := doc.Frontmatter["topic"].(string)
	if topic == "" {
		topic = artifactPath
	}

	switch artType {
	case "plan":
		if status == "active" {
			content, err := os.ReadFile(artifactPath)
			if err != nil {
				return nil, fmt.Errorf("read plan: %w", err)
			}
			checkboxes := parseCheckboxes(string(content))
			if checkboxes.Unchecked > 0 {
				return &Suggestion{
					Action:    fmt.Sprintf("/rpi-implement %s", artifactPath),
					Reasoning: fmt.Sprintf("Active plan '%s' has %d unchecked items remaining", topic, checkboxes.Unchecked),
					Artifact:  artifactPath,
				}, nil
			}
			specPath := resolveSpecFromPlan(artifactPath)
			if specPath != "" {
				return &Suggestion{
					Action:    fmt.Sprintf("/rpi-verify %s", specPath),
					Reasoning: fmt.Sprintf("Active plan '%s' has all items checked — verify against spec", topic),
					Artifact:  artifactPath,
				}, nil
			}
			return &Suggestion{
				Action:    fmt.Sprintf("/rpi-verify (plan: %s)", artifactPath),
				Reasoning: fmt.Sprintf("Active plan '%s' has all items checked — verify completion", topic),
				Artifact:  artifactPath,
			}, nil
		}
		return &Suggestion{
			Action:    fmt.Sprintf("Review plan at %s", artifactPath),
			Reasoning: fmt.Sprintf("Plan '%s' is in %s status", topic, status),
			Artifact:  artifactPath,
		}, nil

	case "design":
		if status == "active" {
			return &Suggestion{
				Action:    fmt.Sprintf("/rpi-plan %s", artifactPath),
				Reasoning: fmt.Sprintf("Active design '%s' needs an implementation plan", topic),
				Artifact:  artifactPath,
			}, nil
		}
		if status == "draft" {
			return &Suggestion{
				Action:    fmt.Sprintf("Review and approve design at %s", artifactPath),
				Reasoning: fmt.Sprintf("Draft design '%s' needs review", topic),
				Artifact:  artifactPath,
			}, nil
		}
		return &Suggestion{
			Action:    fmt.Sprintf("Design '%s' is %s — no action needed", topic, status),
			Reasoning: "No further action needed",
			Artifact:  artifactPath,
		}, nil

	case "research":
		if status == "complete" {
			return &Suggestion{
				Action:    fmt.Sprintf("/rpi-propose %s", artifactPath),
				Reasoning: fmt.Sprintf("Complete research '%s' can be used to propose a solution", topic),
				Artifact:  artifactPath,
			}, nil
		}
		if status == "draft" {
			return &Suggestion{
				Action:    fmt.Sprintf("Review and finalize research at %s", artifactPath),
				Reasoning: fmt.Sprintf("Draft research '%s' needs review", topic),
				Artifact:  artifactPath,
			}, nil
		}
		return &Suggestion{
			Action:    fmt.Sprintf("Research '%s' is %s — no action needed", topic, status),
			Reasoning: "No further action needed",
			Artifact:  artifactPath,
		}, nil

	default:
		return &Suggestion{
			Action:    fmt.Sprintf("Review artifact at %s", artifactPath),
			Reasoning: fmt.Sprintf("Artifact type '%s' with status '%s'", artType, status),
			Artifact:  artifactPath,
		}, nil
	}
}

// buildDownstreamMaps reads frontmatter from plans and designs to build lookup
// sets: "designs that have plans" and "research that has designs".
func buildDownstreamMaps(artifacts []scanner.ArtifactInfo) (planDesigns, designResearch map[string]bool) {
	planDesigns = make(map[string]bool)
	designResearch = make(map[string]bool)

	for _, a := range artifacts {
		switch a.Type {
		case "plan":
			doc, err := frontmatter.Parse(a.Path)
			if err != nil {
				continue
			}
			if design, ok := doc.Frontmatter["design"].(string); ok && design != "" {
				planDesigns[design] = true
			}
		case "design":
			doc, err := frontmatter.Parse(a.Path)
			if err != nil {
				continue
			}
			if research, ok := doc.Frontmatter["related_research"].(string); ok && research != "" {
				designResearch[research] = true
			}
		}
	}
	return
}

// filterArtifacts returns artifacts matching the given type and status.
func filterArtifacts(artifacts []scanner.ArtifactInfo, artType, status string) []scanner.ArtifactInfo {
	var result []scanner.ArtifactInfo
	for _, a := range artifacts {
		if a.Type == artType && a.Status != nil && *a.Status == status {
			result = append(result, a)
		}
	}
	return result
}

// sortByDateDesc sorts artifacts by frontmatter date descending (most recent first).
func sortByDateDesc(artifacts []scanner.ArtifactInfo) {
	sort.Slice(artifacts, func(i, j int) bool {
		ti := parseDate(artifacts[i].Date)
		tj := parseDate(artifacts[j].Date)
		return ti.After(tj)
	})
}

// titleOrPath returns the artifact's title if available, otherwise its path.
func titleOrPath(a scanner.ArtifactInfo) string {
	if a.Title != nil && *a.Title != "" {
		return *a.Title
	}
	return a.Path
}

// resolveSpecFromPlan resolves the chain from a plan to find its linked spec path.
func resolveSpecFromPlan(planPath string) string {
	result, err := chain.Resolve(planPath, chain.ResolveOptions{})
	if err != nil {
		return ""
	}
	for _, a := range result.Artifacts {
		if a.Type == "spec" {
			return a.Path
		}
	}
	return ""
}
