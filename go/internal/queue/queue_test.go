package queue

import (
	"path/filepath"
	"testing"
	"time"
)

func newTestQueue(t *testing.T) *Queue {
	t.Helper()
	q, err := NewQueue(filepath.Join(t.TempDir(), "queue.db"))
	if err != nil {
		t.Fatal(err)
	}
	return q
}

func TestNewQueue(t *testing.T) {
	q := newTestQueue(t)
	defer q.Close()

	count, err := q.Count()
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Errorf("expected empty queue, got %d items", count)
	}
}

func TestPushAndCount(t *testing.T) {
	q := newTestQueue(t)
	defer q.Close()

	jobs := []Job{
		{ID: "job1", MangaID: "m1", Title: "Ch1", MangaTitle: "M1",
			Status: StatusPending, Time: time.Now()},
		{ID: "job2", MangaID: "m1", Title: "Ch2", MangaTitle: "M1",
			Status: StatusPending, Time: time.Now()},
	}

	inserted, err := q.Push(jobs)
	if err != nil {
		t.Fatal(err)
	}
	if inserted != 2 {
		t.Errorf("inserted = %d, want 2", inserted)
	}

	count, err := q.Count()
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}
}

func TestPushDuplicateIgnored(t *testing.T) {
	q := newTestQueue(t)
	defer q.Close()

	now := time.Now()
	jobs := []Job{
		{ID: "job1", MangaID: "m1", Title: "Ch1", MangaTitle: "M1",
			Status: StatusPending, Time: now},
	}
	inserted, err := q.Push(jobs)
	if err != nil {
		t.Fatal(err)
	}
	if inserted != 1 {
		t.Errorf("first insert = %d, want 1", inserted)
	}

	// Push same ID again
	inserted, err = q.Push(jobs)
	if err != nil {
		t.Fatal(err)
	}
	if inserted != 0 {
		t.Errorf("duplicate insert = %d, want 0", inserted)
	}
}

func TestSetStatus(t *testing.T) {
	q := newTestQueue(t)
	defer q.Close()

	_, err := q.Push([]Job{{
		ID: "job1", MangaID: "m1", Title: "Ch1", MangaTitle: "M1",
		Status: StatusPending, Time: time.Now(),
	}})
	if err != nil {
		t.Fatal(err)
	}

	if err := q.SetStatus(StatusDownloading, "job1"); err != nil {
		t.Fatal(err)
	}

	count, err := q.CountStatus(StatusDownloading)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Errorf("downloading count = %d, want 1", count)
	}
}

func TestExists(t *testing.T) {
	q := newTestQueue(t)
	defer q.Close()

	exists, err := q.Exists("nonexistent")
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Error("nonexistent job should not exist")
	}

	_, err = q.Push([]Job{{
		ID: "job1", MangaID: "m1", Title: "Ch1", MangaTitle: "M1",
		Status: StatusPending, Time: time.Now(),
	}})
	if err != nil {
		t.Fatal(err)
	}

	exists, err = q.Exists("job1")
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Error("job1 should exist")
	}
}

func TestDelete(t *testing.T) {
	q := newTestQueue(t)
	defer q.Close()

	_, err := q.Push([]Job{{
		ID: "job1", MangaID: "m1", Title: "Ch1", MangaTitle: "M1",
		Status: StatusPending, Time: time.Now(),
	}})
	if err != nil {
		t.Fatal(err)
	}

	if err := q.Delete("job1"); err != nil {
		t.Fatal(err)
	}

	count, err := q.Count()
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Errorf("after delete, count = %d, want 0", count)
	}
}

func TestPendingCount(t *testing.T) {
	q := newTestQueue(t)
	defer q.Close()

	now := time.Now()
	_, err := q.Push([]Job{
		{ID: "job1", MangaID: "m1", Title: "Ch1", MangaTitle: "M1",
			Status: StatusPending, Time: now},
		{ID: "job2", MangaID: "m1", Title: "Ch2", MangaTitle: "M1",
			Status: StatusDownloading, Time: now},
		{ID: "job3", MangaID: "m1", Title: "Ch3", MangaTitle: "M1",
			Status: StatusCompleted, Time: now},
	})
	if err != nil {
		t.Fatal(err)
	}

	pending, err := q.PendingCount()
	if err != nil {
		t.Fatal(err)
	}
	if pending != 2 {
		t.Errorf("pending count = %d, want 2", pending)
	}
}

func TestSetPagesAndIncrement(t *testing.T) {
	q := newTestQueue(t)
	defer q.Close()

	_, err := q.Push([]Job{{
		ID: "job1", MangaID: "m1", Title: "Ch1", MangaTitle: "M1",
		Status: StatusPending, Time: time.Now(),
	}})
	if err != nil {
		t.Fatal(err)
	}

	if err := q.SetPages(10, "job1"); err != nil {
		t.Fatal(err)
	}
	if err := q.AddSuccess("job1"); err != nil {
		t.Fatal(err)
	}
	if err := q.AddFail("job1"); err != nil {
		t.Fatal(err)
	}

	job, err := q.Get("job1")
	if err != nil {
		t.Fatal(err)
	}
	if job == nil {
		t.Fatal("job not found")
	}
	if job.Pages != 10 {
		t.Errorf("pages = %d, want 10", job.Pages)
	}
	if job.SuccessCount != 1 {
		t.Errorf("success = %d, want 1", job.SuccessCount)
	}
	if job.FailCount != 1 {
		t.Errorf("fail = %d, want 1", job.FailCount)
	}
}

func TestAddMessage(t *testing.T) {
	q := newTestQueue(t)
	defer q.Close()

	_, err := q.Push([]Job{{
		ID: "job1", MangaID: "m1", Title: "Ch1", MangaTitle: "M1",
		Status: StatusPending, Time: time.Now(),
	}})
	if err != nil {
		t.Fatal(err)
	}

	if err := q.AddMessage("error: timeout", "job1"); err != nil {
		t.Fatal(err)
	}

	job, err := q.Get("job1")
	if err != nil {
		t.Fatal(err)
	}
	if job.StatusMessage != "\nerror: timeout" {
		t.Errorf("message = %q, want %q", job.StatusMessage, "\nerror: timeout")
	}
}

func TestPopDownloadable(t *testing.T) {
	q := newTestQueue(t)
	defer q.Close()

	now := time.Now()
	// Non-plugin job (simple ID with no dash) should be skipped.
	_, err := q.Push([]Job{
		{ID: "simpleid123", MangaID: "m1", Title: "Ch1", MangaTitle: "M1",
			Status: StatusPending, Time: now},
		{ID: "plugin1-Y2gx", MangaID: "m1", Title: "Ch2", MangaTitle: "M1", // base64 of "ch1"
			Status: StatusPending, Time: now.Add(time.Second)},
	})
	if err != nil {
		t.Fatal(err)
	}

	job, err := q.PopDownloadable()
	if err != nil {
		t.Fatal(err)
	}
	if job == nil {
		t.Fatal("expected a downloadable job")
	}
	if job.ID != "plugin1-Y2gx" {
		t.Errorf("got job ID %q, want plugin1-Y2gx", job.ID)
	}
	if job.PluginID != "plugin1" {
		t.Errorf("plugin_id = %q, want plugin1", job.PluginID)
	}
	if job.PluginChapterID != "ch1" {
		t.Errorf("plugin_chapter_id = %q, want ch1 (decoded from Y2gx)", job.PluginChapterID)
	}
}

func TestList(t *testing.T) {
	q := newTestQueue(t)
	defer q.Close()

	now := time.Now()
	_, err := q.Push([]Job{
		{ID: "job1", MangaID: "m1", Title: "A", MangaTitle: "M1",
			Status: StatusPending, Time: now},
		{ID: "job2", MangaID: "m1", Title: "B", MangaTitle: "M1",
			Status: StatusCompleted, Time: now.Add(time.Second)},
	})
	if err != nil {
		t.Fatal(err)
	}

	jobs, err := q.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(jobs) != 2 {
		t.Fatalf("got %d jobs, want 2", len(jobs))
	}
}

func TestReset(t *testing.T) {
	q := newTestQueue(t)
	defer q.Close()

	_, err := q.Push([]Job{{
		ID: "job1", MangaID: "m1", Title: "Ch1", MangaTitle: "M1",
		Status: StatusError, Pages: 5, SuccessCount: 2, FailCount: 3,
		StatusMessage: "some error", Time: time.Now(),
	}})
	if err != nil {
		t.Fatal(err)
	}

	if err := q.Reset("job1"); err != nil {
		t.Fatal(err)
	}

	job, err := q.Get("job1")
	if err != nil {
		t.Fatal(err)
	}
	if job.Status != StatusPending {
		t.Errorf("status = %d, want Pending", job.Status)
	}
	if job.Pages != 0 || job.SuccessCount != 0 || job.FailCount != 0 {
		t.Error("counters should be zeroed after reset")
	}
}
