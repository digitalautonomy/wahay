package gui

import "sync"

func (u *gtkUI) ensureDependencies(onFinish func()) {
	var wg sync.WaitGroup

	u.ensureTor(&wg)
	u.ensureMumble(&wg)

	wg.Wait()

	onFinish()
}
