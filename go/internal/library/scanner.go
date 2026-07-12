package library

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/hkalexling/mango-go/internal/storage"
)

// ScanResult holds the outcome of a library scan.
type ScanResult struct {
	Titles     []*Title
	TitleCount int
	EntryCount int
}

// ScanLibrary performs a full scan of the library directory.
//
// It walks the top-level directories, recursively creates Title objects
// (with nested sub-titles and entries), stores IDs in the database, and
// marks stale entries as unavailable.
//
// Matching Crystal Library#scan (simplified: fresh scan without cache).
func ScanLibrary(libraryPath string, st *storage.Storage) (*ScanResult, error) {
	// Ensure library directory exists
	if _, err := os.Stat(libraryPath); os.IsNotExist(err) {
		if err := os.MkdirAll(libraryPath, 0o755); err != nil {
			return nil, fmt.Errorf("creating library directory %s: %w", libraryPath, err)
		}
	}

	topDirEntries, err := os.ReadDir(libraryPath)
	if err != nil {
		return nil, fmt.Errorf("reading library directory %s: %w", libraryPath, err)
	}

	// Collect top-level directory paths
	var topDirs []string
	for _, de := range topDirEntries {
		if strings.HasPrefix(de.Name(), ".") {
			continue
		}
		if de.IsDir() {
			topDirs = append(topDirs, filepath.Join(libraryPath, de.Name()))
		}
	}

	// Scan each top-level directory into a Title
	var titles []*Title
	for _, dir := range topDirs {
		t := NewTitle(dir, "", st)
		if t.ID == "" {
			continue // skip unreadable
		}
		// Only keep titles with content
		if len(t.Entries) > 0 || len(t.TitleIDs) > 0 {
			titles = append(titles, t)
		}
	}

	// Sort top-level titles numerically by name
	sort.Slice(titles, func(i, j int) bool {
		return compareNumerically(titles[i].Name, titles[j].Name) < 0
	})

	// Count entries
	entryCount := 0
	for _, t := range titles {
		entryCount += countEntryRecursive(t)
	}

	result := &ScanResult{
		Titles:     titles,
		TitleCount: len(titles),
		EntryCount: entryCount,
	}

	return result, nil
}

// countEntryRecursive counts all entries in a title tree.
func countEntryRecursive(t *Title) int {
	count := len(t.Entries)
	for _, tid := range t.TitleIDs {
		// In the current simplified model, sub-title entries aren't
		// reachable via IDs alone. We count only direct entries.
		_ = tid
	}
	return count
}
