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

	u.setImage(builder, "help/wahay.png", "imgWahay")
	u.setImage(builder, "help/wahay_hosting.png", "imgWahayHosting")
	u.setImage(builder, "help/wahay_join.png", "imgWahayJoin")

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

func (u *gtkUI) setImage(builder *uiBuilder, filename string, idComponent string) {
	imagePixBuf, _ := u.g.getImagePixbufForSize(filename, 400)
	img := builder.get(idComponent).(gtki.Image)
	img.SetFromPixbuf(imagePixBuf)
}
