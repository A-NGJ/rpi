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
	"rpi-blueprint": true,
	"rpi-commit":    true,
	"rpi-diagnose":  true,
	"rpi-explain":   true,
	"rpi-handoff":   true,
	"rpi-implement": true,
	"rpi-plan":      true,
	"rpi-propose":   true,
	"rpi-research":  true,
	"rpi-revise":    true,
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
	switch state.classification() {
	case installStateNothing:
		return nil
	case installStatePluginMode:
		return removePluginMode(w, state)
	case installStateStandalone:
		return removeStandalone(w, state)
	}
	return nil
}

// removeStandalone executes the full standalone uninstall: skill
// directories, RPI-owned keys in settings.json, then the home-rooted
// binary/dir if originally present.
func removeStandalone(w io.Writer, s installState) error {
	for _, dir := range s.standaloneSkills {
		if err := os.RemoveAll(dir); err != nil {
			return fmt.Errorf("remove %s: %w", dir, err)
		}
		fmt.Fprintf(w, "removed skill dir %s\n", dir)
	}

	if s.settingsExists && (s.standaloneMCP || len(s.rpiAllowPatterns) > 0 || len(s.rpiHookMarkers) > 0) {
		if err := scrubSettings(w, s); err != nil {
			return err
		}
	}

	if s.homeRpiDir {
		if err := removeHomeRpi(w, s); err != nil {
			return err
		}
	}
	return nil
}

// removePluginMode deletes ~/.rpi/bin/rpi and ~/.rpi/ if empty. The
// plugin's own files live elsewhere (Claude Code's plugin cache) and
// are untouched.
func removePluginMode(w io.Writer, s installState) error {
	return removeHomeRpi(w, s)
}

// removeHomeRpi removes the binary, then bin/, then ~/.rpi/ — each
// step is best-effort and only proceeds when the directory is empty.
func removeHomeRpi(w io.Writer, s installState) error {
	if _, err := os.Stat(s.pluginBinaryPath); err == nil {
		if err := os.Remove(s.pluginBinaryPath); err != nil {
			return fmt.Errorf("remove %s: %w", s.pluginBinaryPath, err)
		}
		fmt.Fprintf(w, "removed %s\n", s.pluginBinaryPath)
	}
	binDir := filepath.Dir(s.pluginBinaryPath)
	if isDirEmpty(binDir) {
		_ = os.Remove(binDir)
	}
	if isDirEmpty(s.homeRpiPath) {
		if err := os.Remove(s.homeRpiPath); err != nil {
			return fmt.Errorf("remove %s: %w", s.homeRpiPath, err)
		}
		fmt.Fprintf(w, "removed %s\n", s.homeRpiPath)
	}
	return nil
}

func isDirEmpty(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	return len(entries) == 0
}

// scrubSettings rewrites ~/.claude/settings.json in place, removing
// only RPI-owned keys (mcpServers.rpi when pointed at a non-plugin
// path, Bash(rpi …) and mcp__rpi__* permissions, and hook entries
// whose command bodies contain a known RPI marker). Unrelated keys
// are preserved verbatim — the file is parsed and rewritten, never
// regenerated from scratch.
func scrubSettings(w io.Writer, s installState) error {
	data, err := os.ReadFile(s.settingsPath)
	if err != nil {
		return fmt.Errorf("read settings: %w", err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("parse settings: %w", err)
	}
	changed := false

	// mcpServers.rpi — only drop when the command is not the plugin path.
	if mcpRaw, ok := raw["mcpServers"]; ok {
		var servers map[string]json.RawMessage
		if err := json.Unmarshal(mcpRaw, &servers); err == nil {
			if entryRaw, has := servers["rpi"]; has {
				var entry map[string]any
				_ = json.Unmarshal(entryRaw, &entry)
				cmd, _ := entry["command"].(string)
				if !isPluginBinaryPath(cmd, s.pluginBinaryPath) {
					delete(servers, "rpi")
					if len(servers) == 0 {
						delete(raw, "mcpServers")
					} else {
						out, _ := json.Marshal(servers)
						raw["mcpServers"] = out
					}
					fmt.Fprintf(w, "dropped mcpServers.rpi from %s\n", s.settingsPath)
					changed = true
				}
			}
		}
	}

	// permissions.allow — drop only RPI-owned patterns.
	if permsRaw, ok := raw["permissions"]; ok {
		var perms map[string]json.RawMessage
		if err := json.Unmarshal(permsRaw, &perms); err == nil {
			if allowRaw, ok := perms["allow"]; ok {
				var allow []string
				if err := json.Unmarshal(allowRaw, &allow); err == nil {
					var kept []string
					removed := 0
					for _, entry := range allow {
						if isRPIPermission(entry) {
							removed++
							continue
						}
						kept = append(kept, entry)
					}
					if removed > 0 {
						if len(kept) == 0 {
							delete(perms, "allow")
						} else {
							out, _ := json.Marshal(kept)
							perms["allow"] = out
						}
						if len(perms) == 0 {
							delete(raw, "permissions")
						} else {
							out, _ := json.Marshal(perms)
							raw["permissions"] = out
						}
						fmt.Fprintf(w, "dropped %d RPI permission entries from %s\n", removed, s.settingsPath)
						changed = true
					}
				}
			}
		}
	}

	// hooks — drop matcher entries whose hook body contains an RPI marker.
	if hooksRaw, ok := raw["hooks"]; ok {
		var hooks map[string]json.RawMessage
		if err := json.Unmarshal(hooksRaw, &hooks); err == nil {
			removed := 0
			for event, eventRaw := range hooks {
				var entries []json.RawMessage
				if err := json.Unmarshal(eventRaw, &entries); err != nil {
					continue
				}
				kept := make([]json.RawMessage, 0, len(entries))
				for _, e := range entries {
					if containsRPIHookMarker(string(e)) {
						removed++
						continue
					}
					kept = append(kept, e)
				}
				if len(kept) == 0 {
					delete(hooks, event)
				} else {
					out, _ := json.Marshal(kept)
					hooks[event] = out
				}
			}
			if removed > 0 {
				if len(hooks) == 0 {
					delete(raw, "hooks")
				} else {
					out, _ := json.Marshal(hooks)
					raw["hooks"] = out
				}
				fmt.Fprintf(w, "dropped %d RPI hook entries from %s\n", removed, s.settingsPath)
				changed = true
			}
		}
	}

	if !changed {
		return nil
	}

	out, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal settings: %w", err)
	}
	if err := os.WriteFile(s.settingsPath, append(out, '\n'), 0644); err != nil {
		return fmt.Errorf("write settings: %w", err)
	}
	return nil
}

func containsRPIHookMarker(s string) bool {
	for _, m := range rpiHookMarkerSet {
		if strings.Contains(s, m) {
			return true
		}
	}
	return false
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
