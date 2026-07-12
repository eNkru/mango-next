package archive

import (
	"io"
	"strings"
)

type Entry struct {
	Name string
}

type Reader interface {
	Entries() ([]Entry, error)
	ReadEntry(entry Entry) ([]byte, error)
	Close() error
}

func Open(filename string) (Reader, error) {
	lower := strings.ToLower(filename)
	switch {
	case strings.HasSuffix(lower, ".zip"), strings.HasSuffix(lower, ".cbz"):
		return openZip(filename)
	case strings.HasSuffix(lower, ".rar"), strings.HasSuffix(lower, ".cbr"):
		return openRar(filename)
	case strings.HasSuffix(lower, ".7z"):
		return openSevenZip(filename)
	default:
		return openZip(filename)
	}
}

func IsArchive(filename string) bool {
	lower := strings.ToLower(filename)
	for _, ext := range []string{".zip", ".cbz", ".rar", ".cbr", ".7z"} {
		if strings.HasSuffix(lower, ext) {
			return true
		}
	}
	return false
}

type entryReader struct {
	name string
	r    io.ReadCloser
}

func (e *entryReader) Read(p []byte) (int, error) { return e.r.Read(p) }
func (e *entryReader) Close() error               { return e.r.Close() }
