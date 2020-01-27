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

	autoJoinOriginalValue          bool
	persistConfigFileOriginalValue bool
	encryptFileOriginalValue       bool
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

	s.init(u.config)

	return s
}

func (s *settings) init(conf *config.ApplicationConfig) {
	s.autoJoinOriginalValue = conf.GetAutoJoin()
	s.chkAutojoin.SetActive(s.autoJoinOriginalValue)

	s.persistConfigFileOriginalValue = conf.GetPersistentConfiguration()
	s.chkPersistentConfiguration.SetActive(s.persistConfigFileOriginalValue)
	s.lblMessage.SetVisible(!s.persistConfigFileOriginalValue)

	s.encryptFileOriginalValue = conf.ShouldEncrypt()
	s.chkEncryptFile.SetActive(s.encryptFileOriginalValue)
	s.chkEncryptFile.SetSensitive(s.persistConfigFileOriginalValue)
}

var (
	decryptUncheckConfirmText = "If you disable this option, anyone could read your configuration settings"
)

func (u *gtkUI) onSettingsToggleOption(s *settings) {
	if s.chkAutojoin.GetActive() != s.autoJoinOriginalValue {
		u.config.SetAutoJoin(!s.autoJoinOriginalValue)
		s.autoJoinOriginalValue = !s.autoJoinOriginalValue
	}

	if s.chkPersistentConfiguration.GetActive() != s.persistConfigFileOriginalValue {
		s.lblMessage.SetVisible(s.persistConfigFileOriginalValue)
		u.config.SetPersistentConfiguration(!s.persistConfigFileOriginalValue)
		s.persistConfigFileOriginalValue = !s.persistConfigFileOriginalValue
		s.chkEncryptFile.SetSensitive(s.persistConfigFileOriginalValue)
	}

	if s.chkEncryptFile.GetActive() != s.encryptFileOriginalValue {
		if s.encryptFileOriginalValue {
			u.showConfirmation(func(op bool) {
				if op {
					s.encryptFileOriginalValue = false
					u.config.SetShouldEncrypt(false)
					s.chkEncryptFile.SetActive(false)
				} else {
					// We keep the checkbutton checked. Nothing else change.
					s.chkEncryptFile.SetActive(true)
				}
			}, decryptUncheckConfirmText)
		} else {
			u.captureMasterPassword(func() {
				s.encryptFileOriginalValue = true
				u.config.SetShouldEncrypt(true)
				u.saveConfigOnly()
			}, func() {
				s.chkEncryptFile.SetActive(false)
				u.config.SetShouldEncrypt(false)
			})
		}
	}
}

func (u *gtkUI) openSettingsWindow() {
	s := createSettings(u)

	s.b.ConnectSignals(map[string]interface{}{
		"on_toggle_option": func() {
			u.onSettingsToggleOption(s)
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

	if conf.GetPersistentConfiguration() {
		repeat := true
		for repeat {
			repeat, err = conf.Load(configFile, u.keySupplier)
			if err != nil {
				log.Println(err)
			}

			if repeat {
				u.keySupplier.Invalidate()
				u.keySupplier.LastAttemptFailed()
			}
		}
	} else {
		_, err = conf.Load(configFile, u.keySupplier)
		if err != nil {
			log.Fatal(err)
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
