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
	log.Debugf("reportError(%s)", message)

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

func (u *gtkUI) showStatusErrorsWindow(builder *uiBuilder) {
	txt := u.errorHandler.getStatusErrorsText()

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

/* UI error handler utilities */

type errGroupType string

type errGroupTranslator func(err error) string
type errGroupData struct {
	errorList  []string
	translator errGroupTranslator
}

var initStartupErrorsGroups = map[errGroupType]*errGroupData{}

func initStartupErrorGroup(group errGroupType, t errGroupTranslator) {
	if _, ok := initStartupErrorsGroups[group]; ok {
		return
	}

	initStartupErrorsGroups[group] = &errGroupData{
		errorList:  []string{},
		translator: t,
	}
}

type errorHandler struct {
	sync.Mutex
	hasErrors     bool
	startupErrors map[errGroupType]*errGroupData
}

func (u *gtkUI) initErrorsHandler() {
	u.errorHandler = &errorHandler{
		hasErrors:     false,
		startupErrors: initStartupErrorsGroups,
	}
}

func (h *errorHandler) isThereAnyStartupError() bool {
	return h.hasErrors
}

func (h *errorHandler) getStatusErrorsText() string {
	if !h.hasErrors {
		return "" // nothing to show
	}

	txt := []string{}
	for _, v := range h.startupErrors {
		txt = append(txt, v.errorList...)
	}

	return strings.Join(txt, "\n")
}

func (h *errorHandler) addNewStartupError(err error, group errGroupType) {
	h.Lock()
	defer h.Unlock()

	h.hasErrors = true

	h.startupErrors[group].errorList = append(
		h.startupErrors[group].errorList,
		h.startupErrors[group].translator(err),
	)
}
