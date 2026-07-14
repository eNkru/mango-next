package library

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/eNkru/mango-next/internal/storage"
)

const libraryCacheVersion = 1

// libraryCacheFile is the on-disk JSON (gzip) form of the in-memory title tree.
type libraryCacheFile struct {
	Version     int           `json:"version"`
	LibraryPath string        `json:"library_path"`
	Titles      []cachedTitle `json:"titles"`
}

type cachedTitle struct {
	Dir         string        `json:"dir"`
	ParentID    string        `json:"parent_id"`
	ID          string        `json:"id"`
	Signature   uint64        `json:"signature"`
	ContentsSig string        `json:"contents_sig"`
	Name        string        `json:"name"`
	TitleIDs    []string      `json:"title_ids"`
	MtimeUnix   int64         `json:"mtime_unix"`
	Entries     []cachedEntry `json:"entries"`
}

type cachedEntry struct {
	Kind      string   `json:"kind"` // "archive" | "dir"
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Path      string   `json:"path"`
	Pages     int      `json:"pages"`
	MtimeUnix int64    `json:"mtime_unix"`
	Signature uint64   `json:"signature"`
	Files     []string `json:"files,omitempty"` // dir entries only
}

func titlesToCache(libraryPath string, titles []*Title) libraryCacheFile {
	out := libraryCacheFile{
		Version:     libraryCacheVersion,
		LibraryPath: libraryPath,
		Titles:      make([]cachedTitle, 0, len(titles)),
	}
	for _, t := range titles {
		out.Titles = append(out.Titles, titleToCached(t))
	}
	return out
}

func titleToCached(t *Title) cachedTitle {
	ct := cachedTitle{
		Dir:         t.Dir,
		ParentID:    t.ParentID,
		ID:          t.ID,
		Signature:   t.Signature,
		ContentsSig: t.ContentsSig,
		Name:        t.Name,
		TitleIDs:    append([]string(nil), t.TitleIDs...),
		MtimeUnix:   t.Mtime.UnixNano(),
		Entries:     make([]cachedEntry, 0, len(t.Entries)),
	}
	for _, e := range t.Entries {
		ce := cachedEntry{
			ID:        e.ID(),
			Name:      e.Name(),
			Path:      e.Path(),
			Pages:     e.PageCount(),
			MtimeUnix: e.Mtime().UnixNano(),
			Signature: e.Signature(),
		}
		switch v := e.(type) {
		case *ArchiveEntry:
			ce.Kind = "archive"
		case *DirEntry:
			ce.Kind = "dir"
			ce.Files = append([]string(nil), v.files...)
		default:
			ce.Kind = "archive"
		}
		ct.Entries = append(ct.Entries, ce)
	}
	return ct
}

func titlesFromCache(cf libraryCacheFile) ([]*Title, error) {
	if cf.Version != 0 && cf.Version != libraryCacheVersion {
		return nil, fmt.Errorf("unsupported library cache version %d", cf.Version)
	}
	titles := make([]*Title, 0, len(cf.Titles))
	for i := range cf.Titles {
		t, err := titleFromCached(cf.Titles[i])
		if err != nil {
			return nil, err
		}
		if t.ID == "" {
			continue
		}
		titles = append(titles, t)
	}
	return titles, nil
}

func cacheIdentitiesValid(cf libraryCacheFile, st *storage.Storage) (bool, error) {
	for _, ct := range cf.Titles {
		matches, err := st.TitleIdentityMatches(ct.ID, ct.Dir)
		if err != nil || !matches {
			return matches, err
		}
		for _, id := range ct.TitleIDs {
			exists, err := st.TitleIDExists(id)
			if err != nil || !exists {
				return exists, err
			}
		}
		for _, ce := range ct.Entries {
			matches, err := st.EntryIdentityMatches(ce.ID, ce.Path)
			if err != nil || !matches {
				return matches, err
			}
		}
	}
	return true, nil
}

func titleFromCached(ct cachedTitle) (*Title, error) {
	t := &Title{
		Dir:         ct.Dir,
		ParentID:    ct.ParentID,
		ID:          ct.ID,
		Signature:   ct.Signature,
		ContentsSig: ct.ContentsSig,
		Name:        ct.Name,
		TitleIDs:    append([]string(nil), ct.TitleIDs...),
		Mtime:       time.Unix(0, ct.MtimeUnix),
		Entries:     make([]Entry, 0, len(ct.Entries)),
	}
	for _, ce := range ct.Entries {
		switch ce.Kind {
		case "dir":
			e := &DirEntry{
				id:      ce.ID,
				title:   ce.Name,
				path:    ce.Path,
				book:    t,
				pages:   ce.Pages,
				mtime:   time.Unix(0, ce.MtimeUnix),
				sigHash: ce.Signature,
				files:   append([]string(nil), ce.Files...),
			}
			t.Entries = append(t.Entries, e)
		default: // archive
			e := &ArchiveEntry{
				id:    ce.ID,
				title: ce.Name,
				path:  ce.Path,
				book:  t,
				pages: ce.Pages,
				mtime: time.Unix(0, ce.MtimeUnix),
				sig:   ce.Signature,
			}
			t.Entries = append(t.Entries, e)
		}
	}
	return t, nil
}

func writeLibraryCache(path string, cf libraryCacheFile) error {
	if path == "" {
		return fmt.Errorf("empty library cache path")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(cf)
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	zw := gzip.NewWriter(f)
	if _, err := zw.Write(data); err != nil {
		zw.Close()
		f.Close()
		os.Remove(tmp)
		return err
	}
	if err := zw.Close(); err != nil {
		f.Close()
		os.Remove(tmp)
		return err
	}
	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, path)
}

func readLibraryCache(path string) (libraryCacheFile, error) {
	var cf libraryCacheFile
	if path == "" {
		return cf, fmt.Errorf("empty library cache path")
	}
	f, err := os.Open(path)
	if err != nil {
		return cf, err
	}
	defer f.Close()
	zr, err := gzip.NewReader(f)
	if err != nil {
		return cf, err
	}
	defer zr.Close()
	data, err := io.ReadAll(zr)
	if err != nil {
		return cf, err
	}
	if err := json.Unmarshal(data, &cf); err != nil {
		return cf, err
	}
	return cf, nil
}
