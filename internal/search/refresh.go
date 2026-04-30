package search

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
)

// debounceWindow is how long a successful Refresh suppresses subsequent
// invocations on the same collection. Two seconds covers back-to-back skill
// calls (rpi-research → rpi-propose) without delaying refresh after real
// edits.
const debounceWindow = 2 * time.Second

// RefreshResult mirrors qmd's documented update() return shape, plus a
// Warnings slice for partial-success surfacing.
type RefreshResult struct {
	Indexed        int      `json:"indexed"`
	Updated        int      `json:"updated"`
	Unchanged      int      `json:"unchanged"`
	Removed        int      `json:"removed"`
	NeedsEmbedding int      `json:"needsEmbedding"`
	Warnings       []string `json:"warnings,omitempty"`
	Debounced      bool     `json:"debounced,omitempty"`
}

// RefreshError carries a typed error stage so Query can map a refresh
// failure to the correct backend_error response without re-classifying.
type RefreshError struct {
	Stage   ErrorStage
	Message string
}

func (e *RefreshError) Error() string {
	return fmt.Sprintf("%s: %s", e.Stage, e.Message)
}

// nowFn is the package's clock seam — tests swap it for deterministic
// debounce assertions.
var nowFn = time.Now

// debounceState is the per-process record of when each collection last
// refreshed. It is intentionally process-local: cross-process races are
// documented but not handled in v1.
type debounceState struct {
	mu   sync.Mutex
	last map[string]time.Time
}

var debounce = &debounceState{last: make(map[string]time.Time)}

// resetDebounce is a test-only helper that clears the debounce cache.
func resetDebounce() {
	debounce.mu.Lock()
	defer debounce.mu.Unlock()
	debounce.last = make(map[string]time.Time)
}

// shouldRefresh reports whether enough time has passed since the last
// successful refresh of the given collection. It also records the new
// timestamp atomically when the answer is yes, so a concurrent caller
// sees the lock-out immediately.
func (d *debounceState) shouldRefresh(name string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	if last, ok := d.last[name]; ok {
		if nowFn().Sub(last) < debounceWindow {
			return false
		}
	}
	d.last[name] = nowFn()
	return true
}

// Refresh brings the qmd index up to date for the named collection. It runs
// `qmd update`, parses the result, and only runs `qmd embed` when the update
// reports files needing fresh vectors. Back-to-back calls within the
// debounce window are suppressed and return Debounced: true with a zero
// RefreshResult.
//
// The error contract:
//   - nil error  → either a successful refresh (possibly with Warnings on
//     partial update) or a debounced no-op.
//   - *RefreshError with StageUpdate → update failed entirely with no
//     files processed; no embed was attempted.
//   - *RefreshError with StageEmbed  → update succeeded, embed failed.
func (c *Client) Refresh(ctx context.Context, collectionName string) (RefreshResult, error) {
	if !debounce.shouldRefresh(collectionName) {
		return RefreshResult{Debounced: true}, nil
	}

	out, runErr := c.run(ctx, "qmd", "update", "--collection", collectionName, "--json")
	result, parseErr := parseUpdateOutput(out)

	if runErr != nil {
		// Partial success: any files indexed/updated → treat as ok with
		// warnings rather than failure. This matches the design's
		// "partial-update success is `ok`" rule.
		if result.Indexed > 0 || result.Updated > 0 {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("partial update: %s", strings.TrimSpace(string(out))))
		} else {
			return RefreshResult{}, &RefreshError{
				Stage:   StageUpdate,
				Message: fmt.Sprintf("qmd update failed: %v (%s)", runErr, strings.TrimSpace(string(out))),
			}
		}
	} else if parseErr != nil {
		// Update exited zero but we couldn't parse its output — treat as a
		// recoverable warning, not a hard failure (qmd's JSON shape may
		// shift across versions).
		result.Warnings = append(result.Warnings,
			fmt.Sprintf("update output unparseable: %v", parseErr))
	}

	if result.NeedsEmbedding > 0 {
		if eOut, err := c.run(ctx, "qmd", "embed", "-c", collectionName); err != nil {
			return RefreshResult{}, &RefreshError{
				Stage:   StageEmbed,
				Message: fmt.Sprintf("qmd embed failed: %v (%s)", err, strings.TrimSpace(string(eOut))),
			}
		}
	}

	return result, nil
}

// parseUpdateOutput tolerates qmd outputs that mix human-readable preamble
// with a trailing JSON object — common when qmd writes progress before the
// final summary. We extract the last JSON object found and unmarshal it.
func parseUpdateOutput(out []byte) (RefreshResult, error) {
	trimmed := strings.TrimSpace(string(out))
	if trimmed == "" {
		return RefreshResult{}, nil
	}

	// Take the substring from the last '{' to the last '}' so progress
	// lines emitted before the JSON summary don't break parsing.
	openIdx := strings.LastIndex(trimmed, "{")
	closeIdx := strings.LastIndex(trimmed, "}")
	if openIdx == -1 || closeIdx == -1 || openIdx >= closeIdx {
		return RefreshResult{}, fmt.Errorf("no JSON object found in update output")
	}

	var result RefreshResult
	if err := json.Unmarshal([]byte(trimmed[openIdx:closeIdx+1]), &result); err != nil {
		return RefreshResult{}, fmt.Errorf("unmarshal update output: %w", err)
	}
	return result, nil
}
