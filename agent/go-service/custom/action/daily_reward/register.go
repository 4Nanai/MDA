package daily_reward

import "github.com/MaaXYZ/maa-framework-go/v4"

func Register() {
	maa.AgentServerRegisterCustomAction("DailyRewardsPassClick", &DailyRewardsPassClickRunner{})
}
