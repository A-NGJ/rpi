package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/A-NGJ/rpi/internal/workflow"
	"github.com/spf13/cobra"
)

var (
	initNoClaudeMD bool
	initNoTrack    bool
	initTarget     string
	initNoMCP      bool
)

const (
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

var initCmd = &cobra.Command{
	Use:   "init [directory]",
	Short: "Initialize project with workflow skills and rules file",
	Long: `Initialize a project with the RPI workflow structure.

Targets:
  claude      Creates .claude/ with skills and hooks subdirectories
              and a CLAUDE.md rules file (default)
  opencode    Creates .opencode/ with the same structure and an AGENTS.md rules file
  agents-only Creates .agents/skills/ with cross-tool Agent Skills — no tool-specific
              directory, no MCP config

Also creates:
  .rpi/             Artifact directory hierarchy (research, designs, plans, etc.)

Use --no-claude-md to skip rules file generation. Use --no-track to add
.rpi/ to .gitignore (by default, .rpi/ is tracked in git). Use "rpi update"
to sync an existing project with the latest workflow files.`,
	Example: `  # Initialize for Claude Code (default)
  rpi init

  # Initialize for OpenCode
  rpi init --target opencode

  # Initialize cross-tool skills only (no tool-specific setup)
  rpi init --target agents-only

  # Initialize in a specific directory without rules file
  rpi init ./my-project --no-claude-md`,
	Args: cobra.MaximumNArgs(1),
	RunE: runInit,
}

func init() {
	initCmd.Flags().BoolVar(&initNoClaudeMD, "no-claude-md", false, "Skip rules file generation (CLAUDE.md or AGENTS.md)")
	initCmd.Flags().BoolVar(&initNoTrack, "no-track", false, "Add .rpi/ to .gitignore (artifacts not tracked in git)")
	initCmd.Flags().StringVar(&initTarget, "target", "claude", `AI coding tool to initialize for: "claude", "opencode", or "agents-only"`)
	initCmd.Flags().BoolVar(&initNoMCP, "no-mcp", false, "Skip MCP server configuration")
	rootCmd.AddCommand(initCmd)
}

type targetConfig struct {
	toolDir   string // ".claude", ".opencode", or "" for agents-only
	subdirs   []string
	rulesFile string // "CLAUDE.md", "AGENTS.md", or "" for agents-only
	target    workflow.Target
}

func resolveTargetConfig(t string) (targetConfig, error) {
	switch t {
	case "claude":
		return targetConfig{
			toolDir:   ".claude",
			subdirs:   []string{"skills", "hooks"},
			rulesFile: "CLAUDE.md",
			target:    workflow.TargetClaude,
		}, nil
	case "opencode":
		return targetConfig{
			toolDir:   ".opencode",
			subdirs:   []string{"skills", "hooks"},
			rulesFile: "AGENTS.md",
			target:    workflow.TargetOpenCode,
		}, nil
	case "agents-only":
		return targetConfig{
			toolDir:   "",
			subdirs:   nil,
			rulesFile: "",
			target:    workflow.TargetAgentsOnly,
		}, nil
	default:
		return targetConfig{}, fmt.Errorf("unknown target %q: must be \"claude\", \"opencode\", or \"agents-only\"", t)
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

	// Guard: fail if tool dir already exists
	if cfg.target == workflow.TargetAgentsOnly {
		agentsPath := filepath.Join(targetDir, ".agents")
		if _, err := os.Stat(agentsPath); err == nil {
			return fmt.Errorf(".agents/ already exists; use rpi update to sync")
		}
	} else {
		toolDirPath := filepath.Join(targetDir, cfg.toolDir)
		if _, err := os.Stat(toolDirPath); err == nil {
			return fmt.Errorf("%s/ already exists; use rpi update to sync", cfg.toolDir)
		}
	}

	// Create tool subdirs or .agents/skills/ (first-time creation)
	if cfg.toolDir != "" {
		toolDirPath := filepath.Join(targetDir, cfg.toolDir)
		for _, d := range cfg.subdirs {
			path := filepath.Join(toolDirPath, d)
			if err := os.MkdirAll(path, 0755); err != nil {
				return fmt.Errorf("create %s: %w", path, err)
			}
			logSuccess(w, fmt.Sprintf("Created %s/%s/", cfg.toolDir, d))
		}
	} else {
		agentsSkillsDir := filepath.Join(targetDir, ".agents", "skills")
		if err := os.MkdirAll(agentsSkillsDir, 0755); err != nil {
			return fmt.Errorf("create .agents/skills/: %w", err)
		}
		logSuccess(w, "Created .agents/skills/")
	}

	// Manage .gitignore for tool dir (skip for agents-only)
	if cfg.toolDir != "" {
		if err := ensureGitignoreEntry(w, targetDir, cfg.toolDir+"/"); err != nil {
			logWarning(w, fmt.Sprintf("Failed to update .gitignore: %v", err))
		}
	}

	// Configure MCP server (Claude only)
	if !initNoMCP && cfg.target == workflow.TargetClaude {
		configureMCP(w, targetDir)
	}

	// Add .rpi/ to .gitignore only if --no-track is set
	if initNoTrack {
		if err := ensureGitignoreEntry(w, targetDir, ".rpi/"); err != nil {
			logWarning(w, fmt.Sprintf("Failed to update .gitignore: %v", err))
		}
	}

	// Sync shared project structure (dirs, skills, templates, rules, settings)
	return syncProject(syncOptions{
		targetDir: targetDir,
		cfg:       cfg,
		force:     false,
		skipRules: initNoClaudeMD,
		w:         w,
	})
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

// configureSettings ensures .claude/settings.json contains the mcp__rpi__* permission.
// It merges into any existing settings file rather than overwriting.
func configureSettings(w io.Writer, toolDirPath string) {
	settingsPath := filepath.Join(toolDirPath, "settings.json")

	type settingsFile struct {
		Permissions map[string][]string `json:"permissions,omitempty"`
		Extra       map[string]json.RawMessage
	}

	// Read existing settings (if any)
	raw := make(map[string]json.RawMessage)
	if data, err := os.ReadFile(settingsPath); err == nil {
		if err := json.Unmarshal(data, &raw); err != nil {
			logWarning(w, fmt.Sprintf("Failed to parse %s: %v", settingsPath, err))
			return
		}
	}

	// Parse permissions
	var allow []string
	if permsRaw, ok := raw["permissions"]; ok {
		var perms map[string][]string
		if err := json.Unmarshal(permsRaw, &perms); err == nil {
			allow = perms["allow"]
		}
	}

	// Check if mcp__rpi__* already present
	const rpiPattern = "mcp__rpi__*"
	for _, entry := range allow {
		if entry == rpiPattern {
			return // already configured
		}
	}

	// Add the pattern
	allow = append(allow, rpiPattern)
	permsMap := map[string][]string{"allow": allow}
	permsJSON, _ := json.Marshal(permsMap)
	raw["permissions"] = json.RawMessage(permsJSON)

	// Write back with indentation
	out, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		logWarning(w, fmt.Sprintf("Failed to marshal settings: %v", err))
		return
	}
	if err := os.WriteFile(settingsPath, append(out, '\n'), 0644); err != nil {
		logWarning(w, fmt.Sprintf("Failed to write %s: %v", settingsPath, err))
		return
	}
	logSuccess(w, "Configured auto-allow for RPI MCP tools in settings.json")
}

// configureHooks ensures .claude/settings.json contains a PostCompact hook
// that reminds the assistant to call rpi_context_essentials after compaction.
// It merges into any existing settings/hooks rather than overwriting.
func configureHooks(w io.Writer, toolDirPath string) {
	settingsPath := filepath.Join(toolDirPath, "settings.json")

	// Read existing settings (if any)
	raw := make(map[string]json.RawMessage)
	if data, err := os.ReadFile(settingsPath); err == nil {
		if err := json.Unmarshal(data, &raw); err != nil {
			logWarning(w, fmt.Sprintf("Failed to parse %s: %v", settingsPath, err))
			return
		}
	}

	// Parse existing hooks
	hooks := make(map[string]json.RawMessage)
	if hooksRaw, ok := raw["hooks"]; ok {
		if err := json.Unmarshal(hooksRaw, &hooks); err != nil {
			logWarning(w, fmt.Sprintf("Failed to parse hooks in %s: %v", settingsPath, err))
			return
		}
	}

	// Check if PostCompact already has our entry
	const marker = "rpi_context_essentials"
	if pcRaw, ok := hooks["PostCompact"]; ok {
		if strings.Contains(string(pcRaw), marker) {
			return // already configured
		}
	}

	// Build the hook entry
	hookEntry := map[string]string{
		"type":    "command",
		"command": "cat <<'HOOK_EOF'\nIMPORTANT: Context was compacted. Call the rpi_context_essentials MCP tool to restore your implementation context (active plan phase, spec scenarios, constraints).\nHOOK_EOF",
	}

	// Append to existing PostCompact hooks or create new array
	var postCompact []json.RawMessage
	if pcRaw, ok := hooks["PostCompact"]; ok {
		json.Unmarshal(pcRaw, &postCompact)
	}
	entryJSON, _ := json.Marshal(hookEntry)
	postCompact = append(postCompact, json.RawMessage(entryJSON))

	pcJSON, _ := json.Marshal(postCompact)
	hooks["PostCompact"] = json.RawMessage(pcJSON)

	hooksJSON, _ := json.Marshal(hooks)
	raw["hooks"] = json.RawMessage(hooksJSON)

	// Write back with indentation
	out, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		logWarning(w, fmt.Sprintf("Failed to marshal settings: %v", err))
		return
	}
	if err := os.WriteFile(settingsPath, append(out, '\n'), 0644); err != nil {
		logWarning(w, fmt.Sprintf("Failed to write %s: %v", settingsPath, err))
		return
	}
	logSuccess(w, "Configured PostCompact hook for context preservation in settings.json")
}
