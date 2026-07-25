package dailylogin

import (
	"testing"

	"github.com/MaaXYZ/maa-framework-go/v4"
)

func TestActivityListROIFromRedDot(t *testing.T) {
	got := activityListROIFromRedDot(maa.Rect{820, 436, 10, 10})
	want := maa.Rect{447, 436, 383, 130}
	if got != want {
		t.Fatalf("activityListROIFromRedDot() = %#v, want %#v", got, want)
	}
}

func TestInsetActivityListROI(t *testing.T) {
	got := insetActivityListROI(maa.Rect{447, 436, 383, 130})
	want := maa.Rect{467, 456, 343, 90}
	if got != want {
		t.Fatalf("insetActivityListROI() = %#v, want %#v", got, want)
	}
}

func TestActivityListSwipePoints(t *testing.T) {
	begin, end := activityListSwipePoints(maa.Rect{801, 142, 37, 520})
	if want := (maa.Rect{819, 622, 1, 1}); begin != want {
		t.Fatalf("swipe begin = %#v, want %#v", begin, want)
	}
	if want := (maa.Rect{819, 622 - activityListSwipeDistance, 1, 1}); end != want {
		t.Fatalf("swipe end = %#v, want %#v", end, want)
	}
}

func TestActivityListProbeHasNotMoved(t *testing.T) {
	if !activityListProbeHasNotMoved(maa.Rect{471, 569, 181, 30}) {
		t.Fatal("a 7px probe movement should be treated as no movement")
	}
	if activityListProbeHasNotMoved(maa.Rect{471, 570, 181, 30}) {
		t.Fatal("an 8px vertical movement should be treated as movement")
	}
	if activityListProbeHasNotMoved(maa.Rect{479, 562, 181, 30}) {
		t.Fatal("an 8px horizontal movement should be treated as movement")
	}
}

func TestActivityListRecognitionResult(t *testing.T) {
	got := activityListRecognitionResult(maa.Rect{516, 382, 195, 46})
	if want := (maa.Rect{536, 402, 155, 6}); got.Box != want {
		t.Fatalf("recognition box = %#v, want %#v", got.Box, want)
	}
}
