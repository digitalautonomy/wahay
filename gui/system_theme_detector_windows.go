package gui

import (
	"errors"
	"sync"

	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

type colorManager struct {
	monitorCancel    chan struct{}
	monitorWaitGroup *sync.WaitGroup

	registryKey registry.Key
	keyEvent    windows.Handle

	ui *gtkUI
}

const (
	registryPath        = `SOFTWARE\Microsoft\Windows\CurrentVersion\Themes\Personalize`
	registryValueName   = "AppsUseLightTheme"
	registryWaitTimeout = 1000
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

var (
	user32                      = windows.NewLazySystemDLL("user32.dll")
	advapi32                    = windows.NewLazySystemDLL("advapi32.dll")
	procRegNotifyChangeKeyValue = advapi32.NewProc("RegNotifyChangeKeyValue")
)

func (cm *colorManager) init() {
	cm.enableAutomaticThemeChange()
}

func (cm *colorManager) monitorThemeChanges() {
	defer cm.monitorWaitGroup.Done()

	err := cm.initSystemResources()
	if err != nil {
		log.Errorf("Failed to initialize system resources: %v", err)
		return
	}
	defer cm.cleanupResources()

	cm.watchThemeLoop()
}

func (cm *colorManager) initSystemResources() error {
	var err error
	cm.registryKey, err = registry.OpenKey(registry.CURRENT_USER, registryPath, registry.NOTIFY)
	if err != nil {
		return errors.New("failed to open registry key: " + err.Error())
	}

	cm.keyEvent, err = windows.CreateEvent(nil, 0, 0, nil)
	if err != nil {
		cm.registryKey.Close()
		return errors.New("failed to create event: " + err.Error())
	}

	return nil
}

func (cm *colorManager) cleanupResources() {
	cm.registryKey.Close()
	windows.CloseHandle(cm.keyEvent)
}

func (cm *colorManager) watchThemeLoop() {
	for {
		if err := cm.notifyOnRegistryChange(); err != nil {
			log.Errorf("Registry notification failed: %v", err)
			return
		}

		if shouldExit := cm.handleThemeChange(); shouldExit {
			return
		}
	}
}

func (cm *colorManager) notifyOnRegistryChange() error {
	ret, _, err := procRegNotifyChangeKeyValue.Call(
		uintptr(cm.registryKey),
		0,
		uintptr(windows.REG_NOTIFY_CHANGE_LAST_SET),
		uintptr(cm.keyEvent),
		1,
	)

	if ret != 0 && err != nil {
		return errors.New("RegNotifyChangeKeyValue failed: " + err.Error())
	}
	return nil
}

func (cm *colorManager) handleThemeChange() bool {
	select {
	case <-cm.monitorCancel:
		return true
	default:
		waitResult, err := windows.WaitForSingleObject(cm.keyEvent, registryWaitTimeout)
		if err != nil {
			log.Errorf("WaitForSingleObject failed: %v", err)
			return false
		}

		if waitResult == windows.WAIT_OBJECT_0 {
			cm.updateTheme()
		}
		return false
	}
}

func (cm *colorManager) updateTheme() {
	css := "light-mode-gui"
	isDark := cm.isDarkThemeVariant()
	if isDark {
		css = "dark-mode-gui"
	}

	cm.ui.addCSSProvider(css)
}

func (cm *colorManager) disableAutomaticThemeChange() {
	if cm.monitorCancel == nil {
		return
	}

	close(cm.monitorCancel)

	if cm.monitorWaitGroup != nil {
		cm.monitorWaitGroup.Wait()
	}

	cm.monitorCancel = nil
	cm.monitorWaitGroup = nil
}

func (cm *colorManager) enableAutomaticThemeChange() {
	var wg sync.WaitGroup
	cm.monitorCancel = make(chan struct{})
	wg.Add(1)

	cm.updateTheme()

	cm.monitorWaitGroup = &wg
	go cm.monitorThemeChanges()
}
