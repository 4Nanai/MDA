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

type CommonOCRRedDotRecognizerRunner struct{}

var _ maa.CustomRecognitionRunner = &CommonOCRRedDotRecognizerRunner{}

// CommonOCRRedDotRecognizerRunner finds expected text and then verifies the red dot immediately above it.
func (r *CommonOCRRedDotRecognizerRunner) Run(ctx *maa.Context, arg *maa.CustomRecognitionArg) (*maa.CustomRecognitionResult, bool) {
	var param struct {
		Expected []string `json:"expected"`
	}
	if err := json.Unmarshal([]byte(arg.CustomRecognitionParam), &param); err != nil {
		log.Error().Err(err).Msg("failed to unmarshal param")
		return nil, false
	}
	if len(param.Expected) == 0 {
		log.Error().Msg("expected text is required")
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
		return nil, false
	}

	ocrBox := ocrResult.Box
	redDotROI := maa.NewTargetRect(maa.Rect{ocrBox.X(), ocrBox.Y() - 60, ocrBox.Width() + 20, ocrBox.Height() + 60})
	redDotResult, err := ctx.RunRecognitionDirect(maa.RecognitionTypeTemplateMatch, &maa.TemplateMatchParam{
		ROI:      redDotROI,
		Template: []string{"Common/RedDot.png"},
	}, arg.Img)
	if err != nil {
		log.Error().Err(err).Msg("failed to run red dot template match")
		return nil, false
	}
	if redDotResult.Results.Best == nil {
		return nil, false
	}

	return &maa.CustomRecognitionResult{
		Box:    maa.Rect{redDotResult.Box.X() - 18, redDotResult.Box.Y() + 16, 8, 8},
		Detail: "OCR text and red dot recognized",
	}, true
}

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
		Expected   []string `json:"expected"`
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

type CommonTemplateColorMatchRecognizerRunner struct{}

var _ maa.CustomRecognitionRunner = &CommonTemplateColorMatchRecognizerRunner{}

func (r *CommonTemplateColorMatchRecognizerRunner) Run(ctx *maa.Context, arg *maa.CustomRecognitionArg) (*maa.CustomRecognitionResult, bool) {
	var param struct {
		Template  []string             `json:"template"`
		Lower     [][]int              `json:"lower"`
		Upper     [][]int              `json:"upper"`
		Method    maa.ColorMatchMethod `json:"method"`
		Threshold []float64            `json:"threshold"`
		Count     int                  `json:"count"`
	}

	if err := json.Unmarshal([]byte(arg.CustomRecognitionParam), &param); err != nil {
		log.Error().Err(err).Msg("failed to unmarshal param")
		return nil, false
	}
	log.Debug().Interface("param", param).Msg("Running CommonTemplateColorMatchRecognition with param")

	templateMatchResult, err := ctx.RunRecognitionDirect(maa.RecognitionTypeTemplateMatch, &maa.TemplateMatchParam{
		ROI:       maa.NewTargetRect(arg.Roi),
		Template:  param.Template,
		Threshold: param.Threshold,
	}, arg.Img)
	if err != nil {
		log.Error().Err(err).Msg("failed to run template match")
		return nil, false
	}
	if templateMatchResult.Results.Best == nil {
		log.Info().Msg("Could not find expected template")
		return nil, false
	}
	log.Debug().Interface("templateMatchResult box", templateMatchResult.Box).Msg("Template match result for CommonTemplateMatchWithColorMatchRecognition")

	method := param.Method
	if param.Method == 0 {
		method = maa.ColorMatchMethodRGB
	}
	colorMatchResult, err := ctx.RunRecognitionDirect(maa.RecognitionTypeColorMatch, &maa.ColorMatchParam{
		ROI:    maa.NewTargetRect(templateMatchResult.Box),
		Lower:  param.Lower,
		Upper:  param.Upper,
		Method: method,
		Count:  param.Count,
	}, arg.Img)
	if err != nil {
		log.Error().Err(err).Msg("failed to run color match")
		return nil, false
	}
	if colorMatchResult.Results.Best == nil {
		log.Info().Msg("Could not find expected color match")
		return nil, false
	}
	log.Debug().Interface("colorMatchResult box", colorMatchResult.Box).Msg("Color match result for CommonTemplateMatchWithColorMatchRecognition")

	return &maa.CustomRecognitionResult{
		Box:    templateMatchResult.Box,
		Detail: "Color match result for CommonTemplateMatchWithColorMatchRecognition",
	}, true
}
