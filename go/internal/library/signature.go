package library

import (
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"hash/fnv"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ---------------------------------------------------------------------------
// Signature helpers
// mirrors Crystal src/util/signature.cr (intent)
//
// Decision (07-12-go-signature): uint64 file/dir/entry signatures use FNV of
// path+mtime+size (portable), NOT Unix inode / CRC32. ContentsSignature uses
// SHA1 of supported basenames — same intent as Crystal Dir.contents_signature.
// DB IDs stay stable via storage path-only fallback when signature changes.
// ---------------------------------------------------------------------------

// fileSignature computes a uint64 for a regular file (path + mtime + size).
// Substitute for Crystal File.signature (inode).
func fileSignature(path string, fi os.FileInfo) uint64 {
	h := fnv.New64a()
	h.Write([]byte(filepath.ToSlash(path)))
	binary.Write(h, binary.LittleEndian, fi.ModTime().UnixNano())
	binary.Write(h, binary.LittleEndian, fi.Size())
	return h.Sum64()
}

// FileSignature returns a uint64 signature for a supported archive/image file,
// or 0 if the path is unsupported / unreadable.
// mirrors Crystal File.signature
func FileSignature(path string) uint64 {
	if !isSupportedArchive(path) && !isSupportedImageFile(path) {
		return 0
	}
	fi, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return fileSignature(path, fi)
}

// dirSignature walks the directory tree and computes a uint64 hash from
// relative paths + mtimes (FNV). Spirit of Crystal Dir.signature (CRC32 of inodes).
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

// DirSignature is the exported form of dirSignature.
// mirrors Crystal Dir.signature (algorithm: FNV, not CRC32/inode — see package comment).
func DirSignature(dirname string) uint64 {
	return dirSignature(dirname)
}

// dirEntrySignature computes a hash of the sorted file mtimes/sizes (FNV).
// Crystal used SHA1 of inode strings; Go keeps FNV for consistency with FileSignature.
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

// DirectoryEntrySignature is the exported form of dirEntrySignature.
func DirectoryEntrySignature(files []string) uint64 {
	return dirEntrySignature(files)
}

// ContentsSignature returns a SHA1 hex digest of the directory's supported
// content names (recursive). Used to decide whether a title needs rescan.
// mirrors Crystal Dir.contents_signature
//
// Algorithm:
//   - skip names starting with '.'
//   - directories: recurse and append child hash
//   - files: if supported archive or image, append basename only
//   - join with no separator, SHA1 hex
//   - cache by absolute/logical dirname key
func ContentsSignature(dirname string, cache map[string]string) (string, error) {
	if cache != nil {
		if h, ok := cache[dirname]; ok {
			return h, nil
		}
	}

	entries, err := os.ReadDir(dirname)
	if err != nil {
		return "", fmt.Errorf("contents signature read %s: %w", dirname, err)
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	var parts []string
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		path := filepath.Join(dirname, name)
		if e.IsDir() {
			child, err := ContentsSignature(path, cache)
			if err != nil {
				return "", err
			}
			parts = append(parts, child)
			continue
		}
		// Crystal: ArchiveEntry.is_valid?(fn) || is_supported_image_file(fn)
		if isSupportedArchive(name) || isSupportedImageFile(name) {
			parts = append(parts, name)
		}
	}

	sum := sha1.Sum([]byte(strings.Join(parts, "")))
	hash := hex.EncodeToString(sum[:])
	if cache != nil {
		cache[dirname] = hash
	}
	return hash, nil
}
