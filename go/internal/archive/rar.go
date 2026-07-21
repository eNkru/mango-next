package archive

import (
	"fmt"
	"io"
	"os"

	"github.com/nwaples/rardecode/v2"
)

type rarArchive struct {
	path string
}

func openRar(path string) (Reader, error) {
	return &rarArchive{path: path}, nil
}

func (r *rarArchive) Entries() ([]Entry, error) {
	f, err := os.Open(r.path)
	if err != nil {
		return nil, fmt.Errorf("open rar %s: %w", r.path, err)
	}
	defer f.Close()

	// v2 bounds dictionary size (mitigates DoS via huge RAR dictionaries).
	rr, err := rardecode.NewReader(f)
	if err != nil {
		return nil, fmt.Errorf("init rar reader %s: %w", r.path, err)
	}

	var entries []Entry
	for {
		h, err := rr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read rar entry %s: %w", r.path, err)
		}
		if !h.IsDir {
			entries = append(entries, Entry{Name: h.Name})
		}
	}
	if entries == nil {
		entries = []Entry{}
	}
	return entries, nil
}

func (r *rarArchive) ReadEntry(entry Entry) ([]byte, error) {
	f, err := os.Open(r.path)
	if err != nil {
		return nil, fmt.Errorf("open rar %s: %w", r.path, err)
	}
	defer f.Close()

	rr, err := rardecode.NewReader(f)
	if err != nil {
		return nil, fmt.Errorf("init rar reader %s: %w", r.path, err)
	}

	for {
		h, err := rr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read rar entry %s: %w", r.path, err)
		}
		if h.Name == entry.Name {
			return io.ReadAll(rr)
		}
	}
	return nil, fmt.Errorf("entry %q not found in %s", entry.Name, r.path)
}

func (r *rarArchive) Close() error { return nil }

var _ Reader = (*rarArchive)(nil)
