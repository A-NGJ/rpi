package main

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestServeCmd_RegisteredOnRoot(t *testing.T) {
	for _, cmd := range rootCmd.Commands() {
		if cmd.Name() == "serve" {
			return
		}
	}
	t.Error("serveCmd not registered on rootCmd")
}

// extractText returns the text content from an MCP tool result.
func extractText(t *testing.T, result *mcp.CallToolResult) string {
	t.Helper()
	if len(result.Content) == 0 {
		t.Fatal("expected non-empty content")
	}
	tc, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("expected *TextContent, got %T", result.Content[0])
	}
	return tc.Text
}

func TestHandleGitContext_ReturnsJSON(t *testing.T) {
	result, _, err := handleGitContext(context.Background(), nil, emptyInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := extractText(t, result)
	var m map[string]any
	if err := json.Unmarshal([]byte(text), &m); err != nil {
		t.Fatalf("invalid JSON: %v\ntext: %s", err, text)
	}
	if _, ok := m["branch"]; !ok {
		t.Error("expected 'branch' key in git context")
	}
}

func TestHandleGitChangedFiles_ReturnsJSON(t *testing.T) {
	result, _, err := handleGitChangedFiles(context.Background(), nil, emptyInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := extractText(t, result)
	var files []string
	if err := json.Unmarshal([]byte(text), &files); err != nil {
		t.Fatalf("invalid JSON: %v\ntext: %s", err, text)
	}
}

func TestHandleGitSensitiveCheck_ReturnsJSON(t *testing.T) {
	result, _, err := handleGitSensitiveCheck(context.Background(), nil, emptyInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := extractText(t, result)
	var matches []any
	if err := json.Unmarshal([]byte(text), &matches); err != nil {
		t.Fatalf("invalid JSON: %v\ntext: %s", err, text)
	}
}

func TestHandleIndexStatus_ReturnsJSON(t *testing.T) {
	result, _, err := handleIndexStatus(context.Background(), nil, emptyInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := extractText(t, result)
	var m map[string]any
	if err := json.Unmarshal([]byte(text), &m); err != nil {
		t.Fatalf("invalid JSON: %v\ntext: %s", err, text)
	}
}

func TestHandleArchiveScan_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	oldFlag := rpiDirFlag
	rpiDirFlag = dir
	defer func() { rpiDirFlag = oldFlag }()

	result, _, err := handleArchiveScan(context.Background(), nil, emptyInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := extractText(t, result)
	var results []archiveScanResult
	if err := json.Unmarshal([]byte(text), &results); err != nil {
		t.Fatalf("invalid JSON: %v\ntext: %s", err, text)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results in empty dir, got %d", len(results))
	}
}

func TestHandleArchiveScan_WithArchivable(t *testing.T) {
	dir := setupArchiveTestDir(t)
	oldFlag := rpiDirFlag
	rpiDirFlag = dir
	defer func() { rpiDirFlag = oldFlag }()

	result, _, err := handleArchiveScan(context.Background(), nil, emptyInput{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := extractText(t, result)
	var results []archiveScanResult
	if err := json.Unmarshal([]byte(text), &results); err != nil {
		t.Fatalf("invalid JSON: %v\ntext: %s", err, text)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 archivable artifacts, got %d", len(results))
	}
}
