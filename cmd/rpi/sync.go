package main

import (
	"bytes"
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
	skillCount, skillBackups, err := workflow.InstallSkills(skillsDir, opts.cfg.target)
	if err != nil {
		return fmt.Errorf("install skills: %w", err)
	}
	if skillCount > 0 {
		logSuccess(opts.w, fmt.Sprintf("Installed %d skill files", skillCount))
	}
	if skillBackups > 0 {
		logSuccess(opts.w, fmt.Sprintf("Backed up %d modified skill files", skillBackups))
	}

	// Install agent definitions (Claude target only)
	if opts.cfg.target == workflow.TargetClaude {
		agentsDir := filepath.Join(opts.targetDir, opts.cfg.toolDir, "agents")
		agentCount, agentBackups, err := workflow.InstallAgents(agentsDir)
		if err != nil {
			return fmt.Errorf("install agents: %w", err)
		}
		if agentCount > 0 {
			logSuccess(opts.w, fmt.Sprintf("Installed %d agent files", agentCount))
		}
		if agentBackups > 0 {
			logSuccess(opts.w, fmt.Sprintf("Backed up %d modified agent files", agentBackups))
		}
	}

	// Install scaffold templates
	templatesDir := filepath.Join(rpiDir, "templates")
	tplCount, tplBackups, err := workflow.InstallTemplates(templatesDir)
	if err != nil {
		return fmt.Errorf("install templates: %w", err)
	}
	if tplCount > 0 {
		logSuccess(opts.w, fmt.Sprintf("Installed %d template files", tplCount))
	}
	if tplBackups > 0 {
		logSuccess(opts.w, fmt.Sprintf("Backed up %d modified template files", tplBackups))
	}

	// Write rules file — always install latest, back up if existing content differs
	if !opts.skipRules && opts.cfg.rulesFile != "" {
		rulesPath := filepath.Join(opts.targetDir, opts.cfg.rulesFile)
		content, tplErr := templates.Get(opts.cfg.rulesFile)
		if tplErr != nil {
			logWarning(opts.w, fmt.Sprintf("get %s template: %v", opts.cfg.rulesFile, tplErr))
		} else {
			newData := []byte(content)
			existing, readErr := os.ReadFile(rulesPath)
			if readErr == nil && !bytes.Equal(existing, newData) {
				if bakErr := os.WriteFile(rulesPath+".bak", existing, 0644); bakErr != nil {
					logWarning(opts.w, fmt.Sprintf("backup %s: %v", opts.cfg.rulesFile, bakErr))
				} else {
					logSuccess(opts.w, fmt.Sprintf("Backed up %s to %s.bak", opts.cfg.rulesFile, opts.cfg.rulesFile))
				}
			}
			if readErr != nil || !bytes.Equal(existing, newData) {
				if writeErr := os.WriteFile(rulesPath, newData, 0644); writeErr != nil {
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
		configureHooks(opts.w, toolDirPath)
	}

	return nil
}
