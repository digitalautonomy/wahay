package gui

import (
	"fmt"
	"os"
	"runtime"

	"github.com/coyim/gotk3adapter/glibi"
	"github.com/coyim/gotk3adapter/gtki"
)

const (
	programName   = "Tonio"
	applicationID = "digital.autonomia.Tonio"
)

// Graphics represent the graphic configuration
type Graphics struct {
	gtk gtki.Gtk
}

// CreateGraphics creates a Graphic representation from the given arguments
func CreateGraphics(gtkVal gtki.Gtk) Graphics {
	return Graphics{
		gtk: gtkVal,
	}
}

// UI is the user interface functionality exposed to main
type UI interface {
	Loop()
}

func argsWithApplicationName() *[]string {
	newSlice := make([]string, len(os.Args))
	copy(newSlice, os.Args)
	newSlice[0] = programName
	return &newSlice
}

type gtkUI struct {
	app gtki.Application
	g   Graphics
}

// NewGTK returns a new client for a GTK ui
func NewGTK(gx Graphics) UI {
	runtime.LockOSThread()
	gx.gtk.Init(argsWithApplicationName())

	app, err := gx.gtk.ApplicationNew(applicationID, glibi.APPLICATION_FLAGS_NONE)
	if err != nil {
		fatalf("Couldn't create application: %v", err)
	}

	ret := &gtkUI{
		app: app,
		g:   gx,
	}

	return ret
}

func (u *gtkUI) onActivate() {
	u.createMainWindow()
}

func (u *gtkUI) createMainWindow() {
	builder := u.g.uiBuilderFor("MainWindow")
	win := builder.get("mainWindow").(gtki.ApplicationWindow)
	win.SetApplication(u.app)

	builder.ConnectSignals(map[string]interface{}{
		"on_host_meeting": hostMeeting,
		"on_join_meeting": func() {
			u.joinMeeting()
		},
		"on_test_connection": testTorConnection,
		"on_open_settings":   openSettings,
	})

	win.ShowAll()
}

func (u *gtkUI) Loop() {
	// This Connect call returns a signal handle, but that's not useful
	// for us, so we ignore it.
	_, err := u.app.Connect("activate", u.onActivate)
	if err != nil {
		fatalf("Couldn't activate application: %v", err)
	}

	u.app.Run([]string{})
}

/*
Event handler functions for main window buttons
TODO: Move to another file and remove from here.
*/

func hostMeeting() {
	fmt.Printf("Clicked on host meeting button!\n")
}

func (u *gtkUI) joinMeeting() {
	fmt.Printf("Clicked on join meeting button!\n")

	builder := u.g.uiBuilderFor("MainWindow")
	win := builder.get("mainWindow").(gtki.ApplicationWindow)
	win.SetApplication(u.app)
	win.Hide()

	u.openDialog()

}

func testTorConnection() {
	fmt.Printf("Clicked on test connection button!\n")
}

func openSettings() {
	fmt.Printf("Clicked on open settings button!\n")
}
