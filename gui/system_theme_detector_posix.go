//go:build !windows

package gui

import (
	"os"
	"strings"
	"sync"

	"github.com/coyim/gotk3adapter/glibi"
	"github.com/coyim/gotk3adapter/gtki"
)

type colorManager struct {
	themeVariant          string
	calculateThemeVariant sync.Once
	onThemeChange         *callbacksSet
	ui                    *gtkUI
}

type callbacksSet struct {
	callbacks []func()
	sync.Mutex
}

func newCallbacksSet(callbacks ...func()) *callbacksSet {
	return &callbacksSet{
		callbacks: callbacks,
	}
}

func (s *callbacksSet) add(callbacks ...func()) {
	s.Lock()
	defer s.Unlock()
	s.callbacks = append(s.callbacks, callbacks...)
}

func (s *callbacksSet) invokeAll() {
	s.Lock()
	defer s.Unlock()

	for _, cb := range s.callbacks {
		cb()
	}
}

const (
	darkThemeVariantName  = "dark"
	lightThemeVariantName = "light"
)

func (cm *colorManager) init() {
	cm.onThemeChange = newCallbacksSet()

	set := cm.getGSettings()
	_ = set.Connect("changed::gtk-theme", cm.onThemeChange.invokeAll)

	cm.onThemeChange.add(func() {
		cm.calculateThemeVariant = sync.Once{}
	})
}

func (cm *colorManager) isDarkMode() bool {
	return cm.detectDarkThemeFromEnvironmentVariable() ||
		cm.detectDarkThemeFromGTKSettings() ||
		cm.detectDarkThemeFromGTKSettingsThemeName() ||
		cm.detectDarkThemeFromGSettingsThemeName()
}

func (cm *colorManager) detectDarkThemeFromEnvironmentVariable() bool {
	gtkTheme := os.Getenv("GTK_THEME")
	return doesThemeNameIndicateDarkness(gtkTheme)
}

func doesThemeNameIndicateDarkness(themeName string) bool {
	return isDarkVariantNameBasedOnSeparator(themeName, ":") ||
		isDarkVariantNameBasedOnSeparator(themeName, "-") ||
		isDarkVariantNameBasedOnSeparator(themeName, "_")
}

func isDarkVariantNameBasedOnSeparator(name, separator string) bool {
	parts := strings.Split(name, separator)
	if len(parts) < 2 {
		return false
	}
	variant := parts[len(parts)-1]
	return variant == darkThemeVariantName
}

func (cm *colorManager) detectDarkThemeFromGTKSettings() bool {
	// TODO: this might not be safe to do outside the UI thread
	prefDark, _ := cm.getGTKSettings().GetProperty("gtk-application-prefer-dark-theme")
	val, ok := prefDark.(bool)
	return val && ok
}

func (cm *colorManager) getThemeNameFromGTKSettings() string {
	// TODO: this might not be safe to do outside the UI thread
	themeName, _ := cm.getGTKSettings().GetProperty("gtk-theme-name")
	val, _ := themeName.(string)
	return val
}

func (cm *colorManager) getGTKSettings() gtki.Settings {
	settings, err := g.gtk.SettingsGetDefault()
	if err != nil {
		panic(err)
	}
	return settings
}

func (cm *colorManager) detectDarkThemeFromGTKSettingsThemeName() bool {
	return doesThemeNameIndicateDarkness(cm.getThemeNameFromGTKSettings())
}

func (cm *colorManager) detectDarkThemeFromGSettingsThemeName() bool {
	return doesThemeNameIndicateDarkness(cm.getThemeNameFromGSettings())

}

func (cm *colorManager) getThemeNameFromGSettings() string {
	// TODO: this might not be safe to do outside the UI thread
	return cm.getGSettings().GetString("gtk-theme")
}

func (cm *colorManager) getGSettings() glibi.Settings {
	return g.glib.SettingsNew("org.gnome.desktop.interface")
}

func (cm *colorManager) actuallyCalculateThemeVariant() {
	if cm.isDarkMode() {
		cm.themeVariant = darkThemeVariantName
	} else {
		cm.themeVariant = lightThemeVariantName
	}
}

func (cm *colorManager) getThemeVariant() string {
	cm.calculateThemeVariant.Do(cm.actuallyCalculateThemeVariant)
	return cm.themeVariant
}

func (cm *colorManager) isDarkThemeVariant() bool {
	return cm.getThemeVariant() == darkThemeVariantName
}
