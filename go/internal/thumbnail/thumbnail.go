package thumbnail

import (
	"bytes"
	"image"
	"image/jpeg"
	"io"

	"golang.org/x/image/draw"
	"golang.org/x/image/webp"
)

type Image struct {
	Data     []byte
	Mime     string
	Filename string
	Size     int
}

func DecodeConfig(data []byte) (width, height int, err error) {
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		cfg, err = webp.DecodeConfig(bytes.NewReader(data))
		if err != nil {
			return 0, 0, err
		}
	}
	return cfg.Width, cfg.Height, nil
}

func Generate(data []byte, filename string) (*Image, error) {
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		cfg, err = webp.DecodeConfig(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
	}

	var dstWidth, dstHeight int
	if cfg.Height > cfg.Width {
		dstWidth = 200
		dstHeight = cfg.Height * 200 / cfg.Width
	} else {
		dstHeight = 300
		dstWidth = cfg.Width * 300 / cfg.Height
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		img, err = webp.Decode(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
	}

	dst := image.NewRGBA(image.Rect(0, 0, dstWidth, dstHeight))
	draw.BiLinear.Scale(dst, dst.Rect, img, img.Bounds(), draw.Over, nil)

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, dst, &jpeg.Options{Quality: 85}); err != nil {
		return nil, err
	}
	thumbnail := buf.Bytes()

	return &Image{
		Data:     thumbnail,
		Mime:     "image/jpeg",
		Filename: filename,
		Size:     len(thumbnail),
	}, nil
}

func Decode(r io.Reader) (image.Image, string, error) {
	return image.Decode(r)
}
