package gui

import (
	"time"

	"github.com/atotto/clipboard"
	"github.com/coyim/gotk3adapter/gtki"
)

const (
	gmailURL   = "https://mail.google.com/mail/"
	yahooURL   = "http://compose.mail.yahoo.com/"
	outlookURL = "https://dub130.mail.live.com/default.aspx"
)

func (u *gtkUI) switchToMainWindow() {
	if u.currentWindow != nil {
		u.currentWindow.Hide()
		u.currentWindow = nil
	}

	if u.mainWindow != nil {
		u.switchToWindow(u.mainWindow)
	}
}

func (u *gtkUI) switchToWindow(win gtki.ApplicationWindow) {
	win.SetApplication(u.app)
	u.setCurrentWindow(win)
	u.doInUIThread(win.Show)
}

func (u *gtkUI) showMainWindow() {
	if u.mainWindow != nil {
		u.switchToWindow(u.mainWindow)
	}
}

func (u *gtkUI) copyToClipboard(text string) error {
	return clipboard.WriteAll(text)
}

func (u *gtkUI) messageToLabel(label gtki.Label, message string, seconds int) {
	u.doInUIThread(func() {
		label.SetVisible(true)
		label.SetText(message)
	})

	time.Sleep(time.Duration(seconds) * time.Second)

	u.doInUIThread(func() {
		label.SetText("")
		label.SetVisible(false)
	})
}

func (u *gtkUI) enableWindow(win gtki.Window) {
	if win != nil {
		u.doInUIThread(func() {
			win.SetSensitive(true)
		})
	}
}

func (u *gtkUI) disableWindow(win gtki.Window) {
	if win != nil {
		u.doInUIThread(func() {
			win.SetSensitive(false)
		})
	}
}

func (u *gtkUI) disableMainWindow() {
	if u.mainWindow != nil {
		u.doInUIThread(func() {
			u.disableWindow(u.mainWindow)
		})
	}
}

func (u *gtkUI) enableMainWindow() {
	if u.mainWindow != nil {
		u.enableWindow(u.mainWindow)
	}
}

func (u *gtkUI) disableCurrentWindow() {
	if u.currentWindow != nil {
		u.disableWindow(u.currentWindow)
	}
}

func (u *gtkUI) enableCurrentWindow() {
	if u.currentWindow != nil {
		u.enableWindow(u.currentWindow)
	}
}

func (u *gtkUI) setCurrentWindow(win gtki.Window) {
	if u.currentWindow != win {
		u.currentWindow = win
	}
}

func (u *gtkUI) hideMainWindow() {
	if u.mainWindow != nil {
		u.doInUIThread(u.mainWindow.Hide)
	}
}

func (u *gtkUI) hideCurrentWindow() {
	if u.currentWindow != nil {
		u.doInUIThread(u.currentWindow.Hide)
	}
}
