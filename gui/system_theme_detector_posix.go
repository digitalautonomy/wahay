//go:build !windows

package gui

import (
	"bufio"
	"bytes"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
)

func isDarkMode() bool {
	cmd := exec.Command("gsettings", "get", "org.gnome.desktop.interface", "color-scheme")
	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return false
	}

	output := out.String()
	return strings.Contains(output, "dark")
}

func (s *settings) monitorSystemStyleChanges() {
	var css string
	cmd := exec.Command("gsettings", "monitor", "org.gnome.desktop.interface", "color-scheme")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Errorf("Failed to create stdout pipe: %v", err)
		return
	}

	if err := cmd.Start(); err != nil {
		log.Errorf("Failed to start gsettingsmonitor: %v", err)
		return
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		log.Debug(line)
		if strings.Contains(line, "dark") {
			css = "dark-mode-gui"
		} else {
			css = "light-mode-gui"
		}
		s.u.addCSSProvider(css)
	}

	if err := scanner.Err(); err != nil {
		log.Errorf("Failed to read scanner: %v", err)
	}
}
