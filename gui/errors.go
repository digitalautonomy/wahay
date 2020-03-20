package gui

import (
	"fmt"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/coyim/gotk3adapter/gtki"
)

func fatal(v interface{}) {
	panic(fmt.Sprintf("failing on error: %v", v))
}

func fatalf(format string, v ...interface{}) {
	panic(fmt.Sprintf(format, v...))
}

func (u *gtkUI) reportError(message string) {
	// TODO: this should only be logged as debug
	log.Printf("reportError(%s)", message)

	builder := u.g.uiBuilderFor("GeneralError")
	builder.i18nProperties(
		"text", "dialog",
		"secondary_text", "dialog")

	dlg := builder.get("dialog").(gtki.MessageDialog)

	err := dlg.SetProperty("text", i18n.Sprintf("Error"))
	if err != nil {
		panic(fmt.Sprintf("Programmer error #1: %s", err.Error()))
	}

	err = dlg.SetProperty("secondary-text", message)
	if err != nil {
		panic(fmt.Sprintf("Programmer error #2: %s", err.Error()))
	}

	if u.currentWindow != nil {
		dlg.SetTransientFor(u.currentWindow)
	}

	u.doInUIThread(func() {
		dlg.Run()
		dlg.Present()
		dlg.Destroy()
	})
}

func getStatusErrorsText() string {
	if !isThereAnyStartupError() {
		return "" // nothing to show
	}

	txt := []string{}
	for _, v := range startupErrors {
		txt = append(txt, v.errorList...)
	}

	return strings.Join(txt, "\n")
}

func (u *gtkUI) showStatusErrorsWindow(builder *uiBuilder) {
	txt := getStatusErrorsText()

	dialog := builder.get("mainWindowErrors").(gtki.Dialog)
	buffer := builder.get("helpTextBuffer").(gtki.TextBuffer)
	buffer.SetText(txt)

	if u.mainWindow != nil {
		dialog.SetTransientFor(u.mainWindow)
	}

	u.disableMainWindow()
	u.currentWindow = dialog
	dialog.Present()
	dialog.Show()
}

func (u *gtkUI) closeStatusErrorsWindow() {
	u.enableMainWindow()
	if u.currentWindow != nil {
		u.currentWindow.Hide()
	}
}

type errGroupType string

// TODO[OB]: this is not a parser, so this naming is a bit confusing.
type errGroupParser func(err error) string
type errGroupData struct {
	errorList []string
	parser    errGroupParser
}

// TODO[OB]: It seems like all these things, including the muxes and the flag
// should be in its own struct instad of floating free like this.

var startupErrors = map[errGroupType]*errGroupData{}

func initStartupErrorGroup(group errGroupType, parser errGroupParser) {
	if _, ok := startupErrors[group]; ok {
		return
	}

	startupErrors[group] = &errGroupData{
		errorList: []string{},
		parser:    parser,
	}
}

func isThereAnyStartupError() bool {
	return weHaveErrors
}

var startupErrorsIlock sync.Mutex
var weHaveErrors bool

func addNewStartupError(err error, group errGroupType) {
	startupErrorsIlock.Lock()
	defer startupErrorsIlock.Unlock()

	weHaveErrors = true

	startupErrors[group].errorList = append(
		startupErrors[group].errorList,
		startupErrors[group].parser(err),
	)
}
