package thumbnail

import (
	"bytes"
	_ "embed"
	"image"
	"image/jpeg"
	"strings"
	"testing"
)

// Minimal fixtures as raw bytes so this test package does not import image/png
// or image/gif (which would register decoders process-wide and mask missing
// production imports in thumbnail.go).

//go:embed testdata/sample.webp
var webpSample []byte

// 2x3 RGBA PNG generated with image/png.
var png2x3 = []byte{
	0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x00, 0x00, 0x0d,
	0x49, 0x48, 0x44, 0x52, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x03,
	0x08, 0x06, 0x00, 0x00, 0x00, 0xb9, 0xea, 0xde, 0x81, 0x00, 0x00, 0x00,
	0x17, 0x49, 0x44, 0x41, 0x54, 0x78, 0x9c, 0x62, 0xfa, 0xcf, 0xc0, 0xf0,
	0x9f, 0x01, 0x19, 0x30, 0xc1, 0x18, 0x80, 0x00, 0x00, 0x00, 0xff, 0xff,
	0x31, 0x32, 0x02, 0x03, 0x6a, 0x81, 0x50, 0x47, 0x00, 0x00, 0x00, 0x00,
	0x49, 0x45, 0x4e, 0x44, 0xae, 0x42, 0x60, 0x82,
}

// 1x1 GIF89a.
var gif1x1 = []byte{
	0x47, 0x49, 0x46, 0x38, 0x39, 0x61, 0x01, 0x00, 0x01, 0x00, 0x80, 0x00,
	0x00, 0x00, 0x00, 0x00, 0xff, 0xff, 0xff, 0x2c, 0x00, 0x00, 0x00, 0x00,
	0x01, 0x00, 0x01, 0x00, 0x00, 0x02, 0x02, 0x44, 0x01, 0x00, 0x3b,
}

func createJPEG(t testing.TB, width, height int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 85}); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestDecodeConfigJPEG(t *testing.T) {
	data := createJPEG(t, 400, 300)
	w, h, err := DecodeConfig(data)
	if err != nil {
		t.Fatal(err)
	}
	if w != 400 || h != 300 {
		t.Errorf("got %dx%d, want 400x300", w, h)
	}
}

func TestDecodeConfigPNG(t *testing.T) {
	w, h, err := DecodeConfig(png2x3)
	if err != nil {
		t.Fatal(err)
	}
	if w != 2 || h != 3 {
		t.Errorf("got %dx%d, want 2x3", w, h)
	}
}

func TestDecodeConfigGIF(t *testing.T) {
	w, h, err := DecodeConfig(gif1x1)
	if err != nil {
		t.Fatal(err)
	}
	if w != 1 || h != 1 {
		t.Errorf("got %dx%d, want 1x1", w, h)
	}
}

func TestDecodeConfigWebP(t *testing.T) {
	w, h, err := DecodeConfig(webpSample)
	if err != nil {
		t.Fatal(err)
	}
	if w <= 0 || h <= 0 {
		t.Errorf("got %dx%d, want positive dimensions", w, h)
	}
}

func TestDecodeConfigInvalid(t *testing.T) {
	_, _, err := DecodeConfig([]byte("not an image"))
	if err == nil {
		t.Error("expected error for invalid image data")
	}
	if strings.Contains(err.Error(), "riff:") {
		t.Errorf("non-image error should not come from WebP/RIFF path, got %v", err)
	}
}

func TestGeneratePortrait(t *testing.T) {
	data := createJPEG(t, 100, 300)
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
	data := createJPEG(t, 500, 200)
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
	data := createJPEG(t, 200, 200)
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
	// Portrait 2x3 → width clamped to 200, height 300.
	thumb, err := Generate(png2x3, "page.png")
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
	if cfg.Height != 300 {
		t.Errorf("png thumbnail height = %d, want 300", cfg.Height)
	}
	if thumb.Mime != "image/jpeg" {
		t.Errorf("mime = %q, want image/jpeg", thumb.Mime)
	}
}

func TestGenerateGIFInput(t *testing.T) {
	thumb, err := Generate(gif1x1, "page.gif")
	if err != nil {
		t.Fatal(err)
	}
	if len(thumb.Data) == 0 {
		t.Error("thumbnail should not be empty")
	}
	if thumb.Mime != "image/jpeg" {
		t.Errorf("mime = %q, want image/jpeg", thumb.Mime)
	}
}

func TestGenerateWebPInput(t *testing.T) {
	thumb, err := Generate(webpSample, "page.webp")
	if err != nil {
		t.Fatal(err)
	}
	if len(thumb.Data) == 0 {
		t.Error("thumbnail should not be empty")
	}
	if thumb.Mime != "image/jpeg" {
		t.Errorf("mime = %q, want image/jpeg", thumb.Mime)
	}
}

func TestGenerateSmallImage(t *testing.T) {
	data := createJPEG(t, 50, 50)
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
	if strings.Contains(err.Error(), "riff:") {
		t.Errorf("non-image error should not come from WebP/RIFF path, got %v", err)
	}
}

func benchmarkGenerate(b *testing.B, width, height int) {
	data := createJPEG(b, width, height)
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
