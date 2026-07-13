package library

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/hkalexling/mango-go/internal/storage"
)

// setupTestLibrary creates a synthetic manga library in a temp directory.
func setupTestLibrary(t *testing.T) (libDir string) {
	t.Helper()

	libDir = t.TempDir()

	// Manga Title 1 — two archive entries
	title1Dir := filepath.Join(libDir, "Manga Title 1")
	if err := os.MkdirAll(title1Dir, 0o755); err != nil {
		t.Fatal(err)
	}
	createFakeCBZ(t, filepath.Join(title1Dir, "ch01.cbz"), 5)
	createFakeCBZ(t, filepath.Join(title1Dir, "ch02.cbz"), 8)

	// Manga Title 2 — directory entry with loose images in a sub-dir
	title2Dir := filepath.Join(libDir, "Manga Title 2")
	if err := os.MkdirAll(title2Dir, 0o755); err != nil {
		t.Fatal(err)
	}
	vol1Dir := filepath.Join(title2Dir, "Vol. 1")
	if err := os.MkdirAll(vol1Dir, 0o755); err != nil {
		t.Fatal(err)
	}
	createFakeImage(t, filepath.Join(vol1Dir, "page001.jpg"))
	createFakeImage(t, filepath.Join(vol1Dir, "page002.jpg"))
	createFakeImage(t, filepath.Join(vol1Dir, "page003.jpg"))

	// Nested Series — sub-title + direct archive entry
	nsDir := filepath.Join(libDir, "Nested Series")
	if err := os.MkdirAll(nsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	nsVol1Dir := filepath.Join(nsDir, "Vol. 1")
	if err := os.MkdirAll(nsVol1Dir, 0o755); err != nil {
		t.Fatal(err)
	}
	createFakeImage(t, filepath.Join(nsVol1Dir, "001.png"))
	createFakeImage(t, filepath.Join(nsVol1Dir, "002.png"))
	createFakeCBZ(t, filepath.Join(nsDir, "ch01.cbz"), 3)

	return libDir
}

// createFakeCBZ creates a minimal valid ZIP archive with count dummy image entries.
func createFakeCBZ(t *testing.T, path string, count int) {
	t.Helper()
	zw := newZipWriter()
	for i := 1; i <= count; i++ {
		name := sprintf("page%03d.jpg", i)
		content := []byte(sprintf("fake-image-data-%d", i))
		zw.addEntry(name, content)
	}
	if err := zw.close(path); err != nil {
		t.Fatal(err)
	}
}

// sprintf is a helper to avoid importing fmt in test helpers.
func sprintf(format string, args ...any) string {
	// Use a minimal implementation to avoid fmt dependency in helper.
	// We know the formats we'll use.
	result := []byte(format)
	argIdx := 0
	for i := 0; i < len(result); i++ {
		if result[i] == '%' && i+1 < len(result) {
			switch result[i+1] {
			case 'd':
				if argIdx < len(args) {
					if n, ok := args[argIdx].(int); ok {
						digits := itoa(n)
						result = append(result[:i], append([]byte(digits), result[i+2:]...)...)
						i += len(digits) - 1
					}
				}
				argIdx++
			case 's':
				if argIdx < len(args) {
					if s, ok := args[argIdx].(string); ok {
						result = append(result[:i], append([]byte(s), result[i+2:]...)...)
						i += len(s) - 1
					}
				}
				argIdx++
			}
		}
	}
	return string(result)
}

// itoa converts an int to a decimal string without importing strconv.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	var digits []byte
	for n > 0 {
		digits = append(digits, byte('0'+n%10))
		n /= 10
	}
	if neg {
		digits = append(digits, '-')
	}
	// Reverse
	for i, j := 0, len(digits)-1; i < j; i, j = i+1, j-1 {
		digits[i], digits[j] = digits[j], digits[i]
	}
	return string(digits)
}

// createFakeImage creates a minimal JPEG file (1x1 white pixel).
func createFakeImage(t *testing.T, path string) {
	t.Helper()
	// Minimal valid JPEG data
	data := []byte{
		0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00,
		0x01, 0x01, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00,
		0xFF, 0xDB, 0x00, 0x43, 0x00, 0x08, 0x06, 0x06, 0x07, 0x06, 0x05,
		0x08, 0x07, 0x07, 0x07, 0x09, 0x09, 0x08, 0x0A, 0x0C, 0x14, 0x0D,
		0x0C, 0x0B, 0x0B, 0x0C, 0x19, 0x12, 0x13, 0x0F, 0x14, 0x1D, 0x1A,
		0x1F, 0x1E, 0x1D, 0x1A, 0x1C, 0x1C, 0x20, 0x24, 0x2E, 0x27, 0x20,
		0x22, 0x2C, 0x23, 0x1C, 0x1C, 0x28, 0x37, 0x29, 0x2C, 0x30, 0x31,
		0x34, 0x34, 0x34, 0x1F, 0x27, 0x39, 0x3D, 0x38, 0x32, 0x3C, 0x2E,
		0x33, 0x34, 0x32,
		0xFF, 0xC0, 0x00, 0x0B, 0x08, 0x00, 0x01, 0x00, 0x01, 0x01, 0x01,
		0x11, 0x00,
		0xFF, 0xC4, 0x00, 0x1F, 0x00, 0x00, 0x01, 0x05, 0x01, 0x01, 0x01,
		0x01, 0x01, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
		0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B,
		0xFF, 0xDA, 0x00, 0x08, 0x01, 0x01, 0x00, 0x00, 0x3F, 0x00, 0x7B,
		0x94, 0x11, 0x00,
		0xFF, 0xD9,
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestScanFreshLibrary tests a full scan of a synthetic library.
func TestScanFreshLibrary(t *testing.T) {
	libDir := setupTestLibrary(t)

	dbDir := t.TempDir()
	dbPath := filepath.Join(dbDir, "mango.db")

	st, err := storage.Open(dbPath, libDir)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	// Scan
	result, err := ScanLibrary(libDir, st)
	if err != nil {
		t.Fatal(err)
	}

	// Verify counts
	if result.TitleCount != 3 {
		t.Errorf("title count = %d, want 3", result.TitleCount)
	}

	// Verify each title
	findTitle := func(name string) *Title {
		for _, t := range result.Titles {
			if t.Name == name {
				return t
			}
		}
		return nil
	}

	// Manga Title 1 — 2 archive entries
	t1 := findTitle("Manga Title 1")
	if t1 == nil {
		t.Fatal("Manga Title 1 not found")
	}
	if len(t1.Entries) != 2 {
		t.Errorf("Manga Title 1 entries = %d, want 2", len(t1.Entries))
	}
	if t1.ID == "" {
		t.Error("Manga Title 1 ID is empty")
	}

	// Check entries are sorted (ch01 before ch02)
	if len(t1.Entries) >= 2 {
		if t1.Entries[0].Name() != "ch01" {
			t.Errorf("first entry name = %q, want ch01", t1.Entries[0].Name())
		}
		if t1.Entries[1].Name() != "ch02" {
			t.Errorf("second entry name = %q, want ch02", t1.Entries[1].Name())
		}
	}

	// Check archive entry page counts
	ae1 := t1.Entries[0]
	if ae1.PageCount() != 5 {
		t.Errorf("ch01 page count = %d, want 5", ae1.PageCount())
	}
	ae2 := t1.Entries[1]
	if ae2.PageCount() != 8 {
		t.Errorf("ch02 page count = %d, want 8", ae2.PageCount())
	}

	// Read a page from the archive
	img, err := ae1.ReadPage(1)
	if err != nil {
		t.Fatalf("ReadPage(1): %v", err)
	}
	if img == nil {
		t.Fatal("ReadPage returned nil")
	}
	if len(img.Data) == 0 {
		t.Error("image data is empty")
	}

	// Manga Title 2 — 1 dir entry (Vol. 1 with 3 images)
	t2 := findTitle("Manga Title 2")
	if t2 == nil {
		t.Fatal("Manga Title 2 not found")
	}
	if len(t2.Entries) != 1 {
		t.Errorf("Manga Title 2 entries = %d, want 1 (dir entry Vol. 1)", len(t2.Entries))
	} else {
		de := t2.Entries[0]
		if de.Name() != "Vol. 1" {
			t.Errorf("dir entry name = %q, want Vol. 1", de.Name())
		}
		if de.PageCount() != 3 {
			t.Errorf("dir entry page count = %d, want 3", de.PageCount())
		}
		// Read a page
		img, err := de.ReadPage(1)
		if err != nil {
			t.Fatalf("DirEntry ReadPage(1): %v", err)
		}
		if img == nil {
			t.Fatal("DirEntry ReadPage returned nil")
		}
	}
	// Manga Title 2 should have 0 sub-titles (Vol. 1 is a dir entry, not a sub-title)
	if len(t2.TitleIDs) != 0 {
		t.Errorf("Manga Title 2 title IDs = %d, want 0", len(t2.TitleIDs))
	}

	// Nested Series — should have 1 dir entry (Vol. 1) + 1 archive entry (ch01.cbz)
	ns := findTitle("Nested Series")
	if ns == nil {
		t.Fatal("Nested Series not found")
	}
	if len(ns.TitleIDs) != 0 {
		t.Errorf("Nested Series title IDs = %d, want 0 (Vol. 1 is a dir entry)", len(ns.TitleIDs))
	}
	if len(ns.Entries) != 2 {
		t.Errorf("Nested Series entries = %d, want 2 (Vol. 1 dir entry + ch01.cbz)", len(ns.Entries))
	}

	// Verify DB persistence
	dbTitles, err := st.GetAllTitles()
	if err != nil {
		t.Fatal(err)
	}
	// 3 top-level titles
	if len(dbTitles) < 3 {
		t.Errorf("DB titles = %d, want at least 3", len(dbTitles))
	}

	dbEntries, err := st.GetAllEntries()
	if err != nil {
		t.Fatal(err)
	}
	// 2 (Manga Title 1) + 1 (Manga Title 2) + 2 (Nested Series) = 5 entries
	if len(dbEntries) < 5 {
		t.Errorf("DB entries = %d, want at least 5", len(dbEntries))
	}
}

// TestScanEmptyLibrary tests scanning an empty library.
func TestScanEmptyLibrary(t *testing.T) {
	libDir := t.TempDir()
	dbDir := t.TempDir()
	dbPath := filepath.Join(dbDir, "mango.db")

	st, err := storage.Open(dbPath, libDir)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	result, err := ScanLibrary(libDir, st)
	if err != nil {
		t.Fatal(err)
	}
	if result.TitleCount != 0 {
		t.Errorf("title count = %d, want 0", result.TitleCount)
	}
}

// TestScanNoLibraryDir tests that scanning creates the library dir if missing.
func TestScanNoLibraryDir(t *testing.T) {
	dbDir := t.TempDir()
	dbPath := filepath.Join(dbDir, "mango.db")
	libDir := filepath.Join(dbDir, "nonexistent-library")

	st, err := storage.Open(dbPath, libDir)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	result, err := ScanLibrary(libDir, st)
	if err != nil {
		t.Fatal(err)
	}
	if result.TitleCount != 0 {
		t.Errorf("title count = %d, want 0", result.TitleCount)
	}
	if _, err := os.Stat(libDir); os.IsNotExist(err) {
		t.Error("library directory was not created")
	}
}

// TestLibraryScan verifies the Library struct's Scan method.
func TestLibraryScan(t *testing.T) {
	libDir := setupTestLibrary(t)
	dbDir := t.TempDir()
	dbPath := filepath.Join(dbDir, "mango.db")

	st, err := storage.Open(dbPath, libDir)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	lib := NewLibrary(libDir, st)
	result, err := lib.Scan()
	if err != nil {
		t.Fatal(err)
	}
	if result.TitleCount != 3 {
		t.Errorf("title count = %d, want 3", result.TitleCount)
	}

	// Verify hash map
	if len(lib.TitleHash) != 3 {
		t.Errorf("TitleHash size = %d, want 3", len(lib.TitleHash))
	}
	if len(lib.TitleIDs) != 3 {
		t.Errorf("TitleIDs count = %d, want 3", len(lib.TitleIDs))
	}

	// Verify each title is in the hash
	for _, tt := range result.Titles {
		if lib.TitleHash[tt.ID] != tt {
			t.Errorf("Title %q not found in TitleHash", tt.Name)
		}
	}
}

// TestScanIdempotent verifies that scanning twice gives the same result.
func TestScanIdempotent(t *testing.T) {
	libDir := t.TempDir()
	dbDir := t.TempDir()
	dbPath := filepath.Join(dbDir, "mango.db")

	// Create a simple library
	titleDir := filepath.Join(libDir, "Some Manga")
	if err := os.MkdirAll(titleDir, 0o755); err != nil {
		t.Fatal(err)
	}
	createFakeCBZ(t, filepath.Join(titleDir, "ch01.cbz"), 3)
	createFakeCBZ(t, filepath.Join(titleDir, "ch02.cbz"), 4)

	st, err := storage.Open(dbPath, libDir)
	if err != nil {
		t.Fatal(err)
	}

	result1, err := ScanLibrary(libDir, st)
	if err != nil {
		st.Close()
		t.Fatal(err)
	}

	result2, err := ScanLibrary(libDir, st)
	if err != nil {
		st.Close()
		t.Fatal(err)
	}
	st.Close()

	if result1.TitleCount != result2.TitleCount {
		t.Errorf("title count changed: %d -> %d", result1.TitleCount, result2.TitleCount)
	}
	if result1.EntryCount != result2.EntryCount {
		t.Errorf("entry count changed: %d -> %d", result1.EntryCount, result2.EntryCount)
	}
}

// TestCompareNumerically tests the numeric sort.
func TestCompareNumerically(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"ch01", "ch02", -1},
		{"ch02", "ch01", 1},
		{"ch01", "ch01", 0},
		{"ch1", "ch10", -1},
		{"ch10", "ch2", 1},
		{"page001.jpg", "page002.jpg", -1},
		{"page002.jpg", "page001.jpg", 1},
		{"page1.jpg", "page10.jpg", -1},
		{"page10.jpg", "page2.jpg", 1},
		{"abc", "def", -1},
		{"", "a", -1},
		{"a", "", 1},
	}
	for _, tt := range tests {
		t.Run(tt.a+"_"+tt.b, func(t *testing.T) {
			got := compareNumerically(tt.a, tt.b)
			if (got < 0 && tt.want >= 0) || (got > 0 && tt.want <= 0) || (got == 0 && tt.want != 0) {
				t.Errorf("compareNumerically(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

// TestMimeFromFilename verifies MIME type detection.
func TestMimeFromFilename(t *testing.T) {
	tests := []struct {
		filename string
		want     string
	}{
		{"image.jpg", "image/jpeg"},
		{"image.jpeg", "image/jpeg"},
		{"image.png", "image/png"},
		{"image.webp", "image/webp"},
		{"image.gif", "image/gif"},
		{"image.avif", "image/avif"},
		{"image.apng", "image/apng"},
		{"image.svg", "image/svg+xml"},
		{"image.jxl", "image/jxl"},
		{"archive.zip", "application/zip"},
		{"archive.rar", "application/x-rar-compressed"},
		{"archive.cbz", "application/vnd.comicbook+zip"},
		{"archive.cbr", "application/vnd.comicbook-rar"},
		{"unknown.xyz", "application/octet-stream"},
	}
	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got := mimeFromFilename(tt.filename)
			if got != tt.want {
				t.Errorf("mimeFromFilename(%q) = %q, want %q", tt.filename, got, tt.want)
			}
		})
	}
}

// TestSplitByAlphaNumeric verifies the string split function.
func TestSplitByAlphaNumeric(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"ch01", []string{"ch", "01"}},
		{"page001.jpg", []string{"page", "001", ".jpg"}},
		{"Chapter 10 v2", []string{"Chapter ", "10", " v", "2"}},
		{"abc", []string{"abc"}},
		{"123", []string{"123"}},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := splitByAlphaNumeric(tt.input)
			if len(got) != len(tt.want) {
				t.Errorf("splitByAlphaNumeric(%q) = %v (len=%d), want %v (len=%d)",
					tt.input, got, len(got), tt.want, len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("splitByAlphaNumeric(%q)[%d] = %q, want %q",
						tt.input, i, got[i], tt.want[i])
				}
			}
		})
	}
}

// TestScanDoesNotBlockReaders ensures Scan's disk work does not hold mu for the
// whole walk: concurrent RLock readers keep making progress while Scan runs.
func TestScanDoesNotBlockReaders(t *testing.T) {
	libDir := setupTestLibrary(t)
	dbPath := filepath.Join(t.TempDir(), "mango.db")
	st, err := storage.Open(dbPath, libDir)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	lib := NewLibrary(libDir, st)

	// Seed a previous tree so RLock has something to observe mid-scan.
	if _, err := lib.Scan(); err != nil {
		t.Fatal(err)
	}

	doneScan := make(chan error, 1)
	go func() {
		_, err := lib.Scan()
		doneScan <- err
	}()

	// While scan may be running, readers must not stall for the full walk.
	deadline := time.After(2 * time.Second)
	reads := 0
	for reads < 50 {
		select {
		case err := <-doneScan:
			if err != nil {
				t.Fatalf("scan: %v", err)
			}
			// Scan finished early; still require some successful reads.
			if reads == 0 {
				// One more read after scan to ensure API still works.
				lib.RLock()
				_ = len(lib.TitleIDs)
				lib.RUnlock()
			}
			goto after
		case <-deadline:
			t.Fatalf("timed out after %d successful RLock reads; scan may be holding mu too long", reads)
		default:
			lib.RLock()
			_ = len(lib.TitleIDs)
			lib.RUnlock()
			reads++
			time.Sleep(time.Millisecond)
		}
	}
	if err := <-doneScan; err != nil {
		t.Fatalf("scan: %v", err)
	}
after:
	lib.RLock()
	n := len(lib.TitleIDs)
	lib.RUnlock()
	if n != 3 {
		t.Errorf("TitleIDs after scan = %d, want 3", n)
	}
}
