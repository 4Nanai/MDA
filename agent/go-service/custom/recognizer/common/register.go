package common

import "github.com/MaaXYZ/maa-framework-go/v4"

func Register() {
	maa.AgentServerRegisterCustomRecognition("RaidChallengeButtonUnavailableRecognition", &RaidChallengeButtonUnavailableRecognizerRunner{})
}
