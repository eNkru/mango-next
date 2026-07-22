package storage

import (
	"database/sql"
	"strings"
)

// ---------------------------------------------------------------------------
// Hidden titles — matching Crystal Storage hidden-title methods
// ---------------------------------------------------------------------------

// SetTitleHidden sets the hidden flag on a title (0 or 1).
func (s *Storage) SetTitleHidden(titleID string, hidden int) error {
	_, err := s.db.Exec("UPDATE titles SET hidden = ? WHERE id = ?", hidden, titleID)
	return err
}

// GetHiddenTitleIDs returns all title ids that are hidden.
func (s *Storage) GetHiddenTitleIDs() ([]string, error) {
	rows, err := s.db.Query("SELECT id FROM titles WHERE hidden = 1")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

// GetTitleHidden returns the hidden value (0 or 1) for a title.
func (s *Storage) GetTitleHidden(titleID string) (int, error) {
	var hidden int
	err := s.db.QueryRow(
		"SELECT hidden FROM titles WHERE id = ?", titleID,
	).Scan(&hidden)
	if err != nil {
		return 0, err
	}
	return hidden, nil
}

// ---------------------------------------------------------------------------
// Sort titles — matching Crystal Storage sort-title methods
// ---------------------------------------------------------------------------

// GetTitleSortTitle returns the sort_title override for a title.
func (s *Storage) GetTitleSortTitle(titleID string) (*string, error) {
	var sortTitle sql.NullString
	err := s.db.QueryRow(
		"SELECT sort_title FROM titles WHERE id = ?", titleID,
	).Scan(&sortTitle)
	if err != nil {
		return nil, err
	}
	if sortTitle.Valid {
		return &sortTitle.String, nil
	}
	return nil, nil
}

// SetTitleSortTitle sets the sort_title override for a title. An empty string
// or nil clears the value (matching the Crystal behaviour of setting nil when
// sort_title is an empty string).
func (s *Storage) SetTitleSortTitle(titleID string, sortTitle *string) error {
	v := sortTitle
	if v != nil && *v == "" {
		v = nil
	}
	_, err := s.db.Exec(
		"UPDATE titles SET sort_title = ? WHERE id = ?", v, titleID,
	)
	return err
}

// GetEntrySortTitle returns the sort_title override for an entry.
func (s *Storage) GetEntrySortTitle(entryID string) (*string, error) {
	var sortTitle sql.NullString
	err := s.db.QueryRow(
		"SELECT sort_title FROM ids WHERE id = ?", entryID,
	).Scan(&sortTitle)
	if err != nil {
		return nil, err
	}
	if sortTitle.Valid {
		return &sortTitle.String, nil
	}
	return nil, nil
}

// GetEntriesSortTitle returns a map of entry ID to sort_title for the given
// entry IDs, matching Crystal Storage#get_entries_sort_title.
func (s *Storage) GetEntriesSortTitle(ids []string) (map[string]*string, error) {
	if len(ids) == 0 {
		return map[string]*string{}, nil
	}

	// Build placeholders.
	placeholders := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"
		args[i] = id
	}

	query := "SELECT id, sort_title FROM ids WHERE id IN (" +
		strings.Join(placeholders, ",") + ")"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make(map[string]*string)
	for rows.Next() {
		var id string
		var sortTitle sql.NullString
		if err := rows.Scan(&id, &sortTitle); err != nil {
			return nil, err
		}
		if sortTitle.Valid {
			results[id] = &sortTitle.String
		} else {
			results[id] = nil
		}
	}
	return results, rows.Err()
}

// SetEntrySortTitle sets the sort_title override for an entry. An empty
// string or nil clears the value.
func (s *Storage) SetEntrySortTitle(entryID string, sortTitle *string) error {
	v := sortTitle
	if v != nil && *v == "" {
		v = nil
	}
	_, err := s.db.Exec(
		"UPDATE ids SET sort_title = ? WHERE id = ?", v, entryID,
	)
	return err
}

// ---------------------------------------------------------------------------
// Other helpers
// ---------------------------------------------------------------------------

// CountTitles returns the total number of titles in the database.
func (s *Storage) CountTitles() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM titles").Scan(&count)
	return count, err
}

// CountEntries returns the total number of entries in the ids table.
func (s *Storage) CountEntries() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM ids").Scan(&count)
	return count, err
}
