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
	Lower    int                     `json:"lower"`
	Upper    int                     `json:"upper"`
}
type LargeEventStoryPriority int

const (
	StoryI LargeEventStoryPriority = iota
	StoryII
)

const CustomRecognitionResultDetail string = "Custom large event story recognizer result"

var _ maa.CustomRecognitionRunner = &LargeEventStoryRecognizer{}

func (r *LargeEventStoryRecognizer) Run(ctx *maa.Context, arg *maa.CustomRecognitionArg) (*maa.CustomRecognitionResult, bool) {
	var param LargeEventStoryRecognizerParam
	if err := json.Unmarshal([]byte(arg.CustomRecognitionParam), &param); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal custom recognition param for LargeEventStoryRecognizer")
		return nil, false
	}

	ocrResult, err := ctx.RunRecognitionDirect(maa.RecognitionTypeOCR, maa.OCRParam{
		ROI:      maa.NewTargetRect(arg.Roi),
		Expected: param.Expected,
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
		log.Error().Msg("Multiple OCR results found in LargeEventStoryRecognizer, more than 2, returning false")
		maafocus.Print(ctx, "Multiple OCR results found in LargeEventStoryRecognizer, more than 2, this is unexpected, please check the OCR results and the expected text settings")
		return nil, false
	} else if param.Priority == StoryI {
		log.Info().Msg("Multiple OCR results found in LargeEventStoryRecognizer, priority is StoryI, returning the first one")
		storyI, _ := ocrResult.Results.Filtered[0].AsOCR()

		return &maa.CustomRecognitionResult{
			Box:    storyI.Box,
			Detail: CustomRecognitionResultDetail,
		}, true
	}

	storyIRecoResult, _ := ocrResult.Results.Filtered[0].AsOCR()
	storyIIRecoResult, _ := ocrResult.Results.Filtered[1].AsOCR()

	colorRecoResult, err := ctx.RunRecognitionDirect(maa.RecognitionTypeColorMatch, maa.ColorMatchParam{
		ROI:    maa.NewTargetRect(storyIIRecoResult.Box),
		Method: maa.ColorMatchMethodGRAY,
		Lower: [][]int{
			{param.Lower},
		},
		Upper: [][]int{
			{param.Upper},
		},
	}, arg.Img)
	if err != nil {
		log.Error().Err(err).Msg("Failed to run color match recognition in LargeEventStoryRecognizer")
		return nil, false
	}
	if colorRecoResult.Results.Best == nil {
		log.Info().Msg("No color match result found for STORY II in LargeEventStoryRecognizer, fallback to STORY I")
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
