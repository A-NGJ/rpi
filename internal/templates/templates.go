package templates

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/A-NGJ/rpi/internal/workflow"
)

// knownTemplates maps user-facing names to their asset paths inside workflow/assets/.
var knownTemplates = map[string]string{
	"AGENTS.md": "templates/AGENTS.md.template",
	"CLAUDE.md": "templates/CLAUDE.md.template",
}

const contractSlot = "<!-- rpi:contract:slot -->"

// isRulesFile reports whether the named template is one of the project rules
// files that receive the spliced RPI Skill Contract block.
func isRulesFile(name string) bool {
	return name == "CLAUDE.md" || name == "AGENTS.md"
}

// UserTemplatesDir returns the path to ~/.rpi/templates/.
func UserTemplatesDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return filepath.Join(home, ".rpi", "templates"), nil
}

// EnsureUserTemplates creates ~/.rpi/templates/ and copies embedded
// defaults for any templates not already present.
func EnsureUserTemplates() error {
	dir, err := UserTemplatesDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create templates dir: %w", err)
	}
	for name, assetPath := range knownTemplates {
		path := filepath.Join(dir, name+".template")
		if _, err := os.Stat(path); err == nil {
			continue // user file exists, don't overwrite
		}
		content, err := workflow.ReadAsset(assetPath)
		if err != nil {
			return fmt.Errorf("read embedded %s: %w", name, err)
		}
		if err := os.WriteFile(path, content, 0644); err != nil {
			return fmt.Errorf("write %s: %w", path, err)
		}
	}
	return nil
}

// Get returns the template content for the given name.
// Prefers user-customized version from ~/.rpi/templates/ if it exists,
// falls back to embedded default.
// Returns an error if the name is not recognized.
//
// For rules-file templates (CLAUDE.md, AGENTS.md), any `<!-- rpi:contract:slot -->`
// placeholder is replaced with the rendered RPI Skill Contract block. Legacy
// user templates without the placeholder are returned verbatim — the
// fence-aware writer handles the missing-block case on update.
func Get(name string) (string, error) {
	assetPath, ok := knownTemplates[name]
	if !ok {
		return "", fmt.Errorf("unknown template: %s", name)
	}

	var content string

	// Prefer user-customized version
	dir, err := UserTemplatesDir()
	if err == nil {
		userPath := filepath.Join(dir, name+".template")
		if data, err := os.ReadFile(userPath); err == nil {
			content = string(data)
		}
	}

	if content == "" {
		// Fall back to embedded
		data, err := workflow.ReadAsset(assetPath)
		if err != nil {
			return "", fmt.Errorf("read embedded %s: %w", name, err)
		}
		content = string(data)
	}

	if isRulesFile(name) && strings.Contains(content, contractSlot) {
		block, err := workflow.ReadAsset("contracts/rpi-contract.md.template")
		if err != nil {
			return "", fmt.Errorf("read embedded contract: %w", err)
		}
		blockStr := strings.TrimRight(string(block), "\n")
		content = strings.Replace(content, contractSlot, blockStr, 1)
	}

	return content, nil
}

// Names returns all available template names in sorted order.
func Names() []string {
	names := make([]string, 0, len(knownTemplates))
	for name := range knownTemplates {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
