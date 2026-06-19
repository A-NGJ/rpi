package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/A-NGJ/rpi/internal/coverage"
)

// TestVerifyCoveragePreLock runs the built binary against the committed
// internal/coverage fixtures and asserts the JSON verdict + exit behavior.
// The command always exits 0 and reports via the hardFailure flag — it is a
// report, not a gate (the skill decides whether to block).
func TestVerifyCoveragePreLock(t *testing.T) {
	binary := buildBinary(t)

	tests := []struct {
		name            string
		fixture         string
		wantHardFailure bool
		check           func(t *testing.T, r coverage.Result)
	}{
		{
			name:            "clean",
			fixture:         "internal/coverage/testdata/clean.md",
			wantHardFailure: false,
		},
		{
			name:            "forward-ref",
			fixture:         "internal/coverage/testdata/forward-ref.md",
			wantHardFailure: false,
			check: func(t *testing.T, r coverage.Result) {
				if len(r.Ordering.ForwardRefs) != 1 {
					t.Errorf("want 1 forwardRef, got %+v", r.Ordering.ForwardRefs)
				}
			},
		},
		{
			name:            "orphaned-criterion",
			fixture:         "internal/coverage/testdata/orphaned-criterion.md",
			wantHardFailure: true,
		},
		{
			name:            "uncovered-file",
			fixture:         "internal/coverage/testdata/uncovered-file.md",
			wantHardFailure: true,
			check: func(t *testing.T, r coverage.Result) {
				if len(r.Coverage.UncoveredFiles) != 1 {
					t.Errorf("want 1 uncovered file, got %+v", r.Coverage.UncoveredFiles)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stdout, stderr, exitCode := runRPI(t, binary, "verify", "coverage", "--pre-lock", tt.fixture)
			if exitCode != 0 {
				t.Fatalf("exit %d, stderr=%s", exitCode, stderr)
			}
			var r coverage.Result
			if err := json.Unmarshal([]byte(stdout), &r); err != nil {
				t.Fatalf("unmarshal verdict: %v\n%s", err, stdout)
			}
			if r.HardFailure != tt.wantHardFailure {
				t.Errorf("%s: hardFailure=%v, want %v\n%s", tt.name, r.HardFailure, tt.wantHardFailure, stdout)
			}
			if tt.check != nil {
				tt.check(t, r)
			}
		})
	}
}

// TestVerifyCoverageRequiresPreLock asserts the explicit flag contract: coverage
// only runs in --pre-lock mode today.
func TestVerifyCoverageRequiresPreLock(t *testing.T) {
	binary := buildBinary(t)
	_, stderr, exitCode := runRPI(t, binary, "verify", "coverage", "internal/coverage/testdata/clean.md")
	if exitCode == 0 {
		t.Errorf("expected non-zero exit without --pre-lock")
	}
	if !strings.Contains(stderr, "pre-lock") {
		t.Errorf("expected pre-lock hint in stderr, got %q", stderr)
	}
}
