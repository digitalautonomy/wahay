package gui

import "github.com/coyim/gotk3adapter/gtki"

func (u *gtkUI) openHelpWindow() {
	builder := u.g.uiBuilderFor("Help")

	dialog := builder.get("helpWindow").(gtki.ApplicationWindow)

	cleanup := func() {
		dialog.Hide()
		u.enableCurrentWindow()
	}

	builder.ConnectSignals(map[string]interface{}{
		"on_close_window_signal": cleanup,
	})

	u.doInUIThread(func() {
		u.disableCurrentWindow()
		dialog.Show()
	})
}
