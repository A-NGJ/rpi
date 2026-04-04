package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/A-NGJ/rpi/internal/chain"
	"github.com/A-NGJ/rpi/internal/frontmatter"
	"github.com/A-NGJ/rpi/internal/scanner"
	"github.com/spf13/cobra"
)

// nowFunc is overridable for testing staleness against a fixed time.
var nowFunc = time.Now

var staleDays int

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show a dashboard overview of all RPI artifacts",
	Long: `Display a single-screen summary of all .rpi/ artifacts including
counts by type and status, stale artifact warnings, active plan
progress with checkbox completion, and archive readiness.`,
	Example: `  # Show status dashboard
  rpi status

  # Custom staleness threshold
  rpi status --stale-days 7

  # JSON output for scripting
  rpi status --format json`,
	RunE: runStatus,
}

func init() {
	addRpiDirFlag(statusCmd)
	addFormatFlag(statusCmd)
	statusCmd.Flags().IntVar(&staleDays, "stale-days", 14, "Days before an artifact is considered stale")
	rootCmd.AddCommand(statusCmd)
}

var (
	// Canonical type display order (alphabetical by singular name).
	statusTypeOrder = []string{"design", "diagnosis", "plan", "research", "review", "spec"}

	// Plural display names.
	statusTypePlurals = map[string]string{
		"design":    "designs",
		"diagnosis": "diagnoses",
		"plan":      "plans",
		"research":  "research",
		"review":    "reviews",
		"spec":      "specs",
	}

	// Status display order within each type row.
	statusDisplayOrder = []string{"active", "draft", "complete", "superseded"}
)

func runStatus(cmd *cobra.Command, args []string) error {
	if _, err := os.Stat(rpiDirFlag); os.IsNotExist(err) {
		return fmt.Errorf("directory not found: %s", rpiDirFlag)
	}

	artifacts, err := scanner.Scan(rpiDirFlag, scanner.Filters{})
	if err != nil {
		return err
	}

	// Group by type -> status -> count, and collect active artifact names
	summary := make(map[string]map[string]int)
	activeByType := make(map[string][]activeSpec)
	for _, a := range artifacts {
		status := "unknown"
		if a.Status != nil {
			status = *a.Status
		}
		if summary[a.Type] == nil {
			summary[a.Type] = make(map[string]int)
		}
		summary[a.Type][status]++

		if status == "active" && a.Type == "spec" {
			name := strings.TrimSuffix(filepath.Base(a.Path), ".md")
			if a.Title != nil {
				name = *a.Title
			}
			activeByType[a.Type] = append(activeByType[a.Type], activeSpec{
				Name:         name,
				Requirements: countSpecRequirements(a.Path),
			})
		}
	}
	for _, specs := range activeByType {
		sort.Slice(specs, func(i, j int) bool {
			return specs[i].Name < specs[j].Name
		})
	}

	// Staleness detection
	stale := findStaleArtifacts(artifacts, nowFunc(), staleDays)

	// Active plan chains
	plans := findActivePlans(artifacts)

	// Archive readiness (ST-13)
	archivable := findArchivable(rpiDirFlag)

	// Format output
	format := formatFlag
	if format == "" {
		format = "text"
	}

	switch format {
	case "json":
		return renderStatusJSON(cmd, summary, activeByType, stale, plans, archivable)
	case "text":
		return renderStatusText(cmd, summary, activeByType, stale, plans, archivable)
	default:
		return fmt.Errorf("unknown format: %s (expected text or json)", format)
	}
}

func renderStatusText(cmd *cobra.Command, summary map[string]map[string]int, activeByType map[string][]activeSpec, stale []staleArtifact, plans []activePlan, archivable []archivableArtifact) error {
	w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)

	fmt.Fprintln(w, "Artifacts")
	for _, typ := range statusTypeOrder {
		counts, ok := summary[typ]
		if !ok {
			continue // ST-2: omit types with zero artifacts
		}
		parts := formatStatusCounts(counts)
		fmt.Fprintf(w, "  %s:\t%s\n", statusTypePlurals[typ], strings.Join(parts, "  "))
	}

	// Active Plans section (ST-10 revised: no linked artifact sub-rows)
	if len(plans) > 0 {
		fmt.Fprintln(w, "\nActive Plans")
		for _, p := range plans {
			progress := ""
			if p.Total > 0 {
				pct := p.Checked * 100 / p.Total
				progress = fmt.Sprintf("\t%d/%d (%d%%)", p.Checked, p.Total, pct)
			}
			fmt.Fprintf(w, "  %s\t%s%s\n", p.Topic, p.Status, progress)
		}
	}

	// ST-18/ST-19: Active Specs section with requirement counts
	if specs, ok := activeByType["spec"]; ok && len(specs) > 0 {
		fmt.Fprintln(w, "\nActive Specs")
		for _, s := range specs {
			fmt.Fprintf(w, "  %s\t%d requirements\n", s.Name, s.Requirements)
		}
	}

	if len(stale) > 0 {
		fmt.Fprintf(w, "\nStale (no update in %d+ days)\n", staleDays)
		for _, s := range stale {
			fmt.Fprintf(w, "  %s\t%s\t%dd ago\n", s.Path, s.Status, s.Age)
		}
	}

	if len(archivable) > 0 {
		// ST-14: summary count grouped by type
		typeCounts := make(map[string]int)
		for _, a := range archivable {
			typeCounts[a.Type]++
		}
		var parts []string
		seen := make(map[string]bool)
		for _, typ := range statusTypeOrder {
			if c, ok := typeCounts[typ]; ok {
				parts = append(parts, fmt.Sprintf("%d %s", c, statusTypePlurals[typ]))
				seen[typ] = true
			}
		}
		// Include any types not in statusTypeOrder (sorted for determinism)
		var extra []string
		for typ := range typeCounts {
			if !seen[typ] {
				extra = append(extra, typ)
			}
		}
		sort.Strings(extra)
		for _, typ := range extra {
			plural := typ + "s"
			if p, ok := statusTypePlurals[typ]; ok {
				plural = p
			}
			parts = append(parts, fmt.Sprintf("%d %s", typeCounts[typ], plural))
		}
		fmt.Fprintf(w, "\nReady to Archive\n")
		fmt.Fprintf(w, "  %d artifacts (%s) with 0 active references\n", len(archivable), strings.Join(parts, ", "))
	}

	w.Flush()
	return nil
}

func renderStatusJSON(cmd *cobra.Command, summary map[string]map[string]int, activeByType map[string][]activeSpec, stale []staleArtifact, plans []activePlan, archivable []archivableArtifact) error {
	out := statusOutput{
		Summary:     summary,
		Stale:       stale,
		ActivePlans: plans,
		ActiveSpecs: activeByType["spec"],
		Archivable:  archivable,
	}
	if out.Stale == nil {
		out.Stale = []staleArtifact{}
	}
	if out.ActivePlans == nil {
		out.ActivePlans = []activePlan{}
	}
	if out.ActiveSpecs == nil {
		out.ActiveSpecs = []activeSpec{}
	}
	if out.Archivable == nil {
		out.Archivable = []archivableArtifact{}
	}
	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), string(data))
	return nil
}

type staleArtifact struct {
	Path   string `json:"path"`
	Status string `json:"status"`
	Age    int    `json:"age_days"`
}

// findStaleArtifacts returns non-terminal artifacts whose date exceeds the threshold.
func findStaleArtifacts(artifacts []scanner.ArtifactInfo, now time.Time, threshold int) []staleArtifact {
	terminalStatuses := map[string]bool{"complete": true, "superseded": true, "archived": true}
	var stale []staleArtifact

	for _, a := range artifacts {
		if a.Status == nil {
			continue
		}
		status := *a.Status
		if terminalStatuses[status] {
			continue
		}

		doc, err := frontmatter.Parse(a.Path)
		if err != nil {
			continue
		}

		// ST-5: specs use last_updated, others use date
		dateKey := "date"
		if a.Type == "spec" {
			dateKey = "last_updated"
		}

		artifactDate, ok := getDateFromFrontmatter(doc.Frontmatter, dateKey)
		if !ok {
			continue // ST-7: skip missing/unparseable dates
		}

		age := int(now.Sub(artifactDate).Hours() / 24)
		if age >= threshold {
			stale = append(stale, staleArtifact{
				Path:   a.Path,
				Status: status,
				Age:    age,
			})
		}
	}

	sort.Slice(stale, func(i, j int) bool {
		return stale[i].Path < stale[j].Path
	})
	return stale
}

// getDateFromFrontmatter extracts a date from frontmatter, handling both
// time.Time (yaml.v3 auto-parse) and string values.
func getDateFromFrontmatter(fm map[string]interface{}, key string) (time.Time, bool) {
	val, ok := fm[key]
	if !ok {
		return time.Time{}, false
	}
	switch v := val.(type) {
	case time.Time:
		return v, true
	case string:
		t, err := parseArtifactDate(v)
		if err != nil {
			return time.Time{}, false
		}
		return t, true
	default:
		return time.Time{}, false
	}
}

// parseArtifactDate parses a date string in RFC3339 or plain date format.
func parseArtifactDate(s string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	return time.Parse("2006-01-02", s)
}

type activePlan struct {
	Path    string     `json:"path"`
	Topic   string     `json:"topic"`
	Status  string     `json:"status"`
	Checked int        `json:"checked"`
	Total   int        `json:"total"`
	Links   []planLink `json:"links"`
}

type planLink struct {
	Type   string `json:"type"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

type activeSpec struct {
	Name         string `json:"name"`
	Requirements int    `json:"requirements"`
}

var requirementPattern = regexp.MustCompile(`\*\*\w+-\d+\*\*:`)

func countSpecRequirements(path string) int {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	return len(requirementPattern.FindAll(data, -1))
}

type archivableArtifact struct {
	Path   string `json:"path"`
	Type   string `json:"type"`
	Status string `json:"status"`
}

type statusOutput struct {
	Summary     map[string]map[string]int `json:"summary"`
	Stale       []staleArtifact           `json:"stale"`
	ActivePlans []activePlan              `json:"active_plans"`
	ActiveSpecs []activeSpec              `json:"active_specs"`
	Archivable  []archivableArtifact      `json:"archivable"`
}

// findActivePlans resolves one-level chain links and checkbox progress for active/draft plans.
func findActivePlans(artifacts []scanner.ArtifactInfo) []activePlan {
	var plans []activePlan

	for _, a := range artifacts {
		if a.Type != "plan" || a.Status == nil {
			continue
		}
		status := *a.Status
		if status != "active" && status != "draft" {
			continue
		}

		topic := filepath.Base(a.Path)
		if a.Title != nil {
			topic = *a.Title
		}

		p := activePlan{
			Path:   a.Path,
			Topic:  topic,
			Status: status,
			Links:  []planLink{},
		}

		// Resolve chain for one-level links (ST-9)
		result, err := chain.Resolve(a.Path, chain.ResolveOptions{})
		if err == nil && len(result.Artifacts) > 0 {
			root := result.Artifacts[0]
			linkSet := make(map[string]bool)
			for _, l := range root.LinksTo {
				linkSet[l] = true
			}

			for _, linked := range result.Artifacts[1:] {
				if !linkSet[linked.Path] {
					continue // Not a direct link — skip deeper levels
				}
				linkStatus := "unknown"
				if linked.Status != nil {
					linkStatus = *linked.Status
				}
				name := strings.TrimSuffix(filepath.Base(linked.Path), ".md")
				p.Links = append(p.Links, planLink{
					Type:   linked.Type,
					Name:   name,
					Status: linkStatus,
				})
			}
		}

		// Parse checkboxes (ST-11, ST-12)
		content, readErr := os.ReadFile(a.Path)
		if readErr == nil {
			cb := parseCheckboxes(string(content))
			p.Checked = cb.Checked
			p.Total = cb.Total
		}

		plans = append(plans, p)
	}

	sort.Slice(plans, func(i, j int) bool {
		return plans[i].Path < plans[j].Path
	})
	return plans
}

// findArchivable returns artifacts with archivable status and zero active references.
func findArchivable(rpiDir string) []archivableArtifact {
	results, err := scanner.Scan(rpiDir, scanner.Filters{Archivable: true})
	if err != nil {
		return nil
	}

	var archivable []archivableArtifact
	for _, r := range results {
		refPath := r.Path
		if rel, relErr := filepath.Rel(rpiDir, r.Path); relErr == nil {
			refPath = rel
		}
		refCount, countErr := scanner.CountReferences(rpiDir, refPath)
		if countErr != nil || refCount > 0 {
			continue // ST-13: only zero-reference artifacts
		}
		status := "unknown"
		if r.Status != nil {
			status = *r.Status
		}
		archivable = append(archivable, archivableArtifact{
			Path:   r.Path,
			Type:   r.Type,
			Status: status,
		})
	}

	sort.Slice(archivable, func(i, j int) bool {
		return archivable[i].Path < archivable[j].Path
	})
	return archivable
}

// formatStatusCounts renders status counts in canonical order.
func formatStatusCounts(counts map[string]int) []string {
	seen := make(map[string]bool)
	var parts []string
	for _, s := range statusDisplayOrder {
		if c, ok := counts[s]; ok {
			parts = append(parts, fmt.Sprintf("%d %s", c, s))
			seen[s] = true
		}
	}
	// Include any unexpected statuses at the end (sorted for determinism).
	var extra []string
	for s := range counts {
		if !seen[s] {
			extra = append(extra, s)
		}
	}
	sort.Strings(extra)
	for _, s := range extra {
		parts = append(parts, fmt.Sprintf("%d %s", counts[s], s))
	}
	return parts
}
