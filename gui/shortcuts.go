package gui

import (
	"github.com/coyim/gotk3adapter/gtki"
	"github.com/digitalautonomy/wahay/tor"
)

func (u *gtkUI) connectShortcut(accel string, w gtki.Window, action func(gtki.Window)) {
	gr, _ := u.g.gtk.AccelGroupNew()
	key, mod := u.g.gtk.AcceleratorParse(accel)

	gr.Connect2(key, mod, gtki.ACCEL_VISIBLE, func() {
		action(w)
	})

	w.AddAccelGroup(gr)
}

func (u *gtkUI) connectShortcutsMainWindow(w gtki.Window) {
	// <Primary> maps to Command and OS X, but Control on other platforms
	u.connectShortcut("<Primary>q", w, u.closeApplicationWindow)
	u.connectShortcut("<Primary>Q", w, u.closeApplicationWindow)
	u.connectShortcut("<Alt>F4", w, u.closeApplicationWindow)
	u.connectShortcut("Escape", w, u.closeApplicationWindow)
	u.connectShortcut("<Primary>F4", w, u.closeApplicationWindow)
	u.connectShortcut("<Primary>i", w, func(w gtki.Window) {
		u.hostMeetingHandler()
	})
	u.connectShortcut("<Primary>I", w, func(w gtki.Window) {
		u.hostMeetingHandler()
	})
	u.connectShortcut("<Primary>j", w, func(w gtki.Window) {
		u.joinMeeting()
	})
	u.connectShortcut("<Primary>J", w, func(w gtki.Window) {
		u.joinMeeting()
	})
}

func (u *gtkUI) connectShortcutsHostingMeetingConfigurationWindow(w gtki.Window, b *uiBuilder, h *hostData) {
	// <Primary> maps to Command and OS X, but Control on other platforms
	u.connectShortcut("<Primary>q", w, u.closeApplicationWindow)
	u.connectShortcut("<Primary>Q", w, u.closeApplicationWindow)
	u.connectShortcut("<Primary>F4", w, u.closeWindow)
	u.connectShortcut("Escape", w, u.closeWindow)
	u.connectShortcut("<Alt>F4", w, u.closeApplicationWindow)
	u.connectShortcut("<Primary>Return", w, func(w gtki.Window) {
		h.handleOnStartMeeting(b)
	})
}

func (u *gtkUI) connectShortcutCurrentHostMeetingWindow(w gtki.Window, h *hostData) {
	// <Primary> maps to Command and OS X, but Control on other platforms
	u.connectShortcut("<Primary>l", w, func(w gtki.Window) {
		h.leaveHostMeeting()
	})
	u.connectShortcut("<Primary>L", w, func(w gtki.Window) {
		h.leaveHostMeeting()
	})
	u.connectShortcut("<Primary>w", w, func(w gtki.Window) {
		h.finishMeeting()
	})
	u.connectShortcut("<Primary>W", w, func(w gtki.Window) {
		h.finishMeeting()
	})
}

func (u *gtkUI) connectShortcutCurrentMeetingWindow(w gtki.Window, m tor.Service) {
	// <Primary> maps to Command and OS X, but Control on other platforms
	u.connectShortcut("<Primary>q", w, func(w gtki.Window) {
		u.leaveMeeting(m)
	})
	u.connectShortcut("<Primary>Q", w, func(w gtki.Window) {
		u.leaveMeeting(m)
	})
}

func (u *gtkUI) closeApplicationWindow(w gtki.Window) {
	u.quit()
}

func (u *gtkUI) closeWindow(w gtki.Window) {
	u.switchToMainWindow()
}
