package soloraid

import "github.com/MaaXYZ/maa-framework-go/v4"

type SoloRaidRecognitionRunner struct{}

var _ maa.CustomRecognitionRunner = &SoloRaidRecognitionRunner{}

func (r *SoloRaidRecognitionRunner) Run(ctx *maa.Context, arg *maa.CustomRecognitionArg) (*maa.CustomRecognitionResult, bool) {
	return nil, false
}
