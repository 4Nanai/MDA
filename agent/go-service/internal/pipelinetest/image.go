package pipelinetest

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"strings"
)

const (
	requiredWidth  = 1280
	requiredHeight = 720
)

func loadScreenshot(path string) (image.Image, error) {
	ext := strings.ToLower(filepathExtension(path))
	if ext != ".png" && ext != ".jpg" && ext != ".jpeg" {
		return nil, fmt.Errorf("unsupported image extension %q; use PNG or JPEG", ext)
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	img, format, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}
	if format != "png" && format != "jpeg" {
		return nil, fmt.Errorf("decoded format %q is not PNG or JPEG", format)
	}
	bounds := img.Bounds()
	if bounds.Dx() != requiredWidth || bounds.Dy() != requiredHeight {
		return nil, fmt.Errorf("image dimensions are %dx%d; required %dx%d (no resizing is performed)", bounds.Dx(), bounds.Dy(), requiredWidth, requiredHeight)
	}
	return img, nil
}

func filepathExtension(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		switch path[i] {
		case '.', '/', '\\':
			if path[i] == '.' {
				return path[i:]
			}
			return ""
		}
	}
	return ""
}
