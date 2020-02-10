package gui

import (
	"errors"
	"sync"

	"github.com/digitalautonomy/wahay/tor"
)

// TODO: we should also check that either Torify or Torsocks are available
func (u *gtkUI) ensureTor(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()

		instance, e := tor.GetInstance(u.config)
		if e != nil {
			addNewStartupError(e)
			return
		}

		if instance == nil {
			// TODO: implement a proper way to show errors for the final user
			addNewStartupError(errors.New(i18n.Sprintf("tor can't be used")))
			return
		}

		u.tor = instance
	}()
}
