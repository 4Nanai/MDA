package daily_reward

import (
	"encoding/json"
	"image"
	"os"
	"time"

	"github.com/MaaXYZ/maa-framework-go/v4"
	"github.com/rs/zerolog/log"
)

type DailyRewardsPassClickRunner struct{}

type DailyRewardsPassClickParam struct {
	CurrentPassParam PassRecognitionParam `json:"current_pass_param"`
	NextPassParam    PassRecognitionParam `json:"next_pass_param"`
}

type PassRecognitionParam struct {
	ROI      maa.Rect `json:"roi"`
	Template []string `json:"template"`
}

func (r *DailyRewardsPassClickRunner) Run(ctx *maa.Context, arg *maa.CustomActionArg) bool {
	var customParam DailyRewardsPassClickParam
	if err := json.Unmarshal([]byte(arg.CustomActionParam), &customParam); err != nil {
		return false
	}
	log.Debug().Interface("current pass params", customParam.CurrentPassParam).Msg("Running DailyRewardsPassClickRunner")
	log.Debug().Interface("next pass params", customParam.NextPassParam).Msg("Running DailyRewardsPassClickRunner")
	log.Debug().Interface("roi", arg.Box).Msg("Running DailyRewardsPassClickRunner")
	actionResult, err := ctx.RunActionDirect(maa.ActionTypeScreencap, maa.ScreencapParam{}, arg.Box, nil)
	if err != nil {
		log.Error().Err(err).Msg("Failed to run screencap action in DailyRewardsPassClickRunner")
		return false
	}
	screencapResult, ok := actionResult.Result.AsScreencap()
	if !ok {
		log.Error().Msg("Failed to get screencap result in DailyRewardsPassClickRunner")
		return false
	}
	imgFilepath := screencapResult.Filepath
	imgFile, err := os.Open(imgFilepath)
	if err != nil {
		log.Error().Err(err).Str("filepath", imgFilepath).Msg("Failed to open screencap image in DailyRewardsPassClickRunner")
		return false
	}
	defer imgFile.Close()
	img, _, err := image.Decode(imgFile)
	if err != nil {
		log.Error().Err(err).Str("filepath", imgFilepath).Msg("Failed to decode screencap image in DailyRewardsPassClickRunner")
		return false
	}

	currentPassRecoResult, err := ctx.RunRecognitionDirect(maa.RecognitionTypeTemplateMatch, maa.TemplateMatchParam{
		ROI:      maa.NewTargetRect(customParam.CurrentPassParam.ROI),
		Template: customParam.CurrentPassParam.Template,
	}, img)
	if err != nil {
		log.Error().Err(err).Msg("Failed to run recognition in DailyRewardsPassClickRunner")
		return false
	}
	if currentPassRecoResult.Results.Best != nil {
		log.Debug().Interface("box", currentPassRecoResult.Box).Msg("Current pass has reward, clicking the box in DailyRewardsPassClickRunner")
		ctx.RunActionDirect(maa.ActionTypeClick, maa.ClickParam{
			Target:       maa.NewTargetRect(currentPassRecoResult.Box),
			TargetOffset: maa.Rect{-20, 20, 0, 0},
		}, currentPassRecoResult.Box, nil)
		return true
	}
	nextPassRecoResult, err := ctx.RunRecognitionDirect(maa.RecognitionTypeTemplateMatch, maa.TemplateMatchParam{
		ROI:      maa.NewTargetRect(customParam.NextPassParam.ROI),
		Template: customParam.NextPassParam.Template,
	}, img)
	if err != nil {
		log.Error().Err(err).Msg("Failed to run recognition for next pass in DailyRewardsPassClickRunner")
		return false
	}
	if nextPassRecoResult.Results.Best != nil {
		log.Debug().Interface("box", nextPassRecoResult.Box).Msg("Next pass has reward, click switch button first.")
		ctx.RunActionDirect(maa.ActionTypeClick, maa.ClickParam{
			Target:       maa.NewTargetRect(nextPassRecoResult.Box),
			TargetOffset: maa.Rect{-5, 5, 0, 0},
		}, nextPassRecoResult.Box, nil)
		time.Sleep(200 * time.Millisecond)

		ctx.RunActionDirect(maa.ActionTypeClick, maa.ClickParam{
			Target:       maa.NewTargetRect(nextPassRecoResult.Box),
			TargetOffset: maa.Rect{30, 5, 0, 0},
		}, nextPassRecoResult.Box, nil)
		return true
	}
	return false
}
