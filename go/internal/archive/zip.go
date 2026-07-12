package archive

import (
	"archive/zip"
	"fmt"
	"io"
	"path/filepath"
)

type zipArchive struct{ path string }

func openZip(path string) (Reader, error) {
	return &zipArchive{path: path}, nil
}

func (z *zipArchive) Entries() ([]Entry, error) {
	r, err := zip.OpenReader(z.path)
	if err != nil {
		return nil, fmt.Errorf("open zip %s: %w", z.path, err)
	}
	defer r.Close()

	var entries []Entry
	for _, f := range r.File {
		if !f.FileInfo().IsDir() {
			entries = append(entries, Entry{Name: f.Name})
		}
	}
	if entries == nil {
		entries = []Entry{}
	}
	return entries, nil
}

func (z *zipArchive) ReadEntry(entry Entry) ([]byte, error) {
	r, err := zip.OpenReader(z.path)
	if err != nil {
		return nil, fmt.Errorf("open zip %s: %w", z.path, err)
	}
	defer r.Close()

	for _, f := range r.File {
		if filepath.ToSlash(f.Name) == filepath.ToSlash(entry.Name) {
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			defer rc.Close()
			return io.ReadAll(rc)
		}
	}
	return nil, fmt.Errorf("entry %q not found in %s", entry.Name, z.path)
}

func (z *zipArchive) Close() error { return nil }

var _ Reader = (*zipArchive)(nil)
