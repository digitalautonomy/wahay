package gui

import (
	"os"
	"runtime"

	"github.com/coyim/gotk3adapter/gdki"
	"github.com/coyim/gotk3adapter/glibi"
	"github.com/coyim/gotk3adapter/gtki"
	"github.com/digitalautonomy/wahay/client"
	"github.com/digitalautonomy/wahay/config"
	"github.com/digitalautonomy/wahay/hosting"
	"github.com/digitalautonomy/wahay/tor"
)

const (
	programName   = "Wahay"
	applicationID = "digital.autonomia.Wahay"
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
	tor              tor.Instance
	client           client.Instance
	serverCollection hosting.Servers
	keySupplier      config.KeySupplier
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
	}

	// Creates the encryption key suplier for all the crypto-related
	// functionalities of the configuration package
	ret.keySupplier = config.CreateKeySupplier(ret.getMasterPassword)

	// Ensure all components that should be installed
	ret.ensureInstallation()

	return ret
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

func (u *gtkUI) onActivate() {
	u.displayLoadingWindow()

	go u.setGlobalStyles()
	go u.loadConfig()
}

func (u *gtkUI) configLoaded() {
	u.displayLoadingWindow()

	go u.initLogs()

	go u.ensureDependencies(func(success bool) {
		u.hideLoadingWindow()

		u.doInUIThread(func() {
			u.createMainWindow(success)
		})
	})
}

func (u *gtkUI) getMainWindowBuilder() *uiBuilder {
	builder := u.g.uiBuilderFor("MainWindow")

	builder.i18nProperties(
		"label", "lblWelcome",
		"button", "btnSettings",
		"button", "btnJoinMeeting",
		"tooltip", "btnJoinMeeting",
		"button", "btnHostMeeting",
		"tooltip", "btnHostMeeting",
		"label", "lblApplicationStatus",
		"button", "btnStatusShowErrors",
		"button", "btnErrorsAccept")

	return builder
}

func (u *gtkUI) createMainWindow(success bool) {
	builder := u.getMainWindowBuilder()
	win := builder.get("mainWindow").(gtki.ApplicationWindow)
	u.currentWindow = win
	u.mainWindow = win

	win.SetApplication(u.app)

	win.SetIcon(getApplicationIcon().getPixbuf())
	u.g.gtk.WindowSetDefaultIcon(getApplicationIcon().getPixbuf())

	builder.ConnectSignals(map[string]interface{}{
		"on_close_window_signal": u.quit,
		"on_host_meeting":        u.hostMeetingHandler,
		"on_join_meeting":        u.joinMeeting,
		"on_open_settings":       u.openSettingsWindow,
		"on_show_errors": func() {
			u.showStatusErrorsWindow(builder)
		},
		"on_close_window_errors": func() {
			u.closeStatusErrorsWindow()
		},
	})

	if !success {
		u.updateMainWindowStatusBar(builder)
		u.disableMainWindowControls(builder)
	}

	win.Show()
}

func (u *gtkUI) updateMainWindowStatusBar(builder *uiBuilder) {
	lblAppStatus := builder.get("lblApplicationStatus").(gtki.Label)
	btnStatusShow := builder.get("btnStatusShowErrors").(gtki.Button)

	box := builder.get("boxApplicationStatus").(gtki.Widget)
	cntx, err := box.GetStyleContext()

	if weHaveStartupErrors() {
		if err == nil {
			cntx.AddClass("error")
		}
		lblAppStatus.SetLabel(i18n.Sprintf("We've found errors"))
		btnStatusShow.SetVisible(true)
	}
}

func (u *gtkUI) disableMainWindowControls(builder *uiBuilder) {
	btnHostMeeting := builder.get("btnHostMeeting").(gtki.Button)
	btnJoinMeeting := builder.get("btnJoinMeeting").(gtki.Button)

	btnHostMeeting.SetSensitive(false)
	btnJoinMeeting.SetSensitive(false)
}

func (u *gtkUI) setGlobalStyles() {
	if u.g.gdk == nil {
		return
	}

	prov := u.g.cssFor("gui")
	screen, _ := u.g.gdk.ScreenGetDefault()
	u.g.gtk.AddProviderForScreen(screen, prov, uint(gtki.STYLE_PROVIDER_PRIORITY_APPLICATION))
}

func (u *gtkUI) initialSetupWindow() {
	u.saveConfigOnly()
}

func (u *gtkUI) getConfirmWindow() *uiBuilder {
	builder := u.g.uiBuilderFor("Confirm")

	builder.i18nProperties(
		"title", "dialog",
		"label", "lblTitle",
		"label", "lblText",
		"button", "btnCancel",
		"button", "btnConfirm",
	)

	return builder
}

func (u *gtkUI) showConfirmation(onConfirm func(bool), text string) {
	u.disableCurrentWindow()

	builder := u.getConfirmWindow()
	dialog := builder.get("dialog").(gtki.Window)

	if u.currentWindow != nil {
		dialog.SetTransientFor(u.currentWindow)
	}

	if len(text) > 0 {
		lbl, _ := builder.get("lblText").(gtki.Label)
		lbl.SetText(text)
	}

	clean := func(op bool) {
		dialog.Destroy()
		u.enableCurrentWindow()
		onConfirm(op)
	}

	builder.ConnectSignals(map[string]interface{}{
		"on_cancel": func() {
			clean(false)
		},
		"on_confirm": func() {
			clean(true)
		},
	})

	dialog.Present()
	dialog.Show()
}

func (u *gtkUI) cleanUp() {
	if u.tor != nil {
		// TODO: delete any created Onion Service
		u.tor.Destroy()
	}

	if u.client != nil {
		// TODO: we should remove any Mumble command running
		// and we should close the Grumble service if it's running
		u.client.Destroy()
	}
}

func (u *gtkUI) closeApplication() {
	u.quit()
}

func (u *gtkUI) quit() {
	u.cleanUp()
	u.app.Quit()
}
