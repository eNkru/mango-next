package storage

import "database/sql"

// Image mirrors the Crystal Image struct (src/library/types.cr).
type Image struct {
	Data     []byte
	Filename string
	Mime     string
	Size     int
}

// ---------------------------------------------------------------------------
// Thumbnails — matching Crystal Storage thumbnail methods
// ---------------------------------------------------------------------------

// SaveThumbnail inserts a thumbnail blob into the thumbnails table.
func (s *Storage) SaveThumbnail(id string, img *Image) error {
	_, err := s.db.Exec(
		"INSERT INTO thumbnails VALUES (?, ?, ?, ?, ?)",
		id, img.Data, img.Filename, img.Mime, img.Size,
	)
	return err
}

// GetThumbnail retrieves a thumbnail by id, or returns nil if not found.
func (s *Storage) GetThumbnail(id string) (*Image, error) {
	var img Image
	err := s.db.QueryRow(
		"SELECT * FROM thumbnails WHERE id = ?", id,
	).Scan(&id, &img.Data, &img.Filename, &img.Mime, &img.Size)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &img, nil
}

// DeleteThumbnail removes a thumbnail by id.
func (s *Storage) DeleteThumbnail(id string) error {
	_, err := s.db.Exec("DELETE FROM thumbnails WHERE id = ?", id)
	return err
}
