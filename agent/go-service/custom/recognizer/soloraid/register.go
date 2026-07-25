package soloraid

import "github.com/MaaXYZ/maa-framework-go/v4"

func Register() {
	maa.AgentServerRegisterCustomRecognition("SoloRaidSwitchTeam", &SoloRaidSwitchTeamRunner{})
}
