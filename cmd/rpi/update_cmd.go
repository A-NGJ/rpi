package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	updateNoClaudeMD bool
	updateGlobal     bool
	updateTarget     string
)

var updateCmd = &cobra.Command{
	Use:   "update [directory]",
	Short: "Update an initialized project with latest workflow files, directories, and index",
	Long: `Update an existing RPI project to the latest version.

This command brings an already-initialized project up to date:
  - Creates any missing .rpi/ or tool subdirectories
  - Updates skills in .agents/skills/ and tool-specific directory

Modified files are backed up to .bak before overwriting.

Use --global to refresh the user-level install at ~/.claude (or
~/.config/opencode with --target opencode) instead of a project.`,
	Example: `  # Update current project
  rpi update

  # Update without touching the rules file
  rpi update --no-claude-md

  # Update the user-level install (Claude Code)
  rpi update --global

  # Update the user-level install (OpenCode)
  rpi update --global --target opencode`,
	Args: cobra.MaximumNArgs(1),
	RunE: runUpdate,
}

func init() {
	updateCmd.Flags().BoolVar(&updateNoClaudeMD, "no-claude-md", false, "Skip rules file update (CLAUDE.md or AGENTS.md)")
	updateCmd.Flags().BoolVar(&updateGlobal, "global", false, "Update the user-level install (~/.claude or ~/.config/opencode)")
	updateCmd.Flags().StringVar(&updateTarget, "target", "claude", `Target for --global: "claude" or "opencode" (ignored without --global)`)
	rootCmd.AddCommand(updateCmd)
}

// detectTarget determines the target config from existing tool directories.
func detectTarget(targetDir string) (targetConfig, error) {
	if _, err := os.Stat(filepath.Join(targetDir, ".claude")); err == nil {
		return resolveTargetConfig("claude")
	}
	if _, err := os.Stat(filepath.Join(targetDir, ".opencode")); err == nil {
		return resolveTargetConfig("opencode")
	}
	if _, err := os.Stat(filepath.Join(targetDir, ".agents")); err == nil {
		return resolveTargetConfig("agents-only")
	}
	return targetConfig{}, fmt.Errorf("no .claude/, .opencode/, or .agents/ found; run rpi init first")
}

func runUpdate(cmd *cobra.Command, args []string) error {
	if updateGlobal {
		if len(args) > 0 {
			return fmt.Errorf("--global and a directory argument are mutually exclusive")
		}
		cfg, err := resolveTargetConfig(updateTarget)
		if err != nil {
			return err
		}
		targetDir, cfg, err := resolveGlobalTargetDir(cfg)
		if err != nil {
			return err
		}
		return syncProject(syncOptions{
			targetDir: targetDir,
			cfg:       cfg,
			skipRules: updateNoClaudeMD,
			global:    true,
			w:         cmd.OutOrStdout(),
		})
	}

	targetDir := "."
	if len(args) > 0 {
		targetDir = args[0]
	}

	rpiDir := filepath.Join(targetDir, ".rpi")
	if _, err := os.Stat(rpiDir); err != nil {
		return fmt.Errorf("not initialized; run rpi init first")
	}

	cfg, err := detectTarget(targetDir)
	if err != nil {
		return err
	}

	return syncProject(syncOptions{
		targetDir: targetDir,
		cfg:       cfg,
		skipRules: updateNoClaudeMD,
		w:         cmd.OutOrStdout(),
	})
}
