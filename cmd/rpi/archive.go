package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/A-NGJ/rpi/internal/frontmatter"
	"github.com/A-NGJ/rpi/internal/scanner"
	"github.com/spf13/cobra"
)

var archiveCmd = &cobra.Command{
	Use:   "archive",
	Short: "Archive operations for .rpi/ artifacts",
	Long: `Manage the archive lifecycle for .rpi/ artifacts.

An artifact is archivable when its status is "complete" or "superseded" and
it is not already in the archive/ subdirectory. Archived files are moved to
.rpi/archive/YYYY-MM/<type>/ with their frontmatter updated (status set
to "archived", archived_date added).`,
	Example: `  # Discover archivable artifacts
  rpi archive scan

  # Check if an artifact has active references
  rpi archive check-refs .rpi/designs/2026-03-13-auth.md

  # Archive an artifact (fails with exit 3 if active refs exist)
  rpi archive move .rpi/plans/2026-03-13-auth.md

  # Force archive even with active references
  rpi archive move --force .rpi/plans/2026-03-13-auth.md`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var archiveScanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Discover archivable artifacts with reference counts",
	Long: `Find artifacts with status "complete" or "superseded" that are not already
in archive/. Returns each artifact's path, type, status, title, and the
number of other artifacts that reference it.`,
	Example: `  rpi archive scan
  # → [{"path": ".rpi/plans/...", "type": "plan", "status": "complete", "ref_count": 0}]`,
	RunE: runArchiveScan,
}

var archiveCheckRefsCmd = &cobra.Command{
	Use:   "check-refs <path>",
	Short: "Find all files referencing a given path",
	Long: `Search frontmatter fields and body text of all .rpi/ files for
references to the given path. Returns a list of files that reference it.`,
	Example: `  rpi archive check-refs .rpi/designs/2026-03-13-auth.md
  # → [".rpi/plans/2026-03-13-auth.md"]`,
	Args: cobra.ExactArgs(1),
	RunE: runArchiveCheckRefs,
}

var archiveMoveCmd = &cobra.Command{
	Use:   "move <path>",
	Short: "Archive an artifact: update frontmatter and move to archive/",
	Long: `Update the artifact's frontmatter (status → "archived", adds archived_date)
and move the file to .rpi/archive/YYYY-MM/<type>/.

Exits with code 3 if the file has active references. Use --force to override.`,
	Example: `  rpi archive move .rpi/plans/2026-03-13-auth.md
  # → {"from": "...", "to": ".rpi/archive/2026-03/plans/...", "frontmatter_updated": true}

  # Override active reference check
  rpi archive move --force .rpi/plans/2026-03-13-auth.md`,
	Args: cobra.ExactArgs(1),
	RunE: runArchiveMove,
}

var archiveMoveForce bool

type archiveMoveResult struct {
	From               string `json:"from"`
	To                 string `json:"to"`
	FrontmatterUpdated bool   `json:"frontmatter_updated"`
}

// errHasReferences is returned when a file has references and --force is not set.
var errHasReferences = fmt.Errorf("file has active references")

type archiveScanResult struct {
	Path     string  `json:"path"`
	Type     string  `json:"type"`
	Status   *string `json:"status"`
	Title    *string `json:"title"`
	RefCount int     `json:"ref_count"`
}

func init() {
	archiveCmd.PersistentFlags().StringVar(&rpiDirFlag, "rpi-dir", ".rpi", "Path to .rpi/ artifacts directory")
	archiveMoveCmd.Flags().BoolVar(&archiveMoveForce, "force", false, "Skip ref check warning")
	archiveCmd.AddCommand(archiveScanCmd)
	archiveCmd.AddCommand(archiveCheckRefsCmd)
	archiveCmd.AddCommand(archiveMoveCmd)
	rootCmd.AddCommand(archiveCmd)
}

func runArchiveScan(cmd *cobra.Command, args []string) error {
	results, err := scanner.Scan(rpiDirFlag, scanner.Filters{Archivable: true})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	var output []archiveScanResult
	for _, r := range results {
		// Use relative path for ref counting since references are stored as relative paths
		refPath := r.Path
		if rel, err := filepath.Rel(rpiDirFlag, r.Path); err == nil {
			refPath = rel
		}
		refCount, err := scanner.CountReferences(rpiDirFlag, refPath)
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
	refs, err := scanner.FindReferences(rpiDirFlag, args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	data, _ := json.MarshalIndent(refs, "", "  ")
	fmt.Fprintln(cmd.OutOrStdout(), string(data))
	return nil
}

func runArchiveMove(cmd *cobra.Command, args []string) error {
	result, err := doArchiveMove(args[0], rpiDirFlag, archiveMoveForce, time.Now())
	if err == errHasReferences {
		fmt.Fprintln(os.Stderr, "error: file has active references (use --force to override)")
		os.Exit(3)
	}
	if err != nil {
		return err
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	fmt.Fprintln(cmd.OutOrStdout(), string(data))
	return nil
}

func doArchiveMove(targetPath, rpiDir string, force bool, now time.Time) (*archiveMoveResult, error) {
	doc, err := frontmatter.Parse(targetPath)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", targetPath, err)
	}

	// Check references
	refPath := targetPath
	if rel, relErr := filepath.Rel(rpiDir, targetPath); relErr == nil {
		refPath = rel
	}
	refCount, err := scanner.CountReferences(rpiDir, refPath)
	if err != nil {
		return nil, fmt.Errorf("count references: %w", err)
	}
	if refCount > 0 && !force {
		return nil, errHasReferences
	}

	// Update frontmatter
	doc.Frontmatter["status"] = "archived"
	doc.Frontmatter["archived_date"] = now.Format("2006-01-02")
	if err := frontmatter.Write(doc); err != nil {
		return nil, fmt.Errorf("write frontmatter: %w", err)
	}

	// Compute destination
	artifactType := scanner.InferType(targetPath)
	yearMonth := now.Format("2006-01")
	filename := filepath.Base(targetPath)
	destDir := filepath.Join(rpiDir, "archive", yearMonth, artifactType)
	destPath := filepath.Join(destDir, filename)

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return nil, fmt.Errorf("create archive dir: %w", err)
	}
	if err := os.Rename(targetPath, destPath); err != nil {
		return nil, fmt.Errorf("move file: %w", err)
	}

	return &archiveMoveResult{
		From:               targetPath,
		To:                 destPath,
		FrontmatterUpdated: true,
	}, nil
}
