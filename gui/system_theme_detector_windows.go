package gui

import (
	"errors"

	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/windows/registry"
)

type colorManager struct{}

const (
	registryPath      = `SOFTWARE\Microsoft\Windows\CurrentVersion\Themes\Personalize`
	registryValueName = "AppsUseLightTheme"
)

func (cm *colorManager) isDarkThemeVariant() bool {
	isDark, err := isDarkMode()
	if err != nil {
		log.Printf("Error checking dark mode: %v", err)
		return false
	}

	return isDark
}

func isDarkMode() (bool, error) {
	key, err := registry.OpenKey(registry.CURRENT_USER, registryPath, registry.QUERY_VALUE)
	if err != nil {
		return false, errors.New("failed to open registry key")
	}
	defer key.Close()

	value, _, err := key.GetIntegerValue(registryValueName)
	if err != nil {
		return false, errors.New("failed to read registry value")
	}

	return value == 0, nil
}

func (cm *colorManager) init() {}
