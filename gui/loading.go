package gui

import "github.com/coyim/gotk3adapter/gtki"

func (u *gtkUI) displayLoadingWindowWithCallback(cb func()) {
	u.displayLoadingWindowHelper(cb)
}

func (u *gtkUI) displayLoadingWindow() {
	u.displayLoadingWindowHelper(nil)
}

func (u *gtkUI) hideLoadingWindow() {
	if u.loadingWindow == nil {
		return
	}

	u.doInUIThread(u.loadingWindow.Hide)
	u.loadingWindow = nil
}

func (u *gtkUI) displayLoadingWindowHelper(cb func()) {
	if u.loadingWindow != nil {
		return
	}

	builder := u.g.uiBuilderFor("LoadingWindow")
	builder.i18nProperties("label", "lblLoading")

	win := builder.get("loadingWindow").(gtki.ApplicationWindow)

	if cb != nil {
		builder.ConnectSignals(map[string]interface{}{
			"on_close": cb,
		})
	} else {
		_ = win.SetProperty("deletable", false)
	}

	win.SetApplication(u.app)
	u.loadingWindow = win
	u.doInUIThread(win.Show)
}
