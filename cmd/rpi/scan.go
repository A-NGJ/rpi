package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/A-NGJ/ai-agent-research-plan-implement-flow/internal/scanner"
	"github.com/spf13/cobra"
)

var (
	scanStatus     string
	scanType       string
	scanProposal   string
	scanReferences string
	scanArchivable bool
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Discover and filter artifacts in .rpi/",
	Long: `Walk .rpi/ directory (excludes archive/), parse YAML frontmatter from
each .md file, and return artifacts matching filter criteria.

Filters can be combined; all filters must match (AND logic).
Valid types: plan, proposal, research, spec, review
Valid statuses: draft, active, complete, superseded, archived

Output is JSON by default; use --format md for a markdown table.`,
	Example: `  # All draft research artifacts
  rpi scan --type research --status draft

  # Artifacts ready to archive (complete or superseded, not in archive/)
  rpi scan --archivable

  # Find artifacts that reference a specific file
  rpi scan --references .rpi/proposals/2026-03-13-auth.md

  # Sample JSON output:
  # [
  #   {
  #     "path": ".rpi/research/2026-03-13-auth.md",
  #     "type": "research",
  #     "status": "draft",
  #     "title": "auth flow investigation"
  #   }
  # ]`,
	RunE: runScan,
}

func init() {
	addFormatFlag(scanCmd)
	addRpiDirFlag(scanCmd)
	scanCmd.Flags().StringVar(&scanStatus, "status", "", "Filter by frontmatter status")
	scanCmd.Flags().StringVar(&scanType, "type", "", "Filter by artifact type (plan, proposal, research, etc.)")
	scanCmd.Flags().StringVar(&scanProposal, "proposal", "", "Filter by frontmatter proposal field")
	scanCmd.Flags().StringVar(&scanReferences, "references", "", "Find files that reference this path")
	scanCmd.Flags().BoolVar(&scanArchivable, "archivable", false, "Show only complete/superseded artifacts not in archive/")
	rootCmd.AddCommand(scanCmd)
}

func runScan(cmd *cobra.Command, args []string) error {
	filters := scanner.Filters{
		Status:     scanStatus,
		Type:       scanType,
		Proposal:   scanProposal,
		References: scanReferences,
		Archivable: scanArchivable,
	}

	results, err := scanner.Scan(rpiDirFlag, filters)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	format := formatFlag
	if format == "" {
		format = "json"
	}

	switch format {
	case "json":
		data, _ := json.MarshalIndent(results, "", "  ")
		fmt.Println(string(data))
	case "md":
		printScanTable(results)
	default:
		return fmt.Errorf("unknown format: %s (expected json or md)", format)
	}

	return nil
}

func printScanTable(results []scanner.ArtifactInfo) {
	if len(results) == 0 {
		fmt.Println("No artifacts found.")
		return
	}
	fmt.Println("| Path | Type | Status | Title | Date |")
	fmt.Println("|------|------|--------|-------|------|")
	for _, a := range results {
		status := "-"
		if a.Status != nil {
			status = *a.Status
		}
		title := "-"
		if a.Title != nil {
			title = strings.ReplaceAll(*a.Title, "|", "\\|")
		}
		date := "-"
		if a.Date != nil {
			date = *a.Date
		}
		fmt.Printf("| `%s` | %s | %s | %s | %s |\n", a.Path, a.Type, status, title, date)
	}
}
