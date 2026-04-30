package search

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

// CollectionContext is the description attached to RPI's qmd collection. qmd
// returns this alongside hits to help its reranker score relevance.
const CollectionContext = "RPI artifacts: research, designs, behavioral specs, plans, diagnoses"

// runner is the exec indirection used by all qmd shell-outs. Tests substitute
// a stub via WithRunner; production code uses defaultRunner.
type runner func(ctx context.Context, name string, args ...string) ([]byte, error)

func defaultRunner(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	return cmd.CombinedOutput()
}

// Client wraps qmd command execution with a configurable runner so tests can
// substitute stubs without touching PATH.
type Client struct {
	run runner
}

// NewClient returns a Client that uses the real qmd binary on PATH.
func NewClient() *Client {
	return &Client{run: defaultRunner}
}

// WithRunner returns a copy of c with the runner replaced. Used by tests.
func (c *Client) WithRunner(r runner) *Client {
	return &Client{run: r}
}

// availabilityCache is a process-level cache of IsAvailable's result so the
// PATH lookup runs at most once per process.
type availabilityCache struct {
	once   sync.Once
	result bool
}

var availability = &availabilityCache{}

// IsAvailable reports whether the qmd binary is on PATH. The result is cached
// for the process lifetime — callers that want to retry after install must
// restart the process.
func (c *Client) IsAvailable(ctx context.Context) bool {
	availability.once.Do(func() {
		_, err := c.run(ctx, "qmd", "--version")
		availability.result = err == nil
	})
	return availability.result
}

// resetAvailabilityCache is a test-only helper to clear the cached result
// between table-driven cases.
func resetAvailabilityCache() {
	availability = &availabilityCache{}
}

// BackendState captures the readiness of the qmd backend for a Query call.
// All three booleans must be true for a search to proceed end-to-end.
type BackendState struct {
	Installed     bool   `json:"installed"`
	DaemonRunning bool   `json:"daemon_running"`
	ModelsReady   bool   `json:"models_ready"`
	RawStatus     string `json:"raw_status,omitempty"`
}

var (
	mcpRunningRe  = regexp.MustCompile(`(?i)\bMCP\s*:\s*running\b`)
	modelsReadyRe = regexp.MustCompile(`(?i)\b(models?\s*:\s*(ready|loaded)|all\s+models\s+(present|ready|loaded))\b`)
)

// Status probes qmd for daemon and model readiness. The returned BackendState
// always carries RawStatus when qmd was reachable, so callers can surface
// diagnostic context even when parsing fails. An error is returned only when
// qmd itself is missing or the runner fails to invoke it; in that case
// Installed is false.
func (c *Client) Status(ctx context.Context) (BackendState, error) {
	if !c.IsAvailable(ctx) {
		return BackendState{Installed: false}, nil
	}

	out, err := c.run(ctx, "qmd", "status")
	state := BackendState{Installed: true, RawStatus: string(out)}
	if err != nil {
		// qmd exists on PATH but the status command failed — return what we
		// have so the caller can map it to backend_error{parse}.
		return state, err
	}

	state.DaemonRunning = mcpRunningRe.MatchString(state.RawStatus)
	state.ModelsReady = modelsReadyRe.MatchString(state.RawStatus) || modelsCachedOnDisk()
	return state, nil
}

// alreadyExistsMarker is the substring qmd emits when `collection add` is
// called with a name that's already registered. The full message looks like:
//
//	Collection 'rpi-foo-abc123' already exists.
//	Use a different name with --name <name>
//
// Detecting this lets EnsureCollection treat re-bootstrap as a no-op without
// needing a separate list-then-compare step (qmd's `collection list` doesn't
// support --json, and our naming scheme makes drift repair unnecessary —
// each absolute path produces a distinct, deterministic collection name, so
// "name conflict" can only mean "this exact path was already bootstrapped").
const alreadyExistsMarker = "already exists"

// EnsureCollection registers the project's .rpi/ as a qmd collection. The
// collection name is derived from the absolute path (see CollectionName), so
// each project gets a unique, deterministic identifier — different projects
// never collide and a moved repo gets a fresh collection automatically.
//
// The returned name is what should be passed to subsequent qmd calls
// (`-c <name>`). On second and later calls for the same path, qmd reports
// "already exists" — we treat that as success.
func (c *Client) EnsureCollection(ctx context.Context, rpiDir string) (string, error) {
	name, err := CollectionName(rpiDir)
	if err != nil {
		return "", err
	}
	absRpi, err := filepath.Abs(rpiDir)
	if err != nil {
		return "", fmt.Errorf("resolve rpiDir: %w", err)
	}

	out, err := c.run(ctx, "qmd", "collection", "add", absRpi, "--name", name)
	if err != nil {
		// Already-bootstrapped is the steady-state path on repeat calls;
		// treat it as success without re-running context add.
		if strings.Contains(string(out), alreadyExistsMarker) {
			return name, nil
		}
		return "", fmt.Errorf("qmd collection add: %w (%s)", err, strings.TrimSpace(string(out)))
	}

	// First-time creation: also set the descriptive context qmd's reranker
	// uses to bias relevance for our artifacts.
	if cOut, cErr := c.run(ctx, "qmd", "context", "add", "qmd://"+name, CollectionContext); cErr != nil {
		// Don't fail bootstrap if context already exists — the collection
		// itself was just created so the context is the only thing left,
		// and qmd may legitimately treat re-add as an error.
		if strings.Contains(string(cOut), alreadyExistsMarker) {
			return name, nil
		}
		return "", fmt.Errorf("qmd context add: %w (%s)", cErr, strings.TrimSpace(string(cOut)))
	}
	return name, nil
}

// modelsCachedOnDisk is a defensive backstop for when qmd's status output
// doesn't explicitly enumerate model readiness — we check the documented
// cache location for any .gguf file.
func modelsCachedOnDisk() bool {
	cacheRoot := os.Getenv("XDG_CACHE_HOME")
	if cacheRoot == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return false
		}
		cacheRoot = filepath.Join(home, ".cache")
	}
	modelsDir := filepath.Join(cacheRoot, "qmd", "models")
	entries, err := os.ReadDir(modelsDir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(strings.ToLower(e.Name()), ".gguf") {
			return true
		}
	}
	return false
}
