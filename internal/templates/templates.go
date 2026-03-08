package templates

import (
	_ "embed"
	"fmt"
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

// Get returns the template content for the given name.
// Returns an error if the name is not recognized.
func Get(name string) (string, error) {
	content, ok := knownTemplates[name]
	if !ok {
		return "", fmt.Errorf("unknown template: %s", name)
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
