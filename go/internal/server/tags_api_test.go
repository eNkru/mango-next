package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/eNkru/mango-next/internal/config"
	"github.com/eNkru/mango-next/internal/storage"
)

func TestTagsIndexAndDetailAPI(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "mango.db")
	cfg := &config.Config{
		BaseURL:    "/",
		DBPath:     dbPath,
		Port:       9000,
		UploadPath: filepath.Join(dir, "uploads"),
	}
	cfg.SetCurrent()

	st, err := storage.Open(dbPath, filepath.Join(dir, "library"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })

	if err := st.NewUser("tagadmin", "password123", true); err != nil {
		t.Fatal(err)
	}
	token, err := st.VerifyUser("tagadmin", "password123")
	if err != nil {
		t.Fatal(err)
	}

	// Create one title and tag it.
	if _, err := st.DB().Exec(
		`INSERT INTO titles (id, path, signature, unavailable, hidden) VALUES ('t1', 'series/a', '1', 0, 0)`,
	); err != nil {
		t.Fatal(err)
	}
	if err := st.AddTag("t1", "action"); err != nil {
		t.Fatal(err)
	}

	s := NewServer(&Dependencies{Config: cfg, Storage: st})
	// Library can be nil for tags index counts (uses storage only). Detail needs TitleHash.
	// For detail 404 without library title, ensure API returns not found rather than panic.
	s.RegisterRoutes()

	cookie := &http.Cookie{Name: "mango-token-9000", Value: token}

	// Legacy tags list still string array.
	legacyReq := httptest.NewRequest(http.MethodGet, "/api/tags", nil)
	legacyReq.AddCookie(cookie)
	legacyRec := httptest.NewRecorder()
	s.Router.ServeHTTP(legacyRec, legacyReq)
	if legacyRec.Code != http.StatusOK {
		t.Fatalf("legacy tags status=%d body=%s", legacyRec.Code, legacyRec.Body.String())
	}
	var legacy struct {
		Success bool     `json:"success"`
		Data    []string `json:"data"`
	}
	if err := json.Unmarshal(legacyRec.Body.Bytes(), &legacy); err != nil {
		t.Fatal(err)
	}
	if !legacy.Success || len(legacy.Data) == 0 {
		t.Fatalf("legacy tags body=%+v", legacy)
	}

	// Index with counts.
	idxReq := httptest.NewRequest(http.MethodGet, "/api/tags/index", nil)
	idxReq.AddCookie(cookie)
	idxRec := httptest.NewRecorder()
	s.Router.ServeHTTP(idxRec, idxReq)
	if idxRec.Code != http.StatusOK {
		t.Fatalf("index status=%d body=%s", idxRec.Code, idxRec.Body.String())
	}
	var idx struct {
		Success bool `json:"success"`
		Tags    []struct {
			Tag   string `json:"tag"`
			Count int    `json:"count"`
		} `json:"tags"`
	}
	if err := json.Unmarshal(idxRec.Body.Bytes(), &idx); err != nil {
		t.Fatal(err)
	}
	if !idx.Success || len(idx.Tags) == 0 || idx.Tags[0].Count < 1 {
		t.Fatalf("index body=%+v", idx)
	}

	// Detail without library title should 404 (title not present in TitleHash).
	detailReq := httptest.NewRequest(http.MethodGet, "/api/tags/action/titles", nil)
	detailReq.AddCookie(cookie)
	detailRec := httptest.NewRecorder()
	s.Router.ServeHTTP(detailRec, detailReq)
	if detailRec.Code != http.StatusNotFound {
		t.Fatalf("detail without library status=%d body=%s", detailRec.Code, detailRec.Body.String())
	}
}
