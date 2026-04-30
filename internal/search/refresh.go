package search

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
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

// updateSummaryRe matches the canonical qmd update summary line:
//
//	Indexed: 0 new, 0 updated, 103 unchanged, 0 removed
//
// qmd silently ignores --json on `update` (same as on `collection list`),
// so we parse the human text. The regex tolerates extra whitespace.
var updateSummaryRe = regexp.MustCompile(
	`Indexed:\s*(\d+)\s+new,\s*(\d+)\s+updated,\s*(\d+)\s+unchanged,\s*(\d+)\s+removed`,
)

// needsEmbeddingRe matches the trailing hint qmd emits when fresh vectors
// are owed:
//
//	Run 'qmd embed' to update embeddings (103 unique hashes need vectors)
//
// Absent line means everything is embedded.
var needsEmbeddingRe = regexp.MustCompile(`\((\d+)\s+unique\s+hashes\s+need\s+vectors\)`)

// parseUpdateOutput extracts counts from qmd's human-format `update` output.
// Returns a zero-valued RefreshResult with no error when the output is
// empty (qmd printed nothing meaningful) or when no recognizable summary
// line is present — the caller treats that as "we don't know, assume
// nothing changed" rather than a hard failure.
func parseUpdateOutput(out []byte) (RefreshResult, error) {
	trimmed := strings.TrimSpace(string(out))
	if trimmed == "" {
		return RefreshResult{}, nil
	}

	var result RefreshResult
	if m := updateSummaryRe.FindStringSubmatch(trimmed); m != nil {
		result.Indexed = atoi(m[1])
		result.Updated = atoi(m[2])
		result.Unchanged = atoi(m[3])
		result.Removed = atoi(m[4])
	} else {
		return RefreshResult{}, fmt.Errorf("no update summary line found")
	}

	if m := needsEmbeddingRe.FindStringSubmatch(trimmed); m != nil {
		result.NeedsEmbedding = atoi(m[1])
	}

	return result, nil
}

// atoi is a forgiving integer parser — invalid input becomes 0 rather than
// propagating an error, since the caller has already established that the
// surrounding regex matched.
func atoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}
