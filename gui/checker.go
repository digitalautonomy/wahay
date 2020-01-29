package gui

import "sync"

func (u *gtkUI) ensureDependencies(cb func(bool)) {
	var wg sync.WaitGroup

	// TODO: better error handling for startup checking

	u.ensureTor(&wg)
	u.ensureMumble(&wg)

	wg.Wait()

	cb(len(startupErrors) == 0)
}
