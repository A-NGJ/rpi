// Package coverage provides the deterministic pre-lock analysis of a drafted
// plan: coverage mapping, file existence, and phase-ordering signals computed
// purely from the plan's structure (no git diff, no LLM, no network).
//
// It also owns the single checkbox / **File**: extractor that cmd/rpi's
// completeness check delegates to, so there is exactly one plan parser in the
// codebase.
package coverage

import (
	"regexp"
	"strings"
)

// Line-level extraction regexes — the single source of truth for plan parsing.
// cmd/rpi/verify.go delegates to ParseCheckboxes / ExtractPlanFiles rather than
// re-implementing these.
var (
	uncheckedRe = regexp.MustCompile(`^(\s*)- \[ \] (.+)`)
	checkedRe   = regexp.MustCompile(`^(\s*)- \[x\] (.+)`)
	phaseRe     = regexp.MustCompile(`^## Phase \d+`)
	sectionRe   = regexp.MustCompile(`^### (.+)`)
	// planFileRe captures the **File**: path in group 1 and the trailing
	// remainder (e.g. "(NEW)" / "(UPDATED)") in group 2.
	planFileRe = regexp.MustCompile(`\*\*File\*\*:\s*` + "`" + `([^` + "`" + `]+)` + "`" + `(.*)`)
)

// Checkbox is a single - [ ] / - [x] item with its enclosing heading context.
type Checkbox struct {
	Text    string
	Checked bool
	Phase   string
	Section string
}

// ParseCheckboxes extracts every checkbox item with its phase/section context,
// in document order.
func ParseCheckboxes(content string) []Checkbox {
	var items []Checkbox
	var currentPhase, currentSection string

	for _, line := range strings.Split(content, "\n") {
		if phaseRe.MatchString(line) {
			currentPhase = strings.TrimSpace(strings.TrimPrefix(line, "## "))
			currentSection = ""
		} else if m := sectionRe.FindStringSubmatch(line); m != nil {
			currentSection = strings.TrimSpace(m[1])
		}

		if m := uncheckedRe.FindStringSubmatch(line); m != nil {
			items = append(items, Checkbox{Text: m[2], Checked: false, Phase: currentPhase, Section: currentSection})
		} else if m := checkedRe.FindStringSubmatch(line); m != nil {
			items = append(items, Checkbox{Text: m[2], Checked: true, Phase: currentPhase, Section: currentSection})
		}
	}

	return items
}

// ExtractPlanFiles returns the unique **File**: paths in document order.
func ExtractPlanFiles(content string) []string {
	var files []string
	seen := map[string]bool{}

	for _, line := range strings.Split(content, "\n") {
		if m := planFileRe.FindStringSubmatch(line); m != nil {
			path := m[1]
			if !seen[path] {
				files = append(files, path)
				seen[path] = true
			}
		}
	}

	if files == nil {
		return []string{}
	}
	return files
}
