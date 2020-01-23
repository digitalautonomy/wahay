package gui

import (
	"fmt"
	"log"

	"github.com/coyim/gotk3adapter/gtki"
)

func (u *gtkUI) reportError(message string) {
	// TODO: this should only be logged as debug
	log.Printf("reportError(%s)", message)

	builder := u.g.uiBuilderFor("GeneralError")
	dlg := builder.get("dialog").(gtki.MessageDialog)

	err := dlg.SetProperty("text", message)
	if err != nil {
		panic(fmt.Sprintf("Programmer error #1: %s", err.Error()))
	}

	if u.currentWindow != nil {
		dlg.SetTransientFor(u.currentWindow)
	}

	u.doInUIThread(func() {
		dlg.Run()
		dlg.Destroy()
	})
}

func (u *gtkUI) displayStartupError(err error) {
	u.reportError(err.Error())
}

func fatal(v interface{}) {
	panic(fmt.Sprintf("failing on error: %v", v))
}

func fatalf(format string, v ...interface{}) {
	panic(fmt.Sprintf(format, v...))
}
