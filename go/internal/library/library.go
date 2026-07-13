package library

import (
	"fmt"
	"log"
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

// NewLibrary creates a Library instance bound to the given storage.
func (lib *Library) Lock()     { lib.mu.Lock() }
func (lib *Library) Unlock()   { lib.mu.Unlock() }
func (lib *Library) RLock()    { lib.mu.RLock() }
func (lib *Library) RUnlock()  { lib.mu.RUnlock() }

func NewLibrary(libraryPath string, st *storage.Storage) *Library {
	return &Library{
		Dir:       libraryPath,
		TitleIDs:  make([]string, 0),
		TitleHash: make(map[string]*Title),
		St:        st,
	}
}

// Scan performs a full library scan without holding the read/write lock during
// the disk walk. HTTP handlers use RLock() on TitleHash; holding mu for the
// entire ScanLibrary (minutes on large NAS libraries) starved the UI.
// Disk work runs unlocked; only the TitleIDs/TitleHash swap takes a short Lock.
func (lib *Library) Scan() (*ScanResult, error) {
	lib.scanMu.Lock()
	defer lib.scanMu.Unlock()

	// Heavy work: no lib.mu — readers can serve the previous (or empty) tree.
	result, err := ScanLibrary(lib.Dir, lib.St)
	if err != nil {
		return nil, fmt.Errorf("scan library: %w", err)
	}

	ids := make([]string, 0, len(result.Titles))
	hash := make(map[string]*Title, len(result.Titles))
	for _, t := range result.Titles {
		ids = append(ids, t.ID)
		hash[t.ID] = t
	}

	lib.mu.Lock()
	lib.TitleIDs = ids
	lib.TitleHash = hash
	n := len(lib.TitleIDs)
	lib.mu.Unlock()

	log.Printf("Scanned %d titles", n)
	return result, nil
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

	// Collect all entries
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

		// Check if thumbnail already exists
		existing, err := lib.St.GetThumbnail(e.ID())
		if err == nil && existing != nil {
			lib.thumbnailCtx.Increment()
			continue
		}

		// Read first page
		img, err := e.ReadPage(1)
		if err != nil {
			log.Printf("Failed to read page 1 of %s: %v", e.Path(), err)
			lib.thumbnailCtx.Increment()
			continue
		}

		// Generate thumbnail using the existing thumbnail package
		thumb, err := thumbnail.Generate(img.Data, img.Filename)
		if err != nil {
			log.Printf("Failed to generate thumbnail for %s: %v", e.Path(), err)
			lib.thumbnailCtx.Increment()
			continue
		}

		// Convert thumbnail.Image -> storage.Image
		stImg := &storage.Image{
			Data:     thumb.Data,
			Filename: thumb.Filename,
			Mime:     thumb.Mime,
			Size:     thumb.Size,
		}

		if err := lib.St.SaveThumbnail(e.ID(), stImg); err != nil {
			log.Printf("Failed to save thumbnail for %s: %v", e.Path(), err)
		}

		time.Sleep(100 * time.Millisecond) // minimize disk/CPU impact
		lib.thumbnailCtx.Increment()
	}

	log.Println("Thumbnail generation finished")
	lib.thumbnailCtx.Reset()
	return nil
}
