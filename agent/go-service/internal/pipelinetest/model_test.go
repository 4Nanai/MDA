package pipelinetest

import (
	"encoding/json"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"testing"

	maa "github.com/MaaXYZ/maa-framework-go/v4"
)

func TestConvertDetailPreservesCombinedResults(t *testing.T) {
	detail := &maa.RecognitionDetail{
		Name:       "Parent",
		Algorithm:  "And",
		Hit:        true,
		Box:        maa.Rect{1, 2, 3, 4},
		DetailJson: `[{"name":"Child"}]`,
		CombinedResult: []*maa.RecognitionDetail{
			{Name: "Child", Algorithm: "TemplateMatch", Hit: false, Box: maa.Rect{5, 6, 7, 8}, DetailJson: `{"all":[]}`},
		},
	}
	result := convertDetail(detail)
	if len(result.Combined) != 1 || result.Combined[0].Node != "Child" || result.Combined[0].Hit {
		t.Fatalf("combined result was not preserved: %#v", result)
	}
	encoded, err := json.Marshal(result)
	if err != nil || !json.Valid(encoded) {
		t.Fatalf("invalid result JSON: %s (%v)", encoded, err)
	}
}

func TestEffectiveHitRepairsCombinedChildDetails(t *testing.T) {
	hitChild := &maa.RecognitionDetail{
		Algorithm: "TemplateMatch",
		Results:   &maa.RecognitionResults{Best: new(maa.RecognitionResult)},
	}
	missChild := &maa.RecognitionDetail{
		Algorithm: "TemplateMatch",
		Results:   &maa.RecognitionResults{},
	}
	if !effectiveHit(hitChild) || effectiveHit(missChild) {
		t.Fatal("best-result inference is incorrect")
	}
	if !effectiveHit(&maa.RecognitionDetail{Algorithm: "And", CombinedResult: []*maa.RecognitionDetail{hitChild, hitChild}}) {
		t.Fatal("And inference is incorrect")
	}
	if !effectiveHit(&maa.RecognitionDetail{Algorithm: "Or", CombinedResult: []*maa.RecognitionDetail{missChild, hitChild}}) {
		t.Fatal("Or inference is incorrect")
	}
}

func TestSaveDrawArtifacts(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	img.SetRGBA(0, 0, color.RGBA{255, 0, 0, 255})
	detail := &maa.RecognitionDetail{
		Name:  `Parent:Node`,
		Draws: []image.Image{img},
		CombinedResult: []*maa.RecognitionDetail{
			{Name: "Child", Draws: []image.Image{img}},
		},
	}
	dir := t.TempDir()
	paths, err := saveDrawArtifacts(dir, detail)
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != 2 {
		t.Fatalf("paths = %v", paths)
	}
	for _, path := range paths {
		if filepath.Dir(path) != dir {
			t.Fatalf("artifact escaped directory: %s", path)
		}
		if _, err := os.Stat(path); err != nil {
			t.Fatal(err)
		}
	}
}
