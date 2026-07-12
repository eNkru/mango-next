package thumbnail

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"
	"testing"
)

func createTestImage(t testing.TB, width, height int, asPNG bool) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	var buf bytes.Buffer
	if asPNG {
		if err := png.Encode(&buf, img); err != nil {
			t.Fatal(err)
		}
	} else {
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85}); err != nil {
			t.Fatal(err)
		}
	}
	return buf.Bytes()
}

func TestDecodeConfigJPEG(t *testing.T) {
	data := createTestImage(t, 400, 300, false)
	w, h, err := DecodeConfig(data)
	if err != nil {
		t.Fatal(err)
	}
	if w != 400 || h != 300 {
		t.Errorf("got %dx%d, want 400x300", w, h)
	}
}

func TestDecodeConfigPNG(t *testing.T) {
	data := createTestImage(t, 200, 500, true)
	w, h, err := DecodeConfig(data)
	if err != nil {
		t.Fatal(err)
	}
	if w != 200 || h != 500 {
		t.Errorf("got %dx%d, want 200x500", w, h)
	}
}

func TestDecodeConfigInvalid(t *testing.T) {
	_, _, err := DecodeConfig([]byte("not an image"))
	if err == nil {
		t.Error("expected error for invalid image data")
	}
}

func TestGeneratePortrait(t *testing.T) {
	data := createTestImage(t, 100, 300, false)
	thumb, err := Generate(data, "page001.jpg")
	if err != nil {
		t.Fatal(err)
	}

	cfg, _, err := image.DecodeConfig(bytes.NewReader(thumb.Data))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Width > 200 {
		t.Errorf("portrait thumbnail width = %d, want <= 200", cfg.Width)
	}
	aspect := float64(cfg.Width) / float64(cfg.Height)
	if aspect < 0.3 || aspect > 0.36 {
		t.Errorf("portrait thumbnail aspect = %.4f, expected ~1:3", aspect)
	}
	if thumb.Mime != "image/jpeg" {
		t.Errorf("mime = %q, want image/jpeg", thumb.Mime)
	}
}

func TestGenerateLandscape(t *testing.T) {
	data := createTestImage(t, 500, 200, false)
	thumb, err := Generate(data, "wide.jpg")
	if err != nil {
		t.Fatal(err)
	}

	cfg, _, err := image.DecodeConfig(bytes.NewReader(thumb.Data))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Height > 300 {
		t.Errorf("landscape thumbnail height = %d, want <= 300", cfg.Height)
	}
	aspect := float64(cfg.Width) / float64(cfg.Height)
	if aspect < 2.4 || aspect > 2.6 {
		t.Errorf("landscape thumbnail aspect = %.4f, expected ~2.5:1", aspect)
	}
}

func TestGenerateSquare(t *testing.T) {
	data := createTestImage(t, 200, 200, false)
	thumb, err := Generate(data, "square.jpg")
	if err != nil {
		t.Fatal(err)
	}

	cfg, _, err := image.DecodeConfig(bytes.NewReader(thumb.Data))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Height != 300 || cfg.Width != 300 {
		t.Errorf("square thumbnail = %dx%d, want 300x300", cfg.Width, cfg.Height)
	}
}

func TestGeneratePNGInput(t *testing.T) {
	data := createTestImage(t, 150, 400, true)
	thumb, err := Generate(data, "page.png")
	if err != nil {
		t.Fatal(err)
	}

	cfg, _, err := image.DecodeConfig(bytes.NewReader(thumb.Data))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Width != 200 {
		t.Errorf("png thumbnail width = %d, want 200", cfg.Width)
	}
	if thumb.Mime != "image/jpeg" {
		t.Errorf("mime = %q, want image/jpeg", thumb.Mime)
	}
}

func TestGenerateSmallImage(t *testing.T) {
	data := createTestImage(t, 50, 50, false)
	thumb, err := Generate(data, "small.jpg")
	if err != nil {
		t.Fatal(err)
	}
	if len(thumb.Data) == 0 {
		t.Error("thumbnail should not be empty")
	}
}

func TestGenerateInvalidData(t *testing.T) {
	_, err := Generate([]byte("invalid"), "bad.jpg")
	if err == nil {
		t.Error("expected error for invalid image data")
	}
}

func benchmarkGenerate(b *testing.B, width, height int) {
	data := createTestImage(b, width, height, false)
	b.ResetTimer()
	for range b.N {
		_, err := Generate(data, "page.jpg")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGeneratePortrait(b *testing.B) {
	benchmarkGenerate(b, 800, 1200)
}

func BenchmarkGenerateLandscape(b *testing.B) {
	benchmarkGenerate(b, 1600, 900)
}


