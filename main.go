package main

import (
	"fmt"

	"github.com/coyim/gotk3adapter/gdka"
	"github.com/coyim/gotk3adapter/gliba"
	"github.com/coyim/gotk3adapter/gtka"
	"github.com/digitalautonomy/wahay/config"
	"github.com/digitalautonomy/wahay/gui"
	log "github.com/sirupsen/logrus"
)

// BuildCommit contains which commit the build was based on
var BuildCommit = "UNKNOWN"

// BuildShortCommit contains which commit in short format the build was based on
var BuildShortCommit = "UNKNOWN"

// BuildTag contains which tag - if any - the build was based on
var BuildTag = "(no tag)"

// BuildTimestamp contains the timestamp in Ecuador time zone when the build was made
var BuildTimestamp = "UNKNOWN"

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

	if *config.Version {
		fmt.Printf("Wahay (commit: %s (%s) tag: %s built: %s)\n", BuildShortCommit, BuildCommit, BuildTag, BuildTimestamp)
		return
	}

	initializeLogging()

	runClient()
}

func runClient() {
	g := gui.CreateGraphics(gtka.Real, gliba.Real, gdka.Real)
	gui.NewGTK(g).Loop()
}
