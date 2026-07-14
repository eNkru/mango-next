package tasks

import (
	"context"
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/eNkru/mango-next/internal/library"
	"github.com/eNkru/mango-next/internal/storage"
)

func TestInitialThumbnailRunWaitsForScan(t *testing.T) {
	dir := t.TempDir()
	libraryDir := filepath.Join(dir, "library")
	chapterDir := filepath.Join(libraryDir, "Title", "Chapter")
	if err := os.MkdirAll(chapterDir, 0o755); err != nil {
		t.Fatal(err)
	}
	pagePath := filepath.Join(chapterDir, "001.jpg")
	f, err := os.Create(pagePath)
	if err != nil {
		t.Fatal(err)
	}
	img := image.NewRGBA(image.Rect(0, 0, 16, 16))
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			img.Set(x, y, color.RGBA{R: 180, G: 40, B: 60, A: 255})
		}
	}
	if err := jpeg.Encode(f, img, nil); err != nil {
		f.Close()
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	st, err := storage.Open(filepath.Join(dir, "mango.db"), libraryDir)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	lib := library.NewLibrary(libraryDir, st, "")
	runner := NewRunner(lib, 0, 1)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		runner.Start(ctx)
		close(done)
	}()
	t.Cleanup(func() {
		cancel()
		<-done
	})

	deadline := time.Now().Add(5 * time.Second)
	for {
		var count int
		if err := st.DB().QueryRow("SELECT COUNT(*) FROM thumbnails").Scan(&count); err != nil {
			t.Fatal(err)
		}
		if count > 0 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("initial thumbnail run did not follow the initial scan")
		}
		time.Sleep(10 * time.Millisecond)
	}
}
