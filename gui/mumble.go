package gui

import (
	"sync"

	"github.com/digitalautonomy/wahay/client"
	"github.com/digitalautonomy/wahay/hosting"
	"github.com/digitalautonomy/wahay/tor"
)

func (u *gtkUI) ensureMumble(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()

		c := client.InitSystem(u.config)

		if !c.CanBeUsed() {
			addNewStartupError(c.GetLastError(), errGroupMumble)
			return
		}

		u.client = c

		u.client.Log()
	}()
}

func (u *gtkUI) launchMumbleClient(data hosting.MeetingData, onClose func()) (tor.Service, error) {
	s, err := client.LaunchClient(data, onClose)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (u *gtkUI) switchContextWhenMumbleFinish() {
	u.hideCurrentWindow()
	u.switchToMainWindow()
}

const errGroupMumble errGroupType = "mumble"

func init() {
	initStartupErrorGroup(errGroupMumble, parseMumbleError)
}

func parseMumbleError(err error) string {
	return i18n.Sprintf("the Mumble client can not be used because: %s", err.Error())
}
