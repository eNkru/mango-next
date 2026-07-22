package storage

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/eNkru/mango-next/internal/storage/migration"
)

func TestMigrateFreshDB(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "mango.db")

	// Open will init admin since there are no users.
	st, err := Open(dbPath, filepath.Join(dir, "library"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	ver, err := st.Version()
	if err != nil {
		t.Fatal(err)
	}
	wantVersion := migration.LatestVersion()
	if ver != wantVersion {
		t.Errorf("schema version = %d, want %d", ver, wantVersion)
	}

	wantTables := []string{"users", "ids", "titles", "thumbnails", "tags", "md_account", "progress", "entry_dimensions"}
	for _, tbl := range wantTables {
		var name string
		err := st.DB().QueryRow(
			"SELECT name FROM sqlite_master WHERE type='table' AND name=?", tbl,
		).Scan(&name)
		if err != nil {
			t.Errorf("table %q missing: %v", tbl, err)
		}
	}

	var journalMode string
	if err := st.DB().QueryRow("PRAGMA journal_mode").Scan(&journalMode); err != nil {
		t.Errorf("query journal_mode: %v", err)
	} else if strings.ToLower(journalMode) != "wal" {
		t.Errorf("journal_mode = %q, want wal", journalMode)
	}

	var busyTimeout int
	if err := st.DB().QueryRow("PRAGMA busy_timeout").Scan(&busyTimeout); err != nil {
		t.Errorf("query busy_timeout: %v", err)
	} else if busyTimeout != 5000 {
		t.Errorf("busy_timeout = %d, want 5000", busyTimeout)
	}
}

func TestMigrateIsIdempotent(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "mango.db")
	lib := filepath.Join(dir, "library")

	st1, err := Open(dbPath, lib)
	if err != nil {
		t.Fatal(err)
	}
	// Seed a row, then close and reopen — migration must not touch it.
	if _, err := st1.DB().Exec(
		"INSERT INTO users VALUES ('alice','hash',NULL,1)"); err != nil {
		t.Fatal(err)
	}
	st1.Close()

	st2, err := Open(dbPath, lib)
	if err != nil {
		t.Fatal(err)
	}
	defer st2.Close()

	ver, _ := st2.Version()
	wantVersion := migration.LatestVersion()
	if ver != wantVersion {
		t.Errorf("version after reopen = %d, want %d", ver, wantVersion)
	}
	var n int
	st2.DB().QueryRow("SELECT COUNT(*) FROM users").Scan(&n)
	if n != 2 {
		t.Errorf("user row count = %d, want 2 (admin + alice, data must survive)", n)
	}
}

func TestTitlesColumns(t *testing.T) {
	dir := t.TempDir()
	st, err := Open(filepath.Join(dir, "mango.db"), filepath.Join(dir, "library"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	rows, err := st.DB().Query("PRAGMA table_info(titles)")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	cols := map[string]bool{}
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dflt any
		rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk)
		cols[name] = true
	}
	for _, want := range []string{"id", "path", "signature", "unavailable", "sort_title", "hidden"} {
		if !cols[want] {
			t.Errorf("titles missing column %q", want)
		}
	}
}

// ---------------------------------------------------------------------------
// User tests
// ---------------------------------------------------------------------------

func TestInitAdmin(t *testing.T) {
	dir := t.TempDir()
	st, err := Open(filepath.Join(dir, "mango.db"), filepath.Join(dir, "library"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	// Admin should have been auto-created.
	exists, err := st.UsernameExists("admin")
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("admin user was not auto-created")
	}

	admin, err := st.UsernameIsAdmin("admin")
	if err != nil {
		t.Fatal(err)
	}
	if !admin {
		t.Error("admin user should have admin flag")
	}

	// Admin should have a token generated on first login.
	// We need to use the password from the log output, but we can't easily
	// capture that. Instead, verify the user exists with a hashed password
	// we can match against by checking that VerifyUser with wrong pw fails.
	token, err := st.VerifyUser("admin", "wrongpassword")
	if err != nil {
		t.Fatal(err)
	}
	if token != "" {
		t.Error("VerifyUser with wrong password should return empty token")
	}

	// We can now count users — should be exactly 1 (admin).
	count, err := st.CountUsers()
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Errorf("user count = %d, want 1", count)
	}
}

func TestUserCRUD(t *testing.T) {
	dir := t.TempDir()
	st, err := Open(filepath.Join(dir, "mango.db"), filepath.Join(dir, "library"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	// Create a new non-admin user.
	if err := st.NewUser("testuser", "password123", false); err != nil {
		t.Fatal(err)
	}

	exists, err := st.UsernameExists("testuser")
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("testuser should exist")
	}

	// Verify login.
	token, err := st.VerifyUser("testuser", "password123")
	if err != nil {
		t.Fatal(err)
	}
	if token == "" {
		t.Fatal("expected a token for valid password")
	}

	// Verify token returns username.
	username, err := st.VerifyToken(token)
	if err != nil {
		t.Fatal(err)
	}
	if username != "testuser" {
		t.Errorf("verify token got username %q, want %q", username, "testuser")
	}

	// Verify non-admin.
	admin, err := st.VerifyAdmin(token)
	if err != nil {
		t.Fatal(err)
	}
	if admin {
		t.Error("testuser should not be admin")
	}

	// List users.
	users, err := st.ListUsers()
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, u := range users {
		if u.Username == "testuser" {
			found = true
			if u.IsAdmin {
				t.Error("testuser listed as admin")
			}
			break
		}
	}
	if !found {
		t.Error("testuser not found in list")
	}

	// Update user: change username, make admin.
	if err := st.UpdateUser("testuser", "testuser2", "newpass456", true); err != nil {
		t.Fatal(err)
	}
	exists, _ = st.UsernameExists("testuser")
	if exists {
		t.Error("old username should not exist after update")
	}
	admin, _ = st.UsernameIsAdmin("testuser2")
	if !admin {
		t.Error("updated user should be admin")
	}
	// Verify new password works.
	token, err = st.VerifyUser("testuser2", "newpass456")
	if err != nil {
		t.Fatal(err)
	}
	if token == "" {
		t.Fatal("expected token for updated password")
	}

	// Logout.
	if err := st.Logout(token); err != nil {
		t.Fatal(err)
	}
	username, _ = st.VerifyToken(token)
	if username != "" {
		t.Error("token should be invalid after logout")
	}

	// Delete user.
	if err := st.DeleteUser("testuser2"); err != nil {
		t.Fatal(err)
	}
	exists, _ = st.UsernameExists("testuser2")
	if exists {
		t.Error("deleted user should not exist")
	}
}

func TestVerifyUserReturnsExistingToken(t *testing.T) {
	dir := t.TempDir()
	st, err := Open(filepath.Join(dir, "mango.db"), filepath.Join(dir, "library"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	if err := st.NewUser("bob", "password123", false); err != nil {
		t.Fatal(err)
	}

	// First login — generates token.
	token1, err := st.VerifyUser("bob", "password123")
	if err != nil {
		t.Fatal(err)
	}
	if token1 == "" {
		t.Fatal("expected token")
	}

	// Second login — should return same token.
	token2, err := st.VerifyUser("bob", "password123")
	if err != nil {
		t.Fatal(err)
	}
	if token2 != token1 {
		t.Errorf("second login returned different token %q != %q", token2, token1)
	}
}

func TestVerifyWrongPassword(t *testing.T) {
	dir := t.TempDir()
	st, err := Open(filepath.Join(dir, "mango.db"), filepath.Join(dir, "library"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	if err := st.NewUser("charlie", "secret123", false); err != nil {
		t.Fatal(err)
	}

	token, err := st.VerifyUser("charlie", "wrongpass")
	if err != nil {
		t.Fatal(err)
	}
	if token != "" {
		t.Error("wrong password should not return a token")
	}
}

func TestVerifyNonexistentUser(t *testing.T) {
	dir := t.TempDir()
	st, err := Open(filepath.Join(dir, "mango.db"), filepath.Join(dir, "library"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	token, err := st.VerifyUser("nonexistent", "password")
	if err != nil {
		t.Fatal(err)
	}
	if token != "" {
		t.Error("nonexistent user should not return a token")
	}
}

func TestDeleteLastAdminRejected(t *testing.T) {
	dir := t.TempDir()
	st, err := Open(filepath.Join(dir, "mango.db"), filepath.Join(dir, "library"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	// There should be exactly 1 admin (auto-created).
	if err := st.DeleteUser("admin"); err == nil {
		t.Fatal("expected error when deleting last admin")
	} else if !strings.Contains(err.Error(), "last admin") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDeleteLastAdminViaUpdateRejected(t *testing.T) {
	dir := t.TempDir()
	st, err := Open(filepath.Join(dir, "mango.db"), filepath.Join(dir, "library"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	// Try to update the last admin to remove admin flag.
	if err := st.UpdateUser("admin", "admin", "", false); err == nil {
		t.Fatal("expected error when removing last admin's admin flag")
	} else if !strings.Contains(err.Error(), "last admin") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		wantErr  bool
	}{
		{"valid", "testuser", false},
		{"with underscore", "test_user", false},
		{"starts with underscore", "_testuser", false},
		{"too short", "ab", true},
		{"starts with number", "1test", true},
		{"special chars", "test@user", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateUsername(tt.username)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateUsername(%q) error = %v, wantErr = %v", tt.username, err, tt.wantErr)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"valid", "secret123", false},
		{"too short", "abc12", true},
		{"non-ascii", "héllo123", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePassword(%q) error = %v, wantErr = %v", tt.password, err, tt.wantErr)
			}
		})
	}
}

func TestNewUserValidation(t *testing.T) {
	dir := t.TempDir()
	st, err := Open(filepath.Join(dir, "mango.db"), filepath.Join(dir, "library"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	if err := st.NewUser("ab", "password123", false); err == nil {
		t.Error("expected error for short username")
	}
	if err := st.NewUser("validuser", "short", false); err == nil {
		t.Error("expected error for short password")
	}
}

func TestAdminIsAdmin(t *testing.T) {
	dir := t.TempDir()
	st, err := Open(filepath.Join(dir, "mango.db"), filepath.Join(dir, "library"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	admin, err := st.UsernameIsAdmin("admin")
	if err != nil {
		t.Fatal(err)
	}
	if !admin {
		t.Error("auto-created admin should have admin flag")
	}
}

// ---------------------------------------------------------------------------
// Thumbnail tests
// ---------------------------------------------------------------------------

func TestEntryDimensionsSaveGetAndSignature(t *testing.T) {
	dir := t.TempDir()
	st, err := Open(filepath.Join(dir, "mango.db"), filepath.Join(dir, "library"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	_, err = st.DB().Exec("INSERT INTO ids (id, path, signature) VALUES (?, ?, ?)",
		"entry-dim-1", "book/vol1.cbz", "111")
	if err != nil {
		t.Fatal(err)
	}

	want := []PageDimension{{Width: 100, Height: 200}, {Width: 0, Height: 0}}
	if err := st.SaveEntryDimensions("entry-dim-1", "111", want); err != nil {
		t.Fatal(err)
	}

	got, ok, err := st.GetEntryDimensions("entry-dim-1", "111")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected cache hit")
	}
	if len(got) != 2 || got[0].Width != 100 || got[0].Height != 200 || got[1].Width != 0 {
		t.Fatalf("got %#v, want %#v", got, want)
	}

	_, ok, err = st.GetEntryDimensions("entry-dim-1", "222")
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("expected miss on signature mismatch")
	}

	_, ok, err = st.GetEntryDimensions("missing", "111")
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("expected miss for missing id")
	}

	// Corrupt JSON → miss
	_, err = st.DB().Exec(
		`UPDATE entry_dimensions SET dimensions = ? WHERE id = ?`,
		"not-json", "entry-dim-1",
	)
	if err != nil {
		t.Fatal(err)
	}
	_, ok, err = st.GetEntryDimensions("entry-dim-1", "111")
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("expected miss for corrupt JSON")
	}

	// UPSERT overwrites
	next := []PageDimension{{Width: 10, Height: 20}}
	if err := st.SaveEntryDimensions("entry-dim-1", "333", next); err != nil {
		t.Fatal(err)
	}
	got, ok, err = st.GetEntryDimensions("entry-dim-1", "333")
	if err != nil || !ok {
		t.Fatalf("upsert hit: ok=%v err=%v", ok, err)
	}
	if len(got) != 1 || got[0].Width != 10 || got[0].Height != 20 {
		t.Fatalf("after upsert got %#v", got)
	}
}

func TestSaveGetThumbnail(t *testing.T) {
	dir := t.TempDir()
	st, err := Open(filepath.Join(dir, "mango.db"), filepath.Join(dir, "library"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	// thumbnails has FK to ids, so insert a matching row.
	_, err = st.DB().Exec("INSERT INTO ids (id, path, signature) VALUES (?, ?, ?)",
		"test-id-123", "some/path", "sig")
	if err != nil {
		t.Fatal(err)
	}

	img := &Image{
		Data:     []byte{0xff, 0xd8, 0xff, 0xe0, 0x00, 0x10},
		Filename: "thumb.jpg",
		Mime:     "image/jpeg",
		Size:     6,
	}

	if err := st.SaveThumbnail("test-id-123", img); err != nil {
		t.Fatal(err)
	}

	got, err := st.GetThumbnail("test-id-123")
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("thumbnail not found")
	}
	if string(got.Data) != string(img.Data) {
		t.Errorf("data mismatch")
	}
	if got.Filename != img.Filename {
		t.Errorf("filename = %q, want %q", got.Filename, img.Filename)
	}
	if got.Mime != img.Mime {
		t.Errorf("mime = %q, want %q", got.Mime, img.Mime)
	}
	if got.Size != img.Size {
		t.Errorf("size = %d, want %d", got.Size, img.Size)
	}
}

func TestLibraryIdentityChecksAreReadOnlyAndPathAware(t *testing.T) {
	dir := t.TempDir()
	libraryDir := filepath.Join(dir, "library")
	st, err := Open(filepath.Join(dir, "mango.db"), libraryDir)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	titlePath := filepath.Join(libraryDir, "Title")
	entryPath := filepath.Join(titlePath, "chapter.cbz")
	titleID, err := st.GetOrCreateTitleID(titlePath, 1)
	if err != nil {
		t.Fatal(err)
	}
	entryID, err := st.GetOrCreateEntryID(entryPath, 2)
	if err != nil {
		t.Fatal(err)
	}

	assertMatch := func(name string, got bool, err error, want bool) {
		t.Helper()
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if got != want {
			t.Fatalf("%s = %v, want %v", name, got, want)
		}
	}

	got, err := st.TitleIdentityMatches(titleID, titlePath)
	assertMatch("matching title", got, err, true)
	got, err = st.TitleIdentityMatches(titleID, filepath.Join(libraryDir, "Other"))
	assertMatch("wrong title path", got, err, false)
	got, err = st.EntryIdentityMatches(entryID, entryPath)
	assertMatch("matching entry", got, err, true)
	got, err = st.EntryIdentityMatches(entryID, filepath.Join(titlePath, "other.cbz"))
	assertMatch("wrong entry path", got, err, false)
	got, err = st.TitleIDExists(titleID)
	assertMatch("existing title ID", got, err, true)

	if _, err := st.DB().Exec("UPDATE ids SET unavailable = 1 WHERE id = ?", entryID); err != nil {
		t.Fatal(err)
	}
	got, err = st.EntryIdentityMatches(entryID, entryPath)
	assertMatch("unavailable entry", got, err, false)

	var titleCount, entryCount int
	if err := st.DB().QueryRow("SELECT COUNT(*) FROM titles").Scan(&titleCount); err != nil {
		t.Fatal(err)
	}
	if err := st.DB().QueryRow("SELECT COUNT(*) FROM ids").Scan(&entryCount); err != nil {
		t.Fatal(err)
	}
	if titleCount != 1 || entryCount != 1 {
		t.Fatalf("identity checks changed row counts: titles=%d entries=%d", titleCount, entryCount)
	}
}

func TestGetNonexistentThumbnail(t *testing.T) {
	dir := t.TempDir()
	st, err := Open(filepath.Join(dir, "mango.db"), filepath.Join(dir, "library"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	got, err := st.GetThumbnail("nonexistent")
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Error("expected nil for nonexistent thumbnail")
	}
}

func TestDeleteThumbnail(t *testing.T) {
	dir := t.TempDir()
	st, err := Open(filepath.Join(dir, "mango.db"), filepath.Join(dir, "library"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	// thumbnails has FK to ids, so insert a matching row.
	_, err = st.DB().Exec("INSERT INTO ids (id, path, signature) VALUES (?, ?, ?)",
		"del-test", "some/path", "sig")
	if err != nil {
		t.Fatal(err)
	}

	img := &Image{Data: []byte{1, 2, 3}, Filename: "img.jpg", Mime: "image/jpeg", Size: 3}
	if err := st.SaveThumbnail("del-test", img); err != nil {
		t.Fatal(err)
	}
	if err := st.DeleteThumbnail("del-test"); err != nil {
		t.Fatal(err)
	}
	got, err := st.GetThumbnail("del-test")
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Error("thumbnail should be deleted")
	}
}

// ---------------------------------------------------------------------------
// Tag tests
// ---------------------------------------------------------------------------

func TestAddGetDeleteTag(t *testing.T) {
	dir := t.TempDir()
	st, err := Open(filepath.Join(dir, "mango.db"), filepath.Join(dir, "library"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	// We need a title to exist for tag FK constraints.
	// Insert directly since we have the schema.
	_, err = st.DB().Exec(
		"INSERT INTO titles (id, path, signature, unavailable) VALUES (?, ?, ?, 0)",
		"title-1", "some/path", "sig1",
	)
	if err != nil {
		t.Fatal(err)
	}

	// Add tags.
	if err := st.AddTag("title-1", "action"); err != nil {
		t.Fatal(err)
	}
	if err := st.AddTag("title-1", "comedy"); err != nil {
		t.Fatal(err)
	}

	// Get tags.
	tags, err := st.GetTitleTags("title-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(tags) != 2 {
		t.Fatalf("got %d tags, want 2", len(tags))
	}
	// Should be ordered by tag.
	if tags[0] != "action" || tags[1] != "comedy" {
		t.Errorf("tags = %v, want [action comedy]", tags)
	}

	// Delete one tag.
	if err := st.DeleteTag("title-1", "action"); err != nil {
		t.Fatal(err)
	}
	tags, _ = st.GetTitleTags("title-1")
	if len(tags) != 1 || tags[0] != "comedy" {
		t.Errorf("after delete, tags = %v, want [comedy]", tags)
	}
}

func TestGetTagTitles(t *testing.T) {
	dir := t.TempDir()
	st, err := Open(filepath.Join(dir, "mango.db"), filepath.Join(dir, "library"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	// Create titles.
	for _, id := range []string{"t1", "t2", "t3"} {
		_, err := st.DB().Exec(
			"INSERT INTO titles (id, path, signature, unavailable) VALUES (?, ?, ?, 0)",
			id, "path/"+id, "sig-"+id,
		)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Tag t1 and t2.
	st.AddTag("t1", "shonen")
	st.AddTag("t2", "shonen")
	st.AddTag("t3", "seinen")

	ids, err := st.GetTagTitles("shonen", false)
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 2 {
		t.Fatalf("got %d shonen titles, want 2", len(ids))
	}
}

func TestListTags(t *testing.T) {
	dir := t.TempDir()
	st, err := Open(filepath.Join(dir, "mango.db"), filepath.Join(dir, "library"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	// Create titles with tags.
	_, err = st.DB().Exec(
		"INSERT INTO titles (id, path, signature, unavailable) VALUES (?, ?, ?, 0)",
		"t1", "path/t1", "sig-t1",
	)
	if err != nil {
		t.Fatal(err)
	}
	st.AddTag("t1", "action")
	st.AddTag("t1", "comedy")

	tags, err := st.ListTags()
	if err != nil {
		t.Fatal(err)
	}
	if len(tags) != 2 {
		t.Fatalf("got %d tags, want 2", len(tags))
	}
}

// ---------------------------------------------------------------------------
// Hidden title tests
// ---------------------------------------------------------------------------

func TestSetGetHidden(t *testing.T) {
	dir := t.TempDir()
	st, err := Open(filepath.Join(dir, "mango.db"), filepath.Join(dir, "library"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	_, err = st.DB().Exec(
		"INSERT INTO titles (id, path, signature, unavailable, hidden) VALUES (?, ?, ?, 0, 0)",
		"h-test", "path/h", "sig-h",
	)
	if err != nil {
		t.Fatal(err)
	}

	hidden, err := st.GetTitleHidden("h-test")
	if err != nil {
		t.Fatal(err)
	}
	if hidden != 0 {
		t.Errorf("initial hidden = %d, want 0", hidden)
	}

	if err := st.SetTitleHidden("h-test", 1); err != nil {
		t.Fatal(err)
	}

	hidden, _ = st.GetTitleHidden("h-test")
	if hidden != 1 {
		t.Errorf("after set hidden = %d, want 1", hidden)
	}

	ids, err := st.GetHiddenTitleIDs()
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 1 || ids[0] != "h-test" {
		t.Errorf("hidden ids = %v, want [h-test]", ids)
	}
}

// ---------------------------------------------------------------------------
// Sort title tests
// ---------------------------------------------------------------------------

func TestTitleSortTitle(t *testing.T) {
	dir := t.TempDir()
	st, err := Open(filepath.Join(dir, "mango.db"), filepath.Join(dir, "library"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	_, err = st.DB().Exec(
		"INSERT INTO titles (id, path, signature, unavailable) VALUES (?, ?, ?, 0)",
		"st-test", "path/st", "sig-st",
	)
	if err != nil {
		t.Fatal(err)
	}

	// Initially nil.
	stVal, err := st.GetTitleSortTitle("st-test")
	if err != nil {
		t.Fatal(err)
	}
	if stVal != nil {
		t.Errorf("initial sort_title = %v, want nil", *stVal)
	}

	// Set sort title.
	sortVal := "Akira"
	if err := st.SetTitleSortTitle("st-test", &sortVal); err != nil {
		t.Fatal(err)
	}
	stVal, _ = st.GetTitleSortTitle("st-test")
	if stVal == nil || *stVal != "Akira" {
		t.Errorf("sort_title = %v, want Akira", stVal)
	}

	// Clear with empty string.
	if err := st.SetTitleSortTitle("st-test", strPtr("")); err != nil {
		t.Fatal(err)
	}
	stVal, _ = st.GetTitleSortTitle("st-test")
	if stVal != nil {
		t.Errorf("after empty string, sort_title = %v, want nil", *stVal)
	}
}

func TestEntrySortTitle(t *testing.T) {
	dir := t.TempDir()
	st, err := Open(filepath.Join(dir, "mango.db"), filepath.Join(dir, "library"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	_, err = st.DB().Exec(
		"INSERT INTO ids (id, path, signature) VALUES (?, ?, ?)",
		"entry-1", "path/e1", "sig-e1",
	)
	if err != nil {
		t.Fatal(err)
	}
	_, err = st.DB().Exec(
		"INSERT INTO ids (id, path, signature) VALUES (?, ?, ?)",
		"entry-2", "path/e2", "sig-e2",
	)
	if err != nil {
		t.Fatal(err)
	}

	sortVal := "Chapter 01"
	if err := st.SetEntrySortTitle("entry-1", &sortVal); err != nil {
		t.Fatal(err)
	}

	stVal, err := st.GetEntrySortTitle("entry-1")
	if err != nil {
		t.Fatal(err)
	}
	if stVal == nil || *stVal != "Chapter 01" {
		t.Errorf("sort_title = %v, want Chapter 01", stVal)
	}

	// Get entries sort title.
	results, err := st.GetEntriesSortTitle([]string{"entry-1", "entry-2"})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}
	if results["entry-1"] == nil || *results["entry-1"] != "Chapter 01" {
		t.Errorf("entry-1 sort = %v", results["entry-1"])
	}
	if results["entry-2"] != nil {
		t.Errorf("entry-2 sort = %v, want nil", results["entry-2"])
	}

	// Empty slice.
	empty, err := st.GetEntriesSortTitle([]string{})
	if err != nil {
		t.Fatal(err)
	}
	if len(empty) != 0 {
		t.Errorf("empty input returned %d results", len(empty))
	}
}

// ---------------------------------------------------------------------------
// Count tests
// ---------------------------------------------------------------------------

func TestCountTitles(t *testing.T) {
	dir := t.TempDir()
	st, err := Open(filepath.Join(dir, "mango.db"), filepath.Join(dir, "library"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	count, err := st.CountTitles()
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Errorf("count = %d, want 0", count)
	}

	for _, id := range []string{"t1", "t2"} {
		_, err := st.DB().Exec(
			"INSERT INTO titles (id, path, signature, unavailable) VALUES (?, ?, ?, 0)",
			id, "path/"+id, "sig-"+id,
		)
		if err != nil {
			t.Fatal(err)
		}
	}

	count, _ = st.CountTitles()
	if count != 2 {
		t.Errorf("count = %d, want 2", count)
	}
}

// ---------------------------------------------------------------------------
// Token test
// ---------------------------------------------------------------------------

func TestRandomStrIsUUIDWithoutDashes(t *testing.T) {
	s := randomStr()
	if len(s) != 32 {
		t.Errorf("randomStr length = %d, want 32", len(s))
	}
	// Should not contain dashes.
	if strings.Contains(s, "-") {
		t.Errorf("randomStr contains dashes: %s", s)
	}
}

func TestHashPasswordAndVerify(t *testing.T) {
	hash, err := hashPassword("testpassword")
	if err != nil {
		t.Fatal(err)
	}
	if !verifyPassword(hash, "testpassword") {
		t.Error("verifyPassword should return true for correct password")
	}
	if verifyPassword(hash, "wrong") {
		t.Error("verifyPassword should return false for wrong password")
	}
}

func TestMissingItemsCRUD(t *testing.T) {
	dir := t.TempDir()
	st, err := Open(filepath.Join(dir, "mango.db"), filepath.Join(dir, "library"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	_, err = st.DB().Exec(
		`INSERT INTO titles (id, path, signature, unavailable) VALUES
			('t-missing', 'gone/title', '1', 1),
			('t-ok', 'still/here', '1', 0)`,
	)
	if err != nil {
		t.Fatal(err)
	}
	_, err = st.DB().Exec(
		`INSERT INTO ids (id, path, signature, unavailable) VALUES
			('e-missing', 'gone/entry.cbz', '1', 1),
			('e-ok', 'still/entry.cbz', '1', 0)`,
	)
	if err != nil {
		t.Fatal(err)
	}
	_, err = st.DB().Exec(`INSERT INTO tags (id, tag) VALUES ('t-missing', 'shonen')`)
	if err != nil {
		t.Fatal(err)
	}

	titles, err := st.ListMissingTitles()
	if err != nil {
		t.Fatal(err)
	}
	if len(titles) != 1 || titles[0].ID != "t-missing" {
		t.Fatalf("missing titles = %+v, want only t-missing", titles)
	}
	entries, err := st.ListMissingEntries()
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].ID != "e-missing" {
		t.Fatalf("missing entries = %+v, want only e-missing", entries)
	}

	if err := st.DeleteMissingTitle("t-missing"); err != nil {
		t.Fatal(err)
	}
	if err := st.DeleteMissingEntry("e-missing"); err != nil {
		t.Fatal(err)
	}
	titles, _ = st.ListMissingTitles()
	entries, _ = st.ListMissingEntries()
	if len(titles) != 0 || len(entries) != 0 {
		t.Fatalf("after single deletes titles=%+v entries=%+v", titles, entries)
	}

	// Bulk delete path.
	_, _ = st.DB().Exec(`INSERT INTO titles (id, path, signature, unavailable) VALUES ('t2', 'x', '1', 1)`)
	_, _ = st.DB().Exec(`INSERT INTO ids (id, path, signature, unavailable) VALUES ('e2', 'y', '1', 1)`)
	if err := st.DeleteAllMissingTitles(); err != nil {
		t.Fatal(err)
	}
	if err := st.DeleteAllMissingEntries(); err != nil {
		t.Fatal(err)
	}
	titles, _ = st.ListMissingTitles()
	entries, _ = st.ListMissingEntries()
	if len(titles) != 0 || len(entries) != 0 {
		t.Fatalf("after bulk deletes titles=%+v entries=%+v", titles, entries)
	}

	// Available rows remain.
	var titleCount, entryCount int
	_ = st.DB().QueryRow(`SELECT COUNT(*) FROM titles WHERE id = 't-ok'`).Scan(&titleCount)
	_ = st.DB().QueryRow(`SELECT COUNT(*) FROM ids WHERE id = 'e-ok'`).Scan(&entryCount)
	if titleCount != 1 || entryCount != 1 {
		t.Fatalf("available rows removed: titles=%d entries=%d", titleCount, entryCount)
	}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func strPtr(s string) *string {
	return &s
}
