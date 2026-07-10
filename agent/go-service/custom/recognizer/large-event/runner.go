package largeevent

import (
	"encoding/json"
	"fmt"
	"image"
	"time"

	"github.com/1204244136/MDA/agent/go-service/pkg/maafocus"
	"github.com/MaaXYZ/maa-framework-go/v4"
	"github.com/rs/zerolog/log"
)

type LargeEventStoryRecognizer struct{}
type LargeEventStoryRecognizerParam struct {
	Expected []string                `json:"expected"`
	Priority LargeEventStoryPriority `json:"priority"`
	Template []string                `json:"template"`
}
type LargeEventStoryPriority int

const (
	StoryI LargeEventStoryPriority = iota
	StoryII
)

const CustomRecognitionResultDetail string = "Custom large event story recognizer result"

var _ maa.CustomRecognitionRunner = &LargeEventStoryRecognizer{}

type LargeEventMissionCompletedRecognizer struct{}
type LargeEventMissionCompletedParam struct {
	Expected []string `json:"expected"`
	Template []string `json:"template"`
}

func (r *LargeEventStoryRecognizer) Run(ctx *maa.Context, arg *maa.CustomRecognitionArg) (*maa.CustomRecognitionResult, bool) {
	var param LargeEventStoryRecognizerParam
	if err := json.Unmarshal([]byte(arg.CustomRecognitionParam), &param); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal custom recognition param for LargeEventStoryRecognizer")
		return nil, false
	}

	ocrResult, err := ctx.RunRecognitionDirect(maa.RecognitionTypeOCR, maa.OCRParam{
		ROI:      maa.NewTargetRect(arg.Roi),
		Expected: param.Expected,
		OrderBy:  maa.OCROrderByVertical,
	}, arg.Img)
	if err != nil {
		log.Error().Err(err).Msg("Failed to run OCR recognition in LargeEventStoryRecognizer")
		return nil, false
	}
	if len(ocrResult.Results.Filtered) == 0 {
		log.Info().Msg("No OCR result found in LargeEventStoryRecognizer")
		return nil, false
	} else if len(ocrResult.Results.Filtered) == 1 {
		log.Info().Msg("One OCR result found in LargeEventStoryRecognizer, returning it")
		return &maa.CustomRecognitionResult{
			Box:    ocrResult.Box,
			Detail: CustomRecognitionResultDetail,
		}, true
	} else if len(ocrResult.Results.Filtered) > 2 {
		log.Warn().Int("count", len(ocrResult.Results.Filtered)).Msg("Multiple OCR results found in LargeEventStoryRecognizer, more than 2, this is unexpected")
		maafocus.Print(ctx, "More than 2 'story' found in large event, this is unexpected, please check the OCR results and the expected text settings")
	}

	storyIRecoResult, ok := ocrResult.Results.Filtered[0].AsOCR()
	if !ok {
		log.Error().Msg("Failed to get OCR result for STORY I in LargeEventStoryRecognizer")
		return nil, false
	}
	storyIIRecoResult, ok := ocrResult.Results.Filtered[1].AsOCR()
	if !ok {
		log.Error().Msg("Failed to get OCR result for STORY II in LargeEventStoryRecognizer")
		return nil, false
	}

	extendedStoryIBox := extendBox(storyIRecoResult.Box, -48, -9, 48, 9)
	storyIColorRecoResult, err := runStoryTemplateMatch(ctx, extendedStoryIBox, param.Template, arg.Img)
	if err != nil {
		log.Error().Err(err).Msg("Failed to run template match recognition for STORY I in LargeEventStoryRecognizer")
		return nil, false
	}
	if storyIColorRecoResult.Results.Best == nil {
		log.Info().Msg("No template match result found for STORY I in LargeEventStoryRecognizer, retrying up to 2 times at 1s intervals")
		for attempt := 2; attempt <= 3; attempt++ {
			time.Sleep(time.Second)

			retryImg, err := captureImage(ctx)
			if err != nil {
				log.Error().Err(err).Int("attempt", attempt).Msg("Failed to capture image for STORY I retry in LargeEventStoryRecognizer")
				return nil, false
			}
			storyIColorRecoResult, err = runStoryTemplateMatch(ctx, extendedStoryIBox, param.Template, retryImg)
			if err != nil {
				log.Error().Err(err).Int("attempt", attempt).Msg("Failed to retry template match recognition for STORY I in LargeEventStoryRecognizer")
				return nil, false
			}
			if storyIColorRecoResult.Results.Best != nil {
				break
			}
			log.Info().Int("attempt", attempt).Msg("No template match result found for STORY I attempt in LargeEventStoryRecognizer")
		}
		if storyIColorRecoResult.Results.Best == nil {
			log.Info().Msg("No template match result found for STORY I after 3 attempts in LargeEventStoryRecognizer")
			return nil, false
		}
	}

	time.Sleep(500 * time.Millisecond)
	storyIIImg, err := captureImage(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to capture image for STORY II in LargeEventStoryRecognizer")
		return nil, false
	}
	extendedStoryIIBox := extendBox(storyIIRecoResult.Box, -48, -9, 48, 9)
	storyIIColorRecoResult, err := runStoryTemplateMatch(ctx, extendedStoryIIBox, param.Template, storyIIImg)
	if err != nil {
		log.Error().Err(err).Msg("Failed to run template match recognition for STORY II in LargeEventStoryRecognizer")
		return nil, false
	}
	if storyIIColorRecoResult.Results.Best == nil {
		log.Info().Msg("No template match result found for STORY II in LargeEventStoryRecognizer, returning STORY II OCR result")
		return &maa.CustomRecognitionResult{
			Box:    storyIRecoResult.Box,
			Detail: CustomRecognitionResultDetail + " (STORY II template not found)",
		}, true
	}

	if param.Priority == StoryI {
		log.Info().Msg("Template match results found for both stories, priority is StoryI")
		return &maa.CustomRecognitionResult{
			Box:    storyIColorRecoResult.Box,
			Detail: CustomRecognitionResultDetail,
		}, true
	}

	log.Info().Msg("Template match results found for both stories, priority is StoryII")
	return &maa.CustomRecognitionResult{
		Box:    storyIIColorRecoResult.Box,
		Detail: CustomRecognitionResultDetail,
	}, true
}

func runStoryTemplateMatch(ctx *maa.Context, roi maa.Rect, templates []string, img image.Image) (*maa.RecognitionDetail, error) {
	return ctx.RunRecognitionDirect(maa.RecognitionTypeTemplateMatch, maa.TemplateMatchParam{
		ROI:      maa.NewTargetRect(roi),
		Template: templates,
	}, img)
}

func captureImage(ctx *maa.Context) (image.Image, error) {
	controller := ctx.GetTasker().GetController()
	if !controller.PostScreencap().Wait().Success() {
		return nil, fmt.Errorf("failed to capture a new frame")
	}
	return controller.CacheImage()
}

func (r *LargeEventMissionCompletedRecognizer) Run(ctx *maa.Context, arg *maa.CustomRecognitionArg) (*maa.CustomRecognitionResult, bool) {
	var param LargeEventMissionCompletedParam
	if err := json.Unmarshal([]byte(arg.CustomRecognitionParam), &param); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal custom recognition param for LargeEventMissionCompletedRecognizer")
		return nil, false
	}

	ocrResult, err := ctx.RunRecognitionDirect(maa.RecognitionTypeOCR, maa.OCRParam{
		ROI:      maa.NewTargetRect(arg.Roi),
		Expected: param.Expected,
	}, arg.Img)
	if err != nil {
		log.Error().Err(err).Msg("Failed to run OCR recognition in LargeEventMissionCompletedRecognizer")
		return nil, false
	}
	if len(ocrResult.Results.Filtered) == 0 {
		log.Error().Msg("No OCR result found in LargeEventMissionCompletedRecognizer, returning false")
		return nil, false
	}

	redDotRecoResult, err := ctx.RunRecognitionDirect(maa.RecognitionTypeTemplateMatch, maa.TemplateMatchParam{
		ROI:      maa.NewTargetRect(arg.Roi),
		Template: param.Template,
	}, arg.Img)
	if err != nil {
		log.Error().Err(err).Msg("Failed to run template match recognition in LargeEventMissionCompletedRecognizer")
		return nil, false
	}
	if redDotRecoResult.Results.Best != nil {
		log.Info().Msg("Template match result found in LargeEventMissionCompletedRecognizer, marked as not completed")
		return nil, false
	}
	return &maa.CustomRecognitionResult{
		Box: arg.Roi,
	}, true
}

func extendBox(box maa.Rect, deltaX, deltaY, deltaWidth, deltaHeight int) maa.Rect {
	return maa.Rect{
		box.X() + deltaX,
		box.Y() + deltaY,
		box.Width() + deltaWidth,
		box.Height() + deltaHeight,
	}
}
