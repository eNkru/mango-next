package server

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/eNkru/mango-next/internal/config"
	"github.com/eNkru/mango-next/internal/storage"
)

// setupTest creates a temporary config and storage for testing.
func setupTest(t *testing.T) (*storage.Storage, *config.Config, string) {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "mango.db")

	// Create a test config.
	cfg := &config.Config{
		BaseURL: "/",
		DBPath:  dbPath,
		Port:    9000,
	}
	cfg.SetCurrent()

	// Open storage (will auto-create admin user).
	st, err := storage.Open(dbPath, filepath.Join(dir, "library"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })

	// Create a non-admin test user.
	if err := st.NewUser("testuser", "password123", false); err != nil {
		t.Fatal(err)
	}

	return st, cfg, dir
}

func TestAuthMiddlewareValidToken(t *testing.T) {
	st, _, _ := setupTest(t)

	// Login to get a token.
	token, err := st.VerifyUser("testuser", "password123")
	if err != nil || token == "" {
		t.Fatal("could not get token")
	}

	handler := AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username := GetUsername(r)
		if username != "testuser" {
			t.Errorf("username = %q, want testuser", username)
		}
		if GetIsAdmin(r) {
			t.Error("testuser should not be admin")
		}
		w.WriteHeader(http.StatusOK)
	}), st)

	// Simulate a request with cookie.
	req := httptest.NewRequest(http.MethodGet, "/api/titles", nil)
	req.AddCookie(&http.Cookie{Name: "mango-token-9000", Value: token})

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestAuthMiddlewareInvalidToken(t *testing.T) {
	st, _, _ := setupTest(t)

	handler := AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be reached for invalid token")
	}), st)

	req := httptest.NewRequest(http.MethodGet, "/api/titles", nil)
	req.AddCookie(&http.Cookie{Name: "mango-token-9000", Value: "invalid-token"})

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestAuthMiddlewareBearerToken(t *testing.T) {
	st, _, _ := setupTest(t)

	token, err := st.VerifyUser("testuser", "password123")
	if err != nil || token == "" {
		t.Fatal("could not get token")
	}

	handler := AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username := GetUsername(r)
		if username != "testuser" {
			t.Errorf("username = %q, want testuser", username)
		}
		w.WriteHeader(http.StatusOK)
	}), st)

	req := httptest.NewRequest(http.MethodGet, "/api/titles", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestAuthMiddlewareAdminToken(t *testing.T) {
	st, _, _ := setupTest(t)

	// Admin user exists (auto-created).
	// We need to get the admin's token. Since the password is random,
	// we'll use a workaround: get the admin user directly by verifying
	// the password is not possible. Instead, let's create a known admin.
	if err := st.NewUser("admin2", "adminpass123", true); err != nil {
		t.Fatal(err)
	}

	token, err := st.VerifyUser("admin2", "adminpass123")
	if err != nil || token == "" {
		t.Fatal("could not get token")
	}

	// Test admin middleware.
	baseHandler := AuthMiddleware(AdminMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username := GetUsername(r)
		if username != "admin2" {
			t.Errorf("username = %q, want admin2", username)
		}
		if !GetIsAdmin(r) {
			t.Error("admin2 should be admin")
		}
		w.WriteHeader(http.StatusOK)
	})), st)

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.AddCookie(&http.Cookie{Name: "mango-token-9000", Value: token})

	rec := httptest.NewRecorder()
	baseHandler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestAuthMiddlewareAdminRejected(t *testing.T) {
	st, _, _ := setupTest(t)

	token, err := st.VerifyUser("testuser", "password123")
	if err != nil || token == "" {
		t.Fatal("could not get token")
	}

	handler := AuthMiddleware(AdminMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("non-admin should not reach admin handler")
	})), st)

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req.AddCookie(&http.Cookie{Name: "mango-token-9000", Value: token})

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

func TestAuthMiddlewareOptions(t *testing.T) {
	st, _, _ := setupTest(t)

	called := false
	handler := AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}), st)

	req := httptest.NewRequest(http.MethodOptions, "/api/titles", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !called {
		t.Error("OPTIONS handler should be called")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestAuthMiddlewareDisabledLogin(t *testing.T) {
	st, cfg, _ := setupTest(t)

	// Enable disable_login with default user.
	cfg.DisableLogin = true
	cfg.DefaultUsername = "testuser"

	handler := AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username := GetUsername(r)
		if username != "testuser" {
			t.Errorf("username = %q, want testuser", username)
		}
		w.WriteHeader(http.StatusOK)
	}), st)

	// No auth cookies/headers at all.
	req := httptest.NewRequest(http.MethodGet, "/api/titles", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	// Reset.
	cfg.DisableLogin = false
	cfg.DefaultUsername = ""
}

func TestAuthMiddlewareDisabledLoginMissingDefaultUser(t *testing.T) {
	st, cfg, _ := setupTest(t)

	cfg.DisableLogin = true
	cfg.DefaultUsername = "nonexistent"

	handler := AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be reached")
	}), st)

	req := httptest.NewRequest(http.MethodGet, "/api/titles", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}

	cfg.DisableLogin = false
	cfg.DefaultUsername = ""
}

func TestAuthMiddlewareProxyHeader(t *testing.T) {
	st, cfg, _ := setupTest(t)

	cfg.AuthProxyHeaderName = "X-Auth-User"

	handler := AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username := GetUsername(r)
		if username != "testuser" {
			t.Errorf("username = %q, want testuser", username)
		}
		w.WriteHeader(http.StatusOK)
	}), st)

	req := httptest.NewRequest(http.MethodGet, "/api/titles", nil)
	req.Header.Set("X-Auth-User", "testuser")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	cfg.AuthProxyHeaderName = ""
}

func TestAuthMiddlewareProxyHeaderInvalidUser(t *testing.T) {
	st, cfg, _ := setupTest(t)

	cfg.AuthProxyHeaderName = "X-Auth-User"

	handler := AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be reached")
	}), st)

	req := httptest.NewRequest(http.MethodGet, "/api/titles", nil)
	req.Header.Set("X-Auth-User", "nonexistent")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}

	cfg.AuthProxyHeaderName = ""
}

func TestSetAndClearAuthTokenCookie(t *testing.T) {
	cfg := &config.Config{BaseURL: "/", Port: 9090}

	rec := httptest.NewRecorder()
	SetAuthTokenCookie(rec, cfg, "test-token-123")

	cookies := rec.Result().Cookies()
	var found *http.Cookie
	for _, c := range cookies {
		if strings.Contains(c.Name, "mango-token") {
			found = c
			break
		}
	}
	if found == nil {
		t.Fatal("no auth token cookie set")
	}
	if found.Value != "test-token-123" {
		t.Errorf("cookie value = %q, want test-token-123", found.Value)
	}
	if found.HttpOnly != true {
		t.Error("cookie should be HttpOnly")
	}

	// Clear.
	rec2 := httptest.NewRecorder()
	ClearAuthTokenCookie(rec2, cfg)
	cookies2 := rec2.Result().Cookies()
	for _, c := range cookies2 {
		if strings.Contains(c.Name, "mango-token") {
			if c.MaxAge != -1 {
				t.Errorf("clear cookie maxAge = %d, want -1", c.MaxAge)
			}
			return
		}
	}
	t.Error("no clear cookie found")
}

func TestExtractTokenFromCookie(t *testing.T) {
	cfg := &config.Config{Port: 9000}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "mango-token-9000", Value: "cookie-token"})

	token := extractToken(req, cfg)
	if token != "cookie-token" {
		t.Errorf("extractToken = %q, want cookie-token", token)
	}
}

func TestExtractTokenFromBearer(t *testing.T) {
	cfg := &config.Config{Port: 9000}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer bearer-token")

	token := extractToken(req, cfg)
	if token != "bearer-token" {
		t.Errorf("extractToken = %q, want bearer-token", token)
	}
}

func TestExtractTokenCookiePrecedence(t *testing.T) {
	cfg := &config.Config{Port: 9000}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "mango-token-9000", Value: "cookie-value"})
	req.Header.Set("Authorization", "Bearer bearer-value")

	// Cookie should take precedence.
	token := extractToken(req, cfg)
	if token != "cookie-value" {
		t.Errorf("extractToken = %q, want cookie-value (cookie should take precedence)", token)
	}
}
