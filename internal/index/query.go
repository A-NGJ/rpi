package index

import (
	"os"
	"strings"
	"time"
)

// QueryOptions controls symbol filtering.
type QueryOptions struct {
	Pattern      string
	Kind         string
	ExportedOnly bool
	Signature    string
	Package      string
}

// StatusResult holds index freshness information.
type StatusResult struct {
	Exists         bool           `json:"exists"`
	BuiltAt        time.Time      `json:"built_at,omitempty"`
	AgeSeconds     int64          `json:"age_seconds,omitempty"`
	FileCount      int            `json:"file_count,omitempty"`
	SymbolCount    int            `json:"symbol_count,omitempty"`
	Languages      map[string]int `json:"languages,omitempty"`
	StaleFiles     int            `json:"stale_files,omitempty"`
	IndexPath      string         `json:"index_path,omitempty"`
	IndexSizeBytes int64          `json:"index_size_bytes,omitempty"`
}

// QuerySymbols filters an index's symbols by the given options.
// Pattern matching is case-insensitive substring.
func QuerySymbols(idx *Index, opts QueryOptions) []Symbol {
	var results []Symbol
	pattern := strings.ToLower(opts.Pattern)
	signature := strings.ToLower(opts.Signature)
	pkg := strings.ToLower(opts.Package)
	for _, s := range idx.Symbols {
		if pattern != "" && !strings.Contains(strings.ToLower(s.Name), pattern) {
			continue
		}
		if opts.Kind != "" && s.Kind != opts.Kind {
			continue
		}
		if opts.ExportedOnly && !s.Exported {
			continue
		}
		if signature != "" && !strings.Contains(strings.ToLower(s.Signature), signature) {
			continue
		}
		if pkg != "" && !strings.Contains(strings.ToLower(s.Package), pkg) {
			continue
		}
		results = append(results, s)
	}
	return results
}

// PackageSummary aggregates symbol data for a single package.
type PackageSummary struct {
	Name            string         `json:"name"`
	Files           []string       `json:"files"`
	FileCount       int            `json:"file_count"`
	ExportedSymbols int            `json:"exported_symbols"`
	TotalSymbols    int            `json:"total_symbols"`
	Kinds           map[string]int `json:"kinds"`
}

// QueryPackages aggregates symbols into package-level summaries.
// If pkg is non-empty, only packages matching (case-insensitive substring) are returned.
func QueryPackages(idx *Index, pkg string) []PackageSummary {
	filter := strings.ToLower(pkg)

	type acc struct {
		files    map[string]bool
		exported int
		total    int
		kinds    map[string]int
	}
	packages := make(map[string]*acc)

	for _, s := range idx.Symbols {
		if s.Package == "" {
			continue
		}
		if filter != "" && !strings.Contains(strings.ToLower(s.Package), filter) {
			continue
		}
		a, ok := packages[s.Package]
		if !ok {
			a = &acc{files: make(map[string]bool), kinds: make(map[string]int)}
			packages[s.Package] = a
		}
		a.files[s.File] = true
		a.total++
		if s.Exported {
			a.exported++
		}
		a.kinds[s.Kind]++
	}

	results := make([]PackageSummary, 0, len(packages))
	for name, a := range packages {
		files := make([]string, 0, len(a.files))
		for f := range a.files {
			files = append(files, f)
		}
		results = append(results, PackageSummary{
			Name:            name,
			Files:           files,
			FileCount:       len(files),
			ExportedSymbols: a.exported,
			TotalSymbols:    a.total,
			Kinds:           a.kinds,
		})
	}
	return results
}

// QueryFiles filters an index's files by language. Empty lang returns all files.
func QueryFiles(idx *Index, lang string) []FileEntry {
	if lang == "" {
		return idx.Files
	}
	var results []FileEntry
	for _, f := range idx.Files {
		if f.Language == lang {
			results = append(results, f)
		}
	}
	return results
}

// Status computes freshness information for an index.
func Status(idx *Index, rootPath string) *StatusResult {
	result := &StatusResult{
		Exists:      true,
		BuiltAt:     idx.Metadata.BuiltAt,
		AgeSeconds:  int64(time.Since(idx.Metadata.BuiltAt).Seconds()),
		FileCount:   idx.Metadata.FileCount,
		SymbolCount: idx.Metadata.SymbolCount,
		Languages:   make(map[string]int),
	}

	for _, f := range idx.Files {
		result.Languages[f.Language]++
	}

	// Count stale files by comparing stored mtime against disk.
	for _, f := range idx.Files {
		fullPath := f.Path
		if rootPath != "" {
			fullPath = rootPath + "/" + f.Path
		}
		info, err := os.Stat(fullPath)
		if err != nil {
			result.StaleFiles++
			continue
		}
		if !info.ModTime().Equal(f.Modified) {
			result.StaleFiles++
		}
	}

	return result
}
