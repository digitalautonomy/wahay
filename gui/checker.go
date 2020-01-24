package gui

import (
	"errors"
	"log"
	"sync"
)

func (u *gtkUI) ensureDependencies(cb func(bool)) {
	log.Println("Checking application dependencies")

	success := true
	var wg sync.WaitGroup

	// TODO: better error handling for startup checking

	// Ensure the Tor binary and network
	wg.Add(1)
	go u.ensureTor(&wg)

	// Ensure the Mumble binary
	wg.Add(1)
	go u.ensureMumble(&wg)

	wg.Wait()

	log.Println("Finishing dependencies checking")

	u.hideLoadingWindow()

	if len(startupErrors) != 0 {
		// TODO: show startup errors and give feedback to the user
		success = false
		u.displayStartupError(errors.New("something happened"))
	}

	cb(success)
}
