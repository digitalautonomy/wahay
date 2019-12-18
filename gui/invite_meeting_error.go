package gui

import "github.com/coyim/gotk3adapter/gtki"

func (u *gtkUI) openErrorDialog() {
	builder := u.g.uiBuilderFor("InviteCodeErrorWindow")
	win := builder.get("inviteCodeErrorWindow").(gtki.ApplicationWindow)
	win.SetApplication(u.app)

	builder.ConnectSignals(map[string]interface{}{
		"on_ok": func() {
			win.Hide()
		},
	})
	win.ShowAll()
}
