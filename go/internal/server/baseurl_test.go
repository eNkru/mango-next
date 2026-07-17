package server

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/eNkru/mango-next/internal/config"
	"github.com/eNkru/mango-next/internal/storage"
)

func TestBaseMountPath(t *testing.T) {
	if got := baseMountPath("/"); got != "" {
		t.Errorf("root mount = %q, want empty", got)
	}
	if got := baseMountPath("/mango/"); got != "/mango" {
		t.Errorf("mount = %q, want /mango", got)
	}
}

func TestRegisterRoutesUnderBaseURL(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "mango.db")
	cfg := &config.Config{
		BaseURL:    "/mango/",
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

	s := NewServer(&Dependencies{Config: cfg, Storage: st})
	s.RegisterRoutes()

	// Authenticated API under base path is mounted (401 without token, not 404).
	req := httptest.NewRequest(http.MethodGet, "/mango/api/library", nil)
	rec := httptest.NewRecorder()
	s.Router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("GET /mango/api/library status = %d, want 401", rec.Code)
	}

	// Unprefixed app paths are not mounted.
	req2 := httptest.NewRequest(http.MethodGet, "/api/library", nil)
	rec2 := httptest.NewRecorder()
	s.Router.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusNotFound {
		t.Fatalf("GET /api/library status = %d, want 404", rec2.Code)
	}
}

func TestAppPath(t *testing.T) {
	s := &Server{Deps: &Dependencies{Config: &config.Config{BaseURL: "/mango/"}}}
	if got := s.appPath("login"); got != "/mango/login" {
		t.Errorf("appPath(login) = %q", got)
	}
	if got := s.appPath(""); got != "/mango/" {
		t.Errorf("appPath() = %q", got)
	}
	s2 := &Server{Deps: &Dependencies{Config: &config.Config{BaseURL: "/"}}}
	if got := s2.appPath("login"); got != "/login" {
		t.Errorf("root appPath = %q", got)
	}
}
