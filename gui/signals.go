package gui

func (u *gtkUI) quit() {
	u.cleanup()
	u.app.Quit()
}

func (u *gtkUI) cleanup() {
	if u.tor != nil {
		// TODO: delete any created Onion Service
		u.tor.Destroy()
	}

	if u.client != nil {
		// TODO: we should remove any Mumble command running
		// and we should close the Grumble service if it's running
		u.client.Destroy()
	}
}
