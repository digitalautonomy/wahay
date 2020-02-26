package gui

import (
	"strings"

	"github.com/coyim/gotk3adapter/gtki"
	"github.com/digitalautonomy/wahay/hosting"
	"github.com/digitalautonomy/wahay/tor"
)

func (u *gtkUI) joinMeeting() {
	u.hideMainWindow()
	u.openJoinWindow()
}

func (u *gtkUI) getInviteCodeEntities() (gtki.ApplicationWindow, *uiBuilder) {
	builder := u.g.uiBuilderFor("InviteCodeWindow")

	builder.i18nProperties(
		"title", "inviteWindow",
		"label", "lblMeetingID",
		"label", "lblUsername",
		"label", "lblMeetingPassword",
		"placeholder", "entScreenName",
		"placeholder", "entMeetingID",
		"placeholder", "entMeetingPassword",
		"button", "btnCancel",
		"button", "btnJoin")

	win := builder.get("inviteWindow").(gtki.ApplicationWindow)
	win.SetApplication(u.app)

	return win, builder
}

func (u *gtkUI) getCurrentMeetingWindow() *uiBuilder {
	builder := u.g.uiBuilderFor("CurrentMeetingWindow")

	builder.i18nProperties(
		"text", "leaveMeeting",
		"secondary_text", "leaveMeeting",
		"button", "btnLeaveMeeting",
		"tooltip", "btnLeaveMeeting",
	)

	return builder
}

func (u *gtkUI) openCurrentMeetingWindow(m tor.Service) {
	if m.IsClosed() {
		u.reportError(i18n.Sprintf("The Mumble process is down"))
	}

	u.hideCurrentWindow()

	builder := u.getCurrentMeetingWindow()
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
		u.openErrorDialog(i18n.Sprintf("The Meeting ID cannot be blank"))
		return
	}

	if !isAValidMeetingID(data.MeetingID) {
		u.reportError(i18n.Sprintf("The provided meeting ID is invalid: \n\n%s", data.MeetingID))
		return
	}

	u.hideCurrentWindow()
	u.displayLoadingWindow()

	var mumble tor.Service
	var err error

	finish := make(chan bool)

	go func() {
		mumble, err = u.launchMumbleClient(
			data,
			u.switchContextWhenMumbleFinish,
		)

		finish <- true
	}()

	<-finish // wait until the Mumble client has started

	u.hideLoadingWindow()

	if err != nil {
		u.openErrorDialog(i18n.Sprintf("An error occurred\n\n%s", err.Error()))
		u.showMainWindow()
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
		u.switchToMainWindow()
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

			go u.joinMeetingHandler(data)
		},
		"on_cancel": cleanup,
		"on_close":  cleanup,
	})

	win.Show()
	u.setCurrentWindow(win)
}

const onionServiceLength = 62

// TODO: This function needs to be improved in order to make a real validation of
// the Meeting ID or Onion Address.
// At the moment, this function helps to test the error code window render.
func isAValidMeetingID(meetingID string) bool {
	if len(meetingID) < onionServiceLength {
		return false
	}

	if len(meetingID) == onionServiceLength &&
		!strings.HasSuffix(meetingID, ".onion") {
		return false
	}

	if len(meetingID) > onionServiceLength &&
		len(meetingID[:onionServiceLength]) < onionServiceLength {
		return false
	}

	if len(meetingID) > onionServiceLength &&
		strings.Index(meetingID, ":") != onionServiceLength &&
		strings.Index(meetingID, ".onion") != onionServiceLength-4 {
		return false
	}

	return true
}

func (u *gtkUI) leaveMeeting(m tor.Service) {
	u.wouldYouConfirmLeaveMeeting(func(res bool) {
		if res {
			m.Close()
		}
	})
}

func (u *gtkUI) wouldYouConfirmLeaveMeeting(k func(bool)) {
	builder := u.getCurrentMeetingWindow()
	dialog := builder.get("leaveMeeting").(gtki.MessageDialog)
	dialog.SetDefaultResponse(gtki.RESPONSE_NO)
	responseType := gtki.ResponseType(dialog.Run())
	result := responseType == gtki.RESPONSE_YES
	dialog.Destroy()
	k(result)
}
