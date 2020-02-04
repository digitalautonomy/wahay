package gui

import (
	"fmt"
	"strings"

	"autonomia.digital/tonio/app/hosting"
	"github.com/coyim/gotk3adapter/gtki"
)

func (u *gtkUI) joinMeeting() {
	if u.mainWindow != nil {
		u.mainWindow.Hide()
	}
	u.openJoinWindow()
}

func (u *gtkUI) openMainWindow() {
	u.switchToMainWindow()
}

func (u *gtkUI) getInviteCodeEntities() (gtki.ApplicationWindow, *uiBuilder) {
	builder := u.g.uiBuilderFor("InviteCodeWindow")
	win := builder.get("inviteWindow").(gtki.ApplicationWindow)
	win.SetApplication(u.app)

	return win, builder
}

func (u *gtkUI) openCurrentMeetingWindow(m mumbleService) {
	if m.IsClosed() {
		u.reportError("The Mumble process is down")
	}

	u.hideCurrentWindow()

	builder := u.g.uiBuilderFor("CurrentMeetingWindow")
	win := builder.get("currentMeetingWindow").(gtki.ApplicationWindow)

	builder.ConnectSignals(map[string]interface{}{
		"on_close_window_signal": func() {
			u.leaveMeeting(m)
			u.quit()
		},
		"on_leave_meeting": func() {
			u.leaveMeeting(m)
		},
	})

	u.switchToWindow(win)
}

func (u *gtkUI) joinMeetingHandler(data hosting.MeetingData) {
	if len(data.MeetingID) == 0 {
		u.openErrorDialog("The Meeting ID cannot be blank")
		return
	}

	mumble, err := u.openMumble(data, u.switchContextWhenMumbleFinish)
	if err != nil {
		u.openErrorDialog(fmt.Sprintf("An error occurred\n\n%s", err.Error()))
		return
	}

	u.openCurrentMeetingWindow(mumble)
}

// Test Onion that can be used:
// qvdjpoqcg572ibylv673qr76iwashlazh6spm47ly37w65iwwmkbmtid.onion
func (u *gtkUI) openJoinWindow() {
	win, builder := u.getInviteCodeEntities()

	entMeetingID, _ := builder.get("entMeetingID").(gtki.Entry)
	entScreenName, _ := builder.get("entScreenName").(gtki.Entry)
	entMeetingPassword, _ := builder.get("entMeetingPassword").(gtki.Entry)

	cleanup := func() {
		win.Destroy()
		u.openMainWindow()
	}

	builder.ConnectSignals(map[string]interface{}{
		"on_join": func() {
			meetingID, _ := entMeetingID.GetText()
			username, _ := entScreenName.GetText()
			password, _ := entMeetingPassword.GetText()

			data := hosting.MeetingData{
				MeetingID: meetingID,
				Username:  username,
				Password:  password,
			}

			u.joinMeetingHandler(data)
		},
		"on_cancel": func() {
			cleanup()
		},
		"on_close": func() {
			cleanup()
		},
	})

	win.Show()
	u.setCurrentWindow(win)
}

func (u *gtkUI) openMumble(data hosting.MeetingData, onFinish func()) (mumbleService, error) {
	if !isMeetingIDValid(data.MeetingID) {
		return nil, fmt.Errorf("the provided meeting ID is invalid: \n\n%s", data.MeetingID)
	}
	return u.launchMumbleClient(data, onFinish)
}

const onionServiceLength = 60

// This function needs to be improved in order to make a real validation of the Meeting ID or Onion Address.
// At the moment, this function helps to test the error code window render.
func isMeetingIDValid(meetingID string) bool {
	return len(meetingID) > onionServiceLength && strings.HasSuffix(meetingID, ".onion")
}

func (u *gtkUI) leaveMeeting(m mumbleService) {
	u.wouldYouConfirmLeaveMeeting(func(res bool) {
		if res {
			m.Close()
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
