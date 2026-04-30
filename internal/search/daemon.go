package search

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"
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

// daemonClient issues MCP Streamable-HTTP requests to qmd. Each call is a
// fresh JSON-RPC 2.0 request — the server is documented as stateless.
type daemonClient struct {
	url    string
	http   *http.Client
	nextID atomic.Int64
}

// NewDaemonClient builds a client pointed at the qmd MCP daemon. The HTTP
// timeout is generous because the first call after a fresh embed can run
// inference for several seconds.
func NewDaemonClient() *daemonClient {
	return &daemonClient{
		url:  DaemonURL(),
		http: &http.Client{Timeout: 60 * time.Second},
	}
}

// jsonRPCRequest is the request envelope sent to qmd.
type jsonRPCRequest struct {
	JSONRPC string         `json:"jsonrpc"`
	ID      int64          `json:"id"`
	Method  string         `json:"method"`
	Params  map[string]any `json:"params"`
}

// jsonRPCResponse is the envelope returned by qmd. We only consume Result;
// Error is surfaced as a parse-class failure with the embedded message.
type jsonRPCResponse struct {
	JSONRPC string           `json:"jsonrpc"`
	ID      int64            `json:"id"`
	Result  *toolResultBlock `json:"result,omitempty"`
	Error   *jsonRPCError    `json:"error,omitempty"`
}

type jsonRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// toolResultBlock matches MCP's typed tool-call response — qmd returns its
// structured payload as a single text content block holding JSON.
type toolResultBlock struct {
	Content []toolContent `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

type toolContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Call invokes a tool on the qmd MCP server using JSON-RPC's tools/call
// method. Returns the raw JSON inside the first text content block.
func (c *daemonClient) Call(ctx context.Context, toolName string, args map[string]any) (json.RawMessage, error) {
	id := c.nextID.Add(1)
	body, err := json.Marshal(jsonRPCRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  "tools/call",
		Params: map[string]any{
			"name":      toolName,
			"arguments": args,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		// Distinguish "daemon not listening" from generic transport errors so
		// Query can map it to a specific stage with a useful hint.
		if isConnRefused(err) {
			return nil, errDaemonNotRunning
		}
		return nil, fmt.Errorf("http call: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("%w: HTTP %d: %s", errParse, resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	var env jsonRPCResponse
	if err := json.Unmarshal(respBody, &env); err != nil {
		return nil, fmt.Errorf("%w: %v", errParse, err)
	}
	if env.Error != nil {
		return nil, fmt.Errorf("qmd: %s (code %d)", env.Error.Message, env.Error.Code)
	}
	if env.Result == nil || len(env.Result.Content) == 0 {
		return nil, fmt.Errorf("%w: empty result", errParse)
	}

	// qmd encodes the tool's structured payload as JSON inside a text block.
	return json.RawMessage(env.Result.Content[0].Text), nil
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
