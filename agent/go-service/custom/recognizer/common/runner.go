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

type CommonTemplateOCRRecognizerRunner struct{}

var _ maa.CustomRecognitionRunner = &CommonTemplateOCRRecognizerRunner{}

func (r *CommonTemplateOCRRecognizerRunner) Run(ctx *maa.Context, arg *maa.CustomRecognitionArg) (*maa.CustomRecognitionResult, bool) {
	var param struct {
		Expected []string `json:"expected"`
		Template []string `json:"template"`
	}
	if err := json.Unmarshal([]byte(arg.CustomRecognitionParam), &param); err != nil {
		log.Error().Err(err).Msg("failed to unmarshal param")
		return nil, false
	}
	log.Debug().Interface("param", param).Msg("Running CommonTemplateOCRRecognition with param")

	ocrResult, err := ctx.RunRecognitionDirect(maa.RecognitionTypeOCR, &maa.OCRParam{
		Expected: param.Expected,
		ROI:      maa.NewTargetRect(arg.Roi),
	}, arg.Img)
	if err != nil {
		log.Error().Err(err).Msg("failed to run ocr")
		return nil, false
	}
	if ocrResult.Results.Best == nil {
		log.Info().Msg("Could not find expected text")
		return nil, false
	}
	log.Debug().Interface("ocrResult box", ocrResult.Box).Msg("OCR result for CommonTemplateOCRRecognition")

	templateMatchResult, err := ctx.RunRecognitionDirect(maa.RecognitionTypeTemplateMatch, &maa.TemplateMatchParam{
		ROI:      maa.NewTargetRect(arg.Roi),
		Template: param.Template,
	}, arg.Img)
	if err != nil {
		log.Error().Err(err).Msg("failed to run template match")
		return nil, false
	}
	if templateMatchResult.Results.Best == nil {
		log.Info().Msg("Could not find expected template")
		return nil, false
	}
	log.Debug().Interface("templateMatchResult box", templateMatchResult.Box).Msg("Template match result for CommonTemplateOCRRecognition")

	return &maa.CustomRecognitionResult{
		Box:    ocrResult.Box,
		Detail: "Template match result for CommonTemplateOCRRecognition",
	}, true
}

/**
 * CommonWaitingPageLoadRecognition is a custom recognition that
 * checks if a page has finished loading by verifying the presence of expected text
 * and ensuring that a certain number of non-transparent pixels are present in the specified region of interest (ROI).
**/
type CommonWaitingPageLoadRecognizerRunner struct{}

var _ maa.CustomRecognitionRunner = &CommonWaitingPageLoadRecognizerRunner{}

func (r *CommonWaitingPageLoadRecognizerRunner) Run(ctx *maa.Context, arg *maa.CustomRecognitionArg) (*maa.CustomRecognitionResult, bool) {
	var param struct {
		Expected  []string `json:"expected"`
		Limitation int      `json:"limitation"`
	}
	if err := json.Unmarshal([]byte(arg.CustomRecognitionParam), &param); err != nil {
		log.Error().Err(err).Msg("failed to unmarshal param")
		return nil, false
	}
	log.Debug().Interface("param", param).Msg("Running CommonWaitingPageLoadRecognition with param")

	ocrResult, err := ctx.RunRecognitionDirect(maa.RecognitionTypeOCR, &maa.OCRParam{
		Expected: param.Expected,
		ROI:      maa.NewTargetRect(maa.Rect{0, 15, 70, 25}),
	}, arg.Img)
	if err != nil {
		log.Error().Err(err).Msg("failed to run ocr")
		return nil, false
	}
	if ocrResult.Results.Best == nil {
		log.Info().Msg("Could not find expected text")
		return nil, false
	}
	log.Debug().Interface("ocrResult box", ocrResult.Box).Msg("OCR result for CommonWaitingPageLoadRecognition")

	nodeResult, err := ctx.RunRecognitionDirect(maa.RecognitionTypeAnd, &maa.AndRecognitionParam{
		AllOf: []maa.SubRecognitionItem{
			{
				NodeName: "CommonNoTransparentMask",
			},
		},
	}, arg.Img)
	if err != nil {
		log.Error().Err(err).Msg("failed to run node recognition")
		return nil, false
	}
	if len(nodeResult.CombinedResult) != 1 {
		log.Error().Msg("Unexpected: CommonNoTransparentMask recognition failed")
		return nil, false
	}
	bestResult := nodeResult.CombinedResult[0].Results.Best
	if bestResult == nil {
		log.Info().Msg("CommonNoTransparentMask recognition failed")
		return nil, false
	}
	colorMatchResult, ok := bestResult.AsColorMatch()
	if !ok {
		log.Error().Msg("Unexpected: CommonNoTransparentMask recognition result is not ColorMatchResult")
		return nil, false
	}
	if colorMatchResult.Count < param.Limitation {
		log.Info().Msg("CommonNoTransparentMask recognition failed")
		return nil, false
	}

	log.Debug().Interface("colorMatchResult box", colorMatchResult.Box).Msg("Color match result for CommonWaitingPageLoadRecognition")
	
	return &maa.CustomRecognitionResult{
		Box:    ocrResult.Box,
		Detail: "Color match result for CommonWaitingPageLoadRecognition",
	}, true
}
