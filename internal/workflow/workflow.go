package workflow

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
)

//go:embed all:assets
var assets embed.FS

// ReadAsset reads an embedded file from the assets directory.
// The path should be relative to assets/ (e.g., "templates/CLAUDE.md.template").
func ReadAsset(path string) ([]byte, error) {
	return assets.ReadFile("assets/" + path)
}

// Install copies all embedded workflow files (agents, commands, skills)
// into the target .claude/ directory. Existing files are only overwritten
// when force is true.
func Install(claudeDir string, force bool) (int, error) {
	count := 0
	err := fs.WalkDir(assets, "assets", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		// Strip "assets/" prefix to get the relative path inside .claude/
		rel, _ := filepath.Rel("assets", path)
		target := filepath.Join(claudeDir, rel)

		if d.IsDir() {
			return os.MkdirAll(target, 0755)
		}

		// Skip if file exists and not forcing
		if _, err := os.Stat(target); err == nil && !force {
			return nil
		}

		data, err := assets.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.WriteFile(target, data, 0644); err != nil {
			return err
		}
		count++
		return nil
	})
	return count, err
}
