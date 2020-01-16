package gui

import (
	"fmt"
	"strings"

	"github.com/coyim/gotk3adapter/gtki"
)

// SignalErrorsUpdated is the signal group for errors
const SignalErrorsUpdated = "errors-updated"

// Error is a representation of multiple errors
type Error struct {
	list []string
}

func (e *Error) append(err string) {
	e.list = append(e.list, err)
}

func (e *Error) all() string {
	return strings.Join(e.list, "\n")
}

func (e *Error) empty() bool {
	return len(e.list) == 0
}

func fatal(v interface{}) {
	panic(fmt.Sprintf("failing on error: %v", v))
}

func fatalf(format string, v ...interface{}) {
	panic(fmt.Sprintf(format, v...))
}

func (u *gtkUI) newError(err string, emit bool) {
	u.status.errors.append(err)

	// Send the signal
	if emit {
		u.status.Emit(SignalErrorsUpdated, err)
	}
}

func (u *gtkUI) reportError(message string) {
	builder := u.g.uiBuilderFor("GeneralError")
	dlg := builder.get("dialog").(gtki.MessageDialog)

	err := dlg.SetProperty("text", message)
	if err != nil {
		u.reportError(fmt.Sprintf("Programmer error #1: %s", err.Error()))
	}

	dlg.SetTransientFor(u.currentWindow)
	u.doInUIThread(func() {
		dlg.Run()
		dlg.Destroy()
	})
}
