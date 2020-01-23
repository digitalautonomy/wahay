package gui

import (
	"log"
	"path/filepath"

	"autonomia.digital/tonio/app/config"
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

func (u *gtkUI) loadConfig() {
	//u.config.WhenLoaded(u.configLoaded)

	filename := filepath.Join(config.Dir(), config.FileName)
	var conf *config.ApplicationConfig
	var err error
	if config.FileExists(filename) {
		conf, err = config.LoadOrCreate(filename)
		conf.SetPersistentConfiguration(true)
	} else {
		conf = config.CreateDefaultConfig()
		conf.SetPersistentConfiguration(false)
	}

	if err != nil {
		log.Fatal("We can't load the config file")
	}

	u.config = conf
	u.doInUIThread(u.initialSetupWindow)
}

func (u *gtkUI) saveConfigOnlyInternal() error {
	return u.config.Save()
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
