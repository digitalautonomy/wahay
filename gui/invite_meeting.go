package gui

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/coyim/gotk3adapter/gtki"
)

func (u *gtkUI) getInviteCodeEntities() (gtki.Entry, gtki.ApplicationWindow, *uiBuilder) {
	builder := u.g.uiBuilderFor("InviteCodeWindow")
	url := builder.get("entMeetingID").(gtki.Entry)
	win := builder.get("inviteWindow").(gtki.ApplicationWindow)
	win.SetApplication(u.app)

	return url, win, builder
}

func (u *gtkUI) openCurrentMeetingWindow(state *runningMumble, meetingID string) {
	u.currentWindow.Hide()
	builder := u.g.uiBuilderFor("CurrentMeetingWindow")
	win := builder.get("currentMeetingWindow").(gtki.ApplicationWindow)
	labelMeetingID := builder.get("currentMeetingID").(gtki.Label)
	win.SetApplication(u.app)
	u.currentWindow = win
	builder.ConnectSignals(map[string]interface{}{
		"on_close_window_signal": func() {
			u.leaveMeeting(state)
			u.quit()
		},
		"on_leave_meeting": func() { u.leaveMeeting(state) },
	})

	labelMeetingID.SetText(meetingID)
	u.switchContextWhenMumbleFinished(state)
	win.ShowAll()
}

// Test Onion that can be used: qvdjpoqcg572ibylv673qr76iwashlazh6spm47ly37w65iwwmkbmtid.onion

func (u *gtkUI) openJoinWindow() {
	url, win, builder := u.getInviteCodeEntities()
	u.currentWindow = win

	builder.ConnectSignals(map[string]interface{}{
		"on_join": func() {
			idEntered, err := url.GetText()
			if err != nil {
				u.openErrorDialog()
				return
			}
			state, err := openMumble(idEntered)
			if err != nil {
				u.openErrorDialog()
			}
			u.openCurrentMeetingWindow(state, idEntered)
		},
		"on_cancel": func() {
			win.Hide()
		},
	})
	win.ShowAll()
}

func openMumble(inviteID string) (*runningMumble, error) {
	if !isMeetingIDValid(inviteID) {
		return nil, fmt.Errorf("invalid Onion Address %s", inviteID)
	}

	return launchMumbleClient(inviteID)
}

const onionServiceLength = 60

//This function needs to be improved in order to make a real validation of the Meeting ID or Onion Address.
//At the moment, this function helps to test the error code window render.
func isMeetingIDValid(meetingID string) bool {
	return len(meetingID) > onionServiceLength && strings.HasSuffix(meetingID, ".onion")
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

func launchMumbleClient(inviteID string) (*runningMumble, error) {
	mumbleURL := fmt.Sprintf("mumble://%s", inviteID)

	ctx, cancelFunc := context.WithCancel(context.Background())

	cmd := exec.CommandContext(ctx, "torify", "mumble", mumbleURL)
	if err := cmd.Start(); err != nil {
		cancelFunc()
		return nil, err
	}

	state := &runningMumble{
		cmd:               cmd,
		ctx:               ctx,
		cancelFunc:        cancelFunc,
		finished:          false,
		finishedWithError: nil,
		finishChannel:     make(chan bool, 100),
	}

	go state.waitForFinish()

	return state, nil
}

func (u *gtkUI) switchContextWhenMumbleFinished(state *runningMumble) {
	go func() {
		<-state.finishChannel
		// here, we  could check if the Mumble instance failed with an error
		// and report this
		u.doInUIThread(func() {
			u.currentWindow.Hide()
			u.currentWindow = u.mainWindow
			u.currentWindow.ShowAll()
		})
	}()
}

func (u *gtkUI) leaveMeeting(state *runningMumble) {
	u.wouldYouConfirmLeaveMeeting(func(res bool) {
		if res {
			state.cancelFunc()
		}
	})
}

func (u *gtkUI) wouldYouConfirmLeaveMeeting(k func(bool)) {
	builder := u.g.uiBuilderFor("LeaveMeetingConfirmation")
	dialog := builder.get("leaveMeeting").(gtki.MessageDialog)
	dialog.SetDefaultResponse(gtki.RESPONSE_NO)
	dialog.SetTransientFor(u.mainWindow)
	responseType := gtki.ResponseType(dialog.Run())
	result := responseType == gtki.RESPONSE_YES
	dialog.Destroy()
	k(result)
}
