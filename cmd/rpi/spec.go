package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/A-NGJ/ai-agent-research-plan-implement-flow/internal/frontmatter"
	"github.com/A-NGJ/ai-agent-research-plan-implement-flow/internal/spec"
	"github.com/spf13/cobra"
)

var specCmd = &cobra.Command{
	Use:   "spec",
	Short: "Spec management commands",
}

var specCoverageCmd = &cobra.Command{
	Use:   "coverage <spec-file>",
	Short: "Check test coverage for spec behaviors",
	Long:  `Cross-references spec behavior IDs (XX-N) against // spec:XX-N comments in test files.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runSpecCoverage,
}

func init() {
	specCoverageCmd.Flags().StringVar(&formatFlag, "format", "text", "Output format: text, json")
	specCmd.AddCommand(specCoverageCmd)
	rootCmd.AddCommand(specCmd)
}

func runSpecCoverage(cmd *cobra.Command, args []string) error {
	specPath := args[0]

	// Parse behaviors from spec
	behaviors, prefix, err := spec.ParseBehaviors(specPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}

	// Get spec domain from frontmatter
	doc, err := frontmatter.Parse(specPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(2)
	}
	domain, _ := doc.Frontmatter["domain"].(string)

	// Find project root (walk up to go.mod)
	projectRoot, err := findProjectRoot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// Scan test files
	refs, err := spec.ScanTestFiles(projectRoot, prefix)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// Compute coverage
	report := spec.ComputeCoverage(behaviors, refs, domain, prefix)

	// Output
	switch formatFlag {
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(report); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	default:
		printTextReport(report)
	}

	// Exit code 1 if any missing
	if report.Missing > 0 {
		os.Exit(1)
	}

	return nil
}

func printTextReport(r *spec.CoverageReport) {
	fmt.Printf("Spec: %s (%s)\n", r.SpecDomain, r.Prefix)
	fmt.Printf("Behaviors: %d total, %d covered, %d missing\n", r.Total, r.Covered, r.Missing)

	if len(r.CoveredBehaviors) > 0 {
		fmt.Println("\nCovered:")
		for _, cb := range r.CoveredBehaviors {
			fmt.Printf("  %-6s %s:%d\n", cb.ID, cb.Ref.File, cb.Ref.Line)
		}
	}

	if len(r.MissingBehaviors) > 0 {
		fmt.Println("\nMissing:")
		for _, mb := range r.MissingBehaviors {
			fmt.Printf("  %-6s %s\n", mb.ID, mb.Description)
		}
	}
}

func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not find project root (no go.mod found)")
		}
		dir = parent
	}
}
