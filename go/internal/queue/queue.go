// Package queue provides a download queue backed by a separate SQLite database,
// mirroring Crystal's Queue class (src/queue.cr).
package queue

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// JobStatus mirrors Crystal Queue::JobStatus enum.
type JobStatus int

const (
	StatusPending     JobStatus = 0
	StatusDownloading JobStatus = 1
	StatusError       JobStatus = 2
	StatusCompleted   JobStatus = 3
	StatusMissingPages JobStatus = 4
)

func (s JobStatus) String() string {
	switch s {
	case StatusPending:
		return "Pending"
	case StatusDownloading:
		return "Downloading"
	case StatusError:
		return "Error"
	case StatusCompleted:
		return "Completed"
	case StatusMissingPages:
		return "MissingPages"
	default:
		return "Unknown"
	}
}

// Job mirrors Crystal Queue::Job struct.
type Job struct {
	ID               string
	MangaID          string
	Title            string
	MangaTitle       string
	Status           JobStatus
	StatusMessage    string
	Pages            int
	SuccessCount     int
	FailCount        int
	Time             time.Time
	PluginID         string
	PluginChapterID  string
}

// Queue wraps a SQLite database for managing download jobs.
type Queue struct {
	db *sql.DB
}

// NewQueue opens (or creates) the queue database at dbPath and ensures the
// queue table exists.
func NewQueue(dbPath string) (*Queue, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create queue db directory: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)

	schema := `CREATE TABLE IF NOT EXISTS queue (
		id TEXT,
		manga_id TEXT,
		title TEXT,
		manga_title TEXT,
		status INTEGER,
		status_message TEXT,
		pages INTEGER,
		success_count INTEGER,
		fail_count INTEGER,
		time INTEGER
	);
	CREATE UNIQUE INDEX IF NOT EXISTS id_idx ON queue (id);
	CREATE INDEX IF NOT EXISTS manga_id_idx ON queue (manga_id);
	CREATE INDEX IF NOT EXISTS status_idx ON queue (status);`

	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("init queue schema: %w", err)
	}

	return &Queue{db: db}, nil
}

// Close closes the underlying database.
func (q *Queue) Close() error {
	return q.db.Close()
}

// Push inserts jobs into the queue. Duplicate IDs are ignored.
// Returns the number of jobs actually inserted.
func (q *Queue) Push(jobs []Job) (int, error) {
	startCount, err := q.Count()
	if err != nil {
		return 0, err
	}

	tx, err := q.db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`INSERT OR IGNORE INTO queue
		(id, manga_id, title, manga_title, status, status_message, pages, success_count, fail_count, time)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	for _, job := range jobs {
		_, err := stmt.Exec(
			job.ID, job.MangaID, job.Title, job.MangaTitle,
			int(job.Status), job.StatusMessage, job.Pages,
			job.SuccessCount, job.FailCount, job.Time.UnixMilli(),
		)
		if err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	endCount, err := q.Count()
	if err != nil {
		return 0, err
	}
	return endCount - startCount, nil
}

// Reset resets a single job back to Pending status.
func (q *Queue) Reset(id string) error {
	_, err := q.db.Exec(
		`UPDATE queue SET status = 0, status_message = '', pages = 0, success_count = 0, fail_count = 0 WHERE id = ?`,
		id,
	)
	return err
}

// ResetAll resets all failed jobs (StatusError and StatusMissingPages) back to Pending.
func (q *Queue) ResetAll() error {
	_, err := q.db.Exec(
		`UPDATE queue SET status = 0, status_message = '', pages = 0, success_count = 0, fail_count = 0 WHERE status = 2 OR status = 4`,
	)
	return err
}

// Delete removes a job by ID.
func (q *Queue) Delete(id string) error {
	_, err := q.db.Exec("DELETE FROM queue WHERE id = ?", id)
	return err
}

// Exists returns true if a job with the given ID exists.
func (q *Queue) Exists(id string) (bool, error) {
	var count int
	err := q.db.QueryRow("SELECT COUNT(*) FROM queue WHERE id = ?", id).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// DeleteStatus removes all jobs with the given status.
func (q *Queue) DeleteStatus(status JobStatus) error {
	_, err := q.db.Exec("DELETE FROM queue WHERE status = ?", int(status))
	return err
}

// CountStatus returns the number of jobs with the given status.
func (q *Queue) CountStatus(status JobStatus) (int, error) {
	var count int
	err := q.db.QueryRow("SELECT COUNT(*) FROM queue WHERE status = ?", int(status)).Scan(&count)
	return count, err
}

// Count returns the total number of jobs in the queue.
func (q *Queue) Count() (int, error) {
	var count int
	err := q.db.QueryRow("SELECT COUNT(*) FROM queue").Scan(&count)
	return count, err
}

// PendingCount returns the number of Pending + Downloading jobs.
func (q *Queue) PendingCount() (int, error) {
	var count int
	err := q.db.QueryRow(
		"SELECT COUNT(*) FROM queue WHERE status = 0 OR status = 1",
	).Scan(&count)
	return count, err
}

// SetStatus updates the status of a job.
func (q *Queue) SetStatus(status JobStatus, id string) error {
	_, err := q.db.Exec("UPDATE queue SET status = ? WHERE id = ?", int(status), id)
	return err
}

// SetPages sets the total page count for a job and resets success/fail counters.
func (q *Queue) SetPages(pages int, id string) error {
	_, err := q.db.Exec(
		"UPDATE queue SET pages = ?, success_count = 0, fail_count = 0 WHERE id = ?",
		pages, id,
	)
	return err
}

// AddSuccess increments the success count for a job.
func (q *Queue) AddSuccess(id string) error {
	_, err := q.db.Exec("UPDATE queue SET success_count = success_count + 1 WHERE id = ?", id)
	return err
}

// AddFail increments the fail count for a job.
func (q *Queue) AddFail(id string) error {
	_, err := q.db.Exec("UPDATE queue SET fail_count = fail_count + 1 WHERE id = ?", id)
	return err
}

// AddMessage appends a message to the status_message field.
func (q *Queue) AddMessage(msg, id string) error {
	_, err := q.db.Exec(
		"UPDATE queue SET status_message = status_message || ? || ? WHERE id = ?",
		"\n", msg, id,
	)
	return err
}

// List returns all jobs ordered by time.
func (q *Queue) List() ([]Job, error) {
	rows, err := q.db.Query("SELECT * FROM queue ORDER BY time")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanJobs(rows)
}

// Get returns a single job by ID.
func (q *Queue) Get(id string) (*Job, error) {
	row := q.db.QueryRow("SELECT * FROM queue WHERE id = ?", id)
	return scanJob(row)
}

// PopDownloadable returns the next downloadable job (status Pending or
// Downloading) with a plugin-style ID (contains '-' to separate plugin and
// chapter), ordered by time. This mirrors the Crystal Downloader#pop query.
func (q *Queue) PopDownloadable() (*Job, error) {
	row := q.db.QueryRow(
		`SELECT * FROM queue WHERE id LIKE '%-%' AND (status = 0 OR status = 1) ORDER BY time LIMIT 1`,
	)
	return scanJob(row)
}

// --- internal helpers ---

func scanJobs(rows *sql.Rows) ([]Job, error) {
	var jobs []Job
	for rows.Next() {
		j, err := scanJobFromScanner(rows)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, *j)
	}
	return jobs, rows.Err()
}

type scanner interface {
	Scan(dest ...any) error
}

func scanJobFromScanner(s scanner) (*Job, error) {
	var (
		id, mangaID, title, mangaTitle, statusMsg string
		statusInt                                 int
		pages, successCount, failCount            int
		timeMs                                    int64
	)
	err := s.Scan(&id, &mangaID, &title, &mangaTitle, &statusInt, &statusMsg, &pages, &successCount, &failCount, &timeMs)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	j := &Job{
		ID:            id,
		MangaID:       mangaID,
		Title:         title,
		MangaTitle:    mangaTitle,
		Status:        JobStatus(statusInt),
		StatusMessage: statusMsg,
		Pages:         pages,
		SuccessCount:  successCount,
		FailCount:     failCount,
		Time:          time.UnixMilli(timeMs),
	}

	// Parse plugin_id and plugin_chapter_id from ID (pluginID-base64(chapterID))
	// Same logic as Crystal Queue::Job#parse_query_result and Base64.decode_string
	if len(id) > 0 {
		ary := splitID(id)
		if len(ary) == 2 {
			j.PluginID = ary[0]
			decoded, err := base64.StdEncoding.DecodeString(ary[1])
			if err == nil {
				j.PluginChapterID = string(decoded)
			} else {
				// Backward compatibility: earlier versions didn't encode the chapter ID
				j.PluginChapterID = ary[1]
			}
		}
	}

	return j, nil
}

func scanJob(row *sql.Row) (*Job, error) {
	return scanJobFromScanner(row)
}

// splitID splits an ID like "pluginid-base64chapterid" into two parts.
func splitID(id string) []string {
	for i := 0; i < len(id); i++ {
		if id[i] == '-' {
			return []string{id[:i], id[i+1:]}
		}
	}
	return nil
}
