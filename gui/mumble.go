package gui

import (
	"errors"
	"sync"

	"github.com/digitalautonomy/wahay/client"
	"github.com/digitalautonomy/wahay/hosting"
	"github.com/digitalautonomy/wahay/tor"
)

func (u *gtkUI) ensureMumble(wg *sync.WaitGroup) {
	wg.Add(1)
	u.waitForTorInstance(func(t tor.Instance) {
		go func() {
			defer wg.Done()

			c := client.InitSystem(u.config, t)

			if !c.IsValid() {
				u.errorHandler.addNewStartupError(c.LastError(), errGroupMumble)
				return
			}

			u.client = c
		}()
	})
}

func (u *gtkUI) launchMumbleClient(data hosting.MeetingData, onClose func()) (tor.Service, error) {
	c := u.client

	if !c.IsValid() {
		return nil, errors.New("error: no client to run")
	}

	return c.Launch(data.GenerateURL(), onClose)
}

func (u *gtkUI) switchContextWhenMumbleFinish() {
	u.hideCurrentWindow()
	u.switchToMainWindow()
}

const errGroupMumble errGroupType = "mumble"

func init() {
	initStartupErrorGroup(errGroupMumble, mumbleErrorTranslator)
}

func mumbleErrorTranslator(err error) string {
	return i18n.Sprintf("the Mumble client can not be used because: %s", err.Error())
}
