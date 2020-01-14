package main

import (
	"autonomia.digital/tonio/app/gui"
	"autonomia.digital/tonio/app/tor"
	"github.com/coyim/gotk3adapter/gdka"
	"github.com/coyim/gotk3adapter/gliba"
	"github.com/coyim/gotk3adapter/gtka"
)

func main() {
	tor.LaunchTorInstance()

	runClient()
}

func runClient() {
	g := gui.CreateGraphics(gtka.Real, gliba.Real, gdka.Real)
	gui.NewGTK(g).Loop()
}
