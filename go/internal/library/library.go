package library

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/hkalexling/mango-go/internal/storage"
	"github.com/hkalexling/mango-go/internal/thumbnail"
)

// Library mirrors Crystal's Library class. It manages the title tree and
// provides scan / thumbnail generation / background-job entry points.
type Library struct {
	Dir       string
	TitleIDs  []string
	TitleHash map[string]*Title
	St        *storage.Storage
	CachePath string // gzip JSON cache; empty disables load/save

	mu           sync.RWMutex
	scanMu       sync.Mutex // serializes Scan() so only one disk walk runs
	thumbnailCtx ThumbnailContext
}

// ThumbnailContext tracks thumbnail generation progress (matching Crystal).
type ThumbnailContext struct {
	Current int
	Total   int
	mu      sync.Mutex
}

func (tc *ThumbnailContext) Progress() float64 {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	if tc.Total == 0 {
		return 0
	}
	return float64(tc.Current) / float64(tc.Total)
}

func (tc *ThumbnailContext) Reset() {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.Current = 0
	tc.Total = 0
}

func (tc *ThumbnailContext) Increment() {
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.Current++
}

func (lib *Library) ThumbnailProgress() float64 {
	return lib.thumbnailCtx.Progress()
}

func (lib *Library) Lock()    { lib.mu.Lock() }
func (lib *Library) Unlock()  { lib.mu.Unlock() }
func (lib *Library) RLock()   { lib.mu.RLock() }
func (lib *Library) RUnlock() { lib.mu.RUnlock() }

// NewLibrary creates a Library bound to storage. cachePath is the gzip JSON
// library cache (config library_cache_path); empty string disables cache I/O.
func NewLibrary(libraryPath string, st *storage.Storage, cachePath string) *Library {
	return &Library{
		Dir:       libraryPath,
		TitleIDs:  make([]string, 0),
		TitleHash: make(map[string]*Title),
		St:        st,
		CachePath: cachePath,
	}
}

// LoadFromCache loads a previously saved library tree so the UI can show books
// before the background scan finishes. Invalid/missing cache is a no-op.
func (lib *Library) LoadFromCache() error {
	if lib.CachePath == "" {
		return nil
	}
	cf, err := readLibraryCache(lib.CachePath)
	if err != nil {
		log.Printf("Library cache corrupt or missing (%v); removing and continuing", err)
		if rmErr := os.Remove(lib.CachePath); rmErr != nil && !os.IsNotExist(rmErr) {
			log.Printf("Failed to remove corrupt cache: %v", rmErr)
		}
		return nil
	}
	want, err := filepath.Abs(lib.Dir)
	if err != nil {
		want = lib.Dir
	}
	got, err := filepath.Abs(cf.LibraryPath)
	if err != nil {
		got = cf.LibraryPath
	}
	if got != "" && want != "" && filepath.Clean(got) != filepath.Clean(want) {
		return fmt.Errorf("library cache path mismatch: cache=%q library=%q", cf.LibraryPath, lib.Dir)
	}
	titles, err := titlesFromCache(cf)
	if err != nil {
		return err
	}
	lib.applyTitles(titles)
	log.Printf("Loaded library cache: %d titles from %s", len(titles), lib.CachePath)
	return nil
}

// Scan performs an incremental library scan without holding the RWMutex during
// disk work. Unchanged top-level titles (matching DirSignature) are reused from
// the previous in-memory tree; only new/changed titles call NewTitle.
func (lib *Library) Scan() (*ScanResult, error) {
	lib.scanMu.Lock()
	defer lib.scanMu.Unlock()

	previous := lib.snapshotByDir()

	result, err := ScanLibrary(lib.Dir, lib.St, previous)
	if err != nil {
		return nil, fmt.Errorf("scan library: %w", err)
	}

	lib.applyTitles(result.Titles)

	if err := lib.saveCacheLocked(result.Titles); err != nil {
		log.Printf("Failed to save library cache: %v", err)
	}

	log.Printf("Scanned %d titles (reused %d, rebuilt %d)", len(result.Titles), result.Reused, result.Rebuilt)
	return result, nil
}

func (lib *Library) snapshotByDir() map[string]*Title {
	lib.mu.RLock()
	defer lib.mu.RUnlock()
	out := make(map[string]*Title, len(lib.TitleHash))
	for _, t := range lib.TitleHash {
		if t != nil && t.Dir != "" {
			out[t.Dir] = t
		}
	}
	return out
}

func (lib *Library) applyTitles(titles []*Title) {
	ids := make([]string, 0, len(titles))
	hash := make(map[string]*Title, len(titles))
	for _, t := range titles {
		ids = append(ids, t.ID)
		hash[t.ID] = t
	}
	lib.mu.Lock()
	lib.TitleIDs = ids
	lib.TitleHash = hash
	lib.mu.Unlock()
}

func (lib *Library) saveCacheLocked(titles []*Title) error {
	if lib.CachePath == "" {
		return nil
	}
	return writeLibraryCache(lib.CachePath, titlesToCache(lib.Dir, titles))
}

// GenerateThumbnails iterates over all entries and generates thumbnails for
// those that don't already have one. Matches Crystal Library#generate_thumbnails.
func (lib *Library) GenerateThumbnails() error {
	lib.mu.RLock()
	defer lib.mu.RUnlock()

	if lib.thumbnailCtx.Current > 0 {
		log.Println("Thumbnail generation already in progress")
		return nil
	}

	var allEntries []Entry
	for _, t := range lib.TitleHash {
		allEntries = append(allEntries, t.DeepEntries()...)
	}

	lib.thumbnailCtx.Total = len(allEntries)
	lib.thumbnailCtx.Current = 0

	if lib.thumbnailCtx.Total == 0 {
		return nil
	}

	log.Printf("Starting thumbnail generation for %d entries", lib.thumbnailCtx.Total)

	for _, e := range allEntries {
		if e.Err() != nil {
			lib.thumbnailCtx.Increment()
			continue
		}

		existing, err := lib.St.GetThumbnail(e.ID())
		if err == nil && existing != nil {
			lib.thumbnailCtx.Increment()
			continue
		}

		img, err := e.ReadPage(1)
		if err != nil {
			log.Printf("Failed to read page 1 of %s: %v", e.Path(), err)
			lib.thumbnailCtx.Increment()
			continue
		}

		thumb, err := thumbnail.Generate(img.Data, img.Filename)
		if err != nil {
			log.Printf("Failed to generate thumbnail for %s: %v", e.Path(), err)
			lib.thumbnailCtx.Increment()
			continue
		}

		stImg := &storage.Image{
			Data:     thumb.Data,
			Filename: thumb.Filename,
			Mime:     thumb.Mime,
			Size:     thumb.Size,
		}

		if err := lib.St.SaveThumbnail(e.ID(), stImg); err != nil {
			log.Printf("Failed to save thumbnail for %s: %v", e.Path(), err)
		}

		time.Sleep(100 * time.Millisecond)
		lib.thumbnailCtx.Increment()
	}

	log.Println("Thumbnail generation finished")
	lib.thumbnailCtx.Reset()
	return nil
}
