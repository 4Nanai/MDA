package largeevent

import (
	"encoding/json"

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
		OrderBy: maa.OCROrderByVertical,
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

	if param.Priority == StoryI {
		log.Info().Msg("Multiple OCR results found in LargeEventStoryRecognizer, priority is StoryI, returning the first one")
		storyI, _ := ocrResult.Results.Filtered[0].AsOCR()

		return &maa.CustomRecognitionResult{
			Box:    storyI.Box,
			Detail: CustomRecognitionResultDetail,
		}, true
	}

	// Priority is Story II
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
	extendedStoryIIBox := extendBox(storyIIRecoResult.Box, -48, -9, 48, 9)

	colorRecoResult, err := ctx.RunRecognitionDirect(maa.RecognitionTypeTemplateMatch, maa.TemplateMatchParam{
		ROI:      maa.NewTargetRect(extendedStoryIIBox),
		Template: param.Template,
	}, arg.Img)
	if err != nil {
		log.Error().Err(err).Msg("Failed to run template match recognition in LargeEventStoryRecognizer")
		return nil, false
	}
	if colorRecoResult.Results.Best == nil {
		log.Info().Msg("No template match result found for STORY II in LargeEventStoryRecognizer, fallback to STORY I")
		return &maa.CustomRecognitionResult{
			Box:    storyIRecoResult.Box,
			Detail: CustomRecognitionResultDetail + " (fallback to STORY I)",
		}, true
	}

	return &maa.CustomRecognitionResult{
		Box:    colorRecoResult.Box,
		Detail: CustomRecognitionResultDetail,
	}, true
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
