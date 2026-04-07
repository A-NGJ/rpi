package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/A-NGJ/rpi/internal/templates"
	"github.com/A-NGJ/rpi/internal/workflow"
)

var rpiSubdirs = []string{
	"research", "designs", "diagnoses",
	"plans", "specs", "reviews", "archive",
}

type syncOptions struct {
	targetDir string
	cfg       targetConfig
	force     bool
	skipRules bool
	w         io.Writer
}

func syncProject(opts syncOptions) error {
	rpiDir := filepath.Join(opts.targetDir, ".rpi")

	// Ensure .rpi/ subdirs exist
	for _, d := range rpiSubdirs {
		path := filepath.Join(rpiDir, d)
		if _, statErr := os.Stat(path); statErr != nil {
			if mkErr := os.MkdirAll(path, 0755); mkErr != nil {
				return fmt.Errorf("create %s: %w", path, mkErr)
			}
			logSuccess(opts.w, fmt.Sprintf("Created .rpi/%s/", d))
		}
	}

	// Ensure tool subdirs exist (skip for agents-only)
	if opts.cfg.toolDir != "" {
		toolDirPath := filepath.Join(opts.targetDir, opts.cfg.toolDir)
		for _, d := range opts.cfg.subdirs {
			path := filepath.Join(toolDirPath, d)
			if _, statErr := os.Stat(path); statErr != nil {
				if mkErr := os.MkdirAll(path, 0755); mkErr != nil {
					return fmt.Errorf("create %s: %w", path, mkErr)
				}
				logSuccess(opts.w, fmt.Sprintf("Created %s/%s/", opts.cfg.toolDir, d))
			}
		}
	}

	// Install skills
	var skillsDir string
	if opts.cfg.toolDir != "" {
		skillsDir = filepath.Join(opts.targetDir, opts.cfg.toolDir, "skills")
	} else {
		skillsDir = filepath.Join(opts.targetDir, ".agents", "skills")
	}
	skillCount, err := workflow.InstallSkills(skillsDir, opts.cfg.target, opts.force)
	if err != nil {
		return fmt.Errorf("install skills: %w", err)
	}
	if skillCount > 0 {
		logSuccess(opts.w, fmt.Sprintf("Installed %d skill files", skillCount))
	}

	// Install scaffold templates
	templatesDir := filepath.Join(rpiDir, "templates")
	tplCount, err := workflow.InstallTemplates(templatesDir, opts.force)
	if err != nil {
		return fmt.Errorf("install templates: %w", err)
	}
	if tplCount > 0 {
		logSuccess(opts.w, fmt.Sprintf("Installed %d template files", tplCount))
	}

	// Write rules file (respects force: only overwrite existing if force is true)
	if !opts.skipRules && opts.cfg.rulesFile != "" {
		rulesPath := filepath.Join(opts.targetDir, opts.cfg.rulesFile)
		_, statErr := os.Stat(rulesPath)
		if statErr != nil || opts.force {
			content, tplErr := templates.Get(opts.cfg.rulesFile)
			if tplErr != nil {
				logWarning(opts.w, fmt.Sprintf("get %s template: %v", opts.cfg.rulesFile, tplErr))
			} else {
				if writeErr := os.WriteFile(rulesPath, []byte(content), 0644); writeErr != nil {
					logWarning(opts.w, fmt.Sprintf("write %s: %v", opts.cfg.rulesFile, writeErr))
				} else {
					logSuccess(opts.w, fmt.Sprintf("Installed %s", opts.cfg.rulesFile))
				}
			}
		}
	}

	// Configure settings.json (Claude only)
	if opts.cfg.target == workflow.TargetClaude {
		toolDirPath := filepath.Join(opts.targetDir, opts.cfg.toolDir)
		configureSettings(opts.w, toolDirPath)
		configureHooks(opts.w, toolDirPath, opts.force)
	}

	return nil
}
