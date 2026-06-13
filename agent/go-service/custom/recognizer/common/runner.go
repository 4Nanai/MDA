package common

import (
	"encoding/json"

	"github.com/MaaXYZ/maa-framework-go/v4"
	"github.com/rs/zerolog/log"
)

type RaidChallengeButtonUnavailableRecognizerRunner struct{}
type RaidChallengeButtonUnavailableRecognizerParam struct {
	Expected []string `json:"expected"`
	Lower    []int    `json:"lower"`
	Upper    []int    `json:"upper"`
}

var _ maa.CustomRecognitionRunner = &RaidChallengeButtonUnavailableRecognizerRunner{}

func (r *RaidChallengeButtonUnavailableRecognizerRunner) Run(ctx *maa.Context, arg *maa.CustomRecognitionArg) (*maa.CustomRecognitionResult, bool) {
	var param RaidChallengeButtonUnavailableRecognizerParam
	if err := json.Unmarshal([]byte(arg.CustomRecognitionParam), &param); err != nil {
		log.Error().Err(err).Msg("failed to unmarshal param")
		return nil, false
	}
	log.Debug().Interface("param", param).Msg("Running RaidChallengeButtonUnavailableRecognition with param")

	ocrResult, err := ctx.RunRecognitionDirect(maa.RecognitionTypeOCR, &maa.OCRParam{
		Expected: param.Expected,
		ROI:      maa.NewTargetRect(arg.Roi),
	}, arg.Img)
	if err != nil {
		log.Error().Err(err).Msg("failed to run ocr")
		return nil, false
	}
	if ocrResult.Results.Best == nil {
		log.Info().Msg("Could not find expected challenge button")
		return nil, false
	}
	log.Debug().Interface("ocrResult box", ocrResult.Box).Msg("OCR result for RaidChallengeButtonUnavailableRecognition")

	colorMatchResult, err := ctx.RunRecognitionDirect(maa.RecognitionTypeColorMatch, &maa.ColorMatchParam{
		ROI:   maa.NewTargetRect(ocrResult.Box),
		Lower: [][]int{param.Lower},
		Upper: [][]int{param.Upper},
	}, arg.Img)
	if err != nil {
		log.Error().Err(err).Msg("failed to run color match")
		return nil, false
	}
	if colorMatchResult.Results.Best == nil {
		log.Info().Msg("Challenge button available")
		return nil, false
	}
	log.Debug().Interface("colorMatchResult box", colorMatchResult.Box).Msg("Color match result for RaidChallengeButtonUnavailableRecognition")

	return &maa.CustomRecognitionResult{
		Box: ocrResult.Box,
	}, true
}
