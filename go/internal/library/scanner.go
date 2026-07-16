package library

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/eNkru/mango-next/internal/storage"
)

// ScanResult holds the outcome of a library scan.
type ScanResult struct {
	Titles     []*Title
	TitleCount int
	EntryCount int
	// Reused is how many top-level titles were kept from the previous tree
	// because DirSignature matched (skipped expensive NewTitle).
	Reused int
	// Rebuilt is how many top-level titles were fully scanned with NewTitle.
	Rebuilt int
}

// ScanLibrary scans the library directory.
//
// previous maps absolute title directory path → existing *Title from memory
// or cache. When previous[dir].Signature equals the current DirSignature(dir),
// that title is reused without NewTitle (no archive re-open).
//
// Matching Crystal Library#scan with cache-assisted incremental reuse.
func ScanLibrary(libraryPath string, st *storage.Storage, previous map[string]*Title) (*ScanResult, error) {
	if previous == nil {
		previous = map[string]*Title{}
	}

	if _, err := os.Stat(libraryPath); os.IsNotExist(err) {
		if err := os.MkdirAll(libraryPath, 0o755); err != nil {
			return nil, fmt.Errorf("creating library directory %s: %w", libraryPath, err)
		}
	}

	topDirEntries, err := os.ReadDir(libraryPath)
	if err != nil {
		return nil, fmt.Errorf("reading library directory %s: %w", libraryPath, err)
	}

	var topDirs []string
	for _, de := range topDirEntries {
		if strings.HasPrefix(de.Name(), ".") {
			continue
		}
		if de.IsDir() {
			topDirs = append(topDirs, filepath.Join(libraryPath, de.Name()))
		}
	}

	var titles []*Title
	reused, rebuilt := 0, 0
	for _, dir := range topDirs {
		sig := DirSignature(dir)
		if old, ok := previous[dir]; ok && old != nil && old.ID != "" && old.Signature == sig {
			titles = append(titles, old)
			reused++
			continue
		}
		t := NewTitle(dir, "", st)
		if t.ID == "" {
			continue
		}
		if len(t.Entries) > 0 || len(t.Children) > 0 {
			titles = append(titles, t)
			rebuilt++
		}
	}

	sort.Slice(titles, func(i, j int) bool {
		return compareNumerically(titles[i].Name, titles[j].Name) < 0
	})

	entryCount := 0
	for _, t := range titles {
		entryCount += countEntryRecursive(t)
	}

	return &ScanResult{
		Titles:     titles,
		TitleCount: len(titles),
		EntryCount: entryCount,
		Reused:     reused,
		Rebuilt:    rebuilt,
	}, nil
}

func countEntryRecursive(t *Title) int {
	count := len(t.Entries)
	for _, c := range t.Children {
		count += countEntryRecursive(c)
	}
	return count
}
