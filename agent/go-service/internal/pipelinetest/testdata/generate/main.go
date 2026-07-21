package main

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
)

const (
	width  = 1280
	height = 720
)

func main() {
	base := filepath.Join("internal", "pipelinetest", "testdata")
	marker := makeMarker(false)
	miss := makeMarker(true)
	positive := image.NewRGBA(image.Rect(0, 0, width, height))
	negative := image.NewRGBA(image.Rect(0, 0, width, height))
	fill(positive, color.RGBA{18, 26, 32, 255})
	fill(negative, color.RGBA{18, 26, 32, 255})
	drawAt(positive, marker, image.Pt(100, 120))
	drawTextARENA(positive, image.Pt(200, 300))
	mustWrite(filepath.Join(base, "positive.png"), positive)
	mustWrite(filepath.Join(base, "negative.png"), negative)
	mustWrite(filepath.Join(base, "resource", "image", "pipeline-test-template.png"), marker)
	mustWrite(filepath.Join(base, "resource", "image", "pipeline-test-miss.png"), miss)
}

func makeMarker(invert bool) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, 32, 24))
	for y := 0; y < 24; y++ {
		for x := 0; x < 32; x++ {
			red := (x/4+y/4)%2 == 0
			if invert {
				red = !red
			}
			if red {
				img.SetRGBA(x, y, color.RGBA{245, 45, 45, 255})
			} else {
				img.SetRGBA(x, y, color.RGBA{245, 210, 35, 255})
			}
		}
	}
	return img
}

func fill(img *image.RGBA, value color.RGBA) {
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			img.SetRGBA(x, y, value)
		}
	}
}

func drawAt(dst *image.RGBA, src image.Image, point image.Point) {
	for y := 0; y < src.Bounds().Dy(); y++ {
		for x := 0; x < src.Bounds().Dx(); x++ {
			dst.Set(point.X+x, point.Y+y, src.At(x, y))
		}
	}
}

// A compact block alphabet avoids OS font dependencies in generated OCR fixtures.
func drawTextARENA(img *image.RGBA, origin image.Point) {
	letters := map[rune][]string{
		'A': {"01110", "10001", "10001", "11111", "10001", "10001", "10001"},
		'R': {"11110", "10001", "10001", "11110", "10100", "10010", "10001"},
		'E': {"11111", "10000", "10000", "11110", "10000", "10000", "11111"},
		'N': {"10001", "11001", "10101", "10011", "10001", "10001", "10001"},
	}
	x := origin.X
	for _, letter := range "ARENA" {
		glyph := letters[letter]
		for row, pixels := range glyph {
			for column, pixel := range pixels {
				if pixel != '1' {
					continue
				}
				for dy := 0; dy < 8; dy++ {
					for dx := 0; dx < 8; dx++ {
						img.SetRGBA(x+column*8+dx, origin.Y+row*8+dy, color.RGBA{245, 245, 245, 255})
					}
				}
			}
		}
		x += 48
	}
}

func mustWrite(path string, img image.Image) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		panic(err)
	}
	file, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	if err := png.Encode(file, img); err != nil {
		_ = file.Close()
		panic(err)
	}
	if err := file.Close(); err != nil {
		panic(err)
	}
}
