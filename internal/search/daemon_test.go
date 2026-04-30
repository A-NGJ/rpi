package search

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// newTestDaemonClient builds a daemonClient pointed at a httptest server.
func newTestDaemonClient(url string) *daemonClient {
	c := NewDaemonClient()
	c.url = url
	c.http = &http.Client{Timeout: 5 * time.Second}
	return c
}

func TestDaemonCallSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		var req jsonRPCRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.Method != "tools/call" {
			t.Errorf("expected tools/call method, got %s", req.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(jsonRPCResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: &toolResultBlock{
				Content: []toolContent{{Type: "text", Text: `[{"path":".rpi/designs/x.md","title":"X","score":0.9}]`}},
			},
		})
	}))
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

func TestDaemonCallNon2xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := newTestDaemonClient(srv.URL)
	_, err := c.Call(context.Background(), "query", nil)
	if err == nil {
		t.Fatal("expected error from 500 response")
	}
	if !errors.Is(err, errParse) {
		t.Errorf("expected errParse, got %v", err)
	}
}

func TestDaemonCallMalformedJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("not valid json"))
	}))
	defer srv.Close()

	c := newTestDaemonClient(srv.URL)
	_, err := c.Call(context.Background(), "query", nil)
	if err == nil {
		t.Fatal("expected error from malformed response")
	}
	if !errors.Is(err, errParse) {
		t.Errorf("expected errParse, got %v", err)
	}
}

func TestDaemonCallJSONRPCError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(jsonRPCResponse{
			JSONRPC: "2.0",
			ID:      1,
			Error:   &jsonRPCError{Code: -32600, Message: "bad request"},
		})
	}))
	defer srv.Close()

	c := newTestDaemonClient(srv.URL)
	_, err := c.Call(context.Background(), "query", nil)
	if err == nil {
		t.Fatal("expected error from JSON-RPC error response")
	}
	if !strings.Contains(err.Error(), "bad request") {
		t.Errorf("expected error message to surface, got %v", err)
	}
}

func TestDaemonURLEnvOverride(t *testing.T) {
	t.Setenv("QMD_MCP_URL", "http://example.com/custom")
	if got := DaemonURL(); got != "http://example.com/custom" {
		t.Errorf("expected env override, got %s", got)
	}
}
