package storage

import (
	"compress/gzip"
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/hkalexling/mango-go/internal/storage/migration"
	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

// Storage wraps the SQLite database, mirroring the Crystal Storage class.
type Storage struct {
	db          *sql.DB
	path        string
	libraryPath string
}

// Image mirrors the Crystal Image struct (src/library/types.cr).
type Image struct {
	Data     []byte
	Filename string
	Mime     string
	Size     int
}

// User represents a row in the users table.
type User struct {
	Username string
	Password string // bcrypt hash
	Token    string // nullable
	IsAdmin  bool
}

// Open opens (creating the parent directory if needed) the SQLite database at
// path, runs migrations forward to the latest version, and returns a Storage.
//
// libraryPath is required for the version-8/10 data migrations (no-ops on
// existing or empty databases, see migration package docs).
func Open(path, libraryPath string) (*Storage, error) {
	dir := filepath.Dir(path)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("creating db directory %s: %w", dir, err)
		}
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	// SQLite with a single writer: cap connections to avoid "database is locked".
	db.SetMaxOpenConns(1)

	// Enable foreign keys.
	if _, err := db.Exec("PRAGMA foreign_keys = 1"); err != nil {
		db.Close()
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}

	s := &Storage{db: db, path: path, libraryPath: libraryPath}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("db migration failed: %w", err)
	}

	// Init admin if no users exist (matching Crystal Storage#initialize).
	var userCount int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&userCount); err != nil {
		db.Close()
		return nil, fmt.Errorf("count users: %w", err)
	}
	if userCount == 0 {
		if err := s.InitAdmin(); err != nil {
			db.Close()
			return nil, fmt.Errorf("init admin: %w", err)
		}
	}

	return s, nil
}

// DB exposes the underlying *sql.DB for other packages during migration.
func (s *Storage) DB() *sql.DB { return s.db }

// Close closes the database.
func (s *Storage) Close() error { return s.db.Close() }

// migrate advances the schema to the latest version using SQLite's built-in
// user_version pragma, exactly as MG does for SQLite (version_table/column are
// ignored; user_version stores the version).
func (s *Storage) migrate() error {
	var currentVer int
	if err := s.db.QueryRow("PRAGMA user_version").Scan(&currentVer); err != nil {
		return err
	}

	migrations := migration.All(s.libraryPath)
	for _, m := range migrations {
		if m.Version <= currentVer {
			continue
		}
		tx, err := s.db.Begin()
		if err != nil {
			return err
		}
		if _, err := tx.Exec(m.Up); err != nil {
			tx.Rollback()
			return fmt.Errorf("migration %d (%s): %w", m.Version, m.Name, err)
		}
		// user_version cannot be parameterized; version is a trusted int literal.
		if _, err := tx.Exec(fmt.Sprintf("PRAGMA user_version = %d", m.Version)); err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Commit(); err != nil {
			return err
		}
	}
	return nil
}

// Version returns the current schema version stored in user_version.
func (s *Storage) Version() (int, error) {
	var v int
	err := s.db.QueryRow("PRAGMA user_version").Scan(&v)
	return v, err
}

// ---------------------------------------------------------------------------
// User CRUD — matching Crystal Storage user methods
// ---------------------------------------------------------------------------

// randomStr generates a UUID v4 string without dashes, matching the Crystal
// random_str helper.
func randomStr() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}

// hashPassword returns the bcrypt hash of the given password, matching
// Crypto::Bcrypt::Password.create(pw).to_s in Crystal.
func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// verifyPassword compares a bcrypt hash with a plaintext password, matching
// Crypto::Bcrypt::Password.new(hash).verify(pw) in Crystal.
func verifyPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// validateUsername enforces the same rules as validate_username in
// src/util/validation.cr.
func validateUsername(username string) error {
	if len(username) < 3 {
		return fmt.Errorf("username should contain at least 3 characters")
	}
	// Crystal: /^[a-zA-Z_][a-zA-Z0-9_\-]*$/
	re := regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_\-]*$`)
	if !re.MatchString(username) {
		return fmt.Errorf("username can only contain alphanumeric characters, underscores, and hyphens")
	}
	return nil
}

// validatePassword enforces the same rules as validate_password in
// src/util/validation.cr.
func validatePassword(password string) error {
	if len(password) < 6 {
		return fmt.Errorf("password should contain at least 6 characters")
	}
	// Crystal: /^[[:ascii:]]+$/
	for _, r := range password {
		if r > 127 {
			return fmt.Errorf("password should contain ASCII characters only")
		}
	}
	return nil
}

// InitAdmin creates the initial admin user with a random password, matching
// the init_admin macro in storage.cr.
func (s *Storage) InitAdmin() error {
	pw := randomStr()
	hash, err := hashPassword(pw)
	if err != nil {
		return err
	}
	if _, err := s.db.Exec(
		"INSERT INTO users VALUES (?, ?, ?, ?)",
		"admin", hash, nil, 1,
	); err != nil {
		return fmt.Errorf("create admin user: %w", err)
	}
	log.Printf("Initial user created. You can log in with {\"username\": \"admin\", \"password\": %q}", pw)
	return nil
}

// UsernameExists returns true if a user with the given username exists.
func (s *Storage) UsernameExists(username string) (bool, error) {
	var count int
	err := s.db.QueryRow(
		"SELECT COUNT(*) FROM users WHERE username = ?", username,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// UsernameIsAdmin returns true if the given username has admin privileges.
func (s *Storage) UsernameIsAdmin(username string) (bool, error) {
	var admin int
	err := s.db.QueryRow(
		"SELECT admin FROM users WHERE username = ?", username,
	).Scan(&admin)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return admin > 0, nil
}

// VerifyUser checks the given username/password pair. On success, it ensures a
// token exists for the user (generating a new one if needed) and returns it.
// Returns an empty string and no error if the password doesn't match.
func (s *Storage) VerifyUser(username, password string) (string, error) {
	var hash string
	var token sql.NullString
	err := s.db.QueryRow(
		"SELECT password, token FROM users WHERE username = ?", username,
	).Scan(&hash, &token)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}

	if !verifyPassword(hash, password) {
		return "", nil
	}

	// Return existing token or generate a new one.
	if token.Valid && token.String != "" {
		return token.String, nil
	}

	newToken := randomStr()
	if _, err := s.db.Exec(
		"UPDATE users SET token = ? WHERE username = ?",
		newToken, username,
	); err != nil {
		return "", err
	}
	return newToken, nil
}

// VerifyToken returns the username associated with the given token, or an
// empty string if the token is invalid.
func (s *Storage) VerifyToken(token string) (string, error) {
	var username string
	err := s.db.QueryRow(
		"SELECT username FROM users WHERE token = ?", token,
	).Scan(&username)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	return username, nil
}

// VerifyAdmin returns true if the given token belongs to an admin user.
func (s *Storage) VerifyAdmin(token string) (bool, error) {
	var admin int
	err := s.db.QueryRow(
		"SELECT admin FROM users WHERE token = ?", token,
	).Scan(&admin)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return admin > 0, nil
}

// ListUsers returns all users with their admin status.
func (s *Storage) ListUsers() ([]User, error) {
	rows, err := s.db.Query("SELECT username, admin FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		var admin int
		if err := rows.Scan(&u.Username, &admin); err != nil {
			return nil, err
		}
		u.IsAdmin = admin > 0
		users = append(users, u)
	}
	return users, rows.Err()
}

// NewUser creates a new user with the given username, password, and admin flag.
func (s *Storage) NewUser(username, password string, admin bool) error {
	if err := validateUsername(username); err != nil {
		return err
	}
	if err := validatePassword(password); err != nil {
		return err
	}

	hash, err := hashPassword(password)
	if err != nil {
		return err
	}

	adminInt := 0
	if admin {
		adminInt = 1
	}
	if _, err := s.db.Exec(
		"INSERT INTO users VALUES (?, ?, ?, ?)",
		username, hash, nil, adminInt,
	); err != nil {
		return fmt.Errorf("insert user: %w", err)
	}
	return nil
}

// UpdateUser updates a user's details. If password is empty, it is not changed.
func (s *Storage) UpdateUser(originalUsername, username, password string, admin bool) error {
	if err := validateUsername(username); err != nil {
		return err
	}
	if password != "" {
		if err := validatePassword(password); err != nil {
			return err
		}
	}

	adminInt := 0
	if admin {
		adminInt = 1
	}

	// Check if removing last admin.
	if !admin {
		origAdmin, err := s.UsernameIsAdmin(originalUsername)
		if err != nil {
			return err
		}
		if origAdmin {
			count, err := s.adminCount()
			if err != nil {
				return err
			}
			if count <= 1 {
				return fmt.Errorf("cannot remove the last admin user")
			}
		}
	}

	if password == "" {
		_, err := s.db.Exec(
			"UPDATE users SET username = ?, admin = ? WHERE username = ?",
			username, adminInt, originalUsername,
		)
		return err
	}

	hash, err := hashPassword(password)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(
		"UPDATE users SET username = ?, admin = ?, password = ? WHERE username = ?",
		username, adminInt, hash, originalUsername,
	)
	return err
}

// DeleteUser removes a user. It refuses to delete the last admin.
func (s *Storage) DeleteUser(username string) error {
	isAdmin, err := s.UsernameIsAdmin(username)
	if err != nil {
		return err
	}
	if isAdmin {
		count, err := s.adminCount()
		if err != nil {
			return err
		}
		if count <= 1 {
			return fmt.Errorf("cannot remove the last admin user")
		}
	}

	result, err := s.db.Exec("DELETE FROM users WHERE username = ?", username)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return fmt.Errorf("user %q not found", username)
	}
	return nil
}

// Logout clears the token for the given token value.
func (s *Storage) Logout(token string) error {
	_, err := s.db.Exec("UPDATE users SET token = NULL WHERE token = ?", token)
	return err
}

// adminCount returns the number of users with admin privileges.
func (s *Storage) adminCount() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM users WHERE admin = 1").Scan(&count)
	return count, err
}

// CountUsers returns the total number of users.
func (s *Storage) CountUsers() (int, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	return count, err
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

// ---------------------------------------------------------------------------
// Library cache — gzip-compressed JSON (mirrors Crystal's library.yml.gz)
// ---------------------------------------------------------------------------

// SaveLibraryCache writes gzip-compressed data to the cache path.
func (s *Storage) SaveLibraryCache(data []byte) error {
	// Determine cache path: store next to the library dir (matching Crystal's
	// default library_cache_path).
	cachePath := filepath.Join(filepath.Dir(s.libraryPath), "library.yml.gz")

	dir := filepath.Dir(cachePath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	f, err := os.Create(cachePath)
	if err != nil {
		return err
	}
	defer f.Close()

	w := gzip.NewWriter(f)
	defer w.Close()
	_, err = w.Write(data)
	return err
}

// LoadLibraryCache reads and decompresses the gzip cache file.
func (s *Storage) LoadLibraryCache() ([]byte, error) {
	cachePath := filepath.Join(filepath.Dir(s.libraryPath), "library.yml.gz")
	f, err := os.Open(cachePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return io.ReadAll(r)
}
