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
	result, err := ctx.RunRecognitionDirect(maa.RecognitionTypeOCR, maa.OCRParam{
		ROI:      maa.NewTargetRect(arg.Roi),
		Expected: []string{param.Expected},
	}, arg.Img)
	if err != nil {
		return nil, false
	}
	log.Debug().Interface("ocr result box", result.Box).Msg("OCR recognition completed in SoloRaidRecognition")
	return &maa.CustomRecognitionResult{
		Box:    result.Box,
		Detail: result.Name,
	}, true
}
