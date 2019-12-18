package gui

import (
	"fmt"
	"math/rand"
	"net"

	"autonomia.digital/tonio/app/config"
	"autonomia.digital/tonio/app/hosting"
	"autonomia.digital/tonio/app/tor"
	"github.com/coyim/gotk3adapter/gtki"
)

func (u *gtkUI) hostMeetingHandler() {
	var server hosting.Server
	go func() {
		server = u.createNewConferenceRoom()
	}()

	u.currentWindow.Hide()

	builder := u.g.uiBuilderFor("StartHostingWindow")
	win := builder.get("startHostingWindow").(gtki.ApplicationWindow)
	builder.ConnectSignals(map[string]interface{}{
		"on_close_window_signal": u.quit,
	})
	u.currentWindow = win
	win.SetApplication(u.app)
	win.ShowAll()
}

func isPortAvailable(port int) bool {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))

	if err != nil {
		return false
	}

	err = ln.Close()
	if err != nil {
		return false
	}

	return true
}

func randomPort() int {
	return 10000 + int(rand.Int31n(50000))
}

func (u *gtkUI) reportError(message string) {
	builder := u.g.uiBuilderFor("GeneralError")
	dlg := builder.get("dialog").(gtki.MessageDialog)

	dlg.SetProperty("text", message)
	dlg.SetTransientFor(u.currentWindow)

	dlg.Run()
	dlg.Destroy()
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

func (u *gtkUI) createNewConferenceRoom() hosting.Server {
	port := randomPort()
	for !isPortAvailable(port) {
		port = randomPort()
	}

	u.ensureServerCollection()

	server, e := u.serverCollection.CreateServer(fmt.Sprintf("%d", port))
	if e != nil {
		u.reportError(fmt.Sprintf("Something went wrong: %s", e.Error()))
	}
	e = server.Start()
	if e != nil {
		u.reportError(fmt.Sprintf("Something went wrong: %s", e.Error()))
	}

	torController := tor.CreateController(*config.TorHost, *config.TorPort, *config.TorControlPassword)
	serviceID, e := torController.CreateNewOnionService("127.0.0.1", fmt.Sprintf("%d", port), "64738")
	if e != nil {
		u.reportError(fmt.Sprintf("Something went wrong: %s", e.Error()))
	}

	fmt.Printf("created new conference room at: %s\n", serviceID)

	return server
}
