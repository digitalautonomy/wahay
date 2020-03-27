package gui

import (
	"errors"
	"sync"

	"github.com/digitalautonomy/wahay/client"
	"github.com/digitalautonomy/wahay/hosting"
	"github.com/digitalautonomy/wahay/tor"

	log "github.com/sirupsen/logrus"
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
	}()
}

func (u *gtkUI) launchMumbleClient(data hosting.MeetingData, onClose func()) (tor.Service, error) {
	c := client.Mumble()

	if !c.CanBeUsed() {
		return nil, errors.New("error: no client to run")
	}

	err := c.LoadCertificateFrom(
		data.MeetingID,
		data.Port,
		data.Cert,
		hosting.DefaultCertificateServerPort)
	if err != nil {
		log.WithFields(log.Fields{
			"serviceID": data.MeetingID,
			"port":      hosting.DefaultCertificateServerPort,
		}).Errorf("No valid Mumble certificate available: %s", err)
		return nil, err
	}

	return c.Execute([]string{data.GenerateURL()}, onClose)
}

func (u *gtkUI) switchContextWhenMumbleFinish() {
	u.hideCurrentWindow()
	u.switchToMainWindow()
}

const errGroupMumble errGroupType = "mumble"

func init() {
	initStartupErrorGroup(errGroupMumble, parseMumbleError)
}

// TODO[OB]: this is definitely not a parser...

func parseMumbleError(err error) string {
	return i18n.Sprintf("the Mumble client can not be used because: %s", err.Error())
}
