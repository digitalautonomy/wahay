package gui

import (
	"fmt"
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
			addNewStartupError(fmt.Errorf("the Mumble client can not be used because: %s", c.GetLastError()))
			return
		}

		u.client = c

		u.client.Log()
	}()
}

func (u *gtkUI) launchMumbleClient(data hosting.MeetingData, f func()) (tor.Service, error) {
	s, err := client.LaunchClient(data, f)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (h *hostData) switchToHostOnFinishMeeting() {
	h.u.doInUIThread(func() {
		h.next()
		h.next = func() {}
	})
}

func (u *gtkUI) switchContextWhenMumbleFinish() {
	u.hideCurrentWindow()
	u.openMainWindow()
}
