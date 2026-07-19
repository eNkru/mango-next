package server

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestHandleAdminReactShell(t *testing.T) {
	s, cookie, _ := setupBrowseServer(t)
	rec := browseRequest(t, s, cookie, http.MethodGet, "/admin/", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	html := rec.Body.String()
	if !strings.Contains(html, "pageId") || !strings.Contains(html, "admin") {
		t.Fatalf("expected admin pageId in shell, head=%q", html[:min(400, len(html))])
	}
	if strings.Contains(html, "admin.js") {
		t.Fatal("legacy admin.js should not load on react shell")
	}
}

func TestApiAdminScanProgressIdle(t *testing.T) {
	s, cookie, _ := setupBrowseServer(t)
	rec := browseRequest(t, s, cookie, http.MethodGet, "/api/admin/scan_progress", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	var body struct {
		Success bool   `json:"success"`
		Running bool   `json:"running"`
		Error   string `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if !body.Success || body.Running {
		t.Fatalf("idle progress = %+v", body)
	}
}

func TestApiAdminScanStartAndProgress(t *testing.T) {
	s, cookie, _ := setupBrowseServer(t)

	rec := browseRequest(t, s, cookie, http.MethodPost, "/api/admin/scan", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("scan status = %d body=%s", rec.Code, rec.Body.String())
	}
	var start struct {
		Success bool `json:"success"`
		Running bool `json:"running"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &start); err != nil {
		t.Fatal(err)
	}
	if !start.Success {
		t.Fatalf("start = %+v", start)
	}

	// Wait for background scan (library is tiny in test fixture).
	deadline := time.Now().Add(5 * time.Second)
	var last map[string]any
	for time.Now().Before(deadline) {
		rec = browseRequest(t, s, cookie, http.MethodGet, "/api/admin/scan_progress", nil)
		if rec.Code != http.StatusOK {
			t.Fatalf("progress status = %d", rec.Code)
		}
		last = map[string]any{}
		if err := json.Unmarshal(rec.Body.Bytes(), &last); err != nil {
			t.Fatal(err)
		}
		if running, _ := last["running"].(bool); !running {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if running, _ := last["running"].(bool); running {
		t.Fatalf("scan still running: %+v", last)
	}
	if success, _ := last["success"].(bool); !success {
		t.Fatalf("progress success=false: %+v", last)
	}
	titles, _ := last["titles"].(float64)
	if titles < 1 {
		t.Fatalf("expected titles >= 1, got %+v", last)
	}

	// Duplicate start while idle should start again (or return quickly).
	rec = browseRequest(t, s, cookie, http.MethodPost, "/api/admin/scan", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("second scan status = %d", rec.Code)
	}
}
