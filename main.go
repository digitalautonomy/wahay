package main

import (
	"autonomia.digital/tonio/app/config"
	"autonomia.digital/tonio/app/gui"
	"github.com/coyim/gotk3adapter/gdka"
	"github.com/coyim/gotk3adapter/gliba"
	"github.com/coyim/gotk3adapter/gtka"
	log "github.com/sirupsen/logrus"
)

const debug = false
const debugLogSource = false

func initializeLogging() {
	log.SetLevel(log.InfoLevel)
	if debug {
		log.SetLevel(log.DebugLevel)
	}
	log.SetReportCaller(debugLogSource)
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
