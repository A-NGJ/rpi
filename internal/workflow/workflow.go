package workflow

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Target identifies which AI coding tool the workflow assets are installed for.
type Target string

const (
	TargetClaude     Target = "claude"
	TargetOpenCode   Target = "opencode"
	TargetAgentsOnly Target = "agents-only"
)

// modelMap translates short model aliases used in Claude Code assets to full
// provider-qualified IDs used by OpenCode.
var modelMap = map[string]string{
	"opus":   "anthropic/claude-opus-4-6",
	"sonnet": "anthropic/claude-sonnet-4-6",
	"haiku":  "anthropic/claude-haiku-4-5-20251001",
}

// skillOverrides defines per-skill tool-specific frontmatter fields to inject
// when installing to a tool-specific directory. Skills not listed here get no
// extra fields (model defaults to inherit for Claude, omitted for OpenCode).
var skillOverrides = map[string]map[string]string{
	"rpi-archive": {"model": "haiku", "disable-model-invocation": "true"},
	"rpi-commit":  {"model": "haiku"},
}

//go:embed all:assets
var assets embed.FS

// ReadAsset reads an embedded file from the assets directory.
// The path should be relative to assets/ (e.g., "templates/CLAUDE.md.template").
func ReadAsset(path string) ([]byte, error) {
	return assets.ReadFile("assets/" + path)
}

// Install copies all embedded workflow files into the target .claude/ directory.
// Existing files are only overwritten when force is true.
func Install(claudeDir string, force bool) (int, error) {
	return InstallTo(claudeDir, TargetClaude, force)
}

// InstallTo copies embedded non-skill assets (templates) into targetDir.
// Skills are installed separately via InstallSkills.
// Existing files are only overwritten when force is true.
func InstallTo(targetDir string, _ Target, force bool) (int, error) {
	count := 0
	err := fs.WalkDir(assets, "assets", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel("assets", path)

		// Skip skills — they are installed via InstallSkills.
		if rel == "skills" || strings.HasPrefix(rel, "skills/") || strings.HasPrefix(rel, "skills\\") {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		dest := filepath.Join(targetDir, rel)

		if d.IsDir() {
			return os.MkdirAll(dest, 0755)
		}

		if _, err := os.Stat(dest); err == nil && !force {
			return nil
		}

		data, err := assets.ReadFile(path)
		if err != nil {
			return err
		}

		if err := os.WriteFile(dest, data, 0644); err != nil {
			return err
		}
		count++
		return nil
	})
	return count, err
}

// InstallSkills copies embedded skills to both the cross-tool directory
// (.agents/skills/) and the tool-specific directory (<toolDir>/skills/).
// Canonical SKILL.md files (name+description only) go to agentsDir.
// Enriched copies with tool-specific frontmatter go to toolDir/skills/.
// For TargetAgentsOnly, only the canonical copy is installed.
// Existing files are only overwritten when force is true.
func InstallSkills(agentsDir, toolDir string, target Target, force bool) (int, error) {
	count := 0

	entries, err := fs.ReadDir(assets, "assets/skills")
	if err != nil {
		return 0, fmt.Errorf("read embedded skills: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		skillName := entry.Name()
		srcPath := "assets/skills/" + skillName + "/SKILL.md"

		data, err := assets.ReadFile(srcPath)
		if err != nil {
			return count, fmt.Errorf("read %s: %w", srcPath, err)
		}

		// Install canonical copy to .agents/skills/<name>/SKILL.md
		canonDest := filepath.Join(agentsDir, skillName, "SKILL.md")
		if err := os.MkdirAll(filepath.Dir(canonDest), 0755); err != nil {
			return count, fmt.Errorf("create %s: %w", filepath.Dir(canonDest), err)
		}
		if _, statErr := os.Stat(canonDest); statErr != nil || force {
			if err := os.WriteFile(canonDest, data, 0644); err != nil {
				return count, fmt.Errorf("write %s: %w", canonDest, err)
			}
			count++
		}

		// Install enriched copy to <toolDir>/skills/<name>/SKILL.md
		if target != TargetAgentsOnly && toolDir != "" {
			enriched := data
			if overrides, ok := skillOverrides[skillName]; ok {
				enriched = injectFrontmatter(data, overrides, target)
			}

			toolDest := filepath.Join(toolDir, "skills", skillName, "SKILL.md")
			if err := os.MkdirAll(filepath.Dir(toolDest), 0755); err != nil {
				return count, fmt.Errorf("create %s: %w", filepath.Dir(toolDest), err)
			}
			if _, statErr := os.Stat(toolDest); statErr != nil || force {
				if err := os.WriteFile(toolDest, enriched, 0644); err != nil {
					return count, fmt.Errorf("write %s: %w", toolDest, err)
				}
				count++
			}
		}
	}

	return count, nil
}

// injectFrontmatter inserts additional YAML fields into an existing SKILL.md
// frontmatter block. For OpenCode targets, model aliases are translated to
// full provider-qualified IDs.
func injectFrontmatter(content []byte, fields map[string]string, target Target) []byte {
	s := string(content)
	if !strings.HasPrefix(s, "---\n") {
		return content
	}

	end := strings.Index(s[4:], "\n---")
	if end < 0 {
		return content
	}

	fmEnd := 4 + end // index of \n before closing ---
	existingFM := s[4:fmEnd]
	rest := s[fmEnd:] // "\n---\n..." rest of file

	var extra strings.Builder
	for key, val := range fields {
		if key == "model" && target == TargetOpenCode {
			if fullID, ok := modelMap[val]; ok {
				val = fullID
			}
		}
		extra.WriteString(key + ": " + val + "\n")
	}

	var buf strings.Builder
	buf.WriteString("---\n")
	buf.WriteString(existingFM)
	buf.WriteString("\n")
	buf.WriteString(extra.String())
	buf.WriteString(rest[1:]) // skip the leading \n (already added above)
	return []byte(buf.String())
}
