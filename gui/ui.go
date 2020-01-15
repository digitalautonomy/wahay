package gui

import (
	"context"
	"log"
	"os"
	"os/exec"
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
	g                Graphics
	serverCollection hosting.Servers
	tor              tor.Control

	config *config.ApplicationConfig
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

	go u.loadConfig("")
	go u.ensureTorNetwork()
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
	_ = label.SetProperty("visible", true)
	label.SetText(message)
	time.Sleep(time.Duration(seconds) * time.Second)
	label.SetText("")
	_ = label.SetProperty("visible", false)
}

func (u *gtkUI) loadConfig(configFile string) {
	u.config.WhenLoaded(u.configLoaded)

	var conf *config.ApplicationConfig
	var err error
	conf, err = config.LoadOrCreate(configFile)

	u.config = conf

	if err != nil {
		log.Println("configuration file error:", err.Error())
		u.doInUIThread(u.initialSetupWindow)
		return
	}
}

func (u *gtkUI) configLoaded(c *config.ApplicationConfig) {
	//TODO: do stuffs when config loaded
}

func (u *gtkUI) initialSetupWindow() {
	u.saveConfigOnly()
}

func (u *gtkUI) saveConfigOnlyInternal() error {
	return u.config.Save()
}

func (u *gtkUI) saveConfigOnly() {
	go func() {
		err := u.saveConfigOnlyInternal()
		if err != nil {
			log.Println("Failed to save config file:", err.Error())
		}
	}()
}

func (u *gtkUI) ensureTorNetwork() {
	if !tor.Network.Detect() {
		log.Fatal("tor is not running")
		return
	}

	h := tor.Network.Host()
	p := tor.Network.Port()

	log.Printf("DETECTED TCP HOST: %s\n", h)
	log.Printf("DETECTED TCP PORT: %s\n", p)

	torController := tor.CreateController(h, p, *config.TorControlPassword, "")

	isCompatible, isValid, err := torController.EnsureTorCompatibility()
	if !isCompatible && !isValid {
		log.Fatalf("incompatibility error: %s", err)
	}

	if err != nil {
		log.Println(err)
		instance, err := tor.NewInstance()
		if err != nil {
			log.Println(err)
			return
		}

		// Start our Tor Control Port instance
		err = instance.Start()
		if err != nil {
			log.Println(err)
			return
		}

		// We don't check here the Tor compatibility again because we are using
		// the local Tor for now. Remove this comment or implement this when Tonio
		// has it's own Tor.
		u.tor = tor.CreateController(instance.GetHost(), instance.GetControlPort(), "", instance.GetPreferredAuthType())
		u.tor.SetInstance(instance)
	} else {
		u.tor = torController
	}
}

type runningMumble struct {
	cmd               *exec.Cmd
	ctx               context.Context
	cancelFunc        context.CancelFunc
	finished          bool
	finishedWithError error
	finishChannel     chan bool
}

func (r *runningMumble) close() {
	r.cancelFunc()
}

func (r *runningMumble) waitForFinish() {
	e := r.cmd.Wait()
	r.finished = true
	r.finishedWithError = e
	r.finishChannel <- true
}

func launchMumbleClient(data hosting.MeetingData) (*runningMumble, error) {
	ctx, cancelFunc := context.WithCancel(context.Background())

	cmd := exec.CommandContext(ctx, "torify", "mumble", hosting.GenerateURL(data))
	if err := cmd.Start(); err != nil {
		cancelFunc()
		return nil, err
	}

	state := &runningMumble{
		cmd:               cmd,
		ctx:               ctx,
		cancelFunc:        cancelFunc,
		finished:          false,
		finishedWithError: nil,
		finishChannel:     make(chan bool, 100),
	}

	go state.waitForFinish()

	return state, nil
}

func (u *gtkUI) switchContextWhenMumbleFinished(state *runningMumble) {
	go func() {
		<-state.finishChannel

		// TODO: here, we  could check if the Mumble instance
		// failed with an error and report this
		u.doInUIThread(func() {
			u.openMainWindow()
		})
	}()
}

func (u *gtkUI) cleanUp() {
	u.tor.Close()
	// TODO: delete our onion service if created
	// TODO: close our mumble service if running
}
