package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/A-NGJ/ai-agent-research-plan-implement-flow/internal/index"
	"github.com/A-NGJ/ai-agent-research-plan-implement-flow/internal/templates"
	"github.com/A-NGJ/ai-agent-research-plan-implement-flow/internal/workflow"
	"github.com/spf13/cobra"
)

var (
	initNoClaudeMD bool
	initTrackRpi   bool
	initTarget     string
	initNoMCP      bool
)

const (
	colorRed    = "\033[0;31m"
	colorGreen  = "\033[0;32m"
	colorYellow = "\033[0;33m"
	colorReset  = "\033[0m"
)

func logSuccess(w io.Writer, msg string) {
	fmt.Fprintf(w, "%s✓%s %s\n", colorGreen, colorReset, msg)
}

func logWarning(w io.Writer, msg string) {
	fmt.Fprintf(w, "%s!%s %s\n", colorYellow, colorReset, msg)
}

func logError(w io.Writer, msg string) {
	fmt.Fprintf(w, "%s✗%s %s\n", colorRed, colorReset, msg)
}

func logInfo(w io.Writer, msg string) {
	fmt.Fprintf(w, "  %s\n", msg)
}

var initCmd = &cobra.Command{
	Use:   "init [directory]",
	Short: "Initialize project with workflow directories and rules file for Claude Code or OpenCode",
	Long: `Initialize a project with the RPI workflow structure.

Targets:
  claude    Creates .claude/ with agents, commands, skills, hooks subdirectories
            and a CLAUDE.md rules file (default)
  opencode  Creates .opencode/ with the same structure and an AGENTS.md rules file

Also creates:
  .rpi/             Artifact directory hierarchy (research, designs, plans, etc.)
  .rpi/index.json   Codebase symbol index

Use --no-claude-md to skip rules file generation. Use --track-rpi to keep
.rpi/ tracked in git. Use "rpi update" to sync an existing project with
the latest workflow files.`,
	Example: `  # Initialize for Claude Code (default)
  rpi init

  # Initialize for OpenCode
  rpi init --target opencode

  # Initialize in a specific directory without rules file
  rpi init ./my-project --no-claude-md`,
	Args: cobra.MaximumNArgs(1),
	RunE: runInit,
}

func init() {
	initCmd.Flags().BoolVar(&initNoClaudeMD, "no-claude-md", false, "Skip rules file generation (CLAUDE.md or AGENTS.md)")
	initCmd.Flags().BoolVar(&initTrackRpi, "track-rpi", false, "Do not add .rpi/ to .gitignore")
	initCmd.Flags().StringVar(&initTarget, "target", "claude", `AI coding tool to initialize for: "claude" or "opencode"`)
	initCmd.Flags().BoolVar(&initNoMCP, "no-mcp", false, "Skip MCP server configuration")
	rootCmd.AddCommand(initCmd)
}

type targetConfig struct {
	toolDir   string // ".claude" or ".opencode"
	subdirs   []string
	rulesFile string // "CLAUDE.md" or "AGENTS.md"
	target    workflow.Target
}

func resolveTargetConfig(t string) (targetConfig, error) {
	switch t {
	case "claude":
		return targetConfig{
			toolDir:   ".claude",
			subdirs:   []string{"agents", "commands", "skills", "hooks"},
			rulesFile: "CLAUDE.md",
			target:    workflow.TargetClaude,
		}, nil
	case "opencode":
		return targetConfig{
			toolDir:   ".opencode",
			subdirs:   []string{"agents", "commands", "skills", "hooks"},
			rulesFile: "AGENTS.md",
			target:    workflow.TargetOpenCode,
		}, nil
	default:
		return targetConfig{}, fmt.Errorf("unknown target %q: must be \"claude\" or \"opencode\"", t)
	}
}

func runInit(cmd *cobra.Command, args []string) error {
	targetDir := "."
	if len(args) > 0 {
		targetDir = args[0]
	}

	cfg, err := resolveTargetConfig(initTarget)
	if err != nil {
		return err
	}

	w := cmd.OutOrStdout()

	toolDirPath := filepath.Join(targetDir, cfg.toolDir)

	// Check if already initialized
	if _, err := os.Stat(toolDirPath); err == nil {
		return fmt.Errorf("%s/ already exists; use rpi update to sync", cfg.toolDir)
	}

	// Create tool subdirs
	for _, d := range cfg.subdirs {
		path := filepath.Join(toolDirPath, d)
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("create %s: %w", path, err)
		}
		logSuccess(w, fmt.Sprintf("Created %s/%s/", cfg.toolDir, d))
	}

	// Create .rpi/ artifact subdirs
	rpiDir := filepath.Join(targetDir, ".rpi")
	rpiSubdirs := []string{
		"research", "designs",
		"plans", "specs", "reviews", "archive",
	}
	for _, d := range rpiSubdirs {
		path := filepath.Join(rpiDir, d)
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("create %s: %w", path, err)
		}
		logSuccess(w, fmt.Sprintf("Created .rpi/%s/", d))
	}

	// Generate rules file (CLAUDE.md or AGENTS.md)
	if !initNoClaudeMD {
		rulesPath := filepath.Join(targetDir, cfg.rulesFile)
		content, err := templates.Get(cfg.rulesFile)
		if err != nil {
			return fmt.Errorf("get %s template: %w", cfg.rulesFile, err)
		}
		if err := os.WriteFile(rulesPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("write %s: %w", cfg.rulesFile, err)
		}
		logSuccess(w, fmt.Sprintf("Created %s", cfg.rulesFile))
	}

	// Manage .gitignore
	if err := ensureGitignoreEntry(w, targetDir, cfg.toolDir+"/"); err != nil {
		logWarning(w, fmt.Sprintf("Failed to update .gitignore: %v", err))
	}

	// Install embedded workflow files (agents, commands, skills)
	n, err := workflow.InstallTo(toolDirPath, cfg.target, false)
	if err != nil {
		return fmt.Errorf("install workflow files: %w", err)
	}
	logSuccess(w, fmt.Sprintf("Installed %d workflow files to %s/ (agents, commands, skills)", n, cfg.toolDir))

	// Configure MCP server (Claude only)
	if !initNoMCP && cfg.target == workflow.TargetClaude {
		configureMCP(w, targetDir)
	}

	// Add .rpi/ to .gitignore (unless --track-rpi)
	if !initTrackRpi {
		if err := ensureGitignoreEntry(w, targetDir, ".rpi/"); err != nil {
			logWarning(w, fmt.Sprintf("Failed to update .gitignore: %v", err))
		}
	}

	// Build codebase index
	logInfo(w, "Building codebase index...")
	idx, err := index.Build(targetDir, index.BuildOptions{})
	if err != nil {
		logWarning(w, fmt.Sprintf("Index build failed: %v", err))
	} else {
		if mkErr := os.MkdirAll(rpiDir, 0755); mkErr != nil {
			logWarning(w, fmt.Sprintf("Create .rpi/ failed: %v", mkErr))
		} else {
			indexPath := filepath.Join(rpiDir, "index.json")
			if saveErr := index.Save(idx, indexPath); saveErr != nil {
				logWarning(w, fmt.Sprintf("Index save failed: %v", saveErr))
			} else {
				logSuccess(w, fmt.Sprintf("Built codebase index (%d files, %d symbols)", idx.Metadata.FileCount, idx.Metadata.SymbolCount))
			}

			// Generate CLI reference
			writeCLIReference(w, rpiDir)
		}
	}

	return nil
}

func writeCLIReference(w io.Writer, rpiDir string) {
	cliRef := generateCLIReference(rootCmd)
	cliRefPath := filepath.Join(rpiDir, "cli-reference.md")
	if err := os.WriteFile(cliRefPath, []byte(cliRef), 0644); err != nil {
		logWarning(w, fmt.Sprintf("CLI reference write failed: %v", err))
	} else {
		logSuccess(w, "Generated CLI reference")
	}
}

func ensureGitignoreEntry(w io.Writer, targetDir, entry string) error {
	gitignorePath := filepath.Join(targetDir, ".gitignore")

	// Check if entry already exists
	if data, err := os.ReadFile(gitignorePath); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			if line == entry {
				return nil // already present
			}
		}
	}

	f, err := os.OpenFile(gitignorePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open .gitignore: %w", err)
	}
	defer f.Close()

	if _, err := fmt.Fprintf(f, "\n# RPI workflow\n%s\n", entry); err != nil {
		return fmt.Errorf("write .gitignore: %w", err)
	}
	logSuccess(w, fmt.Sprintf("Added %s to .gitignore", entry))
	return nil
}

// mcpCommandRunner abstracts command execution for testing.
var mcpCommandRunner func(name string, args ...string) ([]byte, error) = func(name string, args ...string) ([]byte, error) {
	return exec.Command(name, args...).CombinedOutput()
}

func configureMCP(w io.Writer, _ string) {
	if _, err := exec.LookPath("claude"); err != nil {
		logWarning(w, "claude not found in PATH — skipping MCP server configuration")
		return
	}
	if _, err := exec.LookPath("rpi"); err != nil {
		logWarning(w, "rpi not found in PATH — skipping MCP server configuration")
		return
	}

	// Check if rpi MCP server is already registered
	if out, err := mcpCommandRunner("claude", "mcp", "get", "rpi"); err == nil {
		_ = out
		logWarning(w, "MCP server 'rpi' already configured")
		return
	}

	// Register the MCP server via claude CLI
	if out, err := mcpCommandRunner("claude", "mcp", "add", "rpi", "--", "rpi", "serve"); err != nil {
		logWarning(w, fmt.Sprintf("Failed to register MCP server: %s", strings.TrimSpace(string(out))))
		return
	}
	logSuccess(w, "Configured MCP server via claude mcp add")
}

// copyDirectory copies all files and subdirectories from src to dest.
// Returns the number of top-level items copied.
func copyDirectory(src, dest string) (int, error) {
	entries, err := os.ReadDir(src)
	if err != nil {
		return 0, err
	}
	if err := os.MkdirAll(dest, 0755); err != nil {
		return 0, err
	}
	count := 0
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		destPath := filepath.Join(dest, entry.Name())
		if entry.IsDir() {
			if err := copyDirRecursive(srcPath, destPath); err != nil {
				return count, err
			}
		} else {
			data, err := os.ReadFile(srcPath)
			if err != nil {
				return count, err
			}
			if err := os.WriteFile(destPath, data, 0644); err != nil {
				return count, err
			}
		}
		count++
	}
	return count, nil
}

func copyDirRecursive(src, dest string) error {
	if err := os.MkdirAll(dest, 0755); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		destPath := filepath.Join(dest, entry.Name())
		if entry.IsDir() {
			if err := copyDirRecursive(srcPath, destPath); err != nil {
				return err
			}
		} else {
			data, err := os.ReadFile(srcPath)
			if err != nil {
				return err
			}
			if err := os.WriteFile(destPath, data, 0644); err != nil {
				return err
			}
		}
	}
	return nil
}
