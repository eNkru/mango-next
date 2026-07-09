package storage

import (
	"database/sql"
	"fmt"
	"time"
)

const createProgressTable = `
CREATE TABLE IF NOT EXISTS progress (
    username TEXT NOT NULL,
    title_id TEXT NOT NULL,
    entry_id TEXT,
    page INTEGER NOT NULL DEFAULT 0,
    updated_at INTEGER NOT NULL,
    PRIMARY KEY (username, title_id, entry_id)
);
`

const migrateProgress = `
INSERT OR REPLACE INTO progress (username, title_id, entry_id, page, updated_at)
SELECT p.username, p.title_id, p.entry_id, p.page, p.updated_at FROM progress;
`

type ProgressRecord struct {
	Username  string
	TitleID   string
	EntryID   *string
	Page      int
	UpdatedAt int64
}

type ContinueReadingItem struct {
	EntryID    string
	EntryName  string
	TitleID    string
	TitleName  string
	Page       int
	Percentage float64
	CoverURL   string
}

type StartReadingItem struct {
	TitleID   string
	TitleName string
	CoverURL  string
}

type RecentlyAddedItem struct {
	EntryID      string
	EntryName    string
	TitleID      string
	TitleName    string
	Percentage   float64
	GroupedCount int
	CoverURL     string
}

func (s *Storage) SaveProgress(username, titleID string, entryID *string, page int) error {
	now := time.Now().Unix()
	_, err := s.db.Exec(
		`INSERT INTO progress (username, title_id, entry_id, page, updated_at)
		 VALUES (?, ?, ?, ?, ?)
		 ON CONFLICT(username, title_id, entry_id) DO UPDATE SET page = ?, updated_at = ?`,
		username, titleID, entryID, page, now, page, now,
	)
	return err
}

func (s *Storage) LoadProgress(username, titleID string, entryID *string) (int, error) {
	var page int
	err := s.db.QueryRow(
		`SELECT page FROM progress WHERE username = ? AND title_id = ? AND entry_id IS ?`,
		username, titleID, entryID,
	).Scan(&page)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return page, nil
}

func (s *Storage) BulkMarkRead(username, titleID string, entryIDs []string) error {
	now := time.Now().Unix()
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, eid := range entryIDs {
		_, err := tx.Exec(
			`INSERT INTO progress (username, title_id, entry_id, page, updated_at)
			 VALUES (?, ?, ?, -1, ?)
			 ON CONFLICT(username, title_id, entry_id) DO UPDATE SET page = -1, updated_at = ?`,
			username, titleID, eid, now, now,
		)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Storage) BulkMarkUnread(username, titleID string, entryIDs []string) error {
	for _, eid := range entryIDs {
		_, err := s.db.Exec(
			`DELETE FROM progress WHERE username = ? AND title_id = ? AND entry_id = ?`,
			username, titleID, eid,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Storage) BulkMarkTitleRead(username, titleID string, pageCount int) error {
	now := time.Now().Unix()
	_, err := s.db.Exec(
		`INSERT INTO progress (username, title_id, entry_id, page, updated_at)
		 VALUES (?, ?, NULL, ?, ?)
		 ON CONFLICT(username, title_id, entry_id) DO UPDATE SET page = ?, updated_at = ?`,
		username, titleID, pageCount, now, pageCount, now,
	)
	return err
}

func (s *Storage) BulkMarkTitleUnread(username, titleID string) error {
	_, err := s.db.Exec(
		`DELETE FROM progress WHERE username = ? AND title_id = ?`,
		username, titleID,
	)
	return err
}

func (s *Storage) GetContinueReading(username string) ([]ContinueReadingItem, error) {
	rows, err := s.db.Query(
		`SELECT p.title_id, p.entry_id, p.page, p.updated_at
		 FROM progress p
		 JOIN titles t ON t.id = p.title_id
		 WHERE p.username = ? AND t.unavailable = 0 AND t.hidden = 0
		 ORDER BY p.updated_at DESC LIMIT 20`,
		username,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []ContinueReadingItem
	for rows.Next() {
		var titleID, entryID sql.NullString
		var page int
		var updatedAt int64
		if err := rows.Scan(&titleID, &entryID, &page, &updatedAt); err != nil {
			return nil, err
		}
		items = append(items, ContinueReadingItem{
			EntryID:  entryID.String,
			TitleID:  titleID.String,
			Page:     page,
		})
	}
	return items, rows.Err()
}

func (s *Storage) GetStartReading(username string) ([]StartReadingItem, error) {
	rows, err := s.db.Query(
		`SELECT t.id, t.path FROM titles t
		 WHERE t.unavailable = 0 AND t.hidden = 0
		 AND t.id NOT IN (SELECT DISTINCT title_id FROM progress WHERE username = ?)
		 ORDER BY t.id`,
		username,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []StartReadingItem
	for rows.Next() {
		var id, path string
		if err := rows.Scan(&id, &path); err != nil {
			return nil, err
		}
		items = append(items, StartReadingItem{TitleID: id})
	}
	return items, rows.Err()
}

func (s *Storage) GetRecentlyAdded(username string) ([]RecentlyAddedItem, error) {
	rows, err := s.db.Query(
		`SELECT t.id, t.path FROM titles t
		 WHERE t.unavailable = 0 AND t.hidden = 0
		 ORDER BY t.id DESC LIMIT 20`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []RecentlyAddedItem
	for rows.Next() {
		var id, path string
		if err := rows.Scan(&id, &path); err != nil {
			return nil, err
		}
		items = append(items, RecentlyAddedItem{TitleID: id})
	}
	return items, rows.Err()
}

func (s *Storage) MigrateProgressTable() error {
	_, err := s.db.Exec(createProgressTable)
	if err != nil {
		return fmt.Errorf("create progress table: %w", err)
	}
	return nil
}
