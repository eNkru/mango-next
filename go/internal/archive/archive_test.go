package archive

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func createZip(t testing.TB, dir string, files map[string]string) string {
	t.Helper()
	path := filepath.Join(dir, "test.cbz")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	w := zip.NewWriter(f)
	for name, content := range files {
		zw, err := w.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := zw.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestOpenZip(t *testing.T) {
	dir := t.TempDir()
	path := createZip(t, dir, map[string]string{
		"page001.jpg":   "image1 data",
		"page002.jpg":   "image2 data",
		"dir/inner.txt": "nested",
	})

	arc, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer arc.Close()

	entries, err := arc.Entries()
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 3 {
		t.Errorf("got %d entries, want 3", len(entries))
	}

	data, err := arc.ReadEntry(Entry{Name: "page001.jpg"})
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "image1 data" {
		t.Errorf("got %q, want %q", string(data), "image1 data")
	}
}

func TestOpenZipByExtension(t *testing.T) {
	dir := t.TempDir()
	path := createZip(t, dir, map[string]string{"page.jpg": "data"})

	extPath := path[:len(path)-4] + "zip"
	if err := os.Rename(path, extPath); err != nil {
		t.Fatal(err)
	}

	arc, err := Open(extPath)
	if err != nil {
		t.Fatal(err)
	}
	defer arc.Close()

	entries, err := arc.Entries()
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Errorf("got %d entries, want 1", len(entries))
	}
}

func TestOpenNonExistent(t *testing.T) {
	arc, err := Open("/nonexistent/file.cbz")
	if err != nil {
		t.Fatal(err)
	}
	defer arc.Close()
	_, err = arc.Entries()
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestIsArchive(t *testing.T) {
	cases := []struct {
		name string
		want bool
	}{
		{"file.zip", true},
		{"file.cbz", true},
		{"file.rar", true},
		{"file.cbr", true},
		{"file.7z", true},
		{"file.txt", false},
		{"file.jpg", false},
		{"file.ZIP", true},
		{"file.CBZ", true},
	}
	for _, tc := range cases {
		if got := IsArchive(tc.name); got != tc.want {
			t.Errorf("IsArchive(%q) = %v, want %v", tc.name, got, tc.want)
		}
	}
}

func TestReadEntryNotFound(t *testing.T) {
	dir := t.TempDir()
	path := createZip(t, dir, map[string]string{"page.jpg": "data"})

	arc, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer arc.Close()

	_, err = arc.ReadEntry(Entry{Name: "nonexistent.jpg"})
	if err == nil {
		t.Error("expected error for missing entry")
	}
}

func TestEmptyZip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "empty.cbz")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	w := zip.NewWriter(f)
	w.Close()
	f.Close()

	arc, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer arc.Close()

	entries, err := arc.Entries()
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Errorf("got %d entries, want 0", len(entries))
	}
}

func TestSevenZipOpenFailsGracefully(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.7z")
	if err := os.WriteFile(path, []byte("not a 7z file"), 0o644); err != nil {
		t.Fatal(err)
	}
	arc, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer arc.Close()
	_, err = arc.Entries()
	if err == nil {
		t.Error("expected error for invalid 7z file")
	}
}

func TestRarOpenFailsGracefully(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.cbr")
	if err := os.WriteFile(path, []byte("not a rar file"), 0o644); err != nil {
		t.Fatal(err)
	}
	arc, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer arc.Close()
	_, err = arc.Entries()
	if err == nil {
		t.Error("expected error for invalid rar file")
	}
}

func BenchmarkOpenZip(b *testing.B) {
	dir := b.TempDir()
	files := make(map[string]string)
	for i := range 100 {
		files[fmt.Sprintf("pages/page%03d.jpg", i)] = fmt.Sprintf("image data for page %d", i)
	}
	path := createZip(b, dir, files)

	b.ResetTimer()
	for range b.N {
		arc, err := Open(path)
		if err != nil {
			b.Fatal(err)
		}
		entries, err := arc.Entries()
		if err != nil {
			b.Fatal(err)
		}
		if len(entries) != 100 {
			b.Fatalf("got %d entries, want 100", len(entries))
		}
		arc.Close()
	}
}
