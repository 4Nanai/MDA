package dailylogin

import (
	"encoding/json"
	"image"
	"image/draw"
	"time"

	"github.com/MaaXYZ/maa-framework-go/v4"
	"github.com/rs/zerolog/log"
)

const (
	activityListWidth         = 383
	activityListHeight        = 130
	activityListInset         = 20
	activityListSwipeDistance = 400
	activityListSwipeDuration = 300 * time.Millisecond
	activityListSwipeEndHold  = 200 * time.Millisecond
	activityListProbeTemplate = "DailyRewards/DailyLogin/TwinkleGreenStar.png"
	activityListEndTolerance  = 8
	activityListMaxScrolls    = 10
)

var (
	activityListContentROI     = maa.Rect{447, 142, 383, 520}
	activityListScrollProbeROI = maa.Rect{471, 562, 181, 30}
)

type OpenActivityListRecognizerRunner struct{}

type OpenActivityListRecognizerParam struct {
	Template []string `json:"template"`
}

var _ maa.CustomRecognitionRunner = &OpenActivityListRecognizerRunner{}

// Run finds a red dot, derives the activity card to its left, and returns its inset box when the configured template matches.
func (r *OpenActivityListRecognizerRunner) Run(ctx *maa.Context, arg *maa.CustomRecognitionArg) (*maa.CustomRecognitionResult, bool) {
	var param OpenActivityListRecognizerParam
	if err := json.Unmarshal([]byte(arg.CustomRecognitionParam), &param); err != nil {
		log.Error().Err(err).Msg("failed to unmarshal daily login activity list recognition parameters")
		return nil, false
	}
	if len(param.Template) == 0 {
		log.Error().Msg("daily login activity list recognition requires at least one template")
		return nil, false
	}

	img := arg.Img
	for scrollCount := 0; ; scrollCount++ {
		if activityListROI, found := findActivityListInImage(ctx, arg.Roi, param.Template, img); found {
			return activityListRecognitionResult(activityListROI), true
		}
		if scrollCount == activityListMaxScrolls {
			log.Debug().Int("max_scrolls", activityListMaxScrolls).Msg("daily login activity list search reached the scroll limit")
			return nil, false
		}

		scrollProbe, ok := cropImage(img, activityListScrollProbeROI)
		if !ok {
			log.Error().Interface("roi", activityListScrollProbeROI).Msg("daily login scroll probe is outside the screenshot")
			return nil, false
		}

		begin, end := activityListSwipePoints(arg.Roi)
		if _, err := ctx.RunActionDirect(maa.ActionTypeSwipe, maa.SwipeParam{
			Begin: maa.NewTargetRect(begin),
			End: []maa.Target{
				maa.NewTargetRect(end),
			},
			Duration: []time.Duration{activityListSwipeDuration},
			EndHold:  []time.Duration{activityListSwipeEndHold},
		}, arg.Roi, nil); err != nil {
			log.Error().Err(err).Msg("failed to swipe daily login activity list")
			return nil, false
		}

		var captured bool
		img, captured = captureActivityListScreenshot(ctx)
		if !captured {
			return nil, false
		}
		if atEnd, err := activityListAtEnd(ctx, scrollProbe, img); err != nil {
			log.Error().Err(err).Msg("failed to determine whether daily login activity list reached the end")
			return nil, false
		} else if atEnd {
			log.Debug().Msg("daily login activity list reached the end")
			return nil, false
		}
	}

}

func captureActivityListScreenshot(ctx *maa.Context) (image.Image, bool) {
	controller := ctx.GetTasker().GetController()
	if !controller.PostScreencap().Wait().Success() {
		log.Error().Msg("failed to capture screenshot during daily login activity list swipe")
		return nil, false
	}
	img, err := controller.CacheImage()
	if err != nil {
		log.Error().Err(err).Msg("failed to get screenshot during daily login activity list swipe")
		return nil, false
	}
	return img, true
}

func findActivityListInImage(ctx *maa.Context, redDotROI maa.Rect, templates []string, img image.Image) (maa.Rect, bool) {
	redDotResult, err := ctx.RunRecognitionDirect(maa.RecognitionTypeTemplateMatch, maa.TemplateMatchParam{
		ROI: maa.NewTargetRect(redDotROI),
		Template: []string{
			"Common/RedDot.png",
		},
	}, img)
	if err != nil || redDotResult.Results == nil {
		if err != nil {
			log.Error().Err(err).Msg("failed to find red dot during daily login activity list swipe")
		}
		return maa.Rect{}, false
	}

	for _, redDot := range redDotResult.Results.Filtered {
		redDotMatch, ok := redDot.AsTemplateMatch()
		if !ok {
			continue
		}
		activityListROI := activityListROIFromRedDot(redDotMatch.Box)
		templateResult, err := ctx.RunRecognitionDirect(maa.RecognitionTypeTemplateMatch, maa.TemplateMatchParam{
			ROI:      maa.NewTargetRect(activityListROI),
			Template: templates,
		}, img)
		if err != nil {
			log.Error().Err(err).Msg("failed to match daily login activity template during swipe")
			return maa.Rect{}, false
		}
		if templateResult.Results == nil || templateResult.Results.Best == nil {
			continue
		}
		templateMatch, ok := templateResult.Results.Best.AsTemplateMatch()
		if !ok {
			log.Warn().Msg("unexpected non-template result while matching daily login activity")
			continue
		}
		return templateMatch.Box, true
	}

	return maa.Rect{}, false
}

func activityListAtEnd(ctx *maa.Context, scrollProbe, img image.Image) (bool, error) {
	if err := ctx.OverrideImage(activityListProbeTemplate, scrollProbe); err != nil {
		return false, err
	}

	probeResult, err := ctx.RunRecognitionDirect(maa.RecognitionTypeTemplateMatch, maa.TemplateMatchParam{
		ROI: maa.NewTargetRect(activityListContentROI),
		Template: []string{
			activityListProbeTemplate,
		},
	}, img)
	if err != nil || probeResult.Results == nil || probeResult.Results.Best == nil {
		return false, err
	}
	probeMatch, ok := probeResult.Results.Best.AsTemplateMatch()
	if !ok {
		return false, nil
	}

	return activityListProbeHasNotMoved(probeMatch.Box), nil
}

func cropImage(img image.Image, roi maa.Rect) (image.Image, bool) {
	cropBounds := image.Rect(roi.X(), roi.Y(), roi.X()+roi.Width(), roi.Y()+roi.Height())
	if !cropBounds.In(img.Bounds()) {
		return nil, false
	}
	crop := image.NewRGBA(image.Rect(0, 0, cropBounds.Dx(), cropBounds.Dy()))
	draw.Draw(crop, crop.Bounds(), img, cropBounds.Min, draw.Src)
	return crop, true
}

func activityListProbeHasNotMoved(probeMatch maa.Rect) bool {
	return absInt(probeMatch.X()-activityListScrollProbeROI.X()) < activityListEndTolerance &&
		absInt(probeMatch.Y()-activityListScrollProbeROI.Y()) < activityListEndTolerance
}

func absInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

func activityListRecognitionResult(templateMatchROI maa.Rect) *maa.CustomRecognitionResult {
	return &maa.CustomRecognitionResult{
		Box:    insetActivityListROI(templateMatchROI),
		Detail: "Daily login activity list opened",
	}
}

func activityListROIFromRedDot(redDotROI maa.Rect) maa.Rect {
	return maa.Rect{
		redDotROI.X() + redDotROI.Width() - activityListWidth,
		redDotROI.Y(),
		activityListWidth,
		activityListHeight,
	}
}

func insetActivityListROI(activityListROI maa.Rect) maa.Rect {
	return maa.Rect{
		activityListROI.X() + activityListInset,
		activityListROI.Y() + activityListInset,
		activityListROI.Width() - activityListInset*2,
		activityListROI.Height() - activityListInset*2,
	}
}

func activityListSwipePoints(roi maa.Rect) (maa.Rect, maa.Rect) {
	x := roi.X() + roi.Width()/2
	beginY := roi.Y() + roi.Height() - 40
	return maa.Rect{x, beginY, 1, 1}, maa.Rect{x, beginY - activityListSwipeDistance, 1, 1}
}
