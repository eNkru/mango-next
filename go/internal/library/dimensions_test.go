package library

import (
	"bytes"
	"image"
	"image/jpeg"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/eNkru/mango-next/internal/storage"
)

func encodeJPEG(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestReadPageDimensionsArchiveSingleOpen(t *testing.T) {
	dir := t.TempDir()
	zipPath := filepath.Join(dir, "ch.cbz")
	zw := newZipWriter()
	zw.addEntry("002.jpg", encodeJPEG(t, 200, 300))
	zw.addEntry("001.jpg", encodeJPEG(t, 100, 150))
	if err := zw.close(zipPath); err != nil {
		t.Fatal(err)
	}

	dbPath := filepath.Join(dir, "mango.db")
	st, err := storage.Open(dbPath, dir)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	book := &Title{Dir: dir, Name: "t"}
	entry := NewArchiveEntry(zipPath, book, st)
	if entry.Err() != nil {
		t.Fatal(entry.Err())
	}
	if entry.PageCount() != 2 {
		t.Fatalf("pages = %d, want 2", entry.PageCount())
	}

	dims, err := ReadPageDimensions(entry)
	if err != nil {
		t.Fatal(err)
	}
	if len(dims) != 2 {
		t.Fatalf("dims len = %d, want 2", len(dims))
	}
	// Numeric sort: 001 then 002
	if dims[0].Width != 100 || dims[0].Height != 150 {
		t.Errorf("page1 = %+v, want 100x150", dims[0])
	}
	if dims[1].Width != 200 || dims[1].Height != 300 {
		t.Errorf("page2 = %+v, want 200x300", dims[1])
	}
}

func TestReadPageDimensionsDirEntry(t *testing.T) {
	dir := t.TempDir()
	vol := filepath.Join(dir, "Vol1")
	if err := os.MkdirAll(vol, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(vol, "01.jpg"), encodeJPEG(t, 50, 60), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(vol, "02.jpg"), encodeJPEG(t, 70, 80), 0o644); err != nil {
		t.Fatal(err)
	}

	dbPath := filepath.Join(dir, "mango.db")
	st, err := storage.Open(dbPath, dir)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	book := &Title{Dir: dir, Name: "t"}
	entry := NewDirEntry(vol, book, st)
	if entry.Err() != nil {
		t.Fatal(entry.Err())
	}

	dims, err := ReadPageDimensions(entry)
	if err != nil {
		t.Fatal(err)
	}
	if len(dims) != 2 {
		t.Fatalf("dims len = %d", len(dims))
	}
	if dims[0].Width != 50 || dims[0].Height != 60 {
		t.Errorf("page1 = %+v", dims[0])
	}
	if dims[1].Width != 70 || dims[1].Height != 80 {
		t.Errorf("page2 = %+v", dims[1])
	}
}

func TestEntryDimensionsCacheRoundTripWithLibrary(t *testing.T) {
	dir := t.TempDir()
	zipPath := filepath.Join(dir, "ch.cbz")
	zw := newZipWriter()
	zw.addEntry("a.jpg", encodeJPEG(t, 10, 20))
	if err := zw.close(zipPath); err != nil {
		t.Fatal(err)
	}

	st, err := storage.Open(filepath.Join(dir, "mango.db"), dir)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	entry := NewArchiveEntry(zipPath, &Title{Dir: dir}, st)
	if entry.Err() != nil {
		t.Fatal(entry.Err())
	}

	dims, err := ReadPageDimensions(entry)
	if err != nil {
		t.Fatal(err)
	}
	stored := make([]storage.PageDimension, len(dims))
	for i, d := range dims {
		stored[i] = storage.PageDimension{Width: d.Width, Height: d.Height}
	}
	sig := strconv.FormatUint(entry.Signature(), 10)
	if err := st.SaveEntryDimensions(entry.ID(), sig, stored); err != nil {
		t.Fatal(err)
	}
	got, ok, err := st.GetEntryDimensions(entry.ID(), sig)
	if err != nil || !ok {
		t.Fatalf("cache: ok=%v err=%v", ok, err)
	}
	if got[0].Width != 10 || got[0].Height != 20 {
		t.Fatalf("got %+v", got[0])
	}
}
