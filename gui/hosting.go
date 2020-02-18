package gui

import (
	"errors"
	"fmt"
	"strconv"

	log "github.com/sirupsen/logrus"

	"github.com/coyim/gotk3adapter/gtki"
	"github.com/digitalautonomy/wahay/config"
	"github.com/digitalautonomy/wahay/hosting"
	"github.com/digitalautonomy/wahay/tor"
)

// defaultPortMumble contains the default port used in mumble
const defaultPortMumble = 64738

type hostData struct {
	u               *gtkUI
	mumble          tor.Service
	serverPort      int
	serverControl   hosting.Server
	serviceID       string
	autoJoin        bool
	meetingUsername string
	meetingPassword string
	currentWindow   gtki.Window
	next            func()
}

func (u *gtkUI) hostMeetingHandler() {
	go u.realHostMeetingHandler()
}

func (u *gtkUI) realHostMeetingHandler() {
	u.hideMainWindow()
	u.displayLoadingWindow()

	h := &hostData{
		u:        u,
		autoJoin: u.config.GetAutoJoin(),
		next:     nil,
	}

	finish := make(chan string)
	go h.createOnionService(finish)

	err := <-finish

	u.doInUIThread(func() {
		u.hideLoadingWindow()
		if len(err) > 0 {
			u.switchToMainWindow()
			h.u.reportError(i18n.Sprintf("Something went wrong: %s", err))
		} else {
			h.showMeetingConfiguration()
		}
	})
}

func (h *hostData) showMeetingControls() {
	builder := h.u.g.uiBuilderFor("StartHostingWindow")
	win := builder.get("startHostingWindow").(gtki.ApplicationWindow)

	builder.i18nProperties(
		"label", "lblHostMeeting",
		"label", "lblMeetingID",
		"label", "lblCopyUrlMessage",
		"button", "btnFinishMeeting",
		"button", "btnJoinMeeting",
		"button", "btnJoinMeeting",
		"button", "btnCopyUrl",
		"tooltip", "btnJoinMeeting")

	builder.ConnectSignals(map[string]interface{}{
		"on_close_window_signal": h.finishMeetingReal,
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
		"on_invite_others": h.onInviteParticipants,
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

	meetingID, err := builder.GetObject("lblMeetingIDValue")
	if err != nil {
		log.Printf("meeting id error: %s", err)
	}
	_ = meetingID.SetProperty("label", h.serviceID)

	h.u.switchToWindow(win)
}

func (h *hostData) joinMeetingHost() {
	var err error
	validOpChannel := make(chan bool)

	go func() {
		data := hosting.MeetingData{
			MeetingID: h.serviceID,
			Password:  h.meetingPassword,
			Username:  h.meetingUsername,
		}

		mumble, err := h.u.launchMumbleClient(data, func() {
			if h.next == nil {
				h.next = h.uiActionFinishMeeting
			}

			h.switchToHostOnFinishMeeting()
		})

		if err != nil {
			log.Errorf("joinMeetingHost() error: %s", err)
			validOpChannel <- false
		} else {
			h.mumble = mumble
			validOpChannel <- true
		}
	}()

	if <-validOpChannel {
		h.openHostJoinMeetingWindow()
		return
	}

	if err == nil {
		err = errors.New(i18n.Sprintf("we couldn't start the meeting"))
	}

	h.u.reportError(err.Error())

	h.u.switchToMainWindow()
}

func (h *hostData) switchToHostOnFinishMeeting() {
	h.u.doInUIThread(func() {
		if h.next != nil {
			h.next()
			h.next = nil
		}
	})
}

func (u *gtkUI) getCurrentHostMeetingWindow() *uiBuilder {
	builder := u.g.uiBuilderFor("CurrentHostMeetingWindow")

	builder.i18nProperties(
		"button", "btnFinishMeeting",
		"button", "btnLeaveMeeting",
		"tooltip", "btnFinishMeeting",
		"tooltip", "btnLeaveMeeting")

	return builder
}

func (h *hostData) openHostJoinMeetingWindow() {
	h.u.hideCurrentWindow()

	builder := h.u.getCurrentHostMeetingWindow()
	win := builder.get("hostMeetingWindow").(gtki.ApplicationWindow)

	builder.ConnectSignals(map[string]interface{}{
		"on_close_window_signal": func() {
			h.leaveHostMeeting()
			h.u.quit()
		},
		"on_leave_meeting":  h.leaveHostMeeting,
		"on_finish_meeting": h.finishMeetingMumble,
	})

	h.u.switchToWindow(win)
}

func (h *hostData) uiActionLeaveMeeting() {
	h.u.currentWindow.Hide()
	h.showMeetingControls()
}

func (h *hostData) uiActionFinishMeeting() {
	h.finishMeetingReal()
}

func (u *gtkUI) ensureServerCollection() {
	if u.serverCollection == nil {
		var e error
		u.serverCollection, e = hosting.Create()
		if e != nil {
			err := e.Error()
			u.reportError(i18n.Sprintf("Something went wrong: %s", err))
		}
	}
}

func (h *hostData) createOnionService(finish chan string) {
	if h.u.tor != nil {
		h.u.ensureServerCollection()

		rp := config.GetRandomPort()

		var err error
		pm := defaultPortMumble
		if h.u.config.GetPortMumble() != "" {
			pm, err = strconv.Atoi(h.u.config.GetPortMumble())
			if err != nil {
				h.u.reportError(i18n.Sprintf("Configured Mumble port is not valid: %s", h.u.config.GetPortMumble()))
				finish <- err.Error()
				return
			}
		}

		controller := h.u.tor.GetController()
		serviceID, e := controller.CreateNewOnionService("127.0.0.1", rp, pm)
		if e != nil {
			finish <- e.Error()
			return
		}

		h.serverPort = rp
		h.serviceID = serviceID
	}

	finish <- ""
}

func (h *hostData) createNewConferenceRoom(complete chan bool) {
	server, e := h.u.serverCollection.CreateServer(fmt.Sprintf("%d", h.serverPort), h.meetingPassword)
	if e != nil {
		err := e.Error()
		h.u.hideLoadingWindow()
		h.u.reportError(i18n.Sprintf("Something went wrong: %s", err))
		complete <- false
		return
	}

	e = server.Start()
	if e != nil {
		err := e.Error()
		h.u.hideLoadingWindow()
		h.u.reportError(i18n.Sprintf("Something went wrong: %s", err))
		complete <- false
		return
	}

	h.serverControl = server

	complete <- true
}

func (h *hostData) finishMeetingReal() {
	// TODO: What happen if two errors occurrs?
	// We need to do a better controlling for each error
	// and if multiple errors occurrs, show all the errors in the
	// same window using the `u.reportError` function
	err := h.serverControl.Stop()
	if err != nil {
		h.u.reportError(i18n.Sprintf("The meeting can't be closed: %s", err))
	}

	h.deleteOnionService()

	if h.currentWindow != nil {
		h.currentWindow.Destroy()
		h.currentWindow = nil
	}

	h.u.switchToMainWindow()
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
		err = errors.New(i18n.Sprintf("internal Tor instance has already been closed"))
	}

	if err != nil {
		h.u.reportError(i18n.Sprintf("The onion service can't be deleted: %s", err))
	}
}

func (h *hostData) finishMeetingMumble() {
	h.wouldYouConfirmFinishMeeting(func(res bool) {
		if res {
			h.next = h.uiActionFinishMeeting
			go h.mumble.Close()
		}
	})
}

func (h *hostData) finishMeeting() {
	h.wouldYouConfirmFinishMeeting(func(res bool) {
		if res {
			h.finishMeetingReal()
		}
	})
}

func (h *hostData) leaveHostMeeting() {
	h.next = h.uiActionLeaveMeeting
	go h.mumble.Close()
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
		h.u.messageToLabel(lblMessage, i18n.Sprintf("The meeting ID has been copied to the clipboard"), 5)
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
		h.u.messageToLabel(lblMessage, i18n.Sprintf("The invitation email has been copied to the clipboard"), 5)
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

func (h *hostData) getInvitationSubject() string {
	return i18n.Sprintf("Join Wahay Meeting")
}

func (h *hostData) getInvitationText() string {
	it := i18n.Sprintf("Please join the Wahay meeting with the following details:") + "%0D%0A%0D%0A"
	if h.serviceID != "" {
		it = i18n.Sprintf("%sMeeting ID: %s", it, h.serviceID)
	}
	return it
}

func (h *hostData) wouldYouConfirmFinishMeeting(k func(bool)) {
	builder := h.u.g.uiBuilderFor("StartHostingWindow")
	dialog := builder.get("finishMeeting").(gtki.MessageDialog)

	builder.i18nProperties(
		"text", "finishMeeting",
		"secondary_text", "finishMeeting")

	dialog.SetDefaultResponse(gtki.RESPONSE_NO)
	dialog.SetTransientFor(h.u.currentWindow)

	h.currentWindow = dialog

	responseType := gtki.ResponseType(dialog.Run())
	result := responseType == gtki.RESPONSE_YES
	dialog.Destroy()
	k(result)
}

func (u *gtkUI) getConfigureMeetingWindow() *uiBuilder {
	builder := u.g.uiBuilderFor("ConfigureMeetingWindow")

	builder.i18nProperties(
		"label", "labelMeetingID",
		"label", "labelUsername",
		"label", "lblMessage",
		"label", "labelMeetingPassword",
		"placeholder", "inpMeetingUsername",
		"placeholder", "inpMeetingPassword",
		"checkbox", "chkAutoJoin",
		"tooltip", "chkAutoJoin",
		"button", "btnCopyMeetingID",
		"button", "btnInviteOthers",
		"button", "btnCancel",
		"button", "btnStartMeeting")

	return builder
}

func (h *hostData) showMeetingConfiguration() {
	builder := h.u.getConfigureMeetingWindow()
	win := builder.get("configureMeetingWindow").(gtki.ApplicationWindow)
	chk := builder.get("chkAutoJoin").(gtki.CheckButton)
	btnStart := builder.get("btnStartMeeting").(gtki.Button)

	builder.i18nProperties(
		"label", "labelMeetingID",
		"label", "labelUsername",
		"label", "labelMeetingPassword",
		"label", "lblMessage")

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
			h.u.switchToMainWindow()
		},
		"on_start_meeting": func() {
			username := builder.get("inpMeetingUsername").(gtki.Entry)
			password := builder.get("inpMeetingPassword").(gtki.Entry)
			h.meetingUsername, _ = username.GetText()
			h.meetingPassword, _ = password.GetText()
			h.startMeetingHandler()
		},
		"on_invite_others": h.onInviteParticipants,
		"on_chkAutoJoin_toggled": func() {
			h.autoJoin = chk.GetActive()
			h.u.config.SetAutoJoin(h.autoJoin)
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

func (h *hostData) changeStartButtonText(btn gtki.Button) {
	if h.autoJoin {
		_ = btn.SetProperty("label", i18n.Sprintf("Start Meeting & Join"))
	} else {
		_ = btn.SetProperty("label", i18n.Sprintf("Start Meeting"))
	}
}

func (h *hostData) startMeetingHandler() {
	h.u.hideCurrentWindow()
	h.u.displayLoadingWindow()
	go h.startMeetingRoutine()
}

func (h *hostData) startMeetingRoutine() {
	complete := make(chan bool)

	go h.createNewConferenceRoom(complete)

	if !<-complete {
		// TODO: Close the meeting window and return to the main window
		return
	}

	h.u.hideLoadingWindow()

	if h.autoJoin {
		h.joinMeetingHost()
	} else {
		h.showMeetingControls()
	}
}

func (h *hostData) onInviteParticipants() {
	builder := h.u.g.uiBuilderFor("InvitePeopleWindow")

	builder.i18nProperties(
		"label", "lblDescription",
		"label", "lblDefaultEmail",
		"label", "lblGmail",
		"label", "lblYahoo",
		"label", "lblOutlook",
		"button", "btnCopyMeetingID",
		"button", "btnCopyInvitation")

	dialog := builder.get("invitePeopleWindow").(gtki.ApplicationWindow)

	btnEmail := builder.get("btnEmail").(gtki.LinkButton)
	btnGmail := builder.get("btnGmail").(gtki.LinkButton)
	btnYahoo := builder.get("btnYahoo").(gtki.LinkButton)
	btnOutlook := builder.get("btnMicrosoft").(gtki.LinkButton)

	_ = btnEmail.SetProperty("uri", h.getInvitationEmailURI())
	_ = btnGmail.SetProperty("uri", h.getInvitationGmailURI())
	_ = btnYahoo.SetProperty("uri", h.getInvitationYahooURI())
	_ = btnOutlook.SetProperty("uri", h.getInvitationMicrosoftURI())

	imagePixBuf, _ := h.u.g.getImagePixbufForSize("email.png")
	widgetImage, _ := h.u.g.gtk.ImageNewFromPixbuf(imagePixBuf)
	btnEmail.SetImage(widgetImage)

	imagePixBuf, _ = h.u.g.getImagePixbufForSize("gmail.png")
	widgetImage, _ = h.u.g.gtk.ImageNewFromPixbuf(imagePixBuf)
	btnGmail.SetImage(widgetImage)

	imagePixBuf, _ = h.u.g.getImagePixbufForSize("yahoo.png")
	widgetImage, _ = h.u.g.gtk.ImageNewFromPixbuf(imagePixBuf)
	btnYahoo.SetImage(widgetImage)

	imagePixBuf, _ = h.u.g.getImagePixbufForSize("outlook.png")
	widgetImage, _ = h.u.g.gtk.ImageNewFromPixbuf(imagePixBuf)
	btnOutlook.SetImage(widgetImage)

	cleanup := func() {
		dialog.Hide()
		h.u.enableCurrentWindow()
	}

	builder.ConnectSignals(map[string]interface{}{
		"on_close_window_signal": cleanup,
		"on_link_clicked":        cleanup,
		"on_copy_meeting_id": func() {
			h.copyMeetingIDToClipboard(builder, "")
		},
		"on_copy_invitation": func() {
			h.copyInvitationToClipboard(builder)
		},
	})

	h.u.doInUIThread(func() {
		h.u.disableCurrentWindow()
		dialog.Show()
	})
}
