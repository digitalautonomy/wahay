package gui

import (
	"fmt"
	"log"
	"math/rand"
	"net"

	"autonomia.digital/tonio/app/config"
	"autonomia.digital/tonio/app/hosting"
	"autonomia.digital/tonio/app/tor"
	"github.com/coyim/gotk3adapter/gtki"
)

func (u *gtkUI) displayLoadingWindow(loaded chan bool) {
	builder := u.g.uiBuilderFor("LoadingWindow")
	win := builder.get("loadingWindow").(gtki.ApplicationWindow)

	u.switchToWindow(win)

	<-loaded

	u.doInUIThread(win.Hide)
}

func (u *gtkUI) hostMeetingHandler() {
	go u.realHostMeetingHandler()
}

func (u *gtkUI) realHostMeetingHandler() {
	u.doInUIThread(u.currentWindow.Hide)

	loaded := make(chan bool)
	go func() {
		u.displayLoadingWindow(loaded)
	}()

	h := u.createNewConferenceRoom()

	loaded <- true

	h.showMeetingControls()
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
		"on_copy_meeting_id": h.copyMeetingIDToClipboard,
	})

	meetingID, err := builder.GetObject("lblMeetingID")
	if err != nil {
		log.Printf("meeting id error: %s", err)
	}
	_ = meetingID.SetProperty("label", h.serviceID)

	h.u.switchToWindow(win)
}

func isPortAvailable(port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))

	if err != nil {
		return false
	}

	return ln.Close() == nil
}

func randomPort() int {
	return 10000 + int(rand.Int31n(50000))
}

func (h *hostData) joinMeetingHost() {
	loaded := make(chan bool)
	go h.u.launchMumbleRoutineStart(loaded)

	state, err := launchMumbleClient(h.serviceID)
	if err != nil {
		h.u.reportError(fmt.Sprintf("Programmer error #1: %s", err.Error()))
		return
	}
	h.runningState = state

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

type hostData struct {
	u             *gtkUI
	runningState  *runningMumble
	serverControl hosting.Server
	torControl    tor.Control
	serviceID     string
	next          func()
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

func (u *gtkUI) createNewConferenceRoom() *hostData {
	port := randomPort()
	for !isPortAvailable(port) {
		port = randomPort()
	}

	u.ensureServerCollection()

	server, e := u.serverCollection.CreateServer(fmt.Sprintf("%d", port))
	if e != nil {
		u.reportError(fmt.Sprintf("Something went wrong: %s", e.Error()))
		return nil
	}
	e = server.Start()
	if e != nil {
		u.reportError(fmt.Sprintf("Something went wrong: %s", e.Error()))
		return nil
	}

	torController := tor.CreateController(*config.TorHost, *config.TorPort, *config.TorControlPassword)
	serviceID, e := torController.CreateNewOnionService("127.0.0.1", fmt.Sprintf("%d", port), "64738")
	if e != nil {
		u.reportError(fmt.Sprintf("Something went wrong: %s", e.Error()))
		return nil
	}

	h := &hostData{
		u:             u,
		serverControl: server,
		torControl:    torController,
		serviceID:     serviceID,
		next:          func() {},
	}
	return h
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

	err = h.torControl.DeleteOnionService(h.serviceID)
	if err != nil {
		h.u.reportError(fmt.Sprintf("The onion service can't be deleted: %s", err))
	}

	h.u.doInUIThread(func() {
		h.u.currentWindow.Hide()
		h.u.currentWindow = h.u.mainWindow
		h.u.mainWindow.ShowAll()
	})
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

func (h *hostData) copyMeetingIDToClipboard() {
	h.u.copyToClipboard(h.serviceID)
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
