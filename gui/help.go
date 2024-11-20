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
		"label", "lblFunctionalities",
		"label", "lblDescFunctionalities",
		"label", "lblHostMeeting",
		"label", "lblDescHostMeeting1",
		"label", "lblDescHostMeeting2",
		"label", "lblJoinMeeting",
		"label", "lblDescJoinMeeting",
	)

	u.setImage(builder, "help/wahay.svg", "imgWahay")
	u.setImage(builder, "help/wahay_hosting.svg", "imgWahayHosting")
	u.setImage(builder, "help/wahay_join.svg", "imgWahayJoin")

	dialog := builder.get("helpWindow").(gtki.Window)

	builder.ConnectSignals(map[string]interface{}{
		"on_close_window_signal": func() {
			u.closeHelpWindow(dialog)
		},
	})

	u.connectShortcutsHelpWindow(dialog)

	u.doInUIThread(func() {
		u.disableCurrentWindow()
		dialog.Show()
	})
}

func (u *gtkUI) closeHelpWindow(dialog gtki.Window) {
	dialog.Hide()
	u.enableCurrentWindow()
}

func (u *gtkUI) setImage(builder *uiBuilder, filename string, idComponent string) {
	imagePixBuf, _ := u.g.getImagePixbufForSize(filename, 400)
	img := builder.get(idComponent).(gtki.Image)
	img.SetFromPixbuf(imagePixBuf)
}
