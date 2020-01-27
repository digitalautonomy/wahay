package gui

import (
	"log"

	"autonomia.digital/tonio/app/config"
	"github.com/coyim/gotk3adapter/gtki"
)

type settings struct {
	u                          *gtkUI
	b                          *uiBuilder
	dialog                     gtki.Window
	chkAutojoin                gtki.CheckButton
	chkPersistentConfiguration gtki.CheckButton
	chkEncryptFile             gtki.CheckButton
	lblMessage                 gtki.Label
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
		"chkEncryptFile", &s.chkEncryptFile,
		"lblMessage", &s.lblMessage,
	)

	return s
}

func (u *gtkUI) openSettingsWindow() {
	s := createSettings(u)

	autoJoinOriginalValue := u.config.GetAutoJoin()
	s.chkAutojoin.SetActive(autoJoinOriginalValue)

	persistConfigFileOriginalValue := u.config.GetPersistentConfiguration()
	s.chkPersistentConfiguration.SetActive(persistConfigFileOriginalValue)
	s.lblMessage.SetVisible(!persistConfigFileOriginalValue)

	encryptFileOriginalValue := u.config.ShouldEncrypt()
	s.chkEncryptFile.SetActive(encryptFileOriginalValue)
	s.chkEncryptFile.SetSensitive(persistConfigFileOriginalValue)

	s.b.ConnectSignals(map[string]interface{}{
		"on_toggle_option": func() {
			if s.chkAutojoin.GetActive() != autoJoinOriginalValue {
				u.config.SetAutoJoin(!autoJoinOriginalValue)
				autoJoinOriginalValue = !autoJoinOriginalValue
			}

			if s.chkPersistentConfiguration.GetActive() != persistConfigFileOriginalValue {
				s.lblMessage.SetVisible(persistConfigFileOriginalValue)
				u.config.SetPersistentConfiguration(!persistConfigFileOriginalValue)
				persistConfigFileOriginalValue = !persistConfigFileOriginalValue
				s.chkEncryptFile.SetSensitive(persistConfigFileOriginalValue)
			}

			if s.chkEncryptFile.GetActive() != encryptFileOriginalValue {
				encryptFileOriginalValue = !encryptFileOriginalValue
				u.config.SetShouldEncrypt(encryptFileOriginalValue)
				if encryptFileOriginalValue {
					u.captureMasterPassword(u.saveConfigOnly, func() {
						s.chkEncryptFile.SetActive(false)
						u.config.SetShouldEncrypt(false)
					})
				}
			}
		},
		"on_save": func() {
			u.saveConfigOnly()
			s.dialog.Destroy()
		},
		"on_close_window": func() {
			u.enableWindow(u.mainWindow)
			s.dialog.Destroy()
		},
		"on_destroy": func() {
			u.enableWindow(u.mainWindow)
		},
	})

	s.dialog.SetTransientFor(u.mainWindow)
	u.disableWindow(u.mainWindow)
	u.switchToWindow(s.dialog)
}

func (u *gtkUI) loadConfig() {
	var err error

	conf := config.New()

	conf.WhenLoaded(func(c *config.ApplicationConfig) {
		u.config = c
		u.doInUIThread(u.initialSetupWindow)
		u.configLoaded()
	})

	configFile, err := conf.Init()
	if err != nil {
		log.Fatal("the configuration file can't be initialized")
	}

	repeat := true
	for repeat {
		repeat, err = conf.Load(configFile, u.keySupplier)
		if repeat {
			u.keySupplier.Invalidate()
			u.keySupplier.LastAttemptFailed()
		}
	}
}

func (u *gtkUI) saveConfigOnlyInternal() error {
	return u.config.Save(u.keySupplier)
}

func (u *gtkUI) saveConfigOnly() {
	// Don't save the configuration file if the user doesn't want it
	if !u.config.GetPersistentConfiguration() {
		u.config.DeleteFileIfExists()
		return
	}

	go func() {
		err := u.saveConfigOnlyInternal()
		if err != nil {
			log.Println("Failed to save config file:", err.Error())
		}
	}()
}
