package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/eNkru/mango-next/internal/library"
	"github.com/eNkru/mango-next/internal/storage"
)

type progressBlockingEntry struct {
	started chan struct{}
	release chan struct{}
	once    sync.Once
}

func (e *progressBlockingEntry) ReadPage(int) (*storage.Image, error) {
	e.once.Do(func() { close(e.started) })
	<-e.release
	return nil, errors.New("test read failure")
}

func (e *progressBlockingEntry) PageCount() int       { return 1 }
func (e *progressBlockingEntry) ID() string           { return "progress-entry" }
func (e *progressBlockingEntry) Name() string         { return "progress-entry" }
func (e *progressBlockingEntry) Path() string         { return "progress-entry" }
func (e *progressBlockingEntry) Mtime() time.Time     { return time.Time{} }
func (e *progressBlockingEntry) Err() error           { return nil }
func (e *progressBlockingEntry) Signature() uint64    { return 1 }
func (e *progressBlockingEntry) Book() *library.Title { return nil }

func TestThumbnailProgressReportsRunningAtZero(t *testing.T) {
	st, cfg, dir := setupTest(t)
	lib := library.NewLibrary(filepath.Join(dir, "library"), st, "")
	entry := &progressBlockingEntry{
		started: make(chan struct{}),
		release: make(chan struct{}),
	}
	lib.TitleIDs = []string{"title"}
	lib.TitleHash = map[string]*library.Title{
		"title": {ID: "title", Entries: []library.Entry{entry}},
	}

	generationDone := make(chan error, 1)
	go func() { generationDone <- lib.GenerateThumbnails() }()
	select {
	case <-entry.started:
	case <-time.After(time.Second):
		t.Fatal("thumbnail generation did not start")
	}

	s := NewServer(&Dependencies{Config: cfg, Storage: st, Library: lib})
	req := httptest.NewRequest(http.MethodGet, "/api/admin/thumbnail_progress", nil)
	rec := httptest.NewRecorder()
	s.apiAdminThumbnailProgress(rec, req)

	var body struct {
		Success  bool    `json:"success"`
		Progress float64 `json:"progress"`
		Running  bool    `json:"running"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if !body.Success || !body.Running || body.Progress != 0 {
		t.Fatalf("progress response = %+v, want success/running at 0", body)
	}

	close(entry.release)
	if err := <-generationDone; err != nil {
		t.Fatal(err)
	}
}
