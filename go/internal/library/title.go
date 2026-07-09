package library

import (
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/hkalexling/mango-go/internal/archive"
	"github.com/hkalexling/mango-go/internal/storage"
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

// ---------------------------------------------------------------------------
// Signature helpers (matching Crystal src/util/signature.cr)
// ---------------------------------------------------------------------------

// fileSignature computes a uint64 signature for a regular file, using a hash
// of path + modtime + size as a practical substitute for inode number (which
// is platform-specific in Go).
func fileSignature(path string, fi os.FileInfo) uint64 {
	h := fnv.New64a()
	h.Write([]byte(filepath.ToSlash(path)))
	binary.Write(h, binary.LittleEndian, fi.ModTime().UnixNano())
	binary.Write(h, binary.LittleEndian, fi.Size())
	return h.Sum64()
}

// dirSignature walks the directory tree and computes a uint64 hash from
// relative filenames + modtimes, matching the spirit of Crystal's
// Dir.signature (CRC32 of sorted inode numbers).
func dirSignature(dirname string) uint64 {
	h := fnv.New64a()
	filepath.WalkDir(dirname, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return filepath.SkipDir
		}
		rel, _ := filepath.Rel(dirname, path)
		h.Write([]byte(filepath.ToSlash(rel)))
		info, iErr := d.Info()
		if iErr == nil {
			binary.Write(h, binary.LittleEndian, info.ModTime().UnixNano())
		}
		return nil
	})
	return h.Sum64()
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

	// Filter and sort image entries
	var imgEntries []archive.Entry
	for _, entry := range entries {
		if isSupportedImage(mimeFromFilename(entry.Name)) {
			imgEntries = append(imgEntries, entry)
		}
	}
	sort.Slice(imgEntries, func(i, j int) bool {
		return compareNumerically(imgEntries[i].Name, imgEntries[j].Name) < 0
	})

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

// dirEntrySignature computes a hash of the sorted file signatures.
func dirEntrySignature(files []string) uint64 {
	h := fnv.New64a()
	for _, f := range files {
		fi, err := os.Stat(f)
		if err != nil {
			continue
		}
		binary.Write(h, binary.LittleEndian, fi.ModTime().UnixNano())
		binary.Write(h, binary.LittleEndian, fi.Size())
	}
	return h.Sum64()
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
	// Signature is the dir signature for change detection.
	Signature uint64
	Name      string
	TitleIDs  []string
	Entries   []Entry
	Mtime     time.Time
}

// NewTitle creates a Title for the given directory, scanning its contents
// recursively. It assigns DB IDs via st.GetOrCreateTitleID / GetOrCreateEntryID.
func NewTitle(absPath, parentID string, st *storage.Storage) *Title {
	t := &Title{
		Dir:      absPath,
		ParentID: parentID,
		Name:     filepath.Base(absPath),
	}

	t.Signature = dirSignature(absPath)

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

			// Recurse as sub-title
			sub := NewTitle(childPath, t.ID, st)
			if sub.ID != "" && (len(sub.Entries) > 0 || len(sub.TitleIDs) > 0) {
				t.TitleIDs = append(t.TitleIDs, sub.ID)
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

	// Sort sub-titles numerically by name
	sort.Slice(t.TitleIDs, func(i, j int) bool {
		return compareNumerically(t.TitleIDs[i], t.TitleIDs[j]) < 0
	})

	// Sort entries by name (numeric sort)
	chSorter := newChapterSorter(t.entryNames())
	sort.Slice(t.Entries, func(i, j int) bool {
		return chSorter.compare(t.Entries[i].Name(), t.Entries[j].Name()) < 0
	})

	// Recompute mtime as max of self and child entries
	for _, e := range t.Entries {
		if e.Mtime().After(t.Mtime) {
			t.Mtime = e.Mtime()
		}
	}
	// For sub-titles we can't easily look up their mtimes without a hash map
	// but the title's own mtime reflects directory changes

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
	if len(t.TitleIDs) == 0 {
		return t.Entries
	}
	// Phase 2 simplification: returns direct entries only.
	// Full recursive entry collection will be added when the Library
	// hash map is available (future phase).
	return t.Entries
}
