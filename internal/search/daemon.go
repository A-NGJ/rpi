package search

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// defaultDaemonURL matches qmd's default HTTP MCP endpoint when launched
// with `qmd mcp --http --daemon`.
const defaultDaemonURL = "http://localhost:8181/mcp"

// DaemonURL returns the qmd MCP endpoint, honoring the QMD_MCP_URL
// environment variable when set.
func DaemonURL() string {
	if v := os.Getenv("QMD_MCP_URL"); v != "" {
		return v
	}
	return defaultDaemonURL
}

// Daemon is the seam between Query and the qmd HTTP MCP server. The
// production implementation is *daemonClient; tests substitute a stub.
type Daemon interface {
	Call(ctx context.Context, toolName string, args map[string]any) (json.RawMessage, error)
}

// errDaemonNotRunning is returned when the qmd daemon refuses connections.
// Query maps this to StatusBackendError{StageDaemonNotRunning}.
var errDaemonNotRunning = errors.New("qmd daemon not running")

// errParse is returned when the daemon's response can't be decoded into the
// expected MCP envelope. Query maps this to StatusBackendError{StageParse}.
var errParse = errors.New("qmd response parse failed")

// daemonClient speaks MCP's Streamable HTTP transport against qmd's
// `qmd mcp --http --daemon` server. Per the MCP spec the protocol requires
// an `initialize` handshake that yields an `Mcp-Session-Id` to be carried on
// every subsequent request — we delegate that bookkeeping to the SDK's
// StreamableClientTransport rather than reimplementing it.
//
// Each Call opens its own session because qmd indexes inference work per
// query and there's no shared client-side state worth amortizing across
// calls; the initialize round-trip is cheap relative to the actual query
// (qmd reports ~1s for context recreation, milliseconds for handshake).
type daemonClient struct {
	endpoint   string
	httpClient *http.Client
	impl       *mcp.Implementation
}

// NewDaemonClient builds a client pointed at the qmd MCP daemon. The HTTP
// timeout is generous because the first call after a fresh embed can run
// inference for several seconds.
func NewDaemonClient() *daemonClient {
	return &daemonClient{
		endpoint:   DaemonURL(),
		httpClient: &http.Client{Timeout: 120 * time.Second},
		impl:       &mcp.Implementation{Name: "rpi-search", Version: "dev"},
	}
}

// Call opens an MCP session, invokes the named tool with args, and returns
// the JSON payload from the first text content block. Connection-refused
// surfaces as errDaemonNotRunning; all other transport-class failures
// surface as errParse so Query can map them to specific stages.
func (c *daemonClient) Call(ctx context.Context, toolName string, args map[string]any) (json.RawMessage, error) {
	client := mcp.NewClient(c.impl, nil)
	transport := &mcp.StreamableClientTransport{
		Endpoint:   c.endpoint,
		HTTPClient: c.httpClient,
		// We only need request-response semantics; qmd doesn't push
		// server-initiated notifications we'd want to consume, and disabling
		// SSE avoids holding a long-lived background connection per Call.
		DisableStandaloneSSE: true,
	}

	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		if isConnRefused(err) {
			return nil, errDaemonNotRunning
		}
		return nil, fmt.Errorf("connect to qmd daemon: %w", err)
	}
	defer session.Close()

	result, err := session.CallTool(ctx, &mcp.CallToolParams{
		Name:      toolName,
		Arguments: args,
	})
	if err != nil {
		return nil, fmt.Errorf("call %s: %w", toolName, err)
	}
	if result.IsError {
		return nil, fmt.Errorf("qmd reported tool error for %s: %s", toolName, contentSummary(result.Content))
	}

	// Prefer the structured payload — qmd ships the typed result there. Fall
	// back to the first text content block (which qmd uses for the human-
	// formatted summary) only when the structured field is empty.
	if result.StructuredContent != nil {
		raw, err := json.Marshal(result.StructuredContent)
		if err != nil {
			return nil, fmt.Errorf("%w: re-marshal structured content: %v", errParse, err)
		}
		return raw, nil
	}

	if len(result.Content) == 0 {
		return nil, fmt.Errorf("%w: empty result", errParse)
	}
	text, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		return nil, fmt.Errorf("%w: first content block is %T, expected TextContent",
			errParse, result.Content[0])
	}
	return json.RawMessage(text.Text), nil
}

// contentSummary renders MCP content for embedding in error messages.
func contentSummary(content []mcp.Content) string {
	var parts []string
	for _, c := range content {
		if t, ok := c.(*mcp.TextContent); ok {
			parts = append(parts, t.Text)
		}
	}
	return strings.Join(parts, " ")
}

// isConnRefused detects the kernel-level "no listener" signal for a wide
// range of OS error wrappings. Matching by net.OpError + syscall codes
// keeps this portable across darwin/linux without reaching into syscall.
func isConnRefused(err error) bool {
	if err == nil {
		return false
	}
	var netErr *net.OpError
	if errors.As(err, &netErr) {
		// On both linux and darwin, the inner Err is a *os.SyscallError
		// whose Err is syscall.ECONNREFUSED — but rather than import
		// syscall, just look at the rendered message. This is robust
		// enough for our "is the daemon up?" check.
		s := netErr.Err.Error()
		if strings.Contains(s, "connection refused") {
			return true
		}
	}
	return strings.Contains(err.Error(), "connection refused")
}
