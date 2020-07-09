package gui

import (
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
)

type cleanupHandler struct {
	u         *gtkUI
	callbacks []func()
}

func (u *gtkUI) initCleanupHandler() {
	u.cleanupHandler = &cleanupHandler{
		u: u,
	}

	u.cleanupHandler.initInterruptHandler()
}

func (h *cleanupHandler) initInterruptHandler() {
	c := make(chan os.Signal, 1)

	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		h.exitOnInterrupt()
	}()
}

func (h *cleanupHandler) exitOnInterrupt() {
	h.u.quit()
	os.Exit(0)
}

func (h *cleanupHandler) doCleanup(cb func()) {
	log.Debug("Cleaning Wahay...")

	if len(h.callbacks) != 0 {
		for _, cb := range h.callbacks {
			cb()
		}
	}

	cb()
}

func (h *cleanupHandler) add(cb func()) {
	h.callbacks = append(h.callbacks, cb)
}

func (u *gtkUI) onExit(cb func()) {
	u.cleanupHandler.add(cb)
}
