package gui

import "github.com/coyim/gotk3adapter/gtki"

func (u *gtkUI) openDialog() {

	builder := u.g.uiBuilderFor("inviteWindow")
	win := builder.get("inviteWindow").(gtki.ApplicationWindow)
	win.SetApplication(u.app)

	/*builder.ConnectSignals(map[string]interface{}{
		"on_host_meeting":    hostMeeting,
		"on_join_meeting":    joinMeeting,
		"on_test_connection": testTorConnection,
		"on_open_settings":   openSettings,
	})*/

	win.ShowAll()

}

