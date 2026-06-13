package unionraid

import "github.com/MaaXYZ/maa-framework-go/v4"

func Register() {
	maa.AgentServerRegisterCustomRecognition("UnionRaidEntryRecognition", &UnionRaidEntryRecognizerRunner{})
	maa.AgentServerRegisterCustomRecognition("UnionRaidOpenedRecognition", &UnionRaidOpenedRecognizerRunner{})
}
