package search

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type stubResponse struct {
	out string
	err error
}

// routeStub builds a runner that dispatches commands by full argv string,
// joined with single spaces. Tests register the exact command line they
// expect to see; unexpected commands fail the test loudly so we never
// silently match the wrong call.
func routeStub(t *testing.T, routes map[string]stubResponse) (runner, *[]string) {
	t.Helper()
	var seen []string
	return func(_ context.Context, name string, args ...string) ([]byte, error) {
		key := name + " " + strings.Join(args, " ")
		key = strings.TrimSpace(key)
		seen = append(seen, key)
		r, ok := routes[key]
		if !ok {
			t.Fatalf("unexpected command: %s", key)
		}
		return []byte(r.out), r.err
	}, &seen
}

func TestIsAvailable(t *testing.T) {
	t.Run("qmd on path returns true and caches", func(t *testing.T) {
		resetAvailabilityCache()
		t.Cleanup(resetAvailabilityCache)

		callCount := 0
		c := NewClient().WithRunner(func(_ context.Context, name string, args ...string) ([]byte, error) {
			callCount++
			return []byte("qmd 1.0.0"), nil
		})

		if !c.IsAvailable(context.Background()) {
			t.Fatal("expected IsAvailable to be true")
		}
		// Second call: cached, runner not re-invoked.
		if !c.IsAvailable(context.Background()) {
			t.Fatal("expected IsAvailable to be true on second call")
		}
		if callCount != 1 {
			t.Fatalf("expected runner to be called once (cached), got %d", callCount)
		}
	})

	t.Run("qmd missing returns false", func(t *testing.T) {
		resetAvailabilityCache()
		t.Cleanup(resetAvailabilityCache)

		c := NewClient().WithRunner(func(_ context.Context, name string, args ...string) ([]byte, error) {
			return nil, errors.New("exec: \"qmd\": executable file not found in $PATH")
		})

		if c.IsAvailable(context.Background()) {
			t.Fatal("expected IsAvailable to be false when qmd missing")
		}
	})
}

func TestStatusParser(t *testing.T) {
	// Isolate from any host filesystem cache — these tests assert solely on
	// the parsed qmd-status output, with the on-disk fallback neutralized.
	dir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", dir)

	cases := []struct {
		name              string
		statusOut         string
		statusErr         error
		isAvailable       bool
		wantInstalled     bool
		wantDaemonRunning bool
		wantModelsReady   bool
		wantErr           bool
	}{
		{
			name:              "mcp running with models ready",
			statusOut:         "MCP: running (PID 123)\nModels: ready\n",
			isAvailable:       true,
			wantInstalled:     true,
			wantDaemonRunning: true,
			wantModelsReady:   true,
		},
		{
			name:              "mcp not running, models ready",
			statusOut:         "MCP: stopped\nModels: ready\n",
			isAvailable:       true,
			wantInstalled:     true,
			wantDaemonRunning: false,
			wantModelsReady:   true,
		},
		{
			name:              "mcp running, models not present",
			statusOut:         "MCP: running\n",
			isAvailable:       true,
			wantInstalled:     true,
			wantDaemonRunning: true,
			wantModelsReady:   false,
		},
		{
			name:          "qmd missing",
			isAvailable:   false,
			wantInstalled: false,
		},
		{
			name:              "malformed status output",
			statusOut:         "garbage data with no recognizable markers",
			isAvailable:       true,
			wantInstalled:     true,
			wantDaemonRunning: false,
			wantModelsReady:   false,
		},
		{
			name:          "qmd status returns error",
			statusOut:     "partial output",
			statusErr:     errors.New("qmd status: exit 1"),
			isAvailable:   true,
			wantInstalled: true,
			wantErr:       true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resetAvailabilityCache()
			t.Cleanup(resetAvailabilityCache)

			run := func(_ context.Context, name string, args ...string) ([]byte, error) {
				if len(args) > 0 && args[0] == "--version" {
					if tc.isAvailable {
						return []byte("qmd 1.0.0"), nil
					}
					return nil, errors.New("not found")
				}
				if len(args) > 0 && args[0] == "status" {
					return []byte(tc.statusOut), tc.statusErr
				}
				t.Fatalf("unexpected command: %s %v", name, args)
				return nil, nil
			}

			c := NewClient().WithRunner(run)
			state, err := c.Status(context.Background())

			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error from Status, got nil")
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if state.Installed != tc.wantInstalled {
				t.Errorf("Installed: got %v, want %v", state.Installed, tc.wantInstalled)
			}
			if state.DaemonRunning != tc.wantDaemonRunning {
				t.Errorf("DaemonRunning: got %v, want %v", state.DaemonRunning, tc.wantDaemonRunning)
			}
			if state.ModelsReady != tc.wantModelsReady {
				t.Errorf("ModelsReady: got %v, want %v", state.ModelsReady, tc.wantModelsReady)
			}
		})
	}
}

func TestEnsureCollection(t *testing.T) {
	// Pre-warm the availability cache so tests don't need to mock --version.
	resetAvailabilityCache()
	t.Cleanup(resetAvailabilityCache)

	makeDir := func(t *testing.T) (rpiDir, name string) {
		t.Helper()
		dir := t.TempDir()
		rpiDir = filepath.Join(dir, "myrepo", ".rpi")
		if err := os.MkdirAll(rpiDir, 0o755); err != nil {
			t.Fatal(err)
		}
		n, err := CollectionName(rpiDir)
		if err != nil {
			t.Fatal(err)
		}
		return rpiDir, n
	}

	t.Run("creates collection and sets context on first run", func(t *testing.T) {
		rpiDir, name := makeDir(t)
		abs, _ := filepath.Abs(rpiDir)

		routes := map[string]stubResponse{
			"qmd collection add " + abs + " --name " + name:           {out: "Collection '" + name + "' created successfully"},
			"qmd context add qmd://" + name + " " + CollectionContext: {out: "ctx added"},
		}
		run, seen := routeStub(t, routes)
		c := NewClient().WithRunner(run)

		got, err := c.EnsureCollection(context.Background(), rpiDir)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != name {
			t.Errorf("got name %q, want %q", got, name)
		}
		if len(*seen) != 2 {
			t.Errorf("expected add+context calls, got %d: %v", len(*seen), *seen)
		}
	})

	// Real-world fixture: this is the exact stderr+stdout text qmd emits when
	// `qmd collection add . --name <existing-name>` is called twice in a row.
	// Captured manually from `qmd 1.x` on 2026-04-30. EnsureCollection must
	// treat this as a successful no-op rather than a hard failure — repeat
	// invocations are the steady state once a project has bootstrapped once.
	t.Run("treats 'already exists' as no-op on repeat run", func(t *testing.T) {
		rpiDir, name := makeDir(t)
		abs, _ := filepath.Abs(rpiDir)

		alreadyExistsOutput := "Collection '" + name + "' already exists.\nUse a different name with --name <name>\n"
		routes := map[string]stubResponse{
			"qmd collection add " + abs + " --name " + name: {
				out: alreadyExistsOutput,
				err: errors.New("exit status 1"),
			},
		}
		run, seen := routeStub(t, routes)
		c := NewClient().WithRunner(run)

		got, err := c.EnsureCollection(context.Background(), rpiDir)
		if err != nil {
			t.Fatalf("expected no error on already-exists, got %v", err)
		}
		if got != name {
			t.Errorf("got name %q, want %q", got, name)
		}
		// Context add must NOT run on the already-exists path — the context
		// was set on the original creation.
		if len(*seen) != 1 {
			t.Errorf("expected only the add call, got %d: %v", len(*seen), *seen)
		}
	})

	t.Run("propagates other add failures", func(t *testing.T) {
		rpiDir, name := makeDir(t)
		abs, _ := filepath.Abs(rpiDir)

		routes := map[string]stubResponse{
			"qmd collection add " + abs + " --name " + name: {
				out: "Error: database locked",
				err: errors.New("exit status 1"),
			},
		}
		run, _ := routeStub(t, routes)
		c := NewClient().WithRunner(run)

		_, err := c.EnsureCollection(context.Background(), rpiDir)
		if err == nil {
			t.Fatal("expected error to propagate")
		}
	})

	t.Run("tolerates context add already-exists", func(t *testing.T) {
		// Defensive: if a prior bootstrap created the collection but a second
		// invocation hits the context-add step (e.g., qmd collection was
		// removed but context wasn't), don't fail bootstrap on the context's
		// already-exists report.
		rpiDir, name := makeDir(t)
		abs, _ := filepath.Abs(rpiDir)

		routes := map[string]stubResponse{
			"qmd collection add " + abs + " --name " + name: {out: "added"},
			"qmd context add qmd://" + name + " " + CollectionContext: {
				out: "Context for 'qmd://" + name + "' already exists",
				err: errors.New("exit status 1"),
			},
		}
		run, _ := routeStub(t, routes)
		c := NewClient().WithRunner(run)

		got, err := c.EnsureCollection(context.Background(), rpiDir)
		if err != nil {
			t.Fatalf("expected context already-exists to be tolerated, got %v", err)
		}
		if got != name {
			t.Errorf("got name %q, want %q", got, name)
		}
	})
}

func TestModelsCachedOnDisk(t *testing.T) {
	t.Run("returns true when gguf file present", func(t *testing.T) {
		dir := t.TempDir()
		t.Setenv("XDG_CACHE_HOME", dir)
		modelsDir := filepath.Join(dir, "qmd", "models")
		if err := os.MkdirAll(modelsDir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(modelsDir, "embed.gguf"), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
		if !modelsCachedOnDisk() {
			t.Fatal("expected true with gguf file present")
		}
	})

	t.Run("returns false when models dir empty", func(t *testing.T) {
		dir := t.TempDir()
		t.Setenv("XDG_CACHE_HOME", dir)
		if err := os.MkdirAll(filepath.Join(dir, "qmd", "models"), 0o755); err != nil {
			t.Fatal(err)
		}
		if modelsCachedOnDisk() {
			t.Fatal("expected false when models dir empty")
		}
	})

	t.Run("returns false when models dir missing", func(t *testing.T) {
		dir := t.TempDir()
		t.Setenv("XDG_CACHE_HOME", dir)
		if modelsCachedOnDisk() {
			t.Fatal("expected false when models dir missing")
		}
	})
}
