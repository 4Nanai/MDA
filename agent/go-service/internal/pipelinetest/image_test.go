package pipelinetest

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadScreenshot(t *testing.T) {
	dir := t.TempDir()
	valid := filepath.Join(dir, "valid.png")
	writeTestPNG(t, valid, 1280, 720)
	img, err := loadScreenshot(valid)
	if err != nil || img.Bounds().Dx() != 1280 {
		t.Fatalf("image=%v error=%v", img, err)
	}

	invalid := filepath.Join(dir, "invalid.png")
	writeTestPNG(t, invalid, 640, 360)
	_, err = loadScreenshot(invalid)
	if err == nil || !strings.Contains(err.Error(), "640x360") || !strings.Contains(err.Error(), "1280x720") {
		t.Fatalf("unexpected dimension error: %v", err)
	}

	unreadable := filepath.Join(dir, "unreadable.jpg")
	mustWrite(t, unreadable, "not an image")
	if _, err := loadScreenshot(unreadable); err == nil || !strings.Contains(err.Error(), "decode image") {
		t.Fatalf("unexpected decode error: %v", err)
	}
	if _, err := loadScreenshot(filepath.Join(dir, "shot.gif")); err == nil || !strings.Contains(err.Error(), "unsupported") {
		t.Fatalf("unexpected extension error: %v", err)
	}
}

func TestRuntimeInputStages(t *testing.T) {
	dir := t.TempDir()
	imagePath := filepath.Join(dir, "valid.png")
	writeTestPNG(t, imagePath, 1280, 720)
	executor := runtimeExecutor{}

	_, err := executor.Execute(Config{ImagePath: imagePath, LibraryDir: filepath.Join(dir, "missing-lib"), ResourcePath: dir})
	assertStage(t, err, "framework")

	lib := filepath.Join(dir, "lib")
	mustMkdir(t, lib)
	_, err = executor.Execute(Config{ImagePath: imagePath, LibraryDir: lib, ResourcePath: filepath.Join(dir, "missing-resource")})
	assertStage(t, err, "resource")
}

func writeTestPNG(t *testing.T, path string, width, height int) {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	img.SetRGBA(0, 0, color.RGBA{255, 0, 0, 255})
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(file, img); err != nil {
		_ = file.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
}

func assertStage(t *testing.T, err error, stage string) {
	t.Helper()
	staged, ok := err.(*StageError)
	if !ok || staged.Stage != stage {
		t.Fatalf("error = %#v, want stage %q", err, stage)
	}
}
