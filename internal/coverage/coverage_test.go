package coverage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func analyzeFixture(t *testing.T, name string) Result {
	t.Helper()
	content, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("reading fixture %s: %v", name, err)
	}
	// An empty temp dir as root: no fixture relies on real files existing on
	// disk, so edit-target existence is deterministically governed by the
	// plan's own create declarations.
	return Analyze(string(content), t.TempDir())
}

func totalFindings(r Result) int {
	return len(r.Coverage.OrphanedCriteria) +
		len(r.Coverage.UncoveredFiles) +
		len(r.Coverage.UnjustifiedFiles) +
		len(r.Ordering.ForwardRefs) +
		len(r.Ordering.Cycles) +
		len(r.Existence.MissingEditTargets) +
		len(r.Existence.DoubleCreated)
}

func TestAnalyze_Fixtures(t *testing.T) {
	tests := []struct {
		name            string
		wantHardFailure bool
		check           func(t *testing.T, r Result)
	}{
		{
			name:            "clean.md",
			wantHardFailure: false,
			check: func(t *testing.T, r Result) {
				if n := totalFindings(r); n != 0 {
					t.Errorf("clean: expected 0 findings, got %d: %+v", n, r)
				}
			},
		},
		{
			name:            "single-phase.md",
			wantHardFailure: false,
			check: func(t *testing.T, r Result) {
				if n := totalFindings(r); n != 0 {
					t.Errorf("single-phase: expected 0 findings, got %d: %+v", n, r)
				}
				if len(r.Ordering.ForwardRefs) != 0 || len(r.Existence.DoubleCreated) != 0 {
					t.Errorf("single-phase: cross-slice findings are impossible, got %+v", r)
				}
			},
		},
		{
			name:            "forward-ref.md",
			wantHardFailure: false,
			check: func(t *testing.T, r Result) {
				if len(r.Ordering.ForwardRefs) != 1 {
					t.Fatalf("forward-ref: want 1 forwardRef, got %+v", r.Ordering.ForwardRefs)
				}
				fr := r.Ordering.ForwardRefs[0]
				if fr.File != "pkg/c.go" || fr.EditPhase != 1 || fr.CreatePhase != 2 {
					t.Errorf("forward-ref: unexpected forward ref %+v", fr)
				}
			},
		},
		{
			name:            "orphaned-criterion.md",
			wantHardFailure: true,
			check: func(t *testing.T, r Result) {
				if len(r.Coverage.OrphanedCriteria) != 1 {
					t.Fatalf("orphaned: want 1 orphaned criterion, got %+v", r.Coverage.OrphanedCriteria)
				}
				if r.Coverage.OrphanedCriteria[0] != "documentation is published" {
					t.Errorf("orphaned: unexpected text %q", r.Coverage.OrphanedCriteria[0])
				}
			},
		},
		{
			name:            "double-create.md",
			wantHardFailure: false,
			check: func(t *testing.T, r Result) {
				if len(r.Existence.DoubleCreated) != 1 || r.Existence.DoubleCreated[0] != "pkg/e.go" {
					t.Errorf("double-create: want [pkg/e.go], got %+v", r.Existence.DoubleCreated)
				}
			},
		},
		{
			name:            "uncovered-file.md",
			wantHardFailure: true,
			check: func(t *testing.T, r Result) {
				if len(r.Coverage.UncoveredFiles) != 1 || r.Coverage.UncoveredFiles[0] != "pkg/ghost.go" {
					t.Errorf("uncovered: want [pkg/ghost.go], got %+v", r.Coverage.UncoveredFiles)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := analyzeFixture(t, tt.name)
			if r.HardFailure != tt.wantHardFailure {
				t.Errorf("%s: hardFailure = %v, want %v (%+v)", tt.name, r.HardFailure, tt.wantHardFailure, r)
			}
			tt.check(t, r)
		})
	}
}

// TestAnalyze_Deterministic asserts that the same plan against the same tree
// yields a byte-identical verdict across repeated runs (map iteration order
// must not leak into the output).
func TestAnalyze_Deterministic(t *testing.T) {
	fixtures := []string{"clean.md", "forward-ref.md", "orphaned-criterion.md", "double-create.md", "uncovered-file.md"}
	for _, name := range fixtures {
		content, err := os.ReadFile(filepath.Join("testdata", name))
		if err != nil {
			t.Fatalf("reading fixture %s: %v", name, err)
		}
		var prev string
		for i := 0; i < 5; i++ {
			r := Analyze(string(content), "")
			b, err := json.Marshal(r)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			if i > 0 && string(b) != prev {
				t.Fatalf("%s: non-deterministic output:\n%s\nvs\n%s", name, prev, string(b))
			}
			prev = string(b)
		}
	}
}
