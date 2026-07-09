package library

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
)

// testZipEntry represents a file inside a test zip archive.
type testZipEntry struct {
	name    string
	content []byte
}

// testZipWriter builds a minimal ZIP archive in memory and writes it to disk.
type testZipWriter struct {
	entries []testZipEntry
}

func newZipWriter() *testZipWriter {
	return &testZipWriter{}
}

func (zw *testZipWriter) addEntry(name string, content []byte) {
	zw.entries = append(zw.entries, testZipEntry{name: name, content: content})
}

func (zw *testZipWriter) close(path string) error {
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	for _, e := range zw.entries {
		f, err := w.Create(e.name)
		if err != nil {
			return fmt.Errorf("create zip entry %s: %w", e.name, err)
		}
		if _, err := io.Copy(f, bytes.NewReader(e.content)); err != nil {
			return fmt.Errorf("write zip entry %s: %w", e.name, err)
		}
	}
	if err := w.Close(); err != nil {
		return err
	}
	return os.WriteFile(path, buf.Bytes(), 0o644)
}
