package gui

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"text/template"

	"autonomia.digital/tonio/app/config"
	"autonomia.digital/tonio/app/hosting"
	"github.com/coyim/gotk3adapter/gtki"
)

type hostData struct {
	u               *gtkUI
	runningState    *runningMumble
	serverPort      int
	serverControl   hosting.Server
	serviceID       string
	autoJoin        bool
	meetingUsername string
	meetingPassword string
	next            func()
}

func (u *gtkUI) hostMeetingHandler() {
	go u.realHostMeetingHandler()
}

func (u *gtkUI) realHostMeetingHandler() {
	u.doInUIThread(func() {
		u.currentWindow.Hide()
		u.displayLoadingWindow()
	})

	h := &hostData{
		u:               u,
		autoJoin:        u.config.GetAutoJoin(),
		meetingPassword: "",
		next:            func() {},
	}

	ch := make(chan bool)

	go h.createOnionService(ch)

	<-ch

	u.doInUIThread(func() {
		u.hideLoadingWindow()
		h.showMeetingConfiguration()
	})
}

func (h *hostData) showMeetingControls() {
	builder := h.u.g.uiBuilderFor("StartHostingWindow")
	win := builder.get("startHostingWindow").(gtki.ApplicationWindow)
	builder.ConnectSignals(map[string]interface{}{
		"on_close_window_signal": h.u.quit,
		"on_finish_meeting": func() {
			if h.serverControl != nil {
				h.finishMeeting()
			} else {
				log.Print("server is nil")
			}
		},
		"on_join_meeting": func() {
			if h.serverControl != nil {
				h.joinMeetingHost()
			} else {
				log.Print("server is nil")
			}
		},
		"on_invite_others": func() {
			h.showInvitePeopleWindow(builder)
		},
		"on_copy_meeting_id": func() {
			h.copyMeetingIDToClipboard(builder, "")
		},
		"on_send_by_email": func() {
			h.sendInvitationByEmail(builder)
		},
		"on_copy_meeting_url": func() {
			h.copyMeetingIDToClipboard(builder, "lblCopyUrlMessage")
		},
	})

	meetingID, err := builder.GetObject("lblMeetingID")
	if err != nil {
		log.Printf("meeting id error: %s", err)
	}
	_ = meetingID.SetProperty("label", h.serviceID)

	h.u.switchToWindow(win)
}

func (h *hostData) joinMeetingHost() {
	loaded := make(chan bool)

	go func() {
		data := hosting.MeetingData{
			MeetingID: h.serviceID,
			Password:  h.meetingPassword,
			Username:  h.meetingUsername,
		}
		state, err := h.u.launchMumbleClient(data)
		if err != nil {
			h.u.reportError(fmt.Sprintf("Programmer error #1: %s", err.Error()))
			return
		}
		h.runningState = state
		loaded <- true
	}()

	go func() {
		<-loaded
		h.openHostJoinMeetingWindow()
	}()
}

func (h *hostData) openHostJoinMeetingWindow() {
	h.u.doInUIThread(func() {
		h.u.currentWindow.Hide()
	})

	builder := h.u.g.uiBuilderFor("CurrentHostMeetingWindow")
	win := builder.get("hostMeetingWindow").(gtki.ApplicationWindow)

	builder.ConnectSignals(map[string]interface{}{
		"on_close_window_signal": func() {
			h.leaveHostMeeting()
			h.u.quit()
		},
		"on_leave_meeting":  h.leaveHostMeeting,
		"on_finish_meeting": h.finishMeetingMumble,
	})

	h.switchToHostOnFinishMeeting()
	h.u.switchToWindow(win)
}

func (h *hostData) uiActionLeaveMeeting() {
	h.u.currentWindow.Hide()
	h.showMeetingControls()
}

func (h *hostData) uiActionFinishMeeting() {
	h.finishMeetingReal()
}

func (h *hostData) switchToHostOnFinishMeeting() {
	go func() {
		<-h.runningState.finishChannel

		// TODO: here, we  could check if the Mumble instance
		// failed with an error and report this
		h.u.doInUIThread(func() {
			h.next()
			h.next = func() {}
		})
	}()
}

func (u *gtkUI) ensureServerCollection() {
	if u.serverCollection == nil {
		var e error
		u.serverCollection, e = hosting.Create()
		if e != nil {
			u.reportError(fmt.Sprintf("Something went wrong: %s", e.Error()))
		}
	}
}

func (h *hostData) createOnionService(ch chan bool) {
	if h.u.tor != nil {
		h.u.ensureServerCollection()

		port := config.GetRandomPort()

		controller := h.u.tor.GetController()
		serviceID, e := controller.CreateNewOnionService("127.0.0.1", port, 64738)
		if e != nil {
			h.u.reportError(fmt.Sprintf("Something went wrong: %s", e.Error()))
			return
		}

		h.serverPort = port
		h.serviceID = serviceID
	}

	ch <- true
}

func (h *hostData) createNewConferenceRoom(complete chan bool) {
	server, e := h.u.serverCollection.CreateServer(fmt.Sprintf("%d", h.serverPort), h.meetingPassword)
	if e != nil {
		h.u.reportError(fmt.Sprintf("Something went wrong: %s", e.Error()))
		complete <- true
		return
	}
	e = server.Start()
	if e != nil {
		h.u.reportError(fmt.Sprintf("Something went wrong: %s", e.Error()))
		complete <- true
		return
	}

	h.serverControl = server

	complete <- true
}

func (h *hostData) finishMeetingReal() {
	// Hide the current window
	h.u.doInUIThread(h.u.currentWindow.Hide)

	// TODO: What happen if two errors occurrs?
	// We need to do a better controlling for each error
	// and if multiple errors occurrs, show all the errors in the
	// same window using the `u.reportError` function
	err := h.serverControl.Stop()
	if err != nil {
		h.u.reportError(fmt.Sprintf("The meeting can't be closed: %s", err))
	}

	h.deleteOnionService()

	h.u.doInUIThread(func() {
		h.u.currentWindow.Hide()
		h.u.currentWindow = h.u.mainWindow
		h.u.mainWindow.Show()
	})
}

func (h *hostData) deleteOnionService() {
	var err error
	if h.u.tor != nil && len(h.serviceID) != 0 {
		controller := h.u.tor.GetController()
		err = controller.DeleteOnionService(h.serviceID)
		if err == nil {
			h.serviceID = ""
		}
	} else {
		err = errors.New("tor is not defined")
	}

	if err != nil {
		h.u.reportError(fmt.Sprintf("The onion service can't be deleted: %s", err))
	}
}

func (h *hostData) finishMeetingMumble() {
	h.u.wouldYouConfirmFinishMeeting(func(res bool) {
		if res {
			h.next = h.uiActionFinishMeeting
			go h.runningState.close()
		}
	})
}

func (h *hostData) finishMeeting() {
	h.u.wouldYouConfirmFinishMeeting(func(res bool) {
		if res {
			h.finishMeetingReal()
		}
	})
}

func (h *hostData) leaveHostMeeting() {
	h.next = h.uiActionLeaveMeeting
	go h.runningState.close()
}

func (h *hostData) copyMeetingIDToClipboard(builder *uiBuilder, label string) {
	err := h.u.copyToClipboard(h.serviceID)
	if err != nil {
		fatal("clipboard copying error")
	}

	var lblMessage gtki.Label
	if len(label) == 0 {
		lblMessage = builder.get("lblMessage").(gtki.Label)
	} else {
		lblMessage = builder.get(label).(gtki.Label)
	}
	_ = lblMessage.SetProperty("visible", false)

	go func() {
		h.u.messageToLabel(lblMessage, "The meeting ID has been copied to Clipboard", 5)
	}()
}

func (h *hostData) copyInvitationToClipboard(builder *uiBuilder) {
	err := h.u.copyToClipboard(h.getInvitationText())
	if err != nil {
		fatal("clipboard copying error")
	}
	lblMessage := builder.get("lblMessage").(gtki.Label)
	_ = lblMessage.SetProperty("visible", false)

	go func() {
		h.u.messageToLabel(lblMessage, "The invitation email has been copied to Clipboard", 5)
	}()
}

func (h *hostData) sendInvitationByEmail(builder *uiBuilder) {
	lnkEmail := builder.get("lnkEmail").(gtki.LinkButton)
	_ = lnkEmail.SetProperty("uri", h.getInvitationEmailURI())
	_, _ = lnkEmail.Emit("clicked")
}

func (h *hostData) getInvitationEmailURI() string {
	subject := h.getInvitationSubject()
	body := h.getInvitationText()
	uri := fmt.Sprintf("mailto:?subject=%s&body=%s", subject, body)
	return uri
}

func (h *hostData) getInvitationGmailURI() string {
	subject := h.getInvitationSubject()
	body := h.getInvitationText()
	uri := fmt.Sprintf("%s?view=cm&fs=1&tf=1&to=&su=%s&body=%s", gmailURL, subject, body)
	return uri
}

func (h *hostData) getInvitationYahooURI() string {
	subject := h.getInvitationSubject()
	body := h.getInvitationText()
	uri := fmt.Sprintf("%s?To=&Subj=%s&Body=%s", yahooURL, subject, body)
	return uri
}

func (h *hostData) getInvitationMicrosoftURI() string {
	subject := h.getInvitationSubject()
	body := h.getInvitationText()
	uri := fmt.Sprintf("%s?rru=compose&subject=%s&body=%s&to=#page=Compose", outlookURL, subject, body)
	return uri
}

const invitationTextTemplate = `
Please join Tonio meeting with the following details:%0D%0A%0D%0A
{{ if .MeetingID }}
Meeting ID: {{ .MeetingID }}%0D%0A
{{ end }}
`

// Invitation is the information of the meeting
type Invitation struct {
	MeetingID string
}

func (h *hostData) getInvitationSubject() string {
	return "Join Tonio Meeting"
}

func (h *hostData) getInvitationText() string {
	data := Invitation{h.serviceID}
	tmpl := template.Must(template.New("invitation").Parse(invitationTextTemplate))

	var b bytes.Buffer
	err := tmpl.Execute(&b, &data)
	if err != nil {
		fatal("An error occurred while parsing the invitation template")
	}

	return b.String()
}

func (u *gtkUI) wouldYouConfirmFinishMeeting(k func(bool)) {
	builder := u.g.uiBuilderFor("StartHostingWindow")
	dialog := builder.get("finishMeeting").(gtki.MessageDialog)
	dialog.SetDefaultResponse(gtki.RESPONSE_NO)
	dialog.SetTransientFor(u.mainWindow)
	responseType := gtki.ResponseType(dialog.Run())
	result := responseType == gtki.RESPONSE_YES
	dialog.Destroy()
	k(result)
}

func (h *hostData) showMeetingConfiguration() {
	builder := h.u.g.uiBuilderFor("ConfigureMeetingWindow")
	win := builder.get("configureMeetingWindow").(gtki.ApplicationWindow)
	chk := builder.get("chkAutoJoin").(gtki.CheckButton)
	btnStart := builder.get("btnStartMeeting").(gtki.Button)

	chk.SetActive(h.autoJoin)
	h.changeStartButtonText(btnStart)

	builder.ConnectSignals(map[string]interface{}{
		"on_copy_meeting_id": func() {
			h.copyMeetingIDToClipboard(builder, "")
		},
		"on_send_by_email": func() {
			h.sendInvitationByEmail(builder)
		},
		"on_cancel": func() {
			h.deleteOnionService()
			h.u.currentWindow.Hide()
			h.u.switchToMainWindow()
		},
		"on_start_meeting": func() {
			username := builder.get("inpMeetingUsername").(gtki.Entry)
			password := builder.get("inpMeetingPassword").(gtki.Entry)
			h.meetingUsername, _ = username.GetText()
			h.meetingPassword, _ = password.GetText()
			go h.startMeetingHandler()
		},
		"on_invite_others": h.onInviteParticipants,
		"on_chkAutoJoin_toggled": func() {
			h.autoJoin = chk.GetActive()
			h.u.config.SetAutoJoin(h.autoJoin)
			h.u.saveConfigOnly()
			h.changeStartButtonText(btnStart)
		},
	})

	meetingID, err := builder.GetObject("inpMeetingID")
	if err != nil {
		log.Printf("meeting id error: %s", err)
	}
	_ = meetingID.SetProperty("text", h.serviceID)

	h.u.switchToWindow(win)
}

func (h *hostData) showInvitePeopleWindow(builder *uiBuilder) {
	h.u.currentWindow.Hide()
	win := builder.get("invitePeopleWindow").(gtki.ApplicationWindow)
	h.u.switchToWindow(win)
}

func (h *hostData) changeStartButtonText(btn gtki.Button) {
	if h.autoJoin {
		_ = btn.SetProperty("label", "Start Meeting & Join")
	} else {
		_ = btn.SetProperty("label", "Start Meeting")
	}
}

func (h *hostData) startMeetingHandler() {
	h.u.currentWindow.Hide()

	complete := make(chan bool)

	go func() {
		h.createNewConferenceRoom(complete)
	}()

	<-complete

	h.u.doInUIThread(func() {
		if h.autoJoin {
			h.joinMeetingHost()
		} else {
			h.showMeetingControls()
		}
	})
}

func (h *hostData) onInviteParticipants() {
	builder := h.u.g.uiBuilderFor("InvitePeopleWindow")
	win := builder.get("invitePeopleWindow").(gtki.ApplicationWindow)

	btnEmail := builder.get("btnEmail").(gtki.LinkButton)
	btnGmail := builder.get("btnGmail").(gtki.LinkButton)
	btnYahoo := builder.get("btnYahoo").(gtki.LinkButton)
	btnOutlook := builder.get("btnMicrosoft").(gtki.LinkButton)

	_ = btnEmail.SetProperty("uri", h.getInvitationEmailURI())
	_ = btnGmail.SetProperty("uri", h.getInvitationGmailURI())
	_ = btnYahoo.SetProperty("uri", h.getInvitationYahooURI())
	_ = btnOutlook.SetProperty("uri", h.getInvitationMicrosoftURI())

	builder.ConnectSignals(map[string]interface{}{
		"on_close_window_signal": win.Hide,
		"on_link_clicked":        win.Hide,
		"on_copy_meeting_id": func() {
			h.copyMeetingIDToClipboard(builder, "")
		},
		"on_copy_invitation": func() {
			h.copyInvitationToClipboard(builder)
		},
	})

	h.u.doInUIThread(win.Show)
}
