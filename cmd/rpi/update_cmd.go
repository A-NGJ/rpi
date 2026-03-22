package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/A-NGJ/ai-agent-research-plan-implement-flow/internal/index"
	"github.com/A-NGJ/ai-agent-research-plan-implement-flow/internal/templates"
	"github.com/A-NGJ/ai-agent-research-plan-implement-flow/internal/workflow"
	"github.com/spf13/cobra"
)

var (
	updateForce      bool
	updateNoClaudeMD bool
)

var updateCmd = &cobra.Command{
	Use:   "update [directory]",
	Short: "Update an initialized project with latest workflow files, directories, and index",
	Long: `Update an existing RPI project to the latest version.

This command brings an already-initialized project up to date:
  - Creates any missing .rpi/ or tool subdirectories
  - Updates skills in .agents/skills/ and tool-specific directory
  - Rebuilds the codebase index
  - Updates the rules file (CLAUDE.md or AGENTS.md)

Workflow files (skills) are only overwritten with --force.`,
	Example: `  # Update current project
  rpi update

  # Force-overwrite workflow files with latest embedded versions
  rpi update --force

  # Update without touching the rules file
  rpi update --no-claude-md`,
	Args: cobra.MaximumNArgs(1),
	RunE: runUpdate,
}

func init() {
	updateCmd.Flags().BoolVar(&updateForce, "force", false, "Overwrite existing workflow files with latest embedded versions")
	updateCmd.Flags().BoolVar(&updateNoClaudeMD, "no-claude-md", false, "Skip rules file update (CLAUDE.md or AGENTS.md)")
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
	targetDir := "."
	if len(args) > 0 {
		targetDir = args[0]
	}

	w := cmd.OutOrStdout()
	rpiDir := filepath.Join(targetDir, ".rpi")

	if _, err := os.Stat(rpiDir); err != nil {
		return fmt.Errorf("not initialized; run rpi init first")
	}

	cfg, err := detectTarget(targetDir)
	if err != nil {
		return err
	}

	// Ensure tool subdirs exist (skip for agents-only)
	if cfg.toolDir != "" {
		toolDirPath := filepath.Join(targetDir, cfg.toolDir)
		for _, d := range cfg.subdirs {
			path := filepath.Join(toolDirPath, d)
			if _, statErr := os.Stat(path); statErr != nil {
				if mkErr := os.MkdirAll(path, 0755); mkErr != nil {
					return fmt.Errorf("create %s: %w", path, mkErr)
				}
				logSuccess(w, fmt.Sprintf("Created %s/%s/", cfg.toolDir, d))
			}
		}
	}

	// Ensure .rpi/ subdirs exist
	rpiSubdirs := []string{
		"research", "designs", "diagnoses",
		"plans", "specs", "reviews", "archive",
	}
	for _, d := range rpiSubdirs {
		path := filepath.Join(rpiDir, d)
		if _, statErr := os.Stat(path); statErr != nil {
			if mkErr := os.MkdirAll(path, 0755); mkErr != nil {
				return fmt.Errorf("create %s: %w", path, mkErr)
			}
			logSuccess(w, fmt.Sprintf("Created .rpi/%s/", d))
		}
	}

	// Install skills to tool-specific or .agents/ skills dir
	var skillsDir string
	if cfg.toolDir != "" {
		skillsDir = filepath.Join(targetDir, cfg.toolDir, "skills")
	} else {
		skillsDir = filepath.Join(targetDir, ".agents", "skills")
	}
	skillCount, err := workflow.InstallSkills(skillsDir, cfg.target, updateForce)
	if err != nil {
		return fmt.Errorf("install skills: %w", err)
	}
	if skillCount > 0 {
		logSuccess(w, fmt.Sprintf("Updated %d skill files", skillCount))
	}

	// Update rules file (skip for agents-only)
	if !updateNoClaudeMD && cfg.rulesFile != "" {
		rulesPath := filepath.Join(targetDir, cfg.rulesFile)
		content, tplErr := templates.Get(cfg.rulesFile)
		if tplErr != nil {
			logWarning(w, fmt.Sprintf("get %s template: %v", cfg.rulesFile, tplErr))
		} else {
			if writeErr := os.WriteFile(rulesPath, []byte(content), 0644); writeErr != nil {
				logWarning(w, fmt.Sprintf("write %s: %v", cfg.rulesFile, writeErr))
			} else {
				logSuccess(w, fmt.Sprintf("Updated %s", cfg.rulesFile))
			}
		}
	}

	// Ensure settings.json has MCP tool permissions (Claude only)
	if cfg.target == workflow.TargetClaude {
		configureSettings(w, filepath.Join(targetDir, cfg.toolDir))
	}

	// Rebuild codebase index
	logInfo(w, "Rebuilding codebase index...")
	idx, err := index.Build(targetDir, index.BuildOptions{})
	if err != nil {
		logWarning(w, fmt.Sprintf("Index build failed: %v", err))
	} else {
		indexPath := filepath.Join(rpiDir, "index.json")
		if saveErr := index.Save(idx, indexPath); saveErr != nil {
			logWarning(w, fmt.Sprintf("Index save failed: %v", saveErr))
		} else {
			logSuccess(w, fmt.Sprintf("Rebuilt codebase index (%d files, %d symbols)", idx.Metadata.FileCount, idx.Metadata.SymbolCount))
		}
	}

	return nil
}
