package gui

func (u *gtkUI) openErrorDialog(message string) {
	finalMessage := message
	if finalMessage == "" {
		finalMessage = "The meeting ID is invalid"
	}
	u.reportError(finalMessage)
	// builder := u.g.uiBuilderFor("InviteCodeErrorWindow")
	// win := builder.get("inviteCodeErrorWindow").(gtki.ApplicationWindow)
	// win.SetApplication(u.app)

	// builder.ConnectSignals(map[string]interface{}{
	// 	"on_ok": func() {
	// 		win.Hide()
	// 	},
	// })
	// win.ShowAll()
}
