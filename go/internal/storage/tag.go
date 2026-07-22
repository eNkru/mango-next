package storage

// ---------------------------------------------------------------------------
// Tags — matching Crystal Storage tag methods
// ---------------------------------------------------------------------------

// GetTitleTags returns all tags for a given title id, ordered by tag.
func (s *Storage) GetTitleTags(id string) ([]string, error) {
	rows, err := s.db.Query(
		"SELECT tag FROM tags WHERE id = ? ORDER BY tag", id,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	return tags, rows.Err()
}

// GetTagTitles returns all title ids that have the given tag, optionally
// including hidden titles.
func (s *Storage) GetTagTitles(tag string, showHidden bool) ([]string, error) {
	query := "SELECT tags.id FROM tags NATURAL JOIN titles WHERE tag = ? AND unavailable = 0"
	if !showHidden {
		query += " AND hidden = 0"
	}
	rows, err := s.db.Query(query, tag)
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

// ListTags returns all distinct tags (excluding titles that are unavailable or
// hidden), matching Crystal Storage#list_tags.
func (s *Storage) ListTags() ([]string, error) {
	rows, err := s.db.Query(
		"SELECT DISTINCT tag FROM tags NATURAL JOIN titles WHERE unavailable = 0 AND hidden = 0",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	return tags, rows.Err()
}

// AddTag adds a tag to a title.
func (s *Storage) AddTag(id, tag string) error {
	_, err := s.db.Exec("INSERT INTO tags VALUES (?, ?)", id, tag)
	return err
}

// DeleteTag removes a tag from a title.
func (s *Storage) DeleteTag(id, tag string) error {
	_, err := s.db.Exec("DELETE FROM tags WHERE id = ? AND tag = ?", id, tag)
	return err
}
