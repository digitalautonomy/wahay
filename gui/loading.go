package gui

import "github.com/coyim/gotk3adapter/gtki"

func (u *gtkUI) displayLoadingWindow() {
	if u.loadingWindow == nil {
		builder := u.g.uiBuilderFor("LoadingWindow")
		builder.i18nProperties("label", "lblLoading")
		win := builder.get("loadingWindow").(gtki.ApplicationWindow)
		u.loadingWindow = win
		win.SetApplication(u.app)
		u.doInUIThread(win.Show)
	}
}

func (u *gtkUI) hideLoadingWindow() {
	if u.loadingWindow != nil {
		u.doInUIThread(u.loadingWindow.Hide)
		u.loadingWindow = nil
	}
}
