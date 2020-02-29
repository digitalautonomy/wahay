package gui

import "github.com/coyim/gotk3adapter/gtki"

func (u *gtkUI) openHelpWindow() {
	builder := u.g.uiBuilderFor("Help")

	builder.i18nProperties(
		"label", "lblWhatIsWahay",
		"label", "lblDescWhatIsWahay",
		"label", "lblDescWhatIsWahay_2",
		"label", "lblWhatIsTor",
		"label", "lblDescWhatIsTor",
		"label", "lblWhatIsMumble",
		"label", "lblDescWhatIsMumble",
	)

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
