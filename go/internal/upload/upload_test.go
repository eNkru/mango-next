package upload

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSaveAndPathToURL(t *testing.T) {
	root := t.TempDir()
	u, err := New(root)
	if err != nil {
		t.Fatal(err)
	}

	path, err := u.Save("img", ".png", strings.NewReader("fake-png-bytes"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("saved file missing: %v", err)
	}
	if !strings.HasSuffix(path, ".png") {
		t.Fatalf("expected .png suffix, got %s", path)
	}
	if !strings.Contains(path, filepath.Join("img")) {
		t.Fatalf("expected img subdir in path %s", path)
	}

	url, ok := u.PathToURL(path)
	if !ok {
		t.Fatal("PathToURL should succeed for saved file")
	}
	if !strings.HasPrefix(url, URLPrefix+"/img/") {
		t.Fatalf("unexpected url: %s", url)
	}
	if strings.Contains(url, "\\") {
		t.Fatalf("url must use forward slashes: %s", url)
	}
}

func TestPathToURLOutside(t *testing.T) {
	root := t.TempDir()
	u, err := New(root)
	if err != nil {
		t.Fatal(err)
	}
	outside := filepath.Join(t.TempDir(), "other.png")
	if err := os.WriteFile(outside, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, ok := u.PathToURL(outside); ok {
		t.Fatal("PathToURL should fail for path outside upload root")
	}
}

func TestNewCreatesDir(t *testing.T) {
	root := filepath.Join(t.TempDir(), "missing", "uploads")
	u, err := New(root)
	if err != nil {
		t.Fatal(err)
	}
	if u.Dir() != root {
		t.Fatalf("dir = %s", u.Dir())
	}
	if _, err := os.Stat(root); err != nil {
		t.Fatal(err)
	}
}

func TestIsSupportedImageMIME(t *testing.T) {
	if !IsSupportedImageMIME("image/png") {
		t.Fatal("png should be supported")
	}
	if IsSupportedImageMIME("application/pdf") {
		t.Fatal("pdf should not be supported")
	}
}

func TestMIMEFromFilename(t *testing.T) {
	cases := map[string]string{
		"a.jpg":  "image/jpeg",
		"a.JPEG": "image/jpeg",
		"a.png":  "image/png",
		"a.webp": "image/webp",
		"a.gif":  "image/gif",
		"a.jxl":  "image/jxl",
		"a.txt":  "text/plain",
		"a":      "",
	}
	for name, want := range cases {
		got := MIMEFromFilename(name)
		if want == "text/plain" {
			// TypeByExtension may return text/plain; either empty or text/plain is fine
			if got != "" && !strings.HasPrefix(got, "text/") {
				t.Errorf("%s: got %q", name, got)
			}
			continue
		}
		if got != want {
			t.Errorf("%s: got %q want %q", name, got, want)
		}
	}
}
