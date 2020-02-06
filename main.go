package main

import (
	"github.com/coyim/gotk3adapter/gdka"
	"github.com/coyim/gotk3adapter/gliba"
	"github.com/coyim/gotk3adapter/gtka"
	"github.com/digitalautonomy/wahay/config"
	"github.com/digitalautonomy/wahay/gui"
	log "github.com/sirupsen/logrus"
)

func initializeLogging() {
	log.SetLevel(log.InfoLevel)
	if *config.Debug {
		log.SetLevel(log.DebugLevel)
	}
	if *config.Trace {
		log.SetLevel(log.TraceLevel)
	}
	log.SetReportCaller(*config.DebugFunctionCalls)
}

func main() {
	config.ProcessCommandLineArguments()

	initializeLogging()

	runClient()
}

func runClient() {
	g := gui.CreateGraphics(gtka.Real, gliba.Real, gdka.Real)
	gui.NewGTK(g).Loop()
}
