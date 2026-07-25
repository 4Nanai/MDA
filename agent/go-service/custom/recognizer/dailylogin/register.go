package dailylogin

import "github.com/MaaXYZ/maa-framework-go/v4"

func Register() {
	maa.AgentServerRegisterCustomRecognition("DailyLoginOpenActivityListRecognition", &OpenActivityListRecognizerRunner{})
}
