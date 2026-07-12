package library

import (
	"os"
	"path/filepath"
	"testing"
)

func TestContentsSignatureStableAndSensitive(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "a.cbz"), []byte("z"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "readme.txt"), []byte("nope"), 0o644); err != nil {
		t.Fatal(err)
	}

	cache := map[string]string{}
	h1, err := ContentsSignature(root, cache)
	if err != nil {
		t.Fatal(err)
	}
	h1b, err := ContentsSignature(root, cache)
	if err != nil {
		t.Fatal(err)
	}
	if h1 != h1b || h1 == "" {
		t.Fatalf("cache/stable: %q vs %q", h1, h1b)
	}

	// unsupported file should not change hash
	if err := os.WriteFile(filepath.Join(root, "notes.md"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	h2, err := ContentsSignature(root, nil)
	if err != nil {
		t.Fatal(err)
	}
	if h2 != h1 {
		t.Fatalf("unsupported file changed hash: %s -> %s", h1, h2)
	}

	// supported file changes hash
	if err := os.WriteFile(filepath.Join(root, "b.zip"), []byte("z2"), 0o644); err != nil {
		t.Fatal(err)
	}
	h3, err := ContentsSignature(root, nil)
	if err != nil {
		t.Fatal(err)
	}
	if h3 == h1 {
		t.Fatal("adding supported archive should change contents signature")
	}
}

func TestContentsSignatureNestedAndDotfiles(t *testing.T) {
	root := t.TempDir()
	sub := filepath.Join(root, "vol1")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "ch1.cbz"), []byte("1"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".hidden.cbz"), []byte("h"), 0o644); err != nil {
		t.Fatal(err)
	}

	h1, err := ContentsSignature(root, nil)
	if err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(sub, "ch2.cbz"), []byte("2"), 0o644); err != nil {
		t.Fatal(err)
	}
	h2, err := ContentsSignature(root, nil)
	if err != nil {
		t.Fatal(err)
	}
	if h1 == h2 {
		t.Fatal("nested supported file should change parent contents signature")
	}
}

func TestFileSignatureSupported(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "x.cbz")
	if err := os.WriteFile(p, []byte("data"), 0o644); err != nil {
		t.Fatal(err)
	}
	if FileSignature(p) == 0 {
		t.Fatal("expected non-zero for archive")
	}
	if FileSignature(filepath.Join(dir, "x.txt")) != 0 {
		t.Fatal("expected zero for unsupported")
	}
}

func TestDirSignatureAndDirectoryEntrySignature(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "a.jpg")
	if err := os.WriteFile(p, []byte("img"), 0o644); err != nil {
		t.Fatal(err)
	}
	if DirSignature(dir) == 0 {
		// empty walk of only files still hashes paths — should be non-zero
		t.Fatal("expected non-zero dir signature")
	}
	if DirectoryEntrySignature([]string{p}) == 0 {
		t.Fatal("expected non-zero entry signature")
	}
}
