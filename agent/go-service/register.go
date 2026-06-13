package main

import (
	"github.com/1204244136/MDA/agent/go-service/common/myaction"
	"github.com/1204244136/MDA/agent/go-service/common/myreco"
	"github.com/1204244136/MDA/agent/go-service/custom/action/daily_reward"
	"github.com/1204244136/MDA/agent/go-service/custom/recognizer/common"
	largeevent "github.com/1204244136/MDA/agent/go-service/custom/recognizer/large-event"
	"github.com/1204244136/MDA/agent/go-service/custom/recognizer/soloraid"
	"github.com/1204244136/MDA/agent/go-service/custom/recognizer/unionraid"
	"github.com/1204244136/MDA/agent/go-service/pkg/resource"
	"github.com/1204244136/MDA/agent/go-service/taskersink/aspectratio"
	"github.com/1204244136/MDA/agent/go-service/taskersink/hdrcheck"
	"github.com/1204244136/MDA/agent/go-service/taskersink/membership"
	"github.com/1204244136/MDA/agent/go-service/taskersink/processcheck"
	"github.com/rs/zerolog/log"
)

func registerAll() {
	// Resource Sink
	resource.EnsureResourcePathSink()

	// Pre-Check Custom
	aspectratio.Register()
	hdrcheck.Register()
	processcheck.Register()
	membership.Register()

	// Custom Actions
	myaction.Register()

	// Custom Recognitions
	common.Register()
	myreco.Register()
	soloraid.Register()
	daily_reward.Register()
	largeevent.Register()
	unionraid.Register()

	log.Info().
		Msg("All custom components and sinks registered successfully")
}
