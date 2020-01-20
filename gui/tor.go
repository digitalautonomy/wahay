package gui

import (
	"errors"

	"autonomia.digital/tonio/app/tor"
	"github.com/coyim/gotk3adapter/gtki"
)

func (u *gtkUI) ensureTonioNetwork(cb func(bool)) {
	ch := make(chan bool)
	go u.ensureTorNetwork(ch, cb)
	<-ch
	u.hideLoadingWindow()
}

// TODO: we should also check that either Torify or Torsocks are available
func (u *gtkUI) ensureTorNetwork(ch chan bool, cb func(bool)) {
	instance, err := tor.GetSystem()
	ch <- true
	if instance != nil && err == nil {
		u.tor = instance
		cb(true)
	} else {
		cb(false)
		u.displayStartupError(err)
	}
}

func (u *gtkUI) showStatusErrorsWindow(builder *uiBuilder) {
	win := builder.get("mainWindowErrors").(gtki.Dialog)
	txt := builder.get("textContent").(gtki.Label)
	txt.SetMarkup("Show the startup or Tor-related errors")
	u.currentWindow = win
	win.Show()
}

func (u *gtkUI) throughTor(command string, args []string) (*tor.RunningCommand, error) {
	if u.tor == nil {
		return nil, errors.New("no configured Tor found in the system")
	}

	return u.tor.Exec(command, args)
}
