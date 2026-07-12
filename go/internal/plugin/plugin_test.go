package plugin

import (
	"os"
	"path/filepath"
	"testing"
)

func writePlugin(t *testing.T, dir, id, infoJSON, indexJS string) string {
	t.Helper()
	pDir := filepath.Join(dir, id)
	if err := os.MkdirAll(pDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pDir, "info.json"), []byte(infoJSON), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pDir, "index.js"), []byte(indexJS), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestLoadPluginV1(t *testing.T) {
	dir := t.TempDir()
	writePlugin(t, dir, "test-v1", `{
		"id": "test-v1",
		"title": "Test V1",
		"placeholder": "Search...",
		"wait_seconds": 1
	}`, `
		function listChapters(query) {
			return JSON.stringify({
				title: "Series: " + query,
				chapters: [{id: "ch1", title: "Chapter 1"}]
			});
		}
		function selectChapter(id) {
			return JSON.stringify({
				title: "Chapter Title",
				pages: [{url: "https://example.com/page1.jpg", filename: "page1.jpg"}]
			});
		}
		function nextPage() {
			return JSON.stringify({url: "https://example.com/page2.jpg", filename: "page2.jpg"});
		}
	`)

	p, err := LoadPlugin(dir, "test-v1")
	if err != nil {
		t.Fatal(err)
	}

	if p.Info().ID != "test-v1" {
		t.Errorf("id = %q, want test-v1", p.Info().ID)
	}
	if p.Info().APIVersion != 1 {
		t.Errorf("api_version = %d, want 1", p.Info().APIVersion)
	}
	if p.CanSubscribe() {
		t.Error("v1 plugin should not be subscribable")
	}

	result, err := p.ListChapters("naruto")
	if err != nil {
		t.Fatal(err)
	}
	obj, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("listChapters result type = %T, want map", result)
	}
	if obj["title"] != "Series: naruto" {
		t.Errorf("title = %v, want Series: naruto", obj["title"])
	}
	chapters, ok := obj["chapters"].([]any)
	if !ok {
		t.Fatalf("chapters type = %T", obj["chapters"])
	}
	if len(chapters) != 1 {
		t.Fatalf("got %d chapters, want 1", len(chapters))
	}
}

func TestLoadPluginV2(t *testing.T) {
	dir := t.TempDir()
	writePlugin(t, dir, "test-v2", `{
		"id": "test-v2",
		"title": "Test V2",
		"placeholder": "Search manga...",
		"wait_seconds": 2,
		"api_version": 2,
		"settings": {"quality": "high", "lang": "en"}
	}`, `
		function searchManga(query) {
			return JSON.stringify([
				{id: "m1", title: "Manga: " + query},
				{id: "m2", title: "Another: " + query}
			]);
		}
		function listChapters(mangaId) {
			return JSON.stringify([
				{id: "ch1", title: "Chapter 1", pages: 20, manga_title: "Test Manga"},
				{id: "ch2", title: "Chapter 2", pages: 25, manga_title: "Test Manga"}
			]);
		}
		function selectChapter(id) {
			return JSON.stringify({
				id: id,
				title: "Chapter Title",
				pages: [
					{url: "https://example.com/001.jpg", filename: "001.jpg"},
					{url: "https://example.com/002.jpg", filename: "002.jpg"}
				],
				manga_title: "Test Manga"
			});
		}
		function newChapters(mangaId, afterMs) {
			return JSON.stringify([]);
		}
	`)

	p, err := LoadPlugin(dir, "test-v2")
	if err != nil {
		t.Fatal(err)
	}

	if p.Info().APIVersion != 2 {
		t.Errorf("api_version = %d, want 2", p.Info().APIVersion)
	}
	if !p.CanSubscribe() {
		t.Error("v2 plugin with newChapters should be subscribable")
	}

	mangas, err := p.SearchManga("naruto")
	if err != nil {
		t.Fatal(err)
	}
	mangaArr, ok := mangas.([]any)
	if !ok {
		t.Fatalf("searchManga result type = %T, want []any", mangas)
	}
	if len(mangaArr) != 2 {
		t.Fatalf("got %d manga, want 2", len(mangaArr))
	}
	m1 := mangaArr[0].(map[string]any)
	if m1["id"] != "m1" || m1["title"] != "Manga: naruto" {
		t.Errorf("manga[0] = %v", m1)
	}

	chapters, err := p.ListChapters("m1")
	if err != nil {
		t.Fatal(err)
	}
	chArr, ok := chapters.([]any)
	if !ok {
		t.Fatalf("listChapters result type = %T, want []any", chapters)
	}
	if len(chArr) != 2 {
		t.Fatalf("got %d chapters, want 2", len(chArr))
	}
}

func TestSearchMangaV1Error(t *testing.T) {
	dir := t.TempDir()
	writePlugin(t, dir, "v1-plugin", `{
		"id": "v1-plugin",
		"title": "V1 Only",
		"placeholder": "x",
		"wait_seconds": 1
	}`, `function listChapters(q) { return JSON.stringify({title:q,chapters:[]}); }`)

	p, err := LoadPlugin(dir, "v1-plugin")
	if err != nil {
		t.Fatal(err)
	}
	_, err = p.SearchManga("naruto")
	if err == nil {
		t.Error("expected error for v1 searchManga")
	}
}

func TestMangoStorage(t *testing.T) {
	dir := t.TempDir()
	writePlugin(t, dir, "test-store", `{
		"id": "test-store",
		"title": "Storage Test",
		"placeholder": "x",
		"wait_seconds": 1
	}`, `
		function testSet(k, v) { mango.storage(k, v); return "ok"; }
		function testGet(k) { return mango.storage(k); }
	`)

	p, err := LoadPlugin(dir, "test-store")
	if err != nil {
		t.Fatal(err)
	}

	_, err = p.Eval("testSet('key1', 'value1')")
	if err != nil {
		t.Fatal(err)
	}

	val, err := p.Eval("testGet('key1')")
	if err != nil {
		t.Fatal(err)
	}
	if val != "value1" {
		t.Errorf("storage get = %v, want value1", val)
	}

	val, err = p.Eval("testGet('nonexistent')")
	if err != nil {
		t.Fatal(err)
	}
	if val != nil {
		t.Errorf("storage get nonexistent = %v, want nil", val)
	}
}

func TestMangoSettings(t *testing.T) {
	dir := t.TempDir()
	writePlugin(t, dir, "test-settings", `{
		"id": "test-settings",
		"title": "Settings Test",
		"placeholder": "x",
		"wait_seconds": 1,
		"api_version": 2,
		"settings": {"quality": "high", "lang": "en"}
	}`, `
		function getQuality() { return mango.settings('quality'); }
		function getLang() { return mango.settings('lang'); }
		function getMissing() { return mango.settings('nonexistent'); }
	`)

	p, err := LoadPlugin(dir, "test-settings")
	if err != nil {
		t.Fatal(err)
	}

	val, err := p.Eval("getQuality()")
	if err != nil {
		t.Fatal(err)
	}
	if val != "high" {
		t.Errorf("settings quality = %v, want high", val)
	}

	val, err = p.Eval("getLang()")
	if err != nil {
		t.Fatal(err)
	}
	if val != "en" {
		t.Errorf("settings lang = %v, want en", val)
	}

	val, err = p.Eval("getMissing()")
	if err != nil {
		t.Fatal(err)
	}
	if val != nil {
		t.Errorf("settings missing = %v, want nil", val)
	}
}

func TestSelectChapterV2(t *testing.T) {
	dir := t.TempDir()
	writePlugin(t, dir, "v2-chapter", `{
		"id": "v2-chapter",
		"title": "Chapter Test",
		"placeholder": "x",
		"wait_seconds": 1,
		"api_version": 2
	}`, `
		function selectChapter(id) {
			return JSON.stringify({
				id: id,
				title: "Awesome Chapter",
				pages: [
					{url: "https://ex.com/p1.jpg", filename: "p1.jpg"},
					{url: "https://ex.com/p2.jpg", filename: "p2.jpg"},
					{url: "https://ex.com/p3.jpg", filename: "p3.jpg"}
				],
				manga_title: "My Manga"
			});
		}
	`)

	p, err := LoadPlugin(dir, "v2-chapter")
	if err != nil {
		t.Fatal(err)
	}

	result, err := p.SelectChapter("ch-123")
	if err != nil {
		t.Fatal(err)
	}
	ch, ok := result.(map[string]any)
	if !ok {
		t.Fatalf("selectChapter result type = %T", result)
	}
	if ch["id"] != "ch-123" {
		t.Errorf("id = %v, want ch-123", ch["id"])
	}
	pages, ok := ch["pages"].([]any)
	if !ok {
		t.Fatalf("pages type = %T", ch["pages"])
	}
	if len(pages) != 3 {
		t.Fatalf("got %d pages, want 3", len(pages))
	}
}

func TestLoadPluginMissingInfo(t *testing.T) {
	dir := t.TempDir()
	_, err := LoadPlugin(dir, "nonexistent")
	if err == nil {
		t.Error("expected error for missing plugin")
	}
}

func TestLoadPluginMissingJS(t *testing.T) {
	dir := t.TempDir()
	pDir := filepath.Join(dir, "nojs")
	os.MkdirAll(pDir, 0o755)
	os.WriteFile(filepath.Join(pDir, "info.json"), []byte(`{"id":"nojs","title":"No JS","placeholder":"x","wait_seconds":1}`), 0o644)

	_, err := LoadPlugin(dir, "nojs")
	if err == nil {
		t.Error("expected error for missing index.js")
	}
}

func TestNewChaptersV1Error(t *testing.T) {
	dir := t.TempDir()
	writePlugin(t, dir, "v1-nc", `{
		"id": "v1-nc",
		"title": "V1",
		"placeholder": "x",
		"wait_seconds": 1
	}`, `function listChapters(q) { return JSON.stringify({title:q,chapters:[]}); }`)

	p, err := LoadPlugin(dir, "v1-nc")
	if err != nil {
		t.Fatal(err)
	}
	if p.CanSubscribe() {
		t.Error("v1 without newChapters should not be subscribable")
	}
}
