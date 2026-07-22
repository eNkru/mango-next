package storage

import (
	"compress/gzip"
	"io"
	"os"
	"path/filepath"
)

// ---------------------------------------------------------------------------
// Library cache — gzip-compressed JSON (mirrors Crystal's library.yml.gz)
// ---------------------------------------------------------------------------

// SaveLibraryCache writes gzip-compressed data to the cache path.
func (s *Storage) SaveLibraryCache(data []byte) error {
	// Determine cache path: store next to the library dir (matching Crystal's
	// default library_cache_path).
	cachePath := filepath.Join(filepath.Dir(s.libraryPath), "library.yml.gz")

	dir := filepath.Dir(cachePath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	f, err := os.Create(cachePath)
	if err != nil {
		return err
	}
	defer f.Close()

	w := gzip.NewWriter(f)
	defer w.Close()
	_, err = w.Write(data)
	return err
}

// LoadLibraryCache reads and decompresses the gzip cache file.
func (s *Storage) LoadLibraryCache() ([]byte, error) {
	cachePath := filepath.Join(filepath.Dir(s.libraryPath), "library.yml.gz")
	f, err := os.Open(cachePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	return io.ReadAll(r)
}
