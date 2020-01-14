package gui

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"autonomia.digital/tonio/app/hosting"
	"github.com/coyim/gotk3adapter/gtki"
)

func (u *gtkUI) openMainWindow() {
	u.currentWindow.Hide()
	u.currentWindow = u.mainWindow
	u.currentWindow.ShowAll()
}

func (u *gtkUI) getInviteCodeEntities() (gtki.Entry, gtki.ApplicationWindow, *uiBuilder) {
	builder := u.g.uiBuilderFor("InviteCodeWindow")
	url := builder.get("entMeetingID").(gtki.Entry)
	win := builder.get("inviteWindow").(gtki.ApplicationWindow)
	win.SetApplication(u.app)

	return url, win, builder
}

func (u *gtkUI) openCurrentMeetingWindow(state *runningMumble) {
	u.doInUIThread(func() {
		u.currentWindow.Hide()
	})

	builder := u.g.uiBuilderFor("CurrentMeetingWindow")
	win := builder.get("currentMeetingWindow").(gtki.ApplicationWindow)

	builder.ConnectSignals(map[string]interface{}{
		"on_close_window_signal": func() {
			u.leaveMeeting(state)
			u.quit()
		},
		"on_leave_meeting": func() {
			u.leaveMeeting(state)
		},
	})

	u.switchToWindow(win)
}

func (u *gtkUI) joinMeetingHandler(meetingID string) {
	if len(meetingID) == 0 {
		u.openErrorDialog("The Meeting ID cannot be blank")
		return
	}

	//loaded := make(chan bool)

	state, err := openMumble(meetingID)
	if err != nil {
		u.openErrorDialog(fmt.Sprintf("An error occurred\n\n%s", err.Error()))
		return
	}
	//loaded <- true

	u.switchContextWhenMumbleFinished(state)

	u.currentWindow.Hide()
	u.openCurrentMeetingWindow(state)

	// go func() {
	// 	//<-loaded
	// 	//u.openCurrentMeetingWindow(state)
	// }()
}

// Test Onion that can be used:
// qvdjpoqcg572ibylv673qr76iwashlazh6spm47ly37w65iwwmkbmtid.onion

func (u *gtkUI) openJoinWindow() {
	url, win, builder := u.getInviteCodeEntities()
	u.currentWindow = win

	builder.ConnectSignals(map[string]interface{}{
		"on_join": func() {
			meetingID, _ := url.GetText()
			u.joinMeetingHandler(meetingID)
		},
		"on_cancel": func() {
			win.Hide()
			u.openMainWindow()
		},
		"on_close": func() {
			u.openMainWindow()
		},
	})

	win.ShowAll()
}

func openMumble(meetingID string) (*runningMumble, error) {
	if !isMeetingIDValid(meetingID) {
		return nil, fmt.Errorf("the provided meeting ID is invalid: \n\n%s", meetingID)
	}
	data := hosting.MeetingData{
		MeetingID: meetingID,
		Password:  "",
		Username:  "",
	}
	return launchMumbleClient(data)
}

const onionServiceLength = 60

// This function needs to be improved in order to make a real validation of the Meeting ID or Onion Address.
// At the moment, this function helps to test the error code window render.
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

func launchMumbleClient(data hosting.MeetingData) (*runningMumble, error) {
	ctx, cancelFunc := context.WithCancel(context.Background())

	cmd := exec.CommandContext(ctx, "torify", "mumble", hosting.GenerateURL(data))
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

		// TODO: here, we  could check if the Mumble instance
		// failed with an error and report this
		u.doInUIThread(func() {
			u.openMainWindow()
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
	builder := u.g.uiBuilderFor("CurrentMeetingWindow")
	dialog := builder.get("leaveMeeting").(gtki.MessageDialog)
	dialog.SetDefaultResponse(gtki.RESPONSE_NO)
	responseType := gtki.ResponseType(dialog.Run())
	result := responseType == gtki.RESPONSE_YES
	dialog.Destroy()
	k(result)
}
