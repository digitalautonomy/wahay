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

func (u *gtkUI) displayLoadingWindow(finished chan bool) {
	builder := u.g.uiBuilderFor("LoadingWindow")
	win := builder.get("loadingWindow").(gtki.ApplicationWindow)

	u.currentWindow = win
	win.SetApplication(u.app)
	u.doInUIThread(win.ShowAll)

	<-finished

	u.doInUIThread(win.Hide)
}

func (u *gtkUI) hostMeetingHandler() {
	go u.realHostMeetingHandler()
}

func (u *gtkUI) realHostMeetingHandler() {
	u.doInUIThread(u.currentWindow.Hide)

	finished := make(chan bool)
	go func() {
		u.displayLoadingWindow(finished)
	}()

	server, tor, serviceID := u.createNewConferenceRoom()

	finished <- true

	builder := u.g.uiBuilderFor("StartHostingWindow")
	win := builder.get("startHostingWindow").(gtki.ApplicationWindow)
	builder.ConnectSignals(map[string]interface{}{
		"on_close_window_signal": u.quit,
		"on_finish_meeting": func() {
			if server != nil {
				u.finishMeeting(server, tor, serviceID)
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

	log.Println("Main: Completed")
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

func (u *gtkUI) reportError(message string) {
	builder := u.g.uiBuilderFor("GeneralError")
	dlg := builder.get("dialog").(gtki.MessageDialog)

	err := dlg.SetProperty("text", message)
	if err != nil {
		u.reportError(fmt.Sprintf("Programmer error #1: %s", err.Error()))
	}

	dlg.SetTransientFor(u.currentWindow)
	u.doInUIThread(func() {
		dlg.Run()
		dlg.Destroy()
	})
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

func (u *gtkUI) finishMeeting(s hosting.Server, cntrl tor.Control, serviceID string) {
	u.wouldYouConfirmFinishMeeting(func(res bool) {
		if res {
			log.Print("Close meeting...")
			err := s.Stop()
			if err != nil {
				log.Println(err)
				u.reportError(fmt.Sprintf("The meeting can't be closed: %s", err))
			}

			err = cntrl.DeleteOnionService(serviceID)
			if err != nil {
				log.Println(err)
				u.reportError(fmt.Sprintf("The onion service can't be deleted: %s", err))
			}

			u.doInUIThread(func() {
				u.currentWindow.Hide()
				u.currentWindow = u.mainWindow
				u.mainWindow.ShowAll()
			})
		}
	})
}

func (u *gtkUI) wouldYouConfirmFinishMeeting(k func(bool)) {
	builder := u.g.uiBuilderFor("FinishMeetingConfirmation")
	dialog := builder.get("finishMeeting").(gtki.MessageDialog)
	dialog.SetDefaultResponse(gtki.RESPONSE_NO)
	dialog.SetTransientFor(u.mainWindow)
	responseType := gtki.ResponseType(dialog.Run())
	result := responseType == gtki.RESPONSE_YES
	dialog.Destroy()
	k(result)
}
