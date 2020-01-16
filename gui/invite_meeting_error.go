package gui

func (u *gtkUI) openErrorDialog(message string) {
	finalMessage := message
	if finalMessage == "" {
		finalMessage = "The meeting ID is invalid"
	}
	u.reportError(finalMessage)
}
