package gui

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/coyim/gotk3adapter/gtki"
	"github.com/digitalautonomy/wahay/hosting"
	"github.com/digitalautonomy/wahay/tor"
)

type hostData struct {
	u                 *gtkUI
	mumble            tor.Service
	service           hosting.Service
	asSuperUser       bool
	superUserPassword string
	autoJoin          bool
	meetingUsername   string
	meetingPassword   string
	currentWindow     gtki.Window
	next              func()
}

func (u *gtkUI) hostMeetingHandler() {
	go u.realHostMeetingHandler()
}

func (u *gtkUI) realHostMeetingHandler() {
	u.hideMainWindow()
	u.displayLoadingWindow()

	if u.servers == nil {
		var err error
		u.servers, err = hosting.CreateServerCollection()
		if err != nil {
			// TODO: should we check if u.servers !== nil here?
			u.reportError(i18n().Sprintf("Something went wrong: %s", err))
			u.switchToMainWindow()
			return
		}
	}

	h := &hostData{
		u:           u,
		asSuperUser: u.config.GetAsSuperUser(),
		autoJoin:    u.config.GetAutoJoin(),
		next:        nil,
	}

	echan := make(chan error)

	go h.createNewService(echan)

	err := <-echan

	u.hideLoadingWindow()

	if err != nil {
		// TODO: we should check if u.servers !== nil to reset it
		h.u.reportError(i18n().Sprintf("Something went wrong: %s", err))
		u.switchToMainWindow()
		return
	}

	u.doInUIThread(h.showMeetingConfiguration)
}

func (h *hostData) showMeetingControls() {
	builder := h.u.g.uiBuilderFor("StartHostingWindow")
	win := builder.get("startHostingWindow").(gtki.ApplicationWindow)

	onInviteOpen := func(d gtki.Window) {
		h.currentWindow = d
		win.Hide()
	}

	onInviteClose := func(gtki.Window) {
		win.Show()
		h.currentWindow = nil
	}

	builder.i18nProperties(
		"label", "lblHostMeeting",
		"label", "lblInfoHost",
		"label", "lblInfoPassword",
		"label", "lblInfoMeetingID",
		"button", "btnFinishMeeting",
		"button", "btnJoinMeeting",
		"button", "btnJoinMeeting",
		"button", "btnInviteOthers",
		"button", "btnCopyMeetingID",
		"tooltip", "btnJoinMeeting",
		"tooltip", "btnInviteOthers")

	builder.ConnectSignals(map[string]interface{}{
		"on_close_window_signal": h.finishMeetingReal,
		"on_finish_meeting":      h.finishMeeting,
		"on_join_meeting": func() {
			h.u.hideCurrentWindow()
			go h.joinMeetingHost()
		},
		"on_invite_others": func() {
			h.onInviteParticipants(onInviteOpen, onInviteClose)
		},
		"on_copy_meeting_id": func() {
			h.copyMeetingIDToClipboard(builder, "")
		},
		"on_send_by_email": func() {
			h.sendInvitationByEmail(builder)
		},
	})

	lblValueHost := builder.get("lblValueHost").(gtki.Label)
	lblValuePassword := builder.get("lblValuePassword").(gtki.Label)
	lblValueMeetingID := builder.get("lblValueMeetingID").(gtki.Label)

	_ = lblValueHost.SetProperty("label", h.meetingUsername)
	_ = lblValuePassword.SetProperty("label", h.meetingPassword)
	_ = lblValueMeetingID.SetProperty("label", h.service.ID())
	h.u.connectShortcutsStartHostingWindow(win, h)
	h.u.switchToWindow(win)
}

func (h *hostData) joinMeetingHost() {
	h.u.displayLoadingWindow()

	validOpChannel := make(chan bool)

	go h.joinMeetingHostHelper(validOpChannel)

	if <-validOpChannel {
		h.openHostJoinMeetingWindow()
		return
	}

	// TODO: we should give more information to the user
	h.u.reportError(i18n().Sprintf("we couldn't start the meeting"))
	h.u.switchToMainWindow()
}

func (h *hostData) joinMeetingHostHelper(validOpChannel chan bool) {
	data := hosting.MeetingData{
		MeetingID: h.service.ID(),
		Port:      h.service.Port(),
		Password: func() string {
			if h.asSuperUser {
				return h.superUserPassword
			}
			return h.meetingPassword
		}(),
		Username: h.meetingUsername,
		IsHost:   true,
	}

	var err error
	var mumble tor.Service

	finish := make(chan bool)

	go func() {
		mumble, err = h.u.launchMumbleClient(
			data,
			// Callback to be executed when the client is closed
			func() {
				if h.next == nil {
					h.next = h.uiActionFinishMeeting
				}
				h.switchToHostOnFinishMeeting()
			})

		finish <- true
	}()

	<-finish // Wait for Mumble to start

	h.u.hideLoadingWindow()

	if err != nil {
		log.Errorf("joinMeetingHost() error: %s", err)
		validOpChannel <- false
	} else {
		h.mumble = mumble
		validOpChannel <- true
	}
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
		"tooltip", "btnLeaveMeeting",
		"button", "btnInviteOthers",
		"label", "lblTipPush",
	)

	return builder
}

func (h *hostData) openHostJoinMeetingWindow() {
	h.u.hideCurrentWindow()

	builder := h.u.getCurrentHostMeetingWindow()
	win := builder.get("hostMeetingWindow").(gtki.ApplicationWindow)
	builder.i18nProperties(
		"button", "btnInviteOthers",
		"tooltip", "btnInviteOthers")

	onInviteOpen := func(d gtki.Window) {
		h.currentWindow = d
		// Hide the current window because we don't want
		// lot of windows there in the user's screen
		win.Hide()
	}

	onInviteClose := func(gtki.Window) {
		win.Show()
		h.currentWindow = nil
	}

	builder.ConnectSignals(map[string]interface{}{
		"on_close_window_signal": func() {
			h.leaveHostMeeting()
			h.u.quit()
		},
		"on_leave_meeting":  h.leaveHostMeeting,
		"on_finish_meeting": h.finishMeetingMumble,
		"on_invite_others": func() {
			h.onInviteParticipants(onInviteOpen, onInviteClose)
		},
	})

	h.u.connectShortcutCurrentHostMeetingWindow(win, h)

	h.u.switchToWindow(win)
}

func (h *hostData) uiActionLeaveMeeting() {
	h.u.currentWindow.Hide()
	h.showMeetingControls()
}

func (h *hostData) uiActionFinishMeeting() {
	h.finishMeetingReal()
}

func (h *hostData) createNewService(err chan error) {
	var port string

	configuredPort := h.u.config.GetPortMumble()
	if configuredPort != "" {
		port = configuredPort
	}

	h.u.waitForTorInstance(func(t tor.Instance) {
		s, e := h.u.servers.NewService(port, t)
		if e != nil {
			log.Errorf("createNewService(): %s", e)
			err <- e
			return
		}

		s.SetWelcomeText(i18n().Sprintf("Welcome to this server running <b>Wahay</b>."))

		h.service = s

		err <- nil
	})
}

func (h *hostData) createNewConferenceRoom(complete chan bool) {
	var su hosting.SuperUserData
	if h.asSuperUser {
		su = hosting.SuperUserData{
			Username: h.meetingUsername,
			Password: h.superUserPassword,
		}
	}

	err := h.service.NewConferenceRoom(h.meetingPassword, su)
	if err != nil {
		h.u.hideLoadingWindow()
		h.u.reportError(i18n().Sprintf("Something went wrong: %s", err))
		complete <- false
		return
	}

	complete <- true
}

func (h *hostData) finishMeetingReal() {
	// TODO: What happens if two errors occurrs?
	// We need to do a better controlling for each error
	// and if multiple errors occurrs, show all the errors in the
	// same window using the `u.reportError` function
	err := h.service.Close()
	if err != nil {
		h.u.reportError(i18n().Sprintf("The meeting can't be closed: %s", err))
	}

	if h.currentWindow != nil {
		h.currentWindow.Destroy()
		h.currentWindow = nil
	}

	h.u.servers = nil

	h.u.switchToMainWindow()
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
	err := h.u.copyToClipboard(h.service.URL())
	if err != nil {
		h.u.reportError(err.Error())
		return
	}

	go func() {
		var lblMessage gtki.Label

		if len(label) == 0 {
			lblMessage = builder.get("lblMessage").(gtki.Label)
		} else {
			lblMessage = builder.get(label).(gtki.Label)
		}

		err = lblMessage.SetProperty("visible", false)
		if err != nil {
			panic(fmt.Sprintf("programmer error: %s", err))
		}

		h.u.messageToLabel(lblMessage, i18n().Sprintf("The meeting ID has been copied to the clipboard"), 5)
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
		h.u.messageToLabel(lblMessage, i18n().Sprintf("The invitation email has been copied to the clipboard"), 5)
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
	return i18n().Sprintf("Join Wahay Meeting")
}

func (h *hostData) getInvitationText() string {
	it := i18n().Sprintf("Please join the Wahay meeting with the following details:") + "%0D%0A%0D%0A"
	if h.service.URL() != "" {
		it = i18n().Sprintf("%sMeeting ID: %s", it, h.service.URL())
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
		"checkbox", "chkAutoJoinSuperUser",
		"tooltip", "chkAutoJoin",
		"tooltip", "chkAutoJoinSuperUser",
		"button", "btnCopyMeetingID",
		"button", "btnInviteOthers",
		"button", "btnCancel",
		"button", "btnStartMeeting")

	return builder
}

func (h *hostData) showMeetingConfiguration() {
	builder := h.u.getConfigureMeetingWindow()
	win := builder.get("configureMeetingWindow").(gtki.ApplicationWindow)
	chkAutoJoin := builder.get("chkAutoJoin").(gtki.CheckButton)
	chkAutoJoinSuperUser := builder.get("chkAutoJoinSuperUser").(gtki.CheckButton)
	btnStart := builder.get("btnStartMeeting").(gtki.Button)

	onInviteOpen := func(d gtki.Window) {
		h.currentWindow = d
		win.Hide()
	}

	onInviteClose := func(gtki.Window) {
		win.Show()
		h.currentWindow = nil
	}

	h.seti18nProperties(builder)

	chkAutoJoin.SetActive(h.autoJoin)
	chkAutoJoinSuperUser.SetActive(h.asSuperUser)
	h.changeStartButtonText(btnStart)

	btnCopyMeetingID := builder.get("btnCopyMeetingID").(gtki.Button)
	btnCopyMeetingID.SetVisible(h.u.isCopyToClipboardSupported())

	builder.ConnectSignals(map[string]interface{}{
		"on_copy_meeting_id": func() { h.copyMeetingIDToClipboard(builder, "") },
		"on_send_by_email":   func() { h.sendInvitationByEmail(builder) },
		"on_cancel": func() {
			h.handlerOnCancel()
		},
		"on_start_meeting": func() {
			h.handleOnStartMeeting(builder)
		},
		"on_invite_others": func() {
			h.onInviteParticipants(onInviteOpen, onInviteClose)
		},
		"on_chkAutoJoin_toggled": func() {
			h.handlerOnAutoJoinToggled(chkAutoJoin, btnStart)
		},
		"on_chkAutoJoinSuperUser_toggled": func() {
			h.handlerOnAutoJoinSuperUserToggled(chkAutoJoinSuperUser)
		},
	})

	h.u.connectShortcutsHostingMeetingConfigurationWindow(win, builder, h)

	meetingID, err := builder.GetObject("lblMeetingID")
	if err != nil {
		log.Printf("meeting id error: %s", err)
	}
	_ = meetingID.SetProperty("label", h.service.URL())

	h.u.switchToWindow(win)
}

func (h *hostData) handleOnStartMeeting(b *uiBuilder) {
	username := b.get("inpMeetingUsername").(gtki.Entry)
	password := b.get("inpMeetingPassword").(gtki.Entry)

	if h.asSuperUser {
		u, _ := username.GetText()
		if len(u) == 0 {
			h.u.reportError(i18n().Sprintf("The username is required"))
			return
		}
	}

	h.handlerOnStartMeeting(username, password)
}

func (h *hostData) seti18nProperties(b *uiBuilder) {
	b.i18nProperties(
		"label", "labelMeetingID",
		"label", "labelUsername",
		"label", "labelMeetingPassword",
		"label", "lblMessage")
}

func (h *hostData) handlerOnCancel() {
	_ = h.service.Close()
	h.u.servers = nil
	h.u.switchToMainWindow()
}

func (h *hostData) handlerOnAutoJoinSuperUserToggled(ch gtki.CheckButton) {
	h.asSuperUser = ch.GetActive()
	h.superUserPassword = generateRandomPassword()
	h.u.config.SetAutoJoinSuperUser(h.asSuperUser)
}

func (h *hostData) handlerOnAutoJoinToggled(ch gtki.CheckButton, b gtki.Button) {
	h.autoJoin = ch.GetActive()
	h.u.config.SetAutoJoin(h.autoJoin)
	h.changeStartButtonText(b)
}

func (h *hostData) handlerOnStartMeeting(u, p gtki.Entry) {
	h.meetingUsername, _ = u.GetText()
	h.meetingPassword, _ = p.GetText()

	h.startMeetingHandler()
}

func (h *hostData) changeStartButtonText(b gtki.Button) {
	if h.autoJoin {
		_ = b.SetProperty("label", i18n().Sprintf("Start Meeting & Join"))
		b.SetTooltipText(i18n().Sprintf("Start a new meeting \u0026 join"))
	} else {
		_ = b.SetProperty("label", i18n().Sprintf("Start Meeting"))
		b.SetTooltipText(i18n().Sprintf("Start a new meeting"))
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

	r := <-complete

	h.u.hideLoadingWindow()

	if !r {
		// TODO: show more useful information
		h.u.reportError(i18n().Sprintf("we couldn't start the meeting"))
		h.u.switchToMainWindow()
		return
	}

	if h.autoJoin {
		h.joinMeetingHost()
	} else {
		h.showMeetingControls()
	}
}

func (h *hostData) getInvitePeopleBuilder() *uiBuilder {
	builder := h.u.g.uiBuilderFor("InvitePeopleWindow")

	btnCopyMeetingID := builder.get("btnCopyMeetingID").(gtki.Button)
	btnCopyMeetingID.SetVisible(h.u.isCopyToClipboardSupported())

	btnCopyInvitation := builder.get("btnCopyInvitation").(gtki.Button)
	btnCopyInvitation.SetVisible(h.u.isCopyToClipboardSupported())

	builder.i18nProperties(
		"label", "lblDescription",
		"label", "lblDefaultEmail",
		"label", "lblGmail",
		"label", "lblYahoo",
		"label", "lblOutlook",
		"button", "btnCopyMeetingID",
		"button", "btnCopyInvitation")

	btnEmail := builder.get("btnEmail").(gtki.LinkButton)
	btnGmail := builder.get("btnGmail").(gtki.LinkButton)
	btnYahoo := builder.get("btnYahoo").(gtki.LinkButton)
	btnOutlook := builder.get("btnMicrosoft").(gtki.LinkButton)

	_ = btnEmail.SetProperty("uri", h.getInvitationEmailURI())
	_ = btnGmail.SetProperty("uri", h.getInvitationGmailURI())
	_ = btnYahoo.SetProperty("uri", h.getInvitationYahooURI())
	_ = btnOutlook.SetProperty("uri", h.getInvitationMicrosoftURI())

	imagePixBuf, _ := h.u.g.getImagePixbufForSize("email.png", 100)
	widgetImage, _ := h.u.g.gtk.ImageNewFromPixbuf(imagePixBuf)
	btnEmail.SetImage(widgetImage)

	imagePixBuf, _ = h.u.g.getImagePixbufForSize("gmail.png", 100)
	widgetImage, _ = h.u.g.gtk.ImageNewFromPixbuf(imagePixBuf)
	btnGmail.SetImage(widgetImage)

	imagePixBuf, _ = h.u.g.getImagePixbufForSize("yahoo.png", 100)
	widgetImage, _ = h.u.g.gtk.ImageNewFromPixbuf(imagePixBuf)
	btnYahoo.SetImage(widgetImage)

	imagePixBuf, _ = h.u.g.getImagePixbufForSize("outlook.png", 100)
	widgetImage, _ = h.u.g.gtk.ImageNewFromPixbuf(imagePixBuf)
	btnOutlook.SetImage(widgetImage)

	return builder
}

// TODO: review this function and make a more pretty solution
func (h *hostData) onInviteParticipants(onOpen func(d gtki.Window), onClose func(d gtki.Window)) {
	builder := h.getInvitePeopleBuilder()

	dialog := builder.get("invitePeopleWindow").(gtki.Window)

	if onClose == nil {
		onClose = func(gtki.Window) {
			h.u.enableCurrentWindow()
		}
	}

	cleanup := func() {
		dialog.Hide()
		onClose(dialog)
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

	if onOpen == nil {
		onOpen = func(gtki.Window) {
			h.u.disableCurrentWindow()
		}
	}

	h.u.doInUIThread(func() {
		dialog.Present()
		dialog.Show()

		onOpen(dialog)
	})
}

func generateRandomPassword() string {
	rand.Seed(time.Now().UnixNano())
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789")
	length := 8
	var b strings.Builder
	for i := 0; i < length; i++ {
		/* #nosec G404 */
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	return b.String()
}
