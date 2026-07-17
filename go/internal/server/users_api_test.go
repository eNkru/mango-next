package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/eNkru/mango-next/internal/config"
	"github.com/eNkru/mango-next/internal/storage"
)

func TestAdminUsersAPI(t *testing.T) {
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

	if err := st.NewUser("rootadmin", "password123", true); err != nil {
		t.Fatalf("create rootadmin: %v", err)
	}
	token, err := st.VerifyUser("rootadmin", "password123")
	if err != nil || token == "" {
		t.Fatalf("login rootadmin: %v", err)
	}

	s := NewServer(&Dependencies{Config: cfg, Storage: st})
	s.RegisterRoutes()

	cookie := &http.Cookie{Name: "mango-token-9000", Value: token}

	// List
	listReq := httptest.NewRequest(http.MethodGet, "/api/admin/users", nil)
	listReq.AddCookie(cookie)
	listRec := httptest.NewRecorder()
	s.Router.ServeHTTP(listRec, listReq)
	if listRec.Code != http.StatusOK {
		t.Fatalf("list status = %d body=%s", listRec.Code, listRec.Body.String())
	}
	var listBody struct {
		Success         bool `json:"success"`
		Users           []struct {
			Username string `json:"username"`
			Admin    bool   `json:"admin"`
		} `json:"users"`
		CurrentUsername string `json:"current_username"`
	}
	if err := json.Unmarshal(listRec.Body.Bytes(), &listBody); err != nil {
		t.Fatal(err)
	}
	if !listBody.Success || listBody.CurrentUsername != "rootadmin" {
		t.Fatalf("list body = %+v", listBody)
	}

	// Create
	createReq := httptest.NewRequest(
		http.MethodPost,
		"/api/admin/users",
		bytes.NewBufferString(`{"username":"alice","password":"password123","admin":false}`),
	)
	createReq.Header.Set("Content-Type", "application/json")
	createReq.AddCookie(cookie)
	createRec := httptest.NewRecorder()
	s.Router.ServeHTTP(createRec, createReq)
	if createRec.Code != http.StatusOK {
		t.Fatalf("create status = %d body=%s", createRec.Code, createRec.Body.String())
	}

	// Update rename + promote
	updateReq := httptest.NewRequest(
		http.MethodPut,
		"/api/admin/users/alice",
		bytes.NewBufferString(`{"username":"alice2","password":"","admin":true}`),
	)
	updateReq.Header.Set("Content-Type", "application/json")
	updateReq.AddCookie(cookie)
	updateRec := httptest.NewRecorder()
	s.Router.ServeHTTP(updateRec, updateReq)
	if updateRec.Code != http.StatusOK {
		t.Fatalf("update status = %d body=%s", updateRec.Code, updateRec.Body.String())
	}

	// Self-delete blocked
	selfReq := httptest.NewRequest(http.MethodDelete, "/api/admin/user/delete/rootadmin", nil)
	selfReq.AddCookie(cookie)
	selfRec := httptest.NewRecorder()
	s.Router.ServeHTTP(selfRec, selfReq)
	if selfRec.Code == http.StatusOK {
		t.Fatal("expected self-delete to fail")
	}

	// Delete other user
	delReq := httptest.NewRequest(http.MethodDelete, "/api/admin/user/delete/alice2", nil)
	delReq.AddCookie(cookie)
	delRec := httptest.NewRecorder()
	s.Router.ServeHTTP(delRec, delReq)
	if delRec.Code != http.StatusOK {
		t.Fatalf("delete status = %d body=%s", delRec.Code, delRec.Body.String())
	}
}
