package spec

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/A-NGJ/ai-agent-research-plan-implement-flow/internal/frontmatter"
)

// Behavior represents a single behavioral guarantee from a spec.
type Behavior struct {
	ID          string
	Description string
}

// TestRef represents a spec behavior reference found in a test file.
type TestRef struct {
	ID   string
	File string
	Line int
}

// CoverageReport summarizes spec behavior coverage in test files.
type CoverageReport struct {
	SpecDomain       string
	Prefix           string
	Total            int
	Covered          int
	Missing          int
	CoveredBehaviors []CoveredBehavior
	MissingBehaviors []Behavior
}

// CoveredBehavior pairs a behavior with the test reference that covers it.
type CoveredBehavior struct {
	Behavior
	Ref TestRef
}

// behaviorRe matches **XX-N**: patterns in spec files.
var behaviorRe = regexp.MustCompile(`\*\*([A-Z]+-\d+)\*\*:\s*(.*)`)

// specLineRe matches lines containing // spec: comments.
var specLineRe = regexp.MustCompile(`//.*spec:`)

// specRefRe extracts individual spec:XX-N references from a comment line.
var specRefRe = regexp.MustCompile(`spec:([A-Z]+-\d+)`)

// skipDirs are directories to skip when scanning for test files.
var skipDirs = map[string]bool{
	"vendor":       true,
	".git":         true,
	"node_modules": true,
}

// ParseBehaviors reads a spec file and extracts behavior IDs and descriptions.
func ParseBehaviors(specPath string) ([]Behavior, string, error) {
	doc, err := frontmatter.Parse(specPath)
	if err != nil {
		return nil, "", fmt.Errorf("parse spec: %w", err)
	}

	id, ok := doc.Frontmatter["id"].(string)
	if !ok || id == "" {
		return nil, "", fmt.Errorf("spec %s missing required 'id' field in frontmatter", specPath)
	}

	var behaviors []Behavior
	for _, match := range behaviorRe.FindAllStringSubmatch(doc.Body, -1) {
		behaviors = append(behaviors, Behavior{
			ID:          match[1],
			Description: strings.TrimSpace(match[2]),
		})
	}

	return behaviors, id, nil
}

// ScanTestFiles walks projectRoot looking for *_test.go files containing
// // spec:XX-N comments matching the given prefix.
func ScanTestFiles(projectRoot string, prefix string) ([]TestRef, error) {
	var refs []TestRef

	err := filepath.Walk(projectRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() && skipDirs[info.Name()] {
			return filepath.SkipDir
		}

		if info.IsDir() || !strings.HasSuffix(path, "_test.go") {
			return nil
		}

		fileRefs, err := scanFile(path, prefix)
		if err != nil {
			return nil // skip unreadable files
		}
		refs = append(refs, fileRefs...)

		return nil
	})

	return refs, err
}

func scanFile(path string, prefix string) ([]TestRef, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var refs []TestRef
	scanner := bufio.NewScanner(f)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		if !specLineRe.MatchString(line) {
			continue
		}
		for _, match := range specRefRe.FindAllStringSubmatch(line, -1) {
			if strings.HasPrefix(match[1], prefix+"-") {
				refs = append(refs, TestRef{
					ID:   match[1],
					File: path,
					Line: lineNum,
				})
			}
		}
	}

	return refs, scanner.Err()
}

// ComputeCoverage matches behaviors against test references and produces a report.
func ComputeCoverage(behaviors []Behavior, refs []TestRef, domain string, prefix string) *CoverageReport {
	refMap := make(map[string]TestRef)
	for _, r := range refs {
		if _, exists := refMap[r.ID]; !exists {
			refMap[r.ID] = r
		}
	}

	report := &CoverageReport{
		SpecDomain: domain,
		Prefix:     prefix,
		Total:      len(behaviors),
	}

	for _, b := range behaviors {
		if ref, ok := refMap[b.ID]; ok {
			report.Covered++
			report.CoveredBehaviors = append(report.CoveredBehaviors, CoveredBehavior{
				Behavior: b,
				Ref:      ref,
			})
		} else {
			report.Missing++
			report.MissingBehaviors = append(report.MissingBehaviors, b)
		}
	}

	return report
}
