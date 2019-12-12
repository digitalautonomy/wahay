package gui

import (
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

// CreateGraphics creates a Graphic represention from the given arguments
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
		panic(err)
	}

	ret := &gtkUI{
		app: app,
		g:   gx,
	}

	return ret
}

func (u *gtkUI) onActivate() {
	u.mainWindow()
}

func (u *gtkUI) mainWindow() {
	builder, err := u.g.gtk.BuilderNew()
	if err != nil {
		panic(err)
	}

	err = builder.AddFromString("<interface>" +
		"  <object class=\"GtkApplicationWindow\" id=\"mainWindow\">" +
		"    <property name=\"can_focus\">False</property>" +
		"    <property name=\"title\">Tonio!</property>" +
		"    <property name=\"default_width\">600</property>" +
		"    <property name=\"default_height\">400</property>" +
		"  </object>" +
		"</interface>")

	obj, _ := builder.GetObject("mainWindow")

	win := obj.(gtki.ApplicationWindow)
	win.SetApplication(u.app)

	win.ShowAll()
}

func (u *gtkUI) Loop() {
	u.app.Connect("activate", u.onActivate)
	u.app.Run([]string{})
}
