package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/A-NGJ/rpi/internal/workflow"
	"github.com/spf13/cobra"
)

var bootstrapCmd = &cobra.Command{
	Use:   "bootstrap",
	Short: "Auto-initialize the current project's .rpi/ tree if a global install is present",
	Long: `Silent, idempotent project bootstrap.

Run from the top of every rpi-* skill prompt (and safe to invoke manually).
Performs a lite project init — creating .rpi/, the rules file, and
.gitignore entries at the git root — but only when:

  - the project is not already initialized (no .rpi/ found in cwd or
    any ancestor), AND
  - a user-level install is present at ~/.claude/skills/rpi-research/
    or ~/.config/opencode/skills/rpi-research/, AND
  - the current directory is inside a git repository.

In any no-op path the command exits silently with status 0. On the
success path, exactly one line is written to stderr:

  ✓ Auto-initialized .rpi/ in <git-root> — skills inherited from <global-path>

Skills, agents, MCP server registration, and settings.json hooks are
inherited from the global install — bootstrap never duplicates them
into the project.`,
	Args: cobra.NoArgs,
	RunE: runBootstrap,
}

func init() {
	rootCmd.AddCommand(bootstrapCmd)
}

func runBootstrap(cmd *cobra.Command, _ []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getwd: %w", err)
	}

	// 1. Already initialized? Walk up looking for .rpi/.
	if findAncestor(cwd, ".rpi") != "" {
		return nil
	}

	// 2. Global install present?
	home, err := os.UserHomeDir()
	if err != nil {
		// Without HOME we cannot detect a global install — silent no-op
		// to match the design's "no surprises" contract.
		return nil
	}
	globalPath, target, ok := detectGlobalInstall(home)
	if !ok {
		return nil
	}

	// 3. Inside a git repo?
	gitRoot := findAncestor(cwd, ".git")
	if gitRoot == "" {
		return nil
	}

	// 4. Lite-init at the git root.
	cfg, cfgErr := resolveTargetConfig(string(target))
	if cfgErr != nil {
		return cfgErr
	}
	if err := liteSyncProject(syncOptions{
		targetDir: gitRoot,
		cfg:       cfg,
		w:         io.Discard,
	}); err != nil {
		return err
	}

	fmt.Fprintf(cmd.ErrOrStderr(), "✓ Auto-initialized .rpi/ in %s — skills inherited from %s\n", gitRoot, globalPath)
	return nil
}

// findAncestor walks up from start looking for a directory containing the
// given child entry. Returns the first ancestor path that contains it, or
// "" if none found before reaching the filesystem root.
func findAncestor(start, child string) string {
	dir := start
	for {
		if _, err := os.Stat(filepath.Join(dir, child)); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}

// detectGlobalInstall probes the canonical user-level install paths and
// returns the install root and the target it represents. Claude is
// preferred when both are present.
func detectGlobalInstall(home string) (path string, target workflow.Target, ok bool) {
	claudePath := filepath.Join(home, ".claude")
	if _, err := os.Stat(filepath.Join(claudePath, "skills", "rpi-research")); err == nil {
		return claudePath, workflow.TargetClaude, true
	}
	opencodePath := filepath.Join(home, ".config", "opencode")
	if _, err := os.Stat(filepath.Join(opencodePath, "skills", "rpi-research")); err == nil {
		return opencodePath, workflow.TargetOpenCode, true
	}
	return "", "", false
}
