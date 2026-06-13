package unionraid

import (
	"encoding/json"

	"github.com/MaaXYZ/maa-framework-go/v4"
	"github.com/rs/zerolog/log"
)

type UnionRaidEntryRecognizerRunner struct{}
type UnionRaidEntryRecognizerParam struct {
	Expected []string `json:"expected"`
	Template []string `json:"template"`
}

type UnionRaidOpenedRecognizerRunner struct{}
type UnionRaidOpenedRecognizerParam struct {
	Expected []string `json:"expected"`
	Template []string `json:"template"`
}

var _ maa.CustomRecognitionRunner = &UnionRaidEntryRecognizerRunner{}
var _ maa.CustomRecognitionRunner = &UnionRaidOpenedRecognizerRunner{}

func (r *UnionRaidEntryRecognizerRunner) Run(ctx *maa.Context, arg *maa.CustomRecognitionArg) (*maa.CustomRecognitionResult, bool) {
	var param UnionRaidEntryRecognizerParam
	if err := json.Unmarshal([]byte(arg.CustomRecognitionParam), &param); err != nil {
		log.Error().Err(err).Msg("failed to unmarshal param")
		return nil, false
	}

	ocrResult, err := ctx.RunRecognitionDirect(maa.RecognitionTypeOCR, &maa.OCRParam{
		Expected: param.Expected,
		ROI:      maa.NewTargetRect(arg.Roi),
	}, arg.Img)
	if err != nil {
		log.Error().Err(err).Msg("failed to run ocr")
		return nil, false
	}
	if ocrResult.Results.Best == nil {
		log.Warn().Msg("ocr result is empty")
		return nil, false
	}

	redDotResult, err := ctx.RunRecognitionDirect(maa.RecognitionTypeTemplateMatch, &maa.TemplateMatchParam{
		Template: param.Template,
		ROI:      maa.NewTargetRect(arg.Roi),
	}, arg.Img)
	if err != nil {
		log.Error().Err(err).Msg("failed to run template match")
		return nil, false
	}
	if redDotResult.Results.Best == nil {
		log.Info().Msg("red dot not found, no union raid opened or already cleared")
		return nil, false
	}
	// If we reach here, it means we found a opened union raid

	// Shrink the box by 10 pixels on each side
	shrinkedBox := maa.Rect{
		arg.Roi.X() + 10,
		arg.Roi.Y() + 10,
		arg.Roi.Width() - 20,
		arg.Roi.Height() - 20,
	}

	return &maa.CustomRecognitionResult{
		Box:    shrinkedBox,
		Detail: "Union Raid Entry Detected",
	}, true
}

func (r *UnionRaidOpenedRecognizerRunner) Run(ctx *maa.Context, arg *maa.CustomRecognitionArg) (*maa.CustomRecognitionResult, bool) {
	var param UnionRaidOpenedRecognizerParam
	if err := json.Unmarshal([]byte(arg.CustomRecognitionParam), &param); err != nil {
		log.Error().Err(err).Msg("failed to unmarshal param")
		return nil, false
	}

	ocrResult, err := ctx.RunRecognitionDirect(maa.RecognitionTypeOCR, &maa.OCRParam{
		Expected: param.Expected,
		ROI:      maa.NewTargetRect(arg.Roi),
	}, arg.Img)
	if err != nil {
		log.Error().Err(err).Msg("failed to run ocr")
		return nil, false
	}
	if ocrResult.Results.Best == nil {
		log.Warn().Msg("ocr result is empty")
		return nil, false
	}

	redDotResult, err := ctx.RunRecognitionDirect(maa.RecognitionTypeTemplateMatch, &maa.TemplateMatchParam{
		Template: param.Template,
		ROI:      maa.NewTargetRect(arg.Roi),
	}, arg.Img)
	if err != nil {
		log.Error().Err(err).Msg("failed to run template match")
		return nil, false
	}
	if redDotResult.Results.Best == nil {
		log.Info().Msg("red dot not found")
		return nil, false
	}

	// Shrink the box by 10 pixels on each side
	shrinkedBox := maa.Rect{
		arg.Roi.X() + 10,
		arg.Roi.Y() + 10,
		arg.Roi.Width() - 20,
		arg.Roi.Height() - 20,
	}

	return &maa.CustomRecognitionResult{
		Box:    shrinkedBox,
		Detail: "Union Raid Opened Detected",
	}, true
}
