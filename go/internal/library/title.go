package library

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/eNkru/mango-next/internal/archive"
	"github.com/eNkru/mango-next/internal/storage"
	"github.com/eNkru/mango-next/internal/thumbnail"
)

// ---------------------------------------------------------------------------
// MIME / extension helpers (matching Crystal src/util/util.cr)
// ---------------------------------------------------------------------------

var supportedArchiveExts = map[string]bool{
	".zip": true, ".cbz": true, ".rar": true, ".cbr": true, ".7z": true,
}

// mimeFromFilename maps file extensions to MIME types, matching Crystal's
// MIME.from_filename plus custom registrations.
func mimeFromFilename(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	case ".gif":
		return "image/gif"
	case ".avif":
		return "image/avif"
	case ".apng":
		return "image/apng"
	case ".svg", ".svgz":
		return "image/svg+xml"
	case ".jxl":
		return "image/jxl"
	case ".zip":
		return "application/zip"
	case ".rar":
		return "application/x-rar-compressed"
	case ".cbz":
		return "application/vnd.comicbook+zip"
	case ".cbr":
		return "application/vnd.comicbook-rar"
	}
	return "application/octet-stream"
}

// isSupportedArchive returns true if the file has a supported archive extension.
func isSupportedArchive(path string) bool {
	return supportedArchiveExts[strings.ToLower(filepath.Ext(path))]
}

// isSupportedImage returns true if the MIME type is a supported image type.
func isSupportedImage(mimeType string) bool {
	switch mimeType {
	case "image/jpeg", "image/png", "image/webp", "image/gif",
		"image/apng", "image/avif", "image/svg+xml", "image/jxl":
		return true
	}
	return false
}

// isSupportedImageFile returns true if filename has a supported image extension.
func isSupportedImageFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp", ".gif",
		".avif", ".apng", ".svg", ".svgz", ".jxl":
		return true
	}
	return false
}

// isValidDirEntry returns true if the directory contains at least one
// supported image file (matching Crystal DirEntry.is_valid?).
func isValidDirEntry(dirPath string) bool {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if e.IsDir() || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		if isSupportedImageFile(e.Name()) {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// Entry interface (matching Crystal's abstract Entry class)
// ---------------------------------------------------------------------------

// Entry represents either an archive file (ArchiveEntry) or a directory of
// loose images (DirEntry).
type Entry interface {
	ReadPage(n int) (*storage.Image, error)
	PageCount() int
	ID() string
	Name() string
	Path() string
	Mtime() time.Time
	Err() error
	Signature() uint64
	Book() *Title
}

// ---------------------------------------------------------------------------
// ArchiveEntry — a single archive file (zip/cbz/rar/cbr/7z)
// ---------------------------------------------------------------------------

// ArchiveEntry wraps an archive file containing manga pages.
type ArchiveEntry struct {
	id    string
	title string
	path  string
	book  *Title
	pages int
	mtime time.Time
	err   error
	sig   uint64
}

// NewArchiveEntry creates an ArchiveEntry for the given archive file path.
func NewArchiveEntry(absPath string, book *Title, st *storage.Storage) *ArchiveEntry {
	e := &ArchiveEntry{
		title: strings.TrimSuffix(filepath.Base(absPath), filepath.Ext(absPath)),
		path:  absPath,
		book:  book,
	}

	fi, fiErr := os.Stat(absPath)
	if fiErr != nil {
		e.err = fmt.Errorf("file %s is not readable: %w", absPath, fiErr)
		return e
	}
	e.mtime = fi.ModTime()
	e.sig = fileSignature(absPath, fi)

	// Validate archive and count image pages
	arc, aErr := archive.Open(absPath)
	if aErr != nil {
		e.err = fmt.Errorf("archive error: %w", aErr)
		return e
	}

	entries, entErr := arc.Entries()
	arc.Close()
	if entErr != nil {
		e.err = fmt.Errorf("archive error: %w", entErr)
		return e
	}

	for _, entry := range entries {
		if isSupportedImage(mimeFromFilename(entry.Name)) {
			e.pages++
		}
	}

	// DB ID
	id, dbErr := st.GetOrCreateEntryID(absPath, e.sig)
	if dbErr != nil {
		e.err = fmt.Errorf("db error: %w", dbErr)
		return e
	}
	e.id = id

	return e
}

// sortedImageEntries filters archive entries to supported images and sorts
// them in reading order (shared by ReadPage and ReadPageDimensions).
func sortedImageEntries(entries []archive.Entry) []archive.Entry {
	var imgEntries []archive.Entry
	for _, entry := range entries {
		if isSupportedImage(mimeFromFilename(entry.Name)) {
			imgEntries = append(imgEntries, entry)
		}
	}
	sort.Slice(imgEntries, func(i, j int) bool {
		return compareNumerically(imgEntries[i].Name, imgEntries[j].Name) < 0
	})
	return imgEntries
}

func (e *ArchiveEntry) ReadPage(n int) (*storage.Image, error) {
	if e.err != nil {
		return nil, fmt.Errorf("unreadable archive: %w", e.err)
	}

	arc, err := archive.Open(e.path)
	if err != nil {
		return nil, err
	}
	defer arc.Close()

	entries, err := arc.Entries()
	if err != nil {
		return nil, err
	}

	imgEntries := sortedImageEntries(entries)

	if n < 1 || n > len(imgEntries) {
		return nil, fmt.Errorf("page %d out of range (1-%d)", n, len(imgEntries))
	}

	pageEntry := imgEntries[n-1]
	data, err := arc.ReadEntry(pageEntry)
	if err != nil {
		return nil, err
	}

	return &storage.Image{
		Data:     data,
		Filename: pageEntry.Name,
		Mime:     mimeFromFilename(pageEntry.Name),
		Size:     len(data),
	}, nil
}

// PageDimension is width/height for one page in reading order.
type PageDimension struct {
	Width  int
	Height int
}

// ReadPageDimensions returns width/height for every page with a single
// archive open (or one pass over directory files). Decode failures yield 0,0.
func ReadPageDimensions(e Entry) ([]PageDimension, error) {
	switch v := e.(type) {
	case *ArchiveEntry:
		return v.readPageDimensions()
	case *DirEntry:
		return v.readPageDimensions()
	default:
		// Fallback: per-page ReadPage (unknown entry kinds).
		dims := make([]PageDimension, 0, e.PageCount())
		for i := 1; i <= e.PageCount(); i++ {
			img, err := e.ReadPage(i)
			if err != nil {
				dims = append(dims, PageDimension{})
				continue
			}
			w, h, err := thumbnail.DecodeConfig(img.Data)
			if err != nil {
				w, h = 0, 0
			}
			dims = append(dims, PageDimension{Width: w, Height: h})
		}
		return dims, nil
	}
}

func (e *ArchiveEntry) readPageDimensions() ([]PageDimension, error) {
	if e.err != nil {
		return nil, fmt.Errorf("unreadable archive: %w", e.err)
	}

	arc, err := archive.Open(e.path)
	if err != nil {
		return nil, err
	}
	defer arc.Close()

	entries, err := arc.Entries()
	if err != nil {
		return nil, err
	}

	imgEntries := sortedImageEntries(entries)
	dims := make([]PageDimension, 0, len(imgEntries))
	for _, pageEntry := range imgEntries {
		data, err := arc.ReadEntry(pageEntry)
		if err != nil {
			dims = append(dims, PageDimension{})
			continue
		}
		w, h, err := thumbnail.DecodeConfig(data)
		if err != nil {
			w, h = 0, 0
		}
		dims = append(dims, PageDimension{Width: w, Height: h})
	}
	return dims, nil
}

func (e *DirEntry) readPageDimensions() ([]PageDimension, error) {
	if e.err != nil {
		return nil, fmt.Errorf("unreadable directory entry: %w", e.err)
	}
	dims := make([]PageDimension, 0, len(e.files))
	for _, path := range e.files {
		data, err := os.ReadFile(path)
		if err != nil {
			dims = append(dims, PageDimension{})
			continue
		}
		w, h, err := thumbnail.DecodeConfig(data)
		if err != nil {
			w, h = 0, 0
		}
		dims = append(dims, PageDimension{Width: w, Height: h})
	}
	return dims, nil
}

func (e *ArchiveEntry) PageCount() int    { return e.pages }
func (e *ArchiveEntry) ID() string        { return e.id }
func (e *ArchiveEntry) Name() string      { return e.title }
func (e *ArchiveEntry) Path() string      { return e.path }
func (e *ArchiveEntry) Mtime() time.Time  { return e.mtime }
func (e *ArchiveEntry) Err() error        { return e.err }
func (e *ArchiveEntry) Signature() uint64 { return e.sig }
func (e *ArchiveEntry) Book() *Title      { return e.book }

// ---------------------------------------------------------------------------
// DirEntry — a directory of loose images
// ---------------------------------------------------------------------------

// DirEntry wraps a directory containing loose image files.
type DirEntry struct {
	id      string
	title   string
	path    string
	book    *Title
	pages   int
	mtime   time.Time
	err     error
	sigHash uint64
	files   []string // sorted absolute paths of image files
}

// NewDirEntry creates a DirEntry for the given directory path.
func NewDirEntry(absPath string, book *Title, st *storage.Storage) *DirEntry {
	e := &DirEntry{
		title: filepath.Base(absPath),
		path:  absPath,
		book:  book,
	}

	fi, fiErr := os.Stat(absPath)
	if fiErr != nil {
		e.err = fmt.Errorf("directory %s is not readable: %w", absPath, fiErr)
		return e
	}
	if !fi.IsDir() {
		e.err = fmt.Errorf("path %s is not a directory", absPath)
		return e
	}

	dirEntries, rErr := os.ReadDir(absPath)
	if rErr != nil {
		e.err = fmt.Errorf("reading directory %s: %w", absPath, rErr)
		return e
	}

	for _, de := range dirEntries {
		if de.IsDir() || strings.HasPrefix(de.Name(), ".") {
			continue
		}
		fullPath := filepath.Join(absPath, de.Name())
		if isSupportedImageFile(fullPath) {
			e.files = append(e.files, fullPath)
		}
	}

	if len(e.files) == 0 {
		e.err = fmt.Errorf("no valid image files in %s", absPath)
		return e
	}

	// Sort numerically by base name
	sort.Slice(e.files, func(i, j int) bool {
		return compareNumerically(filepath.Base(e.files[i]), filepath.Base(e.files[j])) < 0
	})

	e.pages = len(e.files)

	// Max mtime
	var maxMtime time.Time
	for _, f := range e.files {
		info, iErr := os.Stat(f)
		if iErr == nil && info.ModTime().After(maxMtime) {
			maxMtime = info.ModTime()
		}
	}
	e.mtime = maxMtime

	// Compute signature hash (hash of sorted file signatures)
	e.sigHash = dirEntrySignature(e.files)

	// DB ID
	id, dbErr := st.GetOrCreateEntryID(absPath, e.sigHash)
	if dbErr != nil {
		e.err = fmt.Errorf("db error: %w", dbErr)
		return e
	}
	e.id = id

	return e
}

func (e *DirEntry) ReadPage(n int) (*storage.Image, error) {
	if e.err != nil {
		return nil, fmt.Errorf("unreadable directory entry: %w", e.err)
	}
	if n < 1 || n > len(e.files) {
		return nil, fmt.Errorf("page %d out of range (1-%d)", n, len(e.files))
	}

	data, err := os.ReadFile(e.files[n-1])
	if err != nil {
		return nil, err
	}

	return &storage.Image{
		Data:     data,
		Filename: filepath.Base(e.files[n-1]),
		Mime:     mimeFromFilename(e.files[n-1]),
		Size:     len(data),
	}, nil
}

func (e *DirEntry) PageCount() int    { return e.pages }
func (e *DirEntry) ID() string        { return e.id }
func (e *DirEntry) Name() string      { return e.title }
func (e *DirEntry) Path() string      { return e.path }
func (e *DirEntry) Mtime() time.Time  { return e.mtime }
func (e *DirEntry) Err() error        { return e.err }
func (e *DirEntry) Signature() uint64 { return e.sigHash }
func (e *DirEntry) Book() *Title      { return e.book }

// ---------------------------------------------------------------------------
// Title — a directory representing a manga series
// ---------------------------------------------------------------------------

// Title represents a directory containing manga content. It may contain
// sub-titles (nested directories) and/or entries (archives or loose images).
type Title struct {
	Dir      string
	ParentID string
	ID       string
	// Signature is the dir signature for change detection (FNV; see signature.go).
	Signature uint64
	// ContentsSig is SHA1 of supported content names; used for rescan checks.
	// mirrors Crystal Title @contents_signature
	ContentsSig string
	Name        string
	// TitleIDs lists nested title IDs in display order.
	TitleIDs []string
	// Children holds nested title objects produced by NewTitle / cache load.
	// Kept in sync with TitleIDs so handlers can resolve via TitleHash and
	// DeepEntries can walk the subtree without a library map.
	Children []*Title
	Entries  []Entry
	Mtime    time.Time
}

// NewTitle creates a Title for the given directory, scanning its contents
// recursively. It assigns DB IDs via st.GetOrCreateTitleID / GetOrCreateEntryID.
func NewTitle(absPath, parentID string, st *storage.Storage) *Title {
	t := &Title{
		Dir:      absPath,
		ParentID: parentID,
		Name:     filepath.Base(absPath),
	}

	t.Signature = DirSignature(absPath)
	if cs, err := ContentsSignature(absPath, nil); err == nil {
		t.ContentsSig = cs
	}

	id, err := st.GetOrCreateTitleID(absPath, t.Signature)
	if err != nil {
		return t
	}
	t.ID = id

	fi, _ := os.Stat(absPath)
	if fi != nil {
		t.Mtime = fi.ModTime()
	}

	dirEntries, err := os.ReadDir(absPath)
	if err != nil {
		return t
	}

	for _, de := range dirEntries {
		if strings.HasPrefix(de.Name(), ".") {
			continue
		}
		childPath := filepath.Join(absPath, de.Name())

		if de.IsDir() {
			// Check for loose images in this directory
			if isValidDirEntry(childPath) {
				de := NewDirEntry(childPath, t, st)
				if de.PageCount() > 0 || de.Err() != nil {
					t.Entries = append(t.Entries, de)
				}
			}

			// Recurse as sub-title (even when also a dir entry: nested archives
			// deeper than this folder still need a title node).
			sub := NewTitle(childPath, t.ID, st)
			if sub.ID != "" && (len(sub.Entries) > 0 || len(sub.Children) > 0) {
				t.Children = append(t.Children, sub)
			}
			continue
		}

		// Archive file
		if isSupportedArchive(childPath) {
			ae := NewArchiveEntry(childPath, t, st)
			if ae.PageCount() > 0 || ae.Err() != nil {
				t.Entries = append(t.Entries, ae)
			}
		}
	}

	// Sort sub-titles by name, then rebuild TitleIDs from Children.
	sort.Slice(t.Children, func(i, j int) bool {
		return compareNumerically(t.Children[i].Name, t.Children[j].Name) < 0
	})
	t.TitleIDs = make([]string, len(t.Children))
	for i, c := range t.Children {
		t.TitleIDs[i] = c.ID
	}

	// Sort entries by name (numeric sort)
	chSorter := newChapterSorter(t.entryNames())
	sort.Slice(t.Entries, func(i, j int) bool {
		return chSorter.compare(t.Entries[i].Name(), t.Entries[j].Name()) < 0
	})

	// Recompute mtime as max of self, entries, and nested titles
	for _, e := range t.Entries {
		if e.Mtime().After(t.Mtime) {
			t.Mtime = e.Mtime()
		}
	}
	for _, c := range t.Children {
		if c.Mtime.After(t.Mtime) {
			t.Mtime = c.Mtime
		}
	}

	// Metadata is applied after filesystem-name sorting so presentation
	// overrides do not silently change natural reading order.
	_ = t.ApplyDisplayNames()

	return t
}

// entryNames returns the Name() of each direct entry for sorting.
func (t *Title) entryNames() []string {
	names := make([]string, len(t.Entries))
	for i, e := range t.Entries {
		names[i] = e.Name()
	}
	return names
}

// DeepEntries returns all entries including those in nested sub-titles.
func (t *Title) DeepEntries() []Entry {
	if len(t.Children) == 0 {
		return t.Entries
	}
	out := make([]Entry, 0, len(t.Entries))
	out = append(out, t.Entries...)
	for _, c := range t.Children {
		out = append(out, c.DeepEntries()...)
	}
	return out
}
