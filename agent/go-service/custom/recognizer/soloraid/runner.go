package soloraid

import (
	"github.com/MaaXYZ/maa-framework-go/v4"
	"github.com/rs/zerolog/log"
)

type SoloRaidSwitchTeamRunner struct{}

var _ maa.CustomRecognitionRunner = &SoloRaidSwitchTeamRunner{}

type RelativeHorizontalPosition int

const (
	Left RelativeHorizontalPosition = iota
	Right
	Inside
)

func (r *SoloRaidSwitchTeamRunner) Run(ctx *maa.Context, arg *maa.CustomRecognitionArg) (*maa.CustomRecognitionResult, bool) {
	colorMatchResult, err := ctx.RunRecognitionDirect(maa.RecognitionTypeColorMatch, maa.ColorMatchParam{
		ROI: maa.NewTargetRect(arg.Roi),
		Lower: [][]int{
			{0, 161, 234},
		},
		Upper: [][]int{
			{13, 182, 254},
		},
	}, arg.Img)
	if err != nil {
		return nil, false
	}
	if colorMatchResult.Results.Best == nil {
		return nil, false
	}

	teamSetResult, err := ctx.RunRecognitionDirect(maa.RecognitionTypeOCR, maa.OCRParam{
		ROI:     maa.NewTargetRect(arg.Roi),
		OrderBy: maa.OCROrderByHorizontal,
	}, arg.Img)
	if err != nil {
		return nil, false
	}
	if teamSetResult.Results.All == nil {
		return nil, false
	}

	log.Debug().Interface("OCR box", teamSetResult.Box).Interface("ocr all results", teamSetResult.Results.All).Msg("Recognition completed in SoloRaidSwitchTeam")
	if len(teamSetResult.Results.All) != 5 {
		log.Warn().Msg("Expected to find 5 team sets in SoloRaidSwitchTeam, but found a different number")
	}
	ocrResults := make([]*maa.OCRResult, 0, 5)
	for index, reco := range teamSetResult.Results.All {
		ocrResult, ok := reco.AsOCR()
		if !ok {
			return nil, false
		}
		log.Debug().Int("index", index).Interface("ocr result", ocrResult).Msg("Processing OCR result in SoloRaidSwitchTeam")
		ocrResults = append(ocrResults, ocrResult)
	}
	for index, ocrResult := range ocrResults {
		switch boxRelativeHorizontalPosition(colorMatchResult.Box, ocrResult.Box) {
		case Inside:
			log.Debug().Int("index", index).Interface("ocr box", ocrResult.Box).Msg("Found current selected team set, select next one")
			return &maa.CustomRecognitionResult{
				Box:    ocrResults[(index+1)%len(ocrResults)].Box,
				Detail: "SoloRaid switch to next team set",
			}, true
		default:
			log.Debug().Int("index", index).Interface("ocr box", ocrResult.Box).Msg("This team set is not selected")
		}
	}
	return nil, false
}

func boxRelativeHorizontalPosition(anchor, target maa.Rect) RelativeHorizontalPosition {
	if target.X()+target.Width() < anchor.X() {
		return Left
	} else if target.X() > anchor.X()+anchor.Width() {
		return Right
	} else if target.X() >= anchor.X() && target.X()+target.Width() <= anchor.X()+anchor.Width() {
		return Inside
	} else {
		return -1
	}
}
