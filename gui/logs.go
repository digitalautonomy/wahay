package gui

import (
	"io"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/digitalautonomy/wahay/config"
)

func (u *gtkUI) initLogs() {
	if u.config == nil {
		return
	}

	if !u.config.IsLogsEnabled() {
		return
	}

	rawLogFile := ensureLogFile(u.config)
	if rawLogFile != nil {
		log.SetOutput(rawLogFile)
	}
}

func ensureLogFile(conf *config.ApplicationConfig) io.Writer {
	configuredLogFile := conf.GetRawLogFile()

	if len(configuredLogFile) == 0 {
		configuredLogFile = config.GetDefaultLogFile()
	}

	rawLogFile, err := ensureLogFileDirectory(configuredLogFile)
	if len(rawLogFile) == 0 || err != nil {
		return nil
	}

	file, err := os.OpenFile(rawLogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return nil
	}

	return file
}

func ensureLogFileDirectory(rawLogFile string) (string, error) {
	fileInfo, err := getLogFileInfo(rawLogFile, config.GetDefaultLogFile())
	if err != nil {
		return "", err
	}

	if fileInfo.IsDir() {
		return filepath.Join(rawLogFile, config.GetDefaultLogFileName()), nil
	}

	return rawLogFile, nil
}

func getLogFileInfo(rawLogFile string, defaultFile string) (os.FileInfo, error) {
	if len(rawLogFile) == 0 {
		rawLogFile = defaultFile
	}

	fileInfo, err := os.Stat(rawLogFile)
	if fileInfo == nil || err != nil {
		_, err := os.Create(rawLogFile)
		if err != nil {
			return nil, err
		}
	}

	return os.Stat(rawLogFile)
}
