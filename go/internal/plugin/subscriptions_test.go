package plugin

import (
	"os"
	"path/filepath"
	"testing"
)

func writePluginWithInfo(t *testing.T, dir, id, infoJSON string) string {
	t.Helper()
	pDir := filepath.Join(dir, id)
	if err := os.MkdirAll(pDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pDir, "info.json"), []byte(infoJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	return pDir
}

func TestNewSubscription(t *testing.T) {
	sub := NewSubscription("plugin1", "manga-123", "Naruto", "Naruto Sub")

	if sub.ID == "" {
		t.Error("expected non-empty ID")
	}
	if sub.PluginID != "plugin1" {
		t.Errorf("plugin_id = %q, want plugin1", sub.PluginID)
	}
	if sub.MangaID != "manga-123" {
		t.Errorf("manga_id = %q, want manga-123", sub.MangaID)
	}
	if sub.MangaTitle != "Naruto" {
		t.Errorf("manga_title = %q, want Naruto", sub.MangaTitle)
	}
	if sub.Name != "Naruto Sub" {
		t.Errorf("name = %q, want Naruto Sub", sub.Name)
	}
	if sub.CreatedAt == 0 {
		t.Error("created_at should be set")
	}
	if sub.LastChecked == 0 {
		t.Error("last_checked should be set")
	}
}

func TestSubscriptionListSaveLoad(t *testing.T) {
	dir := t.TempDir()

	// Create a plugin directory.
	pluginDir := filepath.Join(dir, "test-plugin")
	if err := os.MkdirAll(pluginDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Empty list should load without error.
	list, err := LoadSubscriptionList(pluginDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(list.Subscriptions) != 0 {
		t.Errorf("expected empty list, got %d items", len(list.Subscriptions))
	}

	// Add a subscription.
	sub := NewSubscription("test-plugin", "m1", "Manga Title", "My Sub")
	if err := list.Add(sub); err != nil {
		t.Fatal(err)
	}
	if len(list.Subscriptions) != 1 {
		t.Fatalf("expected 1 subscription, got %d", len(list.Subscriptions))
	}

	// Verify the file was written.
	subPath := filepath.Join(pluginDir, "subscriptions.json")
	if _, err := os.Stat(subPath); err != nil {
		t.Fatalf("subscriptions.json was not created: %v", err)
	}

	// Load again from disk.
	list2, err := LoadSubscriptionList(pluginDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(list2.Subscriptions) != 1 {
		t.Fatalf("reloaded: expected 1 subscription, got %d", len(list2.Subscriptions))
	}
	if list2.Subscriptions[0].ID != sub.ID {
		t.Errorf("reloaded subscription has different ID")
	}
	if list2.Subscriptions[0].MangaTitle != "Manga Title" {
		t.Errorf("reloaded manga_title = %q", list2.Subscriptions[0].MangaTitle)
	}
}

func TestSubscriptionListRemove(t *testing.T) {
	dir := t.TempDir()

	list, err := LoadSubscriptionList(dir)
	if err != nil {
		t.Fatal(err)
	}

	sub1 := NewSubscription("p1", "m1", "M1", "Sub 1")
	sub2 := NewSubscription("p1", "m2", "M2", "Sub 2")

	if err := list.Add(sub1); err != nil {
		t.Fatal(err)
	}
	if err := list.Add(sub2); err != nil {
		t.Fatal(err)
	}

	if len(list.Subscriptions) != 2 {
		t.Fatalf("expected 2 subscriptions, got %d", len(list.Subscriptions))
	}

	// Remove the first one.
	if err := list.Remove(sub1.ID); err != nil {
		t.Fatal(err)
	}
	if len(list.Subscriptions) != 1 {
		t.Fatalf("expected 1 after remove, got %d", len(list.Subscriptions))
	}
	if list.Subscriptions[0].ID != sub2.ID {
		t.Error("wrong subscription survived removal")
	}
}

func TestPluginSubscribeListUnsubscribe(t *testing.T) {
	dir := t.TempDir()

	// Create a v2 plugin that supports subscribing.
	writePluginWithInfo(t, dir, "test-sub", `{
		"id": "test-sub",
		"title": "Test Sub",
		"placeholder": "Search...",
		"wait_seconds": 1,
		"api_version": 2,
		"settings": {}
	}`)

	// Write index.js with newChapters
	if err := os.WriteFile(filepath.Join(dir, "test-sub", "index.js"), []byte(`
		function newChapters(mangaId, afterMs) {
			return JSON.stringify([]);
		}
	`), 0o644); err != nil {
		t.Fatal(err)
	}

	p, err := LoadPlugin(dir, "test-sub")
	if err != nil {
		t.Fatal(err)
	}

	if !p.CanSubscribe() {
		t.Fatal("v2 plugin with newChapters should be subscribable")
	}

	// Subscribe.
	sub, err := p.Subscribe("manga-1", "Naruto", "My Naruto Subscription")
	if err != nil {
		t.Fatal(err)
	}
	if sub == nil {
		t.Fatal("expected a subscription")
	}

	// List.
	subs, err := p.ListSubscriptions()
	if err != nil {
		t.Fatal(err)
	}
	if len(subs) != 1 {
		t.Fatalf("expected 1 subscription, got %d", len(subs))
	}
	if subs[0].MangaID != "manga-1" {
		t.Errorf("manga_id = %q", subs[0].MangaID)
	}

	// Unsubscribe.
	if err := p.Unsubscribe(sub.ID); err != nil {
		t.Fatal(err)
	}
	subs, _ = p.ListSubscriptions()
	if len(subs) != 0 {
		t.Errorf("expected 0 subscriptions after unsubscribe, got %d", len(subs))
	}
}

func TestV1PluginCannotSubscribe(t *testing.T) {
	dir := t.TempDir()

	writePluginWithInfo(t, dir, "v1-nosub", `{
		"id": "v1-nosub",
		"title": "V1 No Sub",
		"placeholder": "x",
		"wait_seconds": 1
	}`)

	if err := os.WriteFile(filepath.Join(dir, "v1-nosub", "index.js"), []byte(`
		function listChapters(q) { return JSON.stringify({title: q, chapters: []}); }
	`), 0o644); err != nil {
		t.Fatal(err)
	}

	p, err := LoadPlugin(dir, "v1-nosub")
	if err != nil {
		t.Fatal(err)
	}

	if p.CanSubscribe() {
		t.Error("v1 without newChapters should not be subscribable")
	}
}

func TestFilterJSONRoundtrip(t *testing.T) {
	f := Filter{Key: "lang", Value: "en", Type: FilterString}
	data, err := f.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}

	var f2 Filter
	if err := f2.UnmarshalJSON(data); err != nil {
		t.Fatal(err)
	}
	if f2.Key != "lang" || f2.Type != FilterString {
		t.Errorf("roundtrip failed: got %+v", f2)
	}
}

func TestFilterTypeString(t *testing.T) {
	tests := []struct {
		ft   FilterType
		want string
	}{
		{FilterString, "string"},
		{FilterNumMin, "number-min"},
		{FilterNumMax, "number-max"},
		{FilterDateMin, "date-min"},
		{FilterDateMax, "date-max"},
		{FilterArray, "array"},
	}
	for _, tt := range tests {
		if got := tt.ft.String(); got != tt.want {
			t.Errorf("FilterType(%d).String() = %q, want %q", tt.ft, got, tt.want)
		}
	}
}

func TestFilterTypeFromString(t *testing.T) {
	tests := []struct {
		s    string
		want FilterType
	}{
		{"string", FilterString},
		{"number-min", FilterNumMin},
		{"number-max", FilterNumMax},
		{"date-min", FilterDateMin},
		{"date-max", FilterDateMax},
		{"array", FilterArray},
	}
	for _, tt := range tests {
		got, err := FilterTypeFromString(tt.s)
		if err != nil {
			t.Errorf("FilterTypeFromString(%q) error: %v", tt.s, err)
			continue
		}
		if got != tt.want {
			t.Errorf("FilterTypeFromString(%q) = %d, want %d", tt.s, got, tt.want)
		}
	}
	// Unknown type.
	_, err := FilterTypeFromString("unknown")
	if err == nil {
		t.Error("expected error for unknown filter type")
	}
}

func TestSubscriptionFilters(t *testing.T) {
	tests := []struct {
		name    string
		filter  Filter
		chapter map[string]any
		want    bool
	}{
		{
			name:    "no filter always matches",
			filter:  Filter{Type: FilterString, Key: "lang", Value: nil},
			chapter: map[string]any{"lang": "en"},
			want:    true,
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

func TestSubscriptionWithFilters(t *testing.T) {
	dir := t.TempDir()

	list, err := LoadSubscriptionList(dir)
	if err != nil {
		t.Fatal(err)
	}

	sub := NewSubscription("p1", "m1", "M1", "Sub with filters")
	sub.Filters = []Filter{
		{Key: "lang", Value: "en", Type: FilterString},
	}

	if err := list.Add(sub); err != nil {
		t.Fatal(err)
	}

	// Reload and verify filters are preserved.
	list2, err := LoadSubscriptionList(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(list2.Subscriptions) != 1 {
		t.Fatalf("expected 1 sub, got %d", len(list2.Subscriptions))
	}
	if len(list2.Subscriptions[0].Filters) != 1 {
		t.Fatalf("expected 1 filter, got %d", len(list2.Subscriptions[0].Filters))
	}
	if list2.Subscriptions[0].Filters[0].Key != "lang" {
		t.Errorf("filter key = %q", list2.Subscriptions[0].Filters[0].Key)
	}
	if list2.Subscriptions[0].Filters[0].Type != FilterString {
		t.Errorf("filter type = %v, want FilterString", list2.Subscriptions[0].Filters[0].Type)
	}
}
