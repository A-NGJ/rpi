package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	formatFlag       string
	thoughtsDirFlag  string
	templatesDirFlag string
)

var rootCmd = &cobra.Command{
	Use:   "rpi",
	Short: "RPI workflow CLI — context-offloading tool for .thoughts/ artifacts",
	Long: `RPI workflow CLI — manages .thoughts/ artifacts for the
Research → Propose → Plan → Implement → Verify → Archive pipeline.

Each stage produces a markdown artifact in .thoughts/ with YAML frontmatter.
Artifacts link to each other via frontmatter fields (research, proposal,
depends_on), forming dependency chains that rpi can resolve.`,
	Example: `  # Scaffold a research artifact
  rpi scaffold research --topic "auth flow" --write

  # Check what artifacts exist
  rpi scan --type research --status draft

  # Resolve the full chain from a plan back to its research
  rpi chain .thoughts/plans/2026-03-13-auth.md

  # Check plan progress during implementation
  rpi verify completeness .thoughts/plans/2026-03-13-auth.md

  # Update artifact status
  rpi frontmatter transition .thoughts/plans/2026-03-13-auth.md complete

  # Find and archive completed artifacts
  rpi archive scan
  rpi archive move .thoughts/plans/2026-03-13-auth.md`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func addFormatFlag(cmd *cobra.Command) {
	cmd.Flags().StringVar(&formatFlag, "format", "", "Output format: json, md, text")
}

func addThoughtsDirFlag(cmd *cobra.Command) {
	cmd.Flags().StringVar(&thoughtsDirFlag, "thoughts-dir", ".thoughts", "Path to .thoughts/ directory")
}

func addTemplatesDirFlag(cmd *cobra.Command) {
	cmd.Flags().StringVar(&templatesDirFlag, "templates-dir", ".claude/templates", "Path to templates directory")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
