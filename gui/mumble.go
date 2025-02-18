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

			u.onExit(c.Destroy)

			u.client = c
		}()
	})
}

func (u *gtkUI) launchMumbleClient(data hosting.MeetingData, onClose func()) (tor.Service, error) {
	c := u.client

	if !c.IsValid() {
		return nil, errors.New("error: no client to run")
	}

	return c.Launch(data, onClose)
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
	switch err {
	case client.ErrNoClientInConfiguredPath:
		return i18n().Sprintf("The configured path to the Mumble binary is not valid or can't be used." +
			" Please configure another path.")
	case client.ErrBinaryUnavailable:
		return i18n().Sprintf("No valid Mumble binary found on the system.")
	}

	return err.Error()
}
