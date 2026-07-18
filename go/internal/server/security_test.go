package server

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSafeRedirectPath(t *testing.T) {
	cases := []struct {
		in, want string
	}{
		{"", "/"},
		{"/library", "/library"},
		{"/book/x?y=1", "/book/x?y=1"},
		{"//evil.example", "/"},
		{"https://evil.example/", "/"},
		{"http://evil.example/path", "/"},
		{"/\\evil", "/"},
		{"library", "/"},
		{"./library", "/"},
	}
	for _, tc := range cases {
		if got := safeRedirectPath(tc.in); got != tc.want {
			t.Errorf("safeRedirectPath(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestRequestIsHTTPS(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	if requestIsHTTPS(req) {
		t.Error("plain HTTP should not be HTTPS")
	}
	req.Header.Set("X-Forwarded-Proto", "https")
	if !requestIsHTTPS(req) {
		t.Error("X-Forwarded-Proto=https should be HTTPS")
	}
	req2 := httptest.NewRequest(http.MethodGet, "https://example/", nil)
	req2.TLS = &tls.ConnectionState{}
	if !requestIsHTTPS(req2) {
		t.Error("r.TLS set should be HTTPS")
	}
}

func TestClientIPUsesRemoteAddr(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "203.0.113.9:54321"
	req.Header.Set("X-Forwarded-For", "1.2.3.4")
	if got := clientIP(req); got != "203.0.113.9" {
		t.Errorf("clientIP = %q, want 203.0.113.9", got)
	}
}

func TestLoginLimiter(t *testing.T) {
	l := newLoginLimiter(3, time.Minute)
	base := time.Unix(1_700_000_000, 0)
	now := base
	l.now = func() time.Time { return now }

	ip := "10.0.0.1"
	for i := 0; i < 3; i++ {
		if !l.allowed(ip) {
			t.Fatalf("attempt %d should be allowed", i+1)
		}
		l.recordFailure(ip)
	}
	if l.allowed(ip) {
		t.Fatal("4th attempt should be blocked")
	}
	now = base.Add(time.Minute + time.Second)
	if !l.allowed(ip) {
		t.Fatal("after window failures should expire")
	}
	l.recordFailure(ip)
	l.clear(ip)
	if !l.allowed(ip) {
		t.Fatal("clear should reset failures")
	}
}

func TestHandleLogoutRevokesToken(t *testing.T) {
	st, cfg, _ := setupTest(t)
	token, err := st.VerifyUser("testuser", "password123")
	if err != nil || token == "" {
		t.Fatal(err)
	}

	s := &Server{Deps: &Dependencies{Config: cfg, Storage: st}}
	req := httptest.NewRequest(http.MethodGet, "/logout", nil)
	req.AddCookie(&http.Cookie{Name: "mango-token-9000", Value: token})
	rec := httptest.NewRecorder()
	s.handleLogout(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusFound)
	}
	username, err := st.VerifyToken(token)
	if err != nil {
		t.Fatal(err)
	}
	if username != "" {
		t.Fatal("token should be invalid after logout")
	}
}

func TestHandleLoginRejectsOpenRedirect(t *testing.T) {
	st, cfg, _ := setupTest(t)
	s := &Server{Deps: &Dependencies{Config: cfg, Storage: st}}

	// Use a unique RemoteAddr so we don't trip the process-wide limiter.
	body := strings.NewReader("username=testuser&password=password123&callback=https://evil.example/")
	req := httptest.NewRequest(http.MethodPost, "/login", body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.RemoteAddr = "198.51.100.10:40001"
	rec := httptest.NewRecorder()
	s.handleLogin(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusFound)
	}
	if loc := rec.Header().Get("Location"); loc != "/" {
		t.Fatalf("Location = %q, want /", loc)
	}
}

func TestAPILoginGenericErrorAndRateLimit(t *testing.T) {
	st, cfg, _ := setupTest(t)
	s := &Server{Deps: &Dependencies{Config: cfg, Storage: st}}
	ip := "198.51.100.20:40002"

	post := func() *httptest.ResponseRecorder {
		payload, _ := json.Marshal(map[string]string{"username": "testuser", "password": "wrong"})
		req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		req.RemoteAddr = ip
		rec := httptest.NewRecorder()
		s.apiLogin(rec, req)
		return rec
	}

	for i := 0; i < loginRateLimit; i++ {
		rec := post()
		if rec.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want 403", rec.Code)
		}
		var resp map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatal(err)
		}
		if resp["error"] != "login failed" {
			t.Fatalf("error = %v, want login failed", resp["error"])
		}
	}
	rec := post()
	if rec.Code != http.StatusForbidden {
		t.Fatalf("over-limit status = %d, want 403", rec.Code)
	}
	var resp map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp["error"] != "login failed" {
		t.Fatalf("over-limit error = %v", resp["error"])
	}
}

func TestRequireAuthPageRedirectIncludesCallback(t *testing.T) {
	_, cfg, _ := setupTest(t)
	cfg.SetCurrent()

	req := httptest.NewRequest(http.MethodGet, "/library?sort=name", nil)
	rec := httptest.NewRecorder()
	requireAuth(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusFound)
	}
	loc := rec.Header().Get("Location")
	if !strings.HasPrefix(loc, "/login?callback=") {
		t.Fatalf("Location = %q, want /login?callback=...", loc)
	}
	if !strings.Contains(loc, "callback=%2Flibrary") && !strings.Contains(loc, "callback=/library") {
		// QueryEscape encodes / as %2F
		if !strings.Contains(loc, "%2Flibrary") {
			t.Fatalf("Location missing library callback: %q", loc)
		}
	}
}

func TestRequireAuthAPIStillUnauthorized(t *testing.T) {
	_, cfg, _ := setupTest(t)
	cfg.SetCurrent()

	req := httptest.NewRequest(http.MethodGet, "/api/library", nil)
	rec := httptest.NewRecorder()
	requireAuth(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestHandleLoginPageRedirectsWhenAuthenticated(t *testing.T) {
	st, cfg, _ := setupTest(t)
	token, err := st.VerifyUser("testuser", "password123")
	if err != nil || token == "" {
		t.Fatal(err)
	}
	s := &Server{Deps: &Dependencies{Config: cfg, Storage: st}}

	req := httptest.NewRequest(http.MethodGet, "/login?callback=/library", nil)
	req.AddCookie(&http.Cookie{Name: "mango-token-9000", Value: token})
	rec := httptest.NewRecorder()
	s.handleLoginPage(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusFound)
	}
	if loc := rec.Header().Get("Location"); loc != "/library" {
		t.Fatalf("Location = %q, want /library", loc)
	}
}

func TestHandleLoginPageRejectsOpenRedirectWhenAuthenticated(t *testing.T) {
	st, cfg, _ := setupTest(t)
	token, err := st.VerifyUser("testuser", "password123")
	if err != nil || token == "" {
		t.Fatal(err)
	}
	s := &Server{Deps: &Dependencies{Config: cfg, Storage: st}}

	req := httptest.NewRequest(http.MethodGet, "/login?callback=https://evil.example/", nil)
	req.AddCookie(&http.Cookie{Name: "mango-token-9000", Value: token})
	rec := httptest.NewRecorder()
	s.handleLoginPage(rec, req)

	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusFound)
	}
	if loc := rec.Header().Get("Location"); loc != "/" {
		t.Fatalf("Location = %q, want /", loc)
	}
}

func TestCORSAndSecurityHeaders(t *testing.T) {
	h := SecurityHeadersMiddleware(CORSMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("Access-Control-Allow-Origin = %q, want empty", got)
	}
	if rec.Header().Get("X-Content-Type-Options") != "nosniff" {
		t.Error("missing nosniff")
	}
	if rec.Header().Get("X-Frame-Options") != "SAMEORIGIN" {
		t.Error("missing SAMEORIGIN")
	}
	if rec.Header().Get("Referrer-Policy") != "strict-origin-when-cross-origin" {
		t.Error("missing Referrer-Policy")
	}

	opt := httptest.NewRequest(http.MethodOptions, "/", nil)
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, opt)
	if rec2.Code != http.StatusNoContent {
		t.Errorf("OPTIONS status = %d, want 204", rec2.Code)
	}
}

func TestUploadHandlerPathContainment(t *testing.T) {
	root := t.TempDir()
	sibling := root + "-evil"
	if err := os.MkdirAll(sibling, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sibling, "secret.txt"), []byte("nope"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "ok.txt"), []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}

	h := UploadHandler(root)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))

	// Valid file.
	req := httptest.NewRequest(http.MethodGet, "/uploads/ok.txt", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("valid upload status = %d", rec.Code)
	}
	body, _ := io.ReadAll(rec.Body)
	if string(body) != "ok" {
		t.Fatalf("body = %q", body)
	}

	// Sibling prefix must not escape via string-prefix bugs.
	req2 := httptest.NewRequest(http.MethodGet, "/uploads/../"+filepath.Base(sibling)+"/secret.txt", nil)
	// Build a path that resolves under sibling when joined poorly.
	// UploadHandler joins uploadPath with TrimPrefix of /uploads/, so try traversal.
	req2 = httptest.NewRequest(http.MethodGet, "/uploads/../"+filepath.Base(sibling)+"/secret.txt", nil)
	rec2 := httptest.NewRecorder()
	h.ServeHTTP(rec2, req2)
	if rec2.Code == http.StatusOK {
		t.Fatal("sibling traversal should not be served")
	}
}

func TestUnderUploadRoot(t *testing.T) {
	root := "/tmp/uploads"
	if !underUploadRoot("/tmp/uploads/a.png", root) {
		t.Error("child should be under root")
	}
	if underUploadRoot("/tmp/uploads-evil/a.png", root) {
		t.Error("sibling prefix must not pass")
	}
	if underUploadRoot("/tmp/other/a.png", root) {
		t.Error("outside path must not pass")
	}
}
