package gui

import (
	"context"
	"fmt"
	"os/exec"
	"sync"

	"autonomia.digital/tonio/app/client"
	"autonomia.digital/tonio/app/hosting"
)

func (u *gtkUI) ensureMumble(wg *sync.WaitGroup) {
	wg.Add(1)
	go func() {
		defer wg.Done()

		c, e := client.InitSystem()
		if e != nil {
			addNewStartupError(e)
			return
		}

		if !c.CanBeUsed() {
			addNewStartupError(fmt.Errorf("the client can not be used because: %s", c.GetLastError()))
			return
		}

		u.client = c
	}()
}

func (u *gtkUI) launchMumbleClient(data hosting.MeetingData) (*runningMumble, error) {
	rc, err := u.throughTor(u.client.GetBinary(), []string{
		hosting.GenerateURL(data),
	})
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	state := &runningMumble{
		cmd:               rc.Cmd,
		ctx:               rc.Ctx,
		cancelFunc:        rc.CancelFunc,
		finished:          false,
		finishedWithError: nil,
		finishChannel:     make(chan bool, 100),
	}

	go state.waitForFinish()

	return state, nil
}

func (u *gtkUI) switchContextWhenMumbleFinish(state *runningMumble) {
	go func() {
		<-state.finishChannel

		// TODO: here, we  could check if the Mumble instance
		// failed with an error and report this
		u.doInUIThread(func() {
			u.openMainWindow()
		})
	}()
}

type runningMumble struct {
	cmd               *exec.Cmd
	ctx               context.Context
	cancelFunc        context.CancelFunc
	finished          bool
	finishedWithError error
	finishChannel     chan bool
}

func (r *runningMumble) close() {
	r.cancelFunc()
}

func (r *runningMumble) waitForFinish() {
	e := r.cmd.Wait()
	r.finished = true
	r.finishedWithError = e
	r.finishChannel <- true
}
