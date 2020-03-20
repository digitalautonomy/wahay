package gui

import (
	"github.com/coyim/gotk3adapter/gtki"
)

// TODO[OB]: Why is this an exported type?

// FileType is the type for identifiying mime-types
type FileType string

func (u *gtkUI) setCustomFilePathFor(
	entry gtki.Entry,
	originalValue string,
	onSuccess func(string)) {
	go func() {
		ok, filename := u.getCustomFilePath()

		// The file chooser has been closed or no file has been selected
		if !ok {
			return
		}

		if originalValue != filename {
			onSuccess(filename)
			u.doInUIThread(func() {
				entry.SetText(filename)
			})
		}
	}()
}

func (u *gtkUI) getCustomFilePath() (ok bool, path string) {
	channel := make(chan string)
	errChannel := make(chan bool)
	go u.showCustomFilePathDialog(channel, errChannel)
	select {
	case v := <-channel:
		return true, v
	case <-errChannel:
		return false, ""
	}
}

func (u *gtkUI) showCustomFilePathDialog(channel chan string, errChannel chan bool) {
	u.doInUIThread(func() {
		dialog, err := u.g.gtk.FileChooserDialogNewWith2Buttons(
			i18n.Sprintf("Open file"),
			u.currentWindow,
			gtki.FILE_CHOOSER_ACTION_OPEN,
			i18n.Sprintf("Cancel"),
			gtki.RESPONSE_CANCEL,
			i18n.Sprintf("Open"),
			gtki.RESPONSE_ACCEPT)

		if err != nil {
			errChannel <- true
			return
		}

		chooser := (dialog).(gtki.FileChooser)
		chooser.SetDoOverwriteConfirmation(true)

		if u.currentWindow != nil {
			dialog.SetTransientFor(u.currentWindow)
		}

		dialog.Present()

		u.disableCurrentWindow()

		res := dialog.Run()

		if gtki.ResponseType(res) == gtki.RESPONSE_ACCEPT {
			channel <- dialog.GetFilename()
		} else {
			errChannel <- true
		}

		u.enableCurrentWindow()

		dialog.Destroy()
	})
}
