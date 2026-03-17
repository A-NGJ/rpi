package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/A-NGJ/ai-agent-research-plan-implement-flow/internal/git"
	"github.com/A-NGJ/ai-agent-research-plan-implement-flow/internal/index"
	"github.com/A-NGJ/ai-agent-research-plan-implement-flow/internal/scanner"
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
}

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
