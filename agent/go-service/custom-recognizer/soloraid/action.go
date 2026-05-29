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

var _ maa.CustomRecognitionRunner = &SoloRaidRecognitionRunner{}

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
	if err != nil || textResult.Results.Best == nil {
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
	if err != nil || redDotResult.Results.Best == nil {
		return nil, false
	}
	log.Debug().Interface("red dot recognition box", redDotResult.Box).Interface("red dot recognition best result", redDotResult.Results.Best).Msg("Red dot recognition completed in SoloRaidRecognition")
	clickBox := maa.Rect{redDotResult.Box.X() - 18, redDotResult.Box.Y() + 16, 8, 8}
	return &maa.CustomRecognitionResult{
		Box:    clickBox,
		Detail: "SoloRaid recognized, click the red dot",
	}, true
}
