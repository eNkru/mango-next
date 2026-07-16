package server

import (
	"archive/zip"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/eNkru/mango-next/internal/library"
	"github.com/eNkru/mango-next/internal/storage"
)

func TestFirstEntryIDNestedVolume(t *testing.T) {
	libDir := t.TempDir()
	partDir := filepath.Join(libDir, "JOJO Series", "Part 7")
	if err := os.MkdirAll(partDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := writeTestCBZ(filepath.Join(partDir, "Vol.15.cbz"), 2); err != nil {
		t.Fatal(err)
	}

	st, err := storage.Open(filepath.Join(t.TempDir(), "mango.db"), libDir)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	lib := library.NewLibrary(libDir, st, "")
	if _, err := lib.Scan(); err != nil {
		t.Fatal(err)
	}

	lib.RLock()
	defer lib.RUnlock()
	if len(lib.TitleIDs) != 1 {
		t.Fatalf("TitleIDs = %d, want 1", len(lib.TitleIDs))
	}
	series := lib.TitleHash[lib.TitleIDs[0]]
	if series == nil {
		t.Fatal("series missing")
	}

	eid := firstEntryID(series)
	if eid == "" {
		t.Fatal("firstEntryID empty for nested volume tree")
	}
	if len(series.TitleIDs) > 0 && eid == series.TitleIDs[0] {
		t.Fatalf("firstEntryID returned sub-title id %q", eid)
	}
	found := false
	for _, e := range series.DeepEntries() {
		if e.ID() == eid {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("firstEntryID %q not in DeepEntries", eid)
	}
}

func writeTestCBZ(path string, pages int) error {
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	for i := 1; i <= pages; i++ {
		name := fmt.Sprintf("page%03d.jpg", i)
		f, err := w.Create(name)
		if err != nil {
			return err
		}
		if _, err := f.Write([]byte(fmt.Sprintf("fake-%d", i))); err != nil {
			return err
		}
	}
	if err := w.Close(); err != nil {
		return err
	}
	return os.WriteFile(path, buf.Bytes(), 0o644)
}
