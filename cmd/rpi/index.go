package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/A-NGJ/rpi/internal/index"
	"github.com/spf13/cobra"
)

var (
	indexLangFlag      string
	indexPathFlag      string
	indexForceFlag     bool
	indexKindFlag      string
	indexExportedFlag  bool
	indexFormatFlag    string
	indexSignatureFlag string
	indexPackageFlag   string
)

var indexCmd = &cobra.Command{
	Use:   "index",
	Short: "Build and query a codebase symbol index",
	Long: `Build and query a regex-based symbol index stored at .rpi/index.json.

Supports Go, Python, JavaScript/TypeScript, and Rust. Extracts function,
class, struct, interface, and type alias definitions.`,
	Example: `  # Build the index
  rpi index build

  # Search for symbols
  rpi index query "HandleRequest"

  # Check index freshness
  rpi index status`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var indexBuildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build a symbol index of the codebase",
	Long: `Walk the codebase, extract function/class/struct/interface definitions using
regex patterns, and save the index to .rpi/index.json.

Use --lang to restrict indexing to specific languages (comma-separated).`,
	Example: `  # Index the entire codebase
  rpi index build

  # Index only Go and Python files
  rpi index build --lang go,py

  # Force a full rebuild
  rpi index build --force`,
	RunE: runIndexBuild,
}

var indexQueryCmd = &cobra.Command{
	Use:   "query <pattern>",
	Short: "Search for symbols in the index",
	Long: `Case-insensitive substring match on symbol names. Use --kind to filter by
symbol kind (function, method, class, struct, interface, type_alias) and
--exported to show only exported symbols.`,
	Example: `  # Find all symbols matching "resolve"
  rpi index query resolve

  # Find exported functions only
  rpi index query handler --kind function --exported

  # Output as markdown table
  rpi index query scan --format md`,
	Args: cobra.ExactArgs(1),
	RunE: runIndexQuery,
}

var indexFilesCmd = &cobra.Command{
	Use:   "files",
	Short: "List all indexed files",
	Long: `List indexed files with their language, symbol count, and file size.
Use --lang to filter by language.`,
	Example: `  # List all indexed files
  rpi index files

  # List only Go files
  rpi index files --lang go

  # Output as markdown table
  rpi index files --format md`,
	RunE: runIndexFiles,
}

var indexStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show index metadata and freshness",
	Long: `Show index metadata: age, file count, stale files, languages, symbol count,
and index size. Output is plain text by default; use --format json for JSON.`,
	Example: `  rpi index status
  # Index: .rpi/index.json
  # Built: 2026-03-13T10:00:00Z (120s ago)
  # Files: 42 (3 stale)
  # Symbols: 384
  # Languages: go: 30, py: 12`,
	RunE: runIndexStatus,
}

var indexPackagesCmd = &cobra.Command{
	Use:   "packages",
	Short: "List package summaries from the index",
	Long: `Aggregate symbols into package-level summaries showing file count, symbol
counts by kind, and exported symbol count. Use --package to filter.`,
	Example: `  # List all packages
  rpi index packages

  # Filter by package name
  rpi index packages --package index

  # Output as markdown table
  rpi index packages --format md`,
	RunE: runIndexPackages,
}

var indexImportsCmd = &cobra.Command{
	Use:   "imports <file>",
	Short: "List imports for a file",
	Long: `Show all import statements for files matching the given path (case-insensitive
substring match). Returns import paths, aliases, and line numbers.`,
	Example: `  # Show imports for a specific file
  rpi index imports serve.go

  # Output as markdown table
  rpi index imports serve.go --format md`,
	Args: cobra.ExactArgs(1),
	RunE: runIndexImports,
}

var indexImportersCmd = &cobra.Command{
	Use:   "importers <import_path>",
	Short: "Find files that import a given path",
	Long: `Find all files that import a given path (case-insensitive substring match).
Use this to assess the blast radius of changes to a package or module.`,
	Example: `  # Find files importing "internal/index"
  rpi index importers internal/index

  # Output as markdown table
  rpi index importers react --format md`,
	Args: cobra.ExactArgs(1),
	RunE: runIndexImporters,
}

func init() {
	indexBuildCmd.Flags().StringVar(&indexLangFlag, "lang", "", "Comma-separated languages to index (e.g., go,py,ts)")
	indexBuildCmd.Flags().StringVar(&indexPathFlag, "path", ".", "Root path to index")
	indexBuildCmd.Flags().BoolVar(&indexForceFlag, "force", false, "Force full rebuild")

	indexQueryCmd.Flags().StringVar(&indexKindFlag, "kind", "", "Filter by symbol kind (function, method, class, struct, interface, type_alias)")
	indexQueryCmd.Flags().BoolVar(&indexExportedFlag, "exported", false, "Show only exported symbols")
	indexQueryCmd.Flags().StringVar(&indexSignatureFlag, "signature", "", "Filter by substring in symbol signature")
	indexQueryCmd.Flags().StringVar(&indexPackageFlag, "package", "", "Filter by package name")
	indexQueryCmd.Flags().StringVar(&indexFormatFlag, "format", "json", "Output format: json, md")

	indexFilesCmd.Flags().StringVar(&indexLangFlag, "lang", "", "Filter by language")
	indexFilesCmd.Flags().StringVar(&indexFormatFlag, "format", "json", "Output format: json, md")

	indexStatusCmd.Flags().StringVar(&indexFormatFlag, "format", "text", "Output format: json, text")

	indexPackagesCmd.Flags().StringVar(&indexPackageFlag, "package", "", "Filter by package name")
	indexPackagesCmd.Flags().StringVar(&indexFormatFlag, "format", "json", "Output format: json, md")

	indexImportsCmd.Flags().StringVar(&indexFormatFlag, "format", "json", "Output format: json, md")

	indexImportersCmd.Flags().StringVar(&indexFormatFlag, "format", "json", "Output format: json, md")

	indexCmd.AddCommand(indexBuildCmd)
	indexCmd.AddCommand(indexQueryCmd)
	indexCmd.AddCommand(indexFilesCmd)
	indexCmd.AddCommand(indexStatusCmd)
	indexCmd.AddCommand(indexPackagesCmd)
	indexCmd.AddCommand(indexImportsCmd)
	indexCmd.AddCommand(indexImportersCmd)
	rootCmd.AddCommand(indexCmd)
}

func runIndexBuild(cmd *cobra.Command, args []string) error {
	start := time.Now()

	opts := index.BuildOptions{
		ForceRebuild: indexForceFlag,
	}
	if indexLangFlag != "" {
		opts.Languages = strings.Split(indexLangFlag, ",")
	}

	idx, err := index.Build(indexPathFlag, opts)
	if err != nil {
		return fmt.Errorf("build index: %w", err)
	}

	absPath, _ := filepath.Abs(indexPathFlag)
	indexPath := filepath.Join(absPath, index.DefaultIndexPath)
	if err := index.Save(idx, indexPath); err != nil {
		return fmt.Errorf("save index: %w", err)
	}

	if !index.IsGitignored(absPath) {
		fmt.Fprintln(os.Stderr, "warning: .rpi/ is not in .gitignore — index may be accidentally committed")
	}

	elapsed := time.Since(start).Seconds()
	fmt.Fprintf(cmd.OutOrStdout(), "Indexed %d files (%d symbols) in %.1fs. Written to %s\n",
		idx.Metadata.FileCount, idx.Metadata.SymbolCount, elapsed, index.DefaultIndexPath)
	return nil
}

func runIndexQuery(cmd *cobra.Command, args []string) error {
	idx, err := loadIndex()
	if err != nil {
		return err
	}

	results := index.QuerySymbols(idx, index.QueryOptions{
		Pattern:      args[0],
		Kind:         indexKindFlag,
		ExportedOnly: indexExportedFlag,
		Signature:    indexSignatureFlag,
		Package:      indexPackageFlag,
	})

	if results == nil {
		results = []index.Symbol{}
	}

	switch indexFormatFlag {
	case "md":
		printSymbolsMarkdown(cmd, results)
	default:
		data, _ := json.MarshalIndent(results, "", "  ")
		fmt.Fprintln(cmd.OutOrStdout(), string(data))
	}
	return nil
}

func runIndexFiles(cmd *cobra.Command, args []string) error {
	idx, err := loadIndex()
	if err != nil {
		return err
	}

	results := index.QueryFiles(idx, indexLangFlag)
	if results == nil {
		results = []index.FileEntry{}
	}

	type fileOutput struct {
		Path     string `json:"path"`
		Language string `json:"language"`
		Symbols  int    `json:"symbols"`
		Size     int64  `json:"size"`
	}

	// Count symbols per file.
	symCounts := make(map[string]int)
	for _, s := range idx.Symbols {
		symCounts[s.File]++
	}

	var out []fileOutput
	for _, f := range results {
		out = append(out, fileOutput{
			Path:     f.Path,
			Language: f.Language,
			Symbols:  symCounts[f.Path],
			Size:     f.Size,
		})
	}
	if out == nil {
		out = []fileOutput{}
	}

	switch indexFormatFlag {
	case "md":
		fmt.Fprintln(cmd.OutOrStdout(), "| Path | Language | Symbols | Size |")
		fmt.Fprintln(cmd.OutOrStdout(), "|---|---|---|---|")
		for _, f := range out {
			fmt.Fprintf(cmd.OutOrStdout(), "| %s | %s | %d | %d |\n", f.Path, f.Language, f.Symbols, f.Size)
		}
	default:
		data, _ := json.MarshalIndent(out, "", "  ")
		fmt.Fprintln(cmd.OutOrStdout(), string(data))
	}
	return nil
}

func runIndexStatus(cmd *cobra.Command, args []string) error {
	indexPath := index.DefaultIndexPath

	idx, err := index.Load(indexPath)
	if err != nil {
		// Index doesn't exist or is unreadable — not an error for status.
		switch indexFormatFlag {
		case "json":
			fmt.Fprintln(cmd.OutOrStdout(), `{"exists": false}`)
		default:
			fmt.Fprintln(cmd.OutOrStdout(), "No index found. Run 'rpi index build' to create one.")
		}
		return nil
	}

	result := index.Status(idx, idx.Metadata.RootPath)
	result.IndexPath = indexPath
	if info, err := os.Stat(indexPath); err == nil {
		result.IndexSizeBytes = info.Size()
	}

	switch indexFormatFlag {
	case "json":
		data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Fprintln(cmd.OutOrStdout(), string(data))
	default:
		fmt.Fprintf(cmd.OutOrStdout(), "Index: %s\n", result.IndexPath)
		fmt.Fprintf(cmd.OutOrStdout(), "Built: %s (%.0fs ago)\n", result.BuiltAt.Format(time.RFC3339), float64(result.AgeSeconds))
		fmt.Fprintf(cmd.OutOrStdout(), "Files: %d (%d stale)\n", result.FileCount, result.StaleFiles)
		fmt.Fprintf(cmd.OutOrStdout(), "Symbols: %d\n", result.SymbolCount)
		langs := make([]string, 0, len(result.Languages))
		for l, c := range result.Languages {
			langs = append(langs, fmt.Sprintf("%s: %d", l, c))
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Languages: %s\n", strings.Join(langs, ", "))
	}
	return nil
}

func runIndexPackages(cmd *cobra.Command, args []string) error {
	idx, err := loadIndex()
	if err != nil {
		return err
	}

	results := index.QueryPackages(idx, indexPackageFlag)
	if results == nil {
		results = []index.PackageSummary{}
	}

	switch indexFormatFlag {
	case "md":
		fmt.Fprintln(cmd.OutOrStdout(), "| Package | Files | Exported | Total | Kinds |")
		fmt.Fprintln(cmd.OutOrStdout(), "|---|---|---|---|---|")
		for _, p := range results {
			kinds := make([]string, 0, len(p.Kinds))
			for k, v := range p.Kinds {
				kinds = append(kinds, fmt.Sprintf("%s: %d", k, v))
			}
			fmt.Fprintf(cmd.OutOrStdout(), "| %s | %d | %d | %d | %s |\n",
				p.Name, p.FileCount, p.ExportedSymbols, p.TotalSymbols, strings.Join(kinds, ", "))
		}
	default:
		data, _ := json.MarshalIndent(results, "", "  ")
		fmt.Fprintln(cmd.OutOrStdout(), string(data))
	}
	return nil
}

func loadIndex() (*index.Index, error) {
	idx, err := index.Load(index.DefaultIndexPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: no index found — run 'rpi index build' first\n")
		os.Exit(1)
	}
	return idx, nil
}

func runIndexImports(cmd *cobra.Command, args []string) error {
	idx, err := loadIndex()
	if err != nil {
		return err
	}

	results := index.QueryImports(idx, args[0])
	if results == nil {
		results = []index.Import{}
	}

	switch indexFormatFlag {
	case "md":
		fmt.Fprintln(cmd.OutOrStdout(), "| File | Import Path | Alias | Line |")
		fmt.Fprintln(cmd.OutOrStdout(), "|---|---|---|---|")
		for _, imp := range results {
			fmt.Fprintf(cmd.OutOrStdout(), "| %s | %s | %s | %d |\n", imp.File, imp.ImportPath, imp.Alias, imp.Line)
		}
	default:
		data, _ := json.MarshalIndent(results, "", "  ")
		fmt.Fprintln(cmd.OutOrStdout(), string(data))
	}
	return nil
}

func runIndexImporters(cmd *cobra.Command, args []string) error {
	idx, err := loadIndex()
	if err != nil {
		return err
	}

	results := index.QueryImporters(idx, args[0])
	if results == nil {
		results = []string{}
	}

	switch indexFormatFlag {
	case "md":
		fmt.Fprintln(cmd.OutOrStdout(), "| File |")
		fmt.Fprintln(cmd.OutOrStdout(), "|---|")
		for _, f := range results {
			fmt.Fprintf(cmd.OutOrStdout(), "| %s |\n", f)
		}
	default:
		data, _ := json.MarshalIndent(results, "", "  ")
		fmt.Fprintln(cmd.OutOrStdout(), string(data))
	}
	return nil
}

func printSymbolsMarkdown(cmd *cobra.Command, syms []index.Symbol) {
	fmt.Fprintln(cmd.OutOrStdout(), "| Name | Kind | File | Line | Package | Signature |")
	fmt.Fprintln(cmd.OutOrStdout(), "|---|---|---|---|---|---|")
	for _, s := range syms {
		fmt.Fprintf(cmd.OutOrStdout(), "| %s | %s | %s | %d | %s | %s |\n",
			s.Name, s.Kind, s.File, s.Line, s.Package, s.Signature)
	}
}
