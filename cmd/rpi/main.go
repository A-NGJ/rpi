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
	Long:  "RPI workflow CLI — context-offloading tool for .thoughts/ artifacts.\n\nHandles template scaffolding, YAML frontmatter manipulation, artifact chain\nresolution, directory scanning, git context gathering, and archive operations.",
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
