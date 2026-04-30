package search

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// newTestDaemonClient builds a daemonClient pointed at a httptest server.
func newTestDaemonClient(url string) *daemonClient {
	c := NewDaemonClient()
	c.endpoint = url
	c.httpClient = &http.Client{Timeout: 5 * time.Second}
	return c
}

// queryToolHandler is the test stub's tool handler. The handler captures
// each call's args (so tests can assert what was sent) and returns the
// canned content the test wants surfaced.
type queryToolArgs struct {
	Q          string  `json:"q"`
	Collection string  `json:"collection,omitempty"`
	N          int     `json:"n,omitempty"`
	MinScore   float64 `json:"min_score,omitempty"`
}

// newMCPTestServer wires an httptest.Server around the SDK's
// StreamableHTTPHandler and registers a single `query` tool whose
// response and error are controlled by the test's closures. This
// exercises the *real* MCP session protocol — initialize, session ID
// header, and tools/call — so a regression like the qmd "Missing session
// ID" bug couldn't slip through again.
func newMCPTestServer(t *testing.T, response string, toolErr error) *httptest.Server {
	t.Helper()
	server := mcp.NewServer(&mcp.Implementation{Name: "qmd-stub", Version: "test"}, nil)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "query",
		Description: "stub query",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args queryToolArgs) (*mcp.CallToolResult, any, error) {
		if toolErr != nil {
			return nil, nil, toolErr
		}
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: response}},
		}, nil, nil
	})
	handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return server }, nil)
	return httptest.NewServer(handler)
}

func TestDaemonCallSuccessOverRealMCPSession(t *testing.T) {
	// This is the headline regression test. With the SDK's
	// StreamableHTTPHandler we get a server that enforces the MCP session
	// protocol — initialize handshake, Mcp-Session-Id header, the works.
	// If our client ever stops handshaking properly, this test fails with
	// the same `Bad Request: Missing session ID` qmd surfaced in the wild.
	srv := newMCPTestServer(t, `[{"path":".rpi/designs/x.md","title":"X","score":0.9}]`, nil)
	defer srv.Close()

	c := newTestDaemonClient(srv.URL)
	raw, err := c.Call(context.Background(), "query", map[string]any{"q": "test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(raw), `"path":".rpi/designs/x.md"`) {
		t.Errorf("expected hit JSON in raw response, got %s", string(raw))
	}
}

func TestDaemonCallSessionIDIsCarried(t *testing.T) {
	// Direct check that the second-and-later requests in the same Call's
	// session carry an Mcp-Session-Id header. Wraps the SDK handler with
	// a sniffer that records every inbound request's headers.
	server := mcp.NewServer(&mcp.Implementation{Name: "qmd-stub", Version: "test"}, nil)
	mcp.AddTool(server, &mcp.Tool{Name: "query", Description: "stub"},
		func(ctx context.Context, req *mcp.CallToolRequest, args queryToolArgs) (*mcp.CallToolResult, any, error) {
			return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "[]"}}}, nil, nil
		})
	mcpHandler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return server }, nil)

	var requestCount atomic.Int32
	var sawSessionHeader atomic.Bool
	wrapper := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if requestCount.Add(1) > 1 {
			// First request is the initialize POST (sent before the server
			// has assigned a session ID); every later request must carry
			// the header back.
			if r.Header.Get("Mcp-Session-Id") != "" {
				sawSessionHeader.Store(true)
			}
		}
		mcpHandler.ServeHTTP(w, r)
	})
	srv := httptest.NewServer(wrapper)
	defer srv.Close()

	c := newTestDaemonClient(srv.URL)
	if _, err := c.Call(context.Background(), "query", map[string]any{"q": "x"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !sawSessionHeader.Load() {
		t.Error("expected Mcp-Session-Id header on follow-up requests; got none")
	}
}

func TestDaemonCallConnectionRefused(t *testing.T) {
	// Bind to a server then immediately close it so the address has no
	// listener, producing a guaranteed connection-refused.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	url := srv.URL
	srv.Close()

	c := newTestDaemonClient(url)
	_, err := c.Call(context.Background(), "query", nil)
	if err == nil {
		t.Fatal("expected error from closed server")
	}
	if !errors.Is(err, errDaemonNotRunning) {
		t.Errorf("expected errDaemonNotRunning, got %v", err)
	}
}

func TestDaemonCallToolError(t *testing.T) {
	srv := newMCPTestServer(t, "", errors.New("qmd: query failed"))
	defer srv.Close()

	c := newTestDaemonClient(srv.URL)
	_, err := c.Call(context.Background(), "query", map[string]any{"q": "x"})
	if err == nil {
		t.Fatal("expected error from failing tool")
	}
	if !strings.Contains(err.Error(), "query failed") {
		t.Errorf("expected qmd error message to surface, got %v", err)
	}
}

func TestDaemonURLEnvOverride(t *testing.T) {
	t.Setenv("QMD_MCP_URL", "http://example.com/custom")
	if got := DaemonURL(); got != "http://example.com/custom" {
		t.Errorf("expected env override, got %s", got)
	}
}
