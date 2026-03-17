package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/A-NGJ/ai-agent-research-plan-implement-flow/internal/chain"
	"github.com/A-NGJ/ai-agent-research-plan-implement-flow/internal/frontmatter"
	"github.com/A-NGJ/ai-agent-research-plan-implement-flow/internal/git"
	"github.com/A-NGJ/ai-agent-research-plan-implement-flow/internal/index"
	"github.com/A-NGJ/ai-agent-research-plan-implement-flow/internal/scanner"
	"github.com/A-NGJ/ai-agent-research-plan-implement-flow/internal/spec"
	tmpl "github.com/A-NGJ/ai-agent-research-plan-implement-flow/internal/template"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start MCP server exposing all rpi operations as typed tools",
	Long: `Start an MCP (Model Context Protocol) server over stdio.

Exposes all rpi CLI operations as typed MCP tools, enabling AI assistants
to call them with validated schemas instead of constructing shell commands.

Configure your MCP client with:
  {"command": "rpi", "args": ["serve"]}`,
	RunE: runServe,
}

func init() {
	addRpiDirFlag(serveCmd)
	addTemplatesDirFlag(serveCmd)
	rootCmd.AddCommand(serveCmd)
}

func runServe(cmd *cobra.Command, args []string) error {
	s := newRPIServer()
	return s.Run(context.Background(), &mcp.StdioTransport{})
}

func newRPIServer() *mcp.Server {
	s := mcp.NewServer(&mcp.Implementation{
		Name:    "rpi",
		Version: "dev",
	}, nil)
	registerTools(s)
	return s
}

// jsonResult marshals v to JSON and returns a TextContent MCP result.
func jsonResult(v any) (*mcp.CallToolResult, any, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, nil, fmt.Errorf("marshal result: %w", err)
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: string(data)}},
	}, nil, nil
}

type emptyInput struct{}

func registerTools(s *mcp.Server) {
	// No-param tools
	mcp.AddTool(s, &mcp.Tool{
		Name:        "rpi_git_context",
		Description: "Consolidated git state gathering",
	}, handleGitContext)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "rpi_git_changed_files",
		Description: "List files changed vs main branch",
	}, handleGitChangedFiles)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "rpi_git_sensitive_check",
		Description: "Scan staged files for sensitive filenames and content patterns",
	}, handleGitSensitiveCheck)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "rpi_index_status",
		Description: "Show index metadata and freshness",
	}, handleIndexStatus)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "rpi_archive_scan",
		Description: "Discover archivable artifacts with reference counts",
	}, handleArchiveScan)

	// Artifact tools
	mcp.AddTool(s, &mcp.Tool{
		Name:        "rpi_scan",
		Description: "Discover and filter artifacts in .rpi/",
	}, handleScan)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "rpi_scaffold",
		Description: "Generate artifact files from templates",
	}, handleScaffold)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "rpi_frontmatter_get",
		Description: "Read frontmatter fields from an artifact file",
	}, handleFrontmatterGet)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "rpi_frontmatter_set",
		Description: "Set a frontmatter field value in an artifact file",
	}, handleFrontmatterSet)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "rpi_frontmatter_transition",
		Description: "Validated status transition (enforces state machine)",
	}, handleFrontmatterTransition)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "rpi_chain",
		Description: "Resolve artifact cross-reference chain",
	}, handleChain)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "rpi_extract",
		Description: "Extract a section from a markdown file",
	}, handleExtract)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "rpi_extract_list_sections",
		Description: "List all section headings in a markdown file",
	}, handleExtractListSections)

	// Verification + index tools
	mcp.AddTool(s, &mcp.Tool{
		Name:        "rpi_verify_completeness",
		Description: "Check plan progress: checkbox counts and file coverage",
	}, handleVerifyCompleteness)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "rpi_verify_markers",
		Description: "Scan for TODO/FIXME/HACK markers in source files",
	}, handleVerifyMarkers)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "rpi_index_build",
		Description: "Build a symbol index of the codebase",
	}, handleIndexBuild)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "rpi_index_query",
		Description: "Search for symbols in the index",
	}, handleIndexQuery)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "rpi_index_files",
		Description: "List all indexed files",
	}, handleIndexFiles)

	// Archive + spec tools
	mcp.AddTool(s, &mcp.Tool{
		Name:        "rpi_archive_check_refs",
		Description: "Find all files referencing a given path",
	}, handleArchiveCheckRefs)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "rpi_archive_move",
		Description: "Archive an artifact: update frontmatter and move to archive/",
	}, handleArchiveMove)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "rpi_spec_coverage",
		Description: "Check test coverage for spec behaviors",
	}, handleSpecCoverage)
}

// --- No-param tool handlers ---

func handleGitContext(_ context.Context, _ *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, any, error) {
	result, err := git.GatherContext()
	if err != nil {
		return nil, nil, err
	}
	return jsonResult(result)
}

func handleGitChangedFiles(_ context.Context, _ *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, any, error) {
	files, err := git.ChangedFiles()
	if err != nil {
		return nil, nil, err
	}
	return jsonResult(files)
}

func handleGitSensitiveCheck(_ context.Context, _ *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, any, error) {
	matches, err := git.SensitiveCheck()
	if err != nil {
		return nil, nil, err
	}
	return jsonResult(matches)
}

func handleIndexStatus(_ context.Context, _ *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, any, error) {
	idx, err := index.Load(index.DefaultIndexPath)
	if err != nil {
		return jsonResult(map[string]bool{"exists": false})
	}
	result := index.Status(idx, idx.Metadata.RootPath)
	result.IndexPath = index.DefaultIndexPath
	if info, statErr := os.Stat(index.DefaultIndexPath); statErr == nil {
		result.IndexSizeBytes = info.Size()
	}
	return jsonResult(result)
}

func handleArchiveScan(_ context.Context, _ *mcp.CallToolRequest, _ emptyInput) (*mcp.CallToolResult, any, error) {
	results, err := scanner.Scan(rpiDirFlag, scanner.Filters{Archivable: true})
	if err != nil {
		return nil, nil, err
	}

	var output []archiveScanResult
	for _, r := range results {
		refPath := r.Path
		if rel, relErr := filepath.Rel(rpiDirFlag, r.Path); relErr == nil {
			refPath = rel
		}
		refCount, _ := scanner.CountReferences(rpiDirFlag, refPath)
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
	return jsonResult(output)
}

// --- Parameterized tool input structs ---

type scanInput struct {
	Type       string `json:"type,omitempty" jsonschema:"filter by artifact type (plan, proposal, research, spec, review)"`
	Status     string `json:"status,omitempty" jsonschema:"filter by status (draft, active, complete, superseded)"`
	Proposal   string `json:"proposal,omitempty" jsonschema:"filter by frontmatter proposal field"`
	References string `json:"references,omitempty" jsonschema:"find files that reference this path"`
	Archivable bool   `json:"archivable,omitempty" jsonschema:"show only archivable artifacts"`
}

type scaffoldInput struct {
	Type     string `json:"type" jsonschema:"artifact type: research, propose, plan, verify-report, spec"`
	Topic    string `json:"topic" jsonschema:"topic or title for the artifact"`
	Ticket   string `json:"ticket,omitempty" jsonschema:"ticket ID"`
	Research string `json:"research,omitempty" jsonschema:"path to research document"`
	Proposal string `json:"proposal,omitempty" jsonschema:"path to proposal document"`
	Spec     string `json:"spec,omitempty" jsonschema:"path to spec document"`
	Tags     string `json:"tags,omitempty" jsonschema:"comma-separated tags"`
	Write    bool   `json:"write,omitempty" jsonschema:"write to file instead of returning content"`
	Force    bool   `json:"force,omitempty" jsonschema:"allow overwriting existing files"`
}

type frontmatterGetInput struct {
	File  string `json:"file" jsonschema:"path to the artifact file"`
	Field string `json:"field,omitempty" jsonschema:"specific field to read (omit for all frontmatter)"`
}

type frontmatterSetInput struct {
	File  string `json:"file" jsonschema:"path to the artifact file"`
	Field string `json:"field" jsonschema:"frontmatter field name"`
	Value string `json:"value" jsonschema:"value to set"`
}

type frontmatterTransitionInput struct {
	File   string `json:"file" jsonschema:"path to the artifact file"`
	Status string `json:"status" jsonschema:"target status (draft, active, complete, superseded, archived)"`
}

type chainInput struct {
	Path     string `json:"path" jsonschema:"path to the root artifact"`
	Sections string `json:"sections,omitempty" jsonschema:"comma-separated section names to extract from each artifact"`
}

type extractInput struct {
	Path    string `json:"path" jsonschema:"path to the markdown file"`
	Section string `json:"section" jsonschema:"section heading to extract (case-insensitive prefix match)"`
}

type extractListSectionsInput struct {
	Path string `json:"path" jsonschema:"path to the markdown file"`
}

type verifyCompletenessInput struct {
	PlanPath string `json:"plan_path" jsonschema:"path to the plan file"`
}

type verifyMarkersInput struct {
	FilePath string `json:"file_path,omitempty" jsonschema:"specific file to scan (omit to scan git-changed files)"`
}

type indexBuildInput struct {
	Lang  string `json:"lang,omitempty" jsonschema:"comma-separated languages to index (e.g. go,py,ts)"`
	Path  string `json:"path,omitempty" jsonschema:"root path to index (default: current directory)"`
	Force bool   `json:"force,omitempty" jsonschema:"force full rebuild"`
}

type indexQueryInput struct {
	Pattern  string `json:"pattern" jsonschema:"substring pattern to match symbol names"`
	Kind     string `json:"kind,omitempty" jsonschema:"filter by symbol kind (function, method, class, struct, interface, type_alias)"`
	Exported bool   `json:"exported,omitempty" jsonschema:"show only exported symbols"`
}

type indexFilesInput struct {
	Lang string `json:"lang,omitempty" jsonschema:"filter by language"`
}

type archiveCheckRefsInput struct {
	Path string `json:"path" jsonschema:"path to check for references"`
}

type archiveMoveInput struct {
	Path  string `json:"path" jsonschema:"path to the artifact to archive"`
	Force bool   `json:"force,omitempty" jsonschema:"skip active reference check"`
}

type specCoverageInput struct {
	SpecFile string `json:"spec_file" jsonschema:"path to the spec file"`
}

// --- Parameterized tool handlers ---

func handleScan(_ context.Context, _ *mcp.CallToolRequest, input scanInput) (*mcp.CallToolResult, any, error) {
	results, err := scanner.Scan(rpiDirFlag, scanner.Filters{
		Status:     input.Status,
		Type:       input.Type,
		Proposal:   input.Proposal,
		References: input.References,
		Archivable: input.Archivable,
	})
	if err != nil {
		return nil, nil, err
	}
	return jsonResult(results)
}

func handleScaffold(_ context.Context, _ *mcp.CallToolRequest, input scaffoldInput) (*mcp.CallToolResult, any, error) {
	if _, ok := typeDirs[input.Type]; !ok {
		return nil, nil, fmt.Errorf("unknown artifact type %q; valid types: %v", input.Type, validTypes)
	}
	if input.Topic == "" {
		return nil, nil, fmt.Errorf("topic is required")
	}

	ctx := &tmpl.RenderContext{
		Type:     input.Type,
		Topic:    input.Topic,
		Ticket:   input.Ticket,
		Research: input.Research,
		Proposal: input.Proposal,
		Spec:     input.Spec,
		Tags:     input.Tags,
	}

	labels := map[string]string{
		"research": "Research", "plan": "Plan", "propose": "Proposal",
		"verify-report": "Verification Report", "spec": "Spec",
	}
	ctx.TypeLabel = labels[input.Type]

	if err := tmpl.ResolveAutoVars(ctx); err != nil {
		return nil, nil, err
	}
	ctx.Filename = tmpl.GenerateFilename(input.Type, ctx)

	tplDir := templatesDirFlag
	if tplDir == "" {
		tplDir = ".claude/templates"
	}
	output, err := tmpl.RenderTemplate(input.Type, ctx, tplDir)
	if err != nil {
		return nil, nil, err
	}

	if !input.Write {
		return jsonResult(map[string]string{"content": output, "filename": ctx.Filename})
	}

	subdir := typeDirs[input.Type]
	dir := filepath.Join(rpiDirFlag, subdir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, nil, err
	}
	outPath := filepath.Join(dir, ctx.Filename)
	if !input.Force {
		if _, err := os.Stat(outPath); err == nil {
			return nil, nil, fmt.Errorf("file already exists: %s (use force=true to overwrite)", outPath)
		}
	}
	if err := os.WriteFile(outPath, []byte(output), 0644); err != nil {
		return nil, nil, err
	}
	return jsonResult(map[string]string{"path": outPath})
}

func handleFrontmatterGet(_ context.Context, _ *mcp.CallToolRequest, input frontmatterGetInput) (*mcp.CallToolResult, any, error) {
	doc, err := frontmatter.Parse(input.File)
	if err != nil {
		return nil, nil, err
	}
	if input.Field != "" {
		val, ok := doc.Frontmatter[input.Field]
		if !ok {
			return jsonResult(nil)
		}
		return jsonResult(val)
	}
	return jsonResult(doc.Frontmatter)
}

func handleFrontmatterSet(_ context.Context, _ *mcp.CallToolRequest, input frontmatterSetInput) (*mcp.CallToolResult, any, error) {
	doc, err := frontmatter.Parse(input.File)
	if err != nil {
		return nil, nil, err
	}
	doc.Frontmatter[input.Field] = input.Value
	if err := frontmatter.Write(doc); err != nil {
		return nil, nil, err
	}
	return jsonResult(map[string]string{"status": "ok", "file": input.File, "field": input.Field, "value": input.Value})
}

func handleFrontmatterTransition(_ context.Context, _ *mcp.CallToolRequest, input frontmatterTransitionInput) (*mcp.CallToolResult, any, error) {
	doc, err := frontmatter.Parse(input.File)
	if err != nil {
		return nil, nil, err
	}
	if err := frontmatter.Transition(doc, input.Status); err != nil {
		return nil, nil, err
	}
	if err := frontmatter.Write(doc); err != nil {
		return nil, nil, err
	}
	return jsonResult(map[string]string{"status": "ok", "file": input.File, "new_status": input.Status})
}

func handleChain(_ context.Context, _ *mcp.CallToolRequest, input chainInput) (*mcp.CallToolResult, any, error) {
	opts := chain.ResolveOptions{}
	if input.Sections != "" {
		parts := strings.Split(input.Sections, ",")
		for i, p := range parts {
			parts[i] = strings.TrimSpace(p)
		}
		opts.Sections = parts
	}
	result, err := chain.Resolve(input.Path, opts)
	if err != nil {
		return nil, nil, err
	}
	return jsonResult(result)
}

func handleExtract(_ context.Context, _ *mcp.CallToolRequest, input extractInput) (*mcp.CallToolResult, any, error) {
	doc, err := frontmatter.Parse(input.Path)
	if err != nil {
		return nil, nil, err
	}
	content, ok := frontmatter.ExtractSection(doc.Body, input.Section)
	if !ok {
		return nil, nil, fmt.Errorf("section %q not found", input.Section)
	}
	return jsonResult(map[string]string{
		"path": input.Path, "section": input.Section, "content": content,
	})
}

func handleExtractListSections(_ context.Context, _ *mcp.CallToolRequest, input extractListSectionsInput) (*mcp.CallToolResult, any, error) {
	doc, err := frontmatter.Parse(input.Path)
	if err != nil {
		return nil, nil, err
	}
	sections := frontmatter.ListSections(doc.Body)
	return jsonResult(map[string]any{
		"path": input.Path, "sections": sections,
	})
}

func handleVerifyCompleteness(_ context.Context, _ *mcp.CallToolRequest, input verifyCompletenessInput) (*mcp.CallToolResult, any, error) {
	content, err := os.ReadFile(input.PlanPath)
	if err != nil {
		return nil, nil, fmt.Errorf("reading plan: %w", err)
	}
	checkboxes := parseCheckboxes(string(content))
	planFiles := extractPlanFiles(string(content))
	gitFiles, err := git.ChangedFiles()
	if err != nil {
		gitFiles = []string{}
	}
	compare := comparePlanVsGit(planFiles, gitFiles)
	result := CompletenessResult{
		CheckboxResult: checkboxes,
		CompareResult:  compare,
	}
	return jsonResult(result)
}

func handleVerifyMarkers(_ context.Context, _ *mcp.CallToolRequest, input verifyMarkersInput) (*mcp.CallToolResult, any, error) {
	files, err := filesToScan(input.FilePath)
	if err != nil {
		return nil, nil, err
	}
	result := MarkersResult{
		Markers: []Marker{},
		Count:   map[string]int{"TODO": 0, "FIXME": 0, "HACK": 0},
	}
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		markers := scanMarkers(file, string(data))
		result.Markers = append(result.Markers, markers...)
		for _, m := range markers {
			result.Count[m.Type]++
		}
	}
	return jsonResult(result)
}

func handleIndexBuild(_ context.Context, _ *mcp.CallToolRequest, input indexBuildInput) (*mcp.CallToolResult, any, error) {
	opts := index.BuildOptions{
		ForceRebuild: input.Force,
	}
	if input.Lang != "" {
		opts.Languages = strings.Split(input.Lang, ",")
	}
	path := input.Path
	if path == "" {
		path = "."
	}
	idx, err := index.Build(path, opts)
	if err != nil {
		return nil, nil, fmt.Errorf("build index: %w", err)
	}
	absPath, _ := filepath.Abs(path)
	indexPath := filepath.Join(absPath, index.DefaultIndexPath)
	if err := index.Save(idx, indexPath); err != nil {
		return nil, nil, fmt.Errorf("save index: %w", err)
	}
	return jsonResult(map[string]any{
		"files":   idx.Metadata.FileCount,
		"symbols": idx.Metadata.SymbolCount,
		"path":    index.DefaultIndexPath,
	})
}

func handleIndexQuery(_ context.Context, _ *mcp.CallToolRequest, input indexQueryInput) (*mcp.CallToolResult, any, error) {
	idx, err := index.Load(index.DefaultIndexPath)
	if err != nil {
		return nil, nil, fmt.Errorf("no index found — run rpi_index_build first")
	}
	results := index.QuerySymbols(idx, index.QueryOptions{
		Pattern:      input.Pattern,
		Kind:         input.Kind,
		ExportedOnly: input.Exported,
	})
	if results == nil {
		results = []index.Symbol{}
	}
	return jsonResult(results)
}

func handleIndexFiles(_ context.Context, _ *mcp.CallToolRequest, input indexFilesInput) (*mcp.CallToolResult, any, error) {
	idx, err := index.Load(index.DefaultIndexPath)
	if err != nil {
		return nil, nil, fmt.Errorf("no index found — run rpi_index_build first")
	}
	results := index.QueryFiles(idx, input.Lang)
	if results == nil {
		results = []index.FileEntry{}
	}
	return jsonResult(results)
}

func handleArchiveCheckRefs(_ context.Context, _ *mcp.CallToolRequest, input archiveCheckRefsInput) (*mcp.CallToolResult, any, error) {
	refs, err := scanner.FindReferences(rpiDirFlag, input.Path)
	if err != nil {
		return nil, nil, err
	}
	return jsonResult(refs)
}

func handleArchiveMove(_ context.Context, _ *mcp.CallToolRequest, input archiveMoveInput) (*mcp.CallToolResult, any, error) {
	result, err := doArchiveMove(input.Path, rpiDirFlag, input.Force, time.Now())
	if err == errHasReferences {
		return nil, nil, fmt.Errorf("file has active references (use force=true to override)")
	}
	if err != nil {
		return nil, nil, err
	}
	return jsonResult(result)
}

func handleSpecCoverage(_ context.Context, _ *mcp.CallToolRequest, input specCoverageInput) (*mcp.CallToolResult, any, error) {
	behaviors, prefix, err := spec.ParseBehaviors(input.SpecFile)
	if err != nil {
		return nil, nil, err
	}
	doc, err := frontmatter.Parse(input.SpecFile)
	if err != nil {
		return nil, nil, err
	}
	domain, _ := doc.Frontmatter["domain"].(string)
	projectRoot, err := findProjectRoot()
	if err != nil {
		return nil, nil, err
	}
	refs, err := spec.ScanTestFiles(projectRoot, prefix)
	if err != nil {
		return nil, nil, err
	}
	report := spec.ComputeCoverage(behaviors, refs, domain, prefix)
	return jsonResult(report)
}
