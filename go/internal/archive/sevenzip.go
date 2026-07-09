package archive

import (
	"fmt"
	"io"

	"github.com/bodgit/sevenzip"
)

type szipArchive struct {
	path string
}

func openSevenZip(path string) (Reader, error) {
	return &szipArchive{path: path}, nil
}

func (s *szipArchive) Entries() ([]Entry, error) {
	r, err := sevenzip.OpenReader(s.path)
	if err != nil {
		return nil, fmt.Errorf("open 7z %s: %w", s.path, err)
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

func (s *szipArchive) ReadEntry(entry Entry) ([]byte, error) {
	r, err := sevenzip.OpenReader(s.path)
	if err != nil {
		return nil, fmt.Errorf("open 7z %s: %w", s.path, err)
	}
	defer r.Close()

	for _, f := range r.File {
		if f.Name == entry.Name {
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			defer rc.Close()
			return io.ReadAll(rc)
		}
	}
	return nil, fmt.Errorf("entry %q not found in %s", entry.Name, s.path)
}

func (s *szipArchive) Close() error { return nil }

var _ Reader = (*szipArchive)(nil)
