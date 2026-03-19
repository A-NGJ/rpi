package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/A-NGJ/ai-agent-research-plan-implement-flow/internal/chain"
	"github.com/spf13/cobra"
)

var sectionsFlag string

var chainCmd = &cobra.Command{
	Use:   "chain <artifact-path>",
	Short: "Resolve artifact cross-reference chain",
	Long: `Follow frontmatter link fields recursively from a root artifact and return
the full dependency chain as a flat metadata list.

Link fields followed: research, design, ticket, depends_on, related_research.
Recurses up to depth 10 and detects cycles. Falls back to scanning
"## Source Documents" / "## References" sections for .rpi/ paths.

Use --sections to extract specific ## headings from each artifact body,
avoiding separate file reads. Output is JSON by default; use --format md
for a markdown table.`,
	Example: `  # Resolve a plan's full dependency chain
  rpi chain .rpi/plans/2026-03-13-auth.md

  # Resolve and extract Summary + Assessment sections from each linked artifact
  rpi chain .rpi/plans/2026-03-13-auth.md --sections "Summary,Assessment"

  # Sample JSON output:
  # {
  #   "root": ".rpi/plans/2026-03-13-auth.md",
  #   "artifacts": [
  #     {"path": ".rpi/plans/2026-03-13-auth.md", "type": "plan", "status": "draft", ...},
  #     {"path": ".rpi/designs/2026-03-12-auth.md", "type": "design", ...}
  #   ]
  # }`,
	Args: cobra.ExactArgs(1),
	RunE: runChain,
}

func init() {
	addFormatFlag(chainCmd)
	chainCmd.Flags().StringVar(&sectionsFlag, "sections", "", "Comma-separated section names to extract from each artifact")
	rootCmd.AddCommand(chainCmd)
}

func runChain(cmd *cobra.Command, args []string) error {
	opts := chain.ResolveOptions{}
	if sectionsFlag != "" {
		parts := strings.Split(sectionsFlag, ",")
		for i, p := range parts {
			parts[i] = strings.TrimSpace(p)
		}
		opts.Sections = parts
	}

	result, err := chain.Resolve(args[0], opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	for _, w := range result.Warnings {
		fmt.Fprintf(os.Stderr, "warning: %s\n", w)
	}

	format := formatFlag
	if format == "" {
		format = "json"
	}

	switch format {
	case "json":
		data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(data))
	case "md":
		printChainTable(result)
	default:
		return fmt.Errorf("unknown format: %s (expected json or md)", format)
	}

	return nil
}

func printChainTable(result *chain.Result) {
	fmt.Printf("**Chain root**: `%s`\n\n", result.Root)
	fmt.Println("| Path | Type | Status | Title |")
	fmt.Println("|------|------|--------|-------|")
	for _, a := range result.Artifacts {
		status := "-"
		if a.Status != nil {
			status = *a.Status
		}
		title := "-"
		if a.Title != nil {
			title = *a.Title
		}
		path := fmt.Sprintf("`%s`", a.Path)
		fmt.Printf("| %s | %s | %s | %s |\n", path, a.Type, status, strings.ReplaceAll(title, "|", "\\|"))
	}

	// Print sections if any artifact has them
	for _, a := range result.Artifacts {
		if len(a.Sections) > 0 {
			fmt.Printf("\n---\n\n### `%s`\n\n", a.Path)
			for _, content := range a.Sections {
				fmt.Println(content)
			}
		}
	}
}
