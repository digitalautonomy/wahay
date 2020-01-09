package gui

import (
	"os"
	"runtime"
	"time"

	"autonomia.digital/tonio/app/hosting"
	"github.com/atotto/clipboard"
	"github.com/coyim/gotk3adapter/gdki"
	"github.com/coyim/gotk3adapter/glibi"
	"github.com/coyim/gotk3adapter/gtki"
)

const (
	programName   = "Tonio"
	applicationID = "digital.autonomia.Tonio"
)

// Graphics represent the graphic configuration
type Graphics struct {
	gtk  gtki.Gtk
	gdk  gdki.Gdk
	glib glibi.Glib
}

// CreateGraphics creates a Graphic representation from the given arguments
func CreateGraphics(gtkVal gtki.Gtk, glibVal glibi.Glib, gdkVal gdki.Gdk) Graphics {
	return Graphics{
		gtk:  gtkVal,
		gdk:  gdkVal,
		glib: glibVal,
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
	app              gtki.Application
	mainWindow       gtki.ApplicationWindow
	currentWindow    gtki.ApplicationWindow
	g                Graphics
	serverCollection hosting.Servers
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
	u.setGlobalStyles()
}

func (u *gtkUI) createMainWindow() {
	builder := u.g.uiBuilderFor("MainWindow")
	win := builder.get("mainWindow").(gtki.ApplicationWindow)
	u.currentWindow = win
	u.mainWindow = win
	win.SetApplication(u.app)

	builder.ConnectSignals(map[string]interface{}{
		"on_close_window_signal": u.quit,
		"on_host_meeting":        u.hostMeetingHandler,
		"on_join_meeting":        u.joinMeeting,
	})

	win.ShowAll()
}

func (u *gtkUI) setGlobalStyles() {
	if u.g.gdk == nil {
		return
	}
	prov := u.g.cssFor("gui")
	screen, _ := u.g.gdk.ScreenGetDefault()
	u.g.gtk.AddProviderForScreen(screen, prov, uint(gtki.STYLE_PROVIDER_PRIORITY_APPLICATION))
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

func (u *gtkUI) joinMeeting() {
	u.currentWindow.Hide()
	u.openJoinWindow()
}

func (u *gtkUI) switchToWindow(win gtki.ApplicationWindow) {
	u.currentWindow = win
	win.SetApplication(u.app)
	u.doInUIThread(win.ShowAll)
}

func (u *gtkUI) quit() {
	u.app.Quit()
}

func (u *gtkUI) copyToClipboard(text string) error {
	return clipboard.WriteAll(text)
}

func (u *gtkUI) messageToLabel(label gtki.Label, message string, seconds int) {
	label.SetText(message)
	time.Sleep(time.Duration(seconds) * time.Second)
	label.SetText("")
}
