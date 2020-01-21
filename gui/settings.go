package gui

import (
	"github.com/coyim/gotk3adapter/gtki"
)

type settings struct {
	u                          *gtkUI
	b                          *uiBuilder
	dialog                     gtki.Window
	chkAutojoin                gtki.CheckButton
	chkPersistentConfiguration gtki.CheckButton
}

func createSettings(u *gtkUI) *settings {
	builder := u.g.uiBuilderFor("GlobalSettings")
	dialog := builder.get("settingsWindow").(gtki.Window)

	s := &settings{
		u:      u,
		b:      builder,
		dialog: dialog,
	}

	s.b.getItems(
		"chkAutojoin", &s.chkAutojoin,
		"chkPersistentConfiguration", &s.chkPersistentConfiguration,
	)

	return s
}

func (u *gtkUI) openSettingsWindow() {
	s := createSettings(u)

	autoJoinOriginalValue := u.config.GetAutoJoin()
	s.chkAutojoin.SetActive(autoJoinOriginalValue)

	persisConfigFileOriginalValue := u.config.GetPersistentConfiguration()
	s.chkPersistentConfiguration.SetActive(persisConfigFileOriginalValue)

	s.b.ConnectSignals(map[string]interface{}{
		"on_toggle_option": func() {
			if s.chkAutojoin.GetActive() != autoJoinOriginalValue {
				u.config.SetAutoJoin(!autoJoinOriginalValue)
			}

			if s.chkPersistentConfiguration.GetActive() != persisConfigFileOriginalValue {
				u.config.SetPersistentConfiguration(!persisConfigFileOriginalValue)
			}
		},
		"on_save": func() {
			u.saveConfigOnly()
			s.dialog.Destroy()
		},
		"on_close_window": func() {
			s.dialog.Destroy()
		},
	})

	s.dialog.Show()
}
