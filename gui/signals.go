package gui

import (
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
)

type cleanupHandler struct{}

func (u *gtkUI) initCleanupHandler() {
	u.cleanupHandler = &cleanupHandler{}
	u.initInterruptHandler()
}

func (u *gtkUI) initInterruptHandler() {
	c := make(chan os.Signal, 1)

	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		u.exitOnInterrupt()
	}()
}

func (u *gtkUI) exitOnInterrupt() {
	log.Println("Closing Wahay")
	u.quit()
	os.Exit(0)
}

func (u *gtkUI) quit() {
	u.cleanup()
	u.app.Quit()
}

func (u *gtkUI) cleanup() {
	if u.tor != nil {
		// TODO: delete any created Onion Service
		u.tor.Destroy()
	}

	if u.client != nil {
		// TODO: we should remove any Mumble command running
		// and we should close the Grumble service if it's running
		u.client.Destroy()
	}
}
