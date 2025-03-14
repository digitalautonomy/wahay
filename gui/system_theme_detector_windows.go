package gui

import (
	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/windows/registry"
)

func isDarkMode() bool {
	key, err := registry.OpenKey(registry.CURRENT_USER, `SOFTWARE\Microsoft\Windows\CurrentVersion\Themes\Personalize`, registry.QUERY_VALUE)
	if err != nil {
		log.Printf("Failed to open registry key: %v", err)
		return false
	}
	defer key.Close()

	value, _, err := key.GetIntegerValue("AppsUseLightTheme")
	if err != nil {
		log.Printf("Failed to read registry value: %v", err)
		return false
	}

	return value == 0
}

func (s *settings) monitorSystemStyleChanges() {}

func (s *settings) stopMonitoring() {}
