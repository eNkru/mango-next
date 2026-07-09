package migration

// Migration is a single versioned schema change, mirroring MG::Base subclasses
// in the Crystal migration/ directory. Only Up is needed for forward migration
// of existing/new databases; Down is provided for completeness.
type Migration struct {
	Version int
	Name    string
	Up      string
	Down    string
}

// All returns every migration sorted by version. These are transcribed verbatim
// from migration/*.cr so that a database created by Mango-Go is byte-for-byte
// schema-compatible with one created by the Crystal version.
//
// Note: migrations 8 and 10 in Crystal are data migrations that depend on
// Config.current.library_path. They rewrite absolute paths to relative paths.
// Because Mango-Go opens EXISTING databases (already at version >= 13 in the
// wild), these run only for brand-new installs where the tables are empty, so
// the REPLACE/SUBSTR statements are no-ops. We still register them (with the
// library path substituted at runtime) to keep version numbering identical.
func All(libraryPath string) []Migration {
	esc := func(s string) string {
		// Escape single quotes and strip a trailing slash, matching
		// relative_path.8.cr: base.gsub("'", "''").rstrip("/")
		out := ""
		for _, r := range s {
			if r == '\'' {
				out += "''"
			} else {
				out += string(r)
			}
		}
		for len(out) > 0 && out[len(out)-1] == '/' {
			out = out[:len(out)-1]
		}
		return out
	}
	base := esc(libraryPath)

	return []Migration{
		{Version: 1, Name: "CreateUsers", Up: `
CREATE TABLE IF NOT EXISTS users (
  username TEXT NOT NULL,
  password TEXT NOT NULL,
  token TEXT,
  admin INTEGER NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS username_idx ON users (username);
CREATE UNIQUE INDEX IF NOT EXISTS token_idx ON users (token);`},

		{Version: 2, Name: "CreateIds", Up: `
CREATE TABLE IF NOT EXISTS ids (
  path TEXT NOT NULL,
  id TEXT NOT NULL,
  is_title INTEGER NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS path_idx ON ids (path);
CREATE UNIQUE INDEX IF NOT EXISTS id_idx ON ids (id);`},

		{Version: 3, Name: "CreateThumbnails", Up: `
CREATE TABLE IF NOT EXISTS thumbnails (
  id TEXT NOT NULL,
  data BLOB NOT NULL,
  filename TEXT NOT NULL,
  mime TEXT NOT NULL,
  size INTEGER NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS tn_index ON thumbnails (id);`},

		{Version: 4, Name: "CreateTags", Up: `
CREATE TABLE IF NOT EXISTS tags (
  id TEXT NOT NULL,
  tag TEXT NOT NULL,
  UNIQUE (id, tag)
);
CREATE INDEX IF NOT EXISTS tags_id_idx ON tags (id);
CREATE INDEX IF NOT EXISTS tags_tag_idx ON tags (tag);`},

		{Version: 5, Name: "CreateTitles", Up: `
CREATE TABLE titles (
  id TEXT NOT NULL,
  path TEXT NOT NULL,
  signature TEXT
);
CREATE UNIQUE INDEX titles_id_idx on titles (id);
CREATE UNIQUE INDEX titles_path_idx on titles (path);
INSERT INTO titles SELECT id, path, null FROM ids WHERE is_title = 1;
DELETE FROM ids WHERE is_title = 1;
ALTER TABLE ids RENAME TO tmp;
CREATE TABLE ids (
  path TEXT NOT NULL,
  id TEXT NOT NULL
);
INSERT INTO ids SELECT path, id FROM tmp;
DROP TABLE tmp;
CREATE UNIQUE INDEX path_idx ON ids (path);
CREATE UNIQUE INDEX id_idx ON ids (id);`},

		{Version: 6, Name: "ForeignKeys", Up: `
ALTER TABLE tags RENAME TO tmp;
CREATE TABLE tags (
  id TEXT NOT NULL,
  tag TEXT NOT NULL,
  UNIQUE (id, tag),
  FOREIGN KEY (id) REFERENCES titles (id) ON UPDATE CASCADE ON DELETE CASCADE
);
INSERT INTO tags SELECT * FROM tmp;
DROP TABLE tmp;
CREATE INDEX tags_id_idx ON tags (id);
CREATE INDEX tags_tag_idx ON tags (tag);
ALTER TABLE thumbnails RENAME TO tmp;
CREATE TABLE thumbnails (
  id TEXT NOT NULL,
  data BLOB NOT NULL,
  filename TEXT NOT NULL,
  mime TEXT NOT NULL,
  size INTEGER NOT NULL,
  FOREIGN KEY (id) REFERENCES ids (id) ON UPDATE CASCADE ON DELETE CASCADE
);
INSERT INTO thumbnails SELECT * FROM tmp;
DROP TABLE tmp;
CREATE UNIQUE INDEX tn_index ON thumbnails (id);`},

		{Version: 7, Name: "IDSignature", Up: `
ALTER TABLE ids ADD COLUMN signature TEXT;`},

		{Version: 8, Name: "RelativePath", Up: `
UPDATE ids SET path = REPLACE(path, '` + base + `', '');
UPDATE titles SET path = REPLACE(path, '` + base + `', '');`},

		{Version: 9, Name: "UnavailableIDs", Up: `
ALTER TABLE ids ADD COLUMN unavailable INTEGER NOT NULL DEFAULT 0;
ALTER TABLE titles ADD COLUMN unavailable INTEGER NOT NULL DEFAULT 0;`},

		{Version: 10, Name: "RelativePathFix", Up: `
UPDATE ids SET path = SUBSTR(path, 2, LENGTH(path) - 1) WHERE path LIKE '/%';
UPDATE titles SET path = SUBSTR(path, 2, LENGTH(path) - 1) WHERE path LIKE '/%';`},

		{Version: 11, Name: "CreateMangaDexAccount", Up: `
CREATE TABLE md_account (
  username TEXT NOT NULL PRIMARY KEY,
  token TEXT NOT NULL,
  expire INTEGER NOT NULL,
  FOREIGN KEY (username) REFERENCES users (username) ON UPDATE CASCADE ON DELETE CASCADE
);`},

		{Version: 12, Name: "SortTitle", Up: `
ALTER TABLE ids ADD COLUMN sort_title TEXT;
ALTER TABLE titles ADD COLUMN sort_title TEXT;`},

		{Version: 13, Name: "HiddenTitles", Up: `
ALTER TABLE titles ADD COLUMN hidden INTEGER NOT NULL DEFAULT 0;`},

		{Version: 14, Name: "CreateProgress", Up: `
CREATE TABLE IF NOT EXISTS progress (
  username TEXT NOT NULL,
  title_id TEXT NOT NULL,
  entry_id TEXT,
  page INTEGER NOT NULL DEFAULT 0,
  updated_at INTEGER NOT NULL,
  PRIMARY KEY (username, title_id, entry_id),
  FOREIGN KEY (username) REFERENCES users(username) ON UPDATE CASCADE ON DELETE CASCADE,
  FOREIGN KEY (title_id) REFERENCES titles(id) ON UPDATE CASCADE ON DELETE CASCADE
);`},
	}
}

// LatestVersion returns the highest migration version.
func LatestVersion() int { return 14 }
