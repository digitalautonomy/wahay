package gui

import log "github.com/sirupsen/logrus"

// GTK process events in glib event loop (see [1]). In order to keep the UI
// responsive, it is a good practice to not block long running tasks in a signal's
// callback (you dont want a button to keep looking pressed for a couple of seconds).
// doInUIThread schedule the function to run in the next
// 1 - https://developer.gnome.org/glib/unstable/glib-The-Main-Event-Loop.html
func (u *gtkUI) doInUIThread(f func()) {
	_, err := u.g.glib.IdleAdd(f)
	if err != nil {
		log.Errorf("GTK thread error: %s", err)
	}
}
