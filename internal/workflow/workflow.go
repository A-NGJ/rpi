package workflow

import (
	"bytes"
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

//go:embed all:assets
var assets embed.FS

// ReadAsset reads an embedded file from the assets directory.
// The path should be relative to assets/ (e.g., "templates/CLAUDE.md.template").
func ReadAsset(path string) ([]byte, error) {
	return assets.ReadFile("assets/" + path)
}

// backupAndWrite writes newData to dest. If dest already exists with different
// content, a .bak copy is created first. Returns whether the file was written
// and whether a backup was created. Skips writing if content is identical.
func backupAndWrite(dest string, newData []byte) (written, backedUp bool, err error) {
	existing, readErr := os.ReadFile(dest)
	if readErr == nil {
		if bytes.Equal(existing, newData) {
			return false, false, nil
		}
		if writeErr := os.WriteFile(dest+".bak", existing, 0644); writeErr != nil {
			return false, false, fmt.Errorf("backup %s: %w", dest, writeErr)
		}
		backedUp = true
	}
	if writeErr := os.WriteFile(dest, newData, 0644); writeErr != nil {
		return false, backedUp, fmt.Errorf("write %s: %w", dest, writeErr)
	}
	return true, backedUp, nil
}

// InstallSkills copies embedded skills to skillsDir as Agent Skills-compliant
// SKILL.md files. Any sibling files in a skill source dir (e.g. an upstream
// LICENSE for bundled third-party skills) are copied verbatim so attribution
// travels with each deployed copy.
// Existing files that differ are backed up to .bak before overwriting.
func InstallSkills(skillsDir string) (installed int, backedUp int, err error) {
	entries, readErr := fs.ReadDir(assets, "assets/skills")
	if readErr != nil {
		return 0, 0, fmt.Errorf("read embedded skills: %w", readErr)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		skillName := entry.Name()
		srcDir := "assets/skills/" + skillName

		files, readErr := fs.ReadDir(assets, srcDir)
		if readErr != nil {
			return installed, backedUp, fmt.Errorf("read %s: %w", srcDir, readErr)
		}

		destDir := filepath.Join(skillsDir, skillName)
		if mkErr := os.MkdirAll(destDir, 0755); mkErr != nil {
			return installed, backedUp, fmt.Errorf("create %s: %w", destDir, mkErr)
		}

		for _, f := range files {
			if f.IsDir() {
				continue
			}
			fileName := f.Name()
			srcPath := srcDir + "/" + fileName

			data, readErr := assets.ReadFile(srcPath)
			if readErr != nil {
				return installed, backedUp, fmt.Errorf("read %s: %w", srcPath, readErr)
			}

			dest := filepath.Join(destDir, fileName)
			written, backed, writeErr := backupAndWrite(dest, data)
			if writeErr != nil {
				return installed, backedUp, writeErr
			}
			if written {
				installed++
			}
			if backed {
				backedUp++
			}
		}
	}

	return installed, backedUp, nil
}

// InstallTemplates copies embedded .tmpl files to templatesDir.
// Existing files that differ are backed up to .bak before overwriting.
func InstallTemplates(templatesDir string) (installed int, backedUp int, err error) {
	entries, readErr := fs.ReadDir(assets, "assets/templates")
	if readErr != nil {
		return 0, 0, fmt.Errorf("read embedded templates: %w", readErr)
	}

	if mkErr := os.MkdirAll(templatesDir, 0755); mkErr != nil {
		return 0, 0, fmt.Errorf("create %s: %w", templatesDir, mkErr)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".tmpl") {
			continue
		}

		data, readErr := assets.ReadFile("assets/templates/" + entry.Name())
		if readErr != nil {
			return installed, backedUp, fmt.Errorf("read %s: %w", entry.Name(), readErr)
		}

		dest := filepath.Join(templatesDir, entry.Name())
		written, backed, writeErr := backupAndWrite(dest, data)
		if writeErr != nil {
			return installed, backedUp, writeErr
		}
		if written {
			installed++
		}
		if backed {
			backedUp++
		}
	}

	return installed, backedUp, nil
}

// InstallAgents copies embedded agent definitions to agentsDir.
// Existing files that differ are backed up to .bak before overwriting.
func InstallAgents(agentsDir string) (installed int, backedUp int, err error) {
	entries, readErr := fs.ReadDir(assets, "assets/agents")
	if readErr != nil {
		return 0, 0, fmt.Errorf("read embedded agents: %w", readErr)
	}

	if mkErr := os.MkdirAll(agentsDir, 0755); mkErr != nil {
		return 0, 0, fmt.Errorf("create %s: %w", agentsDir, mkErr)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		data, readErr := assets.ReadFile("assets/agents/" + entry.Name())
		if readErr != nil {
			return installed, backedUp, fmt.Errorf("read %s: %w", entry.Name(), readErr)
		}

		dest := filepath.Join(agentsDir, entry.Name())
		written, backed, writeErr := backupAndWrite(dest, data)
		if writeErr != nil {
			return installed, backedUp, writeErr
		}
		if written {
			installed++
		}
		if backed {
			backedUp++
		}
	}

	return installed, backedUp, nil
}
