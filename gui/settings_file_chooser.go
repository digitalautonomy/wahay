package gui

import "github.com/coyim/gotk3adapter/gtki"

func (u *gtkUI) setCustomFilePathFor(entry gtki.Entry, originalValue string, onSuccess func(string)) {
	go func() {
		filename := u.getCustomFilePath()

		if originalValue != filename {
			onSuccess(filename)
			u.doInUIThread(func() {
				entry.SetText(filename)
			})
		}
	}()
}

func (u *gtkUI) getCustomFilePath() string {
	channel := make(chan string)
	go u.showCustomFilePathDialog(channel)
	return <-channel
}

func (u *gtkUI) showCustomFilePathDialog(channel chan string) {
	u.doInUIThread(func() {
		builder := u.g.uiBuilderFor("FileChooser")
		dialog := builder.get("dialog").(gtki.FileChooserDialog)

		if u.currentWindow != nil {
			dialog.SetTransientFor(u.currentWindow)
		}

		close := func() {
			u.enableCurrentWindow()
			dialog.Destroy()
			channel <- "no file"
		}

		builder.ConnectSignals(map[string]interface{}{
			"on_dialog_destroy": close,
		})

		u.disableCurrentWindow()

		dialog.Present()
		dialog.Show()
	})
}
