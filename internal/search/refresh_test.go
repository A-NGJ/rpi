package search

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRefresh(t *testing.T) {
	t.Run("clean update with no embed needed", func(t *testing.T) {
		resetDebounce()
		t.Cleanup(resetDebounce)

		routes := map[string]stubResponse{
			"qmd update --collection test --json": {
				out: `{"indexed":0,"updated":2,"unchanged":10,"removed":0,"needsEmbedding":0}`,
			},
		}
		run, seen := routeStub(t, routes)
		c := NewClient().WithRunner(run)

		result, err := c.Refresh(context.Background(), "test")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Updated != 2 || result.Unchanged != 10 {
			t.Errorf("unexpected counts: %+v", result)
		}
		if len(*seen) != 1 {
			t.Errorf("expected only update call (no embed), got %d: %v", len(*seen), *seen)
		}
	})

	t.Run("update with needsEmbedding triggers embed", func(t *testing.T) {
		resetDebounce()
		t.Cleanup(resetDebounce)

		routes := map[string]stubResponse{
			"qmd update --collection test --json": {
				out: `{"indexed":3,"updated":1,"unchanged":5,"removed":0,"needsEmbedding":4}`,
			},
			"qmd embed -c test": {out: "embedded 4 chunks"},
		}
		run, seen := routeStub(t, routes)
		c := NewClient().WithRunner(run)

		result, err := c.Refresh(context.Background(), "test")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.NeedsEmbedding != 4 {
			t.Errorf("expected NeedsEmbedding=4, got %d", result.NeedsEmbedding)
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
				out: `progress: scanning...
{"indexed":3,"updated":0,"unchanged":2,"removed":0,"needsEmbedding":0}
error: 1 file failed to parse`,
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
			"qmd update --collection test --json": {
				out: `{"indexed":1,"needsEmbedding":1}`,
			},
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
			"qmd update --collection test --json": {
				out: `{"indexed":0,"updated":1,"needsEmbedding":0}`,
			},
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
			name:  "pure json",
			input: `{"indexed":1,"updated":2,"unchanged":3,"removed":0,"needsEmbedding":1}`,
			want:  RefreshResult{Indexed: 1, Updated: 2, Unchanged: 3, NeedsEmbedding: 1},
		},
		{
			name:  "json with preamble",
			input: "scanning files...\n{\"indexed\":1,\"needsEmbedding\":0}",
			want:  RefreshResult{Indexed: 1},
		},
		{
			name:  "empty output",
			input: "",
			want:  RefreshResult{},
		},
		{
			name:    "no json present",
			input:   "just human text",
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
				got.Unchanged != tc.want.Unchanged || got.NeedsEmbedding != tc.want.NeedsEmbedding {
				t.Errorf("got %+v, want %+v", got, tc.want)
			}
		})
	}
}
