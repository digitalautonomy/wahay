package gui

import (
	"os"
	"runtime"
	"sync"

	log "github.com/sirupsen/logrus"

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

var g Graphics

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
	app            gtki.Application
	mainWindow     gtki.ApplicationWindow
	currentWindow  gtki.Window
	loadingWindow  gtki.Window
	g              Graphics
	tor            tor.Instance
	torInitialized *sync.WaitGroup
	client         client.Instance
	keySupplier    config.KeySupplier
	config         *config.ApplicationConfig
	servers        hosting.Servers
	errorHandler   *errorHandler
	cleanupHandler *cleanupHandler
	colorManager
}

// NewGTK returns a new client for a GTK ui
func NewGTK(gx Graphics) UI {
	g = gx
	runtime.LockOSThread()
	g.gtk.Init(argsWithApplicationName())

	app, err := g.gtk.ApplicationNew(applicationID, glibi.APPLICATION_FLAGS_NONE)
	if err != nil {
		fatalf("Couldn't create application: %v", err)
	}

	ret := &gtkUI{
		app: app,
		g:   gx,
	}

	ret.initTasks()

	return ret
}

func (u *gtkUI) Loop() {
	// This Connect call returns a signal handle, but that's not useful
	// for us, so we ignore it.
	_ = u.app.Connect("activate", u.onActivate)

	u.app.Run([]string{})
}

func (u *gtkUI) initTasks() {
	u.initCleanupHandler()
	u.initConfig()
	u.initErrorsHandler()
	u.initColorManager()

	u.torInitialized = &sync.WaitGroup{}
	u.torInitialized.Add(1)

	// Creates the encryption key suplier for all the crypto-related
	// functionalities of the configuration package
	u.keySupplier = config.CreateKeySupplier(u.getMasterPassword)

	u.ensureInstallation()
}

func (u *gtkUI) onActivate() {
	u.displayLoadingWindowWithCallback(u.quit)
	go func() {
		u.loadConfig()
		u.setGlobalStyles()
	}()
}

func (u *gtkUI) quit() {
	log.Println("Closing Wahay...")
	u.cleanupHandler.doCleanup(u.app.Quit)
}

func (u *gtkUI) configLoaded() {
	u.displayLoadingWindow()

	go u.initLogs()

	go u.ensureDependencies(func() {
		u.hideLoadingWindow()

		u.doInUIThread(func() {
			u.createMainWindow()
		})
	})
}

func (u *gtkUI) getMainWindowBuilder() *uiBuilder {
	builder := u.g.uiBuilderFor("MainWindow")

	builder.i18nProperties(
		"button", "btnStatusShowErrors",
		"button", "btnErrorsAccept",
		"tooltip", "btnSettings",
		"tooltip", "btnHelp",
		"tooltip", "btnJoinMeeting",
		"tooltip", "btnHostMeeting",
		"label", "lblWelcome",
		"label", "lblApplicationStatus",
		"label", "lblHostMeeting",
		"label", "lblJoinMeeting",
		"label", "lblSettings",
		"label", "lblHelp")

	imgHostMeeting := builder.get("imgHostMeeting").(gtki.Image)
	imgJoinMeeting := builder.get("imgJoinMeeting").(gtki.Image)
	imgSettings := builder.get("imgSettings").(gtki.Image)
	imgHelp := builder.get("imgHelp").(gtki.Image)

	icon1, _ := u.g.getImagePixbufForSize("host-meeting.svg", 32)
	icon2, _ := u.g.getImagePixbufForSize("join-meeting.svg", 32)

	imgHostMeeting.SetFromPixbuf(icon1)
	imgJoinMeeting.SetFromPixbuf(icon2)
	imgHelp.SetFromIconName("help-contents-symbolic", gtki.ICON_SIZE_LARGE_TOOLBAR)
	imgSettings.SetFromIconName("applications-system-symbolic", gtki.ICON_SIZE_LARGE_TOOLBAR)

	return builder
}

func (u *gtkUI) createMainWindow() {
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
		"on_open_help":           u.openHelpWindow,
		"on_show_errors": func() {
			u.showStatusErrorsWindow(builder)
		},
		"on_close_window_errors": func() {
			u.closeStatusErrorsWindow()
		},
	})

	u.connectShortcutsMainWindow(u.currentWindow)

	u.updateMainWindowStatusBar(builder)
	u.disableMainWindowControls(builder)

	win.Show()
}

func (u *gtkUI) updateMainWindowStatusBar(builder *uiBuilder) {
	if !u.errorHandler.isThereAnyStartupError() {
		return // nothing to do
	}

	lblAppStatus := builder.get("lblApplicationStatus").(gtki.Label)
	btnStatusShow := builder.get("btnStatusShowErrors").(gtki.Button)

	box := builder.get("boxApplicationStatus").(gtki.Widget)
	cntx, err := box.GetStyleContext()
	if err != nil {
		log.WithFields(log.Fields{
			"context": "boxApplicationStatus style context",
		}).Debug("programmer error: updateMainWindowStatusBar()")
	} else {
		cntx.AddClass("error")
	}

	lblAppStatus.SetLabel(i18n().Sprintf("We've found errors"))
	btnStatusShow.SetVisible(true)
}

func (u *gtkUI) disableMainWindowControls(builder *uiBuilder) {
	if !u.errorHandler.isThereAnyStartupError() {
		return // nothing to do
	}
	btnHostMeeting := builder.get("btnHostMeeting").(gtki.Button)
	btnJoinMeeting := builder.get("btnJoinMeeting").(gtki.Button)

	btnHostMeeting.SetSensitive(false)
	btnJoinMeeting.SetSensitive(false)
}

func (u *gtkUI) setGlobalStyles() {
	if u.g.gdk == nil {
		return
	}

	configuredTheme := u.config.GetColorScheme()
	if configuredTheme != "" {
		u.colorManager.disableAutomaticThemeChange()
		u.addCSSProvider(configuredTheme)
		return
	}

	css := "light-mode-gui"
	if u.colorManager.isDarkThemeVariant() {
		css = "dark-mode-gui"
	}

	u.addCSSProvider(css)
}

func (u *gtkUI) addCSSProvider(css string) {
	prov := u.g.cssFor(css)
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

func (u *gtkUI) initColorManager() {
	u.colorManager = colorManager{
		ui: u,
	}
	u.colorManager.init()
}
