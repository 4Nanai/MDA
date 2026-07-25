package common

import "github.com/MaaXYZ/maa-framework-go/v4"

func Register() {
	maa.AgentServerRegisterCustomRecognition("RaidChallengeButtonUnavailableRecognition", &RaidChallengeButtonUnavailableRecognizerRunner{})
	maa.AgentServerRegisterCustomRecognition("CommonOCRRedDotRecognition", &CommonOCRRedDotRecognizerRunner{})
	maa.AgentServerRegisterCustomRecognition("CommonTemplateOCRRecognition", &CommonTemplateOCRRecognizerRunner{})
	maa.AgentServerRegisterCustomRecognition("CommonWaitingPageLoadRecognition", &CommonWaitingPageLoadRecognizerRunner{})
	maa.AgentServerRegisterCustomRecognition("CommonTemplateColorMatchRecognition", &CommonTemplateColorMatchRecognizerRunner{})
	maa.AgentServerRegisterCustomRecognition("CommonPollingRecognition", &CommonPollingRecognizerRunner{})
}
