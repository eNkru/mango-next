package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/eNkru/mango-next/internal/storage/migration"
	_ "modernc.org/sqlite"
)

// Storage wraps the SQLite database, mirroring the Crystal Storage class.
type Storage struct {
	db          *sql.DB
	path        string
	libraryPath string
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

	// SQLite pragmas for performance and concurrency.
	if _, err := db.Exec("PRAGMA journal_mode = WAL;"); err != nil {
		db.Close()
		return nil, fmt.Errorf("enable WAL mode: %w", err)
	}

	if _, err := db.Exec("PRAGMA busy_timeout = 5000;"); err != nil {
		db.Close()
		return nil, fmt.Errorf("set busy timeout: %w", err)
	}

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
