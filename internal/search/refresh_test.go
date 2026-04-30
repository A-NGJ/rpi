package search

import (
	"context"
	"errors"
	"testing"
	"time"
)

// realQmdUpdateOutput is qmd's actual `qmd update --collection X --json`
// output (captured 2026-04-30 from `qmd 1.x`). qmd ignores `--json` here
// and prints the same human text either way — same root cause as the
// `collection list` regression. Used as the canonical fixture.
const realQmdUpdateOutput = `Updating 1 collection(s)...

[1/1] test (**/*.md)
Collection: /tmp/some/path/.rpi (**/*.md)

Indexed: 0 new, 0 updated, 103 unchanged, 0 removed

✓ All collections updated.

Run 'qmd embed' to update embeddings (103 unique hashes need vectors)
`

const realQmdUpdateOutputAllEmbedded = `Updating 1 collection(s)...

[1/1] test (**/*.md)

Indexed: 2 new, 1 updated, 50 unchanged, 0 removed

✓ All collections updated.
`

func TestRefresh(t *testing.T) {
	t.Run("clean update with no embed needed", func(t *testing.T) {
		resetDebounce()
		t.Cleanup(resetDebounce)

		routes := map[string]stubResponse{
			"qmd update --collection test --json": {out: realQmdUpdateOutputAllEmbedded},
		}
		run, seen := routeStub(t, routes)
		c := NewClient().WithRunner(run)

		result, err := c.Refresh(context.Background(), "test")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Indexed != 2 || result.Updated != 1 || result.Unchanged != 50 {
			t.Errorf("unexpected counts: %+v", result)
		}
		if result.NeedsEmbedding != 0 {
			t.Errorf("expected NeedsEmbedding=0, got %d", result.NeedsEmbedding)
		}
		if len(*seen) != 1 {
			t.Errorf("expected only update call (no embed), got %d: %v", len(*seen), *seen)
		}
	})

	t.Run("update with needsEmbedding triggers embed", func(t *testing.T) {
		resetDebounce()
		t.Cleanup(resetDebounce)

		routes := map[string]stubResponse{
			"qmd update --collection test --json": {out: realQmdUpdateOutput},
			"qmd embed -c test":                   {out: "embedded 103 chunks"},
		}
		run, seen := routeStub(t, routes)
		c := NewClient().WithRunner(run)

		result, err := c.Refresh(context.Background(), "test")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.NeedsEmbedding != 103 {
			t.Errorf("expected NeedsEmbedding=103, got %d", result.NeedsEmbedding)
		}
		if len(*seen) != 2 {
			t.Errorf("expected update+embed calls, got %d: %v", len(*seen), *seen)
		}
	})

	t.Run("partial update returns warnings, not error", func(t *testing.T) {
		resetDebounce()
		t.Cleanup(resetDebounce)

		// qmd update exits non-zero but reports some files indexed.
		routes := map[string]stubResponse{
			"qmd update --collection test --json": {
				out: "Indexed: 3 new, 0 updated, 2 unchanged, 0 removed\nerror: 1 file failed to parse",
				err: errors.New("exit status 1"),
			},
		}
		run, _ := routeStub(t, routes)
		c := NewClient().WithRunner(run)

		result, err := c.Refresh(context.Background(), "test")
		if err != nil {
			t.Fatalf("expected no error for partial update, got %v", err)
		}
		if len(result.Warnings) == 0 {
			t.Error("expected warnings on partial update, got none")
		}
		if result.Indexed != 3 {
			t.Errorf("expected Indexed=3 from partial output, got %d", result.Indexed)
		}
	})

	t.Run("total update failure returns typed error", func(t *testing.T) {
		resetDebounce()
		t.Cleanup(resetDebounce)

		routes := map[string]stubResponse{
			"qmd update --collection test --json": {
				out: "fatal: database locked",
				err: errors.New("exit status 1"),
			},
		}
		run, _ := routeStub(t, routes)
		c := NewClient().WithRunner(run)

		_, err := c.Refresh(context.Background(), "test")
		if err == nil {
			t.Fatal("expected error on total update failure")
		}
		var rErr *RefreshError
		if !errors.As(err, &rErr) {
			t.Fatalf("expected *RefreshError, got %T: %v", err, err)
		}
		if rErr.Stage != StageUpdate {
			t.Errorf("expected StageUpdate, got %s", rErr.Stage)
		}
	})

	t.Run("embed failure returns typed error", func(t *testing.T) {
		resetDebounce()
		t.Cleanup(resetDebounce)

		routes := map[string]stubResponse{
			"qmd update --collection test --json": {out: realQmdUpdateOutput},
			"qmd embed -c test": {
				out: "fatal: model load failed",
				err: errors.New("exit status 1"),
			},
		}
		run, _ := routeStub(t, routes)
		c := NewClient().WithRunner(run)

		_, err := c.Refresh(context.Background(), "test")
		if err == nil {
			t.Fatal("expected error on embed failure")
		}
		var rErr *RefreshError
		if !errors.As(err, &rErr) {
			t.Fatalf("expected *RefreshError, got %T: %v", err, err)
		}
		if rErr.Stage != StageEmbed {
			t.Errorf("expected StageEmbed, got %s", rErr.Stage)
		}
	})

	t.Run("debounce skips runner on rapid second call", func(t *testing.T) {
		resetDebounce()
		t.Cleanup(resetDebounce)

		// Freeze time so both calls fall within the debounce window.
		base := time.Now()
		nowFn = func() time.Time { return base }
		t.Cleanup(func() { nowFn = time.Now })

		routes := map[string]stubResponse{
			"qmd update --collection test --json": {out: realQmdUpdateOutputAllEmbedded},
		}
		run, seen := routeStub(t, routes)
		c := NewClient().WithRunner(run)

		// First call: refresh runs.
		if _, err := c.Refresh(context.Background(), "test"); err != nil {
			t.Fatalf("first refresh: %v", err)
		}
		// Second call within debounce window: no runner invocation.
		result, err := c.Refresh(context.Background(), "test")
		if err != nil {
			t.Fatalf("second refresh: %v", err)
		}
		if !result.Debounced {
			t.Error("expected Debounced=true on second call within window")
		}
		if len(*seen) != 1 {
			t.Errorf("expected runner called once, got %d times: %v", len(*seen), *seen)
		}

		// Advance the clock past the debounce window — third call should run.
		nowFn = func() time.Time { return base.Add(debounceWindow + time.Second) }
		if _, err := c.Refresh(context.Background(), "test"); err != nil {
			t.Fatalf("third refresh: %v", err)
		}
		if len(*seen) != 2 {
			t.Errorf("expected runner called twice after window expiry, got %d", len(*seen))
		}
	})
}

func TestParseUpdateOutput(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		want    RefreshResult
		wantErr bool
	}{
		{
			name:  "real qmd output with embeddings owed",
			input: realQmdUpdateOutput,
			want:  RefreshResult{Indexed: 0, Updated: 0, Unchanged: 103, Removed: 0, NeedsEmbedding: 103},
		},
		{
			name:  "real qmd output all embedded",
			input: realQmdUpdateOutputAllEmbedded,
			want:  RefreshResult{Indexed: 2, Updated: 1, Unchanged: 50, NeedsEmbedding: 0},
		},
		{
			name:  "summary line alone",
			input: "Indexed: 1 new, 2 updated, 3 unchanged, 4 removed",
			want:  RefreshResult{Indexed: 1, Updated: 2, Unchanged: 3, Removed: 4},
		},
		{
			name:  "empty output",
			input: "",
			want:  RefreshResult{},
		},
		{
			name:    "no recognizable summary",
			input:   "some unrelated text without an Indexed: line",
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseUpdateOutput([]byte(tc.input))
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Indexed != tc.want.Indexed || got.Updated != tc.want.Updated ||
				got.Unchanged != tc.want.Unchanged || got.Removed != tc.want.Removed ||
				got.NeedsEmbedding != tc.want.NeedsEmbedding {
				t.Errorf("got %+v, want %+v", got, tc.want)
			}
		})
	}
}
