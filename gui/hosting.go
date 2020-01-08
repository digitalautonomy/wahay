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

	u.currentWindow = win
	win.SetApplication(u.app)
	u.doInUIThread(win.ShowAll)

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

	server, tor, serviceID := u.createNewConferenceRoom()

	loaded <- true

	u.showMeetingControls(server, tor, serviceID)
}

func (u *gtkUI) showMeetingControls(server hosting.Server, cntrl tor.Control, serviceID string) {
	builder := u.g.uiBuilderFor("StartHostingWindow")
	win := builder.get("startHostingWindow").(gtki.ApplicationWindow)
	builder.ConnectSignals(map[string]interface{}{
		"on_close_window_signal": u.quit,
		"on_finish_meeting": func() {
			if server != nil {
				u.finishMeeting(server, cntrl, serviceID)
			} else {
				log.Print("server is nil")
			}
		},
		"on_join_meeting": func() {
			if server != nil {
				u.joinMeetingHost(server, cntrl, serviceID)
			} else {
				log.Print("server is nil")
			}
		},
	})

	meetingID, err := builder.GetObject("lblMeetingID")
	if err != nil {
		log.Printf("meeting id error: %s", err)
	}
	_ = meetingID.SetProperty("label", serviceID)

	u.currentWindow = win
	win.SetApplication(u.app)
	u.doInUIThread(win.ShowAll)
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

func (u *gtkUI) joinMeetingHost(s hosting.Server, cntrl tor.Control, serviceID string) {
	if !isMeetingIDValid(serviceID) {
		u.reportError(fmt.Sprintf("invalid Onion Address %s", serviceID))
		return
	}

	state, err := launchMumbleClient(serviceID)
	if err != nil {
		u.reportError(fmt.Sprintf("Programmer error #1: %s", err.Error()))
		return
	}

	u.openHostJoinMeetingWindow(state, s, cntrl, serviceID)
}

func (u *gtkUI) openHostJoinMeetingWindow(state *runningMumble, s hosting.Server, cntrl tor.Control, serviceID string) {
	u.currentWindow.Hide()
	builder := u.g.uiBuilderFor("CurrentHostMeetingWindow")
	win := builder.get("hostMeetingWindow").(gtki.ApplicationWindow)
	win.SetApplication(u.app)
	u.currentWindow = win
	builder.ConnectSignals(map[string]interface{}{
		"on_close_window_signal": func() {
			u.leaveHostMeeting(state, s, cntrl, serviceID)
			u.quit()
		},
		"on_leave_meeting": func() {
			u.leaveHostMeeting(state, s, cntrl, serviceID)
		},
		"on_finish_meeting": func() {
			u.finishMeetingMumble(state, s, cntrl, serviceID)
		},
	})

	u.switchToHostOnFinishMeeting(state, s, cntrl, serviceID)
	win.ShowAll()
}

func (u *gtkUI) switchToHostOnFinishMeeting(
	state *runningMumble,
	s hosting.Server,
	cntrl tor.Control,
	serviceID string) {
	go func() {
		<-state.finishChannel

		// TODO: here, we  could check if the Mumble instance
		// failed with an error and report this
		u.doInUIThread(func() {
			if u.op == UIActionFinishMeeting {
				u.finishMeetingReal(s, cntrl, serviceID)
			} else if u.op == UIActionLeaveMeeting {
				u.currentWindow.Hide()
				u.showMeetingControls(s, cntrl, serviceID)
			} else {
				// Unknow UI action or not required by this phase
			}
			// Reset the custom ui action
			u.op = UIActionNone
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

func (u *gtkUI) createNewConferenceRoom() (hosting.Server, tor.Control, string) {
	port := randomPort()
	for !isPortAvailable(port) {
		port = randomPort()
	}

	u.ensureServerCollection()

	server, e := u.serverCollection.CreateServer(fmt.Sprintf("%d", port))
	if e != nil {
		u.reportError(fmt.Sprintf("Something went wrong: %s", e.Error()))
		return nil, nil, ""
	}
	e = server.Start()
	if e != nil {
		u.reportError(fmt.Sprintf("Something went wrong: %s", e.Error()))
		return nil, nil, ""
	}

	torController := tor.CreateController(*config.TorHost, *config.TorPort, *config.TorControlPassword)
	serviceID, e := torController.CreateNewOnionService("127.0.0.1", fmt.Sprintf("%d", port), "64738")
	if e != nil {
		u.reportError(fmt.Sprintf("Something went wrong: %s", e.Error()))
		return nil, nil, ""
	}

	return server, torController, serviceID
}

func (u *gtkUI) finishMeetingReal(s hosting.Server, cntrl tor.Control, serviceID string) {
	// Hide the current window
	u.doInUIThread(u.currentWindow.Hide)

	// TODO: What happen if two errors occurrs?
	// We need to do a better controlling for each error
	// and if multiple errors occurrs, show all the errors in the
	// same window using the `u.reportError` function

	err := s.Stop()
	if err != nil {
		u.reportError(fmt.Sprintf("The meeting can't be closed: %s", err))
	}

	err = cntrl.DeleteOnionService(serviceID)
	if err != nil {
		u.reportError(fmt.Sprintf("The onion service can't be deleted: %s", err))
	}

	u.doInUIThread(func() {
		u.currentWindow.Hide()
		u.currentWindow = u.mainWindow
		u.mainWindow.ShowAll()
	})
}

func (u *gtkUI) finishMeetingMumble(state *runningMumble, s hosting.Server, cntrl tor.Control, serviceID string) {
	u.wouldYouConfirmFinishMeeting(func(res bool) {
		if res {
			u.op = UIActionFinishMeeting
			go state.close()
		}
	})
}

func (u *gtkUI) finishMeeting(s hosting.Server, cntrl tor.Control, serviceID string) {
	u.wouldYouConfirmFinishMeeting(func(res bool) {
		if res {
			u.finishMeetingReal(s, cntrl, serviceID)
		}
	})
}

func (u *gtkUI) leaveHostMeeting(state *runningMumble, s hosting.Server, cntrl tor.Control, serviceID string) {
	u.op = UIActionLeaveMeeting

	// close the mumble instance
	go state.close()
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
