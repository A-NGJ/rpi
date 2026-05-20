package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// rpiAuthoredSkills is the canonical set of skill folder names that the
// standalone `rpi init` install drops under `~/.claude/skills/`. The
// uninstaller refuses to delete any rpi-prefixed skill directory whose
// basename is not in this set — that way a stray user-created folder
// named like `rpi-foo` is never collateral damage.
var rpiAuthoredSkills = map[string]bool{
	"rpi-archive":   true,
	"rpi-commit":    true,
	"rpi-diagnose":  true,
	"rpi-explain":   true,
	"rpi-handoff":   true,
	"rpi-implement": true,
	"rpi-plan":      true,
	"rpi-propose":   true,
	"rpi-research":  true,
	"rpi-spec-sync": true,
	"rpi-verify":    true,
}

// rpiHookMarkerSet enumerates the marker substrings RPI's configureHooks
// writes into each hook command. Any hook block in settings.json that
// contains one of these strings is treated as RPI-owned.
var rpiHookMarkerSet = []string{
	"rpi_context_essentials",
	"rpi_session_resume",
	"rpi_suggest_next",
	"claude-handoff",
}

const (
	installStateNothing    = "nothing"
	installStatePluginMode = "plugin-mode"
	installStateStandalone = "standalone"
)

// installState is a structured snapshot of the user's home directory
// describing what RPI artifacts were found. Both the detector and the
// remover share it.
type installState struct {
	home             string
	homeRpiPath      string
	pluginBinaryPath string
	settingsPath     string

	// Public booleans called out by the design.
	standaloneSkills     []string
	standaloneMCP        bool
	standaloneHooksPerms bool
	pluginBinary         bool
	homeRpiDir           bool

	// Internal book-keeping populated alongside the booleans, consumed
	// by the Phase 2 remover so the two halves stay aligned.
	settingsExists   bool
	rpiAllowPatterns []string
	rpiHookMarkers   []string
}

var (
	uninstallGlobal bool
	uninstallDryRun bool
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Remove a global RPI install (standalone or plugin-mode)",
	Long: `Remove a previously installed RPI from the user's home directory.

Behavior depends on what is detected on disk:

  Standalone mode  Skills under ~/.claude/skills/rpi-*, hooks and
                   permissions in ~/.claude/settings.json referencing the
                   rpi binary, or an mcpServers.rpi entry pointing at a
                   non-plugin path. Removes only the entries RPI installed.

  Plugin mode      Only ~/.rpi/bin/rpi is present (the plugin owns skills,
                   hooks, and MCP wiring inside its own directory).
                   Removes the binary and the empty ~/.rpi/ directory.

  Nothing          Exits cleanly with a "nothing to remove" message.

Use --dry-run to preview the deletion plan without touching the
filesystem.`,
	Example: `  # Preview what would be removed
  rpi uninstall --global --dry-run

  # Remove a standalone install (or fall through to nothing-to-remove)
  rpi uninstall --global`,
	Args: cobra.NoArgs,
	RunE: runUninstall,
}

func init() {
	uninstallCmd.Flags().BoolVar(&uninstallGlobal, "global", false, "Required: remove the user-level install (~/.claude, ~/.rpi)")
	uninstallCmd.Flags().BoolVar(&uninstallDryRun, "dry-run", false, "Print the deletion plan and exit without touching the filesystem")
	rootCmd.AddCommand(uninstallCmd)
}

func runUninstall(cmd *cobra.Command, _ []string) error {
	if !uninstallGlobal {
		return errors.New("rpi uninstall currently requires --global (per-project uninstall is not supported)")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("resolve home dir: %w", err)
	}
	w := cmd.OutOrStdout()

	state, err := detectInstallState(home)
	if err != nil {
		return err
	}
	printUninstallPlan(w, state)

	if uninstallDryRun {
		return nil
	}
	// Phase 2 will dispatch to removeStandalone / removePluginMode here.
	return nil
}

// detectInstallState inspects the user's home directory for evidence of
// either a standalone `rpi init --global` install or a plugin-mode binary
// install, and returns a snapshot the caller can use to decide what to
// remove. The function is side-effect free.
func detectInstallState(home string) (installState, error) {
	s := installState{
		home:             home,
		homeRpiPath:      filepath.Join(home, ".rpi"),
		pluginBinaryPath: filepath.Join(home, ".rpi", "bin", "rpi"),
		settingsPath:     filepath.Join(home, ".claude", "settings.json"),
	}

	// Standalone skills: only rpi-prefixed folders that pass the
	// authorship marker check are candidates for removal.
	skillsRoot := filepath.Join(home, ".claude", "skills")
	if entries, err := os.ReadDir(skillsRoot); err == nil {
		for _, e := range entries {
			if !e.IsDir() || !strings.HasPrefix(e.Name(), "rpi-") {
				continue
			}
			dir := filepath.Join(skillsRoot, e.Name())
			if isRPIAuthoredSkill(e.Name(), dir) {
				s.standaloneSkills = append(s.standaloneSkills, dir)
			}
		}
	}

	// settings.json scan — mcpServers.rpi (standalone), Bash(rpi *)
	// permissions, and RPI-owned hook markers.
	if data, readErr := os.ReadFile(s.settingsPath); readErr == nil {
		s.settingsExists = true
		var raw map[string]json.RawMessage
		if err := json.Unmarshal(data, &raw); err == nil {
			if mcpRaw, ok := raw["mcpServers"]; ok {
				var servers map[string]map[string]any
				if json.Unmarshal(mcpRaw, &servers) == nil {
					if entry, ok := servers["rpi"]; ok {
						cmd, _ := entry["command"].(string)
						if !isPluginBinaryPath(cmd, s.pluginBinaryPath) {
							s.standaloneMCP = true
						}
					}
				}
			}
			if permsRaw, ok := raw["permissions"]; ok {
				var perms map[string]json.RawMessage
				if json.Unmarshal(permsRaw, &perms) == nil {
					if allowRaw, ok := perms["allow"]; ok {
						var allow []string
						if json.Unmarshal(allowRaw, &allow) == nil {
							for _, p := range allow {
								if isRPIPermission(p) {
									s.rpiAllowPatterns = append(s.rpiAllowPatterns, p)
								}
							}
						}
					}
				}
			}
			if hooksRaw, ok := raw["hooks"]; ok {
				s.rpiHookMarkers = findRPIHookMarkers(string(hooksRaw))
			}
		}
		if len(s.rpiAllowPatterns) > 0 || len(s.rpiHookMarkers) > 0 {
			s.standaloneHooksPerms = true
		}
	}

	// ~/.rpi/ presence checks.
	if info, statErr := os.Stat(s.pluginBinaryPath); statErr == nil && !info.IsDir() {
		if info.Mode()&0111 != 0 {
			s.pluginBinary = true
		}
	}
	if info, statErr := os.Stat(s.homeRpiPath); statErr == nil && info.IsDir() {
		s.homeRpiDir = true
	}

	return s, nil
}

// classification reduces the structured snapshot to one of three
// labels — standalone wins over plugin-mode when both are present so we
// never silently delete user-owned config behind a plugin-mode banner.
func (s installState) classification() string {
	if len(s.standaloneSkills) > 0 || s.standaloneMCP || s.standaloneHooksPerms {
		return installStateStandalone
	}
	if s.pluginBinary {
		return installStatePluginMode
	}
	return installStateNothing
}

func printUninstallPlan(w io.Writer, s installState) {
	switch s.classification() {
	case installStateNothing:
		fmt.Fprintln(w, "Nothing to remove.")
	case installStatePluginMode:
		fmt.Fprintln(w, "RPI plugin-mode install detected. Plan:")
		fmt.Fprintf(w, "  - remove %s\n", s.pluginBinaryPath)
		if s.homeRpiDir {
			fmt.Fprintf(w, "  - remove %s (if empty)\n", s.homeRpiPath)
		}
	case installStateStandalone:
		fmt.Fprintln(w, "RPI standalone install detected. Plan:")
		for _, d := range s.standaloneSkills {
			fmt.Fprintf(w, "  - remove skill dir %s\n", d)
		}
		if s.standaloneMCP {
			fmt.Fprintf(w, "  - drop mcpServers.rpi from %s\n", s.settingsPath)
		}
		if len(s.rpiAllowPatterns) > 0 {
			fmt.Fprintf(w, "  - drop %d permission entries from %s\n", len(s.rpiAllowPatterns), s.settingsPath)
		}
		if len(s.rpiHookMarkers) > 0 {
			fmt.Fprintf(w, "  - drop %d RPI hook entries from %s\n", len(s.rpiHookMarkers), s.settingsPath)
		}
		if s.pluginBinary {
			fmt.Fprintf(w, "  - remove %s\n", s.pluginBinaryPath)
		}
		if s.homeRpiDir {
			fmt.Fprintf(w, "  - remove %s (if empty)\n", s.homeRpiPath)
		}
	}
}

// isRPIAuthoredSkill is the authorship marker: the folder basename must
// be in the canonical RPI skill set AND the SKILL.md inside must declare
// a matching `name:` field in its YAML frontmatter. Both checks must
// pass — a folder named rpi-research without the SKILL.md frontmatter
// match is treated as user-owned.
func isRPIAuthoredSkill(name, dir string) bool {
	if !rpiAuthoredSkills[name] {
		return false
	}
	data, err := os.ReadFile(filepath.Join(dir, "SKILL.md"))
	if err != nil {
		return false
	}
	return frontmatterNameMatches(data, name)
}

func frontmatterNameMatches(data []byte, name string) bool {
	content := string(data)
	if !strings.HasPrefix(content, "---") {
		return false
	}
	end := strings.Index(content[3:], "\n---")
	if end < 0 {
		return false
	}
	expected := "name: " + name
	for _, line := range strings.Split(content[3:3+end], "\n") {
		if strings.TrimSpace(line) == expected {
			return true
		}
	}
	return false
}

// isRPIPermission identifies permission allow-list entries that the
// standalone install would have written. Keep this aligned with
// safeBashPatterns and the MCP wildcard so the remover doesn't strip
// user-added entries.
func isRPIPermission(entry string) bool {
	switch {
	case entry == "mcp__rpi__*":
		return true
	case strings.HasPrefix(entry, "Bash(rpi "):
		return true
	case entry == "Bash(rm /tmp/claude-handoff-*.md)":
		return true
	}
	return false
}

func findRPIHookMarkers(s string) []string {
	var found []string
	for _, m := range rpiHookMarkerSet {
		if strings.Contains(s, m) {
			found = append(found, m)
		}
	}
	return found
}

// isPluginBinaryPath returns true when the MCP command string points at
// the well-known plugin binary location, so the detector can tell apart
// a standalone-registered server from one the plugin manages. Tilde-
// prefixed forms expand against the current $HOME.
func isPluginBinaryPath(cmd, plugin string) bool {
	if cmd == "" {
		return false
	}
	if cmd == plugin {
		return true
	}
	if strings.HasPrefix(cmd, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, cmd[2:]) == plugin
	}
	return false
}
