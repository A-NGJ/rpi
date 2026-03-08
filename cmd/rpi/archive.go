package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/A-NGJ/ai-agent-research-plan-implement-flow/internal/scanner"
	"github.com/spf13/cobra"
)

var archiveCmd = &cobra.Command{
	Use:   "archive",
	Short: "Archive operations for .thoughts/ artifacts",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var archiveScanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Discover archivable artifacts with reference counts",
	RunE:  runArchiveScan,
}

var archiveCheckRefsCmd = &cobra.Command{
	Use:   "check-refs <path>",
	Short: "Find all files referencing a given path",
	Args:  cobra.ExactArgs(1),
	RunE:  runArchiveCheckRefs,
}

type archiveScanResult struct {
	Path     string  `json:"path"`
	Type     string  `json:"type"`
	Status   *string `json:"status"`
	Title    *string `json:"title"`
	RefCount int     `json:"ref_count"`
}

func init() {
	archiveCmd.AddCommand(archiveScanCmd)
	archiveCmd.AddCommand(archiveCheckRefsCmd)
	rootCmd.AddCommand(archiveCmd)
}

func runArchiveScan(cmd *cobra.Command, args []string) error {
	results, err := scanner.Scan(thoughtsDirFlag, scanner.Filters{Archivable: true})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	var output []archiveScanResult
	for _, r := range results {
		// Use relative path for ref counting since references are stored as relative paths
		refPath := r.Path
		if rel, err := filepath.Rel(thoughtsDirFlag, r.Path); err == nil {
			refPath = rel
		}
		refCount, err := scanner.CountReferences(thoughtsDirFlag, refPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "warning: counting refs for %s: %v\n", r.Path, err)
			refCount = 0
		}
		output = append(output, archiveScanResult{
			Path:     r.Path,
			Type:     r.Type,
			Status:   r.Status,
			Title:    r.Title,
			RefCount: refCount,
		})
	}

	if output == nil {
		output = []archiveScanResult{}
	}

	data, _ := json.MarshalIndent(output, "", "  ")
	fmt.Fprintln(cmd.OutOrStdout(), string(data))
	return nil
}

func runArchiveCheckRefs(cmd *cobra.Command, args []string) error {
	refs, err := scanner.FindReferences(thoughtsDirFlag, args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	data, _ := json.MarshalIndent(refs, "", "  ")
	fmt.Fprintln(cmd.OutOrStdout(), string(data))
	return nil
}
