package storage

import (
	"database/sql"
	"path/filepath"
	"strconv"
)

// ---------------------------------------------------------------------------
// Library ID management — matching Crystal Storage get/insert/bulk/unavailable
// ---------------------------------------------------------------------------

// relPath converts an absolute path to a library-relative path using forward
// slashes, matching the Crystal Path.new(p).relative_to(library_path).to_s.
func (s *Storage) relPath(absPath string) string {
	rel, err := filepath.Rel(s.libraryPath, absPath)
	if err != nil {
		return filepath.ToSlash(absPath)
	}
	return filepath.ToSlash(rel)
}

// absPath converts a library-relative path (using forward slashes) to an
// absolute path.
func (s *Storage) absPath(relPath string) string {
	return filepath.Join(s.libraryPath, filepath.FromSlash(relPath))
}

// TitleRecord mirrors a row in the titles table.
type TitleRecord struct {
	ID          string
	Path        string
	Signature   string
	Unavailable int
	SortTitle   *string
	Hidden      int
}

// EntryRecord mirrors a row in the ids table.
type EntryRecord struct {
	ID          string
	Path        string
	Signature   string
	Unavailable int
	SortTitle   *string
}

// GetOrCreateTitleID looks up a title by its absolute path. On match it
// updates the signature (and clears unavailable), then returns the existing
// ID. On miss it creates a new row and returns a fresh UUID.
func (s *Storage) GetOrCreateTitleID(absPath string, sig uint64) (string, error) {
	relPath := s.relPath(absPath)
	sigStr := strconv.FormatUint(sig, 10)

	// 1. Try path + signature exact match.
	var id string
	err := s.db.QueryRow(
		"SELECT id FROM titles WHERE path = ? AND signature = ? AND unavailable = 0",
		relPath, sigStr,
	).Scan(&id)
	if err == nil {
		return id, nil
	}
	if err != sql.ErrNoRows {
		return "", err
	}

	// 2. Try path-only match (signature changed — e.g. new file added to dir).
	err = s.db.QueryRow(
		"SELECT id FROM titles WHERE path = ?", relPath,
	).Scan(&id)
	if err == nil {
		_, err := s.db.Exec(
			"UPDATE titles SET signature = ?, unavailable = 0 WHERE id = ?",
			sigStr, id,
		)
		return id, err
	}
	if err != sql.ErrNoRows {
		return "", err
	}

	// 3. Not found at all — insert fresh row.
	id = randomStr()
	_, err = s.db.Exec(
		"INSERT INTO titles (id, path, signature, unavailable) VALUES (?, ?, ?, 0)",
		id, relPath, sigStr,
	)
	return id, err
}

// GetOrCreateEntryID is the entry (ids table) counterpart of
// GetOrCreateTitleID.
func (s *Storage) GetOrCreateEntryID(absPath string, sig uint64) (string, error) {
	relPath := s.relPath(absPath)
	sigStr := strconv.FormatUint(sig, 10)

	var id string
	err := s.db.QueryRow(
		"SELECT id FROM ids WHERE path = ? AND signature = ? AND unavailable = 0",
		relPath, sigStr,
	).Scan(&id)
	if err == nil {
		return id, nil
	}
	if err != sql.ErrNoRows {
		return "", err
	}

	err = s.db.QueryRow(
		"SELECT id FROM ids WHERE path = ?", relPath,
	).Scan(&id)
	if err == nil {
		_, err := s.db.Exec(
			"UPDATE ids SET signature = ?, unavailable = 0 WHERE id = ?",
			sigStr, id,
		)
		return id, err
	}
	if err != sql.ErrNoRows {
		return "", err
	}

	id = randomStr()
	_, err = s.db.Exec(
		"INSERT INTO ids (id, path, signature, unavailable) VALUES (?, ?, ?, 0)",
		id, relPath, sigStr,
	)
	return id, err
}

// TitleIdentityMatches reports whether an available title ID belongs to the
// expected library path. It is read-only so callers can validate cached data
// without recreating database identities from stale input.
func (s *Storage) TitleIdentityMatches(id, absPath string) (bool, error) {
	return s.identityExists(
		"SELECT EXISTS(SELECT 1 FROM titles WHERE id = ? AND path = ? AND unavailable = 0)",
		id, s.relPath(absPath),
	)
}

// EntryIdentityMatches is the entry counterpart of TitleIdentityMatches.
func (s *Storage) EntryIdentityMatches(id, absPath string) (bool, error) {
	return s.identityExists(
		"SELECT EXISTS(SELECT 1 FROM ids WHERE id = ? AND path = ? AND unavailable = 0)",
		id, s.relPath(absPath),
	)
}

// TitleIDExists reports whether an available title ID exists. Cache payloads
// currently retain nested title IDs without their paths, so those references
// can only be validated by identity.
func (s *Storage) TitleIDExists(id string) (bool, error) {
	return s.identityExists(
		"SELECT EXISTS(SELECT 1 FROM titles WHERE id = ? AND unavailable = 0)", id,
	)
}

func (s *Storage) identityExists(query string, args ...any) (bool, error) {
	var exists bool
	err := s.db.QueryRow(query, args...).Scan(&exists)
	return exists, err
}
