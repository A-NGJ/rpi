package templates

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

//go:embed CLAUDE.md.template
var claudeTemplate string

//go:embed PIPELINE.md.template
var pipelineTemplate string

// knownTemplates maps user-facing names to embedded content.
var knownTemplates map[string]string

func init() {
	knownTemplates = map[string]string{
		"CLAUDE.md":   claudeTemplate,
		"PIPELINE.md": pipelineTemplate,
	}
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
	for name, content := range knownTemplates {
		path := filepath.Join(dir, name+".template")
		if _, err := os.Stat(path); err == nil {
			continue // user file exists, don't overwrite
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			return fmt.Errorf("write %s: %w", path, err)
		}
	}
	return nil
}

// Get returns the template content for the given name.
// Prefers user-customized version from ~/.rpi/templates/ if it exists,
// falls back to embedded default.
// Returns an error if the name is not recognized.
func Get(name string) (string, error) {
	_, ok := knownTemplates[name]
	if !ok {
		return "", fmt.Errorf("unknown template: %s", name)
	}

	// Prefer user-customized version
	dir, err := UserTemplatesDir()
	if err == nil {
		userPath := filepath.Join(dir, name+".template")
		if data, err := os.ReadFile(userPath); err == nil {
			return string(data), nil
		}
	}

	// Fall back to embedded
	return knownTemplates[name], nil
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
