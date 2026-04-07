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

// readOnlyBuiltinTools lists Claude Code built-in tools available to read-only skills.
const readOnlyBuiltinTools = "Read,Glob,Grep,Bash,Agent,LSP"

// mcpTools lists all RPI MCP tools with the mcp__rpi__ prefix.
const mcpTools = "mcp__rpi__rpi_git_context,mcp__rpi__rpi_git_changed_files," +
	"mcp__rpi__rpi_git_sensitive_check,mcp__rpi__rpi_archive_scan," +
	"mcp__rpi__rpi_scan,mcp__rpi__rpi_scaffold," +
	"mcp__rpi__rpi_frontmatter_get,mcp__rpi__rpi_frontmatter_set," +
	"mcp__rpi__rpi_frontmatter_transition,mcp__rpi__rpi_chain," +
	"mcp__rpi__rpi_extract,mcp__rpi__rpi_extract_list_sections," +
	"mcp__rpi__rpi_verify_completeness,mcp__rpi__rpi_verify_markers," +
	"mcp__rpi__rpi_verify_spec,mcp__rpi__rpi_context_essentials," +
	"mcp__rpi__rpi_session_resume,mcp__rpi__rpi_suggest_next," +
	"mcp__rpi__rpi_archive_check_refs,mcp__rpi__rpi_archive_move"

// readOnlyTools is the full allowed-tools value for read-only skills.
const readOnlyTools = readOnlyBuiltinTools + "," + mcpTools

// researchTools adds web access to the read-only tool set.
const researchTools = readOnlyTools + ",WebSearch,WebFetch"

// skillOverrides defines per-skill tool-specific frontmatter fields to inject
// when the target is not agents-only. Skills not listed here get no extra
// fields.
var skillOverrides = map[string]map[string]string{
	"rpi-archive":  {"model": "haiku", "disable-model-invocation": "true"},
	"rpi-commit":   {"model": "haiku"},
	"rpi-research": {"allowed-tools": researchTools, "context": "fork"},
	"rpi-verify":   {"allowed-tools": readOnlyTools},
	"rpi-explain":  {"allowed-tools": readOnlyTools},
}

//go:embed all:assets
var assets embed.FS

// ReadAsset reads an embedded file from the assets directory.
// The path should be relative to assets/ (e.g., "templates/CLAUDE.md.template").
func ReadAsset(path string) ([]byte, error) {
	return assets.ReadFile("assets/" + path)
}

// InstallSkills copies embedded skills to skillsDir as Agent Skills-compliant
// SKILL.md files. For non-agents-only targets, tool-specific frontmatter
// (model, etc.) is injected into the installed copies.
// Existing files are only overwritten when force is true.
func InstallSkills(skillsDir string, target Target, force bool) (int, error) {
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

		// Enrich with tool-specific frontmatter if applicable
		if target != TargetAgentsOnly {
			if overrides, ok := skillOverrides[skillName]; ok {
				data = injectFrontmatter(data, overrides, target)
			}
		}

		dest := filepath.Join(skillsDir, skillName, "SKILL.md")
		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			return count, fmt.Errorf("create %s: %w", filepath.Dir(dest), err)
		}
		if _, statErr := os.Stat(dest); statErr != nil || force {
			if err := os.WriteFile(dest, data, 0644); err != nil {
				return count, fmt.Errorf("write %s: %w", dest, err)
			}
			count++
		}
	}

	return count, nil
}

// InstallTemplates copies embedded .tmpl files to templatesDir.
// Existing files are only overwritten when force is true.
func InstallTemplates(templatesDir string, force bool) (int, error) {
	entries, err := fs.ReadDir(assets, "assets/templates")
	if err != nil {
		return 0, fmt.Errorf("read embedded templates: %w", err)
	}

	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		return 0, fmt.Errorf("create %s: %w", templatesDir, err)
	}

	count := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".tmpl") {
			continue
		}

		data, err := assets.ReadFile("assets/templates/" + entry.Name())
		if err != nil {
			return count, fmt.Errorf("read %s: %w", entry.Name(), err)
		}

		dest := filepath.Join(templatesDir, entry.Name())
		if _, statErr := os.Stat(dest); statErr != nil || force {
			if err := os.WriteFile(dest, data, 0644); err != nil {
				return count, fmt.Errorf("write %s: %w", dest, err)
			}
			count++
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
