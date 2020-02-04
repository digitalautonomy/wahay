package gui

import (
	"fmt"
	"sync"

	"autonomia.digital/tonio/app/client"
	"autonomia.digital/tonio/app/hosting"
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

type mumbleService interface {
	client.Service
}

func (u *gtkUI) launchMumbleClient(data hosting.MeetingData, f func()) (mumbleService, error) {
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
