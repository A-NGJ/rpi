package main

import (
	"fmt"
	"os"

	tmpl "github.com/A-NGJ/ai-agent-research-plan-implement-flow/internal/template"
	"github.com/A-NGJ/ai-agent-research-plan-implement-flow/internal/workflow"
	"github.com/spf13/cobra"
)

var (
	formatFlag       string
	rpiDirFlag       string
	templatesDirFlag string
)

var rootCmd = &cobra.Command{
	Use:   "rpi",
	Short: "RPI workflow CLI — context-offloading tool for .rpi/ artifacts",
	Long: `RPI workflow CLI — manages .rpi/ artifacts for the
Research → Propose → Plan → Implement → Verify → Archive pipeline.

Each stage produces a markdown artifact in .rpi/ with YAML frontmatter.
Artifacts link to each other via frontmatter fields (research, design,
depends_on), forming dependency chains that rpi can resolve.`,
	Example: `  # Scaffold a research artifact
  rpi scaffold research --topic "auth flow" --write

  # Check what artifacts exist
  rpi scan --type research --status draft

  # Resolve the full chain from a plan back to its research
  rpi chain .rpi/plans/2026-03-13-auth.md

  # Check plan progress during implementation
  rpi verify completeness .rpi/plans/2026-03-13-auth.md

  # Update artifact status
  rpi frontmatter transition .rpi/plans/2026-03-13-auth.md complete

  # Find and archive completed artifacts
  rpi archive scan
  rpi archive move .rpi/plans/2026-03-13-auth.md`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func addFormatFlag(cmd *cobra.Command) {
	cmd.Flags().StringVar(&formatFlag, "format", "", "Output format: json, md, text")
}

func addRpiDirFlag(cmd *cobra.Command) {
	cmd.Flags().StringVar(&rpiDirFlag, "rpi-dir", ".rpi", "Path to .rpi/ artifacts directory")
}

func addTemplatesDirFlag(cmd *cobra.Command) {
	cmd.Flags().StringVar(&templatesDirFlag, "templates-dir", ".rpi/templates", "Path to templates directory")
}

func init() {
	tmpl.EmbeddedTemplateReader = func(name string) ([]byte, error) {
		return workflow.ReadAsset("templates/" + name + ".tmpl")
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
