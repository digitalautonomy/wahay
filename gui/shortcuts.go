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
	u.connectShortcut("<Primary>F4", w, u.closeApplicationWindow)
	u.connectShortcut("<Primary>h", w, func(_ gtki.Window) {
		u.openHelpWindow()
	})

	if !u.errorHandler.isThereAnyStartupError() {
		u.connectShortcut("<Primary>i", w, func(_ gtki.Window) {
			u.hostMeetingHandler()
		})
		u.connectShortcut("<Primary>j", w, func(_ gtki.Window) {
			u.joinMeeting()
		})
		u.connectShortcut("<Primary>comma", w, func(_ gtki.Window) {
			u.openSettingsWindow()
		})
	}
}

func (u *gtkUI) connectShortcutsHostingMeetingConfigurationWindow(w gtki.Window, b *uiBuilder, h *hostData) {
	// <Primary> maps to Command and OS X, but Control on other platforms
	u.connectShortcut("<Primary>q", w, u.closeApplicationWindow)
	u.connectShortcut("<Primary>F4", w, u.closeWindow)
	u.connectShortcut("Escape", w, u.closeWindow)
	u.connectShortcut("<Primary>Return", w, func(_ gtki.Window) {
		h.handleOnStartMeeting(b)
	})
}

func (u *gtkUI) connectShortcutsStartHostingWindow(w gtki.Window, h *hostData) {
	// <Primary> maps to Command and OS X, but Control on other platforms
	u.connectShortcut("<Primary>q", w, u.closeApplicationWindow)
	u.connectShortcut("<Primary>F4", w, u.closeWindow)
	u.connectShortcut("Escape", w, u.closeWindow)
	u.connectShortcut("<Primary>w", w, func(_ gtki.Window) {
		h.finishMeeting()
	})
	u.connectShortcut("<Primary>j", w, func(_ gtki.Window) {
		h.u.hideCurrentWindow()
		go h.joinMeetingHost()
	})
	u.connectShortcut("<Primary>i", w, func(_ gtki.Window) {
		onInviteOpen := func(d gtki.Window) {
			h.currentWindow = d
			w.Hide()
		}
		onInviteClose := func(gtki.Window) {
			w.Show()
			h.currentWindow = nil
		}
		h.onInviteParticipants(onInviteOpen, onInviteClose)
	})

}

func (u *gtkUI) connectShortcutsCurrentHostMeetingWindow(w gtki.Window, h *hostData) {
	// <Primary> maps to Command and OS X, but Control on other platforms
	u.connectShortcut("<Primary>l", w, func(_ gtki.Window) {
		h.leaveHostMeeting()
	})
	u.connectShortcut("<Primary>w", w, func(_ gtki.Window) {
		h.finishMeetingMumble()
	})
}

func (u *gtkUI) connectShortcutsCurrentMeetingWindow(w gtki.Window, m tor.Service) {
	// <Primary> maps to Command and OS X, but Control on other platforms
	u.connectShortcut("<Primary>q", w, func(_ gtki.Window) {
		m.Close()
		u.closeApplicationWindow(w)
	})
	u.connectShortcut("<Primary>F4", w, func(_ gtki.Window) {
		u.leaveMeeting(m)
	})
	u.connectShortcut("Escape", w, func(_ gtki.Window) {
		u.leaveMeeting(m)
	})
	u.connectShortcut("<Primary>l", w, func(_ gtki.Window) {
		u.leaveMeeting(m)
	})
}

func (u *gtkUI) connectShortcutsInviteMeetingWindow(w gtki.Window, b *uiBuilder) {
	// <Primary> maps to Command and OS X, but Control on other platforms
	u.connectShortcut("<Primary>q", w, u.closeApplicationWindow)
	u.connectShortcut("<Primary>F4", w, u.closeWindow)
	u.connectShortcut("Escape", w, u.closeWindow)
	u.connectShortcut("<Primary>j", w, func(_ gtki.Window) {
		u.handleOnJoinMeeting(b)
	})
}

func (u *gtkUI) connectShortcutsSettingsWindow(w gtki.Window) {
	// <Primary> maps to Command and OS X, but Control on other platforms
	u.connectShortcut("<Primary>q", w, u.closeApplicationWindow)
}

func (u *gtkUI) connectShortcutsHelpWindow(w gtki.Window) {
	// <Primary> maps to Command and OS X, but Control on other platforms
	u.connectShortcut("<Primary>q", w, u.closeApplicationWindow)
	u.connectShortcut("<Primary>F4", w, func(_ gtki.Window) {
		u.closeHelpWindow(w)
	})
	u.connectShortcut("Escape", w, func(_ gtki.Window) {
		u.closeHelpWindow(w)
	})
}

func (u *gtkUI) closeApplicationWindow(_ gtki.Window) {
	u.quit()
}

func (u *gtkUI) closeWindow(_ gtki.Window) {
	u.switchToMainWindow()
}
