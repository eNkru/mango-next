package plugin

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/eNkru/mango-next/internal/queue"
)

func TestListPluginIDs(t *testing.T) {
	dir := t.TempDir()

	// Create some plugin directories.
	for _, id := range []string{"plugin-a", "plugin-b", "not-a-plugin"} {
		pDir := filepath.Join(dir, id)
		os.MkdirAll(pDir, 0o755)
	}

	// Only plugin-a and plugin-b have info.json
	os.WriteFile(filepath.Join(dir, "plugin-a", "info.json"), []byte(`{"id":"a","title":"A","placeholder":"x","wait_seconds":1}`), 0o644)
	os.WriteFile(filepath.Join(dir, "plugin-b", "info.json"), []byte(`{"id":"b","title":"B","placeholder":"x","wait_seconds":1}`), 0o644)

	ids := listPluginIDs(dir)
	if len(ids) != 2 {
		t.Fatalf("expected 2 plugin IDs, got %d: %v", len(ids), ids)
	}

	m := make(map[string]bool)
	for _, id := range ids {
		m[id] = true
	}
	if !m["plugin-a"] || !m["plugin-b"] {
		t.Errorf("missing expected plugins, got %v", ids)
	}
}

func TestUpdaterDisabledWhenIntervalZero(t *testing.T) {
	dir := t.TempDir()
	q, err := queue.NewQueue(filepath.Join(dir, "queue.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer q.Close()

	u := NewUpdater(dir, q, 0)
	if u.intervalHours != 0 {
		t.Errorf("expected interval 0")
	}
}

func TestFilterMatchChapter(t *testing.T) {
	tests := []struct {
		name    string
		filter  Filter
		chapter map[string]any
		want    bool
	}{
		{
			name:    "nil value always matches",
			filter:  Filter{Key: "lang", Value: nil, Type: FilterString},
			chapter: map[string]any{"lang": "en"},
			want:    true,
		},
		{
			name:    "string match",
			filter:  Filter{Key: "lang", Value: "en", Type: FilterString},
			chapter: map[string]any{"lang": "en"},
			want:    true,
		},
		{
			name:    "string mismatch",
			filter:  Filter{Key: "lang", Value: "jp", Type: FilterString},
			chapter: map[string]any{"lang": "en"},
			want:    false,
		},
		{
			name:    "num-min match",
			filter:  Filter{Key: "pages", Value: float64(5), Type: FilterNumMin},
			chapter: map[string]any{"pages": "10"},
			want:    true,
		},
		{
			name:    "num-min mismatch",
			filter:  Filter{Key: "pages", Value: float64(15), Type: FilterNumMin},
			chapter: map[string]any{"pages": "10"},
			want:    false,
		},
		{
			name:    "num-max match",
			filter:  Filter{Key: "pages", Value: float64(20), Type: FilterNumMax},
			chapter: map[string]any{"pages": "10"},
			want:    true,
		},
		{
			name:    "date-min match",
			filter:  Filter{Key: "date", Value: float64(1000), Type: FilterDateMin},
			chapter: map[string]any{"date": "2000"},
			want:    true,
		},
		{
			name:    "array match",
			filter:  Filter{Key: "genres", Value: "action", Type: FilterArray},
			chapter: map[string]any{"genres": "action, comedy"},
			want:    true,
		},
		{
			name:    "array mismatch",
			filter:  Filter{Key: "genres", Value: "romance", Type: FilterArray},
			chapter: map[string]any{"genres": "action, comedy"},
			want:    false,
		},
		{
			name:    "array 'all' matches everything",
			filter:  Filter{Key: "genres", Value: "all", Type: FilterArray},
			chapter: map[string]any{"genres": "anything"},
			want:    true,
		},
		{
			name:    "missing key does not match",
			filter:  Filter{Key: "nonexistent", Value: "val", Type: FilterString},
			chapter: map[string]any{"lang": "en"},
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.filter.matchChapter(tt.chapter)
			if got != tt.want {
				t.Errorf("matchChapter = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSubscriptionMatchesChapter(t *testing.T) {
	sub := NewSubscription("p1", "m1", "M1", "Sub")
	sub.Filters = []Filter{
		{Key: "lang", Value: "en", Type: FilterString},
	}

	if !sub.matchesChapter(map[string]any{"lang": "en", "id": "ch1"}) {
		t.Error("should match en chapter")
	}
	if sub.matchesChapter(map[string]any{"lang": "jp", "id": "ch2"}) {
		t.Error("should not match jp chapter")
	}
}

func TestUpdaterCheckPlugin(t *testing.T) {
	// This is a lightweight integration test that verifies the updater
	// can load a plugin with subscriptions and check for new chapters.
	dir := t.TempDir()

	// Create a v2 plugin.
	pluginDir := filepath.Join(dir, "testplugin")
	os.MkdirAll(pluginDir, 0o755)
	os.WriteFile(filepath.Join(pluginDir, "info.json"), []byte(`{
		"id": "testplugin",
		"title": "Test Plugin",
		"placeholder": "Search...",
		"wait_seconds": 1,
		"api_version": 2
	}`), 0o644)
	os.WriteFile(filepath.Join(pluginDir, "index.js"), []byte(`
		function newChapters(mangaId, afterMs) {
			return JSON.stringify([
				{id: "ch-new-1", title: "New Chapter 1", pages: 10, manga_title: "Test Manga"},
				{id: "ch-new-2", title: "New Chapter 2", pages: 15, manga_title: "Test Manga"}
			]);
		}
	`), 0o644)

	// Create a subscription.
	list, err := LoadSubscriptionList(pluginDir)
	if err != nil {
		t.Fatal(err)
	}
	sub := NewSubscription("testplugin", "manga-1", "Test Manga", "My Subscription")
	if err := list.Add(sub); err != nil {
		t.Fatal(err)
	}

	// Set last_checked far in the past to trigger new chapters.
	sub.LastChecked = 1000 // way in the past
	list.Save()

	// Create queue.
	q, err := queue.NewQueue(filepath.Join(dir, "queue.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer q.Close()

	// Create updater and check the plugin.
	u := NewUpdater(dir, q, 24)
	u.checkPlugin("testplugin")

	// Wait a moment for async operations.
	time.Sleep(100 * time.Millisecond)

	// Check that jobs were added to the queue.
	count, err := q.Count()
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Errorf("expected 2 queue jobs, got %d", count)
	}
}

func TestUpdaterSkipsV1Plugins(t *testing.T) {
	dir := t.TempDir()

	pluginDir := filepath.Join(dir, "v1plugin")
	os.MkdirAll(pluginDir, 0o755)
	os.WriteFile(filepath.Join(pluginDir, "info.json"), []byte(`{
		"id": "v1plugin",
		"title": "V1 Plugin",
		"placeholder": "x",
		"wait_seconds": 1
	}`), 0o644)
	os.WriteFile(filepath.Join(pluginDir, "index.js"), []byte(`
		function listChapters(q) { return JSON.stringify({title: q, chapters: []}); }
	`), 0o644)

	q, err := queue.NewQueue(filepath.Join(dir, "queue.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer q.Close()

	u := NewUpdater(dir, q, 24)
	// Should not panic or error.
	u.checkPlugin("v1plugin")

	count, _ := q.Count()
	if count != 0 {
		t.Errorf("expected 0 queue jobs for v1 plugin, got %d", count)
	}
}

func TestUpdaterStartStop(t *testing.T) {
	dir := t.TempDir()
	q, err := queue.NewQueue(filepath.Join(dir, "queue.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer q.Close()

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	u := NewUpdater(dir, q, 24)
	// Start blocks until ctx cancelled; should not hang.
	done := make(chan struct{})
	go func() {
		u.Start(ctx)
		close(done)
	}()

	select {
	case <-done:
		// OK
	case <-time.After(2 * time.Second):
		t.Fatal("updater.Start did not return after cancellation")
	}
}

func TestDownloaderStartStop(t *testing.T) {
	dir := t.TempDir()
	q, err := queue.NewQueue(filepath.Join(dir, "queue.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer q.Close()

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	d := NewDownloader(q, filepath.Join(dir, "library"), dir)
	done := make(chan struct{})
	go func() {
		d.Start(ctx)
		close(done)
	}()

	select {
	case <-done:
		// OK
	case <-time.After(2 * time.Second):
		t.Fatal("downloader.Start did not return after cancellation")
	}
}
