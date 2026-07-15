package storage

import (
	"database/sql"
	"encoding/json"
	"time"
)

// PageDimension is width/height for one reader page (JSON: width, height).
type PageDimension struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

// GetEntryDimensions returns cached page dimensions when a row exists for id
// with a matching signature. ok is false on miss, signature mismatch, or
// corrupt payload (err is nil for those cases).
func (s *Storage) GetEntryDimensions(id, signature string) (dims []PageDimension, ok bool, err error) {
	var storedSig, raw string
	var pageCount int
	err = s.db.QueryRow(
		`SELECT signature, dimensions, page_count FROM entry_dimensions WHERE id = ?`,
		id,
	).Scan(&storedSig, &raw, &pageCount)
	if err == sql.ErrNoRows {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, err
	}
	if storedSig != signature {
		return nil, false, nil
	}
	if err := json.Unmarshal([]byte(raw), &dims); err != nil {
		return nil, false, nil
	}
	if pageCount != len(dims) {
		return nil, false, nil
	}
	return dims, true, nil
}

// SaveEntryDimensions upserts dimensions for an entry id.
func (s *Storage) SaveEntryDimensions(id, signature string, dims []PageDimension) error {
	raw, err := json.Marshal(dims)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(
		`INSERT INTO entry_dimensions (id, signature, dimensions, page_count, updated_at)
		 VALUES (?, ?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET
		   signature = excluded.signature,
		   dimensions = excluded.dimensions,
		   page_count = excluded.page_count,
		   updated_at = excluded.updated_at`,
		id, signature, string(raw), len(dims), time.Now().Unix(),
	)
	return err
}
