package gui

import (
	"os"
	"runtime"
	"time"

	"autonomia.digital/tonio/app/config"
	"autonomia.digital/tonio/app/hosting"
	"autonomia.digital/tonio/app/tor"
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
	loadingWindow    gtki.ApplicationWindow
	g                Graphics
	serverCollection hosting.Servers
	tor              tor.Instance
	config           *config.ApplicationConfig
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
		tor: nil,
	}

	return ret
}

func (u *gtkUI) onActivate() {
	u.setGlobalStyles()
	u.displayLoadingWindow()
	go u.loadConfig()
	go u.ensureTonioNetwork(func(startupSuccess bool) {
		u.doInUIThread(func() {
			u.createMainWindow(startupSuccess)
		})
	})
}

func (u *gtkUI) createMainWindow(startupSuccess bool) {
	builder := u.g.uiBuilderFor("MainWindow")
	win := builder.get("mainWindow").(gtki.ApplicationWindow)
	u.currentWindow = win
	u.mainWindow = win

	win.SetApplication(u.app)

	builder.ConnectSignals(map[string]interface{}{
		"on_close_window_signal": u.quit,
		"on_host_meeting":        u.hostMeetingHandler,
		"on_join_meeting":        u.joinMeeting,
		"on_open_settings": func() {
			u.openSettingsWindow()
		},
		"on_show_errors": func() {
			u.showStatusErrorsWindow(builder)
		},
		"on_close_window_errors": func() {
			u.currentWindow.Hide()
		},
	})

	if !startupSuccess {
		u.updateStatusBar(builder)
		u.disableControls(builder)
	}

	win.Show()
}

func (u *gtkUI) updateStatusBar(builder *uiBuilder) {
	lblAppStatus := builder.get("lblApplicationStatus").(gtki.Label)
	btnStatusShow := builder.get("btnStatusShowErrors").(gtki.Button)

	if u.tor == nil {
		lblAppStatus.SetLabel("We've found errors")
		btnStatusShow.SetVisible(true)
	}
}

func (u *gtkUI) disableControls(builder *uiBuilder) {
	btnHostMeeting := builder.get("btnHostMeeting").(gtki.Button)
	btnJoinMeeting := builder.get("btnJoinMeeting").(gtki.Button)

	if u.tor == nil {
		btnHostMeeting.SetSensitive(false)
		btnJoinMeeting.SetSensitive(false)
		btnHostMeeting.SetTooltipText("You can't host a meeting without Tor")
		btnJoinMeeting.SetTooltipText("You can't join a meeting without Tor")
	}
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

func (u *gtkUI) switchToMainWindow() {
	u.switchToWindow(u.mainWindow)
}

func (u *gtkUI) switchToWindow(win gtki.ApplicationWindow) {
	u.currentWindow = win
	win.SetApplication(u.app)
	u.doInUIThread(win.Show)
}

func (u *gtkUI) quit() {
	u.cleanUp()
	u.app.Quit()
}

func (u *gtkUI) copyToClipboard(text string) error {
	return clipboard.WriteAll(text)
}

func (u *gtkUI) messageToLabel(label gtki.Label, message string, seconds int) {
	label.SetVisible(true)
	label.SetText(message)
	time.Sleep(time.Duration(seconds) * time.Second)
	label.SetText("")
	label.SetVisible(false)
}

func (u *gtkUI) initialSetupWindow() {
	u.saveConfigOnly()
}

func (u *gtkUI) cleanUp() {
	if u.tor != nil {
		u.tor.Destroy()
	}

	// TODO: delete our onion service if created
	// TODO: close our mumble service if running
}
