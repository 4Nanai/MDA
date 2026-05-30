package soloraid

import (
	"encoding/json"

	"github.com/MaaXYZ/maa-framework-go/v4"
	"github.com/rs/zerolog/log"
)

type SoloRaidRecognitionRunner struct{}

type SoloRaidRecognitionParam struct {
	Expected string `json:"expected"`
}

type SoloRaidSwitchTeamRunner struct{}

var _ maa.CustomRecognitionRunner = &SoloRaidRecognitionRunner{}

var _ maa.CustomRecognitionRunner = &SoloRaidSwitchTeamRunner{}

type RelativeHorizontalPosition int

const (
	Left RelativeHorizontalPosition = iota
	Right
	Inside
)

// SoloRaidRecognitionRunner first runs OCR to find the expected text, then checks if there is a red dot above the text. If both are found, it returns the click box for the red dot.
func (r *SoloRaidRecognitionRunner) Run(ctx *maa.Context, arg *maa.CustomRecognitionArg) (*maa.CustomRecognitionResult, bool) {
	var param SoloRaidRecognitionParam
	if err := json.Unmarshal([]byte(arg.CustomRecognitionParam), &param); err != nil {
		return nil, false
	}
	log.Debug().Str("custom param", arg.CustomRecognitionParam).Interface("roi", arg.Roi).Msg("Running SoloRaidRecognition")
	textResult, err := ctx.RunRecognitionDirect(maa.RecognitionTypeOCR, maa.OCRParam{
		ROI:      maa.NewTargetRect(arg.Roi),
		Expected: []string{param.Expected},
	}, arg.Img)
	if err != nil {
		return nil, false
	}
	// If cannot find expected text, consider recognition failed
	if textResult.Results.Best == nil {
		return nil, false
	}

	log.Debug().Interface("ocr result box", textResult.Box).Interface("ocr best result", textResult.Results.Best).Msg("OCR recognition completed in SoloRaidRecognition")
	textBox := textResult.Box
	redDotRoi := maa.NewTargetRect(maa.Rect{textBox.X(), textBox.Y() - 60, textBox.Width() + 20, textBox.Height() + 60})
	log.Debug().Interface("red dot roi", redDotRoi).Msg("Checking for red dot in SoloRaidRecognition")
	redDotResult, err := ctx.RunRecognitionDirect(maa.RecognitionTypeTemplateMatch, maa.TemplateMatchParam{
		ROI: redDotRoi,
		Template: []string{
			"Common/RedDot.png",
		},
	}, arg.Img)
	if err != nil {
		return nil, false
	}
	// If cannot find red dot, consider task finished
	if redDotResult.Results.Best == nil {
		return nil, false
	}

	log.Debug().Interface("red dot recognition box", redDotResult.Box).Interface("red dot recognition best result", redDotResult.Results.Best).Msg("Red dot recognition completed in SoloRaidRecognition")
	clickBox := maa.Rect{redDotResult.Box.X() - 18, redDotResult.Box.Y() + 16, 8, 8}
	return &maa.CustomRecognitionResult{
		Box:    clickBox,
		Detail: "SoloRaid recognized, click the red dot",
	}, true
}

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
