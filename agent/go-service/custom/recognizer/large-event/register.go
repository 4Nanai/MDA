package largeevent

import "github.com/MaaXYZ/maa-framework-go/v4"

func Register() {
	maa.AgentServerRegisterCustomRecognition("LargeEventStoryRecognition", &LargeEventStoryRecognizer{})
	maa.AgentServerRegisterCustomRecognition("LargeEventMissionCompletedRecognition", &LargeEventMissionCompletedRecognizer{})
}
