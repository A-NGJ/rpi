package index

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func sampleIndex() *Index {
	return &Index{
		Metadata: Metadata{
			Version:     CurrentVersion,
			BuiltAt:     time.Now(),
			RootPath:    "/project",
			FileCount:   3,
			SymbolCount: 5,
		},
		Files: []FileEntry{
			{Path: "main.go", Language: "go", Size: 100, Modified: time.Now()},
			{Path: "lib.py", Language: "python", Size: 200, Modified: time.Now()},
			{Path: "app.ts", Language: "typescript", Size: 300, Modified: time.Now()},
		},
		Symbols: []Symbol{
			{Name: "HandleRequest", Kind: "function", File: "main.go", Line: 10, Package: "main", Signature: "func HandleRequest(ctx context.Context, req *Request) (*Response, error)", Exported: true},
			{Name: "helperFunc", Kind: "function", File: "main.go", Line: 20, Package: "main", Signature: "func helperFunc()", Exported: false},
			{Name: "UserService", Kind: "class", File: "lib.py", Line: 5, Package: "lib", Signature: "class UserService:", Exported: true},
			{Name: "get_user", Kind: "method", File: "lib.py", Line: 15, Package: "lib", Signature: "def get_user(self, user_id):", Exported: true},
			{Name: "AppComponent", Kind: "class", File: "app.ts", Line: 3, Package: "app", Signature: "export class AppComponent", Exported: true},
		},
	}
}

func TestQuerySymbolsSubstring(t *testing.T) {
	idx := sampleIndex()

	results := QuerySymbols(idx, QueryOptions{Pattern: "handle"})
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	if results[0].Name != "HandleRequest" {
		t.Errorf("got %q, want HandleRequest", results[0].Name)
	}
}

func TestQuerySymbolsCaseInsensitive(t *testing.T) {
	idx := sampleIndex()

	results := QuerySymbols(idx, QueryOptions{Pattern: "HANDLEREQUEST"})
	if len(results) != 1 || results[0].Name != "HandleRequest" {
		t.Errorf("case-insensitive match failed: got %+v", results)
	}
}

func TestQuerySymbolsKindFilter(t *testing.T) {
	idx := sampleIndex()

	results := QuerySymbols(idx, QueryOptions{Kind: "class"})
	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}
}

func TestQuerySymbolsExportedOnly(t *testing.T) {
	idx := sampleIndex()

	results := QuerySymbols(idx, QueryOptions{ExportedOnly: true})
	if len(results) != 4 {
		t.Fatalf("got %d results, want 4", len(results))
	}
	for _, r := range results {
		if !r.Exported {
			t.Errorf("got unexported symbol %q in exported-only results", r.Name)
		}
	}
}

func TestQuerySymbolsEmptyPattern(t *testing.T) {
	idx := sampleIndex()

	results := QuerySymbols(idx, QueryOptions{})
	if len(results) != 5 {
		t.Fatalf("got %d results, want 5 (all)", len(results))
	}
}

func TestQuerySymbolsNoMatch(t *testing.T) {
	idx := sampleIndex()

	results := QuerySymbols(idx, QueryOptions{Pattern: "nonexistent"})
	if len(results) != 0 {
		t.Fatalf("got %d results, want 0", len(results))
	}
}

func TestQueryFilesAll(t *testing.T) {
	idx := sampleIndex()

	results := QueryFiles(idx, "")
	if len(results) != 3 {
		t.Fatalf("got %d files, want 3", len(results))
	}
}

func TestQueryFilesLanguageFilter(t *testing.T) {
	idx := sampleIndex()

	results := QueryFiles(idx, "go")
	if len(results) != 1 {
		t.Fatalf("got %d files, want 1", len(results))
	}
	if results[0].Path != "main.go" {
		t.Errorf("got %q, want main.go", results[0].Path)
	}
}

func TestQueryFilesNoMatch(t *testing.T) {
	idx := sampleIndex()

	results := QueryFiles(idx, "rust")
	if len(results) != 0 {
		t.Fatalf("got %d files, want 0", len(results))
	}
}

func TestQuerySymbolsSignatureFilter(t *testing.T) {
	idx := sampleIndex()

	results := QuerySymbols(idx, QueryOptions{Signature: "context.Context"})
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	if results[0].Name != "HandleRequest" {
		t.Errorf("got %q, want HandleRequest", results[0].Name)
	}
}

func TestQuerySymbolsSignatureAndPatternCompose(t *testing.T) {
	idx := sampleIndex()

	// Both match
	results := QuerySymbols(idx, QueryOptions{Pattern: "handle", Signature: "context.Context"})
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	if results[0].Name != "HandleRequest" {
		t.Errorf("got %q, want HandleRequest", results[0].Name)
	}

	// Pattern doesn't match
	results = QuerySymbols(idx, QueryOptions{Pattern: "query", Signature: "context.Context"})
	if len(results) != 0 {
		t.Fatalf("got %d results, want 0", len(results))
	}
}

func TestQuerySymbolsPackageFilter(t *testing.T) {
	idx := sampleIndex()

	results := QuerySymbols(idx, QueryOptions{Package: "main"})
	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}
	for _, r := range results {
		if r.Package != "main" {
			t.Errorf("got package %q, want main", r.Package)
		}
	}

	// Case-insensitive
	results = QuerySymbols(idx, QueryOptions{Package: "LIB"})
	if len(results) != 2 {
		t.Fatalf("got %d results, want 2 (lib package)", len(results))
	}
}

func TestQueryPackages(t *testing.T) {
	idx := sampleIndex()

	pkgs := QueryPackages(idx, "")
	if len(pkgs) != 3 {
		t.Fatalf("got %d packages, want 3", len(pkgs))
	}

	// Find the "main" package.
	var mainPkg *PackageSummary
	for i := range pkgs {
		if pkgs[i].Name == "main" {
			mainPkg = &pkgs[i]
			break
		}
	}
	if mainPkg == nil {
		t.Fatal("missing 'main' package")
	}
	if mainPkg.FileCount != 1 {
		t.Errorf("main FileCount = %d, want 1", mainPkg.FileCount)
	}
	if mainPkg.ExportedSymbols != 1 {
		t.Errorf("main ExportedSymbols = %d, want 1", mainPkg.ExportedSymbols)
	}
	if mainPkg.TotalSymbols != 2 {
		t.Errorf("main TotalSymbols = %d, want 2", mainPkg.TotalSymbols)
	}
	if mainPkg.Kinds["function"] != 2 {
		t.Errorf("main Kinds[function] = %d, want 2", mainPkg.Kinds["function"])
	}
}

func TestQueryPackagesFilter(t *testing.T) {
	idx := sampleIndex()

	pkgs := QueryPackages(idx, "lib")
	if len(pkgs) != 1 {
		t.Fatalf("got %d packages, want 1", len(pkgs))
	}
	if pkgs[0].Name != "lib" {
		t.Errorf("got package %q, want lib", pkgs[0].Name)
	}

	// Case-insensitive
	pkgs = QueryPackages(idx, "LIB")
	if len(pkgs) != 1 {
		t.Fatalf("got %d packages, want 1 (case-insensitive)", len(pkgs))
	}
}

func TestStatusFresh(t *testing.T) {
	dir := t.TempDir()
	now := time.Now()

	path := filepath.Join(dir, "main.go")
	if err := os.WriteFile(path, []byte("package main"), 0644); err != nil {
		t.Fatal(err)
	}
	info, _ := os.Stat(path)

	idx := &Index{
		Metadata: Metadata{
			Version:     CurrentVersion,
			BuiltAt:     now,
			FileCount:   1,
			SymbolCount: 1,
		},
		Files: []FileEntry{
			{Path: "main.go", Language: "go", Modified: info.ModTime()},
		},
	}

	result := Status(idx, dir)
	if !result.Exists {
		t.Error("expected Exists = true")
	}
	if result.FileCount != 1 {
		t.Errorf("FileCount = %d, want 1", result.FileCount)
	}
	if result.StaleFiles != 0 {
		t.Errorf("StaleFiles = %d, want 0", result.StaleFiles)
	}
}

func TestStatusStaleFiles(t *testing.T) {
	dir := t.TempDir()

	path := filepath.Join(dir, "main.go")
	if err := os.WriteFile(path, []byte("package main"), 0644); err != nil {
		t.Fatal(err)
	}

	idx := &Index{
		Metadata: Metadata{
			Version:     CurrentVersion,
			BuiltAt:     time.Now().Add(-time.Hour),
			FileCount:   1,
			SymbolCount: 1,
		},
		Files: []FileEntry{
			{Path: "main.go", Language: "go", Modified: time.Now().Add(-2 * time.Hour)},
		},
	}

	result := Status(idx, dir)
	if result.StaleFiles != 1 {
		t.Errorf("StaleFiles = %d, want 1", result.StaleFiles)
	}
}

func TestStatusMissingFile(t *testing.T) {
	dir := t.TempDir()

	idx := &Index{
		Metadata: Metadata{
			Version:     CurrentVersion,
			BuiltAt:     time.Now(),
			FileCount:   1,
			SymbolCount: 0,
		},
		Files: []FileEntry{
			{Path: "deleted.go", Language: "go", Modified: time.Now()},
		},
	}

	result := Status(idx, dir)
	if result.StaleFiles != 1 {
		t.Errorf("StaleFiles = %d, want 1 (deleted file)", result.StaleFiles)
	}
}

func TestStatusLanguages(t *testing.T) {
	idx := sampleIndex()
	result := Status(idx, "")

	if result.Languages["go"] != 1 {
		t.Errorf("go count = %d, want 1", result.Languages["go"])
	}
	if result.Languages["python"] != 1 {
		t.Errorf("python count = %d, want 1", result.Languages["python"])
	}
	if result.Languages["typescript"] != 1 {
		t.Errorf("typescript count = %d, want 1", result.Languages["typescript"])
	}
}
