package server

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestApiReaderBootstrapSuccess(t *testing.T) {
	s, cookie, title := setupBrowseServer(t)
	if len(title.Entries) == 0 {
		t.Fatal("expected at least one entry")
	}
	eid := title.Entries[0].ID()

	rec := browseRequest(t, s, cookie, http.MethodGet, "/api/reader/"+title.ID+"/"+eid, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}

	var body struct {
		Success bool `json:"success"`
		Data    struct {
			Title struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"title"`
			Entry struct {
				ID       string `json:"id"`
				Name     string `json:"name"`
				Pages    int    `json:"pages"`
				Progress int    `json:"progress"`
			} `json:"entry"`
			Dimensions []struct {
				Width  int `json:"width"`
				Height int `json:"height"`
			} `json:"dimensions"`
			Entries []struct {
				ID string `json:"id"`
			} `json:"entries"`
			ExitURL           string `json:"exit_url"`
			NextEntryURL      string `json:"next_entry_url"`
			PreviousEntryURL  string `json:"previous_entry_url"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if !body.Success {
		t.Fatalf("success=false body=%s", rec.Body.String())
	}
	if body.Data.Title.ID != title.ID {
		t.Fatalf("title id = %q want %q", body.Data.Title.ID, title.ID)
	}
	if body.Data.Entry.ID != eid {
		t.Fatalf("entry id = %q want %q", body.Data.Entry.ID, eid)
	}
	if body.Data.Entry.Pages <= 0 {
		t.Fatalf("pages = %d", body.Data.Entry.Pages)
	}
	if len(body.Data.Dimensions) != body.Data.Entry.Pages {
		t.Fatalf("dimensions=%d pages=%d", len(body.Data.Dimensions), body.Data.Entry.Pages)
	}
	if !strings.Contains(body.Data.ExitURL, "/book/"+title.ID) {
		t.Fatalf("exit_url = %q", body.Data.ExitURL)
	}
	if len(body.Data.Entries) == 0 {
		t.Fatal("expected sibling entries")
	}
}

func TestApiReaderBootstrapNotFound(t *testing.T) {
	s, cookie, _ := setupBrowseServer(t)
	rec := browseRequest(t, s, cookie, http.MethodGet, "/api/reader/missing-title/missing-entry", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	var body struct {
		Success bool   `json:"success"`
		Error   string `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Success {
		t.Fatal("expected success=false")
	}
	if body.Error == "" {
		t.Fatal("expected error message")
	}
}

func TestHandleReaderReactShell(t *testing.T) {
	s, cookie, title := setupBrowseServer(t)
	if len(title.Entries) == 0 {
		t.Fatal("expected at least one entry")
	}
	eid := title.Entries[0].ID()
	path := "/reader/" + title.ID + "/" + eid + "/2"
	rec := browseRequest(t, s, cookie, http.MethodGet, path, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	html := rec.Body.String()
	if !strings.Contains(html, `"pageId":"reader"`) && !strings.Contains(html, `"pageId": "reader"`) {
		// boot JSON is embedded; tolerate either spacing from encoding/json
		if !strings.Contains(html, "reader") || !strings.Contains(html, "pageId") {
			t.Fatalf("expected reader pageId boot in shell, body head=%q", html[:min(400, len(html))])
		}
	}
	if !strings.Contains(html, title.ID) || !strings.Contains(html, eid) {
		t.Fatalf("expected tid/eid in boot, body head=%q", html[:min(400, len(html))])
	}
	if strings.Contains(html, "reader.js") {
		t.Fatal("legacy reader.js should not be on react shell")
	}
}

func TestHandleReaderNoPageRedirect(t *testing.T) {
	s, cookie, title := setupBrowseServer(t)
	if len(title.Entries) == 0 {
		t.Fatal("expected at least one entry")
	}
	eid := title.Entries[0].ID()
	rec := browseRequest(t, s, cookie, http.MethodGet, "/reader/"+title.ID+"/"+eid, nil)
	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d", rec.Code)
	}
	loc := rec.Header().Get("Location")
	wantSuffix := "/reader/" + title.ID + "/" + eid + "/1"
	if !strings.HasSuffix(loc, wantSuffix) {
		t.Fatalf("Location = %q want suffix %q", loc, wantSuffix)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
