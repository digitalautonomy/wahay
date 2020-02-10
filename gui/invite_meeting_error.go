package gui

func (u *gtkUI) openErrorDialog(message string) {
	finalMessage := message
	if finalMessage == "" {
		finalMessage = i18n.Sprintf("The meeting ID is invalid")
	}
	u.reportError(finalMessage)
}
