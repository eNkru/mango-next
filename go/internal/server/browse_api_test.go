package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/eNkru/mango-next/internal/config"
	"github.com/eNkru/mango-next/internal/library"
	"github.com/eNkru/mango-next/internal/storage"
	"github.com/eNkru/mango-next/web"
)

func setupBrowseServer(t *testing.T) (*Server, *http.Cookie, *library.Title) {
	t.Helper()
	dir := t.TempDir()
	libraryPath := filepath.Join(dir, "library")
	titlePath := filepath.Join(libraryPath, "Series")
	if err := os.MkdirAll(titlePath, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := writeTestCBZ(filepath.Join(titlePath, "Chapter 01.cbz"), 3); err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{
		BaseURL: "/", DBPath: filepath.Join(dir, "mango.db"), Port: 9000,
		LibraryPath: libraryPath, UploadPath: filepath.Join(dir, "uploads"),
	}
	cfg.SetCurrent()
	st, err := storage.Open(cfg.DBPath, libraryPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	if err := st.NewUser("browseadmin", "password123", true); err != nil {
		t.Fatal(err)
	}
	token, err := st.VerifyUser("browseadmin", "password123")
	if err != nil {
		t.Fatal(err)
	}
	lib := library.NewLibrary(libraryPath, st, "")
	if _, err := lib.Scan(); err != nil {
		t.Fatal(err)
	}
	lib.RLock()
	title := lib.TitleHash[lib.TitleIDs[0]]
	lib.RUnlock()
	templates, err := NewTemplateManager(web.Views())
	if err != nil {
		t.Fatal(err)
	}
	s := NewServer(&Dependencies{Config: cfg, Storage: st, Library: lib, Templates: templates})
	s.RegisterRoutes()
	return s, &http.Cookie{Name: "mango-token-9000", Value: token}, title
}

func browseRequest(t *testing.T, s *Server, cookie *http.Cookie, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var data []byte
	if body != nil {
		var err error
		data, err = json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(data))
	req.AddCookie(cookie)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	s.Router.ServeHTTP(rec, req)
	return rec
}

func TestBrowseRoutesAndMutationContracts(t *testing.T) {
	s, cookie, title := setupBrowseServer(t)
	if title == nil || len(title.Entries) != 1 {
		t.Fatal("scanned title or entry missing")
	}
	eid := title.Entries[0].ID()

	for path, pageID := range map[string]string{"/": "home", "/library": "library", "/book/" + title.ID: "title-detail"} {
		rec := browseRequest(t, s, cookie, http.MethodGet, path, nil)
		if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"pageId":"`+pageID+`"`) {
			t.Fatalf("GET %s status=%d body=%s", path, rec.Code, rec.Body.String())
		}
		if strings.Contains(rec.Body.String(), "jquery") || strings.Contains(rec.Body.String(), "uikit") {
			t.Fatalf("GET %s loaded legacy scripts", path)
		}
	}

	display := browseRequest(t, s, cookie, http.MethodPut, "/api/admin/display_name/"+title.ID, map[string]any{"name": "Display Series"})
	if display.Code != http.StatusOK {
		t.Fatalf("display status=%d body=%s", display.Code, display.Body.String())
	}
	entryDisplay := browseRequest(t, s, cookie, http.MethodPut, "/api/admin/display_name/"+title.ID, map[string]any{"name": "Display Chapter", "eid": eid})
	if entryDisplay.Code != http.StatusOK {
		t.Fatalf("entry display status=%d body=%s", entryDisplay.Code, entryDisplay.Body.String())
	}
	sortName := browseRequest(t, s, cookie, http.MethodPut, "/api/admin/sort_title/"+title.ID, map[string]any{"sort_name": "001", "eid": eid})
	if sortName.Code != http.StatusOK {
		t.Fatalf("entry sort status=%d body=%s", sortName.Code, sortName.Body.String())
	}

	book := browseRequest(t, s, cookie, http.MethodGet, "/api/book/"+title.ID, nil)
	if book.Code != http.StatusOK {
		t.Fatalf("book status=%d body=%s", book.Code, book.Body.String())
	}
	var payload struct {
		Title   browseTitle   `json:"title"`
		Entries []browseEntry `json:"entries"`
	}
	if err := json.Unmarshal(book.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Title.Name != "Display Series" || len(payload.Entries) != 1 || payload.Entries[0].Name != "Display Chapter" || payload.Entries[0].SortName != "001" {
		t.Fatalf("book payload=%+v entries=%+v", payload.Title, payload.Entries)
	}

	if rec := browseRequest(t, s, cookie, http.MethodPut, "/api/admin/display_name/"+title.ID, map[string]any{"name": " "}); rec.Code != http.StatusBadRequest {
		t.Fatalf("empty display status=%d body=%s", rec.Code, rec.Body.String())
	}
}

func TestBrowseLibraryHiddenFiltering(t *testing.T) {
	s, cookie, title := setupBrowseServer(t)
	if err := s.Deps.Storage.SetTitleHidden(title.ID, 1); err != nil {
		t.Fatal(err)
	}
	visible := browseRequest(t, s, cookie, http.MethodGet, "/api/library", nil)
	shown := browseRequest(t, s, cookie, http.MethodGet, "/api/library?show_hidden=1", nil)
	var first, second struct {
		Titles []browseTitle `json:"titles"`
	}
	if err := json.Unmarshal(visible.Body.Bytes(), &first); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(shown.Body.Bytes(), &second); err != nil {
		t.Fatal(err)
	}
	if len(first.Titles) != 0 || len(second.Titles) != 1 || !second.Titles[0].Hidden {
		t.Fatalf("hidden filtering default=%+v shown=%+v", first.Titles, second.Titles)
	}
}

func TestContinueReadingFiltersFinishedEntries(t *testing.T) {
	dir := t.TempDir()
	libraryPath := filepath.Join(dir, "library")
	titlePath := filepath.Join(libraryPath, "Series")
	if err := os.MkdirAll(titlePath, 0o755); err != nil {
		t.Fatal(err)
	}
	// Create three entries: one in-progress, one fully read (page == pages), one bulk-marked (page == -1).
	if err := writeTestCBZ(filepath.Join(titlePath, "Chapter 01.cbz"), 10); err != nil {
		t.Fatal(err)
	}
	if err := writeTestCBZ(filepath.Join(titlePath, "Chapter 02.cbz"), 10); err != nil {
		t.Fatal(err)
	}
	if err := writeTestCBZ(filepath.Join(titlePath, "Chapter 03.cbz"), 10); err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{
		BaseURL: "/", DBPath: filepath.Join(dir, "mango.db"), Port: 9000,
		LibraryPath: libraryPath, UploadPath: filepath.Join(dir, "uploads"),
	}
	cfg.SetCurrent()
	st, err := storage.Open(cfg.DBPath, libraryPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	if err := st.NewUser("testuser", "password123", true); err != nil {
		t.Fatal(err)
	}
	token, err := st.VerifyUser("testuser", "password123")
	if err != nil {
		t.Fatal(err)
	}
	lib := library.NewLibrary(libraryPath, st, "")
	if _, err := lib.Scan(); err != nil {
		t.Fatal(err)
	}
	lib.RLock()
	title := lib.TitleHash[lib.TitleIDs[0]]
	lib.RUnlock()
	if title == nil || len(title.Entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(title.Entries))
	}

	// Sort entries by name so we have deterministic order.
	entries := make([]library.Entry, len(title.Entries))
	copy(entries, title.Entries)
	// Use the entries as scanned; they are ordered by filename.
	eid0 := entries[0].ID() // Chapter 01
	eid1 := entries[1].ID() // Chapter 02
	eid2 := entries[2].ID() // Chapter 03

	// Chapter 01: in-progress (page 5 of 10).
	if err := st.SaveProgress("testuser", title.ID, strPtr(eid0), 5); err != nil {
		t.Fatal(err)
	}
	// Chapter 02: fully read (page == pageCount, 10 of 10).
	if err := st.SaveProgress("testuser", title.ID, strPtr(eid1), 10); err != nil {
		t.Fatal(err)
	}
	// Chapter 03: bulk-marked read (page == -1).
	if err := st.BulkMarkRead("testuser", title.ID, []string{eid2}); err != nil {
		t.Fatal(err)
	}

	templates, err := NewTemplateManager(web.Views())
	if err != nil {
		t.Fatal(err)
	}
	s := NewServer(&Dependencies{Config: cfg, Storage: st, Library: lib, Templates: templates})
	s.RegisterRoutes()
	cookie := &http.Cookie{Name: "mango-token-9000", Value: token}

	// Verify apiHome filters finished entries.
	homeRec := browseRequest(t, s, cookie, http.MethodGet, "/api/home", nil)
	if homeRec.Code != http.StatusOK {
		t.Fatalf("home status=%d body=%s", homeRec.Code, homeRec.Body.String())
	}
	var homePayload struct {
		ContinueReading []browseEntry `json:"continue_reading"`
	}
	if err := json.Unmarshal(homeRec.Body.Bytes(), &homePayload); err != nil {
		t.Fatal(err)
	}
	if len(homePayload.ContinueReading) != 1 {
		t.Fatalf("expected 1 continue_reading entry (in-progress only), got %d: %+v", len(homePayload.ContinueReading), homePayload.ContinueReading)
	}
	if homePayload.ContinueReading[0].ID != eid0 {
		t.Fatalf("expected in-progress entry %s, got %s", eid0, homePayload.ContinueReading[0].ID)
	}
	if homePayload.ContinueReading[0].Page != 5 {
		t.Fatalf("expected page 5, got %d", homePayload.ContinueReading[0].Page)
	}

	// Verify standalone endpoint still returns all entries (including finished).
	libRec := browseRequest(t, s, cookie, http.MethodGet, "/api/library/continue_reading", nil)
	if libRec.Code != http.StatusOK {
		t.Fatalf("library continue status=%d body=%s", libRec.Code, libRec.Body.String())
	}
	var libPayload struct {
		Data []struct {
			EntryID string `json:"entry_id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(libRec.Body.Bytes(), &libPayload); err != nil {
		t.Fatal(err)
	}
	if len(libPayload.Data) != 3 {
		t.Fatalf("standalone endpoint should return all 3 entries, got %d", len(libPayload.Data))
	}
}
