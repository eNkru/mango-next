// Package upload mirrors Crystal src/upload.cr (Upload class).
package upload

import (
	"fmt"
	"io"
	"log"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// URLPrefix is the public URL path prefix for uploaded files.
// mirrors Crystal UPLOAD_URL_PREFIX in src/util/util.cr
const URLPrefix = "/uploads"

// SupportedImgTypes matches Crystal SUPPORTED_IMG_TYPES (src/util/util.cr).
var SupportedImgTypes = map[string]struct{}{
	"image/jpeg":    {},
	"image/png":     {},
	"image/webp":    {},
	"image/apng":    {},
	"image/avif":    {},
	"image/gif":     {},
	"image/svg+xml": {},
	"image/jxl":     {},
}

// IsSupportedImageMIME reports whether mimeType is in the Crystal whitelist.
func IsSupportedImageMIME(mimeType string) bool {
	_, ok := SupportedImgTypes[mimeType]
	return ok
}

// MIMEFromFilename returns the MIME type from the file extension, matching
// Crystal MIME.from_filename?(filename). Empty string if unknown.
func MIMEFromFilename(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	// Crystal registers extra types; ensure common comic image ext map.
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".webp":
		return "image/webp"
	case ".apng":
		return "image/apng"
	case ".avif":
		return "image/avif"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".jxl":
		return "image/jxl"
	}
	if t := mime.TypeByExtension(ext); t != "" {
		// strip "; charset=..." if present
		if i := strings.Index(t, ";"); i >= 0 {
			t = strings.TrimSpace(t[:i])
		}
		return t
	}
	return ""
}

// Upload manages files under an uploads directory.
// mirrors Crystal class Upload
type Upload struct {
	dir string
}

// New creates an Upload root. Creates dir if missing.
// mirrors Crystal Upload#initialize
func New(dir string) (*Upload, error) {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		log.Printf("The uploads directory %s does not exist. Attempting to create it", dir)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("create upload dir %s: %w", dir, err)
		}
	}
	return &Upload{dir: dir}, nil
}

// Dir returns the upload root directory.
func (u *Upload) Dir() string { return u.dir }

// Save writes r to {dir}/{subDir}/{random_str}{ext} and returns the full path.
// mirrors Crystal Upload#save
func (u *Upload) Save(subDir, ext string, r io.Reader) (string, error) {
	fullDir := filepath.Join(u.dir, subDir)
	if err := os.MkdirAll(fullDir, 0o755); err != nil {
		return "", fmt.Errorf("create subdir %s: %w", fullDir, err)
	}
	filename := randomStr() + ext
	filePath := filepath.Join(fullDir, filename)
	f, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err := io.Copy(f, r); err != nil {
		return "", err
	}
	return filePath, nil
}

// PathToURL converts a filesystem path under the upload root to a public URL
// path. Returns ("", false) if path is outside the upload directory.
// mirrors Crystal Upload#path_to_url
func (u *Upload) PathToURL(path string) (string, bool) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", false
	}
	absDir, err := filepath.Abs(u.dir)
	if err != nil {
		return "", false
	}
	rel, err := filepath.Rel(absDir, absPath)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		log.Printf("File %s is not in the upload directory %s", path, u.dir)
		return "", false
	}
	// URL path always uses forward slashes
	urlPath := URLPrefix + "/" + filepath.ToSlash(rel)
	return urlPath, true
}

// randomStr matches Crystal random_str / storage.randomStr (UUID without dashes).
func randomStr() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}
