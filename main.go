package main

import (
	"autonomia.digital/tonio/app/config"
	"autonomia.digital/tonio/app/gui"
	"github.com/coyim/gotk3adapter/gliba"
	"github.com/coyim/gotk3adapter/gtka"
)

func main() {
	config.ProcessCommandLineArguments()
	runClient()
}

func runClient() {
	g := gui.CreateGraphics(gtka.Real, gliba.Real)
	gui.NewGTK(g).Loop()
}
