package gui

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/coyim/gotk3adapter/gtki"
)

var (
	startupErrors      []string
	startupErrorsIlock sync.Mutex
)

func weHaveStartupErrors() bool {
	return len(startupErrors) != 0
}

func addNewStartupError(err error) {
	startupErrorsIlock.Lock()
	defer startupErrorsIlock.Unlock()

	startupErrors = append(startupErrors, err.Error())
}

func (u *gtkUI) reportError(message string) {
	// TODO: this should only be logged as debug
	log.Printf("reportError(%s)", message)

	builder := u.g.uiBuilderFor("GeneralError")
	dlg := builder.get("dialog").(gtki.MessageDialog)

	err := dlg.SetProperty("text", "Error")
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

func (u *gtkUI) showStatusErrorsWindow(builder *uiBuilder) {
	if len(startupErrors) == 0 {
		return // nothing to show
	}

	win := builder.get("mainWindowErrors").(gtki.Dialog)
	txt := builder.get("textContent").(gtki.Label)
	txt.SetMarkup(strings.Join(startupErrors, "\n"))

	u.currentWindow = win
	win.Show()
}

func fatal(v interface{}) {
	panic(fmt.Sprintf("failing on error: %v", v))
}

func fatalf(format string, v ...interface{}) {
	panic(fmt.Sprintf(format, v...))
}
