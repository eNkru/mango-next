package storage

import (
	"database/sql"
	"os"
)

// MarkUnavailable sets unavailable=1 on entries and titles whose files no
// longer exist. The candidate slices narrow which rows to check (may be nil
// to check all). Matching Crystal Storage#mark_unavailable.
func (s *Storage) MarkUnavailable(entryIDs, titleIDs []string) error {
	// Entries
	if len(entryIDs) > 0 {
		for _, id := range entryIDs {
			var relPath string
			err := s.db.QueryRow(
				"SELECT path FROM ids WHERE id = ? AND unavailable = 0", id,
			).Scan(&relPath)
			if err == sql.ErrNoRows {
				continue
			}
			if err != nil {
				return err
			}
			fullPath := s.absPath(relPath)
			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				_, err := s.db.Exec(
					"UPDATE ids SET unavailable = 1 WHERE id = ?", id,
				)
				if err != nil {
					return err
				}
			}
		}
	}

	// Titles
	if len(titleIDs) > 0 {
		for _, id := range titleIDs {
			var relPath string
			err := s.db.QueryRow(
				"SELECT path FROM titles WHERE id = ? AND unavailable = 0", id,
			).Scan(&relPath)
			if err == sql.ErrNoRows {
				continue
			}
			if err != nil {
				return err
			}
			fullPath := s.absPath(relPath)
			if _, err := os.Stat(fullPath); os.IsNotExist(err) {
				_, err := s.db.Exec(
					"UPDATE titles SET unavailable = 1 WHERE id = ?", id,
				)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// GetAllTitles returns all rows from the titles table where unavailable = 0.
func (s *Storage) GetAllTitles() ([]TitleRecord, error) {
	rows, err := s.db.Query(
		"SELECT id, path, signature, unavailable, sort_title, hidden FROM titles WHERE unavailable = 0",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []TitleRecord
	for rows.Next() {
		var r TitleRecord
		var sortTitle sql.NullString
		if err := rows.Scan(&r.ID, &r.Path, &r.Signature, &r.Unavailable, &sortTitle, &r.Hidden); err != nil {
			return nil, err
		}
		if sortTitle.Valid {
			r.SortTitle = &sortTitle.String
		}
		records = append(records, r)
	}
	return records, rows.Err()
}

// GetAllEntries returns all rows from the ids table where unavailable = 0.
func (s *Storage) GetAllEntries() ([]EntryRecord, error) {
	rows, err := s.db.Query(
		"SELECT id, path, signature, unavailable, sort_title FROM ids WHERE unavailable = 0",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []EntryRecord
	for rows.Next() {
		var r EntryRecord
		var sortTitle sql.NullString
		if err := rows.Scan(&r.ID, &r.Path, &r.Signature, &r.Unavailable, &sortTitle); err != nil {
			return nil, err
		}
		if sortTitle.Valid {
			r.SortTitle = &sortTitle.String
		}
		records = append(records, r)
	}
	return records, rows.Err()
}

// MissingItem is a title or entry marked unavailable (file no longer present).
type MissingItem struct {
	ID   string `json:"id"`
	Path string `json:"path"`
}

// ListMissingTitles returns unavailable title rows for the admin missing page.
func (s *Storage) ListMissingTitles() ([]MissingItem, error) {
	rows, err := s.db.Query(
		"SELECT id, path FROM titles WHERE unavailable = 1 ORDER BY path",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []MissingItem
	for rows.Next() {
		var item MissingItem
		if err := rows.Scan(&item.ID, &item.Path); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	if out == nil {
		out = []MissingItem{}
	}
	return out, rows.Err()
}

// ListMissingEntries returns unavailable entry rows for the admin missing page.
func (s *Storage) ListMissingEntries() ([]MissingItem, error) {
	rows, err := s.db.Query(
		"SELECT id, path FROM ids WHERE unavailable = 1 ORDER BY path",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []MissingItem
	for rows.Next() {
		var item MissingItem
		if err := rows.Scan(&item.ID, &item.Path); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	if out == nil {
		out = []MissingItem{}
	}
	return out, rows.Err()
}

// DeleteMissingTitle permanently removes one unavailable title and related
// metadata (tags/thumbnails cascade via FK where defined).
func (s *Storage) DeleteMissingTitle(id string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM tags WHERE id = ?", id); err != nil {
		return err
	}
	if _, err := tx.Exec("DELETE FROM thumbnails WHERE id = ?", id); err != nil {
		return err
	}
	if _, err := tx.Exec("DELETE FROM titles WHERE id = ? AND unavailable = 1", id); err != nil {
		return err
	}
	return tx.Commit()
}

// DeleteMissingEntry permanently removes one unavailable entry and its thumbnail.
func (s *Storage) DeleteMissingEntry(id string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM thumbnails WHERE id = ?", id); err != nil {
		return err
	}
	if _, err := tx.Exec("DELETE FROM ids WHERE id = ? AND unavailable = 1", id); err != nil {
		return err
	}
	return tx.Commit()
}

// DeleteAllMissingTitles removes every unavailable title and related metadata.
func (s *Storage) DeleteAllMissingTitles() error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(
		`DELETE FROM tags WHERE id IN (SELECT id FROM titles WHERE unavailable = 1)`,
	); err != nil {
		return err
	}
	if _, err := tx.Exec(
		`DELETE FROM thumbnails WHERE id IN (SELECT id FROM titles WHERE unavailable = 1)`,
	); err != nil {
		return err
	}
	if _, err := tx.Exec("DELETE FROM titles WHERE unavailable = 1"); err != nil {
		return err
	}
	return tx.Commit()
}

// DeleteAllMissingEntries removes every unavailable entry and related thumbnails.
func (s *Storage) DeleteAllMissingEntries() error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(
		`DELETE FROM thumbnails WHERE id IN (SELECT id FROM ids WHERE unavailable = 1)`,
	); err != nil {
		return err
	}
	if _, err := tx.Exec("DELETE FROM ids WHERE unavailable = 1"); err != nil {
		return err
	}
	return tx.Commit()
}
